// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_server "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
)

// Example resources for testing.
var (
	// Example workload member resources.
	exampleAPIWorkloadMemberInternal = &computev1.WorkloadMember{
		ResourceId:       "workload-member-12345678",
		WorkloadMemberId: "workload-member-12345678", // Alias of ResourceId
		InstanceId:       "instance-12345678",
		Instance: &computev1.InstanceResource{
			ResourceId: "instance-12345678",
		},
	}

	exampleInvWorkloadMemberInternal = &inv_computev1.WorkloadMember{
		ResourceId: "workload-member-12345678",
		Instance: &inv_computev1.InstanceResource{
			ResourceId: "instance-12345678",
		},
	}

	// Example workload resources.
	exampleAPIWorkload = &computev1.WorkloadResource{
		ResourceId: "workload-12345678",
		WorkloadId: "workload-12345678", // Alias of ResourceId
		Kind:       computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
		Name:       "example-workload",
		ExternalId: "external-id-12345678",
		Status:     "running",
		Members:    []*computev1.WorkloadMember{exampleAPIWorkloadMemberInternal},
	}

	exampleInvWorkload = &inv_computev1.WorkloadResource{
		ResourceId:   "workload-12345678",
		Kind:         inv_computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
		Name:         "example-workload",
		ExternalId:   "external-id-12345678",
		Status:       "running",
		DesiredState: inv_computev1.WorkloadState_WORKLOAD_STATE_PROVISIONED,
		Members:      []*inv_computev1.WorkloadMember{exampleInvWorkloadMemberInternal},
		CreatedAt:    "2025-04-22T10:00:00Z",
		UpdatedAt:    "2025-04-22T10:30:00Z",
		TenantId:     "tenant-987654",
	}
)

func TestWorkload_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateWorkloadRequest
		wantErr bool
	}{
		{
			name: "Create Workload",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Workload{
								Workload: &inv_computev1.WorkloadResource{
									ResourceId: "workload-12345678",
									Name:       "example-workload",
									Kind:       inv_computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateWorkloadRequest{
				Workload: &computev1.WorkloadResource{
					Name: "example-workload",
					Kind: computev1.WorkloadKind_WORKLOAD_KIND_CLUSTER,
				},
			},
			wantErr: false,
		},
		{
			name: "Create Workload with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateWorkloadRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateWorkload(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateWorkload() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateWorkload() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateWorkload() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetWorkload(), reply)
		})
	}
}

func TestWorkload_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetWorkloadRequest
		wantErr bool
	}{
		{
			name: "Get Workload",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "workload-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_Workload{
									Workload: exampleInvWorkload,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetWorkloadRequest{
				ResourceId: "workload-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get Workload with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "workload-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetWorkloadRequest{
				ResourceId: "workload-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetWorkload(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetWorkload() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetWorkload() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetWorkload() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIWorkload, reply)
		})
	}
}

func TestWorkload_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListWorkloadsRequest
		wantErr bool
	}{
		{
			name: "List Workloads",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_Workload{
											Workload: exampleInvWorkload,
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
			req: &restv1.ListWorkloadsRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List Workloads with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListWorkloadsRequest{
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

			reply, err := server.ListWorkloads(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListWorkloads() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListWorkloads() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListWorkloads() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetWorkloads()) != 1 {
				t.Errorf("ListWorkloads() got %v workloads, want 1", len(reply.GetWorkloads()))
			}
			compareProtoMessages(t, exampleAPIWorkload, reply.GetWorkloads()[0])
		})
	}
}

func TestWorkload_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateWorkloadRequest
		wantErr bool
	}{
		{
			name: "Update Workload",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "workload-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Workload{
								Workload: exampleInvWorkload,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateWorkloadRequest{
				ResourceId: "workload-12345678",
				Workload:   exampleAPIWorkload,
			},
			wantErr: false,
		},
		{
			name: "Update Workload with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "workload-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateWorkloadRequest{
				ResourceId: "workload-12345678",
				Workload:   exampleAPIWorkload,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateWorkload(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateWorkload() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateWorkload() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateWorkload() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIWorkload, reply)
		})
	}
}

func TestWorkload_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteWorkloadRequest
		wantErr bool
	}{
		{
			name: "Delete Workload",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "workload-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteWorkloadRequest{
				ResourceId: "workload-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete Workload with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "workload-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteWorkloadRequest{
				ResourceId: "workload-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteWorkload(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteWorkload() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteWorkload() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteWorkload() got reply = nil, want non-nil")
				return
			}
		})
	}
}
