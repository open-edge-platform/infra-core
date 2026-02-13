// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"strings"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	localaccount_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/localaccount/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	network_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	ou_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/ou/v1"
	provider_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	remoteaccessv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/remoteaccess/v1"
	schedule_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	telemetry_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
)

// This file contains helper functions to convert Ent schema objects to their
// protobuf equivalents.

func entTelemetryProfileToProtoTelemetryProfile(
	telemetryProfile *ent.TelemetryProfile,
) *telemetry_v1.TelemetryProfile {
	if telemetryProfile == nil {
		return nil
	}
	kind := telemetry_v1.TelemetryResourceKind_value[telemetryProfile.Kind.String()]
	logLevel := telemetry_v1.SeverityLevel_value[telemetryProfile.LogLevel.String()]
	protoTelemetryProfile := &telemetry_v1.TelemetryProfile{
		ResourceId:      telemetryProfile.ResourceID,
		Kind:            telemetry_v1.TelemetryResourceKind(kind),
		LogLevel:        telemetry_v1.SeverityLevel(logLevel),
		MetricsInterval: telemetryProfile.MetricsInterval,
		TenantId:        telemetryProfile.TenantID,
		CreatedAt:       telemetryProfile.CreatedAt,
		UpdatedAt:       telemetryProfile.UpdatedAt,
	}

	if inst, qerr := telemetryProfile.Edges.InstanceOrErr(); qerr == nil {
		protoTelemetryProfile.Relation = &telemetry_v1.TelemetryProfile_Instance{
			Instance: entInstanceResourceToProtoInstanceResource(inst),
		}
	}
	if site, qerr := telemetryProfile.Edges.SiteOrErr(); qerr == nil {
		protoTelemetryProfile.Relation = &telemetry_v1.TelemetryProfile_Site{
			Site: entSiteResourceToProtoSiteResource(site),
		}
	}
	if region, qerr := telemetryProfile.Edges.RegionOrErr(); qerr == nil {
		protoTelemetryProfile.Relation = &telemetry_v1.TelemetryProfile_Region{
			Region: entRegionResourceToProtoRegionResource(region),
		}
	}
	if telGroup, qerr := telemetryProfile.Edges.GroupOrErr(); qerr == nil {
		protoTelemetryProfile.Group = entTelemetryGroupResourceToProtoTelemetryGroupResource(telGroup)
	}

	return protoTelemetryProfile
}

func entTelemetryGroupResourceToProtoTelemetryGroupResource(
	telemetryGroup *ent.TelemetryGroupResource,
) *telemetry_v1.TelemetryGroupResource {
	if telemetryGroup == nil {
		return nil
	}
	kind := telemetry_v1.TelemetryResourceKind_value[telemetryGroup.Kind.String()]
	collectorKind := telemetry_v1.CollectorKind_value[telemetryGroup.CollectorKind.String()]
	protoTelemetry := &telemetry_v1.TelemetryGroupResource{
		ResourceId:    telemetryGroup.ResourceID,
		Kind:          telemetry_v1.TelemetryResourceKind(kind),
		CollectorKind: telemetry_v1.CollectorKind(collectorKind),
		Name:          telemetryGroup.Name,
		TenantId:      telemetryGroup.TenantID,
		CreatedAt:     telemetryGroup.CreatedAt,
		UpdatedAt:     telemetryGroup.UpdatedAt,
	}

	if telemetryGroup.Groups != "" {
		protoTelemetry.Groups = strings.Split(telemetryGroup.Groups, "|")
	}

	if profiles, qerr := telemetryGroup.Edges.ProfilesOrErr(); qerr == nil {
		for _, p := range profiles {
			protoTelemetry.Profiles = append(protoTelemetry.Profiles, entTelemetryProfileToProtoTelemetryProfile(p))
		}
	}

	return protoTelemetry
}

func entSingleScheduleResourceToProtoSingleScheduleResource(singleschedule *ent.SingleScheduleResource,
) *schedule_v1.SingleScheduleResource {
	if singleschedule == nil {
		return nil
	}
	// Convert the fields directly.
	status := schedule_v1.ScheduleStatus_value[singleschedule.ScheduleStatus.String()] // Defaults to 0 if not found
	protoSingle := &schedule_v1.SingleScheduleResource{
		ResourceId:     singleschedule.ResourceID,
		ScheduleStatus: schedule_v1.ScheduleStatus(status),
		Name:           singleschedule.Name,
		StartSeconds:   singleschedule.StartSeconds,
		EndSeconds:     singleschedule.EndSeconds,
		TenantId:       singleschedule.TenantID,
		CreatedAt:      singleschedule.CreatedAt,
		UpdatedAt:      singleschedule.UpdatedAt,
	}

	// Convert the edges recursively.
	if targetHost, qerr := singleschedule.Edges.TargetHostOrErr(); qerr == nil {
		protoSingle.Relation = &schedule_v1.SingleScheduleResource_TargetHost{
			TargetHost: entHostResourceToProtoHostResource(targetHost),
		}
	}

	if targetRegion, err := singleschedule.Edges.TargetRegionOrErr(); err == nil {
		protoSingle.Relation = &schedule_v1.SingleScheduleResource_TargetRegion{
			TargetRegion: entRegionResourceToProtoRegionResource(targetRegion),
		}
	}
	if targetSite, qerr := singleschedule.Edges.TargetSiteOrErr(); qerr == nil {
		protoSingle.Relation = &schedule_v1.SingleScheduleResource_TargetSite{
			TargetSite: entSiteResourceToProtoSiteResource(targetSite),
		}
	}

	if targetWorkload, qerr := singleschedule.Edges.TargetWorkloadOrErr(); qerr == nil {
		protoSingle.Relation = &schedule_v1.SingleScheduleResource_TargetWorkload{
			TargetWorkload: entWorkloadResourceToProtoWorkloadResource(targetWorkload),
		}
	}

	return protoSingle
}

