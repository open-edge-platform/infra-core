// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	network_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	os_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	ou_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/ou/v1"
	providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	schedule_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	telemetryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

var (
	metadataResource   = `[{"key":"key1", "value":"region_key1_lvl1"}, {"key":"key2", "value":"region_key2_lvl1"}, {"key":"key3", "value":"region_key3_lvl1"}]`
	SingleSchedResID   = "singlesche-12345678"
	RepeatedSchedResID = "repeatedsche-12345678"

	repeatedscheduleResource = &schedule_v1.RepeatedScheduleResource{
		Name: "for unit testing purposes",
		Relation: &schedule_v1.RepeatedScheduleResource_TargetSite{
			TargetSite: siteResource,
		},
		ScheduleStatus:  schedule_v1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
		DurationSeconds: uint32(100),
		CronMinutes:     "3",
		CronHours:       "4",
		CronDayMonth:    "5",
		CronMonth:       "6",
		CronDayWeek:     "0",
	}

	singlescheduleResource = &schedule_v1.SingleScheduleResource{
		Name: "for unit testing purposes",
		Relation: &schedule_v1.SingleScheduleResource_TargetHost{
			TargetHost: hostResource,
		},
		ScheduleStatus: schedule_v1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
		StartSeconds:   1686240123,
		EndSeconds:     1686240599,
	}

	osrResource = &os_v1.OperatingSystemResource{
		Name:     "for unit testing purposes",
		ImageUrl: "Repo URL Test",
	}

	parentOUResource = &ou_v1.OuResource{
		Name:     "for unit testing purposes",
		OuKind:   "test BU",
		Metadata: metadataResource,
	}

	ouResource = &ou_v1.OuResource{
		Name:     "for unit testing purposes",
		OuKind:   "test BU",
		ParentOu: parentOUResource,
		Metadata: metadataResource,
	}

	parentRegionResource = &location_v1.RegionResource{
		Name:       "for unit testing purposes",
		RegionKind: "test region",
		Metadata:   metadataResource,
	}

	regionResource = &location_v1.RegionResource{
		Name:         "for unit testing purposes",
		RegionKind:   "test region",
		ParentRegion: parentRegionResource,
		Metadata:     metadataResource,
	}

	siteResource = &location_v1.SiteResource{
		Name:             "for unit testing purposes",
		Region:           regionResource,
		Ou:               ouResource,
		Address:          "",
		SiteLat:          0,
		SiteLng:          0,
		DnsServers:       []string{},
		DockerRegistries: []string{},
		MetricsEndpoint:  "",
		HttpProxy:        "",
		HttpsProxy:       "",
		FtpProxy:         "",
		NoProxy:          "",
	}

	hostResource = &computev1.HostResource{
		Name:         "for unit testing purposes",
		DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,

		Site: siteResource,

		HardwareKind: "XDgen2",
		SerialNumber: "12345678",
		Uuid:         "E5E53D99-708D-4AF5-8378-63880FF62712",
		MemoryBytes:  64 * util.Gigabyte,

		CpuModel:        "12th Gen Intel(R) Core(TM) i9-12900",
		CpuSockets:      1,
		CpuCores:        14,
		CpuCapabilities: "",
		CpuArchitecture: "x86_64",
		CpuThreads:      10,

		MgmtIp: "192.168.10.10",

		BmcKind:     computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
		BmcIp:       "10.0.0.10",
		BmcUsername: "user",
		BmcPassword: "pass",
		PxeMac:      "90:49:fa:ff:ff:ff",

		Hostname: "testhost1",

		DesiredPowerState: computev1.PowerState_POWER_STATE_ON,

		HostNics: []*computev1.HostnicResource{
			hostResourceNic,
			hostResourceNicNoIp,
		},
		HostGpus: []*computev1.HostgpuResource{
			hostResourceGpu,
		},
	}

	hostResourceNic = &computev1.HostnicResource{
		ResourceId:   "hostnic-12345678",
		DeviceName:   "eth0",
		MacAddr:      "00:11:22:33:44:55",
		Mtu:          1500,
		LinkState:    computev1.NetworkInterfaceLinkState_NETWORK_INTERFACE_LINK_STATE_UP,
		BmcInterface: false,
	}
	hostResourceNicNoIp = &computev1.HostnicResource{
		ResourceId:   "hostnic-12345679",
		DeviceName:   "eth1",
		MacAddr:      "00:11:22:33:44:66",
		Mtu:          1500,
		LinkState:    computev1.NetworkInterfaceLinkState_NETWORK_INTERFACE_LINK_STATE_DOWN,
		BmcInterface: false,
	}
	ipAddressResource = &network_v1.IPAddressResource{
		ResourceId:   "ipaddr-12345678",
		Address:      "10.0.0.1/24",
		DesiredState: network_v1.IPAddressState_IP_ADDRESS_STATE_CONFIGURED,
		Status:       network_v1.IPAddressStatus_IP_ADDRESS_STATUS_CONFIGURED,
		StatusDetail: "Specifically I am fine",
		ConfigMethod: network_v1.IPAddressConfigMethod_IP_ADDRESS_CONFIG_METHOD_DYNAMIC,
		Nic:          hostResourceNic,
	}

	GPUDescription  = "some desc"
	GPUModel        = "Model XYZ"
	GPUVendor       = "Intel"
	GPUPciID        = "00:00.1"
	GPUName         = "gpu0"
	GPUResourceID   = "hostgpu-12345678"
	hostResourceGpu = &computev1.HostgpuResource{
		ResourceId:  GPUResourceID,
		DeviceName:  GPUName,
		PciId:       GPUPciID,
		Product:     GPUModel,
		Vendor:      GPUVendor,
		Description: GPUDescription,
	}

	hostResourceStorage    = &computev1.HoststorageResource{}
	hostResourceUsb        = &computev1.HostusbResource{}
	workloadResource       = &computev1.WorkloadResource{}
	workloadMember         = &computev1.WorkloadMember{}
	instanceResource       = &computev1.InstanceResource{}
	telemetryGroupResource = &telemetryv1.TelemetryGroupResource{}
	telemetryProfile       = &telemetryv1.TelemetryProfile{}
	providerResource       = &providerv1.ProviderResource{}
)
