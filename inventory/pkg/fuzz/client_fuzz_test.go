// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package fuzz_test

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	invv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	localaccountv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/localaccount/v1"
	locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	networkv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	ouv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/ou/v1"
	providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	remotev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/remoteaccess/v1"
	schedulev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	telemetryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
)

var log = logging.GetLogger("go-fuzz")

var (
	err01 = "unexpected end of JSON input"
	err02 = "cannot unmarshal"
	err03 = "error while marshaling"
	err04 = "invalid character"
	err05 = "invalid byte sequence"
	err06 = "code = InvalidArgument desc"
	err07 = "code = PermissionDenied desc"
	err08 = "unknown resource type"
	err09 = "expected colon after"
	err10 = "expected comma after"
)

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// Currently unused
	flag.String(
		"policyBundle",
		wd+"/../../out/policy_bundle.tar.gz",
		"Path of policy rego file",
	)
	flag.Parse()
	projectRoot := filepath.Dir(filepath.Dir(wd))

	policyPath := projectRoot + "/out"
	certPath := projectRoot + "/cert/certificates"
	migrationsDir := projectRoot + "/out"

	inv_testing.StartTestingEnvironment(policyPath, certPath, migrationsDir)
	rand.New(rand.NewSource(time.Now().UnixNano())) // seed random number generator for fuzzing
	run := m.Run()                                  // run all tests
	inv_testing.StopTestingEnvironment()

	os.Exit(run)
}

func getTestClient(f *testing.F) client.InventoryClient {
	f.Helper()
	invClient := inv_testing.TestClients[inv_testing.APIClient]
	return invClient
}

func checkError(err error) bool {
	return !strings.Contains(err.Error(), err01) &&
		!strings.Contains(err.Error(), err02) &&
		!strings.Contains(err.Error(), err03) &&
		!strings.Contains(err.Error(), err04) &&
		!strings.Contains(err.Error(), err05) &&
		!strings.Contains(err.Error(), err06) &&
		!strings.Contains(err.Error(), err07) &&
		!strings.Contains(err.Error(), err08) &&
		!strings.Contains(err.Error(), err09) &&
		!strings.Contains(err.Error(), err10)
}

