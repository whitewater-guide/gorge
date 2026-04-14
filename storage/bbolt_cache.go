package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/whitewater-guide/gorge/core"
	bbolt "go.etcd.io/bbolt"
)

// BboltCacheManager is a cache manager that persists data to a bbolt database file.
type BboltCacheManager struct {
	db   *bbolt.DB
	path string
	log  *logrus.Entry
}

func formatBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// Start implements CacheManager interface.
func (cache *BboltCacheManager) Start() error {
	db, err := bbolt.Open(cache.path, 0600, nil)
	if err != nil {
		return core.WrapErr(err, "failed to open bbolt cache")
	}
	cache.db = db

	err = db.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(NSStatus)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(NSLatest)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return core.WrapErr(err, "failed to initialise bbolt buckets")
	}

	if cache.log != nil {
		if info, statErr := os.Stat(cache.path); statErr == nil {
			cache.log.Infof("bbolt cache opened: %s (%s)", cache.path, formatBytes(info.Size()))
		}
	}
	return nil
}

// Close implements CacheManager interface.
func (cache *BboltCacheManager) Close() error {
	return cache.db.Close()
}

func (cache *BboltCacheManager) saveStatusAt(jobID, code string, err error, count int, ts time.Time) error {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	var subBucketName, prefix string
	if code == "" {
		subBucketName = "jobs"
		prefix = jobID
	} else {
		subBucketName = jobID
		prefix = code
	}

	return cache.db.Update(func(tx *bbolt.Tx) error {
		statusBucket := tx.Bucket([]byte(NSStatus))
		if statusBucket == nil {
			return errors.New("status bucket not found")
		}
		sub, e := statusBucket.CreateBucketIfNotExists([]byte(subBucketName))
		if e != nil {
			return e
		}
		if e := sub.Put([]byte(prefix+":time"), []byte(ts.Format(time.RFC3339))); e != nil {
			return e
		}
		if e := sub.Put([]byte(prefix+":count"), []byte(strconv.Itoa(count))); e != nil {
			return e
		}
		if e := sub.Put([]byte(prefix+":error"), []byte(errStr)); e != nil {
			return e
		}
		if count > 0 {
			if e := sub.Put([]byte(prefix+":success"), []byte(ts.Format(time.RFC3339))); e != nil {
				return e
			}
		}
		return nil
	})
}

// SaveStatus implements CacheManager interface.
func (cache *BboltCacheManager) SaveStatus(jobID, code string, err error, count int) error {
	return cache.saveStatusAt(jobID, code, err, count, time.Now().UTC())
}

func (cache *BboltCacheManager) loadStatuses(subBucketName string) (map[string]core.Status, error) {
	m := make(map[string]string)
	err := cache.db.View(func(tx *bbolt.Tx) error {
		statusBucket := tx.Bucket([]byte(NSStatus))
		if statusBucket == nil {
			return nil
		}
		sub := statusBucket.Bucket([]byte(subBucketName))
		if sub == nil {
			return nil
		}
		c := sub.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			m[string(k)] = string(v)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return parseStatusFields(m)
}

// LoadJobStatuses implements CacheManager interface.
func (cache *BboltCacheManager) LoadJobStatuses() (map[string]core.Status, error) {
	return cache.loadStatuses("jobs")
}

// LoadGaugeStatuses implements CacheManager interface.
func (cache *BboltCacheManager) LoadGaugeStatuses(jobID string) (map[string]core.Status, error) {
	return cache.loadStatuses(jobID)
}

// LoadLatestMeasurements implements CacheManager interface.
func (cache *BboltCacheManager) LoadLatestMeasurements(from map[string]core.StringSet) (map[core.GaugeID]core.Measurement, error) {
	result := make(map[core.GaugeID]core.Measurement)
	if len(from) == 0 {
		return result, nil
	}

	err := cache.db.View(func(tx *bbolt.Tx) error {
		latestBucket := tx.Bucket([]byte(NSLatest))
		if latestBucket == nil {
			return nil
		}
		for script, codes := range from {
			scriptBucket := latestBucket.Bucket([]byte(script))
			if scriptBucket == nil {
				continue
			}
			if len(codes) == 0 {
				c := scriptBucket.Cursor()
				for k, v := c.First(); k != nil; k, v = c.Next() {
					var m core.Measurement
					if err := json.Unmarshal(v, &m); err != nil {
						return core.WrapErr(err, "failed to unmarshal measurement from bbolt").With("key", string(k))
					}
					result[m.GaugeID] = m
				}
			} else {
				for code := range codes {
					v := scriptBucket.Get([]byte(code))
					if v == nil {
						continue
					}
					var m core.Measurement
					if err := json.Unmarshal(v, &m); err != nil {
						return core.WrapErr(err, "failed to unmarshal measurement from bbolt").With("code", code)
					}
					result[m.GaugeID] = m
				}
			}
		}
		return nil
	})
	return result, err
}

// SaveLatestMeasurements implements CacheManager interface.
func (cache *BboltCacheManager) SaveLatestMeasurements(ctx context.Context, in <-chan *core.Measurement) <-chan error {
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

		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		default:
		}

		errCh <- cache.db.Update(func(tx *bbolt.Tx) error {
			latestBucket := tx.Bucket([]byte(NSLatest))
			if latestBucket == nil {
				return errors.New("latest bucket not found")
			}
			for id, m := range byGauge {
				scriptBucket, err := latestBucket.CreateBucketIfNotExists([]byte(id.Script))
				if err != nil {
					return err
				}
				raw, _ := json.Marshal(m)
				if err := scriptBucket.Put([]byte(id.Code), raw); err != nil {
					return err
				}
			}
			return nil
		})
	}()
	return errCh
}
