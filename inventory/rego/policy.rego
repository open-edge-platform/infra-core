# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package abac

import rego.v1

default deny := false
default isException := false
default abac := false
default resourceRule := false
default allowRule := false
default deleteRule := false


# deny if a host resource is created via northbound API without UUID and SN
deny if {
	input.Method == "CREATE"
	input.resource.host
	not input.resource.host.uuid
	not input.resource.host.serialNumber
	input.ClientKind == "CLIENT_KIND_API"
}

# deny if a host resource is created via southbound API without UUID and SN
deny if {
	input.Method == "CREATE"
	input.resource.host
	not input.resource.host.uuid
	not input.resource.host.serialNumber
	input.ClientKind == "CLIENT_KIND_RESOURCE_MANAGER"
}

# deny if a host resource is updated via northbound API by UUID
deny if {
	input.Method == "UPDATE"
	input.resource.host
	input.resource.host.uuid
	input.ClientKind == "CLIENT_KIND_API"
}

# deny if a host resource is updated via northbound API by SN
deny if {
	input.Method == "UPDATE"
	input.resource.host
	input.resource.host.serialNumber
	input.ClientKind == "CLIENT_KIND_API"
}

# deny if a tenant resource is created/updated by other than TC client
deny if {
	input.resource.tenant
	not input.ClientKind == "CLIENT_KIND_TENANT_CONTROLLER"
}

# deny updates to the CurrentState via northbound API
deny if {
	input.CurrentState
	input.ClientKind == "CLIENT_KIND_API"
}

# deny updates to the DesiredState via southbound API
deny if {
	input.DesiredState
	input.ClientKind == "CLIENT_KIND_RESOURCE_MANAGER"
}

# deny updates to the currentPowerState via northbound API
deny if {
	input.resource.host
	input.resource.host.currentPowerState
	input.ClientKind == "CLIENT_KIND_API"
}

# deny updates to the desiredPowerState via southbound API
deny if {
	input.resource.host
	input.resource.host.desiredPowerState
	input.ClientKind == "CLIENT_KIND_RESOURCE_MANAGER"
}

# deny if TC client tries to create a resource that is not a tenant, provider or telemetryGroup
deny if {
    not input.resource.tenant
    not input.resource.provider
    not input.resource.telemetryGroup
    input.ClientKind == "CLIENT_KIND_TENANT_CONTROLLER"
}

# Exception 1
# Instance specific rules for supporting ZTP with default OS
# This rule allows southbound API to CREATE a new Instance resource with desiredState set to RUNNING
# and of kind METAL. All other options for the mentioned fields are not supported
isException if {
	input.Method == "CREATE"
	input.DesiredState
	input.resource.instance
	input.resource.instance.kind == "INSTANCE_KIND_METAL"
	input.resource.instance.desiredState == "INSTANCE_STATE_RUNNING"
	input.ClientKind == "CLIENT_KIND_RESOURCE_MANAGER"
}

# Exception 2
# This rule allows southbound API to UPDATE only the watcherOsmanager field in the Tenant resource
isException if {
	input.Method == "UPDATE"
	input.resource.tenant
	input.resource.tenant.watcherOsmanager
	not input.resource.tenant.desiredState
	not input.resource.tenant.currentState
	input.ClientKind == "CLIENT_KIND_RESOURCE_MANAGER"
}

# Output rule: Determines if ABAC applies for CREATE operations
abac if {
	input.Method == "CREATE"
	resourceRule
}

# Output rule: Determines if ABAC applies for UPDATE operations
abac if {
	input.resourceId
	input.Method == "UPDATE"
	resourceRule
}

# Output rule: Determines if ABAC applies for DELETE operations
abac if {
	input.Method == "DELETE"
	not input.resource # in Delete message Resource field is not initialized (at all) an thus treated as a simple type
	deleteRule
}

resourceRule if {
	input.resource # this is to make sure that the Resource field is not empty.
	count(input.resource) != 0 # handling the case when Resource field is initialized as an empty structure, which is being converted into an empty map in JSON
	allowRule
}

# Allow request when no deny rules apply
allowRule if {
	not deny
}

# Allow request despite deny rules if an exception applies
allowRule if {
	isException
}

deleteRule if {
	input.resourceId
	input.ClientKind in {"CLIENT_KIND_API", "CLIENT_KIND_RESOURCE_MANAGER"}
}

deleteRule if {
	input.tenantId
	input.resourceKind
	input.ClientKind == "CLIENT_KIND_TENANT_CONTROLLER"
}

deleteRule if {
	input.ClientKind == "CLIENT_KIND_TENANT_CONTROLLER"
	startswith(input.resourceId, "tenant")
}
