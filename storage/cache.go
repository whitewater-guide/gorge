package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/whitewater-guide/gorge/core"
)

// RedisCacheManager is cache manager that uses real redis
type RedisCacheManager struct {
	pool    *redis.Pool
	address string
}

// EmbeddedCacheManager is cache manager that uses embedded redis https://github.com/alicebob/miniredis
type EmbeddedCacheManager struct {
	srv *miniredis.Miniredis
	RedisCacheManager
}

const (
	// NSStatus is redis namespace prefix for job/gauge statuses
	// Cache structure is following hash:
	// key             -> field               -> value
	// -------------------- one-by-one jobs ---------------------
	// status:<jobId>  -> <code>:time         -> ISO time string of latest job/code execution
	//                 -> <code>:success      -> ISO time string of latest SUCCESSFULL job/code execution
	//                 -> <code>:count        -> number of measurements harvested last time
	//                 -> <code>:error        -> latest execution error, or empty string in case of success
	//
	// -------------------- jobs --------------------
	// status:jobs     -> <jobId>:time        -> ISO time string of latest job/code execution
	//                 -> <jobId>:success     -> ISO time string of latest SUCCESSFULL job/code execution
	//                 -> <jobId>:count       -> number of measurements harvested last time
	//                 -> <jobId>:error       -> latest execution error, or empty string in case of success
	NSStatus = "status"
	// NSLatest is redis namespace prefix for latest measurements
	NSLatest = "latest"
)

func (cache *RedisCacheManager) loadStatuses(jobID string) (map[string]core.Status, error) {
	conn := cache.pool.Get()
	defer conn.Close()
	var key string
	if jobID != "" {
		key = NSStatus + ":" + jobID // get gauges statuses of this jobID
	} else {
		key = NSStatus + ":" + "jobs" // get statuses of all jobs
	}
	m, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		return nil, core.WrapErr(err, "failed to get statuses")
	}

	result := make(map[string]core.Status)
	for field, value := range m {
		pts := strings.Split(field, ":")
		id, prop := pts[0], pts[1] // id is job_id or code, prop is time|sucess... etc.

		var status core.Status
		var ok bool
		if status, ok = result[id]; !ok {
			status = core.Status{}
		}

		switch prop {
		case "time":
			ts, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return nil, core.WrapErr(err, fmt.Sprintf("failed parse time '%s' from redis", value))
			}
			status.LastRun = core.HTime{Time: ts}
		case "success":
			ts, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return nil, core.WrapErr(err, fmt.Sprintf("failed parse success time '%s' from redis", value))
			}
			status.LastSuccess = &core.HTime{Time: ts}
		case "count":
			count, err := strconv.Atoi(value)
			if err != nil {
				return nil, core.WrapErr(err, fmt.Sprintf("failed parse count '%s' from redis", value))
			}
			status.Count = count
		case "error":
			status.Error = value
		}

		result[id] = status
	}
	return result, nil
}

// LoadJobStatuses implements CacheManager interface
func (cache *RedisCacheManager) LoadJobStatuses() (map[string]core.Status, error) {
	return cache.loadStatuses("")
}

// LoadGaugeStatuses implements CacheManager interface
func (cache *RedisCacheManager) LoadGaugeStatuses(jobID string) (map[string]core.Status, error) {
	return cache.loadStatuses(jobID)
}

func (cache *RedisCacheManager) saveStatusWithTime(jobID, code string, err error, count int, ts time.Time) error {
	conn := cache.pool.Get()
	defer conn.Close()
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	// status := core.Status{
	// 	Success:   success,
	// 	Timestamp: core.HTime{Time: ts},
	// 	Error:     errStr,
	// 	Count:     count,
	// }
	// raw, err := json.Marshal(status)
	// if err != nil {
	// 	return core.WrapErr(err, "failed to marshal status")
	// }
	var key, prefix string
	if code == "" {
		key = fmt.Sprintf("%s:jobs", NSStatus)
		prefix = jobID
	} else {
		key = fmt.Sprintf("%s:%s", NSStatus, jobID)
		prefix = code
	}

	hmset := []interface{}{
		key,
		fmt.Sprintf("%s:time", prefix), ts.Format(time.RFC3339),
		fmt.Sprintf("%s:count", prefix), strconv.Itoa(count),
		fmt.Sprintf("%s:error", prefix), errStr, // always set to override previous error
	}
	// in case of error count is 0, so we do not overwrite success, keeping last success timestamp
	if count > 0 {
		hmset = append(hmset, fmt.Sprintf("%s:success", prefix), ts.Format(time.RFC3339))
	}

	err = conn.Send("HMSET", hmset...)
	if err != nil {
		return core.WrapErr(err, "failed save status").With("jobID", jobID).With("code", code)
	}
	return nil
}

