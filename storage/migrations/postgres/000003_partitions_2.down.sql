CALL partman.undo_partition_proc(
    'public.measurements',
    p_interval := 'daily'::text,
    p_batch := 500,
    p_target_table := 'public.new_measurements',
    p_keep_table := false
);