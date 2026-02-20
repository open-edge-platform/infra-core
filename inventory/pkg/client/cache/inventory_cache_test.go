// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package cache_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client/cache"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

const defaultCacheTTLTest = 30 * time.Second

const (
	tenant1 = "11111111-1111-1111-1111-111111111111"
	tenant2 = "22222222-2222-2222-2222-222222222222"
)

func TestGetCacheSusbcriptionResourceKind(t *testing.T) {
	c := cache.NewInventoryCache(defaultCacheTTLTest, 0)
	evt := c.GetCacheSusbcriptionResourceKind()
	exp := []inv_v1.ResourceKind{
		inv_v1.ResourceKind_RESOURCE_KIND_SITE,
		inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER,
		inv_v1.ResourceKind_RESOURCE_KIND_HOST,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTDEVICE,
		inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
		inv_v1.ResourceKind_RESOURCE_KIND_OS,
	}

	assert.Equal(t, true, reflect.DeepEqual(exp, evt))
}

func TestGetCacheUUIDSusbcriptionResourceKind(t *testing.T) {
	c := cache.NewInventoryCache(defaultCacheTTLTest, 0)
	evt := c.GetCacheUUIDSubscriptionResourceKind()
	exp := []inv_v1.ResourceKind{
		inv_v1.ResourceKind_RESOURCE_KIND_HOST,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTUSB,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTGPU,
		inv_v1.ResourceKind_RESOURCE_KIND_HOSTDEVICE,
		inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
	}

	assert.Equal(t, true, reflect.DeepEqual(exp, evt))
}

func TestHostStoreGetAndInvalidateCacheByID(t *testing.T) {
	// TestStoreGetAndInvalidateResourceByID
	// 1. store dummy host resource in cache.
	// 2. get resource by ID from Cache and verify that exists.
	// 3. invalidate resource entry in cache by ID.
	// 4. verify that resource doesn't exist in cache now.
	// create resources
	hostT1 := createDummyHost("TestHost-1", nil, nil, WithTenantID(tenant1))
	hostT2 := createDummyHost("TestHost-1", nil, nil, WithTenantID(tenant2))
	// Host
	resT1 := &inv_v1.GetResourceResponse{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{
				Host: hostT1.Host,
			},
		},
	}
	resT2 := &inv_v1.GetResourceResponse{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{
				Host: hostT2.Host,
			},
		},
	}

	c := cache.NewInventoryCache(defaultCacheTTLTest, 0)

	// host caching.
	c.StoreResourceByID(resT1)
	c.StoreResourceByID(resT2)

	// fetch the resource for validation.
	rsp1, err1 := c.GetResourceByID(hostT1.Host.TenantId, hostT1.Host.ResourceId)
	assert.NoError(t, err1)
	hostRes1, ok := rsp1.Resource.Resource.(*inv_v1.Resource_Host)
	assert.True(t, ok)
	assert.Equal(t, hostT1.Host.ResourceId, hostRes1.Host.ResourceId)

	rsp2, err2 := c.GetResourceByID(hostT2.Host.TenantId, hostT2.Host.ResourceId)
	assert.NoError(t, err2)
	hostRes2, ok := rsp2.Resource.Resource.(*inv_v1.Resource_Host)
	assert.True(t, ok)
	assert.Equal(t, hostT2.Host.ResourceId, hostRes2.Host.ResourceId)

	// invalidate resources.
	c.InvalidateCacheEntryByID(hostT1.Host.TenantId, hostT1.Host.ResourceId)
	c.InvalidateCacheEntryByID(hostT2.Host.TenantId, hostT2.Host.ResourceId)
}

