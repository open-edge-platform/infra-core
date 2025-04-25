// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"

	"golang.org/x/exp/maps"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	commonv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/common/v1"
	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	locationv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/location/v1"
	statusv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/status/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPIHostToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs HostTemplate and HostBmManagementInfo defined in
// edge-infra-manager-openapi-types.gen.go.
// Here we should have only fields that are writable from the API.
var OpenAPIHostToProto = map[string]string{
	"Name":     inv_computev1.HostResourceFieldName,
	"SiteId":   inv_computev1.HostResourceEdgeSite,
	"Metadata": inv_computev1.HostResourceFieldMetadata,
}

func toInvHost(host *computev1.HostResource) (*inv_computev1.HostResource, error) {
	if host == nil {
		return &inv_computev1.HostResource{}, nil
	}

	metadata, err := toInvMetadata(host.GetMetadata())
	if err != nil {
		return nil, err
	}

	invHost := &inv_computev1.HostResource{
		Name:         host.GetName(),
		Uuid:         host.GetUuid(),
		SerialNumber: host.GetSerialNumber(),
		DesiredState: inv_computev1.HostState_HOST_STATE_ONBOARDED,
		Metadata:     metadata,
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
		Name:     host.GetName(),
		Metadata: metadata,
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

func fromInvHost(
	invHost *inv_computev1.HostResource,
	resMeta *inventory.GetResourceResponse_ResourceMetadata,
) (*computev1.HostResource, error) {
	if invHost == nil {
		return &computev1.HostResource{}, nil
	}

	metadata, err := fromInvMetadata(invHost.GetMetadata())
	if err != nil {
		return nil, err
	}
	var hostInstance *computev1.InstanceResource
	if invHost.GetInstance() != nil {
		hostInstance, err = fromInvInstance(invHost.GetInstance())
		if err != nil {
			return nil, err
		}
	}
	var hostSite *locationv1.SiteResource
	if invHost.GetSite() != nil {
		hostSite, err = fromInvSite(invHost.GetSite(), nil)
		if err != nil {
			return nil, err
		}
	}

	hostStatus := invHost.GetHostStatus()
	hostStatusIndicator := statusv1.StatusIndication(invHost.GetHostStatusIndicator())
	hostStatusTimestamp := fmt.Sprintf("%d", invHost.GetHostStatusTimestamp())

	onboardingStatus := invHost.GetOnboardingStatus()
	onboardingStatusIndicator := statusv1.StatusIndication(invHost.GetOnboardingStatusIndicator())
	onboardingStatusTimestamp := fmt.Sprintf("%d", invHost.GetOnboardingStatusTimestamp())

	registrationStatus := invHost.GetRegistrationStatus()
	registrationStatusIndicator := statusv1.StatusIndication(invHost.GetRegistrationStatusIndicator())
	registrationStatusTimestamp := fmt.Sprintf("%d", invHost.GetRegistrationStatusTimestamp())

	host := &computev1.HostResource{
		ResourceId:                  invHost.GetResourceId(),
		HostId:                      invHost.GetResourceId(),
		Name:                        invHost.GetName(),
		DesiredState:                computev1.HostState(invHost.GetDesiredState()),
		CurrentState:                computev1.HostState(invHost.GetCurrentState()),
		SiteId:                      invHost.GetSite().GetResourceId(),
		Site:                        hostSite,
		Note:                        invHost.GetNote(),
		SerialNumber:                invHost.GetSerialNumber(),
		MemoryBytes:                 fmt.Sprintf("%d", invHost.GetMemoryBytes()),
		CpuModel:                    invHost.GetCpuModel(),
		CpuSockets:                  invHost.GetCpuSockets(),
		CpuCores:                    invHost.GetCpuCores(),
		CpuCapabilities:             invHost.GetCpuCapabilities(),
		CpuArchitecture:             invHost.GetCpuArchitecture(),
		CpuThreads:                  invHost.GetCpuThreads(),
		CpuTopology:                 invHost.GetCpuTopology(),
		BmcKind:                     computev1.BaremetalControllerKind(invHost.GetBmcKind()),
		BmcIp:                       invHost.GetBmcIp(),
		Hostname:                    invHost.GetHostname(),
		ProductName:                 invHost.GetProductName(),
		BiosVersion:                 invHost.GetBiosVersion(),
		BiosReleaseDate:             invHost.GetBiosReleaseDate(),
		BiosVendor:                  invHost.GetBiosVendor(),
		HostStatus:                  hostStatus,
		HostStatusIndicator:         hostStatusIndicator,
		HostStatusTimestamp:         hostStatusTimestamp,
		OnboardingStatus:            onboardingStatus,
		OnboardingStatusIndicator:   onboardingStatusIndicator,
		OnboardingStatusTimestamp:   onboardingStatusTimestamp,
		RegistrationStatus:          registrationStatus,
		RegistrationStatusIndicator: registrationStatusIndicator,
		RegistrationStatusTimestamp: registrationStatusTimestamp,
		HostStorages:                fromInvHostStorages(invHost.GetHostStorages()),
		HostNics:                    fromInvHostNics(invHost.GetHostNics()),
		HostUsbs:                    fromInvHostUsbs(invHost.GetHostUsbs()),
		HostGpus:                    fromInvHostGpus(invHost.GetHostGpus()),
		Instance:                    hostInstance,
		Metadata:                    metadata,
		InheritedMetadata:           []*commonv1.MetadataItem{},
		CreatedAt:                   invHost.GetCreatedAt(),
		UpdatedAt:                   invHost.GetUpdatedAt(),
	}

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
			ResourceId:    storage.GetResourceId(),
			Wwid:          storage.GetWwid(),
			Serial:        storage.GetSerial(),
			Vendor:        storage.GetVendor(),
			Model:         storage.GetModel(),
			CapacityBytes: fmt.Sprintf("%d", storage.GetCapacityBytes()),
			DeviceName:    storage.GetDeviceName(),
			CreatedAt:     storage.GetCreatedAt(),
			UpdatedAt:     storage.GetUpdatedAt(),
		})
	}
	return hostStorages
}

