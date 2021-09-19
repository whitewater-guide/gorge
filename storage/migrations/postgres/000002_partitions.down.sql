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

CALL partman.undo_partition_proc('public.measurements', p_interval := 'daily'::text, p_batch := 500, p_target_table := 'public.new_measurements', p_keep_table := false);

DROP TABLE IF EXISTS measurements;

ALTER TABLE new_measurements RENAME TO measurements;

