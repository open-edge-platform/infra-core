-- Drop index "operatingsystemresource_tenant_id" from table: "operating_system_resources"
DROP INDEX "operatingsystemresource_tenant_id";
-- Create index "operatingsystemresource_tenant_id" to table: "operating_system_resources"
CREATE UNIQUE INDEX "operatingsystemresource_tenant_id" ON "operating_system_resources" ("tenant_id");
