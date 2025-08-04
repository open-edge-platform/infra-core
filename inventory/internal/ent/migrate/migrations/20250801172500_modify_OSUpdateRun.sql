-- Modify columns from timestamp → bigint (epoch in ms)
ALTER TABLE os_update_run_resources
    ALTER COLUMN status_timestamp TYPE bigint USING (extract(epoch FROM status_timestamp) * 1000)::bigint,
    ALTER COLUMN start_time TYPE bigint USING (extract(epoch FROM start_time) * 1000)::bigint,
    ALTER COLUMN end_time TYPE bigint USING (extract(epoch FROM end_time) * 1000)::bigint;
