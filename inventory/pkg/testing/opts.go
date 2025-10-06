// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	compute_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	network_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	ou_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/ou/v1"
	provider_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	schedule_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	status_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	telemetry_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
)

type Opt[T any] func(*T)

func RSRStatus(status schedule_v1.ScheduleStatus) Opt[schedule_v1.RepeatedScheduleResource] {
	return func(s *schedule_v1.RepeatedScheduleResource) {
		s.ScheduleStatus = status
	}
}

func RSRRegion(region *location_v1.RegionResource) Opt[schedule_v1.RepeatedScheduleResource] {
	return func(s *schedule_v1.RepeatedScheduleResource) {
		if region == nil {
			return
		}
		s.Relation = &schedule_v1.RepeatedScheduleResource_TargetRegion{
			TargetRegion: region,
		}
	}
}

func RSRTargetWorkload(w *compute_v1.WorkloadResource) Opt[schedule_v1.RepeatedScheduleResource] {
	return func(s *schedule_v1.RepeatedScheduleResource) {
		if w == nil {
			return
		}
		s.Relation = &schedule_v1.RepeatedScheduleResource_TargetWorkload{
			TargetWorkload: w,
		}
	}
}

func RSRTargetHost(host *compute_v1.HostResource) Opt[schedule_v1.RepeatedScheduleResource] {
	return func(s *schedule_v1.RepeatedScheduleResource) {
		if host == nil {
			return
		}
		s.Relation = &schedule_v1.RepeatedScheduleResource_TargetHost{
			TargetHost: host,
		}
	}
}

func RSRTargetSite(site *location_v1.SiteResource) Opt[schedule_v1.RepeatedScheduleResource] {
	return func(s *schedule_v1.RepeatedScheduleResource) {
		if site == nil {
			return
		}
		s.Relation = &schedule_v1.RepeatedScheduleResource_TargetSite{
			TargetSite: site,
		}
	}
}

func RSRDayWeek(dayWeek string) Opt[schedule_v1.RepeatedScheduleResource] {
	return func(s *schedule_v1.RepeatedScheduleResource) {
		s.CronDayWeek = dayWeek
	}
}

func RSRMonth(month string) Opt[schedule_v1.RepeatedScheduleResource] {
	return func(s *schedule_v1.RepeatedScheduleResource) {
		s.CronMonth = month
	}
}

func RSRDayMonth(dayMonth string) Opt[schedule_v1.RepeatedScheduleResource] {
	return func(s *schedule_v1.RepeatedScheduleResource) {
		s.CronDayMonth = dayMonth
	}
}

func RSRHours(hours string) Opt[schedule_v1.RepeatedScheduleResource] {
	return func(s *schedule_v1.RepeatedScheduleResource) {
		s.CronHours = hours
	}
}

func RSRMinutes(minutes string) Opt[schedule_v1.RepeatedScheduleResource] {
	return func(s *schedule_v1.RepeatedScheduleResource) {
		s.CronMinutes = minutes
	}
}

func RSRDuration(duration uint32) Opt[schedule_v1.RepeatedScheduleResource] {
	return func(s *schedule_v1.RepeatedScheduleResource) {
		s.DurationSeconds = duration
	}
}

func SSRRegion(region *location_v1.RegionResource) Opt[schedule_v1.SingleScheduleResource] {
	return func(s *schedule_v1.SingleScheduleResource) {
		s.Relation = &schedule_v1.SingleScheduleResource_TargetRegion{
			TargetRegion: region,
		}
	}
}

func SSRTargetWorkload(w *compute_v1.WorkloadResource) Opt[schedule_v1.SingleScheduleResource] {
	return func(s *schedule_v1.SingleScheduleResource) {
		if w == nil {
			return
		}
		s.Relation = &schedule_v1.SingleScheduleResource_TargetWorkload{
			TargetWorkload: w,
		}
	}
}

func SSRTargetHost(host *compute_v1.HostResource) Opt[schedule_v1.SingleScheduleResource] {
	return func(s *schedule_v1.SingleScheduleResource) {
		if host == nil {
			return
		}
		s.Relation = &schedule_v1.SingleScheduleResource_TargetHost{
			TargetHost: host,
		}
	}
}

