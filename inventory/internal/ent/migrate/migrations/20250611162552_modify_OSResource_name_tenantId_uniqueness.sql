-- Create index "operatingsystemresource_name_tenant_id" to table: "operating_system_resources"
CREATE UNIQUE INDEX "operatingsystemresource_name_tenant_id" ON "operating_system_resources" ("name", "tenant_id");
