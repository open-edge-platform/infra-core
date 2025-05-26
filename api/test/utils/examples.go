// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
)

var (
	UserNameOne   = "userone"
	UserNameTwo   = "usertwo"
	UserUnexistID = "user-00000000"
	UserWrongID   = "User-123"

	UserPubKey    = "empty"
	UserPubKeyTwo = "empty-too"

	Host1Name            = "Host-One"
	Host2Name            = "Host-Two"
	Host2bName           = "Host-TwoB"
	Host3Name            = "Host-Three"
	Host4Name            = "Host-Four"
	HostUnexistID        = "host-00000000"
	HostWrongID          = "HOST-123"
	HostNameNonPrintable = "0x73t116 0x74r114 0x67…&#8230 \u2026⟶​9 0x65U+200B&#8203; \u200Bh104"
	HostGUIDNonPrintable = "\u2026⟶​9 0x65U+200B&#8203; \u200Bh104"

	Region1Name     = "region-12345678"
	Region2Name     = "region-23456789"
	Region3Name     = "region-00000003"
	RegionUnexistID = "region-00000000"
	RegionWrongID   = "REGION-123"

	Site1Name     = "site-12345678"
	Site2Name     = "site-12345679"
	Site3Name     = "site-12345670"
	SiteUnexistID = "site-00000000"
	SiteWrongID   = "SITE-123"

	OU1Name     = "ou-12345678"
	OU2Name     = "ou-12345679"
	OU3Name     = "ou-12345670"
	OUUnexistID = "ou-00000000"
	OUWrongID   = "OU-1234"
	EmptyString = ""

	SschedName1 = "singleSched1"
	SschedName2 = "singleSched3"
	SschedName3 = "singleSched3"

	now            = int(time.Now().Unix())
	FutureEpoch    = time.Unix(int64(now), 0).Add(1801 * time.Second)
	SschedStart1   = now + 1800
	SschedStart2   = now + 1800
	SschedStart3   = now + 1800
	SschedEnd1     = now + 3600
	SschedEndError = now - 1800

	cronDayMonth = "10"
	CronAny      = "*"

	SingleScheduleWrongID   = "singlesche-XXXXXX"
	SingleScheduleUnexistID = "singlesche-12345678"

	RepeatedScheduleWrongID   = "repeatedsche-XXXXXX"
	RepeatedScheduleUnexistID = "repeatedsche-12345678"

	OSName1             = "OSName1"
	OSName2             = "OSName2"
	OSName3             = "OSName3"
	OSArch1             = "x86_64"
	OSArch2             = "arch2"
	OSArch3             = "arch3"
	OSKernel1           = "k1"
	OSRepo1             = "OSRepo1"
	OSRepo2             = "OSRepo2"
	OSRepo3             = "OSRepo3"
	OSProfileName1      = "Test OS profile"
	OSInstalledPackages = "intel-opencl-icd\nintel-level-zero-gpu\nlevel-zero"
	OSSecurityFeature1  = api.SECURITYFEATURENONE
	OSSecurityFeature2  = api.SECURITYFEATURESECUREBOOTANDFULLDISKENCRYPTION
	OSTypeImmutable     = api.OPERATINGSYSTEMTYPEIMMUTABLE
	OSProviderInfra     = api.OPERATINGSYSTEMPROVIDERINFRA

	OSResourceWrongID   = "os-XXXXXX"
	OSResourceUnexistID = "os-00000000"

	WorkloadName1   = "WorkloadName1"
	WorkloadStatus1 = "WorkloadStatus1"
	WorkloadName2   = "WorkloadName2"
	WorkloadStatus2 = "WorkloadStatus2"
	WorkloadStatus3 = "WorkloadStatus3"

	WorkloadUnexistID       = "workload-00000000"
	WorkloadWrongID         = "workload-XXXXXX"
	WorkloadMemberUnexistID = "workloadmember-00000000"
	WorkloadMemberWrongID   = "workloadmember-XXXXXX"

	InstanceUnexistID   = "inst-00000000"
	InstanceWrongID     = "inst-XXXXXXXX"
	Inst1Name           = "inst1Name"
	Inst2Name           = "inst2Name"
	inst3Name           = "inst3Name"
	inst4Name           = "inst4Name"
	instHostID          = ""
	instOSID            = ""
	instKind            = api.INSTANCEKINDMETAL
	instSecurityFeature = api.SECURITYFEATURESECUREBOOTANDFULLDISKENCRYPTION

	providerKind1           = api.PROVIDERKINDBAREMETAL
	providerVendor1         = api.PROVIDERVENDORLENOVOLXCA
	ProviderName1           = "SC LXCA"
	providerApiEndpoint1    = "https://192.168.201.3/"
	providerApiCredentials1 = []string{"v1/lxca/user", "v1/lxca/password"}
	providerConfig1         = "Some config string"

	providerKind2           = api.PROVIDERKINDBAREMETAL
	providerVendor2         = api.PROVIDERVENDORLENOVOLOCA
	ProviderName2           = "SC LOCA"
	providerApiEndpoint2    = "https://192.168.202.4/"
	providerApiCredentials2 = []string{"v1/loca/user-admin", "v1/loca/password-pass"}

	providerKind3        = api.PROVIDERKINDBAREMETAL
	ProviderName3        = "Intel"
	providerApiEndpoint3 = "https://192.168.204.4/"

	ProviderUnexistID         = "provider-00000000"
	ProviderWrongID           = "proider-12345678"
	providerBadApiCredentials = []string{"%as", "v1/lxca/password"}

	MetadataOU1 = api.Metadata{
		{
			Key:   "examplekey",
			Value: "ou1",
		}, {
			Key:   "examplekey2",
			Value: "ou1",
		},
	}
	MetadataOU2 = api.Metadata{
		{
			Key:   "examplekey",
			Value: "ou2",
		}, {
			Key:   "examplekey2",
			Value: "ou2",
		},
	}
	MetadataOU3 = api.Metadata{
		{
			Key:   "examplekey2",
			Value: "ou3",
		},
		{
			Key:   "examplekey3",
			Value: "ou3",
		},
	}

	MetadataOU3Rendered = api.Metadata{
		{
			Key:   "examplekey",
			Value: "ou2",
		},
	}

	MetadataR1 = api.Metadata{
		{
			Key:   "examplekey",
			Value: "r1",
		}, {
			Key:   "examplekey2",
			Value: "r1",
		},
	}
	MetadataR2 = api.Metadata{
		{
			Key:   "examplekey",
			Value: "r2",
		}, {
			Key:   "examplekey2",
			Value: "r2",
		},
	}
	MetadataR3 = api.Metadata{
		{
			Key:   "examplekey",
			Value: "r3",
		},
	}
	MetadataR3Inherited = api.Metadata{
		{
			Key:   "examplekey2",
			Value: "r2",
		},
	}

	MetadataSite2 = api.Metadata{
		{
			Key:   "examplekey2",
			Value: "site1",
		},
	}

	MetadataHost1 = api.Metadata{
		{
			Key:   "examplekey1",
			Value: "host1",
		},
	}

	MetadataHost2 = api.Metadata{
		{
			Key:   "examplekey1",
			Value: "host2",
		},
		{
			Key:   "examplekey3",
			Value: "host2",
		},
	}

	MetadataRightPattern = api.Metadata{
		{
			Key:   "asd/ad.123",
			Value: "site1-.ad",
		},
		{
			Key:   "city",
			Value: "test-region",
		},
	}

	MetadataWrongPattern = api.Metadata{
		{
			Key:   "/examplekey2",
			Value: "-site1",
		},
	}

	Host1UUID1, _       = uuid.Parse("BFD3B398-9A4B-480D-AB53-4050ED108F5C")
	Host4UUID1, _       = uuid.Parse("BFD3B398-9A4C-481D-AB53-4050ED108F5D")
	Host1UUIDPatched, _ = uuid.Parse("BFD3B398-9A4B-480D-AB53-4050ED108F5E")
	HostUUIDUnexists, _ = uuid.Parse("BFD3B398-9A4B-480D-AB53-4050ED108F5F")
	HostUUIDError       = "BFD3B398-9A4B-480D-AB53-4050ED108F5FKK"
	Host2UUID           = uuid.New()
	Host3UUID           = uuid.New()
	Host5UUID           = uuid.New()

	HostSerialNumber1 = "SN001"
	HostSerialNumber2 = "SN002"
	HostSerialNumber3 = "SN003"

	DnsServers = []string{"10.10.10.10"}

	Region1Request = api.Region{
		Name:     &Region1Name,
		Metadata: &MetadataR1,
	}

	Region1RequestWrong = api.Region{}

	Region1RequestMetadataOK = api.Region{
		Name:     &Region1Name,
		Metadata: &MetadataRightPattern,
	}

	Region1RequestMetadataNOK = api.Region{
		Name:     &Region1Name,
		Metadata: &MetadataWrongPattern,
	}

	Region2Request = api.Region{
		Name:     &Region2Name,
		Metadata: &MetadataR2,
	}

	Region3Request = api.Region{
		Name:     &Region3Name,
		Metadata: &MetadataR3,
	}

	OU1Request = api.OU{
		Name:     OU1Name,
		Metadata: &MetadataOU1,
	}

	OU2Request = api.OU{
		Name:     OU2Name,
		Metadata: &MetadataOU2,
	}

	OU3Request = api.OU{
		Name:     OU3Name,
		Metadata: &MetadataOU3,
	}

	SiteListRequest = api.Site{
		Name:       &Site1Name,
		DnsServers: &DnsServers,
	}

	SiteListRequest1 = api.Site{
		Name:       &Site1Name,
		DnsServers: &DnsServers,
	}

	SiteListRequest2 = api.Site{
		Name:       &Site2Name,
		DnsServers: &DnsServers,
	}

	SiteListRequest3 = api.Site{
		Name:       &Site3Name,
		DnsServers: &DnsServers,
	}

	Site1Request = api.Site{
		Name:       &Site1Name,
		DnsServers: &DnsServers,
	}

	Site1RequestUpdate = api.Site{
		Name:       &Site1Name,
		DnsServers: &DnsServers,
	}

	Site1RequestUpdatePatch = api.Site{
		Name:       &Site2Name,
		DnsServers: &DnsServers,
	}
	Site2Request = api.Site{
		Name:       &Site2Name,
		DnsServers: &DnsServers,
		Metadata:   &MetadataSite2,
	}
	Site3Request = api.Site{
		Name:       &Site3Name,
		DnsServers: &DnsServers,
		Metadata:   &MetadataSite2,
	}

	metadata = api.Metadata{
		{
			Key:   "examplekey",
			Value: "examplevalue",
		}, {
			Key:   "examplekey2",
			Value: "examplevalue2",
		},
	}

	metadata1 = api.Metadata{
		{
			Key:   "filtermetakey1",
			Value: "filtermetavalue1",
		}, {
			Key:   "filtermetakey2",
			Value: "filtermetavalue2",
		},
	}

	metadata2 = api.Metadata{
		{
			Key:   "filtermetakey1",
			Value: "filtermetavalue1",
		}, {
			Key:   "filtermetakey2",
			Value: "filtermetavalue2_mod",
		},
	}

	Host1Request = api.Host{
		Name:     Host1Name,
		Metadata: &metadata,
		Uuid:     &Host1UUID1,
	}

	Host1RequestPut = api.Host{
		Name:     Host1Name,
		Metadata: &metadata,
	}

	Host1RequestUpdate = api.Host{
		Metadata: &api.Metadata{
			{
				Key:   "examplekey",
				Value: "examplevalue",
			}, {
				Key:   "examplekey2",
				Value: "examplevalue2",
			},
		},
		Name:   Host2Name,
		SiteId: &Site2Name,
	}

	powerHostON               = api.POWERSTATEON
	Host1RequestUpdatePowerON = api.Host{
		DesiredPowerState: &powerHostON,
	}

	powerHostOFF               = api.POWERSTATEOFF
	Host1RequestUpdatePowerOFF = api.Host{
		DesiredPowerState: &powerHostOFF,
	}

	Host1RequestPatch = api.Host{
		Metadata: &api.Metadata{
			{
				Key:   "examplekey",
				Value: "examplevalue",
			}, {
				Key:   "examplekey2",
				Value: "examplevalue2",
			},
		},
		Name:   Host3Name,
		SiteId: &Site2Name,
	}

	Host2Request = api.Host{
		Name:     Host2Name,
		Metadata: &metadata,
		Uuid:     &Host2UUID,
	}

	HostReqFilterMeta1 = api.Host{
		Metadata: &metadata1,
		Uuid:     &Host1UUID1,
	}
	HostReqFilterMeta2 = api.Host{
		Metadata: &metadata2,
		Uuid:     &Host2UUID,
	}

	Host3Request = api.Host{
		Metadata: &api.Metadata{
			{
				Key:   "examplekey",
				Value: "examplevalue",
			}, {
				Key:   "examplekey2",
				Value: "examplevalue2",
			},
		},
		Name:   Host1Name,
		SiteId: &Site1Name,
		Uuid:   &Host3UUID,
	}

	Host4Request = api.Host{
		Name: Host4Name,
		Uuid: &Host4UUID1,
	}

	Host4RequestPut = api.Host{
		Metadata: &api.Metadata{
			{
				Key:   "examplekey",
				Value: "examplevalue",
			}, {
				Key:   "examplekey2",
				Value: "examplevalue2",
			},
		},
		Name:   Host4Name,
		SiteId: &Site1Name,
	}

	Host4RequestPatch = api.Host{
		Metadata: &api.Metadata{
			{
				Key:   "examplekey",
				Value: "examplevalue",
			}, {
				Key:   "examplekey2",
				Value: "examplevalue2",
			},
		},
		Name:   Host4Name,
		SiteId: &Site1Name,
	}

	HostNonPrintable = api.Host{
		Metadata: &api.Metadata{
			{
				Key:   "examplekey",
				Value: "examplevalue",
			}, {
				Key:   "examplekey2",
				Value: "examplevalue2",
			},
		},
		Name:   HostNameNonPrintable,
		SiteId: &Site1Name,
		Uuid:   &Host1UUID1,
	}

	HostRegister = api.HostRegisterInfo{
		Name:         &Host1Name,
		Uuid:         &Host2UUID,
		SerialNumber: &HostSerialNumber1,
	}

	AutoOnboardTrue         bool = true
	AutoOnboardFalse        bool = false
	HostRegisterAutoOnboard      = api.HostRegisterInfo{
		Name:         &Host2Name,
		Uuid:         &Host3UUID,
		SerialNumber: &HostSerialNumber2,
		AutoOnboard:  &AutoOnboardTrue,
	}

	HostRegisterPatch = api.PatchComputeHostsHostIDRegisterJSONRequestBody{
		Name: &Host3Name,
	}

	HostRegisterPatchAutoOnboard = api.PatchComputeHostsHostIDRegisterJSONRequestBody{
		AutoOnboard: &AutoOnboardTrue,
	}

	SingleSchedule1Request = api.SingleSchedule{
		Name:           &SschedName1,
		StartSeconds:   SschedStart1,
		EndSeconds:     &SschedEnd1,
		ScheduleStatus: api.SCHEDULESTATUSMAINTENANCE,
	}
	SingleSchedule2Request = api.SingleSchedule{
		Name:           &SschedName2,
		StartSeconds:   SschedStart2,
		ScheduleStatus: api.SCHEDULESTATUSOSUPDATE,
	}
	SingleSchedule3Request = api.SingleSchedule{
		Name:           &SschedName3,
		StartSeconds:   SschedStart3,
		ScheduleStatus: api.SCHEDULESTATUSOSUPDATE,
	}
	SingleScheduleError = api.SingleSchedule{
		Name:           &SschedName3,
		StartSeconds:   SschedStart3,
		ScheduleStatus: api.SCHEDULESTATUSOSUPDATE,
	}
	SingleScheduleErrorSeconds = api.SingleSchedule{
		Name:           &SschedName3,
		StartSeconds:   SschedStart3,
		EndSeconds:     &SschedEndError,
		ScheduleStatus: api.SCHEDULESTATUSOSUPDATE,
	}

	RepeatedSchedule1Request = api.RepeatedSchedule{
		Name:            &SschedName1,
		DurationSeconds: 1,
		CronDayMonth:    cronDayMonth,
		// don't care the following
		CronMinutes:    CronAny,
		CronHours:      CronAny,
		CronMonth:      CronAny,
		CronDayWeek:    CronAny,
		ScheduleStatus: api.SCHEDULESTATUSMAINTENANCE,
	}
	RepeatedSchedule2Request = api.RepeatedSchedule{
		Name:            &SschedName2,
		DurationSeconds: 5,
		// don't care the following
		CronMinutes:    CronAny,
		CronHours:      CronAny,
		CronDayMonth:   CronAny,
		CronMonth:      CronAny,
		CronDayWeek:    CronAny,
		ScheduleStatus: api.SCHEDULESTATUSOSUPDATE,
	}
	RepeatedSchedule3Request = api.RepeatedSchedule{
		Name:            &SschedName3,
		DurationSeconds: 86400,
		// don't care the following
		CronMinutes:    CronAny,
		CronHours:      CronAny,
		CronDayMonth:   CronAny,
		CronMonth:      CronAny,
		CronDayWeek:    CronAny,
		ScheduleStatus: api.SCHEDULESTATUSOSUPDATE,
	}
	RepeatedScheduleError = api.RepeatedSchedule{
		Name:            &SschedName3,
		DurationSeconds: 86400,
		// don't care the following
		CronMinutes:    CronAny,
		CronHours:      CronAny,
		CronDayMonth:   CronAny,
		CronMonth:      CronAny,
		CronDayWeek:    CronAny,
		ScheduleStatus: api.SCHEDULESTATUSOSUPDATE,
	}
	RepeatedMissingRequest = api.RepeatedSchedule{
		Name:            &SschedName3,
		DurationSeconds: 86400,
		// don't care the following
		CronMinutes:    CronAny,
		CronHours:      CronAny,
		CronDayMonth:   CronAny,
		CronMonth:      CronAny,
		ScheduleStatus: api.SCHEDULESTATUSOSUPDATE,
	}
	RepeatedScheduleCronReqErr = api.RepeatedSchedule{
		Name:            &SschedName3,
		DurationSeconds: 86400,
		CronMinutes:     "/5",
		CronHours:       "*",
		CronDayMonth:    "*",
		CronMonth:       "*",
		CronDayWeek:     "*",
		ScheduleStatus:  api.SCHEDULESTATUSMAINTENANCE,
	}

	SingleScheduleAlwaysRequest = api.SingleSchedule{
		Name:           &SschedName1,
		StartSeconds:   SschedStart2,
		ScheduleStatus: api.SCHEDULESTATUSMAINTENANCE,
	}
	SingleScheduleNever = api.SingleSchedule{
		Name:           &SschedName2,
		StartSeconds:   SschedStart2,
		EndSeconds:     &SschedEnd1,
		ScheduleStatus: api.SCHEDULESTATUSMAINTENANCE,
	}
	RepeatedScheduleAlwaysRequest = api.RepeatedSchedule{
		Name:            &SschedName1,
		DurationSeconds: 120,
		CronMinutes:     CronAny,
		CronHours:       CronAny,
		CronDayMonth:    CronAny,
		CronMonth:       CronAny,
		CronDayWeek:     CronAny,
		ScheduleStatus:  api.SCHEDULESTATUSMAINTENANCE,
	}

	OSResource1Request = api.OperatingSystemResource{
		Name:            &OSName1,
		KernelCommand:   &OSKernel1,
		Architecture:    &OSArch1,
		UpdateSources:   []string{"sourcesList"},
		RepoUrl:         &OSRepo1,
		Sha256:          inv_testing.GenerateRandomSha256(),
		SecurityFeature: &OSSecurityFeature1,
		OsType:          &OSTypeImmutable,
		OsProvider:      &OSProviderInfra,
	}
	OSResource2Request = api.OperatingSystemResource{
		Name:            &OSName2,
		Architecture:    &OSArch2,
		UpdateSources:   []string{"sourcesList"},
		RepoUrl:         &OSRepo2,
		Sha256:          inv_testing.GenerateRandomSha256(),
		ProfileName:     &OSProfileName1,
		SecurityFeature: &OSSecurityFeature2,
		OsType:          &OSTypeImmutable,
		OsProvider:      &OSProviderInfra,
	}

	OSResourceRequestInvalidSha256 = api.OperatingSystemResource{
		Name:          &OSName3,
		Architecture:  &OSArch3,
		UpdateSources: []string{"sourcesList"},
		RepoUrl:       &OSRepo3,
		Sha256:        strings.ToUpper(inv_testing.GenerateRandomSha256()),
	}

	OSResourceRequestNoUpdateSources = api.OperatingSystemResource{
		Name:         &OSName3,
		Architecture: &OSArch3,
		RepoUrl:      &OSRepo3,
		Sha256:       inv_testing.GenerateRandomSha256(),
	}

	OSResourceRequestNoRepoURL = api.OperatingSystemResource{
		Name:          &OSName3,
		Architecture:  &OSArch3,
		UpdateSources: []string{"sourcesList"},
		Sha256:        inv_testing.GenerateRandomSha256(),
	}

	OSResourceRequestNoSha = api.OperatingSystemResource{
		Name:          &OSName3,
		Architecture:  &OSArch3,
		RepoUrl:       &OSRepo3,
		UpdateSources: []string{"sourcesList"},
	}
	OSResource1ReqwithInstallPackages = api.OperatingSystemResource{
		Name:              &OSName1,
		KernelCommand:     &OSKernel1,
		Architecture:      &OSArch1,
		UpdateSources:     []string{"sourcesList"},
		RepoUrl:           &OSRepo1,
		Sha256:            inv_testing.GenerateRandomSha256(),
		InstalledPackages: &OSInstalledPackages,
		OsType:            &OSTypeImmutable,
		OsProvider:        &OSProviderInfra,
	}

	clusterUuid1            = uuid.NewString()
	WorkloadCluster1Request = api.Workload{
		Name:       &WorkloadName1,
		Kind:       api.WORKLOADKINDCLUSTER,
		Status:     &WorkloadStatus1,
		ExternalId: &clusterUuid1,
	}
	WorkloadCluster2Request = api.Workload{
		Name:   &WorkloadName2,
		Kind:   api.WORKLOADKINDCLUSTER,
		Status: &WorkloadStatus2,
	}
	WorkloadCluster3Request = api.Workload{
		Kind:   api.WORKLOADKINDCLUSTER,
		Status: &WorkloadStatus2,
	}
	WorkloadNoKind = api.Workload{
		Name:   &WorkloadName2,
		Status: &WorkloadStatus2,
	}

	Instance1Request = api.Instance{
		HostID: &instHostID,
		OsID:   &instOSID,
		Kind:   &instKind,
		Name:   &Inst1Name,
	}

	Instance2Request = api.Instance{
		HostID:          &instHostID,
		OsID:            &instOSID,
		Kind:            &instKind,
		Name:            &Inst2Name,
		SecurityFeature: &instSecurityFeature,
	}

	InstanceRequestPatch = api.Instance{
		Kind:            &instKind,
		Name:            &Inst2Name,
		SecurityFeature: &instSecurityFeature,
	}

	InstanceRequestNoOSID = api.Instance{
		HostID: &instHostID,
		Kind:   &instKind,
		Name:   &Inst2Name,
	}

	InstanceRequestNoHostID = api.Instance{
		OsID: &instOSID,
		Kind: &instKind,
		Name: &Inst2Name,
	}

	TelemetryLogsGroup1Request = api.TelemetryLogsGroup{
		Name:          "HW Usage",
		CollectorKind: api.TELEMETRYCOLLECTORKINDHOST,
		Groups: []string{
			"syslog",
		},
	}
	TelemetryMetricsGroup1Request = api.TelemetryMetricsGroup{
		Name:          "Network Usage",
		CollectorKind: api.TELEMETRYCOLLECTORKINDHOST,
		Groups: []string{
			"net", "netstat", "ethtool",
		},
	}

	Provider1Request = api.Provider{
		ProviderKind:   providerKind1,
		ProviderVendor: &providerVendor1,
		Name:           ProviderName1,
		ApiEndpoint:    providerApiEndpoint1,
		ApiCredentials: &providerApiCredentials1,
		Config:         &providerConfig1,
	}

	Provider2Request = api.Provider{
		ProviderKind:   providerKind2,
		ProviderVendor: &providerVendor2,
		Name:           ProviderName2,
		ApiEndpoint:    providerApiEndpoint2,
		ApiCredentials: &providerApiCredentials2,
	}

	Provider3Request = api.Provider{
		ProviderKind: providerKind3,
		Name:         ProviderName3,
		ApiEndpoint:  providerApiEndpoint3,
	}

	ProviderNoKind = api.Provider{
		ProviderVendor: &providerVendor1,
		Name:           ProviderName1,
		ApiEndpoint:    providerApiEndpoint1,
		ApiCredentials: &providerApiCredentials1,
	}

	ProviderNoName = api.Provider{
		ProviderKind:   providerKind1,
		ProviderVendor: &providerVendor1,
		ApiEndpoint:    providerApiEndpoint1,
		ApiCredentials: &providerApiCredentials1,
	}

	ProviderNoApiEndpoint = api.Provider{
		ProviderKind:   providerKind1,
		ProviderVendor: &providerVendor1,
		Name:           ProviderName1,
		ApiCredentials: &providerApiCredentials1,
	}

	ProviderBadCredentials = api.Provider{
		ProviderKind:   providerKind1,
		ProviderVendor: &providerVendor1,
		Name:           ProviderName1,
		ApiCredentials: &providerBadApiCredentials,
	}
)