func TestCacheUUIDStoreGetAndInvalidate(t *testing.T) {
	hostT1 := createDummyHost("TestHost-1", nil, nil, WithTenantID(tenant1))
	hostT2 := createDummyHost("TestHost-1", nil, nil, WithTenantID(tenant2))
	uuidT1 := hostT1.Host.Uuid
	uuidT2 := hostT2.Host.Uuid
	hostUsbT1 := &computev1.HostusbResource{ResourceId: "hostusb-12345678"}
	hostGpuT1 := &computev1.HostgpuResource{ResourceId: "hostgpu-12345678"}
	hostNicT1 := &computev1.HostnicResource{ResourceId: "hostnic-12345678"}
	hostStorageT1 := &computev1.HoststorageResource{ResourceId: "hoststorage-12345678"}
	hostDeviceT1 := &computev1.HostdeviceResource{ResourceId: "hostdevice-12345678"}
	hostT1.Host.HostUsbs = []*computev1.HostusbResource{hostUsbT1}
	hostT1.Host.HostGpus = []*computev1.HostgpuResource{hostGpuT1}
	hostT1.Host.HostNics = []*computev1.HostnicResource{hostNicT1}
	hostT1.Host.HostStorages = []*computev1.HoststorageResource{hostStorageT1}
	hostT1.Host.HostDevice = hostDeviceT1
	hostUsbT2 := &computev1.HostusbResource{ResourceId: "hostusb-87654321"}
	hostGpuT2 := &computev1.HostgpuResource{ResourceId: "hostgpu-87654321"}
	hostNicT2 := &computev1.HostnicResource{ResourceId: "hostnic-87654321"}
	hostStorageT2 := &computev1.HoststorageResource{ResourceId: "hoststorage-87654321"}
	hostDeviceT2 := &computev1.HostdeviceResource{ResourceId: "hostdevice-87654321"}
	hostT2.Host.HostUsbs = []*computev1.HostusbResource{hostUsbT2}
	hostT2.Host.HostGpus = []*computev1.HostgpuResource{hostGpuT2}
	hostT2.Host.HostNics = []*computev1.HostnicResource{hostNicT2}
	hostT2.Host.HostStorages = []*computev1.HoststorageResource{hostStorageT2}
	hostT2.Host.HostDevice = hostDeviceT2

	c := cache.NewInventoryCache(defaultCacheTTLTest, 0)

	// host caching.
	c.StoreHostByUUID(uuidT1, hostT1.Host)
	c.StoreHostByUUID(uuidT2, hostT2.Host)

	// fetch the resource for validation.
	rsp, err := c.GetHostByUUID(tenant1, uuidT1)
	assert.NoError(t, err)
	assert.Equal(t, hostT1.Host.ResourceId, rsp.ResourceId)

	rsp, err = c.GetHostByUUID(tenant2, uuidT2)
	assert.NoError(t, err)
	assert.Equal(t, hostT2.Host.ResourceId, rsp.ResourceId)

	// invalidate resources.
	c.InvalidateCacheUUIDByResourceID(tenant1, hostT1.Host.ResourceId)
	// Invalidated resource should not be there anymore.
	res1, err1 := c.GetHostByUUID(tenant1, uuidT1)
	assert.Nil(t, res1)
	assertCacheNotFound(t, err1)

	// Tenant2 resource should still be there
	rsp, err = c.GetHostByUUID(tenant2, uuidT2)
	assert.NoError(t, err)
	assert.Equal(t, hostT2.Host.ResourceId, rsp.ResourceId)

	// Invalidate and check resource from tenant2.
	c.InvalidateCacheUUIDByResourceID(tenant2, hostT2.Host.ResourceId)
	res2, err2 := c.GetHostByUUID(tenant2, uuidT2)
	assert.Nil(t, res2)
	assertCacheNotFound(t, err2)
}