func entRepeatedScheduleResourceToProtoRepeatedScheduleResource(repeatedschedule *ent.RepeatedScheduleResource,
) *schedule_v1.RepeatedScheduleResource {
	if repeatedschedule == nil {
		return nil
	}
	// Convert the fields directly.
	status := schedule_v1.ScheduleStatus_value[repeatedschedule.ScheduleStatus.String()] // Defaults to 0 if not found
	protoRepeated := &schedule_v1.RepeatedScheduleResource{
		ResourceId:      repeatedschedule.ResourceID,
		ScheduleStatus:  schedule_v1.ScheduleStatus(status),
		Name:            repeatedschedule.Name,
		DurationSeconds: repeatedschedule.DurationSeconds,
		CronMinutes:     repeatedschedule.CronMinutes,
		CronHours:       repeatedschedule.CronHours,
		CronDayMonth:    repeatedschedule.CronDayMonth,
		CronMonth:       repeatedschedule.CronMonth,
		CronDayWeek:     repeatedschedule.CronDayWeek,
		TenantId:        repeatedschedule.TenantID,
		CreatedAt:       repeatedschedule.CreatedAt,
		UpdatedAt:       repeatedschedule.UpdatedAt,
	}
	// Convert the edges recursively.
	if targetHost, qerr := repeatedschedule.Edges.TargetHostOrErr(); qerr == nil {
		protoRepeated.Relation = &schedule_v1.RepeatedScheduleResource_TargetHost{
			TargetHost: entHostResourceToProtoHostResource(targetHost),
		}
	}
	if targetRegion, err := repeatedschedule.Edges.TargetRegionOrErr(); err == nil {
		protoRepeated.Relation = &schedule_v1.RepeatedScheduleResource_TargetRegion{
			TargetRegion: entRegionResourceToProtoRegionResource(targetRegion),
		}
	}
	if targetSite, qerr := repeatedschedule.Edges.TargetSiteOrErr(); qerr == nil {
		protoRepeated.Relation = &schedule_v1.RepeatedScheduleResource_TargetSite{
			TargetSite: entSiteResourceToProtoSiteResource(targetSite),
		}
	}
	if targetWorkload, qerr := repeatedschedule.Edges.TargetWorkloadOrErr(); qerr == nil {
		protoRepeated.Relation = &schedule_v1.RepeatedScheduleResource_TargetWorkload{
			TargetWorkload: entWorkloadResourceToProtoWorkloadResource(targetWorkload),
		}
	}

	return protoRepeated
}

func entOperatingSystemResourceToProtoOperatingSystemResource(os *ent.OperatingSystemResource,
) *osv1.OperatingSystemResource {
	if os == nil {
		return nil
	}
	// Convert the fields directly.
	securityFeatures := osv1.SecurityFeature_value[os.SecurityFeature.String()] // Defaults to 0 if not found
	osType := osv1.OsType_value[os.OsType.String()]                             // Defaults to 0 if not found
	osProvider := osv1.OsProviderKind_value[os.OsProvider.String()]             // Defaults to 0 if not found
	protoUpdate := &osv1.OperatingSystemResource{
		ResourceId:           os.ResourceID,
		Name:                 os.Name,
		Architecture:         os.Architecture,
		ImageUrl:             os.ImageURL,
		ImageId:              os.ImageID,
		Sha256:               os.Sha256,
		ProfileName:          os.ProfileName,
		ProfileVersion:       os.ProfileVersion,
		InstalledPackages:    os.InstalledPackages,
		InstalledPackagesUrl: os.InstalledPackagesURL,
		SecurityFeature:      osv1.SecurityFeature(securityFeatures),
		OsType:               osv1.OsType(osType),
		OsProvider:           osv1.OsProviderKind(osProvider),
		PlatformBundle:       os.PlatformBundle,
		Description:          os.Description,
		ExistingCvesUrl:      os.ExistingCvesURL,
		ExistingCves:         os.ExistingCves,
		FixedCvesUrl:         os.FixedCvesURL,
		FixedCves:            os.FixedCves,
		TenantId:             os.TenantID,
		CreatedAt:            os.CreatedAt,
		UpdatedAt:            os.UpdatedAt,
		Metadata:             os.Metadata,
		TlsCaCert:            os.TLSCaCert,
	}

	return protoUpdate
}

func entOuResourceToProtoOuResource(ou *ent.OuResource) *ou_v1.OuResource {
	if ou == nil {
		return nil
	}
	// Convert the fields directly.
	protoOu := &ou_v1.OuResource{
		ResourceId: ou.ResourceID,
		Name:       ou.Name,
		OuKind:     ou.OuKind,
		Metadata:   ou.Metadata,
		TenantId:   ou.TenantID,
		CreatedAt:  ou.CreatedAt,
		UpdatedAt:  ou.UpdatedAt,
	}
	// Convert the edges recursively. Ent only returns the first level.
	if parentOu, qerr := ou.Edges.ParentOuOrErr(); qerr == nil {
		protoOu.ParentOu = entOuResourceToProtoOuResource(parentOu)
	}

	return protoOu
}

