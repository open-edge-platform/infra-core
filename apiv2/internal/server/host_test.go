// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	commonv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/common/v1"
	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	locationv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/location/v1"
	statusv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/status/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_server "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	inv_networkv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	inv_statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

// Write an example of inventory resource with a Host resource filled with all fields.
var exampleInvHostResource = &inv_computev1.HostResource{
	ResourceId:   "host-12345678",
	Name:         "example-host",
	DesiredState: inv_computev1.HostState_HOST_STATE_REGISTERED,
	CurrentState: inv_computev1.HostState_HOST_STATE_ONBOARDED,
	Site: &inv_locationv1.SiteResource{
		ResourceId: "site-12345678",
	},
	Note:            "Example note",
	SerialNumber:    "SN12345678",
	Uuid:            "uuid-1234-5678-9012-3456",
	MemoryBytes:     16384,
	CpuModel:        "Intel Xeon",
	CpuSockets:      2,
	CpuCores:        16,
	CpuCapabilities: "capability1,capability2",
	CpuArchitecture: "x86_64",
	CpuThreads:      32,
	CpuTopology:     "topology-json",
	BmcKind:         inv_computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_IPMI,
	BmcIp:           "192.168.0.1",
	Hostname:        "example-hostname",
	ProductName:     "Example Product",
	BiosVersion:     "1.0.0",
	BiosReleaseDate: "2023-01-01",
	BiosVendor:      "Example Vendor",
	HostStorages: []*inv_computev1.HoststorageResource{
		{
			ResourceId:    "storage-12345678",
			Wwid:          "wwid-1234",
			Serial:        "serial-1234",
			Vendor:        "vendor-1234",
			Model:         "model-1234",
			CapacityBytes: 1024,
			DeviceName:    "sda",
		},
	},
	HostNics: []*inv_computev1.HostnicResource{
		{
			ResourceId:    "nic-12345678",
			DeviceName:    "eth0",
			PciIdentifier: "pci-1234",
			MacAddr:       "00:11:22:33:44:55",
			SriovEnabled:  true,
			SriovVfsNum:   8,
			SriovVfsTotal: 16,
			Features:      "feature1,feature2",
			Mtu:           1500,
			LinkState:     inv_computev1.NetworkInterfaceLinkState_NETWORK_INTERFACE_LINK_STATE_UP,
			BmcInterface:  true,
		},
	},
	HostUsbs: []*inv_computev1.HostusbResource{
		{
			ResourceId: "usb-12345678",
			Idvendor:   "vendor-1234",
			Idproduct:  "product-1234",
			Bus:        123,
			Addr:       123,
			Class:      "class-1234",
			Serial:     "serial-1234",
			DeviceName: "usb0",
		},
	},
	HostGpus: []*inv_computev1.HostgpuResource{
		{
			ResourceId:  "gpu-12345678",
			PciId:       "pci-1234",
			Product:     "product-1234",
			Vendor:      "vendor-1234",
			Description: "description-1234",
			DeviceName:  "gpu0",
			Features:    "feature1,feature2",
		},
	},
	Instance: &inv_computev1.InstanceResource{
		ResourceId:                  "instance-12345678",
		ProvisioningStatus:          "provisioned",
		ProvisioningStatusIndicator: inv_statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		ProvisioningStatusTimestamp: 1234567890,
		UpdateStatus:                "updating",
		UpdateStatusIndicator:       inv_statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		UpdateStatusTimestamp:       1234567890,
		InstanceStatus:              "running",
		InstanceStatusIndicator:     inv_statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		InstanceStatusTimestamp:     1234567890,
	},
	Metadata:                    `[{"key":"key1","value":"value1"}]`,
	OnboardingStatus:            "onboarding",
	OnboardingStatusIndicator:   inv_statusv1.StatusIndication_STATUS_INDICATION_IDLE,
	OnboardingStatusTimestamp:   1234567890,
	RegistrationStatus:          "registered",
	RegistrationStatusIndicator: inv_statusv1.StatusIndication_STATUS_INDICATION_IDLE,
	RegistrationStatusTimestamp: 1234567890,
	HostStatus:                  "running",
	HostStatusIndicator:         inv_statusv1.StatusIndication_STATUS_INDICATION_IDLE,
	HostStatusTimestamp:         1234567890,
}

