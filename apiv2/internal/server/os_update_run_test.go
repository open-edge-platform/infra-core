// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	computev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/compute/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	invserver "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
)

// Example resources for testing.
var (
	exampleAPIOSUpdateRunResource = &computev1.OSUpdateRun{
		ResourceId:  "osupdaterun-12345678",
		Name:        "example-run",
		Description: "An example OS update run",
	}
	exampleInvOSUpdateRunResource = &inv_computev1.OSUpdateRunResource{
		ResourceId:  "osupdaterun-12345678",
		Name:        "example-run",
		Description: "An example OS update run",
		StartTime:   uint64(time.Now().Unix()),
	}
)

func TestListOSUpdateRun(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListOSUpdateRunRequest
		wantErr bool
	}{
		{
			name: "List OSUpdateRun",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_OsUpdateRun{
											OsUpdateRun: exampleInvOSUpdateRunResource,
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
			req: &restv1.ListOSUpdateRunRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List OSUpdateRun with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListOSUpdateRunRequest{
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

			reply, err := server.ListOSUpdateRun(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListOSUpdateRun() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListOSUpdateRun() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListOSUpdateRun() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetOsUpdateRuns()) != 1 {
				t.Errorf("ListOSUpdateRun() got %v OSUpdatePolicies, want 1", len(reply.GetOsUpdateRuns()))
			}
			compareProtoMessages(t, exampleAPIOSUpdateRunResource, reply.GetOsUpdateRuns()[0])
		})
	}
}

func TestGetOSUpdateRun(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetOSUpdateRunRequest
		wantErr bool
	}{
		{
			name: "Get OSUpdateRun",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "osupdaterun-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_OsUpdateRun{
									OsUpdateRun: exampleInvOSUpdateRunResource,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetOSUpdateRunRequest{
				ResourceId: "osupdaterun-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get OSUpdateRun with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "osupdaterun-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetOSUpdateRunRequest{
				ResourceId: "osupdaterun-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetOSUpdateRun(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetOSUpdateRun() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetOSUpdateRun() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetOSUpdateRun() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIOSUpdateRunResource, reply)
		})
	}
}

func TestDeleteOSUpdateRun(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := invserver.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteOSUpdateRunRequest
		wantErr bool
	}{
		{
			name: "Delete OSUpdateRun",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "osupdaterun-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteOSUpdateRunRequest{
				ResourceId: "osupdaterun-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete OSUpdateRun with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "osupdaterun-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteOSUpdateRunRequest{
				ResourceId: "osupdaterun-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteOSUpdateRun(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteOSUpdateRun() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteOSUpdateRun() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteOSUpdateRun() got reply = nil, want non-nil")
				return
			}
		})
	}
}