func entRegionResourceToProtoRegionResource(region *ent.RegionResource) *location_v1.RegionResource {
	if region == nil {
		return nil
	}
	// Convert the fields directly.
	protoRegion := &location_v1.RegionResource{
		ResourceId: region.ResourceID,
		Name:       region.Name,
		RegionKind: region.RegionKind,
		Metadata:   region.Metadata,
		TenantId:   region.TenantID,
		CreatedAt:  region.CreatedAt,
		UpdatedAt:  region.UpdatedAt,
	}
	// Convert the edges recursively.
	if parentRegion, qerr := region.Edges.ParentRegionOrErr(); qerr == nil {
		protoRegion.ParentRegion = entRegionResourceToProtoRegionResource(parentRegion)
	}

	return protoRegion
}

func entSiteResourceToProtoSiteResource(site *ent.SiteResource) *location_v1.SiteResource {
	if site == nil {
		return nil
	}
	// Convert the fields directly.
	protoSite := &location_v1.SiteResource{
		ResourceId:      site.ResourceID,
		Name:            site.Name,
		Address:         site.Address,
		SiteLat:         site.SiteLat,
		SiteLng:         site.SiteLng,
		MetricsEndpoint: site.MetricsEndpoint,
		HttpProxy:       site.HTTPProxy,
		HttpsProxy:      site.HTTPSProxy,
		FtpProxy:        site.FtpProxy,
		NoProxy:         site.NoProxy,
		Metadata:        site.Metadata,
		TenantId:        site.TenantID,
		CreatedAt:       site.CreatedAt,
		UpdatedAt:       site.UpdatedAt,
	}
	// We need to handle the special case of empty string, which should not
	// result in a slice of length 1 with an empty string, but an empty slice.
	if site.DNSServers != "" {
		protoSite.DnsServers = strings.Split(site.DNSServers, "|")
	}
	if site.DockerRegistries != "" {
		protoSite.DockerRegistries = strings.Split(site.DockerRegistries, "|")
	}
	// Convert the edges recursively.
	if region, qerr := site.Edges.RegionOrErr(); qerr == nil {
		protoSite.Region = entRegionResourceToProtoRegionResource(region)
	}
	if ou, qerr := site.Edges.OuOrErr(); qerr == nil {
		protoSite.Ou = entOuResourceToProtoOuResource(ou)
	}
	if prov, qerr := site.Edges.ProviderOrErr(); qerr == nil {
		protoSite.Provider = entProviderResourceToProtoProviderResource(prov)
	}

	return protoSite
}

func entProviderResourceToProtoProviderResource(provider *ent.ProviderResource) *provider_v1.ProviderResource {
	if provider == nil {
		return nil
	}
	// Convert the fields directly.
	kind := provider_v1.ProviderKind_value[provider.ProviderKind.String()]       // Defaults to 0 if not found
	vendor := provider_v1.ProviderVendor_value[provider.ProviderVendor.String()] // Defaults to 0 if not found
	protoProvider := &provider_v1.ProviderResource{
		ResourceId:     provider.ResourceID,
		ProviderKind:   provider_v1.ProviderKind(kind),
		ProviderVendor: provider_v1.ProviderVendor(vendor),
		Name:           provider.Name,
		ApiEndpoint:    provider.APIEndpoint,
		Config:         provider.Config,
		TenantId:       provider.TenantID,
		CreatedAt:      provider.CreatedAt,
		UpdatedAt:      provider.UpdatedAt,
	}

	if provider.APICredentials != "" {
		protoProvider.ApiCredentials = strings.Split(provider.APICredentials, "|")
	}

	return protoProvider
}

func entLocalAccountResourceToProtoLocalAccountResource(
	localaccount *ent.LocalAccountResource,
) *localaccount_v1.LocalAccountResource {
	if localaccount == nil {
		return nil
	}
	protoLocalAccount := &localaccount_v1.LocalAccountResource{
		ResourceId: localaccount.ResourceID,
		Username:   localaccount.Username,
		SshKey:     localaccount.SSHKey,
		TenantId:   localaccount.TenantID,
		CreatedAt:  localaccount.CreatedAt,
	}
	return protoLocalAccount
}