func FuzzCreateRegion(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("name", `[{"key":"key1-test","value":"region_key1_lvl4-test"}]`, "regionKind", "tenant-12345678")
	f.Add("kkkkkk", `[{"key":"key2-test","value":"region_key2_lvl4-test"}]`, "region1", "tenant-87654321")
	f.Add("456", `[{"key":"key3-test","value":"region_key3_lvl4-test"}]`, "abcdef", "tenant-11223344")
	f.Add(" ", `[{"key":"key4-test","value":"region_key4_lvl4-test"}]`, "123456", "tenant-55667788")
	f.Fuzz(func(t *testing.T,
		name string, metadata string, regionKind string, tenantId string,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_Region{
				Region: &locationv1.RegionResource{
					Name:       name,
					Metadata:   metadata,
					RegionKind: regionKind,
					TenantId:   tenantId,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")

			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateSite(f *testing.F) {
	invClient := getTestClient(f)
	f.Add(
		"name",
		`[{"key":"key1-test","value":"region_key1_lvl4-test"}]`,
		"address",
		int32(100000),
		int32(100000),
		"metricsEndpoint",
		"httpProxy",
		"httpsProxy",
		"ftpProxy",
		"noProxy",
		"tenant-12345678",
	)
	f.Add(
		"site1",
		`[{"key":"key2-test","value":"region_key2_lvl4-test"}]`,
		"address1",
		int32(200000),
		int32(200000),
		"http://123.456.789:8000",
		"123456",
		"abc",
		"bndasdasd",
		"789asd",
		"tenant-87654321",
	)
	f.Add(
		"site2",
		`[{"key":"key3-test","value":"region_key3_lvl4-test"}]`,
		"address2",
		int32(300000),
		int32(300000),
		" ",
		"adasdasdassd",
		"2",
		" ",
		"///",
		"tenant-11223344",
	)
	f.Add(
		"site3",
		`[{"key":"key4-test","value":"region_key4_lvl4-test"}]`,
		"address3",
		int32(400000),
		int32(400000),
		"metricsEndpoint3",
		" ",
		"//:",
		"abcsdsad",
		"123456",
		"tenant-55667788",
	)
	f.Fuzz(
		func(t *testing.T,
			name string,
			metadata string,
			address string,
			siteLat int32,
			siteLng int32,
			metricsEndpoint string,
			httpProxy string,
			httpsProxy string,
			ftpProxy string,
			noProxy string,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_Site{
					Site: &locationv1.SiteResource{
						Name:            name,
						Address:         address,
						SiteLat:         siteLat,
						SiteLng:         siteLng,
						MetricsEndpoint: metricsEndpoint,
						HttpProxy:       httpProxy,
						HttpsProxy:      httpsProxy,
						FtpProxy:        ftpProxy,
						NoProxy:         noProxy,
						TenantId:        tenantId,
						Metadata:        metadata,
					},
				},
			}

			ctx := context.Background()

			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}

func FuzzCreateOu(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("name", "ouKind", `[{"key":"key1-test","value":"region_key1_lvl4-test"}]`, "tenant-12345678")
	f.Add("ou", "abca", `[{"key":"key2-test","value":"region_key2_lvl4-test"}]`, "tenant-87654321")
	f.Add("2", "123456", `[{"key":"key3-test","value":"region_key3_lvl4-test"}]`, "tenant-11223344")
	f.Add("dsadasdas", " ", `[{"key":"key4-test","value":"region_key4_lvl4-test"}]`, "tenant-55667788")
	f.Fuzz(func(t *testing.T,
		name string, ouKind string, metadata string, tenantId string,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_Ou{
				Ou: &ouv1.OuResource{
					Name:     name,
					OuKind:   ouKind,
					Metadata: metadata,
					TenantId: tenantId,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")
			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateProvider(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("name", "http://123.456.789:8000", "config", "tenant-12345678")
	f.Add("dasdsadas", "http://987.654.321:8000", "456", "tenant-87654321")
	f.Add("123123", "http://111.222.333:8000", "AAA2", "tenant-11223344")
	f.Add(" ", "http://444.555.666:8000", "BCAS", "tenant-55667788")
	f.Fuzz(func(t *testing.T,
		name string,
		apiEndpoint string,
		config string,
		tenantId string,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_Provider{
				Provider: &providerv1.ProviderResource{
					Name:        name,
					ApiEndpoint: apiEndpoint,
					Config:      config,
					TenantId:    tenantId,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")
			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateHostResource(f *testing.F) {
	invClient := getTestClient(f)
	f.Add(
		"name",
		"kind",
		"note",
		"hardwareKind",
		"serialNumber",
		"0000-0000-0000-0000",
		uint64(100000),
		"cpuModel",
		uint32(100),
		uint32(100),
		"cpuCapabilities",
		"cpuArchitecture",
		uint32(100),
		"cpuTopology",
		"123.142.123.123",
		"bmcIp",
		"dasdasd",
		"123123",
		"pxeMac",
		"hostname",
		"productName",
		"biosVersion",
		"biosReleaseDate",
		"biosVendor",
		`[{"key":"key1-test","value":"region_key1_lvl4-test"}]`,
		"hostStatus",
		uint64(100000),
		"onboardingStatus",
		uint64(100000),
		"registrationStatus",
		uint64(100000),
		"tenant-12345678",
	)
	f.Add(
		"host1",
		"kind1",
		"note1",
		"hardwareKind1",
		"serialNumber1",
		"1111-1111-1111-1111",
		uint64(200000),
		"cpuModel1",
		uint32(200),
		uint32(200),
		"cpuCapabilities1",
		"cpuArchitecture1",
		uint32(200),
		"cpuTopology1",
		"mgmtIp1",
		" ",
		"bmcUsername1",
		"bmcPassword1",
		"pxeMac1",
		"hostname1",
		"productName1",
		"addsad",
		"biosReleaseDate1",
		"biosVendor1",
		`[{"key":"key2-test","value":"region_key2_lvl4-test"}]`,
		"hostStatus1",
		uint64(200000),
		"onboardingStatus1",
		uint64(200000),
		"registrationStatus1",
		uint64(200000),
		"tenant-87654321",
	)
	f.Add(
		"host2",
		"aaaaaa",
		"note2",
		"hardwareKind2",
		"123456",
		"2222-2222-2222-2222",
		uint64(300000),
		"cpuModel2",
		uint32(300),
		uint32(300),
		"cpuCapabilities2",
		"cpuArchitecture2",
		uint32(300),
		"cpuTopology2",
		"mgmtIp2",
		"123.",
		"bmcUsername2",
		"bmcPassword2",
		"01:02:03:04:05:06",
		"hostname2",
		"4",
		"biosVersion2",
		"05050505",
		"aaaaaaaaaa",
		`[{"key":"key3-test","value":"region_key3_lvl4-test"}]`,
		"hostStatus2",
		uint64(300000),
		"onboardingStatus2",
		uint64(300000),
		"registrationStatus2",
		uint64(300000),
		"tenant-11223344",
	)
	f.Add(
		"host3",
		"kind3",
		"note3",
		"hardwareKind3",
		"serialNumber3",
		"3333-3333-3333-3333",
		uint64(400000),
		"cpuModel3",
		uint32(400),
		uint32(400),
		"cpuCapabilities3",
		"cpuArchitecture3",
		uint32(400),
		"cpuTopology3",
		" ",
		"456.",
		"bmcUsername3",
		"bmcPassword3",
		"pxeMac3",
		"hostname3",
		"productName3",
		"AAAAAA",
		"biosReleaseDate3",
		"CCCCC",
		`[{"key":"key4-test","value":"region_key4_lvl4-test"}]`,
		"hostStatus3",
		uint64(400000),
		"onboardingStatus3",
		uint64(400000),
		"registrationStatus3",
		uint64(400000),
		"tenant-55667788",
	)
	f.Fuzz(
		func(t *testing.T,
			name string,
			kind string,
			note string,
			hardwareKind string,
			serialNumber string,
			uuid string,
			memoryBytes uint64,
			cpuModel string,
			cpuSockets uint32,
			cpuCores uint32,
			cpuCapabilities string,
			cpuArchitecture string,
			cpuThreads uint32,
			cpuTopology string,
			mgmtIp string,
			bmcIp string,
			bmcUsername string,
			bmcPassword string,
			pxeMac string,
			hostname string,
			productName string,
			biosVersion string,
			biosReleaseDate string,
			biosVendor string,
			metadata string,
			hostStatus string,
			hostStatusTimestamp uint64,
			onboardingStatus string,
			onboardingStatusTimestamp uint64,
			registrationStatus string,
			registrationStatusTimestamp uint64,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_Host{
					Host: &computev1.HostResource{
						Name:                        name,
						Kind:                        kind,
						Note:                        note,
						HardwareKind:                hardwareKind,
						SerialNumber:                serialNumber,
						Uuid:                        uuid,
						MemoryBytes:                 memoryBytes,
						CpuModel:                    cpuModel,
						CpuSockets:                  cpuSockets,
						CpuCores:                    cpuCores,
						CpuCapabilities:             cpuCapabilities,
						CpuArchitecture:             cpuArchitecture,
						CpuThreads:                  cpuThreads,
						CpuTopology:                 cpuTopology,
						MgmtIp:                      mgmtIp,
						BmcIp:                       bmcIp,
						BmcUsername:                 bmcUsername,
						BmcPassword:                 bmcPassword,
						PxeMac:                      pxeMac,
						Hostname:                    hostname,
						ProductName:                 productName,
						BiosVersion:                 biosVersion,
						BiosReleaseDate:             biosReleaseDate,
						BiosVendor:                  biosVendor,
						Metadata:                    metadata,
						HostStatus:                  hostStatus,
						HostStatusTimestamp:         hostStatusTimestamp,
						OnboardingStatus:            onboardingStatus,
						OnboardingStatusTimestamp:   onboardingStatusTimestamp,
						RegistrationStatus:          registrationStatus,
						RegistrationStatusTimestamp: registrationStatusTimestamp,
						TenantId:                    tenantId,
					},
				},
			}

			ctx := context.Background()

			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}

func FuzzCreateHoststorage(f *testing.F) {
	invClient := getTestClient(f)
	f.Add(
		"kind",
		"providerStatus",
		"wwid",
		"serial",
		"vendor",
		"model",
		uint64(100000),
		"deviceName",
		"tenant-12345678",
	)
	f.Add(
		"kind1",
		"providerStatus1",
		"wwww",
		"aaaa",
		"ADASD",
		"model1",
		uint64(200000),
		"deviceName1",
		"tenant-87654321",
	)
	f.Add(
		"kind2",
		"providerStatus2",
		"11111",
		"123123",
		"vendor2",
		"model2",
		uint64(300000),
		" ",
		"tenant-11223344",
	)
	f.Add(
		"kind3",
		"aaaaaaaa",
		"wwid3",
		"serial3",
		"31231",
		"1231 ",
		uint64(400000),
		"deviceName3",
		"tenant-55667788",
	)
	f.Fuzz(
		func(t *testing.T,
			kind string,
			providerStatus string,
			wwid string,
			serial string,
			vendor string,
			model string,
			capacityBytes uint64,
			deviceName string,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_Hoststorage{
					Hoststorage: &computev1.HoststorageResource{
						Kind:           kind,
						ProviderStatus: providerStatus,
						Wwid:           wwid,
						Serial:         serial,
						Vendor:         vendor,
						Model:          model,
						CapacityBytes:  capacityBytes,
						DeviceName:     deviceName,
						TenantId:       tenantId,
					},
				},
			}

			ctx := context.Background()

			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}

func FuzzCreateHostnic(f *testing.F) {
	invClient := getTestClient(f)
	f.Add(
		"kind",
		"providerStatus",
		"deviceName",
		"123123",
		"00:00:00:00:00:00",
		true,
		uint32(100),
		uint32(100),
		"peerName",
		"peerDescription",
		"11:11:11:11:11:11",
		"10.10.10.10",
		"123456",
		"supportedLinkMode",
		"advertisingLinkMode",
		uint64(100000),
		"currentDuplex",
		"features",
		uint32(100),
		true,
		"tenant-12345678",
	)
	f.Add(
		"kind1",
		"providerStatus1",
		"asdasdas",
		"1 ",
		"22:22:22:22:22:22",
		true,
		uint32(200),
		uint32(200),
		" ",
		"peerDescription1",
		"33:33:33:33:33:33",
		"20.20.20.20",
		"654321",
		"supportedLinkMode1",
		"aaa",
		uint64(200000),
		"currentDuplex1",
		"1111",
		uint32(200),
		true,
		"tenant-87654321",
	)
	f.Add(
		"kind2",
		"providerStatus2",
		"deviceName2",
		"pciIdentifier2",
		"44:44:44:44:44:44",
		true,
		uint32(300),
		uint32(300),
		"peerName2",
		" ",
		"55:55:55:55:55:55",
		"30.30.30.30",
		"789012",
		"supportedLinkMode2",
		"advertisingLinkMode2",
		uint64(300000),
		"currentDuplex2",
		"features2",
		uint32(300),
		true,
		"tenant-11223344",
	)
	f.Add(
		"kind3",
		"providerStatus3",
		"123123123",
		"pciIdentifier3",
		"66:66:66:66:66:66",
		true,
		uint32(400),
		uint32(400),
		"peerName3",
		"peerDescription3",
		"77:77:77:77:77:77",
		"40.40.40.40",
		"210987",
		" ",
		"nnnn",
		uint64(400000),
		"currentDuplex3",
		" ",
		uint32(400),
		true,
		"tenant-55667788",
	)
	f.Fuzz(
		func(t *testing.T,
			kind string,
			providerStatus string,
			deviceName string,
			pciIdentifier string,
			macAddr string,
			sriovEnabled bool,
			sriovVfsNum uint32,
			sriovVfsTotal uint32,
			peerName string,
			peerDescription string,
			peerMac string,
			peerMgmtIp string,
			peerPort string,
			supportedLinkMode string,
			advertisingLinkMode string,
			currentSpeedBps uint64,
			currentDuplex string,
			features string,
			mtu uint32,
			bmcInterface bool,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_Hostnic{
					Hostnic: &computev1.HostnicResource{
						DeviceName:          deviceName,
						Kind:                kind,
						ProviderStatus:      providerStatus,
						PciIdentifier:       pciIdentifier,
						MacAddr:             macAddr,
						SriovEnabled:        sriovEnabled,
						SriovVfsNum:         sriovVfsNum,
						SriovVfsTotal:       sriovVfsTotal,
						PeerName:            peerName,
						PeerDescription:     peerDescription,
						PeerMac:             peerMac,
						PeerMgmtIp:          peerMgmtIp,
						PeerPort:            peerPort,
						SupportedLinkMode:   supportedLinkMode,
						AdvertisingLinkMode: advertisingLinkMode,
						CurrentSpeedBps:     currentSpeedBps,
						CurrentDuplex:       currentDuplex,
						Features:            features,
						Mtu:                 mtu,
						BmcInterface:        bmcInterface,
						TenantId:            tenantId,
					},
				},
			}

			ctx := context.Background()

			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}

func FuzzCreateHostusb(f *testing.F) {
	invClient := getTestClient(f)
	f.Add(
		"kind",
		"ownerId",
		"idvendor",
		"idproduct",
		uint32(100),
		uint32(100),
		"class",
		"serial",
		"deviceName",
		"tenant-12345678",
	)
	f.Add(
		" ",
		"ownerId1",
		"asdasdasd",
		"idproduct1",
		uint32(200),
		uint32(200),
		"12",
		"3333333",
		"deviceName1",
		"tenant-87654321",
	)
	f.Add(
		"kind2",
		"ownerId2",
		"idvendor2",
		"44444",
		uint32(300),
		uint32(300),
		"class2",
		" ",
		"deviceName2",
		"tenant-11223344",
	)
	f.Add(
		" ",
		"ownerId3",
		"idvendor3",
		"123.",
		uint32(400),
		uint32(400),
		"class3",
		"serial3",
		"///",
		"tenant-55667788",
	)
	f.Fuzz(
		func(t *testing.T,
			kind string,
			ownerId string,
			idvendor string,
			idproduct string,
			bus uint32,
			addr uint32,
			class string,
			serial string,
			deviceName string,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_Hostusb{
					Hostusb: &computev1.HostusbResource{
						Kind:       kind,
						OwnerId:    ownerId,
						Idvendor:   idvendor,
						Idproduct:  idproduct,
						Bus:        bus,
						Addr:       addr,
						Class:      class,
						Serial:     serial,
						DeviceName: deviceName,
						TenantId:   tenantId,
					},
				},
			}

			ctx := context.Background()

			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}

func FuzzCreateHostgpu(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("pciId", "product", "vendor", "description", "deviceName", "features", "tenant-12345678")
	f.Add("pciId1", "product1", "vendor1", "description1", "deviceName1", "features1", "tenant-87654321")
	f.Add("pciId2", "product2", "vendor2", "description2", "deviceName2", "features2", "tenant-11223344")
	f.Add("pciId3", "product3", "vendor3", "description3", "deviceName3", "features3", "tenant-55667788")
	f.Fuzz(
		func(t *testing.T,
			pciId string,
			product string,
			vendor string,
			description string,
			deviceName string,
			features string,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_Hostgpu{
					Hostgpu: &computev1.HostgpuResource{
						PciId:       pciId,
						Product:     product,
						Vendor:      vendor,
						Description: description,
						DeviceName:  deviceName,
						Features:    features,
						TenantId:    tenantId,
					},
				},
			}

			ctx := context.Background()

			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}

func FuzzCreateInstance(f *testing.F) {
	invClient := getTestClient(f)
	f.Add(
		"name",
		uint64(100000),
		uint32(100),
		uint64(100000),
		"12313123",
		uint64(100000),
		"1111",
		uint64(100000),
		"updateStatus",
		uint64(100000),
		"updateStatusDetail",
		"tenant-12345678",
	)
	f.Add(
		" ",
		uint64(200000),
		uint32(200),
		uint64(200000),
		"instanceStatus1",
		uint64(200000),
		"AAAAAA",
		uint64(200000),
		"updateStatus1",
		uint64(200000),
		"updateStatusDetail1",
		"tenant-87654321",
	)
	f.Add(
		"n",
		uint64(300000),
		uint32(300),
		uint64(300000),
		"instanceStatus2",
		uint64(300000),
		"provisioningStatus2",
		uint64(300000),
		"updateStatus2",
		uint64(300000),
		"updateStatusDetail2",
		"tenant-11223344",
	)
	f.Add(
		"CCCC",
		uint64(400000),
		uint32(400),
		uint64(400000),
		"instanceStatus3",
		uint64(400000),
		"provisioningStatus3",
		uint64(400000),
		"updateStatus3",
		uint64(400000),
		" ",
		"tenant-55667788",
	)
	f.Fuzz(
		func(t *testing.T,
			name string,
			vmMemoryBytes uint64,
			vmCpuCores uint32,
			vmStorageBytes uint64,
			instanceStatus string,
			instanceStatusTimestamp uint64,
			provisioningStatus string,
			provisioningStatusTimestamp uint64,
			updateStatus string,
			updateStatusTimestamp uint64,
			updateStatusDetail string,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_Instance{
					Instance: &computev1.InstanceResource{
						Name:                        name,
						VmMemoryBytes:               vmMemoryBytes,
						VmCpuCores:                  vmCpuCores,
						VmStorageBytes:              vmStorageBytes,
						InstanceStatus:              instanceStatus,
						InstanceStatusTimestamp:     instanceStatusTimestamp,
						ProvisioningStatus:          provisioningStatus,
						ProvisioningStatusTimestamp: provisioningStatusTimestamp,
						UpdateStatus:                updateStatus,
						UpdateStatusTimestamp:       updateStatusTimestamp,
						TenantId:                    tenantId,
					},
				},
			}

			ctx := context.Background()

			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}

func FuzzCreateIpaddress(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("10.10.10.10", "statusDetail", "tenant-12345678")
	f.Add("20.20.20.20", "aAAA", "tenant-87654321")
	f.Add("30.30.30.30", "12", "tenant-11223344")
	f.Add("40.40.40.40", " ", "tenant-55667788")
	f.Fuzz(func(t *testing.T,
		address string, statusDetail string, tenantId string,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_Ipaddress{
				Ipaddress: &networkv1.IPAddressResource{
					Address:      address,
					StatusDetail: statusDetail,
					TenantId:     tenantId,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")
			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateNetworkSegment(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("name", int32(100000), "tenant-12345678")
	f.Add("aAAA", int32(200000), "tenant-87654321")
	f.Add(" ", int32(300000), "tenant-11223344")
	f.Add("2", int32(400000), "tenant-55667788")
	f.Fuzz(func(t *testing.T,
		name string, vlanId int32, tenantId string,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_NetworkSegment{
				NetworkSegment: &networkv1.NetworkSegment{
					Name:     name,
					VlanId:   vlanId,
					TenantId: tenantId,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")
			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateNetlink(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("name", "providerStatus", "tenant-12345678")
	f.Add("BBBB", "providerStatus1", "tenant-87654321")
	f.Add("2", "providerStatus2", "tenant-11223344")
	f.Add("asd", "providerStatus3", "tenant-55667788")
	f.Fuzz(func(t *testing.T,
		name, providerStatus, tenantId string,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_Netlink{
				Netlink: &networkv1.NetlinkResource{
					Name:           name,
					ProviderStatus: providerStatus,
					TenantId:       tenantId,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")
			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateEndpoint(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("name", "kind", "tenant-12345678")
	f.Add("AAA", "kind1", "tenant-87654321")
	f.Add("2", "2222", "tenant-11223344")
	f.Add(" ", "333", "tenant-55667788")
	f.Fuzz(func(t *testing.T,
		name, kind, tenantId string,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_Endpoint{
				Endpoint: &networkv1.EndpointResource{
					Name:     name,
					Kind:     kind,
					TenantId: tenantId,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")
			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateOs(f *testing.F) {
	invClient := getTestClient(f)
	f.Add(
		"name",
		"architecture",
		"kernelCommand",
		"http://",
		"123132",
		inv_testing.GenerateRandomSha256(),
		"profileName",
		"profileVersion",
		"installedPackages",
		"tenant-12345678",
	)
	f.Add(
		"name1",
		"architecture1",
		"-asdda/dsa ",
		"1111",
		" ",
		inv_testing.GenerateRandomSha256(),
		"profileName1",
		"profileVersion1",
		"installedPackages1",
		"tenant-87654321",
	)
	f.Add(
		"name2",
		"architecture2",
		"kernelCommand2",
		"imageUrl2",
		"BBBB",
		inv_testing.GenerateRandomSha256(),
		"profileName2",
		"profileVersion2",
		"installedPackages2",
		"tenant-11223344",
	)
	f.Add(
		"name3",
		"xxxxx",
		"dasdasd",
		"imageUrl3",
		" ",
		inv_testing.GenerateRandomSha256(),
		"profileName3",
		"adasda ",
		"installedPackages3",
		"tenant-55667788",
	)
	f.Fuzz(
		func(t *testing.T,
			name string,
			architecture string,
			kernelCommand string,
			imageUrl string,
			imageId string,
			sha256 string,
			profileName string,
			profileVersion string,
			installedPackages string,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_Os{
					Os: &osv1.OperatingSystemResource{
						Name:              name,
						Architecture:      architecture,
						ImageUrl:          imageUrl,
						ImageId:           imageId,
						Sha256:            sha256,
						ProfileName:       profileName,
						ProfileVersion:    profileVersion,
						InstalledPackages: installedPackages,
						TenantId:          tenantId,
					},
				},
			}

			ctx := context.Background()

			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}

func FuzzCreateSingleschedule(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("name", uint64(100000), uint64(100000), "tenant-12345678")
	f.Add("na me1", uint64(200000), uint64(200000), "tenant-87654321")
	f.Add("1", uint64(300000), uint64(300000), "tenant-11223344")
	f.Add("AAA", uint64(400000), uint64(400000), "tenant-55667788")
	f.Fuzz(func(t *testing.T,
		name string,
		startSeconds uint64,
		endSeconds uint64,
		tenantId string,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_Singleschedule{
				Singleschedule: &schedulev1.SingleScheduleResource{
					Name:         name,
					StartSeconds: startSeconds,
					EndSeconds:   endSeconds,
					TenantId:     tenantId,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")
			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateRepeatedschedule(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("name", uint32(100000), "10", "20", "1", "10", ".", "tenant-12345678")
	f.Add("AAAA", uint32(200000), "11", "AAAA", " ", "11", "432432423423", "tenant-87654321")
	f.Add("2", uint32(300000), "12", "22", "3", "BBBB", "7", "tenant-11223344")
	f.Add(" ", uint32(400000), "13", "12312313", "4", "13", "8", "tenant-55667788")
	f.Fuzz(
		func(t *testing.T,
			name string,
			durationSeconds uint32,
			cronMinutes string,
			cronHours string,
			cronDayMonth string,
			cronMonth string,
			cronDayWeek string,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_Repeatedschedule{
					Repeatedschedule: &schedulev1.RepeatedScheduleResource{
						Name:            name,
						DurationSeconds: durationSeconds,
						CronMinutes:     cronMinutes,
						CronHours:       cronHours,
						CronDayMonth:    cronDayMonth,
						CronMonth:       cronMonth,
						CronDayWeek:     cronDayWeek,
						TenantId:        tenantId,
					},
				},
			}

			ctx := context.Background()

			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}

func FuzzCreateWorkloadResource(f *testing.F) {
	invClient := getTestClient(f)
	f.Add(
		"name",
		"externalId",
		"status",
		`[{"key":"key1-test","value":"region_key1_lvl4-test"}]`,
		"tenant-12345678",
	)
	f.Add(
		"name1",
		"adadsadsa",
		" ",
		`[{"key":"key2-test","value":"region_key2_lvl4-test"}]`,
		"tenant-87654321",
	)
	f.Add(
		"name2",
		"AAAAAA",
		"2",
		`[{"key":"key3-test","value":"region_key3_lvl4-test"}]`,
		"tenant-11223344",
	)
	f.Add(
		"name3",
		" ",
		"3",
		`[{"key":"key4-test","value":"region_key4_lvl4-test"}]`,
		"tenant-55667788",
	)
	f.Fuzz(
		func(t *testing.T,
			name string,
			externalId string,
			status string,
			metadata string,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_Workload{
					Workload: &computev1.WorkloadResource{
						Name:       name,
						ExternalId: externalId,
						Status:     status,
						Metadata:   metadata,
						TenantId:   tenantId,
					},
				},
			}

			ctx := context.Background()

			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}

func FuzzCreateWorkloadMember(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("tenant-12345678")
	f.Add("tenant-87654321")
	f.Add("tenant-11223344")
	f.Add("tenant-55667788")
	f.Fuzz(func(t *testing.T,
		tenantId string,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_WorkloadMember{
				WorkloadMember: &computev1.WorkloadMember{
					TenantId: tenantId,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")
			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateTelemetryGroup(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("name", "tenant-12345678")
	f.Add(" ", "tenant-87654321")
	f.Add("s", "tenant-11223344")
	f.Add("2222", "tenant-55667788")
	f.Fuzz(func(t *testing.T, name string,
		tenantId string,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_TelemetryGroup{
				TelemetryGroup: &telemetryv1.TelemetryGroupResource{
					Name:     name,
					TenantId: tenantId,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")
			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateTelemetryProfile(f *testing.F) {
	invClient := getTestClient(f)
	f.Add(uint32(100000), "tenant-12345678")
	f.Add(uint32(200000), "tenant-87654321")
	f.Add(uint32(300000), "tenant-11223344")
	f.Add(uint32(400000), "tenant-55667788")
	f.Fuzz(func(t *testing.T, metricsInterval uint32,
		tenantId string,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_TelemetryProfile{
				TelemetryProfile: &telemetryv1.TelemetryProfile{
					MetricsInterval: metricsInterval,
					TenantId:        tenantId,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")
			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateTenant(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("tenant-12345678", true)
	f.Add("tenant-87654321", false)
	f.Add("tenant-11223344", true)
	f.Add("tenant-55667788", false)
	f.Fuzz(func(t *testing.T, name string,
		watcherOsmanager bool,
	) {
		res := &invv1.Resource{
			Resource: &invv1.Resource_Tenant{
				Tenant: &tenantv1.Tenant{
					TenantId:         name,
					WatcherOsmanager: watcherOsmanager,
				},
			},
		}

		ctx := context.Background()
		created, err := invClient.Create(ctx, res)
		if created == nil || err != nil {
			log.Debug().Msg("create fuzz failed")
			if checkError(err) {
				t.Errorf("%v", err.Error())
			}
		} else {
			log.Debug().Msg("create fuzz ok")
			require.NoError(t, err)
		}
	})
}

func FuzzCreateRemoteAccess(f *testing.F) {
	invClient := getTestClient(f)
	f.Add(uint64(100000), uint32(100000), "user", "configurationStatus", uint64(100000), "tenant-12345678")
	f.Add(uint64(200000), uint32(200000), "    ", "adsadas", uint64(200000), "tenant-87654321")
	f.Add(uint64(300000), uint32(300000), "1", "@@@@", uint64(300000), "tenant-11223344")
	f.Add(uint64(400000), uint32(400000), " ", "2222", uint64(400000), "tenant-55667788")
	f.Fuzz(
		func(t *testing.T,
			expirationTimestamp uint64,
			localPort uint32,
			user string,
			configurationStatus string,
			configurationStatusTimestamp uint64,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_RemoteAccess{
					RemoteAccess: &remotev1.RemoteAccessConfiguration{
						ExpirationTimestamp:          expirationTimestamp,
						LocalPort:                    localPort,
						User:                         user,
						ConfigurationStatus:          configurationStatus,
						ConfigurationStatusTimestamp: configurationStatusTimestamp,
						TenantId:                     tenantId,
					},
				},
			}

			ctx := context.Background()
			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}

func FuzzCreateLocalAccount(f *testing.F) {
	invClient := getTestClient(f)
	f.Add("user", "configurationStatus", "tenant-12345678")
	f.Add("    ", "adsadas", "tenant-87654321")
	f.Add("1", "@@@@", "tenant-11223344")
	f.Add(" ", "2222", "tenant-55667788")
	f.Fuzz(
		func(t *testing.T,
			username string,
			sshKey string,
			tenantId string,
		) {
			res := &invv1.Resource{
				Resource: &invv1.Resource_LocalAccount{
					LocalAccount: &localaccountv1.LocalAccountResource{
						Username: username,
						SshKey:   sshKey,
						TenantId: tenantId,
					},
				},
			}

			ctx := context.Background()
			created, err := invClient.Create(ctx, res)
			if created == nil || err != nil {
				log.Debug().Msg("create fuzz failed")
				if checkError(err) {
					t.Errorf("%v", err.Error())
				}
			} else {
				log.Debug().Msg("create fuzz ok")
				require.NoError(t, err)
			}
		},
	)
}
