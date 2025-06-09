// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package policy_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	network_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	telemetry_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/policy"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
)

var (
	bundlePath       = "out/policy_bundle.tar.gz"
	regionID         = "region-12345678"
	siteID           = "site-12345678"
	hostID           = "host-12345678"
	ipaddrID         = "ipaddr-12345678"
	tenantID         = "tenant-12345678"
	providerID       = "provider-12345678"
	telemetrygroupID = "telemetrygroup-12345678"
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

	inv_testing.CreatePolicyBundle(projectRoot + "/out")
	run := m.Run() // run all tests
	os.Exit(run)
}

func loadPolicyBundle(bundlePath string) (*policy.Policy, error) {
	pwd, err := os.Getwd()
	if err != nil {
		err = fmt.Errorf("could not get current directory pwd error %s", err.Error())
		return nil, err
	}
	bundleFullPath := filepath.Join(pwd, "..", "..", bundlePath)
	pol, err := policy.New(bundleFullPath)
	if err != nil {
		return nil, err
	}

	return pol, nil
}

func TestPolicyVerifyCreate(t *testing.T) { // table-driven test
	testCases := map[string]struct {
		cliendKind inv_v1.ClientKind
		resource   *inv_v1.Resource
		resourceID string
		valid      bool
	}{
		"Test_ClientAPI_Unset_Fail": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource:   &inv_v1.Resource{},
			valid:      false,
		},
		"Test_ClientRM_Unset_Fail": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource:   &inv_v1.Resource{},
			valid:      false,
		},
		"Test_ClientAPI_Create_Region_Success": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Region{
					Region: &location_v1.RegionResource{
						Name: "region",
					},
				},
			},
			resourceID: regionID,
			valid:      true,
		},
		// this should fail eventually
		"Test_ClientRM_Create_Region_Fail": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Region{
					Region: &location_v1.RegionResource{
						Name: "region",
					},
				},
			},
			resourceID: regionID,
			valid:      true,
		},
		"Test_ClientAPI_Create_Site_Success": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Site{
					Site: &location_v1.SiteResource{
						Name:   "site",
						Region: &location_v1.RegionResource{ResourceId: "region-12345678"},
					},
				},
			},
			resourceID: siteID,
			valid:      true,
		},
		// this should fail eventually
		"Test_ClientRM_Create_Site_Fail": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Site{
					Site: &location_v1.SiteResource{
						Name:   "site",
						Region: &location_v1.RegionResource{ResourceId: "region-12345678"},
					},
				},
			},
			resourceID: siteID,
			valid:      true,
		},
		"Test_ClientAPI_Create_Host_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Create_Host_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Create_Host_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
						CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Create_Host_Success2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
						CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Create_Host_Success2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
						DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Create_Host_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
						DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Create_Host_Fail2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						SerialNumber:      "12345678",
						CurrentPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Create_Host_Success3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						SerialNumber:      "12345678",
						CurrentPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Create_Host_Success3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						SerialNumber:      "12345678",
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Create_Host_Fail3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						SerialNumber:      "12345678",
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Create_Host_Fail3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
						CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Create_Host_Fail4": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
						CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Create_Host_Fail4": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						SerialNumber:      "12345678",
						CurrentState:      computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Create_Host_Fail5": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						SerialNumber:      "12345678",
						CurrentState:      computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Create_Host_Fail5": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						SerialNumber:      "12345678",
						CurrentState:      computev1.HostState_HOST_STATE_ONBOARDED,
						CurrentPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Create_Host_Success4": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						SerialNumber:      "12345678",
						CurrentState:      computev1.HostState_HOST_STATE_ONBOARDED,
						CurrentPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Create_Host_Success4": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						SerialNumber:      "12345678",
						DesiredState:      computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Create_Host_Fail6": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						SerialNumber:      "12345678",
						DesiredState:      computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Create_Host_Success5": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Create_Host_Success5": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Register_Host_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						Uuid:         "12345678",
						SerialNumber: "12345678",
						DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Register_Host_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						Uuid:         "12345678",
						SerialNumber: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Register_Host_Success2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Register_Host_Success2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Register_Host_Success3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name: "host",
						Uuid: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Register_Host_Success3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name: "host",
						Uuid: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Register_Host_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name: "host",
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Register_Host_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name: "host",
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Create_IPAddr_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Ipaddress{
					Ipaddress: &network_v1.IPAddressResource{
						Address:      "192.168.1.1/24",
						CurrentState: network_v1.IPAddressState_IP_ADDRESS_STATE_CONFIGURED,
					},
				},
			},
			resourceID: ipaddrID,
			valid:      false,
		},
		"Test_ClientRM_Create_IPAddr_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Ipaddress{
					Ipaddress: &network_v1.IPAddressResource{
						Address:      "192.168.1.1/24",
						CurrentState: network_v1.IPAddressState_IP_ADDRESS_STATE_CONFIGURED,
					},
				},
			},
			resourceID: ipaddrID,
			valid:      true,
		},
		"Test_ClientAPI_Create_IPAddr_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Ipaddress{
					Ipaddress: &network_v1.IPAddressResource{
						Address:      "192.168.1.1/24",
						DesiredState: network_v1.IPAddressState_IP_ADDRESS_STATE_CONFIGURED,
					},
				},
			},
			resourceID: ipaddrID,
			valid:      true,
		},
		"Test_ClientRM_Create_IPAddr_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Ipaddress{
					Ipaddress: &network_v1.IPAddressResource{
						Address:      "192.168.1.1/24",
						DesiredState: network_v1.IPAddressState_IP_ADDRESS_STATE_CONFIGURED,
					},
				},
			},
			resourceID: ipaddrID,
			valid:      false,
		},
		"Test_ClientAPI_Create_Tenant_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Tenant{
					Tenant: &tenantv1.Tenant{
						DesiredState:     tenantv1.TenantState_TENANT_STATE_CREATED,
						WatcherOsmanager: true,
					},
				},
			},
			resourceID: tenantID,
			valid:      false,
		},
		"Test_ClientRM_Create_Tenant_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Tenant{
					Tenant: &tenantv1.Tenant{
						DesiredState:     tenantv1.TenantState_TENANT_STATE_CREATED,
						WatcherOsmanager: true,
					},
				},
			},
			resourceID: tenantID,
			valid:      false,
		},
		"Test_ClientTC_Create_Tenant_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_TENANT_CONTROLLER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Tenant{
					Tenant: &tenantv1.Tenant{
						DesiredState:     tenantv1.TenantState_TENANT_STATE_CREATED,
						WatcherOsmanager: true,
					},
				},
			},
			resourceID: tenantID,
			valid:      true,
		},
		"Test_ClientTC_Create_Provider_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_TENANT_CONTROLLER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Provider{
					Provider: &providerv1.ProviderResource{
						Name: "test-provider-name",
					},
				},
			},
			resourceID: providerID,
			valid:      true,
		},
		"Test_ClientTC_Create_TelemetryGroup_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_TENANT_CONTROLLER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_TelemetryGroup{
					TelemetryGroup: &telemetry_v1.TelemetryGroupResource{
						Name: "test-tenant-name",
					},
				},
			},
			resourceID: telemetrygroupID,
			valid:      true,
		},
		"Test_ClientTC_Create_IPAddr_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_TENANT_CONTROLLER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Ipaddress{
					Ipaddress: &network_v1.IPAddressResource{
						Address:      "192.168.1.1/24",
						DesiredState: network_v1.IPAddressState_IP_ADDRESS_STATE_CONFIGURED,
					},
				},
			},
			resourceID: ipaddrID,
			valid:      false,
		},
	}

	pol, err := loadPolicyBundle(bundlePath)
	if err != nil {
		t.Errorf("new policy instance create error %s", err.Error())
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			// Verify create first
			creReq := &inv_v1.CreateResourceRequest{Resource: testCase.resource}
			polVerifyErr := pol.Verify(testCase.cliendKind.String(), creReq)
			if testCase.valid && polVerifyErr != nil {
				t.Errorf("policy verification for create error %s", polVerifyErr.Error())
			} else if !testCase.valid && polVerifyErr == nil {
				t.Errorf("policy verification for create should have errored for test case %s", testName)
			}
		})
	}
}

