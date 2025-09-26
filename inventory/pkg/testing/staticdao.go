// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	localaccount_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/localaccount/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	networkv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	ouv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/ou/v1"
	provider_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/remoteaccess/v1"
	schedulev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	telemetryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/function"
)

var (
	once        sync.Once
	daoInstance *InvResourceDAO
)

func getInvResourceDAO() *InvResourceDAO {
	once.Do(
		func() {
			c, e := NewInvResourceDAO()
			if e != nil {
				panic(e)
			}
			daoInstance = c
		},
	)
	return daoInstance
}

func DeleteResource(tb testing.TB, resourceID string) {
	tb.Helper()
	getInvResourceDAO().DeleteResource(tb, client.FakeTenantID, resourceID)
}

func DeleteResourceAndReturnError(tb testing.TB, resourceID string) error {
	tb.Helper()
	return getInvResourceDAO().DeleteResourceAndReturnError(tb, client.FakeTenantID, resourceID)
}

// HardDeleteHost - hard deletes the given host via 2-phase deletion.
func HardDeleteHost(tb testing.TB, resourceID string) {
	tb.Helper()
	getInvResourceDAO().HardDeleteHost(tb, client.FakeTenantID, resourceID)
}

// HardDeleteHostAndReturnError - hard deletes the given host via 2-phase deletion.
func HardDeleteHostAndReturnError(tb testing.TB, resourceID string) error {
	tb.Helper()
	return getInvResourceDAO().HardDeleteHostAndReturnError(tb, client.FakeTenantID, resourceID)
}

// HardDeleteIPAddress - hard delete is done without explicit delete. IPAddresses are removed without an explicit desired state.
func HardDeleteIPAddress(tb testing.TB, resourceID string) {
	tb.Helper()
	getInvResourceDAO().HardDeleteIPAddress(tb, client.FakeTenantID, resourceID)
}

func HardDeleteIPAddressAndReturnError(tb testing.TB, resourceID string) error {
	tb.Helper()
	return getInvResourceDAO().HardDeleteIPAddressAndReturnError(tb, client.FakeTenantID, resourceID)
}

// HardDeleteInstance - hard deletes the given VM via 2-phase deletion.
func HardDeleteInstance(tb testing.TB, resourceID string) {
	tb.Helper()
	getInvResourceDAO().HardDeleteInstance(tb, client.FakeTenantID, resourceID)
}

func HardDeleteInstanceAndReturnError(tb testing.TB, resourceID string) error {
	tb.Helper()
	return getInvResourceDAO().HardDeleteInstanceAndReturnError(tb, client.FakeTenantID, resourceID)
}

// CreateHostWithMetadata - create host with meta. Note this helper is not really meant to be used for the
// test of HostResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateHostWithMetadata(
	tb testing.TB,
	metadata string,
	site *location_v1.SiteResource,
	provider *provider_v1.ProviderResource,
	doCleanup bool,
) (host *computev1.HostResource) {
	tb.Helper()
	return getInvResourceDAO().CreateHostWithOpts(
		tb,
		client.FakeTenantID,
		doCleanup,
		HostMetadata(metadata),
		HostSite(site),
		HostProvider(provider),
	)
}

func CreateHostWithArgs(
	tb testing.TB,
	hostname, hostUUID, serialNumber, md string,
	site *location_v1.SiteResource,
	provider *provider_v1.ProviderResource,
	doCleanup bool,
) *computev1.HostResource {
	tb.Helper()
	return getInvResourceDAO().CreateHostWithOpts(
		tb,
		client.FakeTenantID,
		doCleanup,
		HostHostName(hostname),
		HostUUID(hostUUID),
		HostSerialNumber(serialNumber),
		HostMetadata(md),
		HostSite(site),
		HostProvider(provider),
	)
}

func CreateHost(
	tb testing.TB,
	site *location_v1.SiteResource,
	provider *provider_v1.ProviderResource,
) *computev1.HostResource {
	tb.Helper()
	return getInvResourceDAO().CreateHost(
		tb,
		client.FakeTenantID,
		HostSite(site),
		HostProvider(provider),
	)
}

func CreateHostNoCleanup(
	tb testing.TB,
	site *location_v1.SiteResource,
	provider *provider_v1.ProviderResource,
) *computev1.HostResource {
	tb.Helper()
	return getInvResourceDAO().CreateHostNoCleanup(
		tb, client.FakeTenantID,
		HostSite(site),
		HostProvider(provider),
	)
}

