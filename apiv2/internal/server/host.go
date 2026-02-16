// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	commonv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/common/v1"
	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	locationv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/location/v1"
	networkv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/network/v1"
	providerv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/provider/v1"
	statusv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/status/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	inv_networkv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPIHostToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs HostTemplate and HostBmManagementInfo defined in
// edge-infra-manager-openapi-types.gen.go.
// It handles naming mismatches between API and inventory,
// and includes only mutable fields used in UPDATE/PATCH requests.
var OpenAPIHostToProto = map[string]string{
	computev1.HostResourceFieldName:               inv_computev1.HostResourceFieldName,
	computev1.HostResourceFieldSiteId:             inv_computev1.HostResourceEdgeSite,
	computev1.HostResourceEdgeMetadata:            inv_computev1.HostResourceFieldMetadata,
	computev1.HostResourceFieldDesiredPowerState:  inv_computev1.HostResourceFieldDesiredPowerState,
	computev1.HostResourceFieldDesiredAmtState:    inv_computev1.HostResourceFieldDesiredAmtState,
	computev1.HostResourceFieldPowerCommandPolicy: inv_computev1.HostResourceFieldPowerCommandPolicy,
	computev1.HostResourceFieldAmtControlMode:     inv_computev1.HostResourceFieldAmtControlMode,
	computev1.HostResourceFieldAmtDnsSuffix:       inv_computev1.HostResourceFieldAmtDnsSuffix,
}

var (
	filterIsFailedHostStatusExp = `%s = %s OR %s = %s OR %s = %s OR %s = %s OR %s = %s OR ` +
		`%s.%s = %s OR %s.%s = %s OR %s.%s = %s OR %s.%s = %s`
	filterIsFailedHostStatus = fmt.Sprintf(filterIsFailedHostStatusExp,
		inv_computev1.HostResourceFieldHostStatusIndicator,
		api.STATUSINDICATIONERROR,
		inv_computev1.HostResourceFieldOnboardingStatusIndicator,
		api.STATUSINDICATIONERROR,
		inv_computev1.HostResourceFieldRegistrationStatusIndicator,
		api.STATUSINDICATIONERROR,
		inv_computev1.HostResourceFieldPowerStatusIndicator,
		api.STATUSINDICATIONERROR,
		inv_computev1.HostResourceFieldAmtStatusIndicator,
		api.STATUSINDICATIONERROR,
		inv_computev1.HostResourceEdgeInstance,
		inv_computev1.InstanceResourceFieldInstanceStatusIndicator,
		api.STATUSINDICATIONERROR,
		inv_computev1.HostResourceEdgeInstance,
		inv_computev1.InstanceResourceFieldProvisioningStatusIndicator,
		api.STATUSINDICATIONERROR,
		inv_computev1.HostResourceEdgeInstance,
		inv_computev1.InstanceResourceFieldUpdateStatusIndicator,
		api.STATUSINDICATIONERROR,
		inv_computev1.HostResourceEdgeInstance,
		inv_computev1.InstanceResourceFieldTrustedAttestationStatusIndicator,
		api.STATUSINDICATIONERROR,
	)

	// Create a filter specifically for instance error states.
	filterIsFailedInstanceStatusExp = `%s.%s = %s OR %s.%s = %s OR %s.%s = %s OR %s.%s = %s`
	filterIsFailedInstanceStatus    = fmt.Sprintf(filterIsFailedInstanceStatusExp,
		inv_computev1.HostResourceEdgeInstance,
		inv_computev1.InstanceResourceFieldInstanceStatusIndicator,
		api.STATUSINDICATIONERROR,
		inv_computev1.HostResourceEdgeInstance,
		inv_computev1.InstanceResourceFieldProvisioningStatusIndicator,
		api.STATUSINDICATIONERROR,
		inv_computev1.HostResourceEdgeInstance,
		inv_computev1.InstanceResourceFieldUpdateStatusIndicator,
		api.STATUSINDICATIONERROR,
		inv_computev1.HostResourceEdgeInstance,
		inv_computev1.InstanceResourceFieldTrustedAttestationStatusIndicator,
		api.STATUSINDICATIONERROR,
	)

	// Modify running instance filter to only exclude instance-related errors
	// and not host-related errors.
	filterInstanceRunningExp = `has(%s) AND %s.%s = %s AND NOT (%s)`
	filterInstanceRunning    = fmt.Sprintf(filterInstanceRunningExp,
		inv_computev1.HostResourceEdgeInstance,
		inv_computev1.HostResourceEdgeInstance,
		inv_computev1.InstanceResourceFieldCurrentState,
		inv_computev1.InstanceState_INSTANCE_STATE_RUNNING,
		filterIsFailedInstanceStatus,
	)

	filterIsUnallocatedExp = `NOT has(%s)`
	filterIsUnallocated    = fmt.Sprintf(filterIsUnallocatedExp,
		inv_computev1.HostResourceEdgeSite,
	)
)

