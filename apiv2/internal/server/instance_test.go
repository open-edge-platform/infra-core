// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	osv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/os/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_server "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	inv_statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
)

// Example resources for testing.
var (
	// Example instance resource from the API.
	exampleAPIInstanceResource = &computev1.InstanceResource{
		ResourceId:      "inst-12345678",
		Kind:            computev1.InstanceKind_INSTANCE_KIND_METAL,
		Name:            "example-instance",
		DesiredState:    computev1.InstanceState_INSTANCE_STATE_RUNNING,
		CurrentState:    computev1.InstanceState_INSTANCE_STATE_RUNNING,
		HostID:          "host-87654321",
		SecurityFeature: osv1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION,
		ExistingCves: `[
{
  "cve_id": "CVE-000-000",
  "priority": "critical",
  "affected_packages": [
    "test-package-0.0.0",
    "test-2\test3"
  ]
}]`,
		// Optional fields
		OsID:       "os-12345678",
		InstanceID: "inst-12345678", // Alias of ResourceId
	}

	// Example instance resource from the Inventory.
	exampleInvInstanceResource = &inv_computev1.InstanceResource{
		ResourceId:      "inst-12345678",
		Kind:            inv_computev1.InstanceKind_INSTANCE_KIND_METAL,
		Name:            "example-instance",
		SecurityFeature: inv_osv1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION,
		DesiredState:    inv_computev1.InstanceState_INSTANCE_STATE_RUNNING,
		CurrentState:    inv_computev1.InstanceState_INSTANCE_STATE_RUNNING,
		VmMemoryBytes:   8589934592, // 8GB in bytes
		VmCpuCores:      4,
		VmStorageBytes:  274877906944, // 256GB in bytes

		InstanceStatus:          "Running normally",
		InstanceStatusIndicator: inv_statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		InstanceStatusTimestamp: 1713868800000, // Example timestamp

		ProvisioningStatus:          "Successfully provisioned",
		ProvisioningStatusIndicator: inv_statusv1.StatusIndication_STATUS_INDICATION_IDLE,
		ProvisioningStatusTimestamp: 1713868800000,
		Host: &inv_computev1.HostResource{
			ResourceId: "host-87654321",
		},
		CurrentOs: &inv_osv1.OperatingSystemResource{
			ResourceId: "os-12345678",
		},
		Os: &inv_osv1.OperatingSystemResource{
			ResourceId: "os-12345678",
		},
		ExistingCves: `[
{
  "cve_id": "CVE-000-000",
  "priority": "critical",
  "affected_packages": [
    "test-package-0.0.0",
    "test-2\test3"
  ]
}]`,
		TenantId:  "tenant-987654",
		CreatedAt: "2025-04-22T10:00:00Z",
		UpdatedAt: "2025-04-22T10:30:00Z",
	}
)

func TestInstance_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateInstanceRequest
		wantErr bool
	}{
		{
			name: "Create Instance",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Instance{
								Instance: &inv_computev1.InstanceResource{
									ResourceId: "instance-12345678",
									Name:       "example-instance",
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateInstanceRequest{
				Instance: &computev1.InstanceResource{
					Name: "example-instance",
				},
			},
			wantErr: false,
		},
		{
			name: "Create Instance with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateInstanceRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateInstance(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateInstance() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateInstance() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateInstance() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetInstance(), reply)
		})
	}
}

func TestInstance_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetInstanceRequest
		wantErr bool
	}{
		{
			name: "Get Instance",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "instance-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_Instance{
									Instance: exampleInvInstanceResource,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetInstanceRequest{
				ResourceId: "instance-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get Instance with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "instance-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetInstanceRequest{
				ResourceId: "instance-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetInstance(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetInstance() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetInstance() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetInstance() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIInstanceResource, reply)
		})
	}
}

func TestInstance_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListInstancesRequest
		wantErr bool
	}{
		{
			name: "List Instances",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_Instance{
											Instance: exampleInvInstanceResource,
										},
									},
								},
							},
							TotalElements: 1,
							HasNext:       false,
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListInstancesRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List Instances with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListInstancesRequest{
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

			reply, err := server.ListInstances(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListInstances() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListInstances() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListInstances() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetInstances()) != 1 {
				t.Errorf("ListInstances() got %v instances, want 1", len(reply.GetInstances()))
			}
			compareProtoMessages(t, exampleAPIInstanceResource, reply.GetInstances()[0])
		})
	}
}

func TestInstance_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateInstanceRequest
		wantErr bool
	}{
		{
			name: "Update Instance",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "instance-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Instance{
								Instance: exampleInvInstanceResource,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateInstanceRequest{
				ResourceId: "instance-12345678",
				Instance:   exampleAPIInstanceResource,
			},
			wantErr: false,
		},
		{
			name: "Update Instance with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "instance-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateInstanceRequest{
				ResourceId: "instance-12345678",
				Instance:   exampleAPIInstanceResource,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateInstance(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateInstance() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateInstance() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateInstance() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIInstanceResource, reply)
		})
	}
}

func TestInstance_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteInstanceRequest
		wantErr bool
	}{
		{
			name: "Delete Instance",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "instance-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteInstanceRequest{
				ResourceId: "instance-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete Instance with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "instance-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteInstanceRequest{
				ResourceId: "instance-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteInstance(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteInstance() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteInstance() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteInstance() got reply = nil, want non-nil")
				return
			}
		})
	}
}