// CreateHostNic - creates host nic. Note this helper is not really meant to be used for the
// test of HostNicResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateHostNic(tb testing.TB, host *computev1.HostResource) *computev1.HostnicResource {
	tb.Helper()
	return getInvResourceDAO().CreateHostNic(tb, client.FakeTenantID, host)
}

func CreateHostNicNoCleanup(tb testing.TB, host *computev1.HostResource) (nic *computev1.HostnicResource) {
	tb.Helper()
	return getInvResourceDAO().CreateHostNicNoCleanup(tb, client.FakeTenantID, host)
}

func CreatHostGPU(tb testing.TB, host *computev1.HostResource) (gpu *computev1.HostgpuResource) {
	tb.Helper()
	return getInvResourceDAO().CreateHostGPU(tb, client.FakeTenantID, host)
}

func CreatHostGPUNoCleanup(tb testing.TB, host *computev1.HostResource) (gpu *computev1.HostgpuResource) {
	tb.Helper()
	return getInvResourceDAO().CreateHostGPUNoCleanup(tb, client.FakeTenantID, host)
}

// CreateIPAddress - creates IPAddress. Note this helper is not really meant to be used for the
// test of IPAddressResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateIPAddress(tb testing.TB, hostNic *computev1.HostnicResource, cleanup bool) *networkv1.IPAddressResource {
	tb.Helper()
	return getInvResourceDAO().CreateIPAddress(tb, client.FakeTenantID, hostNic, cleanup)
}

// CreateRepeatedSchedule - creates repeated schedule. Note this helper is not really meant to be used for the
// test of RepeatedScheduleResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateRepeatedSchedule(
	tb testing.TB,
	host *computev1.HostResource,
	site *location_v1.SiteResource,
	scheduleStatus schedulev1.ScheduleStatus,
	opts ...Opt[schedulev1.RepeatedScheduleResource],
) *schedulev1.RepeatedScheduleResource {
	tb.Helper()

	opts = append(opts, RSRTargetHost(host), RSRTargetSite(site), RSRStatus(scheduleStatus))
	return getInvResourceDAO().CreateRepeatedSchedule(tb, client.FakeTenantID, opts...)
}

func CreateRepeatedScheduleNoCleaup(
	tb testing.TB,
	host *computev1.HostResource,
	site *location_v1.SiteResource,
	scheduleStatus schedulev1.ScheduleStatus,
	opts ...Opt[schedulev1.RepeatedScheduleResource],
) *schedulev1.RepeatedScheduleResource {
	tb.Helper()
	opts = append(opts, RSRTargetHost(host), RSRTargetSite(site), RSRStatus(scheduleStatus))
	return getInvResourceDAO().CreateRepeatedScheduleNoCleanup(tb, client.FakeTenantID, opts...)
}

// CreateSingleSchedule - creates single schedule. Note this helper is not really meant to be used for the
// test of SingleScheduleResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateSingleSchedule(
	tb testing.TB,
	host *computev1.HostResource,
	site *location_v1.SiteResource,
	scheduleStatus schedulev1.ScheduleStatus,
	opts ...Opt[schedulev1.SingleScheduleResource],
) *schedulev1.SingleScheduleResource {
	tb.Helper()
	opts = append(opts, SSRTargetHost(host), SSRTargetSite(site), SSRStatus(scheduleStatus))
	return getInvResourceDAO().CreateSingleSchedule(tb, client.FakeTenantID, opts...)
}

func CreateSingleScheduleNoCleanup(
	tb testing.TB,
	host *computev1.HostResource,
	site *location_v1.SiteResource,
	scheduleStatus schedulev1.ScheduleStatus,
	opts ...Opt[schedulev1.SingleScheduleResource],
) *schedulev1.SingleScheduleResource {
	tb.Helper()
	opts = append(opts, SSRTargetHost(host), SSRTargetSite(site), SSRStatus(scheduleStatus))
	return getInvResourceDAO().CreateSingleScheduleNoCleanup(tb, client.FakeTenantID, opts...)
}

