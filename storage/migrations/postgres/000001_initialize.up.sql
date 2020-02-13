BEGIN;

CREATE TABLE measurements
(
    timestamp timestamp with time zone not null,
    script varchar(255) not null,
    code varchar(255) not null,
    flow real,
    level real
);

CREATE INDEX measurements_script_code_index
    ON measurements (script, code);

CREATE UNIQUE INDEX measurements_idx
    ON measurements (script asc, code asc, timestamp desc);

CREATE INDEX measurements_timestamp_idx
    ON measurements (timestamp desc);

SELECT create_hypertable('measurements', 'timestamp');

CREATE TABLE IF NOT EXISTS jobs 
(
    id TEXT PRIMARY KEY,
    description JSON
);

COMMIT;