func TestCacheUUIDInvalidateMultipleKey(t *testing.T) {
	// User dummy resources with overlapping resources ID to ensure that MT isolation is granted in the cache.
	instT1 := createDummyInstance("TestInstance-1", WithTenantID(tenant1))
	instT2 := createDummyInstance("TestInstance-1", WithTenantID(tenant2))

	hostT1 := createDummyHost("TestHost-1", nil, instT1.Instance, WithTenantID(tenant1))
	hostT2 := createDummyHost("TestHost-1", nil, instT2.Instance, WithTenantID(tenant2))
	uuidT1 := hostT1.Host.Uuid
	uuidT2 := hostT2.Host.Uuid
	hostUsb := &computev1.HostusbResource{ResourceId: "hostusb-12345678"}
	hostGpu := &computev1.HostgpuResource{ResourceId: "hostgpu-12345678"}
	hostNic := &computev1.HostnicResource{ResourceId: "hostnic-12345678"}
	hostStorage := &computev1.HoststorageResource{ResourceId: "hoststorage-12345678"}
	hostDevice := &computev1.HostdeviceResource{ResourceId: "hostdevice-12345678"}
	hostT1.Host.HostUsbs = []*computev1.HostusbResource{hostUsb}
	hostT1.Host.HostGpus = []*computev1.HostgpuResource{hostGpu}
	hostT1.Host.HostNics = []*computev1.HostnicResource{hostNic}
	hostT1.Host.HostStorages = []*computev1.HoststorageResource{hostStorage}
	hostT1.Host.HostDevice = hostDevice
	hostT2.Host.HostUsbs = []*computev1.HostusbResource{hostUsb}
	hostT2.Host.HostGpus = []*computev1.HostgpuResource{hostGpu}
	hostT2.Host.HostNics = []*computev1.HostnicResource{hostNic}
	hostT2.Host.HostStorages = []*computev1.HoststorageResource{hostStorage}
	hostT2.Host.HostDevice = hostDevice

	c := cache.NewInventoryCache(defaultCacheTTLTest, 0)

	c.StoreHostByUUID(uuidT2, hostT2.Host)

	testcasesT1 := map[string]struct {
		invalidateResID string
	}{
		"InstanceT1": {
			invalidateResID: instT1.Instance.ResourceId,
		},
		"hostUsbT1": {
			invalidateResID: hostUsb.ResourceId,
		},
		"hostGpuT1": {
			invalidateResID: hostGpu.ResourceId,
		},
		"hostNicT1": {
			invalidateResID: hostNic.ResourceId,
		},
		"hostStorageT1": {
			invalidateResID: hostStorage.ResourceId,
		},
		"hostDeviceT1": {
			invalidateResID: hostDevice.ResourceId,
		},
	}
	for tcname, tc := range testcasesT1 {
		t.Run(tcname, func(t *testing.T) {
			c.StoreHostByUUID(uuidT1, hostT1.Host)

			c.InvalidateCacheUUIDByResourceID(tenant1, tc.invalidateResID)

			res1, err1 := c.GetHostByUUID(tenant1, uuidT1)
			assert.Nil(t, res1)
			assertCacheNotFound(t, err1)

			// Resource from other tenant should be untouched in the cache
			res2, err2 := c.GetHostByUUID(tenant2, uuidT2)
			assert.NoError(t, err2)
			assert.Equal(t, hostT2.Host.ResourceId, res2.ResourceId)
		})
	}
}

func TestGetHostResourceIdFromSubRes(t *testing.T) {
	inst := createDummyInstance("TestInstance-1")
	hr := createDummyHost("TestHost-1", nil, nil)
	inst.Instance.Host = hr.Host

	hostUsb := &computev1.HostusbResource{
		ResourceId: "hostusb-12345678",
		Host:       hr.Host,
	}
	hostGpu := &computev1.HostgpuResource{
		ResourceId: "hostgpu-12345678",
		Host:       hr.Host,
	}
	hostNic := &computev1.HostnicResource{
		ResourceId: "hostnic-12345678",
		Host:       hr.Host,
	}
	hostStorage := &computev1.HoststorageResource{
		ResourceId: "hoststorage-12345678",
		Host:       hr.Host,
	}
	hostDevice := &computev1.HostdeviceResource{
		ResourceId: "hostdevice-12345678",
		Host:       hr.Host,
	}

	c := cache.NewInventoryCache(defaultCacheTTLTest, 0)

	testcases := map[string]struct {
		res *inv_v1.Resource
	}{
		"Host": {
			res: &inv_v1.Resource{Resource: hr},
		},
		"Instance": {
			res: &inv_v1.Resource{Resource: inst},
		},
		"hostUsb": {
			res: &inv_v1.Resource{Resource: &inv_v1.Resource_Hostusb{Hostusb: hostUsb}},
		},
		"hostGpu": {
			res: &inv_v1.Resource{Resource: &inv_v1.Resource_Hostgpu{Hostgpu: hostGpu}},
		},
		"hostNic": {
			res: &inv_v1.Resource{Resource: &inv_v1.Resource_Hostnic{Hostnic: hostNic}},
		},
		"hostStorage": {
			res: &inv_v1.Resource{Resource: &inv_v1.Resource_Hoststorage{Hoststorage: hostStorage}},
		},
		"hostDevice": {
			res: &inv_v1.Resource{Resource: &inv_v1.Resource_HostDevice{HostDevice: hostDevice}},
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			resID := c.GetHostResourceIDFromSubRes(tc.res)
			assert.Equal(t, hr.Host.ResourceId, resID)
		})
	}
}