func CreateTelemetryGroupMetrics(tb testing.TB, cleanup bool) *telemetryv1.TelemetryGroupResource {
	tb.Helper()
	return getInvResourceDAO().CreateTelemetryGroupMetrics(tb, client.FakeTenantID, cleanup)
}

func CreateTelemetryGroupLogs(tb testing.TB, cleanup bool) *telemetryv1.TelemetryGroupResource {
	tb.Helper()
	return getInvResourceDAO().CreateTelemetryGroupLogs(tb, client.FakeTenantID, cleanup)
}

func CreateTelemetryProfile(
	tb testing.TB,
	instance *computev1.InstanceResource,
	site *location_v1.SiteResource,
	region *location_v1.RegionResource,
	group *telemetryv1.TelemetryGroupResource,
	cleanup bool,
) *telemetryv1.TelemetryProfile {
	tb.Helper()

	notNilTargets := collections.Filter([]any{instance, site, region}, function.Not(function.IsNil))
	require.Lenf(tb, notNilTargets, 1,
		"Only one target for TelemetryProfile is allowed, provided: [%v, %v, %v]", instance, site, region)

	var target TelemetryProfileTargetConfigurator
	if instance != nil {
		target = TelemetryProfileTarget(instance)
	}
	if region != nil {
		target = TelemetryProfileTarget(region)
	}
	if site != nil {
		target = TelemetryProfileTarget(site)
	}

	return getInvResourceDAO().CreateTelemetryProfile(tb, client.FakeTenantID, target, group, cleanup)
}

func CreateOsWithArgs(
	tb testing.TB,
	sha256Hex, name, profileName string,
	feature osv1.SecurityFeature,
	osType osv1.OsType,
) *osv1.OperatingSystemResource {
	tb.Helper()
	return getInvResourceDAO().CreateOsWithArgs(tb, client.FakeTenantID, sha256Hex, name, profileName, feature, osType)
}

func CreateOsWithOpts(
	tb testing.TB,
	doCleanup bool,
	opts ...Opt[osv1.OperatingSystemResource],
) *osv1.OperatingSystemResource {
	tb.Helper()
	return getInvResourceDAO().CreateOsWithOpts(tb, client.FakeTenantID, doCleanup, opts...)
}

// CreateOs creates mutable OSResource by default. Use CreateOsWithArgs to customize OS type, if needed.
func CreateOs(tb testing.TB) *osv1.OperatingSystemResource {
	tb.Helper()
	return getInvResourceDAO().CreateOs(tb, client.FakeTenantID)
}

// CreateOsNoCleanup creates mutable OSResource by default (with no cleanup).
// Use CreateOsWithArgs to customize OS type, if needed.
func CreateOsNoCleanup(tb testing.TB) (osr *osv1.OperatingSystemResource) {
	tb.Helper()
	return getInvResourceDAO().CreateOsNoCleanup(tb, client.FakeTenantID)
}

func CreateSite(tb testing.TB, region *location_v1.RegionResource, ou *ouv1.OuResource) *location_v1.SiteResource {
	tb.Helper()
	return getInvResourceDAO().CreateSite(tb, client.FakeTenantID, SiteRegion(region), SiteOu(ou))
}

// CreateSiteWithMeta - creates site with metadata. Note this helper is not really meant to be used for the
// test of SiteResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateSiteWithMeta(
	tb testing.TB,
	md string,
	region *location_v1.RegionResource,
	ou *ouv1.OuResource,
) *location_v1.SiteResource {
	tb.Helper()
	return getInvResourceDAO().CreateSite(tb, client.FakeTenantID, SiteRegion(region), SiteOu(ou), SiteMetadata(md))
}

func CreateSiteWithArgs(
	tb testing.TB,
	name string,
	lat, long int32,
	md string,
	region *location_v1.RegionResource,
	ou *ouv1.OuResource,
	provider *provider_v1.ProviderResource,
) *location_v1.SiteResource {
	tb.Helper()
	return getInvResourceDAO().CreateSite(
		tb,
		client.FakeTenantID,
		SiteRegion(region),
		SiteOu(ou),
		SiteName(name),
		SiteCoordinates(long, lat),
		SiteMetadata(md),
		SiteProvider(provider),
	)
}