func toInvHost(host *computev1.HostResource) (*inv_computev1.HostResource, error) {
	if host == nil {
		return &inv_computev1.HostResource{}, nil
	}

	metadata, err := toInvMetadata(host.GetMetadata())
	if err != nil {
		return nil, err
	}

	invHost := &inv_computev1.HostResource{
		Name:               host.GetName(),
		Uuid:               host.GetUuid(),
		SerialNumber:       host.GetSerialNumber(),
		DesiredState:       inv_computev1.HostState_HOST_STATE_ONBOARDED,
		Metadata:           metadata,
		AmtDnsSuffix:       host.GetAmtDnsSuffix(),
		DesiredPowerState:  inv_computev1.PowerState_POWER_STATE_ON,
		DesiredAmtState:    inv_computev1.AmtState(host.GetDesiredAmtState()),
		AmtSku:             inv_computev1.AmtSku(host.GetAmtSku()),
		AmtControlMode:     inv_computev1.AmtControlMode(host.GetAmtControlMode()),
		PowerCommandPolicy: inv_computev1.PowerCommandPolicy_POWER_COMMAND_POLICY_ORDERED,
		UserLvmSize:        host.GetUserLvmSize(),
	}

	hostSiteID := host.GetSiteId()
	if isSet(&hostSiteID) {
		invHost.Site = &inv_locationv1.SiteResource{
			ResourceId: host.GetSiteId(),
		}
	}

	err = validator.ValidateMessage(invHost)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}

	return invHost, nil
}

func toInvHostUpdate(host *computev1.HostResource) (*inv_computev1.HostResource, error) {
	if host == nil {
		return &inv_computev1.HostResource{}, nil
	}

	metadata, err := toInvMetadata(host.GetMetadata())
	if err != nil {
		return nil, err
	}

	invHost := &inv_computev1.HostResource{
		Name:               host.GetName(),
		Metadata:           metadata,
		DesiredPowerState:  inv_computev1.PowerState(host.GetDesiredPowerState()),
		DesiredAmtState:    inv_computev1.AmtState(host.GetDesiredAmtState()),
		AmtDnsSuffix:       host.GetAmtDnsSuffix(),
		AmtSku:             inv_computev1.AmtSku(host.GetAmtSku()),
		AmtControlMode:     inv_computev1.AmtControlMode(host.GetAmtControlMode()),
		PowerCommandPolicy: inv_computev1.PowerCommandPolicy(host.GetPowerCommandPolicy()),
	}

	hostSiteID := host.GetSiteId()
	if isSet(&hostSiteID) {
		invHost.Site = &inv_locationv1.SiteResource{
			ResourceId: host.GetSiteId(),
		}
	}

	err = validator.ValidateMessage(invHost)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}
	return invHost, nil
}

func fromInvHostStatus(
	invHost *inv_computev1.HostResource,
	host *computev1.HostResource,
) {
	hostStatus := invHost.GetHostStatus()
	hostStatusIndicator := statusv1.StatusIndication(invHost.GetHostStatusIndicator())
	hostStatusTimestamp := TruncateUint64ToUint32(invHost.GetHostStatusTimestamp())

	onboardingStatus := invHost.GetOnboardingStatus()
	onboardingStatusIndicator := statusv1.StatusIndication(invHost.GetOnboardingStatusIndicator())
	onboardingStatusTimestamp := TruncateUint64ToUint32(invHost.GetOnboardingStatusTimestamp())

	registrationStatus := invHost.GetRegistrationStatus()
	registrationStatusIndicator := statusv1.StatusIndication(invHost.GetRegistrationStatusIndicator())
	registrationStatusTimestamp := TruncateUint64ToUint32(invHost.GetRegistrationStatusTimestamp())

	powerStatus := invHost.GetPowerStatus()
	powerStatusIndicator := statusv1.StatusIndication(invHost.GetPowerStatusIndicator())
	powerStatusTimestamp := TruncateUint64ToUint32(invHost.GetPowerStatusTimestamp())

	amtStatus := invHost.GetAmtStatus()
	amtDNSSuffix := invHost.GetAmtDnsSuffix()
	amtStatusIndicator := statusv1.StatusIndication(invHost.GetPowerStatusIndicator())
	amtStatusTimestamp := TruncateUint64ToUint32(invHost.GetAmtStatusTimestamp())

	host.HostStatus = hostStatus
	host.HostStatusIndicator = hostStatusIndicator
	host.HostStatusTimestamp = hostStatusTimestamp
	host.OnboardingStatus = onboardingStatus
	host.OnboardingStatusIndicator = onboardingStatusIndicator
	host.OnboardingStatusTimestamp = onboardingStatusTimestamp
	host.RegistrationStatus = registrationStatus
	host.RegistrationStatusIndicator = registrationStatusIndicator
	host.RegistrationStatusTimestamp = registrationStatusTimestamp
	host.PowerStatus = powerStatus
	host.PowerStatusIndicator = powerStatusIndicator
	host.PowerStatusTimestamp = powerStatusTimestamp
	host.AmtStatus = amtStatus
	host.AmtStatusIndicator = amtStatusIndicator
	host.AmtStatusTimestamp = amtStatusTimestamp
	host.AmtDnsSuffix = amtDNSSuffix
}

