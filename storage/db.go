package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	// Import postgres
	_ "github.com/lib/pq"
	"github.com/whitewater-guide/gorge/core"
)

// DbManager implements DatabaseManager using sql database
type DbManager struct {
	db *sqlx.DB
	// nearest day order by clasue
	nearestDayClause string
	// defaultStart is sql expression for starting period of measurements slice
	defaultStart string
	// saveChunkSize indicates how many measurements will be written in on query
	// when set to 0, no limit will be enforced, which might lead to hitting max variables limits or other limits
	saveChunkSize int
}

const saveMeasurementsQuery = "INSERT INTO measurements (timestamp, script, code, flow, level) VALUES (:timestamp, :script, :code, :flow, :level) ON CONFLICT DO NOTHING"

// obtainConnection waits for postgres to start, because containers start in random order
func obtainConnection(driver, address string, timeout, retries int64) (*sqlx.DB, error) {
	db, err := sqlx.Connect(driver, address)

	if err != nil {
		if retries > 0 {
			time.Sleep(time.Duration(timeout) * time.Second)
			return obtainConnection(driver, address, timeout, retries-1)
		}
		return nil, fmt.Errorf("couldn't wait for postgres anymore, %w", err)
	}

	return db, nil
}

func (mgr *DbManager) saveMeasurementsChunk(chunk []*core.Measurement) (int, error) {
	result, err := mgr.db.NamedExec(saveMeasurementsQuery, chunk)
	if err != nil {
		return 0, core.WrapErr(err, "failed to save measurements").With("count", len(chunk))
	}
	if rowsAffected, err := result.RowsAffected(); err == nil {
		return int(rowsAffected), nil
	}
	return len(chunk), nil
}

// SaveMeasurements implements DatabaseManager interface
func (mgr *DbManager) SaveMeasurements(ctx context.Context, in <-chan *core.Measurement) (<-chan int, <-chan error) {
	savedCh := make(chan int, 1)
	errCh := make(chan error, 1)
	go func() {
		defer close(savedCh)
		defer close(errCh)
		var chunk []*core.Measurement
		total, count := 0, 0
		for m := range core.Cancelable(ctx, in) {
			if m.Flow.Float64Value() == 0.0 && m.Level.Float64Value() == 0.0 || !m.Flow.Valid() && !m.Level.Valid() {
				continue
			}
			chunk = append(chunk, m)
			count++
			if count == mgr.saveChunkSize && mgr.saveChunkSize != 0 {
				saved, err := mgr.saveMeasurementsChunk(chunk)
				if err != nil {
					errCh <- core.WrapErr(err, "failed to save measurements")
					return
				}
				total, count = total+saved, 0
				chunk = nil
			}
		}

		select {
		case <-ctx.Done():
			errCh <- ctx.Err()
			return
		default:
			if count > 0 {
				saved, err := mgr.saveMeasurementsChunk(chunk)
				if err != nil {
					errCh <- core.WrapErr(err, "failed to save measurements")
				}
				total += saved
			}
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
			case savedCh <- total:
			}
		}

	}()
	return savedCh, errCh
}

// GetMeasurements implements DatabaseManager interface
func (mgr *DbManager) GetMeasurements(query MeasurementsQuery) ([]core.Measurement, error) {
	q, args := mgr.getMeasurementsWhereClause(query)
	q = "SELECT * FROM measurements " + q
	rows, err := mgr.db.Queryx(q, args...)
	if err != nil {
		return nil, core.WrapErr(err, "failed to query measurements")
	}
	result := make([]core.Measurement, 0)
	for rows.Next() {
		var m core.Measurement
		err := rows.StructScan(&m)
		if err != nil {
			rows.Close()
			return nil, core.WrapErr(err, "failed to get next measurements row")
		}
		result = append(result, m)
	}

	return result, nil
}

// GetNearestMeasurement implements DatabaseManager interface
func (mgr *DbManager) GetNearestMeasurement(script, code string, to time.Time, tolerance time.Duration) (*core.Measurement, error) {
	q := "SELECT * FROM measurements WHERE script = $1 AND code = $2 ORDER BY " + fmt.Sprintf(mgr.nearestDayClause, "$3") + " LIMIT 1"
	var m core.Measurement
	err := mgr.db.QueryRowx(q, script, code, to).StructScan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, core.WrapErr(err, "failed to query nearest measurement")
	}
	if tolerance != 0 && (m.Timestamp.After(to.Add(tolerance)) || m.Timestamp.Before(to.Add(-tolerance))) {
		return nil, nil
	}
	m.Timestamp = core.HTime{Time: m.Timestamp.Time.UTC()}
	return &m, nil
}

