-- Migrate data
CALL partman.partition_data_proc(
    'public.measurements',
    p_batch := 100,
    p_source_table := 'public.old_measurements'
);
