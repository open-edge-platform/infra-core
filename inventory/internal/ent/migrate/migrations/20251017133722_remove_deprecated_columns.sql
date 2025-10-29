-- Check os_update_policy_resources for non-NULL values in install_packages or kernel_command.
-- Abort if any soon-to-be-dropped columns still contain data.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM os_update_policy_resources
    WHERE install_packages IS NOT NULL OR kernel_command IS NOT NULL
  ) THEN
    RAISE EXCEPTION 'Refusing to drop: os_update_policy_resources has non-NULL deprecated data';
  END IF;
END$$;

-- Check instance_resources for non-NULL values in update_status_detail, instance_resource_desired_os, or instance_resource_current_os
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM instance_resources
    WHERE update_status_detail IS NOT NULL
       OR instance_resource_desired_os IS NOT NULL
       OR instance_resource_current_os IS NOT NULL
  ) THEN
    RAISE EXCEPTION 'Refusing to drop: instance_resources has non-NULL deprecated data';
  END IF;
END$$;

-- Check operating_system_resources for non-NULL values in update_sources or kernel_command
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM operating_system_resources
    WHERE update_sources IS NOT NULL
       OR kernel_command IS NOT NULL
  ) THEN
    RAISE EXCEPTION 'Refusing to drop: operating_system_resources has non-NULL deprecated data';
  END IF;
END$$;

-- Drop the empty, deprecated columns
-- atlas:nolint DS103 - pre-migration “checks.sql” (the recommended fix) requires Atlas Pro
ALTER TABLE "instance_resources"
  DROP COLUMN "update_status_detail",
  DROP COLUMN "instance_resource_desired_os",
  DROP COLUMN "instance_resource_current_os";
-- atlas:nolint DS103 - pre-migration “checks.sql” (the recommended fix) requires Atlas Pro
ALTER TABLE "os_update_policy_resources"
  DROP COLUMN "install_packages",
  DROP COLUMN "kernel_command";
-- atlas:nolint DS103 - pre-migration “checks.sql” (the recommended fix) requires Atlas Pro
ALTER TABLE "operating_system_resources"
  DROP COLUMN "update_sources",
  DROP COLUMN "kernel_command";