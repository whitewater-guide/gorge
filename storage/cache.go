package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/whitewater-guide/gorge/core"
)

// RedisCacheManager is cache manager that uses real redis
type RedisCacheManager struct {
	pool *redis.Pool
}

// EmbeddedCacheManager is cache manager that uses embedded redis https://github.com/alicebob/miniredis
type EmbeddedCacheManager struct {
	srv *miniredis.Miniredis
	RedisCacheManager
}

const (
	// NSStatus is redis namespace prefix for job/gauge statuses
	NSStatus = "status"
	// NSLatest is redis namespace prefix for latest measurements
	NSLatest = "latest"
)

func newCacheManager(address string) (*RedisCacheManager, error) {
	manager := &RedisCacheManager{
		pool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", address)
			},
		},
	}

	return manager, nil
}

// NewRedisCacheManager creates new redis cache manager
func NewRedisCacheManager(host, port string) (*RedisCacheManager, error) {
	return newCacheManager(fmt.Sprintf("%s:%s", host, port))
}

// NewEmbeddedCacheManager creates new miniredis cache manager
func NewEmbeddedCacheManager() (*EmbeddedCacheManager, error) {
	srv, err := miniredis.Run()
	if err != nil {
		return nil, core.WrapErr(err, "failed to start embedded redis")
	}
	manager, err := newCacheManager(srv.Addr())

	return &EmbeddedCacheManager{
		srv:               srv,
		RedisCacheManager: *manager,
	}, err
}

func (cache RedisCacheManager) loadStatuses(jobID string) (map[string]core.Status, error) {
	conn := cache.pool.Get()
	defer conn.Close()
	key := NSStatus // get jobids statuses
	if jobID != "" {
		key = key + ":" + jobID // get gauges statuses of this jobID
	}
	m, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		return nil, core.WrapErr(err, "failed to get statuses")
	}
	i := 0
	result := make(map[string]core.Status)
	for id, rawStatus := range m {
		var status core.Status
		if err := json.Unmarshal([]byte(rawStatus), &status); err != nil {
			return nil, core.WrapErr(err, "failed to unmarshal job status from redis").With("value", rawStatus)
		}
		result[id] = status
		i++
	}
	return result, nil
}

// LoadJobStatuses implements CacheManager interface
func (cache RedisCacheManager) LoadJobStatuses() (map[string]core.Status, error) {
	return cache.loadStatuses("")
}

// LoadGaugeStatuses implements CacheManager interface
func (cache RedisCacheManager) LoadGaugeStatuses(jobID string) (map[string]core.Status, error) {
	return cache.loadStatuses(jobID)
}

func (cache RedisCacheManager) saveStatusWithTime(jobID, code string, err error, count int, ts time.Time) error {
	conn := cache.pool.Get()
	defer conn.Close()
	success, errStr := true, ""
	if err != nil {
		success = false
		errStr = err.Error()
	}
	status := core.Status{
		Success:   success,
		Timestamp: core.HTime{Time: ts},
		Error:     errStr,
		Count:     count,
	}
	raw, err := json.Marshal(status)
	if err != nil {
		return core.WrapErr(err, "failed to marshal status")
	}
	if code != "" {
		err = conn.Send("HSET", fmt.Sprintf("%s:%s", NSStatus, jobID), code, raw)
		if err != nil {
			return core.WrapErr(err, "failed send gauge status")
		}
	}
	_, err = conn.Do("HSET", NSStatus, jobID, raw)
	if err != nil {
		return core.WrapErr(err, "failed to save status").With("jobID", jobID).With("code", code)
	}
	return nil
}

// SaveStatus implements CacheManager interface
func (cache RedisCacheManager) SaveStatus(jobID, code string, err error, count int) error {
	return cache.saveStatusWithTime(jobID, code, err, count, time.Now().UTC())
}

// LoadLatestMeasurements implements CacheManager interface
func (cache RedisCacheManager) LoadLatestMeasurements(from map[string]core.StringSet) (map[core.GaugeID]core.Measurement, error) {
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
func (cache RedisCacheManager) SaveLatestMeasurements(ctx context.Context, in <-chan *core.Measurement) <-chan error {
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

// Close implements CacheManager interface
func (cache RedisCacheManager) Close() {
	cache.pool.Close()
}

// Close implements CacheManager interface
func (cache EmbeddedCacheManager) Close() {
	cache.RedisCacheManager.Close()
	cache.srv.Close()
}
