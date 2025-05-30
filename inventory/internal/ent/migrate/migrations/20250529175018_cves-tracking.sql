-- Modify "instance_resources" table
ALTER TABLE "instance_resources" ADD COLUMN "existing_cves" character varying NULL;
-- Modify "operating_system_resources" table
ALTER TABLE "operating_system_resources" ADD COLUMN "existing_cves" character varying NULL, ADD COLUMN "existing_cves_url" character varying NULL, ADD COLUMN "fixed_cves" character varying NULL, ADD COLUMN "fixed_cves_url" character varying NULL;