func fromInvHostEdges(
	invHost *inv_computev1.HostResource,
	host *computev1.HostResource,
) error {
	var err error
	var hostInstance *computev1.InstanceResource
	if invHost.GetInstance() != nil {
		hostInstance, err = fromInvInstance(invHost.GetInstance())
		if err != nil {
			return err
		}
	}
	var hostSite *locationv1.SiteResource
	if invHost.GetSite() != nil {
		hostSite, err = fromInvSite(invHost.GetSite(), nil)
		if err != nil {
			return err
		}
	}
	var hostProvider *providerv1.ProviderResource
	if invHost.GetProvider() != nil {
		hostProvider = fromInvProvider(invHost.GetProvider())
	}
	host.Instance = hostInstance
	host.Site = hostSite
	host.Provider = hostProvider
	return nil
}

func fromInvHost(
	invHost *inv_computev1.HostResource,
	resMeta *inventory.GetResourceResponse_ResourceMetadata,
	nicToIPAdrresses map[string][]*networkv1.IPAddressResource,
) (*computev1.HostResource, error) {
	if invHost == nil {
		return &computev1.HostResource{}, nil
	}

	metadata, err := fromInvMetadata(invHost.GetMetadata())
	if err != nil {
		return nil, err
	}

	host := &computev1.HostResource{
		ResourceId:         invHost.GetResourceId(),
		Name:               invHost.GetName(),
		DesiredState:       computev1.HostState(invHost.GetDesiredState()),
		CurrentState:       computev1.HostState(invHost.GetCurrentState()),
		SiteId:             invHost.GetSite().GetResourceId(),
		Note:               invHost.GetNote(),
		SerialNumber:       invHost.GetSerialNumber(),
		MemoryBytes:        fmt.Sprintf("%d", invHost.GetMemoryBytes()),
		CpuModel:           invHost.GetCpuModel(),
		CpuSockets:         invHost.GetCpuSockets(),
		CpuCores:           invHost.GetCpuCores(),
		CpuCapabilities:    invHost.GetCpuCapabilities(),
		CpuArchitecture:    invHost.GetCpuArchitecture(),
		CpuThreads:         invHost.GetCpuThreads(),
		CpuTopology:        invHost.GetCpuTopology(),
		BmcKind:            computev1.BaremetalControllerKind(invHost.GetBmcKind()),
		BmcIp:              invHost.GetBmcIp(),
		Hostname:           invHost.GetHostname(),
		ProductName:        invHost.GetProductName(),
		BiosVersion:        invHost.GetBiosVersion(),
		BiosReleaseDate:    invHost.GetBiosReleaseDate(),
		BiosVendor:         invHost.GetBiosVendor(),
		CurrentPowerState:  computev1.PowerState(invHost.GetCurrentPowerState()),
		DesiredPowerState:  computev1.PowerState(invHost.GetDesiredPowerState()),
		PowerCommandPolicy: computev1.PowerCommandPolicy(invHost.GetPowerCommandPolicy()),
		PowerOnTime:        TruncateUint64ToUint32(invHost.GetPowerOnTime()),
		HostStorages:       fromInvHostStorages(invHost.GetHostStorages()),
		HostNics:           fromInvHostNics(invHost.GetHostNics(), nicToIPAdrresses),
		HostUsbs:           fromInvHostUsbs(invHost.GetHostUsbs()),
		HostGpus:           fromInvHostGpus(invHost.GetHostGpus()),
		AmtDnsSuffix:       invHost.GetAmtDnsSuffix(),
		AmtSku:             computev1.AmtSku(invHost.GetAmtSku()),
		AmtControlMode:     computev1.AmtControlMode(invHost.GetAmtControlMode()),
		DesiredAmtState:    computev1.AmtState(invHost.GetDesiredAmtState()),
		CurrentAmtState:    computev1.AmtState(invHost.GetCurrentAmtState()),
		Metadata:           metadata,
		InheritedMetadata:  []*commonv1.MetadataItem{},
		Timestamps:         GrpcToOpenAPITimestamps(invHost),
		UserLvmSize:        invHost.GetUserLvmSize(),
	}

	if err = fromInvHostEdges(invHost, host); err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory host edges")
		return nil, errors.Wrap(err)
	}
	fromInvHostStatus(invHost, host)

	hostUUID := invHost.GetUuid()
	if isSet(&hostUUID) {
		host.Uuid = hostUUID
	}

	if resMeta != nil {
		inheritedMetadata, err := fromInvMetadata(resMeta.GetPhyMetadata())
		if err != nil {
			return nil, err
		}
		host.InheritedMetadata = inheritedMetadata
	}
	return host, nil
}