func TestInstanceStoreGetAndInvalidateCacheByID(t *testing.T) {
	// TestStoreGetAndInvalidateResourceByID
	// 1. store dummy instance resource in cache.
	// 2. get resource by ID from Cache and verify that exists.
	// 3. invalidate resource entry in cache by ID.
	// 4. verify that resource doesn't exist in cache now.
	// create resources
	hostT1 := createDummyHost("TestHost-1", nil, nil, WithTenantID(tenant1))
	hostT2 := createDummyHost("TestHost-1", nil, nil, WithTenantID(tenant2))
	instT1 := createDummyInstance("TestInstance-1", WithTenantID(tenant1))
	instT2 := createDummyInstance("TestInstance-1", WithTenantID(tenant2))
	instT1.Instance.Host = hostT1.Host // Manually add backref to host
	instT2.Instance.Host = hostT2.Host // Manually add backref to host

	// instance
	resT1 := &inv_v1.GetResourceResponse{
		Resource: &inv_v1.Resource{
			Resource: instT1,
		},
	}
	resT2 := &inv_v1.GetResourceResponse{
		Resource: &inv_v1.Resource{
			Resource: instT2,
		},
	}

	c := cache.NewInventoryCache(defaultCacheTTLTest, 0)

	// instance caching.
	c.StoreResourceByID(resT1)
	c.StoreResourceByID(resT2)

	// fetch the resource for validation.
	rsp1, err1 := c.GetResourceByID(tenant1, instT1.Instance.ResourceId)
	assert.NoError(t, err1)
	hostRes1, ok := rsp1.Resource.Resource.(*inv_v1.Resource_Instance)
	assert.True(t, ok)
	assert.Equal(t, instT1.Instance.ResourceId, hostRes1.Instance.ResourceId)
	rsp2, err2 := c.GetResourceByID(tenant2, instT2.Instance.ResourceId)
	assert.NoError(t, err2)
	hostRes2, ok := rsp2.Resource.Resource.(*inv_v1.Resource_Instance)
	assert.True(t, ok)
	assert.Equal(t, instT2.Instance.ResourceId, hostRes2.Instance.ResourceId)

	// invalidate resource from tenant1 (for now we flush the whole cache)
	c.InvalidateCacheEntryByID(tenant1, instT1.Instance.ResourceId)

	// fetch the resource for validation- shouldn't exist anymore.
	res, err := c.GetResourceByID(tenant1, instT1.Instance.ResourceId)
	assert.Nil(t, res)
	assertCacheNotFound(t, err)
	res, err = c.GetResourceByID(tenant2, instT2.Instance.ResourceId)
	assert.Nil(t, res)
	assertCacheNotFound(t, err)
}

func TestStoreGetAndInvalidateCacheByResource(t *testing.T) {
	// TestStoreGetAndInvalidateCacheByResource
	// 1. store dummy host resource in cache.
	// 2. get resource by ID from Cache and verify that exists.
	// 3. invalidate resource entry in cache by resource this time.
	// 4. verify that resource doesn't exist in cache now.
	// create resources
	siteT1 := createDummySite("TestSite-2", nil, WithTenantID(tenant1))
	siteT2 := createDummySite("TestSite-2", nil, WithTenantID(tenant2))
	instT1 := createDummyInstance("TestInstance-2", WithTenantID(tenant1))
	instT2 := createDummyInstance("TestInstance-2", WithTenantID(tenant2))

	hostT1 := createDummyHost("TestHost-1", siteT1, instT1.Instance, WithTenantID(tenant1))
	hostT2 := createDummyHost("TestHost-1", siteT2, instT2.Instance, WithTenantID(tenant2))
	// Host
	resT1 := &inv_v1.GetResourceResponse{
		Resource: &inv_v1.Resource{
			Resource: hostT1,
		},
	}
	resT2 := &inv_v1.GetResourceResponse{
		Resource: &inv_v1.Resource{
			Resource: hostT2,
		},
	}
	c := cache.NewInventoryCache(defaultCacheTTLTest, 0)

	// host caching.
	c.StoreResourceByID(resT1)
	c.StoreResourceByID(resT2)

	// fetch the resource for validation.
	rsp, err := c.GetResourceByID(tenant1, hostT1.Host.ResourceId)
	assert.NoError(t, err)
	hostRes1, ok := rsp.Resource.Resource.(*inv_v1.Resource_Host)
	assert.True(t, ok)
	assert.Equal(t, hostT1.Host.ResourceId, hostRes1.Host.ResourceId)
	rsp, err = c.GetResourceByID(tenant2, hostT2.Host.ResourceId)
	assert.NoError(t, err)
	hostRes2, ok := rsp.Resource.Resource.(*inv_v1.Resource_Host)
	assert.True(t, ok)
	assert.Equal(t, hostT2.Host.ResourceId, hostRes2.Host.ResourceId)

	// invalidate resource from tenant1. For now we flush the whole cache.
	c.InvalidateCacheEntryByResource(resT1.Resource)
	// fetch the resource for validation- shouldn't exist anymore.
	res2, err2 := c.GetResourceByID(tenant1, hostT1.Host.ResourceId)
	assert.Nil(t, res2)
	assertCacheNotFound(t, err2)

	res2, err2 = c.GetResourceByID(tenant2, hostT1.Host.ResourceId)
	assert.Nil(t, res2)
	assertCacheNotFound(t, err2)
}

