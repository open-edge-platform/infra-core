-- Modify "os_update_policy_resources" table
ALTER TABLE "os_update_policy_resources" ADD COLUMN "update_packages" character varying NULL, ADD COLUMN "update_kernel_command" character varying NULL;
