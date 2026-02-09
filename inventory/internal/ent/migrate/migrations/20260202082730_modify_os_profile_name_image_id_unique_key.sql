-- Handle duplicate OS resources before creating unique index
-- Step 1: For each set of duplicates, migrate all foreign key references to the OS resource with the smallest ID
-- This ensures no foreign key constraint violations when we delete duplicates

-- Update instance_resources to point to the kept OS (smallest id)
UPDATE "instance_resources" ir
SET "instance_resource_os" = kept.min_id
FROM (
    SELECT 
        b.id as duplicate_id,
        MIN(a.id) as min_id
    FROM "operating_system_resources" a
    JOIN "operating_system_resources" b 
        ON a.profile_name = b.profile_name 
        AND a.image_id = b.image_id 
        AND a.tenant_id = b.tenant_id
        AND a.id < b.id
    GROUP BY b.id
) AS kept
WHERE ir."instance_resource_os" = kept.duplicate_id;

-- Update os_update_policy_resources to point to the kept OS (smallest id)
UPDATE "os_update_policy_resources" osp
SET "os_update_policy_resource_target_os" = kept.min_id
FROM (
    SELECT 
        b.id as duplicate_id,
        MIN(a.id) as min_id
    FROM "operating_system_resources" a
    JOIN "operating_system_resources" b 
        ON a.profile_name = b.profile_name 
        AND a.image_id = b.image_id 
        AND a.tenant_id = b.tenant_id
        AND a.id < b.id
    GROUP BY b.id
) AS kept
WHERE osp."os_update_policy_resource_target_os" = kept.duplicate_id;

-- Step 2: Now safe to delete duplicate OS resources (keeping the one with smallest id)
DELETE FROM "operating_system_resources" a 
USING "operating_system_resources" b
WHERE a.id > b.id 
AND a.profile_name = b.profile_name 
AND a.image_id = b.image_id 
AND a.tenant_id = b.tenant_id;

-- Step 3: Create the unique index
-- atlas:nolint data_depend
CREATE UNIQUE INDEX "operatingsystemresource_profile_name_image_id_tenant_id" ON "operating_system_resources" ("profile_name", "image_id", "tenant_id");
