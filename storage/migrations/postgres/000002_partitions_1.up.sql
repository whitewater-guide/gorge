BEGIN;

-- Prerequisites: setup partman

CREATE SCHEMA IF NOT EXISTS partman;
CREATE EXTENSION IF NOT EXISTS pg_partman SCHEMA partman;

CREATE ROLE partman WITH LOGIN;
GRANT ALL ON SCHEMA partman TO partman;
GRANT ALL ON ALL TABLES IN SCHEMA partman TO partman;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA partman TO partman;
GRANT EXECUTE ON ALL PROCEDURES IN SCHEMA partman TO partman;
GRANT ALL ON SCHEMA public TO partman;

-- Schema where archived tables will be placed
CREATE SCHEMA IF NOT EXISTS archive;
GRANT ALL ON SCHEMA archive TO partman;

-- End of partman setup
-- Beginning of migration

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

COMMIT;
