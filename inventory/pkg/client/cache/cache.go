// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"math/rand"
	"sync"
	"time"
)

// default cache entry timeout. Used for go-flags.
const (
	DefaultCacheStaleTime             time.Duration = 30 * time.Second
	DefaultCacheStaleTimeOffsetFactor               = 10
)

// Inventory Cache.
// Cache shall hold single map for storing both,
// resourceId and Filter based results.
// e.g. hostID:HostResource for resourceId based caching for Inventory client GET methods.
// e.g. filter-hash: ResourceList result for filter based caching for Inventory client List method.
type InventoryCache struct {
	staleTime       time.Duration
	stateTimeOffset int
	resourceMap     sync.Map
	// Used by UUID-cache only
	reverseMap sync.Map   // Map used to store reverse map from UUID, Host[usb|gpu|storage|nic] ID, Instance ID to HostID
	mu         sync.Mutex // Mutex to access reverseMap
}

// Cache Entry
// This is the entry store in Resource Map as key: Entry{Resource, timestamp}.
type Entry struct {
	Resource any
	Validity time.Time
}

func NewInventoryCache(staleTime time.Duration, staleTimeOffset int) *InventoryCache {
	// init cache if not already done.
	invCache := &InventoryCache{
		staleTime:       DefaultCacheStaleTime,
		stateTimeOffset: DefaultCacheStaleTimeOffsetFactor,
	}

	if staleTime > 0 {
		invCache.staleTime = staleTime
	}
	if staleTimeOffset > 0 {
		invCache.stateTimeOffset = staleTimeOffset
	}

	zlog.Info().Msgf("resource cache configured with stale time <%v> seconds", invCache.staleTime)
	return invCache
}

func (c *InventoryCache) StaleTime() time.Duration {
	return c.staleTime
}

func (c *InventoryCache) StateTimeOffset() int {
	return c.stateTimeOffset
}

func (c *InventoryCache) UpdateStaleTime(staleTime time.Duration) {
	if c.staleTime > 0 {
		c.staleTime = staleTime
	}
	zlog.Info().Msgf("resource cache updated with stale time <%v> seconds", c.staleTime)
}

func (c *InventoryCache) generateCacheEntryStaleTime() time.Duration {
	// have random extension by "cacheStaleTimeOffsetFactor"% of stale time (positive and negative).
	randomTimeOffset := int(c.staleTime.Seconds() / float64(c.stateTimeOffset))

	// if randomTimeOffset is 0 then make offset to 1.
	if randomTimeOffset == 0 {
		randomTimeOffset = 1
	}
	//nolint:gosec,mnd // Not necessary to use secure random generator, also using 2 to have positive and negative offsets.
	return time.Duration(int(c.staleTime) + ((rand.Intn(randomTimeOffset*2) - randomTimeOffset) * int(time.Second)))
}

// isCacheEntryStale check if cache entry is stale.
func (e Entry) isCacheEntryStale() bool {
	return e.Validity.Before(time.Now())
}

// StoreResourceByID stores key: resource mapping.
func (c *InventoryCache) storeCacheEntry(key tenantIDBasedKey, res any) {
	// generate cache entry stale time.
	staleTime := c.generateCacheEntryStaleTime()
	e := Entry{
		Resource: res,
		Validity: time.Now().Add(staleTime),
	}
	c.resourceMap.Store(key, e)
	zlog.Debug().Msgf("cached resource with key <%v>, stale time <%v>, validity <%v>",
		key, staleTime, e.Validity)
}

// GetResourceByResID gets key:resource mapping with stale timestamp check.
func (c *InventoryCache) getCacheEntry(key tenantIDBasedKey) (any, bool) {
	if e, ok := c.resourceMap.Load(key); ok {
		zlog.Debug().Msgf("cached entry found with key <%v>", key)
		entry, isEntry := e.(Entry)
		if !isEntry {
			zlog.Error().Msgf("unexpected type for Entry: %T", e)
			return nil, false
		}
		// check time-stamp
		if !entry.isCacheEntryStale() {
			return entry.Resource, ok
		}

		zlog.Error().Msgf("cached entry found stale with key <%v>", key)
		// delete stale entry
		c.resourceMap.Delete(key)
		return nil, false
	}
	zlog.Debug().Msgf("cached entry not found with key <%v>", key)
	return nil, false
}

// DeleteResourceByResID deletes resource by key.
func (c *InventoryCache) deleteCacheEntry(key tenantIDBasedKey) {
	zlog.Debug().Msgf("deleted cache entry with key <%v>", key)
	c.resourceMap.Delete(key)
}

// loadAndDeleteCacheEntry deletes resource by key and return the stored value.
func (c *InventoryCache) loadAndDeleteCacheEntry(key tenantIDBasedKey) (any, bool) {
	zlog.Debug().Msgf("deleted cache entry with key <%v>", key)
	val, ok := c.resourceMap.LoadAndDelete(key)
	if !ok {
		return nil, ok
	}
	res, ok := val.(Entry)
	if !ok {
		return nil, false
	}
	return res.Resource, true
}

// FlushResourceCache clears all cached resources.
func (c *InventoryCache) flushResourceCache() {
	zlog.Debug().Msgf("flushing full resource cache")
	c.resourceMap.Range(func(key, _ any) bool {
		c.resourceMap.Delete(key)
		return true
	})
	zlog.Info().Msg("Cache: flush")
}
