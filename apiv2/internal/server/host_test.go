// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"fmt"
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
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	inv_networkv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	inv_statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
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

//nolint:funlen // Test functions are long but necessary to test all the cases.
func TestHost_Summary_Comprehensive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	// Setup base resources
	region := inv_testing.CreateRegion(t, nil)
	site1 := inv_testing.CreateSiteWithArgs(t, "SiteA", 100, 100, "", region, nil, nil)
	site2 := inv_testing.CreateSiteWithArgs(t, "SiteB", 200, 200, "", region, nil, nil)
	os := inv_testing.CreateOs(t)

	// Create a server for testing
	server := inv_server.InventorygRPCServer{InvClient: inv_testing.TestClients[inv_testing.APIClient]}

	// SCENARIO 1: Base hosts with different states
	// Host 1: Normal host with instance, running state
	host1 := createTestHost(ctx, t, "Host1-Running", site1, inv_computev1.HostState_HOST_STATE_ONBOARDED)
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host1.GetResourceId()) })
	instance1 := inv_testing.CreateInstance(t, host1, os)

	// Update Instance status to Running
	updateInstanceState(ctx, t, instance1.GetResourceId(), inv_computev1.InstanceState_INSTANCE_STATE_RUNNING)

	// Host 2: Host with no site allocation
	host2 := createTestHost(ctx, t, "Host2-Unallocated", nil, inv_computev1.HostState_HOST_STATE_REGISTERED)
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host2.GetResourceId()) })

	// Host 3: Host with error status
	host3 := createTestHost(ctx, t, "Host3-Error", site2, inv_computev1.HostState_HOST_STATE_ONBOARDED)
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host3.GetResourceId()) })
	updateHostStatus(ctx, t, host3.GetResourceId(), inv_statusv1.StatusIndication_STATUS_INDICATION_ERROR)

	// Host 4: Host with instance but instance has error
	host4 := createTestHost(ctx, t, "Host4-InstanceError", site1, inv_computev1.HostState_HOST_STATE_ONBOARDED)
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host4.GetResourceId()) })
	instance4 := inv_testing.CreateInstance(t, host4, os)
	updateInstanceStatusIndicator(ctx, t, instance4.GetResourceId(), inv_statusv1.StatusIndication_STATUS_INDICATION_ERROR)

	// VERIFICATION 1: All hosts, no filter
	response, err := server.GetHostsSummary(ctx, &restv1.GetHostSummaryRequest{})
	assert.NoError(t, err)
	assert.Equal(t, uint32(4), response.Total, "Should have 4 hosts in total")
	assert.Equal(t, uint32(1), response.Unallocated, "Should have 1 unallocated host")
	assert.Equal(t, uint32(2), response.Error, "Should have 2 hosts in error state")
	assert.Equal(t, uint32(1), response.Running, "Should have 1 host running")

	// VERIFICATION 2: Test with filter for specific site
	siteFilter := fmt.Sprintf("site.resource_id = %q", site1.GetResourceId())
	response, err = server.GetHostsSummary(ctx, &restv1.GetHostSummaryRequest{Filter: siteFilter})
	assert.NoError(t, err)
	assert.Equal(t, uint32(2), response.Total, "Should have 2 hosts in site1")
	assert.Equal(t, uint32(0), response.Unallocated, "Should have 0 unallocated hosts in site1")
	assert.Equal(t, uint32(1), response.Error, "Should have 1 host in error state in site1")
	assert.Equal(t, uint32(1), response.Running, "Should have 1 host running in site1")

	// SCENARIO 3: Test onboarding status error
	host5 := createTestHost(ctx, t, "Host5-OnboardingError", site2, inv_computev1.HostState_HOST_STATE_REGISTERED)
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host5.GetResourceId()) })
	updateHostOnboardingStatus(ctx, t, host5.GetResourceId(), inv_statusv1.StatusIndication_STATUS_INDICATION_ERROR)

	// VERIFICATION 3: Verify onboarding error is counted
	response, err = server.GetHostsSummary(ctx, &restv1.GetHostSummaryRequest{})
	assert.NoError(t, err)
	assert.Equal(t, uint32(5), response.Total, "Should have 5 hosts in total")
	assert.Equal(t, uint32(1), response.Unallocated, "Should have 1 unallocated host")
	assert.Equal(t, uint32(3), response.Error, "Should have 3 hosts in error state")

	// SCENARIO 4: Test registration status error
	host6 := createTestHost(ctx, t, "Host6-RegistrationError", site1, inv_computev1.HostState_HOST_STATE_REGISTERED)
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host6.GetResourceId()) })
	updateHostRegistrationStatus(ctx, t, host6.GetResourceId(), inv_statusv1.StatusIndication_STATUS_INDICATION_ERROR)

	// VERIFICATION 4: Verify registration error is counted
	response, err = server.GetHostsSummary(ctx, &restv1.GetHostSummaryRequest{})
	assert.NoError(t, err)
	assert.Equal(t, uint32(6), response.Total, "Should have 6 hosts in total")
	assert.Equal(t, uint32(1), response.Unallocated, "Should have 1 unallocated host")
	assert.Equal(t, uint32(4), response.Error, "Should have 4 hosts in error state")

	// SCENARIO 5: Test instance provisioning status error
	host7 := createTestHost(ctx, t, "Host7-ProvisioningError", site2, inv_computev1.HostState_HOST_STATE_ONBOARDED)
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host7.GetResourceId()) })
	instance7 := inv_testing.CreateInstance(t, host7, os)
	updateInstanceProvisioningStatusIndicator(
		ctx,
		t,
		instance7.GetResourceId(),
		inv_statusv1.StatusIndication_STATUS_INDICATION_ERROR,
	)

	// VERIFICATION 5: Verify provisioning error is counted
	response, err = server.GetHostsSummary(ctx, &restv1.GetHostSummaryRequest{})
	assert.NoError(t, err)
	assert.Equal(t, uint32(7), response.Total, "Should have 7 hosts in total")
	assert.Equal(t, uint32(1), response.Unallocated, "Should have 1 unallocated host")
	assert.Equal(t, uint32(5), response.Error, "Should have 5 hosts in error state")

	// SCENARIO 6: Test instance update status error
	host8 := createTestHost(ctx, t, "Host8-UpdateError", site1, inv_computev1.HostState_HOST_STATE_ONBOARDED)
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host8.GetResourceId()) })
	instance8 := inv_testing.CreateInstance(t, host8, os)
	updateInstanceUpdateStatusIndicator(ctx, t, instance8.GetResourceId(), inv_statusv1.StatusIndication_STATUS_INDICATION_ERROR)

	// VERIFICATION 6: Verify update error is counted
	response, err = server.GetHostsSummary(ctx, &restv1.GetHostSummaryRequest{})
	assert.NoError(t, err)
	assert.Equal(t, uint32(8), response.Total, "Should have 8 hosts in total")
	assert.Equal(t, uint32(1), response.Unallocated, "Should have 1 unallocated host")
	assert.Equal(t, uint32(6), response.Error, "Should have 6 hosts in error state")

	// SCENARIO 7: Test instance attestation status error
	host9 := createTestHost(ctx, t, "Host9-AttestationError", site2, inv_computev1.HostState_HOST_STATE_ONBOARDED)
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host9.GetResourceId()) })
	instance9 := inv_testing.CreateInstance(t, host9, os)
	updateInstanceAttestationStatusIndicator(
		ctx,
		t,
		instance9.GetResourceId(),
		inv_statusv1.StatusIndication_STATUS_INDICATION_ERROR,
	)

	// VERIFICATION 7: Verify attestation error is counted
	response, err = server.GetHostsSummary(ctx, &restv1.GetHostSummaryRequest{})
	assert.NoError(t, err)
	assert.Equal(t, uint32(9), response.Total, "Should have 9 hosts in total")
	assert.Equal(t, uint32(1), response.Unallocated, "Should have 1 unallocated host")
	assert.Equal(t, uint32(7), response.Error, "Should have 7 hosts in error state")

	// SCENARIO 8: Test multiple running instances
	host10 := createTestHost(ctx, t, "Host10-Running", site1, inv_computev1.HostState_HOST_STATE_ONBOARDED)
	t.Cleanup(func() { inv_testing.HardDeleteHost(t, host10.GetResourceId()) })
	instance10 := inv_testing.CreateInstance(t, host10, os)
	updateInstanceState(ctx, t, instance10.GetResourceId(), inv_computev1.InstanceState_INSTANCE_STATE_RUNNING)

	// VERIFICATION 8: Verify running count increases
	response, err = server.GetHostsSummary(ctx, &restv1.GetHostSummaryRequest{})
	assert.NoError(t, err)
	assert.Equal(t, uint32(10), response.Total, "Should have 10 hosts in total")
	assert.Equal(t, uint32(1), response.Unallocated, "Should have 1 unallocated host")
	assert.Equal(t, uint32(7), response.Error, "Should have 7 hosts in error state")
	assert.Equal(t, uint32(2), response.Running, "Should have 2 hosts running")
}