//nolint:cyclop,funlen // host resource has many edges and fields that need conversion from ent to proto
func entHostResourceToProtoHostResource(host *ent.HostResource) *computev1.HostResource {
	if host == nil {
		return nil
	}
	// Convert the fields directly.
	desiredState := computev1.HostState_value[host.DesiredState.String()]            // Defaults to 0 if not found
	currentState := computev1.HostState_value[host.CurrentState.String()]            // Defaults to 0 if not found
	bcmKind := computev1.BaremetalControllerKind_value[host.BmcKind.String()]        // Defaults to 0 if not found
	desiredPowerState := computev1.PowerState_value[host.DesiredPowerState.String()] // Defaults to 0 if not found
	currentPowerState := computev1.PowerState_value[host.CurrentPowerState.String()] // Defaults to 0 if not found
	hostStatusIndicator := statusv1.StatusIndication_value[host.HostStatusIndicator.String()]
	onboardingStatusIndicator := statusv1.StatusIndication_value[host.OnboardingStatusIndicator.String()]
	registrationStatusIndicator := statusv1.StatusIndication_value[host.RegistrationStatusIndicator.String()]
	powerCommandPolicy := computev1.PowerCommandPolicy_value[host.PowerCommandPolicy.String()]
	amtSku := computev1.AmtSku_value[host.AmtSku.String()]
	AmtControlMode := computev1.AmtControlMode_value[host.AmtControlMode.String()]
	desiredAmtState := computev1.AmtState_value[host.DesiredAmtState.String()]
	currentAmtState := computev1.AmtState_value[host.CurrentAmtState.String()]
	powerStatusIndicator := statusv1.StatusIndication_value[host.PowerStatusIndicator.String()]
	amtStatusIndicator := statusv1.StatusIndication_value[host.AmtStatusIndicator.String()]

	protoHost := &computev1.HostResource{
		ResourceId:                  host.ResourceID,
		Kind:                        host.Kind,
		Name:                        host.Name,
		Note:                        host.Note,
		DesiredState:                computev1.HostState(desiredState),
		CurrentState:                computev1.HostState(currentState),
		HostStatus:                  host.HostStatus,
		HostStatusIndicator:         statusv1.StatusIndication(hostStatusIndicator),
		HostStatusTimestamp:         host.HostStatusTimestamp,
		OnboardingStatus:            host.OnboardingStatus,
		OnboardingStatusIndicator:   statusv1.StatusIndication(onboardingStatusIndicator),
		OnboardingStatusTimestamp:   host.OnboardingStatusTimestamp,
		RegistrationStatus:          host.RegistrationStatus,
		RegistrationStatusIndicator: statusv1.StatusIndication(registrationStatusIndicator),
		RegistrationStatusTimestamp: host.RegistrationStatusTimestamp,
		HardwareKind:                host.HardwareKind,
		SerialNumber:                host.SerialNumber,
		Uuid:                        host.UUID,
		MemoryBytes:                 host.MemoryBytes,
		CpuModel:                    host.CPUModel,
		CpuSockets:                  host.CPUSockets,
		CpuCores:                    host.CPUCores,
		CpuCapabilities:             host.CPUCapabilities,
		CpuArchitecture:             host.CPUArchitecture,
		CpuThreads:                  host.CPUThreads,
		CpuTopology:                 host.CPUTopology,
		MgmtIp:                      host.MgmtIP,
		BmcKind:                     computev1.BaremetalControllerKind(bcmKind),
		BmcIp:                       host.BmcIP,
		BmcUsername:                 host.BmcUsername,
		BmcPassword:                 host.BmcPassword,
		PxeMac:                      host.PxeMAC,
		Hostname:                    host.Hostname,
		ProductName:                 host.ProductName,
		BiosVersion:                 host.BiosVersion,
		BiosReleaseDate:             host.BiosReleaseDate,
		BiosVendor:                  host.BiosVendor,
		Metadata:                    host.Metadata,
		DesiredPowerState:           computev1.PowerState(desiredPowerState),
		CurrentPowerState:           computev1.PowerState(currentPowerState),
		PowerStatus:                 host.PowerStatus,
		PowerStatusIndicator:        statusv1.StatusIndication(powerStatusIndicator),
		PowerStatusTimestamp:        host.PowerStatusTimestamp,
		PowerCommandPolicy:          computev1.PowerCommandPolicy(powerCommandPolicy),
		PowerOnTime:                 host.PowerOnTime,
		AmtSku:                      computev1.AmtSku(amtSku),
		AmtControlMode:              computev1.AmtControlMode(AmtControlMode),
		DesiredAmtState:             computev1.AmtState(desiredAmtState),
		CurrentAmtState:             computev1.AmtState(currentAmtState),
		AmtStatus:                   host.AmtStatus,
		AmtDnsSuffix:                host.AmtDNSSuffix,
		AmtStatusIndicator:          statusv1.StatusIndication(amtStatusIndicator),
		AmtStatusTimestamp:          host.AmtStatusTimestamp,
		UserLvmSize:                 host.UserLvmSize,
		TenantId:                    host.TenantID,
		CreatedAt:                   host.CreatedAt,
		UpdatedAt:                   host.UpdatedAt,
	}
	// Convert the edges recursively.
	if site, qerr := host.Edges.SiteOrErr(); qerr == nil {
		protoHost.Site = entSiteResourceToProtoSiteResource(site)
	}

	if provider, qerr := host.Edges.ProviderOrErr(); qerr == nil {
		protoHost.Provider = entProviderResourceToProtoProviderResource(provider)
	}
	if storages, qerr := host.Edges.HostStoragesOrErr(); qerr == nil {
		for _, s := range storages {
			protoHost.HostStorages = append(protoHost.HostStorages, entHostStorageResourceToProtoHostStorageResource(s))
		}
	}
	if nics, qerr := host.Edges.HostNicsOrErr(); qerr == nil {
		for _, n := range nics {
			protoHost.HostNics = append(protoHost.HostNics, entHostnicResourceToProtoHostnicResource(n))
		}
	}
	if usbs, qerr := host.Edges.HostUsbsOrErr(); qerr == nil {
		for _, u := range usbs {
			protoHost.HostUsbs = append(protoHost.HostUsbs, entHostusbResourceToProtoHostusbResource(u))
		}
	}
	if device, qerr := host.Edges.HostDeviceOrErr(); qerr == nil {
		protoHost.HostDevice = entHostDeviceResourceToProtoHostDeviceResource(device)
	}
	if inst, qerr := host.Edges.InstanceOrErr(); qerr == nil {
		protoHost.Instance = entInstanceResourceToProtoInstanceResource(inst)
	}
	if gpus, qerr := host.Edges.HostGpusOrErr(); qerr == nil {
		for _, i := range gpus {
			protoHost.HostGpus = append(protoHost.HostGpus, entHostgpuResourceToProtoHostgpuResource(i))
		}
	}
	return protoHost
}

