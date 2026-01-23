-- Modify "host_resources" table
ALTER TABLE "host_resources" ADD COLUMN "amt_control_mode" character varying NULL, ADD COLUMN "amt_dns_suffix" character varying NULL;
-- Drop "os_update_policies" table
DROP TABLE "os_update_policies";