// Write an example of API resource with a Host resource filled with all fields.
var exampleAPIHostResource = &computev1.HostResource{
	ResourceId:   "host-12345678",
	Name:         "example-host",
	DesiredState: computev1.HostState_HOST_STATE_REGISTERED,
	CurrentState: computev1.HostState_HOST_STATE_ONBOARDED,
	Site: &locationv1.SiteResource{
		ResourceId: "site-12345678",
	},
	Note:            "Example note",
	SerialNumber:    "SN12345678",
	Uuid:            "uuid-1234-5678-9012-3456",
	MemoryBytes:     "16384",
	CpuModel:        "Intel Xeon",
	CpuSockets:      2,
	CpuCores:        16,
	CpuCapabilities: "capability1,capability2",
	CpuArchitecture: "x86_64",
	CpuThreads:      32,
	CpuTopology:     "topology-json",
	BmcKind:         computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_IPMI,
	BmcIp:           "192.168.0.1",
	Hostname:        "example-hostname",
	ProductName:     "Example Product",
	BiosVersion:     "1.0.0",
	BiosReleaseDate: "2023-01-01",
	BiosVendor:      "Example Vendor",
	HostStorages: []*computev1.HoststorageResource{
		{
			ResourceId:    "storage-12345678",
			Wwid:          "wwid-1234",
			Serial:        "serial-1234",
			Vendor:        "vendor-1234",
			Model:         "model-1234",
			CapacityBytes: "1024",
			DeviceName:    "sda",
		},
	},
	HostNics: []*computev1.HostnicResource{
		{
			ResourceId:    "nic-12345678",
			DeviceName:    "eth0",
			PciIdentifier: "pci-1234",
			MacAddr:       "00:11:22:33:44:55",
			SriovEnabled:  true,
			SriovVfsNum:   8,
			SriovVfsTotal: 16,
			Features:      "feature1,feature2",
			Mtu:           1500,
			LinkState:     computev1.NetworkInterfaceLinkState_NETWORK_INTERFACE_LINK_STATE_UP,
			BmcInterface:  true,
		},
	},
	HostUsbs: []*computev1.HostusbResource{
		{
			ResourceId: "usb-12345678",
			Idvendor:   "vendor-1234",
			Idproduct:  "product-1234",
			Bus:        123,
			Addr:       123,
			Class:      "class-1234",
			Serial:     "serial-1234",
			DeviceName: "usb0",
		},
	},
	HostGpus: []*computev1.HostgpuResource{
		{
			ResourceId:  "gpu-12345678",
			PciId:       "pci-1234",
			Product:     "product-1234",
			Vendor:      "vendor-1234",
			Description: "description-1234",
			DeviceName:  "gpu0",
			Features:    "feature1,feature2",
		},
	},
	Instance: &computev1.InstanceResource{
		ResourceId:                  "instance-12345678",
		InstanceStatus:              "running",
		InstanceStatusIndicator:     statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		InstanceStatusTimestamp:     1234567890,
		ProvisioningStatus:          "provisioned",
		ProvisioningStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		ProvisioningStatusTimestamp: 1234567890,
		UpdateStatus:                "updating",
		UpdateStatusIndicator:       statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		UpdateStatusTimestamp:       1234567890,
	},
	SiteId:                      "site-12345678",
	Metadata:                    []*commonv1.MetadataItem{{Key: "key1", Value: "value1"}},
	OnboardingStatus:            "onboarding",
	OnboardingStatusIndicator:   statusv1.StatusIndication_STATUS_INDICATION_IDLE,
	OnboardingStatusTimestamp:   1234567890,
	RegistrationStatus:          "registered",
	RegistrationStatusIndicator: statusv1.StatusIndication_STATUS_INDICATION_IDLE,
	RegistrationStatusTimestamp: 1234567890,
	HostStatus:                  "running",
	HostStatusIndicator:         statusv1.StatusIndication_STATUS_INDICATION_IDLE,
	HostStatusTimestamp:         1234567890,
}

//nolint:funlen // Test functions are long but necessary to test all the cases.
func TestHost_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateHostRequest
		wantErr bool
	}{
		{
			name: "Create Host",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Host{
								Host: &inv_computev1.HostResource{
									ResourceId: "host-12345678",
									Name:       "example-host",
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateHostRequest{
				Host: &computev1.HostResource{
					Name: "example-host",
				},
			},
			wantErr: false,
		},
		{
			name: "Create Host with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateHostRequest{},
			wantErr: true,
		},
		{
			name: "Create Host with all fields",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Host{
								Host: exampleInvHostResource,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateHostRequest{
				Host: &computev1.HostResource{
					Name: "example-host",
				},
			},
			wantErr: false,
		},
		{
			name: "Create Host with all fields and error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateHostRequest{
				Host: &computev1.HostResource{
					Name: "example-host",
				},
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateHost(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateHost() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateHost() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateHost() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetHost(), reply)
		})
	}
}

func TestHost_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetHostRequest
		wantErr bool
	}{
		{
			name: "Get Host",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "host-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_Host{
									Host: exampleInvHostResource,
								},
							},
						}, nil).Once(),
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_Ipaddress{
											Ipaddress: &inv_networkv1.IPAddressResource{
												ResourceId: "ipaddress-12345678",
											},
										},
									},
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetHostRequest{
				ResourceId: "host-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get Host with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "host-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetHostRequest{
				ResourceId: "host-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetHost(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetHost() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetHost() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetHost() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIHostResource, reply)
		})
	}
}