func entHostnicResourceToProtoHostnicResource(hostnic *ent.HostnicResource) *computev1.HostnicResource {
	if hostnic == nil {
		return nil
	}
	// Convert the fields directly.
	linkState := computev1.NetworkInterfaceLinkState_value[hostnic.LinkState.String()]
	protoNic := &computev1.HostnicResource{
		ResourceId:          hostnic.ResourceID,
		Kind:                hostnic.Kind,
		ProviderStatus:      hostnic.ProviderStatus,
		DeviceName:          hostnic.DeviceName,
		PciIdentifier:       hostnic.PciIdentifier,
		MacAddr:             hostnic.MACAddr,
		SriovEnabled:        hostnic.SriovEnabled,
		SriovVfsNum:         hostnic.SriovVfsNum,
		SriovVfsTotal:       hostnic.SriovVfsTotal,
		PeerName:            hostnic.PeerName,
		PeerDescription:     hostnic.PeerDescription,
		PeerMac:             hostnic.PeerMAC,
		PeerMgmtIp:          hostnic.PeerMgmtIP,
		PeerPort:            hostnic.PeerPort,
		SupportedLinkMode:   hostnic.SupportedLinkMode,
		AdvertisingLinkMode: hostnic.AdvertisingLinkMode,
		CurrentSpeedBps:     hostnic.CurrentSpeedBps,
		CurrentDuplex:       hostnic.CurrentDuplex,
		Features:            hostnic.Features,
		LinkState:           computev1.NetworkInterfaceLinkState(linkState),
		Mtu:                 hostnic.Mtu,
		BmcInterface:        hostnic.BmcInterface,
		TenantId:            hostnic.TenantID,
		CreatedAt:           hostnic.CreatedAt,
		UpdatedAt:           hostnic.UpdatedAt,
	}
	if host, qerr := hostnic.Edges.HostOrErr(); qerr == nil {
		protoNic.Host = entHostResourceToProtoHostResource(host)
	}

	return protoNic
}

func entHostusbResourceToProtoHostusbResource(hostusb *ent.HostusbResource) *computev1.HostusbResource {
	if hostusb == nil {
		return nil
	}
	// Convert the fields directly.
	protoUsb := &computev1.HostusbResource{
		ResourceId: hostusb.ResourceID,
		Kind:       hostusb.Kind,
		OwnerId:    hostusb.OwnerID,
		Idvendor:   hostusb.Idvendor,
		Idproduct:  hostusb.Idproduct,
		Bus:        hostusb.Bus,
		Addr:       hostusb.Addr,
		Class:      hostusb.Class,
		Serial:     hostusb.Serial,
		DeviceName: hostusb.DeviceName,
		TenantId:   hostusb.TenantID,
		CreatedAt:  hostusb.CreatedAt,
		UpdatedAt:  hostusb.UpdatedAt,
	}
	// Convert the edges recursively.
	if host, qerr := hostusb.Edges.HostOrErr(); qerr == nil {
		protoUsb.Host = entHostResourceToProtoHostResource(host)
	}

	return protoUsb
}

func entHostgpuResourceToProtoHostgpuResource(hostgpu *ent.HostgpuResource) *computev1.HostgpuResource {
	if hostgpu == nil {
		return nil
	}

	protoGpu := &computev1.HostgpuResource{
		ResourceId:  hostgpu.ResourceID,
		DeviceName:  hostgpu.DeviceName,
		PciId:       hostgpu.PciID,
		Product:     hostgpu.Product,
		Vendor:      hostgpu.Vendor,
		Description: hostgpu.Description,
		Features:    hostgpu.Features,
		TenantId:    hostgpu.TenantID,
		CreatedAt:   hostgpu.CreatedAt,
		UpdatedAt:   hostgpu.UpdatedAt,
	}

	if host, qerr := hostgpu.Edges.HostOrErr(); qerr == nil {
		protoGpu.Host = entHostResourceToProtoHostResource(host)
	}

	return protoGpu
}

func entHostStorageResourceToProtoHostStorageResource(hostStorage *ent.HoststorageResource) *computev1.HoststorageResource {
	if hostStorage == nil {
		return nil
	}
	// Convert the fields directly.
	protoHostStorage := &computev1.HoststorageResource{
		ResourceId:     hostStorage.ResourceID,
		Kind:           hostStorage.Kind,
		ProviderStatus: hostStorage.ProviderStatus,
		Wwid:           hostStorage.Wwid,
		Serial:         hostStorage.Serial,
		Vendor:         hostStorage.Vendor,
		Model:          hostStorage.Model,
		CapacityBytes:  hostStorage.CapacityBytes,
		DeviceName:     hostStorage.DeviceName,
		TenantId:       hostStorage.TenantID,
		CreatedAt:      hostStorage.CreatedAt,
		UpdatedAt:      hostStorage.UpdatedAt,
	}
	if host, qerr := hostStorage.Edges.HostOrErr(); qerr == nil {
		protoHostStorage.Host = entHostResourceToProtoHostResource(host)
	}

	return protoHostStorage
}

func entHostDeviceResourceToProtoHostDeviceResource(hostDevice *ent.HostdeviceResource) *computev1.HostdeviceResource {
	if hostDevice == nil {
		return nil
	}
	// Convert the fields directly.
	protoHostDevice := &computev1.HostdeviceResource{
		ResourceId:       hostDevice.ResourceID,
		Version:          hostDevice.Version,
		DeviceName:       hostDevice.DeviceName,
		OperationalState: hostDevice.OperationalState,
		BuildNumber:      hostDevice.BuildNumber,
		Sku:              hostDevice.Sku,
		Features:         hostDevice.Features,
		DeviceGuid:       hostDevice.DeviceGUID,
		ControlMode:      hostDevice.ControlMode,
		DnsSuffix:        hostDevice.DNSSuffix,
		NetworkStatus:    hostDevice.NetworkStatus,
		RemoteStatus:     hostDevice.RemoteStatus,
		RemoteTrigger:    hostDevice.RemoteTrigger,
		MpsHostname:      hostDevice.MpsHostname,
		TenantId:         hostDevice.TenantID,
		CreatedAt:        hostDevice.CreatedAt,
		UpdatedAt:        hostDevice.UpdatedAt,
	}
	if host, qerr := hostDevice.Edges.HostOrErr(); qerr == nil {
		protoHostDevice.Host = entHostResourceToProtoHostResource(host)
	}

	return protoHostDevice
}