func SSRTargetSite(site *location_v1.SiteResource) Opt[schedule_v1.SingleScheduleResource] {
	return func(s *schedule_v1.SingleScheduleResource) {
		if site == nil {
			return
		}
		s.Relation = &schedule_v1.SingleScheduleResource_TargetSite{
			TargetSite: site,
		}
	}
}

func SSRStatus(status schedule_v1.ScheduleStatus) Opt[schedule_v1.SingleScheduleResource] {
	return func(s *schedule_v1.SingleScheduleResource) {
		s.ScheduleStatus = status
	}
}

func SSRStart(start uint64) Opt[schedule_v1.SingleScheduleResource] {
	return func(s *schedule_v1.SingleScheduleResource) {
		s.StartSeconds = start
	}
}

func SSREnd(end uint64) Opt[schedule_v1.SingleScheduleResource] {
	return func(s *schedule_v1.SingleScheduleResource) {
		s.EndSeconds = end
	}
}

func HostSerialNumber(sn string) Opt[compute_v1.HostResource] {
	return func(c *compute_v1.HostResource) {
		c.SerialNumber = sn
	}
}

func HostHostName(hn string) Opt[compute_v1.HostResource] {
	return func(c *compute_v1.HostResource) {
		c.Hostname = hn
	}
}

func HostUUID(uuid string) Opt[compute_v1.HostResource] {
	return func(c *compute_v1.HostResource) {
		c.Uuid = uuid
	}
}

func HostMetadata(md string) Opt[compute_v1.HostResource] {
	return func(c *compute_v1.HostResource) {
		c.Metadata = md
	}
}

func HostSite(v *location_v1.SiteResource) Opt[compute_v1.HostResource] {
	return func(c *compute_v1.HostResource) {
		c.Site = v
	}
}

func HostProvider(v *provider_v1.ProviderResource) Opt[compute_v1.HostResource] {
	return func(c *compute_v1.HostResource) {
		c.Provider = v
	}
}

func SiteName(name string) Opt[location_v1.SiteResource] {
	return func(s *location_v1.SiteResource) {
		s.Name = name
	}
}

func SiteMetadata(md string) Opt[location_v1.SiteResource] {
	return func(s *location_v1.SiteResource) {
		s.Metadata = md
	}
}

func SiteCoordinates(lat, long int32) Opt[location_v1.SiteResource] {
	return func(s *location_v1.SiteResource) {
		s.SiteLng = long
		s.SiteLat = lat
	}
}

func SiteRegion(region *location_v1.RegionResource) Opt[location_v1.SiteResource] {
	return func(s *location_v1.SiteResource) {
		s.Region = region
	}
}

func SiteOu(ou *ou_v1.OuResource) Opt[location_v1.SiteResource] {
	return func(s *location_v1.SiteResource) {
		s.Ou = ou
	}
}

func SiteProvider(provider *provider_v1.ProviderResource) Opt[location_v1.SiteResource] {
	return func(s *location_v1.SiteResource) {
		s.Provider = provider
	}
}

func RegionParentRegion(pr *location_v1.RegionResource) Opt[location_v1.RegionResource] {
	return func(r *location_v1.RegionResource) {
		r.ParentRegion = pr
	}
}

func RegionMetadata(md string) Opt[location_v1.RegionResource] {
	return func(r *location_v1.RegionResource) {
		r.Metadata = md
	}
}

func OuParent(parent *ou_v1.OuResource) Opt[ou_v1.OuResource] {
	return func(o *ou_v1.OuResource) {
		o.ParentOu = parent
	}
}

func OuMetadata(md string) Opt[ou_v1.OuResource] {
	return func(o *ou_v1.OuResource) {
		o.Metadata = md
	}
}

func TelemetryProfileTarget[T location_v1.SiteResource | location_v1.RegionResource | compute_v1.InstanceResource](
	target *T,
) TelemetryProfileTargetConfigurator {
	return func(tp *telemetry_v1.TelemetryProfile) {
		switch t := any(target).(type) {
		case *location_v1.RegionResource:
			tp.Relation = &telemetry_v1.TelemetryProfile_Region{
				Region: t,
			}
		case *location_v1.SiteResource:
			tp.Relation = &telemetry_v1.TelemetryProfile_Site{
				Site: t,
			}
		case *compute_v1.InstanceResource:
			tp.Relation = &telemetry_v1.TelemetryProfile_Instance{
				Instance: t,
			}
		}
	}
}

