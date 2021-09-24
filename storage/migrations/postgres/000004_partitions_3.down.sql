BEGIN;

CREATE TABLE new_measurements
(
    timestamp timestamp with time zone not null,
    script varchar(255) not null,
    code varchar(255) not null,
    flow real,
    level real
);

CREATE INDEX measurements_script_code_index
    ON new_measurements (script, code);

CREATE UNIQUE INDEX measurements_idx
    ON new_measurements (script asc, code asc, timestamp desc);

CREATE INDEX measurements_timestamp_idx
    ON new_measurements (timestamp desc);

COMMIT;

