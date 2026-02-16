-- Create the unique index
-- Below comment suppresses Atlas data dependency warnings for this migration. This allows the index creation to proceed without checking for existing data conflicts
-- The orchestrator upgrade might fail if there are existing duplicate records that violate the new unique constraint, so it's important to ensure that the data duplication is clean before applying this migration. 
-- atlas:nolint data_depend
CREATE UNIQUE INDEX "operatingsystemresource_profile_name_image_id_tenant_id" ON "operating_system_resources" ("profile_name", "image_id", "tenant_id");