func entRemoteAccessConfigurationToProto(entity *ent.RemoteAccessConfiguration) *remoteaccessv1.RemoteAccessConfiguration {
	if entity == nil {
		return nil
	}
	statusIndicator := statusv1.StatusIndication_value[entity.ConfigurationStatusIndicator.String()]
	desiredState := remoteaccessv1.RemoteAccessState_value[entity.DesiredState.String()]
	currentState := remoteaccessv1.RemoteAccessState_value[entity.CurrentState.String()]
	protoResource := &remoteaccessv1.RemoteAccessConfiguration{
		ConfigurationStatus:          entity.ConfigurationStatus,
		ConfigurationStatusIndicator: statusv1.StatusIndication(statusIndicator),
		ConfigurationStatusTimestamp: entity.ConfigurationStatusTimestamp,
		ExpirationTimestamp:          entity.ExpirationTimestamp,
		LocalPort:                    entity.LocalPort,
		ResourceId:                   entity.ResourceID,
		DesiredState:                 remoteaccessv1.RemoteAccessState(desiredState),
		CurrentState:                 remoteaccessv1.RemoteAccessState(currentState),
		User:                         entity.User,
		TenantId:                     entity.TenantID,
		CreatedAt:                    entity.CreatedAt,
		UpdatedAt:                    entity.UpdatedAt,
	}

	if inst, err := entity.Edges.InstanceOrErr(); err == nil {
		protoResource.Instance = entInstanceResourceToProtoInstanceResource(inst)
	}
	return protoResource
}

func entEndpointResourceToProtoEndpointResource(endpoint *ent.EndpointResource) *network_v1.EndpointResource {
	if endpoint == nil {
		return nil
	}
	// Convert the fields directly.
	protoEndpoint := &network_v1.EndpointResource{
		ResourceId: endpoint.ResourceID,
		Kind:       endpoint.Kind,
		Name:       endpoint.Name,
		TenantId:   endpoint.TenantID,
		CreatedAt:  endpoint.CreatedAt,
		UpdatedAt:  endpoint.UpdatedAt,
	}
	// Convert the edges recursively.
	if host, qerr := endpoint.Edges.HostOrErr(); qerr == nil {
		protoEndpoint.Host = entHostResourceToProtoHostResource(host)
	}

	return protoEndpoint
}

func entNetworkSegmentToProtoNetworkSegmentResource(network *ent.NetworkSegment) *network_v1.NetworkSegment {
	if network == nil {
		return nil
	}
	// Convert the fields directly.
	protoNetwork := &network_v1.NetworkSegment{
		ResourceId: network.ResourceID,
		Name:       network.Name,
		VlanId:     network.VlanID,
		TenantId:   network.TenantID,
		CreatedAt:  network.CreatedAt,
		UpdatedAt:  network.UpdatedAt,
	}
	// Convert the edges recursively.
	if site, qerr := network.Edges.SiteOrErr(); qerr == nil {
		protoNetwork.Site = entSiteResourceToProtoSiteResource(site)
	}

	return protoNetwork
}

func entNetlinkResourceToProtoNetlinkResource(netlink *ent.NetlinkResource) *network_v1.NetlinkResource {
	if netlink == nil {
		return nil
	}
	// Convert the fields directly.
	desiredState := network_v1.NetlinkState_value[netlink.DesiredState.String()] // Defaults to 0 if not found
	currentState := network_v1.NetlinkState_value[netlink.CurrentState.String()] // Defaults to 0 if not found
	protoNetlink := &network_v1.NetlinkResource{
		ResourceId:     netlink.ResourceID,
		Kind:           netlink.Kind,
		Name:           netlink.Name,
		DesiredState:   network_v1.NetlinkState(desiredState),
		CurrentState:   network_v1.NetlinkState(currentState),
		ProviderStatus: netlink.ProviderStatus,
		TenantId:       netlink.TenantID,
		CreatedAt:      netlink.CreatedAt,
		UpdatedAt:      netlink.UpdatedAt,
	}
	// Convert the edges recursively.
	if src, qerr := netlink.Edges.SrcOrErr(); qerr == nil {
		protoNetlink.Src = entEndpointResourceToProtoEndpointResource(src)
	}
	if dst, qerr := netlink.Edges.DstOrErr(); qerr == nil {
		protoNetlink.Dst = entEndpointResourceToProtoEndpointResource(dst)
	}

	return protoNetlink
}