func TestStoreGetAndInvalidateCacheByTimeout(t *testing.T) {
	// TestStoreGetAndInvalidateCacheByTimeout.
	// 1. Update cache with stale entry timeout to 1 sec.
	// 2. store dummy host resource in cache.
	// 3. add delay of 1 + 1(random delay) so that cache entry expires.
	// 4. verify that resource doesn't exist in cache now.
	// create resources
	region := createDummyRegion("TestRegion-3")
	site := createDummySite("TestSite-3", region)
	inst := createDummyInstance("TestInstance-3")

	hr := createDummyHost("TestHost-3", site, inst.Instance)
	// Host
	res := &inv_v1.GetResourceResponse{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Host{
				Host: hr.Host,
			},
		},
	}

	// update cache config with cache entry stale time to 1 sec.
	c := cache.NewInventoryCache(1*time.Second, 10)
	// validate that stale time is 1 sec now.
	assert.Equal(t, 1, int(c.StaleTime().Seconds()))
	assert.Equal(t, 10, c.StateTimeOffset())

	// host caching.
	c.StoreResourceByID(res)

	// sleep for 1 sec + 10% of 1 sec(max random) and then fetch cache entry.
	// - should have expired after timer expiry.
	time.Sleep(2 * time.Second)

	// fetch the resource for validation- shouldn't exist anymore.
	res, err := c.GetResourceByID(hr.Host.TenantId, hr.Host.ResourceId)
	assert.Nil(t, res)
	assertCacheNotFound(t, err)
}

func TestCacheResourceByFilter(t *testing.T) {
	// make resource with filter.
	region := createDummyRegion("TestRegion-3")
	site := createDummySite("TestSite-3", region)
	inst := createDummyInstance("TestInstance-3")
	hr := createDummyHost("TestHost-3", site, inst.Instance)

	// define filter for search.
	host := &inv_v1.Resource_Host{
		Host: &computev1.HostResource{
			Uuid: hr.Host.Uuid,
		},
	}

	filter := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{Resource: host},
		Limit:    1,
		Offset:   1,
		Filter:   `site.region.name = "foo"`,
	}

	result := &inv_v1.ListResourcesResponse{
		Resources: []*inv_v1.GetResourceResponse{
			{
				Resource: &inv_v1.Resource{
					Resource: hr,
				},
			},
		},
		TotalElements: 1,
	}

	// store the result.
	c := cache.NewInventoryCache(5*time.Second, 0)
	c.StoreResourceByFilter(filter, result)

	// get result back.
	rsp, err := c.GetResourceByFilter(filter)
	require.Equal(t, nil, err, "get resource by filter error")
	require.NotNil(t, rsp, "get resource by filter response is nil")
	resource, ok := rsp.Resources[0].Resource.Resource.(*inv_v1.Resource_Host)
	require.Equal(t, true, ok)
	require.Equal(t, hr.Host.Uuid, resource.Host.Uuid)

	// delete resource.
	c.DeleteResourceByFilter(filter)
	// get result back.
	rsp1, err1 := c.GetResourceByFilter(filter)
	require.Nil(t, rsp1, "get resource by filter response should be nil")
	assertCacheNotFound(t, err1)

	// Invalid filter
	filter.Filter = ""
	_, err2 := c.GetResourceByFilter(filter)
	assertCacheInvalidArgument(t, err2, "only AIP-160 style filter caching is supported")
	// store by invalid filter is no op
	c.StoreResourceByFilter(filter, nil)
}

