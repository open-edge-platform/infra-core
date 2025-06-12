-- First ensure that we won't have any duplicate names for the same tenant.
UPDATE "operating_system_resources"
SET "name" = "name" || '-' || substr(md5(random()::text), 1, 5)
WHERE ("name", "tenant_id") IN (
    SELECT "name", "tenant_id"
    FROM "operating_system_resources"
    GROUP BY "name", "tenant_id"
    HAVING COUNT(*) > 1
);
-- atlas:nolint MF101 the above data migration ensure that there won't be any duplicated.
-- Create index "operatingsystemresource_name_tenant_id" to table: "operating_system_resources"
CREATE UNIQUE INDEX "operatingsystemresource_name_tenant_id" ON "operating_system_resources" ("name", "tenant_id");