func TenantCurrentState(v tenantv1.TenantState) Opt[tenantv1.Tenant] {
	return func(t *tenantv1.Tenant) {
		t.CurrentState = v
	}
}

func TenantCurrentStateDeleted() Opt[tenantv1.Tenant] {
	return func(t *tenantv1.Tenant) {
		t.CurrentState = tenantv1.TenantState_TENANT_STATE_DELETED
	}
}

func TenantCurrentStateCreated() Opt[tenantv1.Tenant] {
	return func(t *tenantv1.Tenant) {
		t.CurrentState = tenantv1.TenantState_TENANT_STATE_CREATED
	}
}

func TenantDesiredStateDeleted() Opt[tenantv1.Tenant] {
	return func(t *tenantv1.Tenant) {
		t.DesiredState = tenantv1.TenantState_TENANT_STATE_DELETED
	}
}

func TenantDesiredStateCreated() Opt[tenantv1.Tenant] {
	return func(t *tenantv1.Tenant) {
		t.DesiredState = tenantv1.TenantState_TENANT_STATE_CREATED
	}
}

func TenantDesiredState(v tenantv1.TenantState) Opt[tenantv1.Tenant] {
	return func(t *tenantv1.Tenant) {
		t.DesiredState = v
	}
}

func TenantWatcherOSManager(v bool) Opt[tenantv1.Tenant] {
	return func(t *tenantv1.Tenant) {
		t.WatcherOsmanager = v
	}
}

func ProviderKind(pk provider_v1.ProviderKind) Opt[provider_v1.ProviderResource] {
	return func(p *provider_v1.ProviderResource) {
		p.ProviderKind = pk
	}
}

func ProviderConfig(pc string) Opt[provider_v1.ProviderResource] {
	return func(p *provider_v1.ProviderResource) {
		p.Config = pc
	}
}

func NetlinkCurrentStateDeleted() Opt[network_v1.NetlinkResource] {
	return func(t *network_v1.NetlinkResource) {
		t.CurrentState = network_v1.NetlinkState_NETLINK_STATE_DELETED
	}
}

func WorkloadKind(wk compute_v1.WorkloadKind) Opt[compute_v1.WorkloadResource] {
	return func(t *compute_v1.WorkloadResource) {
		t.Kind = wk
	}
}

func InstalledPackages(pkgs string) Opt[osv1.OperatingSystemResource] {
	return func(t *osv1.OperatingSystemResource) {
		t.InstalledPackages = pkgs
	}
}

func InstalledPackagesURL(url string) Opt[osv1.OperatingSystemResource] {
	return func(t *osv1.OperatingSystemResource) {
		t.InstalledPackagesUrl = url
	}
}

func Sha256(sha256 string) Opt[osv1.OperatingSystemResource] {
	return func(t *osv1.OperatingSystemResource) {
		t.Sha256 = sha256
	}
}

func ProfileName(profileName string) Opt[osv1.OperatingSystemResource] {
	return func(t *osv1.OperatingSystemResource) {
		t.ProfileName = profileName
	}
}

func ImageID(imageID string) Opt[osv1.OperatingSystemResource] {
	return func(t *osv1.OperatingSystemResource) {
		t.ImageId = imageID
	}
}

func OsName(name string) Opt[osv1.OperatingSystemResource] {
	return func(t *osv1.OperatingSystemResource) {
		t.Name = name
	}
}

func Metadata(metadata string) Opt[osv1.OperatingSystemResource] {
	return func(t *osv1.OperatingSystemResource) {
		t.Metadata = metadata
	}
}

func SecurityFeature(secFeature osv1.SecurityFeature) Opt[osv1.OperatingSystemResource] {
	return func(t *osv1.OperatingSystemResource) {
		t.SecurityFeature = secFeature
	}
}

func OsType(osType osv1.OsType) Opt[osv1.OperatingSystemResource] {
	return func(t *osv1.OperatingSystemResource) {
		t.OsType = osType
	}
}

func OSUpdatePolicyTarget() Opt[compute_v1.OSUpdatePolicyResource] {
	return func(oup *compute_v1.OSUpdatePolicyResource) {
		oup.UpdatePolicy = compute_v1.UpdatePolicy_UPDATE_POLICY_TARGET
	}
}