func fromInvHostNics(nics []*inv_computev1.HostnicResource) []*computev1.HostnicResource {
	// Conversion logic for HostNics
	hostNics := make([]*computev1.HostnicResource, 0, len(nics))
	for _, nic := range nics {
		hostNics = append(hostNics, &computev1.HostnicResource{
			ResourceId:    nic.GetResourceId(),
			DeviceName:    nic.GetDeviceName(),
			PciIdentifier: nic.GetPciIdentifier(),
			MacAddr:       nic.GetMacAddr(),
			SriovEnabled:  nic.GetSriovEnabled(),
			SriovVfsNum:   nic.GetSriovVfsNum(),
			SriovVfsTotal: nic.GetSriovVfsTotal(),
			Features:      nic.GetFeatures(),
			Mtu:           nic.GetMtu(),
			LinkState:     computev1.NetworkInterfaceLinkState(nic.GetLinkState()),
			BmcInterface:  nic.GetBmcInterface(),
			CreatedAt:     nic.GetCreatedAt(),
			UpdatedAt:     nic.GetUpdatedAt(),
		})
	}
	return hostNics
}

func fromInvHostUsbs(usbs []*inv_computev1.HostusbResource) []*computev1.HostusbResource {
	// Conversion logic for HostUsbs
	hostUsbs := make([]*computev1.HostusbResource, 0, len(usbs))
	for _, usb := range usbs {
		hostUsbs = append(hostUsbs, &computev1.HostusbResource{
			ResourceId: usb.GetResourceId(),
			Idvendor:   usb.GetIdvendor(),
			Idproduct:  usb.GetIdproduct(),
			Bus:        usb.GetBus(),
			Addr:       usb.GetAddr(),
			Class:      usb.GetClass(),
			Serial:     usb.GetSerial(),
			DeviceName: usb.GetDeviceName(),
			CreatedAt:  usb.GetCreatedAt(),
			UpdatedAt:  usb.GetUpdatedAt(),
		})
	}
	return hostUsbs
}

func fromInvHostGpus(gpus []*inv_computev1.HostgpuResource) []*computev1.HostgpuResource {
	// Conversion logic for HostGpus
	hostGpus := make([]*computev1.HostgpuResource, 0, len(gpus))
	for _, gpu := range gpus {
		hostGpus = append(hostGpus, &computev1.HostgpuResource{
			ResourceId:  gpu.GetResourceId(),
			PciId:       gpu.GetPciId(),
			Product:     gpu.GetProduct(),
			Vendor:      gpu.GetVendor(),
			Description: gpu.GetDescription(),
			DeviceName:  gpu.GetDeviceName(),
			Features:    gpu.GetFeatures(),
			CreatedAt:   gpu.GetCreatedAt(),
			UpdatedAt:   gpu.GetUpdatedAt(),
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
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: invHost,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to create inventory resource %s", invRes)
		return nil, err
	}

	hostCreated, err := fromInvHost(invResp.GetHost(), nil)
	if err != nil {
		return nil, err
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

	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, err
	}
	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_Host{Host: &inv_computev1.HostResource{}}},
		Offset:   offset,
		Limit:    limit,
		OrderBy:  req.GetOrderBy(),
		Filter:   req.GetFilter(),
	}

	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to list inventory resources %s", filter)
		return nil, err
	}

	invResources := invResp.GetResources()
	hosts := make([]*computev1.HostResource, 0, len(invResources))
	for _, invRes := range invResources {
		host, err := fromInvHost(invRes.GetResource().GetHost(), invRes.GetRenderedMetadata())
		if err != nil {
			return nil, err
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
	host, err := fromInvHost(invHost, invResp.GetRenderedMetadata())
	if err != nil {
		return nil, err
	}
	zlog.Debug().Msgf("Got %s", host)
	return host, nil
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
		return nil, err
	}

	fieldmask, err := fieldmaskpb.New(invHost, maps.Values(OpenAPIHostToProto)...)
	if err != nil {
		return nil, err
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
	invUpRes, err := fromInvHost(invUp, nil)
	if err != nil {
		return nil, err
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
		return nil, err
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
		return nil, err
	}

	_, err = is.InvClient.Update(ctx, req.GetResourceId(), fm, res)
	if err != nil {
		return nil, err
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
		Name:         req.GetHost().GetName(),
		DesiredState: inv_computev1.HostState_HOST_STATE_REGISTERED,
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

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: hostResource,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to create inventory resource %s", invRes)
		return nil, err
	}

	hostResp, err := fromInvHost(invResp.GetHost(), nil)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fm, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}

	invUp := upRes.GetHost()
	invUpRes, err := fromInvHost(invUp, nil)
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Onboarded %s", invUpRes)
	return &restv1.OnboardHostResponse{}, nil
}