func fromInvHostStorages(storages []*inv_computev1.HoststorageResource) []*computev1.HoststorageResource {
	// Conversion logic for HostStorages
	hostStorages := make([]*computev1.HoststorageResource, 0, len(storages))
	for _, storage := range storages {
		hostStorages = append(hostStorages, &computev1.HoststorageResource{
			Wwid:          storage.GetWwid(),
			Serial:        storage.GetSerial(),
			Vendor:        storage.GetVendor(),
			Model:         storage.GetModel(),
			CapacityBytes: fmt.Sprintf("%d", storage.GetCapacityBytes()),
			DeviceName:    storage.GetDeviceName(),
			Timestamps:    GrpcToOpenAPITimestamps(storage),
		})
	}
	return hostStorages
}

func fromInvHostNics(
	nics []*inv_computev1.HostnicResource,
	nicToIPAdrresses map[string][]*networkv1.IPAddressResource,
) []*computev1.HostnicResource {
	// Conversion logic for HostNics
	hostNics := make([]*computev1.HostnicResource, 0, len(nics))
	for _, nic := range nics {
		ipAdresses := []*networkv1.IPAddressResource{}
		if nicToIPAdrresses != nil {
			ipAdresses = nicToIPAdrresses[nic.GetResourceId()]
		}

		linkState := &computev1.NetworkInterfaceLinkState{
			Type: computev1.LinkState(nic.GetLinkState()),
		}
		hostNics = append(hostNics, &computev1.HostnicResource{
			DeviceName:    nic.GetDeviceName(),
			PciIdentifier: nic.GetPciIdentifier(),
			MacAddr:       nic.GetMacAddr(),
			SriovEnabled:  nic.GetSriovEnabled(),
			SriovVfsNum:   nic.GetSriovVfsNum(),
			SriovVfsTotal: nic.GetSriovVfsTotal(),
			Mtu:           nic.GetMtu(),
			LinkState:     linkState,
			BmcInterface:  nic.GetBmcInterface(),
			Ipaddresses:   ipAdresses,
			Timestamps:    GrpcToOpenAPITimestamps(nic),
		})
	}
	return hostNics
}

func fromInvHostUsbs(usbs []*inv_computev1.HostusbResource) []*computev1.HostusbResource {
	// Conversion logic for HostUsbs
	hostUsbs := make([]*computev1.HostusbResource, 0, len(usbs))
	for _, usb := range usbs {
		hostUsbs = append(hostUsbs, &computev1.HostusbResource{
			IdVendor:   usb.GetIdvendor(),
			IdProduct:  usb.GetIdproduct(),
			Bus:        usb.GetBus(),
			Addr:       usb.GetAddr(),
			Class:      usb.GetClass(),
			Serial:     usb.GetSerial(),
			DeviceName: usb.GetDeviceName(),
			Timestamps: GrpcToOpenAPITimestamps(usb),
		})
	}
	return hostUsbs
}

