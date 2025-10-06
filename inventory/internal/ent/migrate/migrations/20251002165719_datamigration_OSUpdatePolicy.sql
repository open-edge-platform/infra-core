-- Backfill new fields from legacy ones if new are NULL/empty.
UPDATE os_update_policy_resources
SET
  update_packages = CASE
    WHEN (update_packages IS NULL OR update_packages = '')
    THEN install_packages ELSE update_packages END,
  update_kernel_command = CASE
    WHEN (update_kernel_command IS NULL OR update_kernel_command = '')
    THEN kernel_command ELSE update_kernel_command END
WHERE
  (install_packages IS NOT NULL AND install_packages <> '')
  OR (kernel_command   IS NOT NULL AND kernel_command   <> '');
