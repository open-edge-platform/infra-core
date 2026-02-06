// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"

	"google.golang.org/grpc/codes"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

func (is *InvStore) ListResources(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*inv_v1.GetResourceResponse, int, error,
) {
	zlog.Debug().Msgf("ListResources: %v", filter)

	resKind := util.GetResourceKindFromResource(filter.GetResource())
	// TODO apply other filters to the resources
	// NOTE we might want to check that the filter we got are applicable for a resource
	mapFindResources := map[inv_v1.ResourceKind]func(context.Context, *inv_v1.ResourceFilter) (
		[]*inv_v1.GetResourceResponse,
		int,
		error,
	){
		inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE:          is.ListInstances,
		inv_v1.ResourceKind_RESOURCE_KIND_HOST:              is.ListHosts,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE:       is.ListHoststorage,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTDEVICE:        is.ListHostdevice,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC:           is.ListHostnics,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB:           is.ListHostusb,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU:           is.ListHostgpus,
		inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT:    is.ListNetworkSegments,
		inv_v1.ResourceKind_RESOURCE_KIND_NETLINK:           is.ListNetlinks,
		inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT:          is.ListEndpoints,
		inv_v1.ResourceKind_RESOURCE_KIND_REGION:            is.ListRegions,
		inv_v1.ResourceKind_RESOURCE_KIND_SITE:              is.ListSites,
		inv_v1.ResourceKind_RESOURCE_KIND_OU:                is.ListOus,
		inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER:          is.ListProviders,
		inv_v1.ResourceKind_RESOURCE_KIND_OS:                is.ListOss,
		inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE:    is.ListSingleSchedules,
		inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE:  is.ListRepeatedSchedules,
		inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP:   is.ListTelemetryGroup,
		inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE: is.ListTelemetryProfile,
		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD:          is.ListWorkload,
		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER:   is.ListWorkloadMember,
		inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS:         is.ListIPAddress,
		inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF:   is.ListRemoteAccessConfig,
		inv_v1.ResourceKind_RESOURCE_KIND_TENANT:            is.ListTenants,
		inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT:      is.ListLocalAccounts,
		inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY:    is.ListOSUpdatePolicies,
		inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG:      is.ListCustomConfigs,
		inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN:       is.ListOSUpdateRuns,
	}

	filterFunc, ok := mapFindResources[resKind]
	if !ok {
		zlog.InfraSec().InfraError("resource kind not found %s", resKind).Msg("")
		return nil, 0, errors.Errorfc(codes.InvalidArgument, "resource kind not found %s", resKind)
	}
	return filterFunc(ctx, filter)
}

func (is *InvStore) FindResources(ctx context.Context, filter *inv_v1.ResourceFilter) (
	[]*client.ResourceTenantIDCarrier, int, error,
) {
	zlog.Debug().Msgf("FindResources: %v", filter)

	resKind := util.GetResourceKindFromResource(filter.GetResource())
	// TODO apply other filters to the resources
	// NOTE we might want to check that the filter we got are applicable for a resource
	mapFindResources := map[inv_v1.ResourceKind]func(context.Context, *inv_v1.ResourceFilter) (
		[]*client.ResourceTenantIDCarrier, int, error,
	){
		inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE:          is.FilterInstances,
		inv_v1.ResourceKind_RESOURCE_KIND_HOST:              is.FilterHosts,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE:       is.FilterHoststorage,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTDEVICE:        is.FilterHostdevice,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC:           is.FilterHostnics,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB:           is.FilterHostusb,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU:           is.FilterHostgpus,
		inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT:    is.FilterNetworkSegments,
		inv_v1.ResourceKind_RESOURCE_KIND_NETLINK:           is.FilterNetlinks,
		inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT:          is.FilterEndpoints,
		inv_v1.ResourceKind_RESOURCE_KIND_REGION:            is.FilterRegions,
		inv_v1.ResourceKind_RESOURCE_KIND_SITE:              is.FilterSites,
		inv_v1.ResourceKind_RESOURCE_KIND_OU:                is.FilterOus,
		inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER:          is.FilterProviders,
		inv_v1.ResourceKind_RESOURCE_KIND_OS:                is.FilterOss,
		inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE:    is.FilterSingleSchedules,
		inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE:  is.FilterRepeatedSchedules,
		inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_GROUP:   is.FilterTelemetryGroup,
		inv_v1.ResourceKind_RESOURCE_KIND_TELEMETRY_PROFILE: is.FilterTelemetryProfile,
		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD:          is.FilterWorkload,
		inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD_MEMBER:   is.FilterWorkloadMember,
		inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS:         is.FilterIPAddress,
		inv_v1.ResourceKind_RESOURCE_KIND_RMT_ACCESS_CONF:   is.FilterRemoteAccessConfig,
		inv_v1.ResourceKind_RESOURCE_KIND_TENANT:            is.FilterTenants,
		inv_v1.ResourceKind_RESOURCE_KIND_LOCALACCOUNT:      is.FilterLocalAccounts,
		inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATEPOLICY:    is.FilterOSUpdatePolicies,
		inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG:      is.FilterCustomConfigs,
		inv_v1.ResourceKind_RESOURCE_KIND_OSUPDATERUN:       is.FilterOSUpdateRuns,
	}

	filterFunc, ok := mapFindResources[resKind]
	if !ok {
		zlog.InfraSec().InfraError("resource kind not found %s", resKind).Msg("")
		return nil, 0, errors.Errorfc(codes.InvalidArgument, "resource kind not found %s", resKind)
	}
	return filterFunc(ctx, filter)
}
