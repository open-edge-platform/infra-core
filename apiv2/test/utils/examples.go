// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
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

	now                   = uint32(time.Now().Unix())
	FutureEpoch           = time.Unix(int64(now), 0).Add(1801 * time.Second)
	SschedStart1   uint32 = now + 1800
	SschedStart2   uint32 = now + 1800
	SschedStart3   uint32 = now + 1800
	SschedEnd1     uint32 = now + 3600
	SschedEndError uint32 = now - 1800

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
	OSSecurityFeature1  = api.OperatingSystemResourceSecurityFeatureSECURITYFEATURENONE
	OSSecurityFeature2  = api.OperatingSystemResourceSecurityFeatureSECURITYFEATURESECUREBOOTANDFULLDISKENCRYPTION

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
	instSecurityFeature = api.InstanceResourceSecurityFeatureSECURITYFEATURESECUREBOOTANDFULLDISKENCRYPTION

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

	MetadataOU1 = []api.MetadataItem{
		{
			Key:   "examplekey",
			Value: "ou1",
		}, {
			Key:   "examplekey2",
			Value: "ou1",
		},
	}
	MetadataOU2 = []api.MetadataItem{
		{
			Key:   "examplekey",
			Value: "ou2",
		}, {
			Key:   "examplekey2",
			Value: "ou2",
		},
	}
	MetadataOU3 = []api.MetadataItem{
		{
			Key:   "examplekey2",
			Value: "ou3",
		},
		{
			Key:   "examplekey3",
			Value: "ou3",
		},
	}

	MetadataOU3Rendered = []api.MetadataItem{
		{
			Key:   "examplekey",
			Value: "ou2",
		},
	}

	MetadataR1 = []api.MetadataItem{
		{
			Key:   "examplekey",
			Value: "r1",
		}, {
			Key:   "examplekey2",
			Value: "r1",
		},
	}
	MetadataR2 = []api.MetadataItem{
		{
			Key:   "examplekey",
			Value: "r2",
		}, {
			Key:   "examplekey2",
			Value: "r2",
		},
	}
	MetadataR3 = []api.MetadataItem{
		{
			Key:   "examplekey",
			Value: "r3",
		},
	}
	MetadataR3Inherited = []api.MetadataItem{
		{
			Key:   "examplekey2",
			Value: "r2",
		},
	}

	MetadataSite2 = []api.MetadataItem{
		{
			Key:   "examplekey2",
			Value: "site1",
		},
	}

	MetadataHost1 = []api.MetadataItem{
		{
			Key:   "examplekey1",
			Value: "host1",
		},
	}

	MetadataHost2 = []api.MetadataItem{
		{
			Key:   "examplekey1",
			Value: "host2",
		},
		{
			Key:   "examplekey3",
			Value: "host2",
		},
	}

	MetadataRightPattern = []api.MetadataItem{
		{
			Key:   "asd/ad.123",
			Value: "site1-.ad",
		},
		{
			Key:   "city",
			Value: "test-region",
		},
	}

	MetadataWrongPattern = []api.MetadataItem{
		{
			Key:   "/examplekey2",
			Value: "-site1",
		},
	}

	Host1UUID1       = "BFD3B398-9A4B-480D-AB53-4050ED108F5C"
	Host4UUID1       = "BFD3B398-9A4C-481D-AB53-4050ED108F5D"
	Host1UUIDPatched = "BFD3B398-9A4B-480D-AB53-4050ED108F5E"
	HostUUIDUnexists = "BFD3B398-9A4B-480D-AB53-4050ED108F5F"
	HostUUIDError    = "BFD3B398-9A4B-480D-AB53-4050ED108F5FKK"
	Host2UUID        = uuid.New().String()
	Host3UUID        = uuid.New().String()
	Host5UUID        = uuid.New().String()

	HostSerialNumber1 = "SN001"
	HostSerialNumber2 = "SN002"
	HostSerialNumber3 = "SN003"

	DnsServers = []string{"10.10.10.10"}

	Region1Request = api.RegionResource{
		Name:     &Region1Name,
		Metadata: &MetadataR1,
	}

	Region1RequestWrong = api.RegionResource{}

	Region1RequestMetadataOK = api.RegionResource{
		Name:     &Region1Name,
		Metadata: &MetadataRightPattern,
	}

	Region1RequestMetadataNOK = api.RegionResource{
		Name:     &Region1Name,
		Metadata: &MetadataWrongPattern,
	}

	Region2Request = api.RegionResource{
		Name:     &Region2Name,
		Metadata: &MetadataR2,
	}

	Region3Request = api.RegionResource{
		Name:     &Region3Name,
		Metadata: &MetadataR3,
	}

	SiteListRequest = api.SiteResource{
		Name: &Site1Name,
	}

	SiteListRequest1 = api.SiteResource{
		Name: &Site1Name,
	}

	SiteListRequest2 = api.SiteResource{
		Name: &Site2Name,
	}

	SiteListRequest3 = api.SiteResource{
		Name: &Site3Name,
	}

	Site1Request = api.SiteResource{
		Name: &Site1Name,
	}

	Site1RequestUpdate = api.SiteResource{
		Name: &Site1Name,
	}

	Site1RequestUpdatePatch = api.SiteResource{
		Name: &Site2Name,
	}
	Site2Request = api.SiteResource{
		Name:     &Site2Name,
		Metadata: &MetadataSite2,
	}
	Site3Request = api.SiteResource{
		Name:     &Site3Name,
		Metadata: &MetadataSite2,
	}

	metadata = []api.MetadataItem{
		{
			Key:   "examplekey",
			Value: "examplevalue",
		}, {
			Key:   "examplekey2",
			Value: "examplevalue2",
		},
	}

	metadata1 = []api.MetadataItem{
		{
			Key:   "filtermetakey1",
			Value: "filtermetavalue1",
		}, {
			Key:   "filtermetakey2",
			Value: "filtermetavalue2",
		},
	}

	metadata2 = []api.MetadataItem{
		{
			Key:   "filtermetakey1",
			Value: "filtermetavalue1",
		}, {
			Key:   "filtermetakey2",
			Value: "filtermetavalue2_mod",
		},
	}
	AutoOnboardTrue  bool = true
	AutoOnboardFalse bool = false

	HostRegisterAutoOnboard = api.HostRegister{
		Name:         &Host2Name,
		Uuid:         &Host3UUID,
		SerialNumber: &HostSerialNumber2,
		AutoOnboard:  &AutoOnboardTrue,
	}
	HostRegister = api.HostRegister{
		Name: &Host1Name,
		Uuid: &Host1UUID1,
	}

	HostRegisterPatch = api.HostRegister{
		Name: &Host2Name,
		Uuid: &Host1UUID1,
	}

	Host1Request = api.HostResource{
		Name:     Host1Name,
		Metadata: &metadata,
		Uuid:     &Host1UUID1,
	}

	Host1RequestUpdate = api.HostResource{
		Metadata: &[]api.MetadataItem{
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
		Uuid:   &Host1UUID1, // Must be the same UUID used in the creation
	}

	Host1RequestPatch = api.HostResource{
		Metadata: &[]api.MetadataItem{
			{
				Key:   "examplekey",
				Value: "examplevalue",
			}, {
				Key:   "examplekey2",
				Value: "examplevalue2",
			},
		},
		Name: Host3Name,
	}

	Host2Request = api.HostResource{
		Name:     Host2Name,
		Metadata: &metadata,
		Uuid:     &Host2UUID,
	}

	HostReqFilterMeta1 = api.HostResource{
		Metadata: &metadata1,
		Uuid:     &Host1UUID1,
	}
	HostReqFilterMeta2 = api.HostResource{
		Metadata: &metadata2,
		Uuid:     &Host2UUID,
	}

	Host3Request = api.HostResource{
		Metadata: &[]api.MetadataItem{
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

	Host4Request = api.HostResource{
		Name: Host4Name,
		Uuid: &Host4UUID1,
	}

	Host4RequestPut = api.HostResource{
		Metadata: &[]api.MetadataItem{
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
		Uuid:   &Host4UUID1,
	}

	Host4RequestPutMissingField = api.HostResource{
		Metadata: &[]api.MetadataItem{
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
		Uuid:   &Host4UUID1,
	}

	Host4RequestPatch = api.HostResource{
		Metadata: &[]api.MetadataItem{
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

	HostNonPrintable = api.HostResource{
		Metadata: &[]api.MetadataItem{
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

	SingleSchedule1Request = api.SingleScheduleResource{
		Name:           &SschedName1,
		StartSeconds:   SschedStart1,
		EndSeconds:     &SschedEnd1,
		ScheduleStatus: api.SingleScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}
	SingleSchedule2Request = api.SingleScheduleResource{
		Name:           &SschedName2,
		StartSeconds:   SschedStart2,
		ScheduleStatus: api.SingleScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}
	SingleSchedule3Request = api.SingleScheduleResource{
		Name:           &SschedName3,
		StartSeconds:   SschedStart3,
		ScheduleStatus: api.SingleScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}
	SingleScheduleError = api.SingleScheduleResource{
		Name:           &SschedName3,
		StartSeconds:   SschedStart3,
		ScheduleStatus: api.SingleScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}
	SingleScheduleErrorSeconds = api.SingleScheduleResource{
		Name:           &SschedName3,
		StartSeconds:   SschedStart3,
		EndSeconds:     &SschedEndError,
		ScheduleStatus: api.SingleScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}

	RepeatedSchedule1Request = api.RepeatedScheduleResource{
		Name:            &SschedName1,
		DurationSeconds: 1,
		CronDayMonth:    cronDayMonth,
		// don't care the following
		CronMinutes:    CronAny,
		CronHours:      CronAny,
		CronMonth:      CronAny,
		CronDayWeek:    CronAny,
		ScheduleStatus: api.RepeatedScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}
	RepeatedSchedule2Request = api.RepeatedScheduleResource{
		Name:            &SschedName2,
		DurationSeconds: 5,
		// don't care the following
		CronMinutes:    CronAny,
		CronHours:      CronAny,
		CronDayMonth:   CronAny,
		CronMonth:      CronAny,
		CronDayWeek:    CronAny,
		ScheduleStatus: api.RepeatedScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}
	RepeatedSchedule3Request = api.RepeatedScheduleResource{
		Name:            &SschedName3,
		DurationSeconds: 86400,
		// don't care the following
		CronMinutes:    CronAny,
		CronHours:      CronAny,
		CronDayMonth:   CronAny,
		CronMonth:      CronAny,
		CronDayWeek:    CronAny,
		ScheduleStatus: api.RepeatedScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}
	RepeatedScheduleError = api.RepeatedScheduleResource{
		Name:            &SschedName3,
		DurationSeconds: 86400,
		// don't care the following
		CronMinutes:    CronAny,
		CronHours:      CronAny,
		CronDayMonth:   CronAny,
		CronMonth:      CronAny,
		CronDayWeek:    CronAny,
		ScheduleStatus: api.RepeatedScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}
	RepeatedMissingRequest = api.RepeatedScheduleResource{
		Name:            &SschedName3,
		DurationSeconds: 86400,
		// don't care the following
		CronMinutes:    CronAny,
		CronHours:      CronAny,
		CronDayMonth:   CronAny,
		CronMonth:      CronAny,
		ScheduleStatus: api.RepeatedScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}
	RepeatedScheduleCronReqErr = api.RepeatedScheduleResource{
		Name:            &SschedName3,
		DurationSeconds: 86400,
		CronMinutes:     "/5",
		CronHours:       "*",
		CronDayMonth:    "*",
		CronMonth:       "*",
		CronDayWeek:     "*",
		ScheduleStatus:  api.RepeatedScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}

	SingleScheduleAlwaysRequest = api.SingleScheduleResource{
		Name:           &SschedName1,
		StartSeconds:   SschedStart2,
		ScheduleStatus: api.SingleScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}
	SingleScheduleNever = api.SingleScheduleResource{
		Name:           &SschedName2,
		StartSeconds:   SschedStart2,
		EndSeconds:     &SschedEnd1,
		ScheduleStatus: api.SingleScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}
	RepeatedScheduleAlwaysRequest = api.RepeatedScheduleResource{
		Name:            &SschedName1,
		DurationSeconds: 120,
		CronMinutes:     CronAny,
		CronHours:       CronAny,
		CronDayMonth:    CronAny,
		CronMonth:       CronAny,
		CronDayWeek:     CronAny,
		ScheduleStatus:  api.RepeatedScheduleResourceScheduleStatusSCHEDULESTATUSMAINTENANCE,
	}

	OsTypeMutable  = api.OSTYPEMUTABLE
	OsTypeImutable = api.OSTYPEIMMUTABLE
	OSProvider     = api.OSPROVIDERKINDINFRA

	OSResource1Request = api.OperatingSystemResource{
		Name:            &OSName1,
		KernelCommand:   &OSKernel1,
		Architecture:    &OSArch1,
		UpdateSources:   []string{"sourcesList"},
		RepoUrl:         &OSRepo1,
		Sha256:          inv_testing.GenerateRandomSha256(),
		SecurityFeature: &OSSecurityFeature1,
		OsType:          &OsTypeMutable,
		OsProvider:      &OSProvider,
	}
	OSResource2Request = api.OperatingSystemResource{
		Name:            &OSName2,
		Architecture:    &OSArch2,
		UpdateSources:   []string{"sourcesList"},
		RepoUrl:         &OSRepo2,
		Sha256:          inv_testing.GenerateRandomSha256(),
		ProfileName:     &OSProfileName1,
		SecurityFeature: &OSSecurityFeature2,
		OsType:          &OsTypeImutable,
		OsProvider:      &OSProvider,
	}

	OSResource3Request = api.OperatingSystemResource{
		Name:          &OSName3,
		Architecture:  &OSArch3,
		UpdateSources: []string{"sourcesList"},
		RepoUrl:       &OSRepo3,
		Sha256:        inv_testing.GenerateRandomSha256(),
		OsType:        &OsTypeMutable,
		OsProvider:    &OSProvider,
	}

	OSResourceRequestInvalidSha256 = api.OperatingSystemResource{
		Name:          &OSName3,
		Architecture:  &OSArch3,
		UpdateSources: []string{"sourcesList"},
		RepoUrl:       &OSRepo3,
		Sha256:        strings.ToUpper(inv_testing.GenerateRandomSha256()),
		OsType:        &OsTypeMutable,
		OsProvider:    &OSProvider,
	}

	OSResourceRequestNoUpdateSources = api.OperatingSystemResource{
		Name:         &OSName3,
		Architecture: &OSArch3,
		RepoUrl:      &OSRepo3,
		Sha256:       inv_testing.GenerateRandomSha256(),
		OsType:       &OsTypeMutable,
		OsProvider:   &OSProvider,
	}

	OSResourceRequestNoRepoURL = api.OperatingSystemResource{
		Name:          &OSName3,
		Architecture:  &OSArch3,
		UpdateSources: []string{"sourcesList"},
		Sha256:        inv_testing.GenerateRandomSha256(),
		OsType:        &OsTypeMutable,
		OsProvider:    &OSProvider,
	}

	OSResourceRequestNoSha = api.OperatingSystemResource{
		Name:          &OSName3,
		Architecture:  &OSArch3,
		RepoUrl:       &OSRepo3,
		UpdateSources: []string{"sourcesList"},
		OsType:        &OsTypeMutable,
		OsProvider:    &OSProvider,
	}
	OSResource1ReqwithInstallPackages = api.OperatingSystemResource{
		Name:              &OSName1,
		KernelCommand:     &OSKernel1,
		Architecture:      &OSArch1,
		UpdateSources:     []string{"sourcesList"},
		RepoUrl:           &OSRepo1,
		Sha256:            inv_testing.GenerateRandomSha256(),
		InstalledPackages: &OSInstalledPackages,
		OsType:            &OsTypeMutable,
		OsProvider:        &OSProvider,
	}

	clusterUuid1            = uuid.NewString()
	WorkloadCluster1Request = api.WorkloadResource{
		Name:       &WorkloadName1,
		Kind:       api.WORKLOADKINDCLUSTER,
		Status:     &WorkloadStatus1,
		ExternalId: &clusterUuid1,
	}
	WorkloadCluster2Request = api.WorkloadResource{
		Name:   &WorkloadName2,
		Kind:   api.WORKLOADKINDCLUSTER,
		Status: &WorkloadStatus2,
	}
	WorkloadCluster3Request = api.WorkloadResource{
		Kind:   api.WORKLOADKINDCLUSTER,
		Status: &WorkloadStatus2,
	}
	WorkloadNoKind = api.WorkloadResource{
		Name:   &WorkloadName2,
		Status: &WorkloadStatus2,
	}

	Instance1Request = api.InstanceResource{
		HostID: &instHostID,
		OsID:   &instOSID,
		Kind:   &instKind,
		Name:   &Inst1Name,
	}

	Instance2Request = api.InstanceResource{
		HostID:          &instHostID,
		OsID:            &instOSID,
		Kind:            &instKind,
		Name:            &Inst2Name,
		SecurityFeature: &instSecurityFeature,
	}

	InstanceRequestPatch = api.InstanceResource{
		Kind:            &instKind,
		Name:            &Inst2Name,
		SecurityFeature: &instSecurityFeature,
	}

	InstanceRequestNoOSID = api.InstanceResource{
		HostID: &instHostID,
		Kind:   &instKind,
		Name:   &Inst2Name,
	}

	InstanceRequestNoHostID = api.InstanceResource{
		OsID: &instOSID,
		Kind: &instKind,
		Name: &Inst2Name,
	}

	TelemetryLogsGroup1Request = api.TelemetryLogsGroupResource{
		Name:          "HW Usage",
		CollectorKind: api.TelemetryLogsGroupResourceCollectorKindCOLLECTORKINDCLUSTER,
		Groups: []string{
			"syslog",
		},
	}
	TelemetryMetricsGroup1Request = api.TelemetryMetricsGroupResource{
		Name:          "Network Usage",
		CollectorKind: api.TelemetryMetricsGroupResourceCollectorKindCOLLECTORKINDHOST,
		Groups: []string{
			"net", "netstat", "ethtool",
		},
	}

	Provider1Request = api.ProviderResource{
		ProviderKind:   providerKind1,
		ProviderVendor: &providerVendor1,
		Name:           ProviderName1,
		ApiEndpoint:    providerApiEndpoint1,
		ApiCredentials: &providerApiCredentials1,
		Config:         &providerConfig1,
	}

	Provider2Request = api.ProviderResource{
		ProviderKind:   providerKind2,
		ProviderVendor: &providerVendor2,
		Name:           ProviderName2,
		ApiEndpoint:    providerApiEndpoint2,
		ApiCredentials: &providerApiCredentials2,
	}

	Provider3Request = api.ProviderResource{
		ProviderKind: providerKind3,
		Name:         ProviderName3,
		ApiEndpoint:  providerApiEndpoint3,
	}

	ProviderNoKind = api.ProviderResource{
		ProviderVendor: &providerVendor1,
		Name:           ProviderName1,
		ApiEndpoint:    providerApiEndpoint1,
		ApiCredentials: &providerApiCredentials1,
	}

	ProviderNoName = api.ProviderResource{
		ProviderKind:   providerKind1,
		ProviderVendor: &providerVendor1,
		ApiEndpoint:    providerApiEndpoint1,
		ApiCredentials: &providerApiCredentials1,
	}

	ProviderNoApiEndpoint = api.ProviderResource{
		ProviderKind:   providerKind1,
		ProviderVendor: &providerVendor1,
		Name:           ProviderName1,
		ApiCredentials: &providerApiCredentials1,
	}

	ProviderBadCredentials = api.ProviderResource{
		ProviderKind:   providerKind1,
		ProviderVendor: &providerVendor1,
		Name:           ProviderName1,
		ApiCredentials: &providerBadApiCredentials,
	}

	// Example LocalAccount resources for testing

	// LocalAccountUnexistID represents a non-existent LocalAccount ID for testing
	LocalAccountUnexistID = "nonexistent-id"

	// Example LocalAccount resources for testing
	LocalAccount1Request = api.LocalAccountResource{
		LocalAccountID: nil, // This will be populated by the API upon creation
		ResourceId:     nil, // This will be populated by the API upon creation
		SshKey:         "ssh-ed25519 AAAAB3NzaC1yc2EAAAABIwAAAQEArandomkey1",
		Timestamps:     nil, // This will be populated by the API upon creation
		Username:       "user1",
	}

	// LocalAccount2Request represents another valid LocalAccount resource for testing
	LocalAccount2Request = api.LocalAccountResource{
		LocalAccountID: nil, // This will be populated by the API upon creation
		ResourceId:     nil, // This will be populated by the API upon creation
		SshKey:         "ssh-ed25519 AAAAB3NzaC1yc2EAAAABIwAAAQEArandomkey2",
		Timestamps:     nil, // This will be populated by the API upon creation
		Username:       "user2",
	}

	// LocalAccountNoName represents an invalid LocalAccount resource with no username
	LocalAccountNoName = api.LocalAccountResource{
		LocalAccountID: nil,
		ResourceId:     nil,
		SshKey:         "ssh-ed25519 AAAAB3NzaC1yc2EAAAABIwAAAQEArandomkey3",
		Timestamps:     nil,
		Username:       "",
	}
)