// Helper functions for test.
func createTestHost(
	ctx context.Context,
	t *testing.T,
	name string,
	site *inv_locationv1.SiteResource,
	state inv_computev1.HostState,
) *inv_computev1.HostResource {
	t.Helper()

	host := &inv_computev1.HostResource{
		Name:         name,
		DesiredState: state,
		SerialNumber: uuid.NewString(),
		Uuid:         uuid.NewString(),
		MemoryBytes:  64 * 1024 * 1024 * 1024, // 64GB
		CpuModel:     "Test CPU Model",
		CpuSockets:   1,
		CpuCores:     8,
	}

	if site != nil {
		host.Site = site
	}

	createReq := &inventory.Resource{
		Resource: &inventory.Resource_Host{
			Host: host,
		},
	}

	response, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, response)

	return response.GetHost()
}

func updateHostStatus(ctx context.Context, t *testing.T, hostID string, status inv_statusv1.StatusIndication) {
	t.Helper()

	updateHost := &inv_computev1.HostResource{
		HostStatusIndicator: status,
	}

	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{inv_computev1.HostResourceFieldHostStatusIndicator},
	}

	updateRes := &inventory.Resource{
		Resource: &inventory.Resource_Host{Host: updateHost},
	}

	_, err := inv_testing.TestClients[inv_testing.APIClient].Update(ctx, hostID, fieldMask, updateRes)
	require.NoError(t, err)
}