func TestCacheResourceByFilterMismatch(t *testing.T) {
	// make resource with filter.
	region := createDummyRegion("TestRegion-4")
	site := createDummySite("TestSite-4", region)
	inst := createDummyInstance("TestInstance-4")
	hr := createDummyHost("TestHost-4", site, inst.Instance)

	// define filter for search.
	host := &inv_v1.Resource_Host{
		Host: &computev1.HostResource{
			Uuid: hr.Host.Uuid,
		},
	}

	filter := &inv_v1.ResourceFilter{
		Resource: &inv_v1.Resource{Resource: host},
		Limit:    1,
		Offset:   1,
		Filter:   `site.region.name = "foo"`,
	}

	result := &inv_v1.ListResourcesResponse{
		Resources: []*inv_v1.GetResourceResponse{
			{
				Resource: &inv_v1.Resource{
					Resource: hr,
				},
			},
		},
		TotalElements: 1,
	}

	// store the result.
	c := cache.NewInventoryCache(5*time.Second, 0)
	c.StoreResourceByFilter(filter, result)

	// Negative tests
	testcases := map[string]struct {
		filter *inv_v1.ResourceFilter
	}{
		"DifferentLimit": {
			filter: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: host},
				Limit:    2,
				Offset:   1,
				Filter:   `site.region.name = "foo"`,
			},
		},
		"DifferentOffset": {
			filter: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: host},
				Limit:    1,
				Offset:   2,
				Filter:   `site.region.name = "foo"`,
			},
		},
		"DifferentFilter": {
			filter: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{Resource: host},
				Limit:    1,
				Offset:   1,
				Filter:   `site.region.name = "change"`,
			},
		},
		"DifferentResourceContent": {
			filter: &inv_v1.ResourceFilter{
				Resource: &inv_v1.Resource{
					Resource: &inv_v1.Resource_Host{
						Host: &computev1.HostResource{
							ResourceId: "host-11abcabc",
						},
					},
				},
				Limit:  1,
				Offset: 1,
				Filter: `site.region.name = "change"`,
			},
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			rsp, err := c.GetResourceByFilter(tc.filter)
			require.Nil(t, rsp, "get resource by filter response should be nil")
			assertCacheNotFound(t, err)
		})
	}
}

func TestUnsupportedResourceCache(t *testing.T) {
	c := cache.NewInventoryCache(5*time.Second, 0)

	for _, val := range inv_v1.ResourceKind_value {
		kind := inv_v1.ResourceKind(val)
		if kind == inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED ||
			kind == inv_v1.ResourceKind_RESOURCE_KIND_HOST ||
			kind == inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE {
			// Skip cached resources
			continue
		}

		res, err := util.GetResourceFromKind(kind)
		require.NoError(t, err)
		t.Run(string(util.ResourceKindToPrefix(kind)), func(t *testing.T) {
			filter := &inv_v1.ResourceFilter{
				Resource: res,
				Filter:   `randomFilter = "test"`,
			}
			rsp, err := c.GetResourceByFilter(filter)
			require.Nil(t, rsp, "get resource by filter response should be nil")
			assertCacheInvalidArgument(t, err, "resource type not supported")
		})
	}
}

func assertCacheNotFound(t *testing.T, err error) {
	t.Helper()
	s := status.Convert(err)
	require.NotNil(t, s)
	assert.Equal(t, codes.NotFound, s.Code())
	assert.Contains(t, s.Message(), "cache entry not found")
}

func assertCacheInvalidArgument(t *testing.T, err error, expErrMsg string) {
	t.Helper()
	s := status.Convert(err)
	require.NotNil(t, s)
	assert.Equal(t, codes.InvalidArgument, s.Code())
	assert.Contains(t, s.Message(), expErrMsg)
}