func fromInvHostGpus(gpus []*inv_computev1.HostgpuResource) []*computev1.HostgpuResource {
	// Conversion logic for HostGpus
	hostGpus := make([]*computev1.HostgpuResource, 0, len(gpus))
	for _, gpu := range gpus {
		gpuCapabilities := strings.Split(gpu.GetFeatures(), ",")
		hostGpus = append(hostGpus, &computev1.HostgpuResource{
			PciId:        gpu.GetPciId(),
			Product:      gpu.GetProduct(),
			Vendor:       gpu.GetVendor(),
			Description:  gpu.GetDescription(),
			DeviceName:   gpu.GetDeviceName(),
			Capabilities: gpuCapabilities,
			Timestamps:   GrpcToOpenAPITimestamps(gpu),
		})
	}
	return hostGpus
}

func (is *InventorygRPCServer) CreateHost(
	ctx context.Context,
	req *restv1.CreateHostRequest,
) (*computev1.HostResource, error) {
	zlog.Debug().Msg("CreateHost")

	host := req.GetHost()
	invHost, err := toInvHost(host)
	if err != nil {
		zlog.Error().Err(err).Msg("toInvHost failed")
		return nil, errors.Wrap(err)
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: invHost,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to create inventory resource %s", invRes)
		return nil, errors.Wrap(err)
	}

	hostCreated, err := fromInvHost(invResp.GetHost(), nil, nil)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Created %s", hostCreated)
	return hostCreated, nil
}

// Get a list of hosts.
func (is *InventorygRPCServer) ListHosts(
	ctx context.Context,
	req *restv1.ListHostsRequest,
) (*restv1.ListHostsResponse, error) {
	zlog.Debug().Msg("ListHosts")

	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_Host{Host: &inv_computev1.HostResource{}}},
		Offset:   req.GetOffset(),
		Limit:    req.GetPageSize(),
		OrderBy:  req.GetOrderBy(),
		Filter:   req.GetFilter(),
	}

	if err := validator.ValidateMessage(filter); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("failed to validate query params")
		return nil, errors.Wrap(err)
	}

	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to list inventory resources %s", filter)
		return nil, errors.Wrap(err)
	}

	invResources := invResp.GetResources()
	hosts := make([]*computev1.HostResource, 0, len(invResources))
	for _, invRes := range invResources {
		nicToIPAddresses, err := is.getInterfaceToIPAddresses(ctx, invRes.GetResource().GetHost())
		if err != nil {
			zlog.Error().Err(err).Msgf("failed to get IP addresses for host %s",
				invRes.GetResource().GetHost().GetResourceId())
			return nil, err
		}

		host, err := fromInvHost(invRes.GetResource().GetHost(), invRes.GetRenderedMetadata(), nicToIPAddresses)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		hosts = append(hosts, host)
	}

	resp := &restv1.ListHostsResponse{
		Hosts:         hosts,
		TotalElements: invResp.GetTotalElements(),
		HasNext:       invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific host.
func (is *InventorygRPCServer) GetHost(ctx context.Context, req *restv1.GetHostRequest) (*computev1.HostResource, error) {
	zlog.Debug().Msg("GetHost")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to get inventory resource %s", req.GetResourceId())
		return nil, err
	}

	invHost := invResp.GetResource().GetHost()
	nicToIPAddresses, err := is.getInterfaceToIPAddresses(ctx, invHost)
	if err != nil {
		zlog.Error().Err(err).Msgf("failed to get IP addresses for host %s",
			invHost.GetResourceId())
		return nil, errors.Wrap(err)
	}

	host, err := fromInvHost(invHost, invResp.GetRenderedMetadata(), nicToIPAddresses)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Got %s", host)
	return host, nil
}

// handleConsecutivePowerReset manages consecutive power reset operations for hosts.
func (is *InventorygRPCServer) handleConsecutivePowerReset(
	ctx context.Context,
	resourceID string,
	invHost *inv_computev1.HostResource,
) {
	if invHost.GetDesiredPowerState() != inv_computev1.PowerState_POWER_STATE_RESET {
		return
	}
	zlog.Info().Msgf("Processing RESET request for host %s", resourceID)
	currentHostRes, getErr := is.InvClient.Get(ctx, resourceID)
	if getErr != nil {
		zlog.Warn().Err(getErr).Msgf("Could not retrieve current host state for %s, proceeding with standard reset",
			resourceID)
		return
	}
	currentHost := currentHostRes.GetResource().GetHost()
	if currentHost == nil {
		return
	}
	currentPowerState := currentHost.GetCurrentPowerState()
	currentDesiredState := currentHost.GetDesiredPowerState()
	zlog.Info().Msgf("Host %s state analysis: current=%v, desired=%v",
		resourceID, currentPowerState, currentDesiredState)

	// Check both current and desired power states for consecutive resets
	switch {
	case currentPowerState == inv_computev1.PowerState_POWER_STATE_RESET ||
		currentDesiredState == inv_computev1.PowerState_POWER_STATE_RESET:
		zlog.Info().Msgf("RESET -> RESET_REPEAT for host %s", resourceID)
		invHost.DesiredPowerState = inv_computev1.PowerState_POWER_STATE_RESET_REPEAT
	case currentPowerState == inv_computev1.PowerState_POWER_STATE_RESET_REPEAT &&
		currentDesiredState == inv_computev1.PowerState_POWER_STATE_RESET_REPEAT:
		zlog.Info().Msgf("RESET_REPEAT -> RESET for host %s", resourceID)
		invHost.DesiredPowerState = inv_computev1.PowerState_POWER_STATE_RESET
	default:
		zlog.Info().Msgf("Standard reset operation for host %s (current=%v, desired=%v)",
			resourceID, currentPowerState, currentDesiredState)
	}

	zlog.Info().Msgf("Final power state being sent to inventory for host %s: %v",
		resourceID, invHost.DesiredPowerState)
}