// SaveStatus implements CacheManager interface
func (cache *RedisCacheManager) SaveStatus(jobID, code string, err error, count int) error {
	return cache.saveStatusWithTime(jobID, code, err, count, time.Now().UTC())
}

// LoadLatestMeasurements implements CacheManager interface
func (cache *RedisCacheManager) LoadLatestMeasurements(from map[string]core.StringSet) (map[core.GaugeID]core.Measurement, error) {
	result := make(map[core.GaugeID]core.Measurement)
	var raws []string
	conn := cache.pool.Get()
	defer conn.Close()

	if len(from) == 0 {
		return result, nil
	}
	isGetAll, i := make([]bool, len(from)), 0

	for script, codes := range from {
		if len(codes) == 0 { // get all gauges for this script
			err := conn.Send("HGETALL", fmt.Sprintf("%s:%s", NSLatest, script))
			if err != nil {
				return result, core.WrapErr(err, "failed to hgetall last measurements")
			}
			isGetAll[i] = true
		} else {
			args, j := make([]interface{}, len(codes)+1), 1
			args[0] = fmt.Sprintf("%s:%s", NSLatest, script)
			for code := range codes {
				args[j] = code
				j++
			}
			err := conn.Send("HMGET", args...)
			if err != nil {
				return result, core.WrapErr(err, "failed to hmget last measurements")
			}
			isGetAll[i] = false
		}
		i++
	}
	reply, err := redis.Values(conn.Do(""))
	if err != nil {
		return result, core.WrapErr(err, "failed to read last measurements")
	}
	if len(reply) != len(from) {
		return result, errors.New("reply length doesn't match input length")
	}
	for i, r := range reply {
		if isGetAll[i] {
			stringMap, _ := redis.StringMap(r, nil)
			for _, v := range stringMap {
				raws = append(raws, v)
			}
		} else {
			strings, _ := redis.Strings(r, nil)
			for _, v := range strings {
				if v != "" {
					raws = append(raws, v)
				}
			}
		}
	}

	for _, jsonStr := range raws {
		var m core.Measurement
		if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
			return result, core.WrapErr(err, "failed to unmarshal last measurement from redis").With("value", jsonStr)
		}
		result[m.GaugeID] = m
	}
	return result, nil
}

// SaveLatestMeasurements implements CacheManager interface
func (cache *RedisCacheManager) SaveLatestMeasurements(ctx context.Context, in <-chan *core.Measurement) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)

		byGauge := map[core.GaugeID]*core.Measurement{}
	outer:
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case m, ok := <-in:
				if !ok {
					break outer
				}
				if !m.Flow.Valid() && !m.Level.Valid() {
					continue
				}
				if e, ok := byGauge[m.GaugeID]; ok {
					if e.Timestamp.Before(m.Timestamp.Time) {
						byGauge[m.GaugeID] = m
					}
				} else {
					byGauge[m.GaugeID] = m
				}
			}
		}

		var conn redis.Conn
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		default:
			conn = cache.pool.Get()
			defer conn.Close()
			var raw []byte

			for id, m := range byGauge {
				raw, _ = json.Marshal(m)
				err := conn.Send(
					"HSET",
					fmt.Sprintf("%s:%s", NSLatest, id.Script),
					id.Code,
					raw,
				)
				if err != nil {
					errCh <- err
					return
				}
			}
		}
		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		default:
			_, err := conn.Do("")
			errCh <- err
		}
	}()
	return errCh
}

// Start implements CacheManager interface
func (cache *RedisCacheManager) Start() error {
	cache.pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", cache.address)
		},
	}
	return nil
}

// Close implements CacheManager interface
func (cache *RedisCacheManager) Close() error {
	return cache.pool.Close()
}

// Start implements CacheManager interface
func (cache *EmbeddedCacheManager) Start() error {
	srv, err := miniredis.Run()
	if err != nil {
		return core.WrapErr(err, "failed to start embedded redis")
	}
	cache.srv = srv
	cache.address = srv.Addr()
	return cache.RedisCacheManager.Start()
}

// Close implements CacheManager interface
func (cache *EmbeddedCacheManager) Close() error {
	err := cache.RedisCacheManager.Close()
	if err != nil {
		return err
	}
	cache.srv.Close()
	return nil
}