func TestPolicyVerifyUpdate(t *testing.T) { // table-driven test
	testCases := map[string]struct {
		cliendKind inv_v1.ClientKind
		resource   *inv_v1.Resource
		resourceID string
		valid      bool
	}{
		"Test_ClientAPI_Unset_Fail": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource:   &inv_v1.Resource{},
			valid:      false,
		},
		"Test_ClientRM_Unset_Fail": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource:   &inv_v1.Resource{},
			valid:      false,
		},
		"Test_ClientAPI_Update_Region_Success": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Region{
					Region: &location_v1.RegionResource{
						Name: "region",
					},
				},
			},
			resourceID: regionID,
			valid:      true,
		},
		// this should fail eventually
		"Test_ClientRM_Update_Region_Fail": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Region{
					Region: &location_v1.RegionResource{
						Name: "region",
					},
				},
			},
			resourceID: regionID,
			valid:      true,
		},
		"Test_ClientAPI_Update_Site_Success": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Site{
					Site: &location_v1.SiteResource{
						Name:   "site",
						Region: &location_v1.RegionResource{ResourceId: "region-12345678"},
					},
				},
			},
			resourceID: siteID,
			valid:      true,
		},
		// this should fail eventually
		"Test_ClientRM_Update_Site_Fail": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Site{
					Site: &location_v1.SiteResource{
						Name:   "site",
						Region: &location_v1.RegionResource{ResourceId: "region-12345678"},
					},
				},
			},
			resourceID: siteID,
			valid:      true,
		},
		"Test_ClientAPI_Update_Host_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name: "host",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Update_Host_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name: "host",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Update_Host_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Update_Host_Success2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Update_Host_Success2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Update_Host_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Update_Host_Fail2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						CurrentPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Update_Host_Success3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						CurrentPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Update_Host_Success3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Update_Host_Fail3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Update_Host_Fail3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Update_Host_Fail4": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Update_Host_Fail4": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						CurrentState:      computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Update_Host_Fail5": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						CurrentState:      computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Update_Host_Fail5": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						CurrentState:      computev1.HostState_HOST_STATE_ONBOARDED,
						CurrentPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Update_Host_Success4": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						CurrentState:      computev1.HostState_HOST_STATE_ONBOARDED,
						CurrentPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Update_Host_Success4": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						DesiredState:      computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Update_Host_Fail6": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:              "host",
						DesiredState:      computev1.HostState_HOST_STATE_ONBOARDED,
						DesiredPowerState: computev1.PowerState_POWER_STATE_ON,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientRM_Update_Host_Success5": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Update_Tenant_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Tenant{
					Tenant: &tenantv1.Tenant{
						DesiredState: tenantv1.TenantState_TENANT_STATE_CREATED,
					},
				},
			},
			resourceID: tenantID,
			valid:      false,
		},
		"Test_ClientRM_Update_Tenant_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Tenant{
					Tenant: &tenantv1.Tenant{
						DesiredState: tenantv1.TenantState_TENANT_STATE_CREATED,
					},
				},
			},
			resourceID: tenantID,
			valid:      false,
		},
		"Test_ClientAPI_Update_Registered_Host_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Update_Registered_Host_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						Uuid:         "12345678",
						SerialNumber: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Update_Registered_Host_Success2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Update_Registered_Host_Success3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name: "host",
						Uuid: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientRM_Update_Registered_Host_Success4": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name: "host",
					},
				},
			},
			resourceID: hostID,
			valid:      true,
		},
		"Test_ClientAPI_Update_Registered_Host_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						Uuid:         "12345678",
						SerialNumber: "12345678",
						DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Update_Registered_Host_Fail2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name: "host",
						Uuid: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Update_Registered_Host_Fail3": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{
					Host: &computev1.HostResource{
						Name:         "host",
						SerialNumber: "12345678",
					},
				},
			},
			resourceID: hostID,
			valid:      false,
		},
		"Test_ClientAPI_Update_IPAddr_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Ipaddress{
					Ipaddress: &network_v1.IPAddressResource{
						Address:      "192.168.1.1/24",
						CurrentState: network_v1.IPAddressState_IP_ADDRESS_STATE_CONFIGURED,
					},
				},
			},
			resourceID: ipaddrID,
			valid:      false,
		},
		"Test_ClientRM_Update_IPAddr_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Ipaddress{
					Ipaddress: &network_v1.IPAddressResource{
						Address:      "192.168.1.1/24",
						CurrentState: network_v1.IPAddressState_IP_ADDRESS_STATE_CONFIGURED,
					},
				},
			},
			resourceID: ipaddrID,
			valid:      true,
		},
		"Test_ClientAPI_Update_IPAddr_Success1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Ipaddress{
					Ipaddress: &network_v1.IPAddressResource{
						Address:      "192.168.1.1/24",
						DesiredState: network_v1.IPAddressState_IP_ADDRESS_STATE_CONFIGURED,
					},
				},
			},
			resourceID: ipaddrID,
			valid:      true,
		},
		"Test_ClientRM_Update_IPAddr_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			resource: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Ipaddress{
					Ipaddress: &network_v1.IPAddressResource{
						Address:      "192.168.1.1/24",
						DesiredState: network_v1.IPAddressState_IP_ADDRESS_STATE_CONFIGURED,
					},
				},
			},
			resourceID: ipaddrID,
			valid:      false,
		},
	}

	pol, err := loadPolicyBundle(bundlePath)
	if err != nil {
		t.Errorf("new policy instance create error %s", err.Error())
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			// Verify update
			updReq := &inv_v1.UpdateResourceRequest{
				Resource:   testCase.resource,
				ResourceId: testCase.resourceID,
			}
			polVerifyErr := pol.Verify(testCase.cliendKind.String(), updReq)
			if testCase.valid && polVerifyErr != nil {
				t.Errorf("policy verification for update error %s", polVerifyErr.Error())
			} else if !testCase.valid && polVerifyErr == nil {
				t.Errorf("policy verification for update should have errored for test case %s", testName)
			}
		})
	}
}