func TestHost_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListHostsRequest
		wantErr bool
	}{
		{
			name: "List Hosts",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_Host{
											Host: exampleInvHostResource,
										},
									},
								},
							},
							TotalElements: 1,
							HasNext:       false,
						}, nil).Once(),
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_Ipaddress{
											Ipaddress: &inv_networkv1.IPAddressResource{
												ResourceId: "ipaddress-12345678",
											},
										},
									},
								},
							},
							TotalElements: 0,
							HasNext:       false,
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListHostsRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List Hosts with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListHostsRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.ListHosts(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListHosts() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListHosts() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListHosts() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetHosts()) != 1 {
				t.Errorf("ListHosts() got %v hosts, want 1", len(reply.GetHosts()))
			}
			compareProtoMessages(t, exampleAPIHostResource, reply.GetHosts()[0])
		})
	}
}

func TestHost_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateHostRequest
		wantErr bool
	}{
		{
			name: "Update Host",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "host-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Host{
								Host: exampleInvHostResource,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateHostRequest{
				ResourceId: "host-12345678",
				Host:       exampleAPIHostResource,
			},
			wantErr: false,
		},
		{
			name: "Update Host with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "host-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateHostRequest{
				ResourceId: "host-12345678",
				Host:       exampleAPIHostResource,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateHost(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateHost() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateHost() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateHost() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIHostResource, reply)
		})
	}
}

func TestHost_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteHostRequest
		wantErr bool
	}{
		{
			name: "Delete Host",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "host-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteHostRequest{
				ResourceId: "host-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete Host with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "host-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteHostRequest{
				ResourceId: "host-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteHost(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteHost() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteHost() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteHost() got reply = nil, want non-nil")
				return
			}
		})
	}
}