// Update a host. (PUT).
func (is *InventorygRPCServer) UpdateHost(
	ctx context.Context,
	req *restv1.UpdateHostRequest,
) (*computev1.HostResource, error) {
	zlog.Debug().Msg("UpdateHost")

	host := req.GetHost()
	invHost, err := toInvHostUpdate(host)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// Handle consecutive power reset operations
	is.handleConsecutivePowerReset(ctx, req.GetResourceId(), invHost)

	fieldmask, err := fieldmaskpb.New(invHost, maps.Values(OpenAPIHostToProto)...)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: invHost,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}
	invUp := upRes.GetHost()
	invUpRes, err := fromInvHost(invUp, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Update a host. (PATCH).
func (is *InventorygRPCServer) PatchHost(
	ctx context.Context,
	req *restv1.PatchHostRequest,
) (*computev1.HostResource, error) {
	zlog.Debug().Msg("PatchHost")

	host := req.GetHost()
	invHost, err := toInvHostUpdate(host)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// Handle consecutive power reset operations
	is.handleConsecutivePowerReset(ctx, req.GetResourceId(), invHost)

	fieldmask, err := parseFielmask(invHost, req.GetFieldMask(), OpenAPIHostToProto)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: invHost,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, errors.Wrap(err)
	}
	invUp := upRes.GetHost()
	invUpRes, err := fromInvHost(invUp, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Delete a host.
func (is *InventorygRPCServer) DeleteHost(
	ctx context.Context,
	req *restv1.DeleteHostRequest,
) (*restv1.DeleteHostResponse, error) {
	zlog.Debug().Msg("DeleteHost")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to delete inventory resource %s", req.GetResourceId())
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteHostResponse{}, nil
}

// Invalidate a host.
func (is *InventorygRPCServer) InvalidateHost(
	ctx context.Context,
	req *restv1.InvalidateHostRequest,
) (*restv1.InvalidateHostResponse, error) {
	zlog.Debug().Msg("InvalidateHost")
	res := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: &inv_computev1.HostResource{
				DesiredState: inv_computev1.HostState_HOST_STATE_UNTRUSTED,
				Note:         req.GetNote(),
			},
		},
	}

	fm, err := fieldmaskpb.New(
		res.GetHost(),
		inv_computev1.HostResourceFieldDesiredState,
		inv_computev1.HostResourceFieldNote,
	)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	_, err = is.InvClient.Update(ctx, req.GetResourceId(), fm, res)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Invalidated %s", req.GetResourceId())
	return &restv1.InvalidateHostResponse{}, nil
}

