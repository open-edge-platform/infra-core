-- ***** First migrate Mutable OS *****
-- Create OsUpdatePolicyResources for every Mutable OperatingSystemResource with update_sources, kernel_commands or installed_packages fields set,
-- needs to be migrated to the new schema by creating OSUpdatePolicy, filling the install_packages, update_sources and kernel_command fields
-- with the one coming from the OSprofile and linking it to the related Instances.

-- First, insert new OS update policies for mutable operating system resources
INSERT INTO os_update_policy_resources (
    resource_id,
    name,
    description,
    install_packages,
    update_sources,
    kernel_command,
    update_policy,
-- temporary using os_update_policy_resource_target_os for storing target OS during migration, but this is removed right after.
    os_update_policy_resource_target_os,
    tenant_id,
    created_at,
    updated_at
)
SELECT
    'osupdatepolicy-' || split_part(os.resource_id, '-', 2) AS resource_id,
    '"' || os.name || '" Update Policy' AS name,
    'Migrated from OS "' || os.resource_id || '"' AS description,
    os.installed_packages,
    os.update_sources,
    os.kernel_command,
    'UPDATE_POLICY_TARGET' AS update_policy,
    os.id,
    os.tenant_id,
    NOW() AS created_at,
    NOW() AS updated_at
FROM
    operating_system_resources os
WHERE
    os.os_type = 'OS_TYPE_MUTABLE' AND (
    (os.installed_packages IS NOT NULL AND os.installed_packages != '') OR
    (os.update_sources IS NOT NULL AND os.update_sources != '') OR
    (os.kernel_command IS NOT NULL AND os.kernel_command != ''));

-- Then, update instance_resources to link to the new OS update policies, use the os_update_policy_resource_target_os as
-- a way to link the OS update policy to the instance via the Instance's current OS.
UPDATE instance_resources AS ir
SET instance_resource_os_update_policy = osp.id
FROM os_update_policy_resources osp
    JOIN operating_system_resources os ON osp.os_update_policy_resource_target_os = os.id
WHERE
    ir.instance_resource_current_os = os.id AND
    (os.os_type = 'OS_TYPE_MUTABLE' AND (
    (os.installed_packages IS NOT NULL AND os.installed_packages != '') OR
    (os.update_sources IS NOT NULL AND os.update_sources != '') OR
    (os.kernel_command IS NOT NULL AND os.kernel_command != '')));

-- Reset os_update_policy_resource_target_os, it's needed only during migration.
UPDATE os_update_policy_resources
SET os_update_policy_resource_target_os = NULL
WHERE os_update_policy_resource_target_os IS NOT NULL;
-- ***** End of Mutable OS migration *****

-- ***** Now migrate Immutable OS *****
-- Given the changes to the instance, and deprecation of current/desired OS, fill the OS edge in the instance with the
-- reference to the current_os.
UPDATE instance_resources AS ir
SET instance_resource_os = ir.instance_resource_current_os, updated_at = NOW()
WHERE ir.instance_resource_os IS NULL AND ir.instance_resource_current_os IS NOT NULL;


-- Create OSUpdatePolicyResources for every Instance with desiredOS different from currentOS
INSERT INTO os_update_policy_resources (
    resource_id,
    name,
    description,
    update_policy,
    tenant_id,
    created_at,
    updated_at,
    os_update_policy_resource_target_os
)
SELECT
    'osupdatepolicy-' || split_part(i.resource_id, '-', 2) AS resource_id,
    '"' || i.resource_id || '" OS Update Policy' AS name,
    'Auto-migrated policy for instance "' || i.resource_id || '" (desired OS differs from current OS)' AS description,
    'UPDATE_POLICY_TARGET' AS update_policy,
    i.tenant_id,
    NOW() AS created_at,
    NOW() AS updated_at,
    i.instance_resource_desired_os
FROM
    instance_resources i
    JOIN operating_system_resources osd ON i.instance_resource_desired_os = osd.id
    JOIN operating_system_resources osc ON i.instance_resource_current_os = osc.id
WHERE
    osd.os_type = 'OS_TYPE_IMMUTABLE' AND
    i.instance_resource_desired_os IS NOT NULL
    AND i.instance_resource_current_os IS NOT NULL
    AND i.instance_resource_desired_os != i.instance_resource_current_os;

-- Link the new OSUpdatePolicyResource to the instance
UPDATE instance_resources AS ir
SET instance_resource_os_update_policy = osp.id, updated_at = NOW()
FROM os_update_policy_resources osp
WHERE
    osp.resource_id = 'osupdatepolicy-' || split_part(ir.resource_id, '-', 2)
    AND ir.instance_resource_desired_os IS NOT NULL
    AND ir.instance_resource_current_os IS NOT NULL
    AND ir.instance_resource_desired_os != ir.instance_resource_current_os;

-- ***** End of Immutable OS migration *****

-- Finally, reset the fields in operating_system_resources that were migrated to os_update_policy_resources
UPDATE operating_system_resources
SET installed_packages = '', update_sources = '', kernel_command = '', updated_at = NOW()
WHERE
    os_type = 'OS_TYPE_MUTABLE' AND (
    (installed_packages IS NOT NULL AND installed_packages != '') OR
    (update_sources IS NOT NULL AND update_sources != '') OR
    (kernel_command IS NOT NULL AND kernel_command != ''));