//nolint:cyclop // InstanceResource has many edges that need to be converted.
func entInstanceResourceToProtoInstanceResource(ins *ent.InstanceResource) *computev1.InstanceResource {
	if ins == nil {
		return nil
	}
	// Convert the fields directly.
	desiredState := computev1.InstanceState_value[ins.DesiredState.String()]     // Defaults to 0 if not found
	currentState := computev1.InstanceState_value[ins.CurrentState.String()]     // Defaults to 0 if not found
	inskind := computev1.InstanceKind_value[ins.Kind.String()]                   // Defaults to 0 if not found
	securityFeatures := osv1.SecurityFeature_value[ins.SecurityFeature.String()] // Defaults to 0 if not found
	insStatusIndicator := statusv1.StatusIndication_value[ins.InstanceStatusIndicator.String()]
	updateStatusIndicator := statusv1.StatusIndication_value[ins.UpdateStatusIndicator.String()]
	provisioningStatusIndicator := statusv1.StatusIndication_value[ins.ProvisioningStatusIndicator.String()]
	trustedAttestationStatusIndicator := statusv1.StatusIndication_value[ins.TrustedAttestationStatusIndicator.String()]
	protoInstance := &computev1.InstanceResource{
		ResourceId:                        ins.ResourceID,
		Kind:                              computev1.InstanceKind(inskind),
		Name:                              ins.Name,
		DesiredState:                      computev1.InstanceState(desiredState),
		CurrentState:                      computev1.InstanceState(currentState),
		VmMemoryBytes:                     ins.VMMemoryBytes,
		VmCpuCores:                        ins.VMCPUCores,
		VmStorageBytes:                    ins.VMStorageBytes,
		SecurityFeature:                   osv1.SecurityFeature(securityFeatures),
		InstanceStatus:                    ins.InstanceStatus,
		InstanceStatusIndicator:           statusv1.StatusIndication(insStatusIndicator),
		InstanceStatusTimestamp:           ins.InstanceStatusTimestamp,
		InstanceStatusDetail:              ins.InstanceStatusDetail,
		UpdateStatus:                      ins.UpdateStatus,
		UpdateStatusIndicator:             statusv1.StatusIndication(updateStatusIndicator),
		UpdateStatusTimestamp:             ins.UpdateStatusTimestamp,
		ProvisioningStatus:                ins.ProvisioningStatus,
		ProvisioningStatusIndicator:       statusv1.StatusIndication(provisioningStatusIndicator),
		ProvisioningStatusTimestamp:       ins.ProvisioningStatusTimestamp,
		TrustedAttestationStatus:          ins.TrustedAttestationStatus,
		TrustedAttestationStatusIndicator: statusv1.StatusIndication(trustedAttestationStatusIndicator),
		TrustedAttestationStatusTimestamp: ins.TrustedAttestationStatusTimestamp,
		ExistingCves:                      ins.ExistingCves,
		RuntimePackages:                   ins.RuntimePackages,
		OsUpdateAvailable:                 ins.OsUpdateAvailable,
		TenantId:                          ins.TenantID,
		CreatedAt:                         ins.CreatedAt,
		UpdatedAt:                         ins.UpdatedAt,
	}
	// Convert the edges recursively.
	if host, qerr := ins.Edges.HostOrErr(); qerr == nil {
		protoInstance.Host = entHostResourceToProtoHostResource(host)
	}
	if os, qerr := ins.Edges.OsOrErr(); qerr == nil {
		protoInstance.Os = entOperatingSystemResourceToProtoOperatingSystemResource(os)
	}
	if wMembers, qerr := ins.Edges.WorkloadMembersOrErr(); qerr == nil {
		for _, m := range wMembers {
			protoInstance.WorkloadMembers = append(protoInstance.WorkloadMembers, entWorkloadMemberToProtoWorkloadMember(m))
		}
	}
	if provider, qerr := ins.Edges.ProviderOrErr(); qerr == nil {
		protoInstance.Provider = entProviderResourceToProtoProviderResource(provider)
	}
	if localaccount, qerr := ins.Edges.LocalaccountOrErr(); qerr == nil {
		protoInstance.Localaccount = entLocalAccountResourceToProtoLocalAccountResource(localaccount)
	}
	if osUpdatePolicy, qerr := ins.Edges.OsUpdatePolicyOrErr(); qerr == nil {
		protoInstance.OsUpdatePolicy = entOSUpdatePolicyResourceToProtoOSUpdatePolicyResource(osUpdatePolicy)
	}
	if customConfig, qerr := ins.Edges.CustomConfigOrErr(); qerr == nil {
		for _, m := range customConfig {
			protoInstance.CustomConfig = append(protoInstance.CustomConfig, entCustomConfigResourceToProtoCustomConfigResource(m))
		}
	}
	return protoInstance
}

func entWorkloadResourceToProtoWorkloadResource(workload *ent.WorkloadResource) *computev1.WorkloadResource {
	if workload == nil {
		return nil
	}
	// Convert the fields directly.
	kind := computev1.WorkloadKind_value[workload.Kind.String()]
	desiredState := computev1.WorkloadState_value[workload.DesiredState.String()]
	currentState := computev1.WorkloadState_value[workload.CurrentState.String()]
	protoWorkload := &computev1.WorkloadResource{
		ResourceId:   workload.ResourceID,
		Kind:         computev1.WorkloadKind(kind),
		Name:         workload.Name,
		DesiredState: computev1.WorkloadState(desiredState),
		CurrentState: computev1.WorkloadState(currentState),
		Status:       workload.Status,
		Metadata:     workload.Metadata,
		ExternalId:   workload.ExternalID,
		TenantId:     workload.TenantID,
		CreatedAt:    workload.CreatedAt,
		UpdatedAt:    workload.UpdatedAt,
	}
	// Convert the edges recursively.
	if members, qerr := workload.Edges.MembersOrErr(); qerr == nil {
		for _, m := range members {
			protoWorkload.Members = append(protoWorkload.Members, entWorkloadMemberToProtoWorkloadMember(m))
		}
	}

	return protoWorkload
}

