# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

package abac

import rego.v1

default allow := false
default deny := false
default abac := false
default resourceRule := false
default deleteRule := false

allow if {
	input.DesiredState
	not input.CurrentState
	not input.resource.host.currentPowerState
	not input.resource.host.currentAmtState
	not input.resource.tenant
	input.ClientKind == "CLIENT_KIND_API"
}

allow if {
	input.CurrentState
	not input.DesiredState
	not input.resource.host.desiredPowerState
	not input.resource.host.desiredAmtState
	input.ClientKind == "CLIENT_KIND_RESOURCE_MANAGER"
}

# Allow the update of the other resources not reconciled
allow if {
	not input.DesiredState
	not input.CurrentState
	not input.resource.host.desiredPowerState
	not input.resource.host.currentPowerState
    not input.resource.host.desiredAmtState
    not input.resource.host.currentAmtState
	input.ClientKind != "CLIENT_KIND_UNSPECIFIED"
}

# Host specific rules
allow if {
	not input.CurrentState
	input.resource.host
	input.resource.host.desiredPowerState
	not input.resource.host.currentPowerState
    input.resource.host.desiredAmtState
    not input.resource.host.currentAmtState
	input.ClientKind == "CLIENT_KIND_API"
}

allow if {
	not input.CurrentState
	input.resource.host
	input.resource.host.desiredPowerState
	not input.resource.host.currentPowerState
    not input.resource.host.desiredAmtState
    not input.resource.host.currentAmtState
	input.ClientKind == "CLIENT_KIND_API"
}

allow if {
	not input.CurrentState
	input.resource.host
	not input.resource.host.desiredPowerState
	not input.resource.host.currentPowerState
    input.resource.host.desiredAmtState
    not input.resource.host.currentAmtState
	input.ClientKind == "CLIENT_KIND_API"
}

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

allow if {
	not input.DesiredState
	input.resource.host
	input.resource.host.currentPowerState
	not input.resource.host.desiredPowerState
    input.resource.host.currentAmtState
    not input.resource.host.desiredAmtState
	input.ClientKind == "CLIENT_KIND_RESOURCE_MANAGER"
}

allow if {
	not input.DesiredState
	input.resource.host
	input.resource.host.currentPowerState
	not input.resource.host.desiredPowerState
    not input.resource.host.currentAmtState
    not input.resource.host.desiredAmtState
	input.ClientKind == "CLIENT_KIND_RESOURCE_MANAGER"
}

allow if {
	not input.DesiredState
	input.resource.host
	not input.resource.host.currentPowerState
	not input.resource.host.desiredPowerState
    input.resource.host.currentAmtState
    not input.resource.host.desiredAmtState
	input.ClientKind == "CLIENT_KIND_RESOURCE_MANAGER"
}

# Instance specific rules for supporting ZTP with default OS
# This rule allows RM to CREATE a new Instance resource with desiredState set to RUNNING
# and of kind METAL. All other options for the mentioned fields are not supported
allow if {
	input.Method == "CREATE"
	input.DesiredState
	input.resource.instance
	input.resource.instance.kind == "INSTANCE_KIND_METAL"
	input.resource.instance.desiredState == "INSTANCE_STATE_RUNNING"
	input.ClientKind == "CLIENT_KIND_RESOURCE_MANAGER"
}

# What should we do here? Do we want to really enforce control on hostStatus ?
#allow if {
#	not input.DesiredState
#	input.resource.host
#	input.resource.host.hostStatus
#	not input.resource.host.desiredPowerState
#	input.ClientKind == "CLIENT_KIND_RESOURCE_MANAGER"
#}

allow if {
	input.ClientKind == "CLIENT_KIND_TENANT_CONTROLLER"
	with input.resource as {"tenant", "provider", "telemetryGroup"}
}

deleteRule if {
	input.Method == "DELETE"
	not input.resource
	input.ClientKind == "CLIENT_KIND_TENANT_CONTROLLER"
	startswith(input.resourceId, "tenant")
}

abac if {
	input.Method == "CREATE"
	resourceRule
}

abac if {
	input.resourceId
	input.Method == "UPDATE"
	resourceRule
}

resourceRule if {
	input.resource # this is to make sure that the Resource field is not empty.
	count(input.resource) != 0 # handling the case when Resource field is initialized as an empty structure, which is being converted into an empty map in JSON
	allow
	not deny
}

abac if {
	input.Method == "DELETE"
	not input.resource # in Delete message Resource field is not initialized (at all) an thus treated as a simple type
	deleteRule
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