// CreateSiteWithNested - creates site with nested objects preserved. Use this when tests need
// to verify the full nested structure including Region and OU data.
func CreateSiteWithNested(tb testing.TB, region *location_v1.RegionResource, ou *ouv1.OuResource) *location_v1.SiteResource {
	tb.Helper()
	return getInvResourceDAO().CreateSiteWithNested(tb, client.FakeTenantID, SiteRegion(region), SiteOu(ou))
}

// CreateSiteWithNestedNoCleanup - creates site with nested objects preserved and no cleanup.
func CreateSiteWithNestedNoCleanup(tb testing.TB, region *location_v1.RegionResource, ou *ouv1.OuResource) *location_v1.SiteResource {
	tb.Helper()
	return getInvResourceDAO().CreateSiteWithNestedNoCleanup(tb, client.FakeTenantID, SiteRegion(region), SiteOu(ou))
}

// CreateInstance - creates instance with a cleanup. Note this helper is not really meant to be used for the
// test of InstanceResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateInstance(
	tb testing.TB,
	hostRes *computev1.HostResource,
	osRes *osv1.OperatingSystemResource,
) *computev1.InstanceResource {
	tb.Helper()
	return getInvResourceDAO().CreateInstance(tb, client.FakeTenantID, hostRes, osRes)
}

func CreateInstanceWithArgs(
	tb testing.TB,
	instanceName string,
	securityFeature osv1.SecurityFeature,
	host *computev1.HostResource,
	os *osv1.OperatingSystemResource,
	provider *provider_v1.ProviderResource,
	localAccount *localaccount_v1.LocalAccountResource,
	cleanup bool,
) (ins *computev1.InstanceResource) {
	tb.Helper()
	return getInvResourceDAO().CreateInstanceWithArgs(
		tb, client.FakeTenantID, instanceName, securityFeature, host, os, provider, localAccount, cleanup)
}

func CreateInstanceWithOpts(
	tb testing.TB,
	hostRes *computev1.HostResource,
	osRes *osv1.OperatingSystemResource,
	doCleanup bool,
	opts ...Opt[computev1.InstanceResource],
) (ins *computev1.InstanceResource) {
	tb.Helper()
	return getInvResourceDAO().CreateInstanceWithOpts(
		tb, client.FakeTenantID, hostRes, osRes, doCleanup, opts...,
	)
}

// CreateInstanceNoCleanup - creates instance with NO cleanup. Note this helper is not really meant to be used for the
// test of InstanceResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateInstanceNoCleanup(
	tb testing.TB, host *computev1.HostResource, os *osv1.OperatingSystemResource,
) *computev1.InstanceResource {
	tb.Helper()
	return getInvResourceDAO().CreateInstanceNoCleanup(tb, client.FakeTenantID, host, os)
}

func CreateInstanceWithProvider(
	tb testing.TB, host *computev1.HostResource, os *osv1.OperatingSystemResource, provider *provider_v1.ProviderResource,
) *computev1.InstanceResource {
	tb.Helper()
	return getInvResourceDAO().CreateInstanceWithProvider(tb, client.FakeTenantID, host, os, provider)
}

func CreateInstanceWithCustomConfig(
	tb testing.TB, host *computev1.HostResource, os *osv1.OperatingSystemResource,
	customConfig []*computev1.CustomConfigResource,
) *computev1.InstanceResource {
	tb.Helper()
	return getInvResourceDAO().CreateInstanceWithCustomConfig(tb, client.FakeTenantID, host, os, customConfig)
}

func CreateInstanceWithLocalAccount(
	tb testing.TB, host *computev1.HostResource, os *osv1.OperatingSystemResource, account *localaccount_v1.LocalAccountResource,
) *computev1.InstanceResource {
	tb.Helper()
	return getInvResourceDAO().CreateInstanceWithLocalAccount(tb, client.FakeTenantID, host, os, account)
}

func CreateInstanceWithOsUpdatePolicy(
	tb testing.TB, host *computev1.HostResource, os *osv1.OperatingSystemResource, oup *computev1.OSUpdatePolicyResource,
) *computev1.InstanceResource {
	tb.Helper()
	return getInvResourceDAO().CreateInstanceWithOsUpdatePolicy(tb, client.FakeTenantID, host, os, oup)
}

