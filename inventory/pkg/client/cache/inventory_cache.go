// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"crypto/sha256"
	"encoding/hex"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_errors "github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

var (
	zlog = logging.GetLogger("InfraAPIClientCache")

	subscribeCacheResourceKinds = []inv_v1.ResourceKind{
		inv_v1.ResourceKind_RESOURCE_KIND_SITE,
		inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER,
		inv_v1.ResourceKind_RESOURCE_KIND_HOST,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTDEVICE,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU,
		inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
		inv_v1.ResourceKind_RESOURCE_KIND_OS,
	}
	subscribeCacheUUIDResourceKinds = []inv_v1.ResourceKind{
		inv_v1.ResourceKind_RESOURCE_KIND_HOST,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTDEVICE,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU,
		inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
	}
)

const cacheFakeTenantID = "fake-tenant"

func (c *InventoryCache) GetCacheSusbcriptionResourceKind() []inv_v1.ResourceKind {
	return subscribeCacheResourceKinds
}

func (c *InventoryCache) GetCacheUUIDSubscriptionResourceKind() []inv_v1.ResourceKind {
	return subscribeCacheUUIDResourceKinds
}

func (c *InventoryCache) StoreResourceByID(
	resource *inv_v1.GetResourceResponse,
) {
	// get the resource type to store in cache.
	switch res := resource.Resource.Resource.(type) {
	case *inv_v1.Resource_Host:
		zlog.Debug().Msgf("store host resource with resource ID <%v>", res.Host.GetResourceId())
		c.storeHostResourceInfo(res.Host)
		zlog.Info().Msgf("ResourceIDCache: store, resourceID=%v", res.Host.GetResourceId())
	case *inv_v1.Resource_Instance:
		zlog.Debug().Msgf("store instance resource with resource ID <%v>", res.Instance.GetResourceId())
		c.storeInstanceResourceInfo(res.Instance)
		zlog.Info().Msgf("ResourceIDCache: store, resourceID=%v", res.Instance.GetResourceId())
	default:
	}
}

func (c *InventoryCache) GetResourceByID(tenantID, resourceID string) (*inv_v1.GetResourceResponse, error) {
	// get resource kind from ID
	if rKind, err := util.GetResourceKindFromResourceID(resourceID); err == nil {
		var (
			resource *inv_v1.GetResourceResponse
			ok       bool
		)
		switch rKind {
		case inv_v1.ResourceKind_RESOURCE_KIND_HOST:
			resource, ok = c.getHostResourceByID(tenantID, resourceID)
		case inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE:
			resource, ok = c.getInstanceResourceByID(tenantID, resourceID)
		default:
			zlog.Info().Msgf("ResourceIDCache: not supported, resourceID=%v", resourceID)
			return nil, inv_errors.Errorfc(codes.InvalidArgument, "resource type not supported <%v> ", rKind)
		}

		if ok {
			zlog.Info().Msgf("ResourceIDCache: hit, resourceID=%v", resourceID)
			return resource, nil
		}
	}
	zlog.Info().Msgf("ResourceIDCache: miss, resourceID=%v", resourceID)
	return nil, inv_errors.Errorfc(codes.NotFound, "cache entry not found for resource Id <%v> ", resourceID)
}

// invalidateCacheUpdateEntryByID invalidates cached info of any given resource and its related resource.
func (c *InventoryCache) InvalidateCacheEntryByID(tenantID, resourceID string) {
	// first check if resource id has any direct resource associated.
	zlog.Debug().Msgf("invalidate resource with resource ID <%v>, tenantID <%v>", resourceID, tenantID)

	if res, ok := c.getCacheEntry(tenantIDBasedKey{tenantID, resourceID}); ok {
		// identify resource kind and trigger delete for relation resources

		// get resource type.
		switch r := res.(type) {
		case *computev1.HostResource:
			c.invalidateHostResourceInfo(r)
		case *computev1.InstanceResource:
			c.invalidateInstanceResourceInfo(r)
		default:
		}
	}

	// for now flush entire cache too
	c.flushResourceCache()
}

