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

-- Drop the deprecated columns now that they have been verified to be empty.

-- atlas:nolint DS103 - pre-migration “checks.sql” (the recommended fix) requires Atlas Pro
ALTER TABLE "instance_resources"
  DROP COLUMN "update_status_detail",
  DROP COLUMN "instance_resource_desired_os",
  DROP COLUMN "instance_resource_current_os";

-- atlas:nolint DS103 - pre-migration “checks.sql” (the recommended fix) requires Atlas Pro
ALTER TABLE "os_update_policy_resources"
  DROP COLUMN "install_packages",
  DROP COLUMN "kernel_command";