func CreateInstanceWithProviderNoCleanup(
	tb testing.TB, host *computev1.HostResource, os *osv1.OperatingSystemResource, provider *provider_v1.ProviderResource,
) *computev1.InstanceResource {
	tb.Helper()
	return getInvResourceDAO().CreateInstanceWithProviderNoCleanup(tb, client.FakeTenantID, host, os, provider)
}

// CreateHostStorage - creates host storage. Note this helper is not really meant to be used for the
// test of HostStorageResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateHostStorage(tb testing.TB, host *computev1.HostResource) *computev1.HoststorageResource {
	tb.Helper()
	return getInvResourceDAO().CreateHostStorage(tb, client.FakeTenantID, host)
}

func CreateHostStorageNoCleanup(tb testing.TB, host *computev1.HostResource) *computev1.HoststorageResource {
	tb.Helper()
	return getInvResourceDAO().CreateHostStorageNoCleanup(tb, client.FakeTenantID, host)
}

// CreateHostusb - creates host usb. Note this helper is not really meant to be used for the
// test of HostUsbResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by
// the eager loading.
func CreateHostusb(tb testing.TB, host *computev1.HostResource) *computev1.HostusbResource {
	tb.Helper()
	return getInvResourceDAO().CreateHostUsb(tb, client.FakeTenantID, host)
}

func CreateHostusbNoCleanup(tb testing.TB, host *computev1.HostResource) *computev1.HostusbResource {
	tb.Helper()
	return getInvResourceDAO().CreateHostUsbNoCleanup(tb, client.FakeTenantID, host)
}

// CreateCustomConfig - creates custom config. Note this helper is not really meant to be used for the
// test of CustomConfigResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateCustomConfig(tb testing.TB, name, description, config string) *computev1.CustomConfigResource {
	tb.Helper()
	return getInvResourceDAO().CreateCustomConfig(tb, client.FakeTenantID, name, description, config)
}

func CreateCustomConfigNoCleanup(tb testing.TB, name, description, config string) *computev1.CustomConfigResource {
	tb.Helper()
	return getInvResourceDAO().CreateCustomConfigNoCleanup(tb, client.FakeTenantID, name, description, config)
}

// CreateLocalAccount - creates local account. Note this helper is not really meant to be used for the
// test of LocalAccountResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateLocalAccount(tb testing.TB, username, sshKey string) *localaccount_v1.LocalAccountResource {
	tb.Helper()
	return getInvResourceDAO().CreateLocalAccount(tb, client.FakeTenantID, username, sshKey)
}

func CreateLocalAccountNoCleanup(tb testing.TB, username, sshKey string) *localaccount_v1.LocalAccountResource {
	tb.Helper()
	return getInvResourceDAO().CreateLocalAccountNoCleanup(tb, client.FakeTenantID, username, sshKey)
}

// CreateProvider - creates provider. Note this helper is not really meant to be used for the
// test of ProviderResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateProvider(tb testing.TB, name string) *provider_v1.ProviderResource {
	tb.Helper()
	return getInvResourceDAO().CreateProvider(tb, client.FakeTenantID, name,
		ProviderKind(provider_v1.ProviderKind_PROVIDER_KIND_BAREMETAL))
}

func CreateProviderbNoCleanup(tb testing.TB, name string) *provider_v1.ProviderResource {
	tb.Helper()
	return getInvResourceDAO().CreateProviderNoCleanup(tb, client.FakeTenantID, name,
		ProviderKind(provider_v1.ProviderKind_PROVIDER_KIND_BAREMETAL))
}

func CreateProviderWithArgs(tb testing.TB,
	name, apiEndpoint string, credentials []string, kind provider_v1.ProviderKind, vendor provider_v1.ProviderVendor,
) *provider_v1.ProviderResource {
	tb.Helper()
	return getInvResourceDAO().CreateProviderWithArgs(
		tb, client.FakeTenantID, name, apiEndpoint, credentials, vendor, ProviderKind(kind),
	)
}

func CreateProviderWithArgsNoCleanup(
	tb testing.TB, name, apiEndpoint string, credentials []string,
	kind provider_v1.ProviderKind, vendor provider_v1.ProviderVendor,
) *provider_v1.ProviderResource {
	tb.Helper()
	return getInvResourceDAO().
		CreateProviderWithArgsNoCleanup(tb, client.FakeTenantID, name, apiEndpoint, credentials, vendor, ProviderKind(kind))
}

