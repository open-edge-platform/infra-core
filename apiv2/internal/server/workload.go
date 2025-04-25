// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPIWorkloadToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs Workload defined in edge-infra-manager-openapi-types.gen.go.
var OpenAPIWorkloadToProto = map[string]string{
	"Kind":       inv_computev1.WorkloadResourceFieldKind,
	"Name":       inv_computev1.WorkloadResourceFieldName,
	"Status":     inv_computev1.WorkloadResourceFieldStatus,
	"ExternalId": inv_computev1.WorkloadResourceFieldExternalId,
}

var OpenAPIWorkloadObjectsNames = map[string]struct{}{
	"Members": {},
}

func toInvWorkload(workload *computev1.WorkloadResource) (*inv_computev1.WorkloadResource, error) {
	if workload == nil {
		return &inv_computev1.WorkloadResource{}, nil
	}
	invWorkload := &inv_computev1.WorkloadResource{
		Kind:         inv_computev1.WorkloadKind(workload.GetKind()),
		Name:         workload.GetName(),
		ExternalId:   workload.GetExternalId(),
		Status:       workload.GetStatus(),
		DesiredState: inv_computev1.WorkloadState_WORKLOAD_STATE_PROVISIONED,
	}

	err := validator.ValidateMessage(invWorkload)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}
	return invWorkload, nil
}

func fromInvWorkload(invWorkload *inv_computev1.WorkloadResource) (*computev1.WorkloadResource, error) {
	if invWorkload == nil {
		return &computev1.WorkloadResource{}, nil
	}
	members, err := fromInvWorkloadMembers(invWorkload.GetMembers())
	if err != nil {
		return nil, err
	}

	workload := &computev1.WorkloadResource{
		ResourceId: invWorkload.GetResourceId(),
		WorkloadId: invWorkload.GetResourceId(),
		Kind:       computev1.WorkloadKind(invWorkload.GetKind()),
		Name:       invWorkload.GetName(),
		ExternalId: invWorkload.GetExternalId(),
		Status:     invWorkload.GetStatus(),
		Members:    members,
		CreatedAt:  invWorkload.GetCreatedAt(),
		UpdatedAt:  invWorkload.GetUpdatedAt(),
	}

	return workload, nil
}

func fromInvWorkloadMembers(members []*inv_computev1.WorkloadMember) ([]*computev1.WorkloadMember, error) {
	// Conversion logic for WorkloadMembers
	workloadMembers := make([]*computev1.WorkloadMember, 0, len(members))
	for _, member := range members {
		workloadMember, err := fromInvWorkloadMember(member)
		if err != nil {
			return nil, err
		}
		workloadMembers = append(workloadMembers, workloadMember)
	}
	return workloadMembers, nil
}

func (is *InventorygRPCServer) CreateWorkload(
	ctx context.Context,
	req *restv1.CreateWorkloadRequest,
) (*computev1.WorkloadResource, error) {
	zlog.Debug().Msg("CreateWorkload")

	workload := req.GetWorkload()
	invWorkload, err := toInvWorkload(workload)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory workload")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Workload{
			Workload: invWorkload,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create workload in inventory")
		return nil, err
	}

	workloadCreated, err := fromInvWorkload(invResp.GetWorkload())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory workload")
		return nil, err
	}

	zlog.Debug().Msgf("Created %s", workloadCreated)
	return workloadCreated, nil
}

// Get a list of workloads.
func (is *InventorygRPCServer) ListWorkloads(
	ctx context.Context,
	req *restv1.ListWorkloadsRequest,
) (*restv1.ListWorkloadsResponse, error) {
	zlog.Debug().Msg("ListWorkloads")
	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, err
	}
	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_Workload{Workload: &inv_computev1.WorkloadResource{}}},
		Offset:   offset,
		Limit:    limit,
		OrderBy:  req.GetOrderBy(),
		Filter:   req.GetFilter(),
	}

	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list workloads from inventory")
		return nil, err
	}

	workloads := []*computev1.WorkloadResource{}
	for _, invRes := range invResp.GetResources() {
		workload, err := fromInvWorkload(invRes.GetResource().GetWorkload())
		if err != nil {
			zlog.InfraErr(err).Msg("Failed to convert from inventory workload")
			return nil, err
		}
		workloads = append(workloads, workload)
	}

	resp := &restv1.ListWorkloadsResponse{
		Workloads:     workloads,
		TotalElements: invResp.GetTotalElements(),
		HasNext:       invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific workload.
func (is *InventorygRPCServer) GetWorkload(
	ctx context.Context,
	req *restv1.GetWorkloadRequest,
) (*computev1.WorkloadResource, error) {
	zlog.Debug().Msg("GetWorkload")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get workload from inventory")
		return nil, err
	}

	invWorkload := invResp.GetResource().GetWorkload()
	workload, err := fromInvWorkload(invWorkload)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert from inventory workload")
		return nil, err
	}
	zlog.Debug().Msgf("Got %s", workload)
	return workload, nil
}

// Update a workload. (PUT).
func (is *InventorygRPCServer) UpdateWorkload(
	ctx context.Context,
	req *restv1.UpdateWorkloadRequest,
) (*computev1.WorkloadResource, error) {
	zlog.Debug().Msg("UpdateWorkload")

	workload := req.GetWorkload()
	invWorkload, err := toInvWorkload(workload)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory workload")
		return nil, err
	}

	fieldmask, err := fieldmaskpb.New(invWorkload, maps.Values(OpenAPIWorkloadToProto)...)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create field mask")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Workload{
			Workload: invWorkload,
		},
	}
	upRes, err := is.InvClient.Update(ctx, req.GetResourceId(), fieldmask, invRes)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to update inventory resource %s %s", req.GetResourceId(), invRes)
		return nil, err
	}
	invUp := upRes.GetWorkload()
	invUpRes, err := fromInvWorkload(invUp)
	if err != nil {
		return nil, err
	}

	zlog.Debug().Msgf("Updated %s", invUpRes)
	return invUpRes, nil
}

// Delete a workload.
func (is *InventorygRPCServer) DeleteWorkload(
	ctx context.Context,
	req *restv1.DeleteWorkloadRequest,
) (*restv1.DeleteWorkloadResponse, error) {
	zlog.Debug().Msg("DeleteWorkload")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete workload from inventory")
		return nil, err
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteWorkloadResponse{}, nil
}
