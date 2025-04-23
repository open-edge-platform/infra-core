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
	// Example workload member resource from the API.
	exampleAPIWorkloadMember = &computev1.WorkloadMember{
		ResourceId:       "workload-member-12345678",
		WorkloadMemberId: "workload-member-12345678", // Alias of ResourceId
		Kind:             computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
		WorkloadId:       "workload-12345678",
		InstanceId:       "inst-12345678",
		Workload:         exampleAPIWorkload,         // Reference to existing example
		Instance:         exampleAPIInstanceResource, // Reference to existing example
		Member:           exampleAPIInstanceResource, // Reference to instance as member
	}

	// Example workload member resource from the Inventory.
	exampleInvWorkloadMember = &inv_computev1.WorkloadMember{
		ResourceId: "workload-member-12345678",
		Kind:       inv_computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
		Workload:   exampleInvWorkload,
		Instance:   exampleInvInstanceResource,
		TenantId:   "tenant-987654",
		CreatedAt:  "2025-04-22T10:00:00Z",
		UpdatedAt:  "2025-04-22T10:30:00Z",
	}
)

func TestWorkloadMember_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateWorkloadMemberRequest
		wantErr bool
	}{
		{
			name: "Create WorkloadMember",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_WorkloadMember{
								WorkloadMember: &inv_computev1.WorkloadMember{
									ResourceId: "workload-member-12345678",
									Kind:       inv_computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
									Instance: &inv_computev1.InstanceResource{
										ResourceId: "inst-12345678",
									},
									Workload: &inv_computev1.WorkloadResource{
										ResourceId: "workload-12345678",
									},
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateWorkloadMemberRequest{
				WorkloadMember: &computev1.WorkloadMember{
					Kind:       computev1.WorkloadMemberKind_WORKLOAD_MEMBER_KIND_CLUSTER_NODE,
					InstanceId: "inst-12345678",
					WorkloadId: "workload-12345678",
				},
			},
			wantErr: false,
		},
		{
			name: "Create WorkloadMember with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateWorkloadMemberRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateWorkloadMember(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateWorkloadMember() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateWorkloadMember() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateWorkloadMember() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetWorkloadMember(), reply)
		})
	}
}

func TestWorkloadMember_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetWorkloadMemberRequest
		wantErr bool
	}{
		{
			name: "Get WorkloadMember",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "workload-member-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_WorkloadMember{
									WorkloadMember: exampleInvWorkloadMember,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetWorkloadMemberRequest{
				ResourceId: "workload-member-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get WorkloadMember with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "workload-member-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetWorkloadMemberRequest{
				ResourceId: "workload-member-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetWorkloadMember(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetWorkloadMember() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetWorkloadMember() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetWorkloadMember() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIWorkloadMember, reply)
		})
	}
}

func TestWorkloadMember_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListWorkloadMembersRequest
		wantErr bool
	}{
		{
			name: "List WorkloadMembers",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_WorkloadMember{
											WorkloadMember: exampleInvWorkloadMember,
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
			req: &restv1.ListWorkloadMembersRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List WorkloadMembers with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListWorkloadMembersRequest{
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

			reply, err := server.ListWorkloadMembers(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListWorkloadMembers() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListWorkloadMembers() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListWorkloadMembers() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetWorkloadMembers()) != 1 {
				t.Errorf("ListWorkloadMembers() got %v workloadMembers, want 1", len(reply.GetWorkloadMembers()))
			}
			compareProtoMessages(t, exampleAPIWorkloadMember, reply.GetWorkloadMembers()[0])
		})
	}
}

func TestWorkloadMember_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateWorkloadMemberRequest
		wantErr bool
	}{
		{
			name: "Update WorkloadMember",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "workload-member-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_WorkloadMember{
								WorkloadMember: exampleInvWorkloadMember,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateWorkloadMemberRequest{
				ResourceId:     "workload-member-12345678",
				WorkloadMember: exampleAPIWorkloadMember,
			},
			wantErr: false,
		},
		{
			name: "Update WorkloadMember with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "workload-member-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateWorkloadMemberRequest{
				ResourceId:     "workload-member-12345678",
				WorkloadMember: exampleAPIWorkloadMember,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateWorkloadMember(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateWorkloadMember() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateWorkloadMember() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateWorkloadMember() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIWorkloadMember, reply)
		})
	}
}

func TestWorkloadMember_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteWorkloadMemberRequest
		wantErr bool
	}{
		{
			name: "Delete WorkloadMember",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "workload-member-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteWorkloadMemberRequest{
				ResourceId: "workload-member-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete WorkloadMember with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "workload-member-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteWorkloadMemberRequest{
				ResourceId: "workload-member-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteWorkloadMember(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteWorkloadMember() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteWorkloadMember() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteWorkloadMember() got reply = nil, want non-nil")
				return
			}
		})
	}
}