func TestVerifyDelete(t *testing.T) {
	testCases := map[string]struct {
		cliendKind inv_v1.ClientKind
		resourceID string
		valid      bool
	}{
		"Test_ClientAPI_Unset_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			valid:      false,
			resourceID: "",
		},
		"Test_ClientRM_Unset_Fail1": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			valid:      false,
			resourceID: "",
		},
		"Test_ClientAPI_Unset_Fail2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			valid:      false,
		},
		"Test_ClientRM_Unset_Fail2": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			valid:      false,
		},
		"Test_ClientAPI_RegionSuccess": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			valid:      true,
			resourceID: regionID,
		},
		"Test_ClientRM_RegionSuccess": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			valid:      true,
			resourceID: regionID,
		},
		"Test_ClientAPI_SiteSuccess": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			valid:      true,
			resourceID: siteID,
		},
		"Test_ClientRM_SiteSuccess": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			valid:      true,
			resourceID: siteID,
		},
		"Test_ClientAPI_HostSuccess": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_API,
			valid:      true,
			resourceID: hostID,
		},
		"Test_ClientRM_HostSuccess": {
			cliendKind: inv_v1.ClientKind_CLIENT_KIND_RESOURCE_MANAGER,
			valid:      true,
			resourceID: hostID,
		},
	}

	pol, err := loadPolicyBundle(bundlePath)
	if err != nil {
		t.Errorf("new policy instance create error %s", err.Error())
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			// Verify delete
			delReq := &inv_v1.DeleteResourceRequest{ResourceId: testCase.resourceID}
			polVerifyErr := pol.Verify(testCase.cliendKind.String(), delReq)
			if testCase.valid && polVerifyErr != nil {
				t.Errorf("policy verification for delete error %s", polVerifyErr.Error())
			} else if !testCase.valid && polVerifyErr == nil {
				t.Errorf("policy verification for delete should have errored for test case %s", testName)
			}
		})
	}
}