// CreateOuAndReturnError - creates ou and return error if any. Note this helper is not really meant to be used for the
// test of OuResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateOuAndReturnError(tb testing.TB, md string, parent *ouv1.OuResource) (ou *ouv1.OuResource, err error) {
	tb.Helper()
	return getInvResourceDAO().CreateOuAndReturnError(tb, client.FakeTenantID, OuMetadata(md), OuParent(parent))
}

// CreateOu - creates ou. Note this helper is not really meant to be used for the
// test of OuResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateOu(tb testing.TB, parent *ouv1.OuResource) *ouv1.OuResource {
	tb.Helper()
	return getInvResourceDAO().CreateOu(tb, client.FakeTenantID, OuParent(parent))
}

func CreateOuNoCleaup(tb testing.TB, parent *ouv1.OuResource) *ouv1.OuResource {
	tb.Helper()
	return getInvResourceDAO().CreateOuNoCleanup(tb, client.FakeTenantID, OuParent(parent))
}

// CreateOuWithMeta - note this helper is not really meant to be used for the
// test of OuResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateOuWithMeta(tb testing.TB, md string, parent *ouv1.OuResource) *ouv1.OuResource {
	tb.Helper()
	return getInvResourceDAO().CreateOu(tb, client.FakeTenantID, OuMetadata(md), OuParent(parent))
}

// CreateOuWithMetaNoCleanup - note this helper is not really meant to be used for the
// test of OuResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateOuWithMetaNoCleanup(tb testing.TB, md string, parent *ouv1.OuResource) *ouv1.OuResource {
	tb.Helper()
	return getInvResourceDAO().CreateOuNoCleanup(tb, client.FakeTenantID, OuMetadata(md), OuParent(parent))
}

// CreateRegion - creates region. Note this helper is not really meant to be used for the
// test of RegionResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateRegion(tb testing.TB, parent *location_v1.RegionResource) *location_v1.RegionResource {
	tb.Helper()
	return getInvResourceDAO().CreateRegion(tb, client.FakeTenantID, RegionParentRegion(parent))
}

// HardDeleteWorkload - hard deletes the given workload via 2-phase deletion if needed.
func HardDeleteWorkload(tb testing.TB, resourceID string, workloadKind computev1.WorkloadKind) {
	tb.Helper()
	getInvResourceDAO().HardDeleteWorkload(tb, client.FakeTenantID, resourceID, workloadKind)
}

// HardDeleteWorkloadAndReturnError - hard deletes the given workload via 2-phase deletion if needed.
func HardDeleteWorkloadAndReturnError(tb testing.TB, resourceID string, workloadKind computev1.WorkloadKind) error {
	tb.Helper()
	return getInvResourceDAO().HardDeleteWorkloadAndReturnError(tb, client.FakeTenantID, resourceID, workloadKind)
}

// CreateWorkload - creates workload. Note this helper is not really meant to be used for the
// test of WorkloadResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateWorkload(tb testing.TB) *computev1.WorkloadResource {
	tb.Helper()
	return getInvResourceDAO().CreateWorkload(tb, client.FakeTenantID)
}

func CreateWorkloadNoCleanup(tb testing.TB) *computev1.WorkloadResource {
	tb.Helper()
	return getInvResourceDAO().CreateWorkloadNoCleanup(tb, client.FakeTenantID)
}

// CreateWorkloadMember - create WorkloadMember. Note this helper is not really meant to be used for the
// test of WorkloadMember, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateWorkloadMember(
	tb testing.TB, workload *computev1.WorkloadResource, instance *computev1.InstanceResource,
) *computev1.WorkloadMember {
	tb.Helper()
	return getInvResourceDAO().CreateWorkloadMember(tb, client.FakeTenantID, workload, instance)
}

func CreateWorkloadMemberNoCleanup(
	tb testing.TB, workload *computev1.WorkloadResource, instance *computev1.InstanceResource,
) *computev1.WorkloadMember {
	tb.Helper()
	return getInvResourceDAO().CreateWorkloadMemberNoCleanup(tb, client.FakeTenantID, workload, instance)
}

// HardDeleteNetlink - hard deletes the given netlink via 2-phase deletion.
func HardDeleteNetlink(tb testing.TB, resourceID string) {
	tb.Helper()
	getInvResourceDAO().HardDeleteNetlink(tb, client.FakeTenantID, resourceID)
}

