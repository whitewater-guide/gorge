BEGIN;

DROP TABLE IF EXISTS measurements;

ALTER TABLE new_measurements RENAME TO measurements;

COMMIT;