func entWorkloadMemberToProtoWorkloadMember(member *ent.WorkloadMember) *computev1.WorkloadMember {
	if member == nil {
		return nil
	}
	// Convert the fields directly.
	kind := computev1.WorkloadMemberKind_value[member.Kind.String()]
	protoMember := &computev1.WorkloadMember{
		ResourceId: member.ResourceID,
		Kind:       computev1.WorkloadMemberKind(kind),
		TenantId:   member.TenantID,
		CreatedAt:  member.CreatedAt,
		UpdatedAt:  member.UpdatedAt,
	}
	// Convert the edges recursively.
	if workload, qerr := member.Edges.WorkloadOrErr(); qerr == nil {
		protoMember.Workload = entWorkloadResourceToProtoWorkloadResource(workload)
	}
	if instance, qerr := member.Edges.InstanceOrErr(); qerr == nil {
		protoMember.Instance = entInstanceResourceToProtoInstanceResource(instance)
	}

	return protoMember
}

func entIPAddressResourceToProtoIPAddressResource(ipaddress *ent.IPAddressResource) *network_v1.IPAddressResource {
	if ipaddress == nil {
		return nil
	}
	// Convert the fields directly.
	desiredState := network_v1.IPAddressState_value[ipaddress.DesiredState.String()]
	currentState := network_v1.IPAddressState_value[ipaddress.CurrentState.String()]
	configMethod := network_v1.IPAddressConfigMethod_value[ipaddress.ConfigMethod.String()]
	status := network_v1.IPAddressStatus_value[ipaddress.Status.String()]
	protoIPAddress := &network_v1.IPAddressResource{
		ResourceId:   ipaddress.ResourceID,
		Address:      ipaddress.Address,
		DesiredState: network_v1.IPAddressState(desiredState),
		CurrentState: network_v1.IPAddressState(currentState),
		Status:       network_v1.IPAddressStatus(status),
		StatusDetail: ipaddress.StatusDetail,
		ConfigMethod: network_v1.IPAddressConfigMethod(configMethod),
		TenantId:     ipaddress.TenantID,
		CreatedAt:    ipaddress.CreatedAt,
		UpdatedAt:    ipaddress.UpdatedAt,
	}
	// Convert the edges recursively.
	if nic, qerr := ipaddress.Edges.NicOrErr(); qerr == nil {
		protoIPAddress.Nic = entHostnicResourceToProtoHostnicResource(nic)
	}

	return protoIPAddress
}

func entOSUpdatePolicyResourceToProtoOSUpdatePolicyResource(osup *ent.OSUpdatePolicyResource,
) *computev1.OSUpdatePolicyResource {
	if osup == nil {
		return nil
	}
	// Convert the fields directly.
	updatePolicy := computev1.UpdatePolicy_value[osup.UpdatePolicy.String()] // Defaults to 0 if not found
	protoOsUpdatePolicy := &computev1.OSUpdatePolicyResource{
		Name:                osup.Name,
		Description:         osup.Description,
		ResourceId:          osup.ResourceID,
		UpdatePackages:      osup.UpdatePackages,
		UpdateKernelCommand: osup.UpdateKernelCommand,
		UpdatePolicy:        computev1.UpdatePolicy(updatePolicy),
		TenantId:            osup.TenantID,
		CreatedAt:           osup.CreatedAt,
		UpdatedAt:           osup.UpdatedAt,
	}
	if osup.UpdateSources != "" {
		protoOsUpdatePolicy.UpdateSources = strings.Split(osup.UpdateSources, "|")
	}
	// Convert the edges recursively.
	if os, err := osup.Edges.TargetOsOrErr(); err == nil {
		protoOsUpdatePolicy.TargetOs = entOperatingSystemResourceToProtoOperatingSystemResource(os)
	}
	return protoOsUpdatePolicy
}

func entCustomConfigResourceToProtoCustomConfigResource(
	customconfig *ent.CustomConfigResource,
) *computev1.CustomConfigResource {
	if customconfig == nil {
		return nil
	}
	protoLocalAccount := &computev1.CustomConfigResource{
		ResourceId:  customconfig.ResourceID,
		Name:        customconfig.Name,
		Description: customconfig.Description,
		Config:      customconfig.Config,
		TenantId:    customconfig.TenantID,
		CreatedAt:   customconfig.CreatedAt,
	}
	return protoLocalAccount
}

func entOSUpdateRunResourceToProtoOSUpdateRunResource(osrun *ent.OSUpdateRunResource) *computev1.OSUpdateRunResource {
	if osrun == nil {
		return nil
	}
	// Convert the fields directly.
	statusIndicator := statusv1.StatusIndication_value[osrun.StatusIndicator.String()] // Defaults to 0 if not found
	protoOsUpdateRun := &computev1.OSUpdateRunResource{
		Name:            osrun.Name,
		Description:     osrun.Description,
		ResourceId:      osrun.ResourceID,
		Status:          osrun.Status,
		StatusDetails:   osrun.StatusDetails,
		StatusTimestamp: osrun.StatusTimestamp,
		StartTime:       osrun.StartTime,
		EndTime:         osrun.EndTime,
		StatusIndicator: statusv1.StatusIndication(statusIndicator),
		TenantId:        osrun.TenantID,
		CreatedAt:       osrun.CreatedAt,
		UpdatedAt:       osrun.UpdatedAt,
	}
	// Convert the edges recursively.
	if osPolicy, err := osrun.Edges.AppliedPolicyOrErr(); err == nil {
		protoOsUpdateRun.AppliedPolicy = entOSUpdatePolicyResourceToProtoOSUpdatePolicyResource(osPolicy)
	}
	if instance, err := osrun.Edges.InstanceOrErr(); err == nil {
		protoOsUpdateRun.Instance = entInstanceResourceToProtoInstanceResource(instance)
	}
	return protoOsUpdateRun
}
