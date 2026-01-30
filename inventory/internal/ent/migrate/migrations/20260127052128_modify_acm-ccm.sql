-- Modify "host_resources" table
ALTER TABLE "host_resources" ADD COLUMN "amt_control_mode" character varying NULL, ADD COLUMN "amt_dns_suffix" character varying NULL;
