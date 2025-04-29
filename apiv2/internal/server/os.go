// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	osv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/os/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPIOSResourceToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs OSResource defined in edge-infra-manager-openapi-types.gen.go.
var OpenAPIOSResourceToProto = map[string]string{
	osv1.OperatingSystemResourceFieldName:              inv_osv1.OperatingSystemResourceFieldName,
	osv1.OperatingSystemResourceFieldArchitecture:      inv_osv1.OperatingSystemResourceFieldArchitecture,
	osv1.OperatingSystemResourceFieldKernelCommand:     inv_osv1.OperatingSystemResourceFieldKernelCommand,
	osv1.OperatingSystemResourceFieldUpdateSources:     inv_osv1.OperatingSystemResourceFieldUpdateSources,
	osv1.OperatingSystemResourceFieldInstalledPackages: inv_osv1.OperatingSystemResourceFieldInstalledPackages,
}

func toInvOSResource(osResource *osv1.OperatingSystemResource) (*inv_osv1.OperatingSystemResource, error) {
	if osResource == nil {
		return &inv_osv1.OperatingSystemResource{}, nil
	}
	invOSResource := &inv_osv1.OperatingSystemResource{
		Name:              osResource.GetName(),
		Architecture:      osResource.GetArchitecture(),
		KernelCommand:     osResource.GetKernelCommand(),
		UpdateSources:     osResource.GetUpdateSources(),
		ImageUrl:          osResource.GetImageUrl(),
		ImageId:           osResource.GetImageId(),
		Sha256:            osResource.GetSha256(),
		ProfileName:       osResource.GetProfileName(),
		ProfileVersion:    osResource.GetProfileVersion(),
		InstalledPackages: osResource.GetInstalledPackages(),
		SecurityFeature:   inv_osv1.SecurityFeature(osResource.GetSecurityFeature()),
		OsType:            inv_osv1.OsType(osResource.GetOsType()),
		OsProvider:        inv_osv1.OsProviderKind(osResource.GetOsProvider()),
	}

	err := validator.ValidateMessage(invOSResource)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}

	return invOSResource, nil
}

func fromInvOSResource(invOSResource *inv_osv1.OperatingSystemResource) *osv1.OperatingSystemResource {
	if invOSResource == nil {
		return &osv1.OperatingSystemResource{}
	}
	osResource := &osv1.OperatingSystemResource{
		ResourceId:        invOSResource.GetResourceId(),
		Name:              invOSResource.GetName(),
		Architecture:      invOSResource.GetArchitecture(),
		KernelCommand:     invOSResource.GetKernelCommand(),
		UpdateSources:     invOSResource.GetUpdateSources(),
		ImageUrl:          invOSResource.GetImageUrl(),
		ImageId:           invOSResource.GetImageId(),
		Sha256:            invOSResource.GetSha256(),
		ProfileName:       invOSResource.GetProfileName(),
		ProfileVersion:    invOSResource.GetProfileVersion(),
		InstalledPackages: invOSResource.GetInstalledPackages(),
		SecurityFeature:   osv1.SecurityFeature(invOSResource.GetSecurityFeature()),
		OsType:            osv1.OsType(invOSResource.GetOsType()),
		OsProvider:        osv1.OsProviderKind(invOSResource.GetOsProvider()),
		OsResourceID:      invOSResource.GetResourceId(),
		Timestamps:        GrpcToOpenAPITimestamps(invOSResource),
		PlatformBundle:    invOSResource.GetPlatformBundle(),
	}

	return osResource
}

func (is *InventorygRPCServer) CreateOperatingSystem(
	ctx context.Context,
	req *restv1.CreateOperatingSystemRequest,
) (*osv1.OperatingSystemResource, error) {
	zlog.Debug().Msg("CreateOSResource")

	osResource := req.GetOs()
	invOSResource, err := toInvOSResource(osResource)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory OS resource")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Os{
			Os: invOSResource,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create OS resource in inventory")
		return nil, err
	}

	osResourceCreated := fromInvOSResource(invResp.GetOs())
	zlog.Debug().Msgf("Created %s", osResourceCreated)
	return osResourceCreated, nil
}

// Get a list of osResources.
func (is *InventorygRPCServer) ListOperatingSystems(
	ctx context.Context,
	req *restv1.ListOperatingSystemsRequest,
) (*restv1.ListOperatingSystemsResponse, error) {
	zlog.Debug().Msg("ListOSResources")
	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, err
	}
	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_Os{Os: &inv_osv1.OperatingSystemResource{}}},
		Offset:   offset,
		Limit:    limit,
		OrderBy:  req.GetOrderBy(),
		Filter:   req.GetFilter(),
	}

	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list OS resources from inventory")
		return nil, err
	}

	osResources := []*osv1.OperatingSystemResource{}
	for _, invRes := range invResp.GetResources() {
		osResource := fromInvOSResource(invRes.GetResource().GetOs())
		osResources = append(osResources, osResource)
	}

	resp := &restv1.ListOperatingSystemsResponse{
		OperatingSystemResources: osResources,
		TotalElements:            invResp.GetTotalElements(),
		HasNext:                  invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific osResource.
func (is *InventorygRPCServer) GetOperatingSystem(
	ctx context.Context,
	req *restv1.GetOperatingSystemRequest,
) (*osv1.OperatingSystemResource, error) {
	zlog.Debug().Msg("GetOSResource")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get OS resource from inventory")
		return nil, err
	}

	invOSResource := invResp.GetResource().GetOs()
	osResource := fromInvOSResource(invOSResource)
	zlog.Debug().Msgf("Got %s", osResource)
	return osResource, nil
}

// Update a osResource. (PUT).
func (is *InventorygRPCServer) UpdateOperatingSystem(
	ctx context.Context,
	req *restv1.UpdateOperatingSystemRequest,
) (*osv1.OperatingSystemResource, error) {
	zlog.Debug().Msg("UpdateOSResource")

	osResource := req.GetOs()
	invOSResource, err := toInvOSResource(osResource)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory OS resource")
		return nil, err
	}

	fieldmask, err := fieldmaskpb.New(invOSResource, maps.Values(OpenAPIOSResourceToProto)...)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create field mask")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Os{
			Os: invOSResource,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}
	invUp := upRes.GetOs()
	invUpRes := fromInvOSResource(invUp)
	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Update a osResource. (PATCH).
func (is *InventorygRPCServer) PatchOperatingSystem(
	ctx context.Context,
	req *restv1.PatchOperatingSystemRequest,
) (*osv1.OperatingSystemResource, error) {
	zlog.Debug().Msg("PatchOperatingSystem")

	osResource := req.GetOs()
	invOSResource, err := toInvOSResource(osResource)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory OS resource")
		return nil, err
	}

	fieldmask, err := parseFielmask(invOSResource, req.GetFieldMask(), OpenAPIOSResourceToProto)
	if err != nil {
		return nil, err
	}
	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Os{
			Os: invOSResource,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}
	invUp := upRes.GetOs()
	invUpRes := fromInvOSResource(invUp)
	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Delete a osResource.
func (is *InventorygRPCServer) DeleteOperatingSystem(
	ctx context.Context,
	req *restv1.DeleteOperatingSystemRequest,
) (*restv1.DeleteOperatingSystemResponse, error) {
	zlog.Debug().Msg("DeleteOSResource")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete OS resource from inventory")
		return nil, err
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteOperatingSystemResponse{}, nil
}