// CreateNetworkSegment - creates network segment. Note this helper is not really meant to be used for the
// test of NetworkSegment, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateNetworkSegment(tb testing.TB, name string, site *location_v1.SiteResource, vlanID int32) *networkv1.NetworkSegment {
	tb.Helper()
	return getInvResourceDAO().CreateNetworkSegment(tb, client.FakeTenantID, name, site, vlanID, true)
}

// CreateEndpoint - Note this helper is not really meant to be used for the
// test of EndpointResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateEndpoint(tb testing.TB, host *computev1.HostResource) (endpoint *networkv1.EndpointResource) {
	tb.Helper()
	return getInvResourceDAO().CreateEndpoint(tb, client.FakeTenantID, host)
}

func CreateRemoteAccessConfiguration(t *testing.T, opts ...Opt[v1.RemoteAccessConfiguration]) *v1.RemoteAccessConfiguration {
	t.Helper()
	return getInvResourceDAO().CreateRemoteAccessConfiguration(t, client.FakeTenantID, opts...)
}

func CreateRemoteAccessConfigurationNoCleanup(
	t *testing.T, opts ...Opt[v1.RemoteAccessConfiguration],
) *v1.RemoteAccessConfiguration {
	t.Helper()
	return getInvResourceDAO().CreateRemoteAccessConfigurationNoCleanup(t, client.FakeTenantID, opts...)
}

func HardDeleteRemoteAccessConfiguration(tb testing.TB, resourceID string) {
	tb.Helper()
	getInvResourceDAO().HardDeleteRemoteAccessConfiguration(tb, client.FakeTenantID, resourceID)
}

func HardDeleteRemoteAccessConfigurationAndReturnError(tb testing.TB, resourceID string) error {
	tb.Helper()
	return getInvResourceDAO().HardDeleteRemoteAccessConfigurationAndReturnError(tb, client.FakeTenantID, resourceID)
}

func CreateSiteNoCleanup(tb testing.TB, region *location_v1.RegionResource, ou *ouv1.OuResource) *location_v1.SiteResource {
	tb.Helper()
	return getInvResourceDAO().CreateSiteNoCleanup(tb, client.FakeTenantID, SiteRegion(region), SiteOu(ou))
}

func CreateRegionNoCleanup(tb testing.TB, parent *location_v1.RegionResource) *location_v1.RegionResource {
	tb.Helper()
	return getInvResourceDAO().CreateRegionNoCleanup(tb, client.FakeTenantID, RegionParentRegion(parent))
}

// CreateRegionWithMeta - creates region with metadata. Note this helper is not really meant to be used for the
// test of RegionResource, but they are typically leveraged in case of wider
// tests involving long chain of relations that are not usually fulfilled by the eager loading.
func CreateRegionWithMeta(tb testing.TB, md string, parent *location_v1.RegionResource) *location_v1.RegionResource {
	tb.Helper()
	return getInvResourceDAO().CreateRegion(tb, client.FakeTenantID, RegionMetadata(md), RegionParentRegion(parent))
}

func CreateRegionWithMetaNoCleanup(tb testing.TB, md string, parent *location_v1.RegionResource) *location_v1.RegionResource {
	tb.Helper()
	return getInvResourceDAO().CreateRegionNoCleanup(tb, client.FakeTenantID, RegionMetadata(md), RegionParentRegion(parent))
}

func CreateTenant(t *testing.T, opts ...Opt[tenantv1.Tenant]) *tenantv1.Tenant {
	t.Helper()
	return getInvResourceDAO().CreateTenantWithOpts(t, client.FakeTenantID, true, opts...)
}

func CreateTenantNoCleanup(t *testing.T, opts ...Opt[tenantv1.Tenant]) *tenantv1.Tenant {
	t.Helper()
	return getInvResourceDAO().CreateTenantWithOpts(t, client.FakeTenantID, false, opts...)
}

// CreateOsUpdatePolicy creates OS UpdatePolicy.
func CreateOsUpdatePolicy(tb testing.TB) *computev1.OSUpdatePolicyResource {
	tb.Helper()
	return getInvResourceDAO().CreateOSUpdatePolicy(tb, client.FakeTenantID)
}
