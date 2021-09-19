-- Prerequisites: setup partman and cron

CREATE SCHEMA IF NOT EXISTS partman;
CREATE EXTENSION IF NOT EXISTS pg_partman SCHEMA partman;


-- Schema where archived tables will be placed
CREATE SCHEMA IF NOT EXISTS archive;

-- First, the original table should be renamed so the partitioned table can be made with the original table's name.
ALTER TABLE measurements RENAME to old_measurements;

-- Recreate original table, but with partitions
CREATE TABLE measurements
(
    timestamp timestamp with time zone not null,
    script varchar(255) not null,
    code varchar(255) not null,
    flow real,
    level real
) PARTITION BY RANGE (timestamp);

CREATE INDEX msmnts_script_code_index
    ON measurements (script, code);

CREATE UNIQUE INDEX msmnts_idx
    ON measurements (script asc, code asc, timestamp desc);

CREATE INDEX msmnts_timestamp_idx
    ON measurements (timestamp desc);

-- Make partman handle this table
SELECT partman.create_parent('public.measurements', 'timestamp', 'native', 'monthly');

-- Migrate data
CALL partman.partition_data_proc('public.measurements', p_interval := '1 day', p_batch := 500, p_source_table := 'public.old_measurements');

-- Delete old table
DROP TABLE IF EXISTS old_measurements;

-- Configure partman maintetance
-- See https://github.com/pgpartman/pg_partman/blob/master/doc/pg_partman.md#tables
UPDATE partman.part_config 
SET infinite_time_partitions = true,
    retention = '13 months', 
    retention_schema = 'archive',
    retention_keep_table = true 
WHERE parent_table = 'public.measurements';
