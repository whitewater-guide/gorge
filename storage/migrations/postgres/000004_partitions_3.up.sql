BEGIN;

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

COMMIT;