// invalidateCacheUpdateEntry invalidates cached info of any given resource and its related resource.
func (c *InventoryCache) InvalidateCacheEntryByResource(r *inv_v1.Resource) {
	tenantID, resourceID, err := util.GetResourceKeyFromResource(r)
	if err != nil {
		// This error should never happen.
		zlog.Err(err).Msgf("Unexpected error")
		return
	}
	// first check if resource id has any direct resource associated.
	c.InvalidateCacheEntryByID(tenantID, resourceID)
	zlog.Info().Msg("ResourceIDCache: invalidate")
}

func (c *InventoryCache) StoreResourceByFilter(filter *inv_v1.ResourceFilter, rsp *inv_v1.ListResourcesResponse) {
	// support only AIP-160 style filtering.
	if filter.Filter == "" {
		return
	}

	// support only Host and Instance resource kind.
	resKind := util.GetResourceKindFromResource(filter.Resource)
	if resKind == inv_v1.ResourceKind_RESOURCE_KIND_HOST ||
		resKind == inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE {
		// derive key from Filter.
		key := makeHashKeyFromFilter(filter)
		zlog.Debug().Msgf("store resource with filter <%v>, hash <%v>", filter.String(), key)

		// make copy and preserve in map.
		rspCopy := proto.Clone(rsp)
		c.storeCacheEntry(tenantIDBasedKey{cacheFakeTenantID, key}, rspCopy)
		zlog.Info().Msg("FilterCache: store")
	}
}

func (c *InventoryCache) GetResourceByFilter(filter *inv_v1.ResourceFilter) (*inv_v1.ListResourcesResponse, error) {
	// support only AIP-160 style filtering.
	if filter.Filter == "" {
		return nil, inv_errors.Errorfc(codes.InvalidArgument, "only AIP-160 style filter caching is supported")
	}

	// support only Host and Instance resource kind.
	resKind := util.GetResourceKindFromResource(filter.Resource)
	if resKind == inv_v1.ResourceKind_RESOURCE_KIND_HOST ||
		resKind == inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE {
		// derive key from Filter.
		key := makeHashKeyFromFilter(filter)
		zlog.Debug().Msgf("get resource with filter <%v>, hash <%v>", filter.String(), key)
		if val, ok := c.getCacheEntry(tenantIDBasedKey{cacheFakeTenantID, key}); ok {
			// make rsp copy and return.
			if rsp, ok := val.(*inv_v1.ListResourcesResponse); ok {
				zlog.Info().Msg("FilterCache: hit")
				zlog.Debug().Msgf("get resource found with filter <%v>, hash <%v>", filter.String(), key)
				return makeFilterResourceRspCopy(rsp), nil
			}
		}
		zlog.Info().Msg("FilterCache: miss")
		return nil, inv_errors.Errorfc(codes.NotFound, "cache entry not found for filter <%s>, hash <%v> ", filter.String(), key)
	}
	zlog.Info().Msg("FilterCache: not supported")
	return nil, inv_errors.Errorfc(codes.InvalidArgument, "resource type not supported <%v> ", resKind)
}

func (c *InventoryCache) DeleteResourceByFilter(filter *inv_v1.ResourceFilter) {
	// derive key from Filter
	zlog.Info().Msg("FilterCache: invalidate")
	key := makeHashKeyFromFilter(filter)
	zlog.Debug().Msgf("delete resource with filter <%v>, hash <%v>", filter.String(), key)
	c.deleteCacheEntry(tenantIDBasedKey{cacheFakeTenantID, key})
}

// makeHashKeyFromFilter generates hash from filter based on following fields.
// Resource, Limit, Offset, OrderBy and API-160 style filter field in string format.
func makeHashKeyFromFilter(filter *inv_v1.ResourceFilter) string {
	// make string of filter fields
	s := filter.String()
	h := sha256.New()
	h.Write([]byte(s))
	hash := h.Sum(nil)
	hashHex := hex.EncodeToString(hash)
	zlog.Debug().Msgf("generate hash from filter <%v> as hash <%v>", filter.String(), hashHex)
	return hashHex
}

func makeFilterResourceRspCopy(resource *inv_v1.ListResourcesResponse) *inv_v1.ListResourcesResponse {
	resourceCopy := proto.Clone(resource)
	if val, ok := resourceCopy.(*inv_v1.ListResourcesResponse); ok {
		return val
	}
	return nil
}
