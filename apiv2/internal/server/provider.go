// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"

	providerv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/provider/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/validator"
)

// OpenAPIProviderToProto maps OpenAPI fields name to Proto fields name.
// The key is derived from the json property respectively of the
// structs Provider defined in edge-infra-manager-openapi-types.gen.go.
var OpenAPIProviderToProto = map[string]string{
	providerv1.ProviderResourceFieldProviderKind:   inv_providerv1.ProviderResourceFieldProviderKind,
	providerv1.ProviderResourceFieldProviderVendor: inv_providerv1.ProviderResourceFieldProviderVendor,
	providerv1.ProviderResourceFieldName:           inv_providerv1.ProviderResourceFieldName,
	providerv1.ProviderResourceFieldApiEndpoint:    inv_providerv1.ProviderResourceFieldApiEndpoint,
	providerv1.ProviderResourceFieldApiCredentials: inv_providerv1.ProviderResourceFieldApiCredentials,
	providerv1.ProviderResourceFieldConfig:         inv_providerv1.ProviderResourceFieldConfig,
}

func toInvProvider(provider *providerv1.ProviderResource) (*inv_providerv1.ProviderResource, error) {
	if provider == nil {
		return &inv_providerv1.ProviderResource{}, nil
	}
	invProvider := &inv_providerv1.ProviderResource{
		ProviderKind:   inv_providerv1.ProviderKind(provider.GetProviderKind()),
		ProviderVendor: inv_providerv1.ProviderVendor(provider.GetProviderVendor()),
		Name:           provider.GetName(),
		ApiEndpoint:    provider.GetApiEndpoint(),
		ApiCredentials: provider.GetApiCredentials(),
		Config:         provider.GetConfig(),
	}

	err := validator.ValidateMessage(invProvider)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to validate inventory resource")
		return nil, err
	}
	return invProvider, nil
}

func fromInvProvider(invProvider *inv_providerv1.ProviderResource) *providerv1.ProviderResource {
	if invProvider == nil {
		return &providerv1.ProviderResource{}
	}
	provider := &providerv1.ProviderResource{
		ResourceId:     invProvider.GetResourceId(),
		ProviderKind:   providerv1.ProviderKind(invProvider.GetProviderKind()),
		ProviderVendor: providerv1.ProviderVendor(invProvider.GetProviderVendor().Number()),
		Name:           invProvider.GetName(),
		ApiEndpoint:    invProvider.GetApiEndpoint(),
		ApiCredentials: invProvider.GetApiCredentials(),
		Config:         invProvider.GetConfig(),
		ProviderID:     invProvider.GetResourceId(),
		Timestamps:     GrpcToOpenAPITimestamps(invProvider),
	}

	return provider
}

func (is *InventorygRPCServer) CreateProvider(
	ctx context.Context,
	req *restv1.CreateProviderRequest,
) (*providerv1.ProviderResource, error) {
	zlog.Debug().Msg("CreateProvider")

	provider := req.GetProvider()
	invProvider, err := toInvProvider(provider)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to convert to inventory provider")
		return nil, err
	}

	invRes := &inventory.Resource{
		Resource: &inventory.Resource_Provider{
			Provider: invProvider,
		},
	}

	invResp, err := is.InvClient.Create(ctx, invRes)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to create provider in inventory")
		return nil, err
	}

	providerCreated := fromInvProvider(invResp.GetProvider())
	zlog.Debug().Msgf("Created %s", providerCreated)
	return providerCreated, nil
}

// Get a list of providers.
func (is *InventorygRPCServer) ListProviders(
	ctx context.Context,
	req *restv1.ListProvidersRequest,
) (*restv1.ListProvidersResponse, error) {
	zlog.Debug().Msg("ListProviders")
	offset, limit, err := parsePagination(req.GetOffset(), req.GetPageSize())
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse pagination %d %d", req.GetOffset(), req.GetPageSize())
		return nil, err
	}
	filter := &inventory.ResourceFilter{
		Resource: &inventory.Resource{Resource: &inventory.Resource_Provider{Provider: &inv_providerv1.ProviderResource{}}},
		Offset:   offset,
		Limit:    limit,
		OrderBy:  req.GetOrderBy(),
		Filter:   req.GetFilter(),
	}
	if err := validator.ValidateMessage(filter); err != nil {
		zlog.InfraSec().InfraErr(err).Msg("failed to validate query params")
		return nil, errors.Wrap(err)
	}
	invResp, err := is.InvClient.List(ctx, filter)
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to list providers from inventory")
		return nil, err
	}

	providers := []*providerv1.ProviderResource{}
	for _, invRes := range invResp.GetResources() {
		provider := fromInvProvider(invRes.GetResource().GetProvider())
		providers = append(providers, provider)
	}

	resp := &restv1.ListProvidersResponse{
		Providers:     providers,
		TotalElements: invResp.GetTotalElements(),
		HasNext:       invResp.GetHasNext(),
	}
	zlog.Debug().Msgf("Listed %s", resp)
	return resp, nil
}

// Get a specific provider.
func (is *InventorygRPCServer) GetProvider(
	ctx context.Context,
	req *restv1.GetProviderRequest,
) (*providerv1.ProviderResource, error) {
	zlog.Debug().Msg("GetProvider")

	invResp, err := is.InvClient.Get(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to get provider from inventory")
		return nil, err
	}

	invProvider := invResp.GetResource().GetProvider()
	provider := fromInvProvider(invProvider)
	zlog.Debug().Msgf("Got %s", provider)
	return provider, nil
}

// Delete a provider.
func (is *InventorygRPCServer) DeleteProvider(
	ctx context.Context,
	req *restv1.DeleteProviderRequest,
) (*restv1.DeleteProviderResponse, error) {
	zlog.Debug().Msg("DeleteProvider")

	_, err := is.InvClient.Delete(ctx, req.GetResourceId())
	if err != nil {
		zlog.InfraErr(err).Msg("Failed to delete provider from inventory")
		return nil, err
	}
	zlog.Debug().Msgf("Deleted %s", req.GetResourceId())
	return &restv1.DeleteProviderResponse{}, nil
}