func TestClientTenantController(t *testing.T) {
	suite.Run(t, new(tenantControllerClientTS))
}

type tenantControllerClientTS struct {
	suite.Suite
	policy     *policy.Policy
	clientType inv_v1.ClientKind
}

func (ts *tenantControllerClientTS) SetupSuite() {
	pb, err := loadPolicyBundle(bundlePath)
	ts.Require().NoError(err, "new policy instance create error")
	ts.policy = pb
	ts.clientType = inv_v1.ClientKind_CLIENT_KIND_TENANT_CONTROLLER
}

func (ts *tenantControllerClientTS) TestCreateTenant() {
	createTenantReq := &inv_v1.CreateResourceRequest{Resource: &inv_v1.Resource{
		Resource: &inv_v1.Resource_Tenant{Tenant: &tenantv1.Tenant{
			DesiredState:     tenantv1.TenantState_TENANT_STATE_CREATED,
			WatcherOsmanager: true,
		}},
	}}
	err := ts.policy.Verify(ts.clientType.String(), createTenantReq)
	ts.Assert().NoError(err)
}

func (ts *tenantControllerClientTS) TestUpdateTenant() {
	updateTenantReq := &inv_v1.UpdateResourceRequest{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Tenant{
				Tenant: &tenantv1.Tenant{CurrentState: tenantv1.TenantState_TENANT_STATE_CREATED},
			},
		},
		ResourceId: "tenant-12345678",
	}
	err := ts.policy.Verify(ts.clientType.String(), updateTenantReq)
	ts.Assert().NoError(err)
}