func TestHost_Summary(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	region1 := inv_testing.CreateRegion(t, nil)
	region2 := inv_testing.CreateRegion(t, region1)
	site1 := inv_testing.CreateSiteWithArgs(t, "SJI1", 100, 100, "", region2, nil, nil)
	provider1 := inv_testing.CreateProvider(t, "Test Provider1")
	os1 := inv_testing.CreateOs(t)
	uuidH1 := uuid.NewString()

	createresreq1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &inv_computev1.HostResource{
				Name:         "Test Host 3",
				DesiredState: inv_computev1.HostState_HOST_STATE_ONBOARDED,

				Site:         site1,
				Provider:     provider1,
				HardwareKind: "XDgen4",
				SerialNumber: "1001",
				MemoryBytes:  64 * util.Gigabyte,
				Uuid:         uuidH1,

				CpuModel:        "12th Gen Intel(R) Core(TM) i9-12900",
				CpuSockets:      1,
				CpuCores:        14,
				CpuCapabilities: "",
				CpuArchitecture: "x86_64",
				CpuThreads:      13,

				MgmtIp: "192.168.10.13",

				BmcKind:     inv_computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
				BmcIp:       "10.0.0.13",
				BmcUsername: "user",
				BmcPassword: "pass",
				PxeMac:      "90:49:fa:ff:ff:f3",

				Hostname: "testhost3",
				Metadata: "",
			},
		},
	}

	createresreq2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{
			Host: &inv_computev1.HostResource{
				Name:              "Test Host 1",
				DesiredState:      inv_computev1.HostState_HOST_STATE_REGISTERED,
				CurrentState:      inv_computev1.HostState_HOST_STATE_UNSPECIFIED,
				HardwareKind:      "XDgen2",
				SerialNumber:      "12345678",
				Uuid:              "E5E53D99-708D-4AF5-8378-63880FF62712",
				MemoryBytes:       64 * util.Gigabyte,
				CpuModel:          "12th Gen Intel(R) Core(TM) i9-12900",
				CpuSockets:        1,
				CpuCores:          14,
				CpuCapabilities:   "",
				CpuArchitecture:   "x86_64",
				CpuThreads:        20,
				CpuTopology:       `{"some_json":[]}`,
				MgmtIp:            "192.168.10.10",
				BmcKind:           inv_computev1.BaremetalControllerKind_BAREMETAL_CONTROLLER_KIND_PDU,
				BmcIp:             "10.0.0.10",
				BmcUsername:       "user",
				BmcPassword:       "pass",
				PxeMac:            "90:49:fa:ff:ff:ff",
				Hostname:          "testhost1",
				ProductName:       "PowerEdge R750",
				BiosVersion:       "1.0.0",
				BiosReleaseDate:   "09/14/2022",
				BiosVendor:        "Dell Inc.",
				DesiredPowerState: inv_computev1.PowerState_POWER_STATE_ON,
				Metadata:          "[{\"key\":\"cluster-name\",\"value\":\"\"},{\"key\":\"app-id\",\"value\":\"\"}]",
			},
		},
	}
	// Create hosts.
	chostResp1, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	expHost1 := *chostResp1.GetHost()
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, expHost1.GetResourceId()) })

	instance1 := inv_testing.CreateInstance(t, &expHost1, os1)

	chostResp2, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq2)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	expHost2 := *chostResp2.GetHost()
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, expHost2.GetResourceId()) })

	// Instantiate server
	server := inv_server.InventorygRPCServer{InvClient: inv_testing.TestClients[inv_testing.APIClient]}

	// Test GetHostsSummary with no filters
	response, err := server.GetHostsSummary(ctx, &restv1.GetHostSummaryRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, uint32(2), response.Total)
	assert.Equal(t, uint32(1), response.Unallocated)
	assert.Equal(t, uint32(0), response.Error)
	assert.Equal(t, uint32(0), response.Running)

	// Update host status to ERROR
	updateHost := &inv_computev1.HostResource{
		HostStatusIndicator: inv_statusv1.StatusIndication_STATUS_INDICATION_ERROR,
	}
	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{computev1.HostResourceFieldHostStatusIndicator},
	}
	updateRes := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Host{Host: updateHost},
	}

	_, err = inv_testing.TestClients[inv_testing.APIClient].Update(ctx, expHost2.GetResourceId(), fieldMask, updateRes)
	assert.NoError(t, err)

	// Test GetHostsSummary with host in error
	response, err = server.GetHostsSummary(ctx, &restv1.GetHostSummaryRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, uint32(2), response.Total)
	assert.Equal(t, uint32(1), response.Unallocated)
	assert.Equal(t, uint32(1), response.Error)
	assert.Equal(t, uint32(0), response.Running)

	// Update Instance status to Running
	updateInstance := &inv_computev1.InstanceResource{
		CurrentState: inv_computev1.InstanceState_INSTANCE_STATE_RUNNING,
	}
	fieldMask = &fieldmaskpb.FieldMask{
		Paths: []string{
			inv_computev1.InstanceResourceFieldCurrentState,
		},
	}
	updateRes = &inv_v1.Resource{
		Resource: &inv_v1.Resource_Instance{Instance: updateInstance},
	}
	updateResponse, err := inv_testing.TestClients[inv_testing.RMClient].Update(ctx, instance1.GetResourceId(), fieldMask, updateRes)
	assert.NoError(t, err)
	require.Equal(t, inv_computev1.InstanceState_INSTANCE_STATE_RUNNING, updateResponse.GetInstance().GetCurrentState())

	getResponse, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, expHost1.GetResourceId())
	assert.NoError(t, err)
	require.Equal(t, inv_statusv1.StatusIndication_STATUS_INDICATION_UNSPECIFIED, getResponse.GetResource().GetHost().GetHostStatusIndicator())
	require.Equal(t, inv_statusv1.StatusIndication_STATUS_INDICATION_UNSPECIFIED, getResponse.GetResource().GetHost().GetOnboardingStatusIndicator())
	require.Equal(t, inv_statusv1.StatusIndication_STATUS_INDICATION_UNSPECIFIED, getResponse.GetResource().GetHost().GetRegistrationStatusIndicator())
	require.Equal(t, inv_statusv1.StatusIndication_STATUS_INDICATION_UNSPECIFIED, getResponse.GetResource().GetHost().GetInstance().GetProvisioningStatusIndicator())
	require.Equal(t, inv_statusv1.StatusIndication_STATUS_INDICATION_UNSPECIFIED, getResponse.GetResource().GetHost().GetInstance().GetUpdateStatusIndicator())
	require.Equal(t, inv_statusv1.StatusIndication_STATUS_INDICATION_UNSPECIFIED, getResponse.GetResource().GetHost().GetInstance().GetTrustedAttestationStatusIndicator())
	require.Equal(t, inv_statusv1.StatusIndication_STATUS_INDICATION_UNSPECIFIED, getResponse.GetResource().GetHost().GetInstance().GetInstanceStatusIndicator())

	// Test GetHostsSummary with host in error and instance running
	response, err = server.GetHostsSummary(ctx, &restv1.GetHostSummaryRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, uint32(2), response.Total)
	assert.Equal(t, uint32(1), response.Unallocated)
	assert.Equal(t, uint32(1), response.Error)
	assert.Equal(t, uint32(1), response.Running)
}