// Register a host.
func (is *InventorygRPCServer) RegisterHost(
	ctx context.Context,
	req *restv1.RegisterHostRequest,
) (*computev1.HostResource, error) {
	zlog.Debug().Msg("RegisterHost")
	hostResource := &inv_computev1.HostResource{
		Name:              req.GetHost().GetName(),
		DesiredState:      inv_computev1.HostState_HOST_STATE_REGISTERED,
		DesiredPowerState: inv_computev1.PowerState_POWER_STATE_ON,
		DesiredAmtState:   inv_computev1.AmtState_AMT_STATE_UNPROVISIONED,
	}

	hostUUID := req.GetHost().GetUuid()
	if isSet(&hostUUID) {
		hostResource.Uuid = hostUUID
	}
	hostSerial := req.GetHost().GetSerialNumber()
	if isSet(&hostSerial) {
		hostResource.SerialNumber = hostSerial
	}

	if isUnset(&hostUUID) && isUnset(&hostSerial) {
		err := errors.Errorfc(codes.InvalidArgument, "either UUID or SerialNumber must be set")
		zlog.InfraErr(err).Msg("Failed to parse register host fields")
		return nil, err
	}

	if req.GetHost().GetAutoOnboard() {
		hostResource.DesiredState = inv_computev1.HostState_HOST_STATE_ONBOARDED
	}

	if req.GetHost().GetEnableVpro() {
		hostResource.DesiredAmtState = inv_computev1.AmtState_AMT_STATE_PROVISIONED
	}

	hostUserLvmSize := req.GetHost().GetUserLvmSize()
	hostResource.UserLvmSize = hostUserLvmSize

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: hostResource,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to create inventory resource %s", invRes)
		return nil, errors.Wrap(err)
	}

	hostResp, err := fromInvHost(invResp.GetHost(), nil, nil)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	zlog.Debug().Msgf("Registered %s", hostResp)
	return hostResp, nil
}

// Onboard a host.
func (is *InventorygRPCServer) OnboardHost(
	ctx context.Context,
	req *restv1.OnboardHostRequest,
) (*restv1.OnboardHostResponse, error) {
	zlog.Debug().Msg("OnboardHost")
	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: &inv_computev1.HostResource{
				DesiredState: inv_computev1.HostState_HOST_STATE_ONBOARDED,
			},
		},
	}

	fm, err := fieldmaskpb.New(
		invRes.GetHost(),
		inv_computev1.HostResourceFieldDesiredState,
	)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fm, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, errors.Wrap(err)
	}

	invUp := upRes.GetHost()
	invUpRes, err := fromInvHost(invUp, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	zlog.Debug().Msgf("Onboarded %s", invUpRes)
	return &restv1.OnboardHostResponse{}, nil
}

// Onboard a host.
func (is *InventorygRPCServer) PatchRegisterHost(
	ctx context.Context,
	req *restv1.RegisterHostRequest,
) (*computev1.HostResource, error) {
	zlog.Debug().Msg("PatchRegisterHost")
	hostResource := &inv_computev1.HostResource{
		Name:            req.GetHost().GetName(),
		DesiredState:    inv_computev1.HostState_HOST_STATE_REGISTERED,
		DesiredAmtState: inv_computev1.AmtState_AMT_STATE_UNPROVISIONED,
	}
	fieldList := []string{inv_computev1.HostResourceFieldName, inv_computev1.HostResourceFieldDesiredState}

	if req.GetHost().GetAutoOnboard() {
		hostResource.DesiredState = inv_computev1.HostState_HOST_STATE_ONBOARDED
	}

	if req.GetHost().GetEnableVpro() {
		hostResource.DesiredAmtState = inv_computev1.AmtState_AMT_STATE_PROVISIONED
	}
	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: hostResource,
		},
	}

	fm, err := fieldmaskpb.New(
		invRes.GetHost(),
		fieldList...,
	)
	if err != nil {
		return nil, err
	}

	invReply, err := is.InvClient.Update(ctx, req.GetResourceId(), fm, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}

	invHost, err := fromInvHost(invReply.GetHost(), nil, nil)
	if err != nil {
		return nil, err
	}
	zlog.Debug().Msgf("Updated %s", invHost)
	return invHost, nil
}

func (is *InventorygRPCServer) totalHosts(ctx context.Context, filter string) (int32, error) {
	var pageSize uint32 = 1
	req := &restv1.ListHostsRequest{
		Filter:   filter,
		PageSize: pageSize,
	}
	hostsList, err := is.ListHosts(ctx, req)
	if err != nil {
		return 0, err
	}
	return hostsList.GetTotalElements(), nil
}

