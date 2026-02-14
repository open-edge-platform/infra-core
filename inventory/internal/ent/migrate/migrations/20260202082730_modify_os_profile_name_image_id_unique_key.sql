-- Create the unique index
-- atlas:nolint data_depend
CREATE UNIQUE INDEX "operatingsystemresource_profile_name_image_id_tenant_id" ON "operating_system_resources" ("profile_name", "image_id", "tenant_id");