// ListJobs implements DatabaseManager interface
func (mgr *DbManager) ListJobs() ([]core.JobDescription, error) {
	rows, err := mgr.db.Query("SELECT id, description FROM jobs")
	if err != nil {
		return nil, err
	}

	result := make([]core.JobDescription, 0)
	for rows.Next() {
		var id string
		var description string
		var job core.JobDescription
		err := rows.Scan(&id, &description)
		if err != nil {
			rows.Close()
			return nil, err
		}
		err = json.Unmarshal([]byte(description), &job)
		if err != nil {
			rows.Close()
			return nil, err
		}
		result = append(result, job)
	}

	return result, nil
}

// GetJob implements DatabaseManager interface
func (mgr *DbManager) GetJob(id string) (*core.JobDescription, error) {
	var result struct {
		ID          string
		Description *core.JobDescription
	}
	err := mgr.db.Get(&result, "SELECT * FROM jobs WHERE id = $1", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return result.Description, err
}

// AddJob implements DatabaseManager interface
func (mgr *DbManager) AddJob(job core.JobDescription, onSave func(job core.JobDescription) error) error {
	descr, err := json.Marshal(job)
	if err != nil {
		return core.WrapErr(err, "failed to marshal job description")
	}
	tx, err := mgr.db.Begin()
	if err != nil {
		return core.WrapErr(err, "failed to begin add job transaction")
	}

	_, err = tx.Exec("INSERT INTO jobs (id, description) VALUES ($1, $2)", job.ID, descr)
	if err != nil {
		tx.Rollback() //nolint:errcheck
		return core.WrapErr(err, "failed to insert job").With("description", string(descr)).With("id", job.ID)
	}

	saveErr := onSave(job)
	if saveErr != nil {
		tx.Rollback() //nolint:errcheck
		return saveErr
	}

	err = tx.Commit()
	if err != nil {
		return core.WrapErr(err, "failed to commit add job transaction")
	}

	return nil
}

// DeleteJob implements DatabaseManager interface
func (mgr *DbManager) DeleteJob(id string, onDelete func(id string) error) error {
	tx, err := mgr.db.Begin()
	if err != nil {
		return core.WrapErr(err, "failed to begin delete job transaction")
	}
	res, err := tx.Exec("DELETE FROM jobs WHERE id = $1", id)
	if err != nil {
		tx.Rollback() //nolint:errcheck
		return core.WrapErr(err, "failed to delete job").With("jobId", id)
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		tx.Rollback() //nolint:errcheck
		return core.WrapErr(err, "failed to count deleted rows").With("jobId", id)
	}

	deleteErr := onDelete(id)
	if deleteErr != nil {
		tx.Rollback() //nolint:errcheck
		return deleteErr
	}

	err = tx.Commit()
	if err != nil {
		return core.WrapErr(err, "failed to commit delete job transaction")
	}
	if cnt == 0 {
		return core.WrapErr(err, "job not found in database")
	}

	return nil
}

// Close implements DatabaseManager interface
func (mgr *DbManager) Close() error {
	return mgr.db.Close()
}

func (mgr *DbManager) getMeasurementsWhereClause(query MeasurementsQuery) (string, []interface{}) {
	var args []interface{} = []interface{}{query.Script}
	fromP := "$2"
	if query.From == nil {
		fromP = mgr.defaultStart
	} else {
		args = append(args, query.From)
	}
	where := fmt.Sprintf("WHERE script = $1 AND timestamp >= %s", fromP)
	if query.To != nil {
		args = append(args, query.To)
		where = fmt.Sprintf("%s AND timestamp <= $%d", where, len(args))
	}
	order := "ORDER BY script ASC, timestamp DESC"
	if query.Code != "" {
		args = append(args, query.Code)
		where = fmt.Sprintf("%s AND code = $%d", where, len(args))
		order = "ORDER BY script ASC, code ASC, timestamp DESC"
	}
	return where + " " + order, args
}
