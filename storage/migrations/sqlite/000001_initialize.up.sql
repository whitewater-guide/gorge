CREATE TABLE IF NOT EXISTS measurements
(
    timestamp TEXT NOT NULL,
    script TEXT NOT NULL,
    code TEXT NOT NULL,
    flow REAL,
    level REAL
);

CREATE INDEX measurements_script_code_index
    ON measurements (script, code);

CREATE UNIQUE INDEX measurements_idx
    ON measurements (script asc, code asc, timestamp desc);

CREATE INDEX measurements_timestamp_idx
    ON measurements (timestamp desc);

CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    description TEXT -- JSON
);
