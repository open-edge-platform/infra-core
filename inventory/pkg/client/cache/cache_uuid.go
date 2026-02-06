// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_errors "github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

type tenantIDBasedKey struct {
	tenantID string
	key      string
}

// StoreHostByUUID caches the given Host resource by the given UUID. Also stores backward mapping from resourceID to UUID.
func (c *InventoryCache) StoreHostByUUID(uuid string, host *computev1.HostResource) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hostID := host.GetResourceId()
	tenantID := host.GetTenantId()
	// The actual hostID to hostResource is stored in the cache
	c.storeHostResourceInfo(host)

	storedIDs := make([]string, 0)

	// We store the following mapping in the reverse map:
	// - uuid -> hostID
	// - hostUsbID -> hostID
	// - hostGpuID -> hostID
	// - hostStorageId -> hostID
	// - hostNic -> hostID
	c.reverseMap.Store(tenantIDBasedKey{tenantID, uuid}, hostID)
	if instanceID := host.GetInstance().GetResourceId(); instanceID != "" {
		c.reverseMap.Store(tenantIDBasedKey{tenantID, instanceID}, hostID)
		storedIDs = append(storedIDs, instanceID)
	}
	for _, hostUsb := range host.GetHostUsbs() {
		if usbID := hostUsb.GetResourceId(); usbID != "" {
			c.reverseMap.Store(tenantIDBasedKey{tenantID, usbID}, hostID)
			storedIDs = append(storedIDs, usbID)
		}
	}
	for _, hostGpu := range host.GetHostGpus() {
		if gpuID := hostGpu.GetResourceId(); gpuID != "" {
			c.reverseMap.Store(tenantIDBasedKey{tenantID, gpuID}, hostID)
			storedIDs = append(storedIDs, gpuID)
		}
	}
	for _, hostStorage := range host.GetHostStorages() {
		if storageID := hostStorage.GetResourceId(); storageID != "" {
			c.reverseMap.Store(tenantIDBasedKey{tenantID, storageID}, hostID)
			storedIDs = append(storedIDs, storageID)
		}
	}
	for _, hostNic := range host.GetHostNics() {
		if nicID := hostNic.GetResourceId(); nicID != "" {
			c.reverseMap.Store(tenantIDBasedKey{tenantID, nicID}, hostID)
			storedIDs = append(storedIDs, nicID)
			// TODO: Consider storing also IPs
		}
	}
	zlog.Info().Msgf("UUIDCache: store, uuid=%v, tenantID=%v", uuid, tenantID)
	zlog.Debug().Msgf("UUIDCache: reverse map stored (tenantID=%v) IDs=%v", tenantID, storedIDs)
}

func (c *InventoryCache) GetHostByUUID(tenantID, uuid string) (*computev1.HostResource, error) {
	var host *computev1.HostResource
	notFoundErr := inv_errors.Errorfc(codes.NotFound, "cache entry not found for UUID=%v, tenantID=%v", uuid, tenantID)
	if val, ok := c.reverseMap.Load(tenantIDBasedKey{tenantID, uuid}); !ok {
		// miss in the reverse Map, ok
		zlog.Info().Msgf("UUIDCache: miss, uuid=%v, tenantID=%v", uuid, tenantID)
		return nil, notFoundErr
	} else if hostID, ok := val.(string); !ok {
		zlog.Error().Msgf("UUIDCache: unexpected, UUID cached value is not a string")
		return nil, notFoundErr
	} else if val, ok = c.getCacheEntry(tenantIDBasedKey{tenantID, hostID}); !ok {
		// Miss in the cache: WHY is this happening? We maybe should cleanup reverse map?
		// All errors below this are bad errors. How should we tackle them?
		zlog.Error().Msgf("UUIDCache: unexpected, value is not in the actual cache")
		return nil, notFoundErr
	} else if tHost, ok := val.(*computev1.HostResource); !ok {
		zlog.Error().Msgf("UUIDCache: unexpected, Host cached value is not a Host")
		return nil, notFoundErr
	} else if host, ok = proto.Clone(tHost).(*computev1.HostResource); !ok {
		zlog.Error().Msgf("UUIDCache: unexpected, Host cached value cannot be copied")
		return nil, notFoundErr
	}
	zlog.Info().Msgf("UUIDCache: hit, uuid=%v, tenantID=%v", uuid, tenantID)
	return host, nil
}

