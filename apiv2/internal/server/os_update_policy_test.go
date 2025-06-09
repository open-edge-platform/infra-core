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
	invserver "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
)

// Example resources for testing.
var (
	exampleAPIOSUpdatePolicyResourceLatest = &computev1.OSUpdatePolicy{
		ResourceId:   "osupdatepolicy-12345678",
		Name:         "example-policy",
		Description:  "An example OS update policy",
		UpdatePolicy: computev1.UpdatePolicy_UPDATE_POLICY_LATEST,
	}
	exampleInvOSUpdatePolicyResourceLatest = &inv_computev1.OSUpdatePolicyResource{
		ResourceId:   "osupdatepolicy-12345678",
		Name:         "example-policy",
		Description:  "An example OS update policy",
		UpdatePolicy: inv_computev1.UpdatePolicy_UPDATE_POLICY_LATEST,
	}
)

func TestCreateOSUpdatePolicy(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateOSUpdatePolicyRequest
		wantErr bool
	}{
		{
			name: "Create OSUpdatePolicy",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_OsUpdatePolicy{
								OsUpdatePolicy: &inv_computev1.OSUpdatePolicyResource{
									ResourceId: "osupdatepolicy-12345678",
									Name:       "example-policy",
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateOSUpdatePolicyRequest{
				OsUpdatePolicy: &computev1.OSUpdatePolicy{
					Name: "example-policy",
				},
			},
			wantErr: false,
		},
		{
			name: "Create OSPolicyUpdate with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateOSUpdatePolicyRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateOSUpdatePolicy(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateOSUpdatePolicy() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateOSUpdatePolicy() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateOSUpdatePolicy() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetOsUpdatePolicy(), reply)
		})
	}
}

func TestListOSUpdatePolicy(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListOSUpdatePolicyRequest
		wantErr bool
	}{
		{
			name: "List OSUpdatePolicy",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_OsUpdatePolicy{
											OsUpdatePolicy: exampleInvOSUpdatePolicyResourceLatest,
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
			req: &restv1.ListOSUpdatePolicyRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List OSUpdatePolicy with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListOSUpdatePolicyRequest{
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

			reply, err := server.ListOSUpdatePolicy(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListOSUpdatePolicy() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListOSUpdatePolicy() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListOSUpdatePolicy() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetOsUpdatePolicies()) != 1 {
				t.Errorf("ListOSUpdatePolicy() got %v OSUpdatePolicies, want 1", len(reply.GetOsUpdatePolicies()))
			}
			compareProtoMessages(t, exampleAPIOSUpdatePolicyResourceLatest, reply.GetOsUpdatePolicies()[0])
		})
	}
}

func TestGetOSUpdatePolicy(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetOSUpdatePolicyRequest
		wantErr bool
	}{
		{
			name: "Get OSUpdatePolicy",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "osupdatepolicy-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_OsUpdatePolicy{
									OsUpdatePolicy: exampleInvOSUpdatePolicyResourceLatest,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetOSUpdatePolicyRequest{
				ResourceId: "osupdatepolicy-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get OSUpdatePolicy with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "osupdatepolicy-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetOSUpdatePolicyRequest{
				ResourceId: "osupdatepolicy-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetOSUpdatePolicy(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetOSUpdatePolicy() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetOSUpdatePolicy() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetOSUpdatePolicy() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIOSUpdatePolicyResourceLatest, reply)
		})
	}
}

func TestDeleteOSUpdatePolicy(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteOSUpdatePolicyRequest
		wantErr bool
	}{
		{
			name: "Delete OSUpdatePolicy",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "osupdatepolicy-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteOSUpdatePolicyRequest{
				ResourceId: "osupdatepolicy-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete OSUpdatePolicy with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "osupdatepolicy-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteOSUpdatePolicyRequest{
				ResourceId: "osupdatepolicy-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteOSUpdatePolicy(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteOSUpdatePolicy() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteOSUpdatePolicy() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteOSUpdatePolicy() got reply = nil, want non-nil")
				return
			}
		})
	}
}