func (ts *tenantControllerClientTS) TestDeleteTenant() {
	deleteTenantReq := &inv_v1.DeleteResourceRequest{ResourceId: "tenant-12345678"}
	err := ts.policy.Verify(ts.clientType.String(), deleteTenantReq)
	ts.Assert().NoError(err)
}

func (ts *tenantControllerClientTS) TestCreateProvider() {
	creationReq := &inv_v1.CreateResourceRequest{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_Provider{
				Provider: &providerv1.ProviderResource{
					Name: "anyProvider",
				},
			},
		},
	}
	ts.Assert().NoError(ts.policy.Verify(ts.clientType.String(), creationReq))
}

func (ts *tenantControllerClientTS) TestCreateTelemetryGroupMetrics() {
	creationReq := &inv_v1.CreateResourceRequest{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_TelemetryGroup{
				TelemetryGroup: &telemetry_v1.TelemetryGroupResource{
					Name: "tg-metrics",
					Kind: telemetry_v1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
				},
			},
		},
	}
	ts.Assert().NoError(ts.policy.Verify(ts.clientType.String(), creationReq))
}

func (ts *tenantControllerClientTS) TestCreateTelemetryGroupLogs() {
	creationReq := &inv_v1.CreateResourceRequest{
		Resource: &inv_v1.Resource{
			Resource: &inv_v1.Resource_TelemetryGroup{
				TelemetryGroup: &telemetry_v1.TelemetryGroupResource{
					Name: "tg-metrics",
					Kind: telemetry_v1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
				},
			},
		},
	}
	ts.Assert().NoError(ts.policy.Verify(ts.clientType.String(), creationReq))
}
