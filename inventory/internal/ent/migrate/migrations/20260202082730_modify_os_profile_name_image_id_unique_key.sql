-- Create index "operatingsystemresource_profile_name_image_id_tenant_id" to table: "operating_system_resources"
CREATE UNIQUE INDEX "operatingsystemresource_profile_name_image_id_tenant_id" ON "operating_system_resources" ("profile_name", "image_id", "tenant_id");