func OSUpdatePolicyLatest() Opt[compute_v1.OSUpdatePolicyResource] {
	return func(oup *compute_v1.OSUpdatePolicyResource) {
		oup.UpdatePolicy = compute_v1.UpdatePolicy_UPDATE_POLICY_LATEST
	}
}

func OsUpdatePolicyName(name string) Opt[compute_v1.OSUpdatePolicyResource] {
	return func(oup *compute_v1.OSUpdatePolicyResource) {
		oup.Name = name
	}
}

func OsUpdatePolicyDescription(desc string) Opt[compute_v1.OSUpdatePolicyResource] {
	return func(oup *compute_v1.OSUpdatePolicyResource) {
		oup.Description = desc
	}
}

func OSUpdatePolicyUpdatePackages(pkgs string) Opt[compute_v1.OSUpdatePolicyResource] {
	return func(oup *compute_v1.OSUpdatePolicyResource) {
		oup.UpdatePackages = pkgs
	}
}

func OSUpdatePolicyUpdateSources(sources []string) Opt[compute_v1.OSUpdatePolicyResource] {
	return func(oup *compute_v1.OSUpdatePolicyResource) {
		oup.UpdateSources = sources
	}
}

func OSUpdatePolicyUpdateKernelCommand(cmd string) Opt[compute_v1.OSUpdatePolicyResource] {
	return func(oup *compute_v1.OSUpdatePolicyResource) {
		oup.UpdateKernelCommand = cmd
	}
}

func OSUpdatePolicyTargetOS(targetOS *osv1.OperatingSystemResource) Opt[compute_v1.OSUpdatePolicyResource] {
	return func(oup *compute_v1.OSUpdatePolicyResource) {
		oup.TargetOs = targetOS
	}
}

func InstanceOsUpdatePolicy(
	policy *compute_v1.OSUpdatePolicyResource,
) Opt[compute_v1.InstanceResource] {
	return func(i *compute_v1.InstanceResource) {
		i.OsUpdatePolicy = policy
	}
}

func OsUpdateRunName(name string) Opt[compute_v1.OSUpdateRunResource] {
	return func(our *compute_v1.OSUpdateRunResource) {
		our.Name = name
	}
}

func OsUpdateRunDescription(desc string) Opt[compute_v1.OSUpdateRunResource] {
	return func(our *compute_v1.OSUpdateRunResource) {
		our.Description = desc
	}
}

func OSUpdateRunAppliedPolicy(appliedPolicy *compute_v1.OSUpdatePolicyResource) Opt[compute_v1.OSUpdateRunResource] {
	return func(our *compute_v1.OSUpdateRunResource) {
		our.AppliedPolicy = appliedPolicy
	}
}

func OSUpdateRunInstance(instance *compute_v1.InstanceResource) Opt[compute_v1.OSUpdateRunResource] {
	return func(our *compute_v1.OSUpdateRunResource) {
		our.Instance = instance
	}
}

func OSUpdateRunStatusIndicator(statusIndicator status_v1.StatusIndication) Opt[compute_v1.OSUpdateRunResource] {
	return func(our *compute_v1.OSUpdateRunResource) {
		our.StatusIndicator = statusIndicator
	}
}

func OSUpdateRunStatus(status string) Opt[compute_v1.OSUpdateRunResource] {
	return func(our *compute_v1.OSUpdateRunResource) {
		our.Status = status
	}
}

func OSUpdateRunStatusDetails(statusDetails string) Opt[compute_v1.OSUpdateRunResource] {
	return func(our *compute_v1.OSUpdateRunResource) {
		our.StatusDetails = statusDetails
	}
}

func OSUpdateRunStatusTimestamp(statusTimestamp uint64) Opt[compute_v1.OSUpdateRunResource] {
	return func(our *compute_v1.OSUpdateRunResource) {
		our.StatusTimestamp = statusTimestamp
	}
}

func OSUpdateRunStartTime(startTime uint64) Opt[compute_v1.OSUpdateRunResource] {
	return func(our *compute_v1.OSUpdateRunResource) {
		our.StartTime = startTime
	}
}

func OSUpdateRunEndTime(endTime uint64) Opt[compute_v1.OSUpdateRunResource] {
	return func(our *compute_v1.OSUpdateRunResource) {
		our.EndTime = endTime
	}
}
