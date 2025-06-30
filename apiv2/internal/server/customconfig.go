// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	customconfigv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/customconfig/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

func toInvCustomConfig(customConfig *customconfigv1.CustomConfigResource) (*inv_computev1.CustomConfigResource, error) {
	if customConfig == nil {
		return &inv_computev1.CustomConfigResource{}, nil
	}

	invCustomConfig := &inv_computev1.CustomConfigResource{
		Name:        customConfig.GetName(),
		Description: customConfig.GetDescription(),
		Config:      customConfig.GetConfig(),
	}

	err := validator.ValidateMessage(invCustomConfig)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory custom config resource")
		return nil, err
	}

	return invCustomConfig, nil
}

func fromInvCustomConfig(invCustomConfig *inv_computev1.CustomConfigResource) *customconfigv1.CustomConfigResource {
	if invCustomConfig == nil {
		return &customconfigv1.CustomConfigResource{}
	}

	return &customconfigv1.CustomConfigResource{
		ResourceId:  invCustomConfig.GetResourceId(),
		Name:        invCustomConfig.GetName(),
		Description: invCustomConfig.GetDescription(),
		Config:      invCustomConfig.GetConfig(),
		Timestamps:  GrpcToOpenAPITimestamps(invCustomConfig),
	}
}

func (is *InventorygRPCServer) CreateCustomConfig(
	ctx context.Context,
	req *restv1.CreateCustomConfigRequest,
) (*customconfigv1.CustomConfigResource, error) {
	zlog.Debug().Msg("CreateCustomConfig")

	invCustomConfig, err := toInvCustomConfig(req.GetCustomConfig())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory custom config")
		return nil, errors.Wrap(err)
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_CustomConfig{
			CustomConfig: invCustomConfig,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create custom config in inventory")
		return nil, errors.Wrap(err)
	}

	customConfigCreated := fromInvCustomConfig(invResp.GetCustomConfig())
	zlog.Debug().Msgf("Created %s", customConfigCreated)
	return customConfigCreated, nil
}

func (is *InventorygRPCServer) ListCustomConfigs(
	ctx context.Context,
	req *restv1.ListCustomConfigsRequest,
) (*restv1.ListCustomConfigsResponse, error) {
	zlog.Debug().Msg("ListCustomConfigs")

	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_CustomConfig{
			CustomConfig: &inv_computev1.CustomConfigResource{},
		}},
		Offset:  req.GetOffset(),
		Limit:   req.GetPageSize(),
		OrderBy: req.GetOrderBy(),
		Filter:  req.GetFilter(),
	}

	if err := validator.ValidateMessage(filter); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("failed to validate query params")
		return nil, errors.Wrap(err)
	}

	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list custom configs from inventory")
		return nil, errors.Wrap(err)
	}

	customConfigs := []*customconfigv1.CustomConfigResource{}
	for _, invRes := range invResp.GetResources() {
		cc := fromInvCustomConfig(invRes.GetResource().GetCustomConfig())
		customConfigs = append(customConfigs, cc)
	}

	resp := &restv1.ListCustomConfigsResponse{
		CustomConfigs: customConfigs,
		TotalElements: invResp.GetTotalElements(),
		HasNext:       invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

func (is *InventorygRPCServer) GetCustomConfig(
	ctx context.Context,
	req *restv1.GetCustomConfigRequest,
) (*customconfigv1.CustomConfigResource, error) {
	zlog.Debug().Msg("GetCustomConfig")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get custom config from inventory")
		return nil, errors.Wrap(err)
	}

	invCustomConfig := invResp.GetResource().GetCustomConfig()
	customConfig := fromInvCustomConfig(invCustomConfig)
	zlog.Debug().Msgf("Got %s", customConfig)
	return customConfig, nil
}

func (is *InventorygRPCServer) DeleteCustomConfig(
	ctx context.Context,
	req *restv1.DeleteCustomConfigRequest,
) (*restv1.DeleteCustomConfigResponse, error) {
	zlog.Debug().Msg("DeleteCustomConfig")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete custom config from inventory")
		return nil, errors.Wrap(err)
	}

	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteCustomConfigResponse{}, nil
}