//nolint:cyclop // high complexity due to nested ifs for error checking
func (c *InventoryCache) invalidateHostByHostID(tenantID, hostID string) {
	var host *computev1.HostResource
	if val, ok := c.loadAndDeleteCacheEntry(tenantIDBasedKey{tenantID, hostID}); !ok {
		return
	} else if host, ok = val.(*computev1.HostResource); !ok {
		return
	}

	uuid := host.GetUuid()
	c.reverseMap.Delete(tenantIDBasedKey{tenantID, uuid})

	if instanceID := host.GetInstance().GetResourceId(); instanceID != "" {
		c.reverseMap.Delete(tenantIDBasedKey{tenantID, instanceID})
	}
	for _, hostUsb := range host.GetHostUsbs() {
		if usbID := hostUsb.GetResourceId(); usbID != "" {
			c.reverseMap.Delete(tenantIDBasedKey{tenantID, usbID})
		}
	}
	for _, hostGpu := range host.GetHostGpus() {
		if gpuID := hostGpu.GetResourceId(); gpuID != "" {
			c.reverseMap.Delete(tenantIDBasedKey{tenantID, gpuID})
		}
	}
	for _, hostStorage := range host.GetHostStorages() {
		if storageID := hostStorage.GetResourceId(); storageID != "" {
			c.reverseMap.Delete(tenantIDBasedKey{tenantID, storageID})
		}
	}
	for _, hostNic := range host.GetHostNics() {
		if nicID := hostNic.GetResourceId(); nicID != "" {
			c.reverseMap.Delete(tenantIDBasedKey{tenantID, nicID})
			// TODO: Consider deleting also IPs
		}
	}
}

func (c *InventoryCache) InvalidateCacheUUIDByResourceID(tenantID, resourceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	resKind, err := util.GetResourceKindFromResourceID(resourceID)
	if err != nil {
		// Error here should never happen
		zlog.Err(err).Msgf("This error should never happen: resourceID=%s", resourceID)
		return
	}
	var hostID string
	switch resKind {
	case inv_v1.ResourceKind_RESOURCE_KIND_HOST:
		hostID = resourceID
	case inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTDEVICE,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB:
		if v, ok := c.reverseMap.Load(tenantIDBasedKey{tenantID, resourceID}); !ok {
			return
		} else if hostID, ok = v.(string); !ok {
			return
		}
	default:
		return
	}
	if hostID != "" {
		zlog.Info().Msgf("UUIDCache: invalidate, resourceID=%v, hostID=%v, tenantID=%v", resourceID, hostID, tenantID)
		c.invalidateHostByHostID(tenantID, hostID)
	}
}

// InvalidateCacheByEvent does smart invalidation based on the received event. If the given event is a create or update
// and the given resource contains eager loaded resource, we manage invalidation via the content of it, otherwise we
// proceed with invalidation by resource ID via InvalidateCacheUUIDByResourceID.
func (c *InventoryCache) InvalidateCacheByEvent(evKind inv_v1.SubscribeEventsResponse_EventKind, resource *inv_v1.Resource) {
	if evKind == inv_v1.SubscribeEventsResponse_EVENT_KIND_UNSPECIFIED {
		// unspecified is never expected
		return
	}

	tenantID, resourceID, err := util.GetResourceKeyFromResource(resource)
	if err != nil {
		// Error here should never happen
		zlog.Err(err).Msgf("This error should never happen: resource=%v", resource)
		return
	}

	if evKind == inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED {
		// If event is deleted, cleanup cache, no need to do smart invalidation.
		c.InvalidateCacheUUIDByResourceID(tenantID, resourceID)
		return
	}

	// If Updated or Created smart invalidation, we need to consider that Instance or sub-resources could be associated
	// to a cached host, and the reverseMap could not contain the expected info, so we need to invalidate the entry
	// by checking also the content of the resource itself that we got in the event.
	if hostID := c.GetHostResourceIDFromSubRes(resource); hostID != "" {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.invalidateHostByHostID(tenantID, hostID)
	} else {
		// Fallback to standard invalidation
		c.InvalidateCacheUUIDByResourceID(tenantID, resourceID)
	}
}

func (c *InventoryCache) GetHostResourceIDFromSubRes(res *inv_v1.Resource) string {
	switch util.GetResourceKindFromResource(res) {
	case inv_v1.ResourceKind_RESOURCE_KIND_HOST:
		return res.GetHost().GetResourceId()
	case inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE:
		return res.GetInstance().GetHost().GetResourceId()
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB:
		return res.GetHostusb().GetHost().GetResourceId()
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU:
		return res.GetHostgpu().GetHost().GetResourceId()
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE:
		return res.GetHoststorage().GetHost().GetResourceId()
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTDEVICE:
		return res.GetHostdevice().GetHost().GetResourceId()
	case inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC:
		// TODO: how to manage IPs?
		return res.GetHostnic().GetHost().GetResourceId()
	default:
		return ""
	}
}