// Onboard a host.
func (is *InventorygRPCServer) RegisterUpdateHost(
	ctx context.Context,
	req *restv1.RegisterHostRequest,
) (*computev1.HostResource, error) {
	zlog.Debug().Msg("RegisterUpdateHost")
	hostResource := &inv_computev1.HostResource{
		Name:         req.GetHost().GetName(),
		DesiredState: inv_computev1.HostState_HOST_STATE_REGISTERED,
	}
	fieldList := []string{inv_computev1.HostResourceFieldName, inv_computev1.HostResourceFieldDesiredState}

	if req.GetHost().GetAutoOnboard() {
		hostResource.DesiredState = inv_computev1.HostState_HOST_STATE_ONBOARDED
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

	invHost, err := fromInvHost(invReply.GetHost(), nil)
	if err != nil {
		return nil, err
	}
	zlog.Debug().Msgf("Updated %s", invHost)
	return invHost, nil
}

func (is *InventorygRPCServer) listAllHosts(ctx context.Context, filter string) ([]*computev1.HostResource, error) {
	var offset int
	var pageSize int32 = 100
	hasNext := true
	hosts := make([]*computev1.HostResource, 0, pageSize)

	for hasNext {
		offsetInt32, err := SafeIntToInt32(offset)
		if err != nil {
			return nil, err
		}
		req := &restv1.ListHostsRequest{
			Filter:   filter,
			PageSize: pageSize,
			Offset:   offsetInt32,
		}
		hostsList, err := is.ListHosts(ctx, req)
		if err != nil {
			return nil, err
		}

		hosts = append(hosts, hostsList.GetHosts()...)
		hasNext = hostsList.GetHasNext()
		offset += len(hostsList.GetHosts())
	}

	return hosts, nil
}

// Get hosts summary.
//
//nolint:cyclop // high cyclomatic complexity due to complex status checking
func (is *InventorygRPCServer) GetHostsSummary(
	ctx context.Context,
	req *restv1.GetHostSummaryRequest,
) (*restv1.GetHostSummaryResponse, error) {
	zlog.Debug().Msg("GetHostsSummary")

	var total uint32
	var errorState uint32
	var runningState uint32
	var unallocatedState uint32

	isFailedHostStatus := func(host *computev1.HostResource) bool {
		hostErr := host.GetHostStatusIndicator() == statusv1.StatusIndication_STATUS_INDICATION_ERROR ||
			host.GetOnboardingStatusIndicator() == statusv1.StatusIndication_STATUS_INDICATION_ERROR

		instanceErr := false
		if host.GetInstance() != nil {
			instanceErr = host.GetInstance().
				GetInstanceStatusIndicator() ==
				statusv1.StatusIndication_STATUS_INDICATION_ERROR ||
				host.GetInstance().GetProvisioningStatusIndicator() == statusv1.StatusIndication_STATUS_INDICATION_ERROR ||
				host.GetInstance().GetUpdateStatusIndicator() == statusv1.StatusIndication_STATUS_INDICATION_ERROR
		}
		return hostErr || instanceErr
	}

	hosts, err := is.listAllHosts(ctx, req.GetFilter())
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		if host.GetSite() == nil {
			unallocatedState++
		}
		if host.GetSite() != nil && host.GetSite().GetResourceId() == "" {
			unallocatedState++
		}

		if isFailedHostStatus(host) {
			errorState++
		}

		// Since IDLE status can be used for multiple status (e.g., Powered off or Invalidated),
		// we use Instance's current state as a source of Running state.
		// To avoid counting hosts in both Running and Error states
		// current state must be RUNNING, but Host/Instance status cannot be a failure status.
		if host.GetInstance().GetCurrentState() == computev1.InstanceState_INSTANCE_STATE_RUNNING &&
			!isFailedHostStatus(host) {
			runningState++
		}
	}

	total, err = SafeIntToUint32(len(hosts))
	if err != nil {
		return nil, err
	}
	// Notice, error and running numbers come from Provider.State
	hostsSummary := &restv1.GetHostSummaryResponse{
		Total:       total,
		Error:       errorState,
		Running:     runningState,
		Unallocated: unallocatedState,
	}

	return hostsSummary, nil
}
