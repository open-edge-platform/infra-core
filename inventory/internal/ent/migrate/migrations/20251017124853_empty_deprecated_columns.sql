-- Clear the legacy columns that were already migrated in 20251002165719_datamigration_OSUpdatePolicy.sql
-- The data has already been copied to update_packages and update_kernel_command
UPDATE os_update_policy_resources 
SET install_packages = NULL, 
    kernel_command = NULL
WHERE install_packages IS NOT NULL 
   OR kernel_command IS NOT NULL;

-- Clear instance_resources deprecated fields, their content is not used anymore  
UPDATE instance_resources 
SET update_status_detail = NULL, 
    instance_resource_desired_os = NULL, 
    instance_resource_current_os = NULL
WHERE update_status_detail IS NOT NULL 
   OR instance_resource_desired_os IS NOT NULL 
   OR instance_resource_current_os IS NOT NULL;
