-- Migrate data
CALL partman.partition_data_proc(
    'public.measurements',
    p_interval := '1 day',
    p_batch := 500,
    p_source_table := 'public.old_measurements'
);