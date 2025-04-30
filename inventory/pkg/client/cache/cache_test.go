// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package cache_test

import (
	"flag"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	cache "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client/cache"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

const fakeTestTenant = "anyTenant"

func TestMain(m *testing.M) {
	// Only needed to suppress the error
	flag.String(
		"policyBundle",
		"/rego/policy_bundle.tar.gz",
		"Path of policy rego file",
	)
	flag.Parse()

	run := m.Run() // run all tests
	os.Exit(run)
}

type Options struct {
	tenantID string
}

type Option func(*Options)

func WithTenantID(tenantID string) Option {
	return func(options *Options) {
		options.tenantID = tenantID
	}
}

// parseOptions parses the given list of Option into an Options.
func parseOptions(options ...Option) *Options {
	opts := &Options{tenantID: fakeTestTenant}
	for _, option := range options {
		option(opts)
	}
	return opts
}

func createDummyRegion(name string, opts ...Option) *location_v1.RegionResource {
	options := parseOptions(opts...)
	region := &location_v1.RegionResource{
		TenantId:     options.tenantID,
		Name:         name,
		RegionKind:   "Test-Region",
		ParentRegion: nil,
		Metadata:     "",
	}
	return region
}

func createDummySite(name string, region *location_v1.RegionResource, opts ...Option) *location_v1.SiteResource {
	options := parseOptions(opts...)
	site := &location_v1.SiteResource{
		TenantId:         options.tenantID,
		Name:             name,
		Region:           region,
		Ou:               nil,
		Address:          "",
		SiteLat:          10,
		SiteLng:          20,
		DnsServers:       []string{},
		DockerRegistries: []string{},
		MetricsEndpoint:  "",
		HttpProxy:        "",
		HttpsProxy:       "",
		FtpProxy:         "",
		NoProxy:          "",
		Metadata:         "",
	}

	return site
}

func createDummyInstance(name string, opts ...Option) *inv_v1.Resource_Instance {
	options := parseOptions(opts...)
	return &inv_v1.Resource_Instance{
		Instance: &computev1.InstanceResource{
			TenantId:   options.tenantID,
			Name:       name,
			ResourceId: "inst-11aabbcc",
		},
	}
}

func createDummyHost(
	name string,
	site *location_v1.SiteResource,
	inst *computev1.InstanceResource,
	opts ...Option,
) *inv_v1.Resource_Host {
	options := parseOptions(opts...)
	return &inv_v1.Resource_Host{
		Host: &computev1.HostResource{
			TenantId:     options.tenantID,
			Name:         name,
			ResourceId:   "host-11abcabc",
			Uuid:         uuid.NewString(),
			DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,

			Site:         site,
			Provider:     nil,
			Instance:     inst,
			HardwareKind: "XDgen4",
			SerialNumber: "1001",
			MemoryBytes:  64 * util.Gigabyte,

			CpuModel:        "12th Gen Intel(R) Core(TM) i9-12900",
			CpuSockets:      1,
			CpuCores:        14,
			CpuCapabilities: "",
			CpuArchitecture: "x86_64",
			CpuThreads:      13,

			MgmtIp: "192.168.10.13",

			BmcKind:     computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
			BmcIp:       "10.0.0.13",
			BmcUsername: "user",
			BmcPassword: "pass",
			PxeMac:      "90:49:fa:ff:ff:f3",

			Hostname: "testhost",
			Metadata: `[{"key":"key1-test","value":"host_key1-test"}]`,
		},
	}
}

func TestInitClientCache(t *testing.T) {
	c := cache.NewInventoryCache(0, 0)
	assert.Equal(t, cache.DefaultCacheStaleTime, c.StaleTime(),
		"cache should have default stale time of %d", cache.DefaultCacheStaleTime)
	assert.Equalf(t, cache.DefaultCacheStaleTimeOffsetFactor, c.StateTimeOffset(),
		"cache should have default stale time offset of %d", cache.DefaultCacheStaleTimeOffsetFactor)
}
