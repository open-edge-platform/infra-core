-- Modify "os_update_run_resources" table
ALTER TABLE "os_update_run_resources" ALTER COLUMN "name" DROP NOT NULL, ALTER COLUMN "status" DROP NOT NULL, ALTER COLUMN "status_details" DROP NOT NULL, ALTER COLUMN "end_time" DROP NOT NULL;
