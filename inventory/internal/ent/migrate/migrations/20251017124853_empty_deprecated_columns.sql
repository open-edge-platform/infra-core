-- Migrate data from deprecated columns to new columns and enforce NOT NULL constraints

-- Copy data from legacy fields to the new os field (if they still exist in DB)
UPDATE instance_resources 
SET instance_resource_os = COALESCE(
    instance_resource_current_os, -- Keep existing value in current_os if set
    instance_resource_desired_os  -- Otherwise use desired_os 
)
WHERE instance_resource_os IS NULL
  AND (instance_resource_current_os IS NOT NULL 
       OR instance_resource_desired_os IS NOT NULL);

-- Validate that all instances now have an OS assigned
-- This runtime check ensures that legacy data does not violate the new NOT NULL constraint during deployments
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM instance_resources 
        WHERE instance_resource_os IS NULL
    ) THEN
        RAISE EXCEPTION 'Cannot add NOT NULL constraint: % instances exist without OS assigned', 
                       (SELECT COUNT(*) FROM instance_resources WHERE instance_resource_os IS NULL);
    END IF;
END $$;

-- Add the foreign key constraint (referential integrity + delete protection)
ALTER TABLE "instance_resources"
    DROP CONSTRAINT IF EXISTS "instance_resources_operating_system_resources_os",
    ADD CONSTRAINT "instance_resources_operating_system_resources_os" 
        FOREIGN KEY ("instance_resource_os") REFERENCES "operating_system_resources" ("id") 
        ON UPDATE NO ACTION ON DELETE NO ACTION;

-- Add the NOT NULL constraint (make OS required)
ALTER TABLE "instance_resources" 
ALTER COLUMN "instance_resource_os" SET NOT NULL;

-- Step 4: Remove NOT NULL constraint from desired_os (make it optional so we can clear it)
ALTER TABLE instance_resources 
ALTER COLUMN instance_resource_desired_os DROP NOT NULL;

-- Clear deprecated fields in instance_resources. Their content is not used anymore.
UPDATE instance_resources 
SET update_status_detail = NULL, 
    instance_resource_desired_os = NULL, 
    instance_resource_current_os = NULL
WHERE update_status_detail IS NOT NULL 
   OR instance_resource_desired_os IS NOT NULL 
   OR instance_resource_current_os IS NOT NULL;

-- Clear operating_system_resources deprecated fields, their content is not used anymore.
UPDATE operating_system_resources 
SET update_sources = NULL, 
    kernel_command = NULL
WHERE update_sources IS NOT NULL
   OR kernel_command IS NOT NULL;

-- Clear the os_update_policy_resources legacy columns that were already migrated in the previous OS update policy data migration.
-- The data has already been copied to update_packages and update_kernel_command.
UPDATE os_update_policy_resources 
SET install_packages = NULL, 
    kernel_command = NULL
WHERE install_packages IS NOT NULL 
   OR kernel_command IS NOT NULL;