// Get hosts summary.
func (is *InventorygRPCServer) GetHostsSummary(
	ctx context.Context,
	req *restv1.GetHostSummaryRequest,
) (*restv1.GetHostSummaryResponse, error) {
	zlog.Debug().Msg("GetHostsSummary")

	var total uint32
	var errorState uint32
	var runningState uint32
	var unallocatedState uint32
	reqFilter := req.GetFilter()

	filterTotal := reqFilter
	filterIsFailedHostStatusParsed := filterIsFailedHostStatus
	filterInstanceRunningParsed := filterInstanceRunning
	filterIsUnallocatedParsed := filterIsUnallocated

	if reqFilter != "" {
		filterIsFailedHostStatusParsed = fmt.Sprintf("%s AND (%s)", reqFilter, filterIsFailedHostStatusParsed)
		filterInstanceRunningParsed = fmt.Sprintf("%s AND (%s)", reqFilter, filterInstanceRunningParsed)
		filterIsUnallocatedParsed = fmt.Sprintf("%s AND (%s)", reqFilter, filterIsUnallocatedParsed)
	}

	totalHosts, err := is.totalHosts(ctx, filterTotal)
	if err != nil {
		return nil, err
	}
	totalHostsError, err := is.totalHosts(ctx, filterIsFailedHostStatusParsed)
	if err != nil {
		return nil, err
	}
	totalHostsUnallocated, err := is.totalHosts(ctx, filterIsUnallocatedParsed)
	if err != nil {
		return nil, err
	}
	totalHostsRunning, err := is.totalHosts(ctx, filterInstanceRunningParsed)
	if err != nil {
		return nil, err
	}

	total, err = SafeInt32ToUint32(totalHosts)
	if err != nil {
		return nil, err
	}
	errorState, err = SafeInt32ToUint32(totalHostsError)
	if err != nil {
		return nil, err
	}
	unallocatedState, err = SafeInt32ToUint32(totalHostsUnallocated)
	if err != nil {
		return nil, err
	}
	runningState, err = SafeInt32ToUint32(totalHostsRunning)
	if err != nil {
		return nil, err
	}
	hostsSummary := &restv1.GetHostSummaryResponse{
		Total:       total,
		Error:       errorState,
		Running:     runningState,
		Unallocated: unallocatedState,
	}

	return hostsSummary, nil
}

func fromInvIPAddresses(
	invIPAddresses []*inv_networkv1.IPAddressResource,
) []*networkv1.IPAddressResource {
	IPAddresses := make([]*networkv1.IPAddressResource, 0, len(invIPAddresses))
	for _, invIPAddress := range invIPAddresses {
		configMode := networkv1.IPAddressConfigMethod(invIPAddress.GetConfigMethod())
		status := networkv1.IPAddressStatus(invIPAddress.GetStatus())
		cidrAddress := invIPAddress.GetAddress()
		statusDetail := invIPAddress.GetStatusDetail()
		ipAddress := &networkv1.IPAddressResource{
			ConfigMethod: configMode,
			Status:       status,
			Address:      cidrAddress,
			StatusDetail: statusDetail,
		}
		IPAddresses = append(IPAddresses, ipAddress)
	}
	return IPAddresses
}

func castToIPAddress(resp *inventory.GetResourceResponse) (*inv_networkv1.IPAddressResource, error) {
	if resp.GetResource().GetIpaddress() != nil {
		return resp.GetResource().GetIpaddress(), nil
	}
	err := errors.Errorfc(codes.Internal, "%s is not an IPAddress", resp.GetResource())
	zlog.InfraErr(err).Msgf("could not cast inventory resource")
	return nil, err
}

func (is *InventorygRPCServer) getInterfaceToIPAddresses(
	ctx context.Context,
	host *inv_computev1.HostResource,
) (map[string][]*networkv1.IPAddressResource, error) {
	nicToIPAddresses := make(map[string][]*networkv1.IPAddressResource)
	hostInterfaces := host.GetHostNics()
	for _, hostInterface := range hostInterfaces {
		ipAddresses := make([]*networkv1.IPAddressResource, 0)
		req := &inventory.ResourceFilter{
			Resource: &inventory.Resource{Resource: &inventory.Resource_Ipaddress{}},
			Filter: fmt.Sprintf("%s.%s = %q", inv_networkv1.IPAddressResourceEdgeNic,
				inv_computev1.HostnicResourceFieldResourceId, hostInterface.GetResourceId()),
		}
		inventoryRes, err := is.InvClient.List(ctx, req)
		if errors.IsNotFound(err) {
			// resp is nil but we can continue in this case
			nicToIPAddresses[hostInterface.GetResourceId()] = ipAddresses
			continue
		}
		if err != nil {
			return nil, err
		}
		invIPAddresses := make([]*inv_networkv1.IPAddressResource, 0)
		for _, ipResp := range inventoryRes.GetResources() {
			invIPAddress, err := castToIPAddress(ipResp)
			if err != nil {
				return nil, err
			}
			invIPAddresses = append(invIPAddresses, invIPAddress)
		}
		IPAddresses := fromInvIPAddresses(invIPAddresses)
		nicToIPAddresses[hostInterface.GetResourceId()] = IPAddresses
	}
	return nicToIPAddresses, nil
}