func updateHostOnboardingStatus(ctx context.Context, t *testing.T, hostID string, status inv_statusv1.StatusIndication) {
	t.Helper()

	updateHost := &inv_computev1.HostResource{
		OnboardingStatusIndicator: status,
	}

	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{inv_computev1.HostResourceFieldOnboardingStatusIndicator},
	}

	updateRes := &inventory.Resource{
		Resource: &inventory.Resource_Host{Host: updateHost},
	}

	_, err := inv_testing.TestClients[inv_testing.APIClient].Update(ctx, hostID, fieldMask, updateRes)
	require.NoError(t, err)
}

func updateHostRegistrationStatus(ctx context.Context, t *testing.T, hostID string, status inv_statusv1.StatusIndication) {
	t.Helper()

	updateHost := &inv_computev1.HostResource{
		RegistrationStatusIndicator: status,
	}

	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{inv_computev1.HostResourceFieldRegistrationStatusIndicator},
	}

	updateRes := &inventory.Resource{
		Resource: &inventory.Resource_Host{Host: updateHost},
	}

	_, err := inv_testing.TestClients[inv_testing.APIClient].Update(ctx, hostID, fieldMask, updateRes)
	require.NoError(t, err)
}

func updateInstanceState(ctx context.Context, t *testing.T, instanceID string, state inv_computev1.InstanceState) {
	t.Helper()

	updateInstance := &inv_computev1.InstanceResource{
		CurrentState: state,
	}

	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{inv_computev1.InstanceResourceFieldCurrentState},
	}

	updateRes := &inventory.Resource{
		Resource: &inventory.Resource_Instance{Instance: updateInstance},
	}

	_, err := inv_testing.TestClients[inv_testing.RMClient].Update(ctx, instanceID, fieldMask, updateRes)
	require.NoError(t, err)
}

func updateInstanceStatusIndicator(ctx context.Context, t *testing.T, instanceID string, status inv_statusv1.StatusIndication) {
	t.Helper()

	updateInstance := &inv_computev1.InstanceResource{
		InstanceStatusIndicator: status,
	}

	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{inv_computev1.InstanceResourceFieldInstanceStatusIndicator},
	}

	updateRes := &inventory.Resource{
		Resource: &inventory.Resource_Instance{Instance: updateInstance},
	}

	_, err := inv_testing.TestClients[inv_testing.RMClient].Update(ctx, instanceID, fieldMask, updateRes)
	require.NoError(t, err)
}

func updateInstanceProvisioningStatusIndicator(
	ctx context.Context,
	t *testing.T,
	instanceID string,
	status inv_statusv1.StatusIndication,
) {
	t.Helper()

	updateInstance := &inv_computev1.InstanceResource{
		ProvisioningStatusIndicator: status,
	}

	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{inv_computev1.InstanceResourceFieldProvisioningStatusIndicator},
	}

	updateRes := &inventory.Resource{
		Resource: &inventory.Resource_Instance{Instance: updateInstance},
	}

	_, err := inv_testing.TestClients[inv_testing.RMClient].Update(ctx, instanceID, fieldMask, updateRes)
	require.NoError(t, err)
}

func updateInstanceUpdateStatusIndicator(
	ctx context.Context,
	t *testing.T,
	instanceID string,
	status inv_statusv1.StatusIndication,
) {
	t.Helper()

	updateInstance := &inv_computev1.InstanceResource{
		UpdateStatusIndicator: status,
	}

	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{inv_computev1.InstanceResourceFieldUpdateStatusIndicator},
	}

	updateRes := &inventory.Resource{
		Resource: &inventory.Resource_Instance{Instance: updateInstance},
	}

	_, err := inv_testing.TestClients[inv_testing.RMClient].Update(ctx, instanceID, fieldMask, updateRes)
	require.NoError(t, err)
}

func updateInstanceAttestationStatusIndicator(
	ctx context.Context,
	t *testing.T,
	instanceID string,
	status inv_statusv1.StatusIndication,
) {
	t.Helper()

	updateInstance := &inv_computev1.InstanceResource{
		TrustedAttestationStatusIndicator: status,
	}

	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{inv_computev1.InstanceResourceFieldTrustedAttestationStatusIndicator},
	}

	updateRes := &inventory.Resource{
		Resource: &inventory.Resource_Instance{Instance: updateInstance},
	}

	_, err := inv_testing.TestClients[inv_testing.RMClient].Update(ctx, instanceID, fieldMask, updateRes)
	require.NoError(t, err)
}
