// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	telemetryv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/telemetry/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_server "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_telemetryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
)

// Example resources for testing.
var (
	// Example telemetry logs group resource from the API.
	exampleAPITelemetryLogsGroupResource = &telemetryv1.TelemetryLogsGroupResource{
		ResourceId:           "telemetry-logs-12345678",
		TelemetryLogsGroupId: "telemetry-logs-12345678", // Alias of ResourceId
		Name:                 "example-logs-group",
		CollectorKind:        telemetryv1.TelemetryCollectorKind_TELEMETRY_COLLECTOR_KIND_CLUSTER,
		Groups:               []string{"system", "application", "security"},
	}

	// Example telemetry group resource from the Inventory.
	// Note: in inventory it's TelemetryGroupResource with Kind=TELEMETRY_RESOURCE_KIND_LOGS.
	exampleInvTelemetryLogsGroupResource = &inv_telemetryv1.TelemetryGroupResource{
		ResourceId:    "telemetry-logs-12345678",
		Name:          "example-logs-group",
		Groups:        []string{"system", "application", "security"},
		Kind:          inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
		CollectorKind: inv_telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
		TenantId:      "tenant-987654",
		CreatedAt:     "2025-04-22T10:00:00Z",
		UpdatedAt:     "2025-04-22T10:30:00Z",
	}
)

func TestTelemetryLogsGroup_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateTelemetryLogsGroupRequest
		wantErr bool
	}{
		{
			name: "Create TelemetryLogsGroup",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_TelemetryGroup{
								TelemetryGroup: &inv_telemetryv1.TelemetryGroupResource{
									ResourceId:    "telemetry-logs-12345678",
									Name:          "example-logs-group",
									Kind:          inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
									CollectorKind: inv_telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
									Groups:        []string{"system", "application"},
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateTelemetryLogsGroupRequest{
				TelemetryLogsGroup: &telemetryv1.TelemetryLogsGroupResource{
					Name:          "example-logs-group",
					CollectorKind: telemetryv1.TelemetryCollectorKind_TELEMETRY_COLLECTOR_KIND_CLUSTER,
					Groups:        []string{"system", "application"},
				},
			},
			wantErr: false,
		},
		{
			name: "Create TelemetryLogsGroup with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateTelemetryLogsGroupRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateTelemetryLogsGroup(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateTelemetryLogsGroup() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateTelemetryLogsGroup() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateTelemetryLogsGroup() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetTelemetryLogsGroup(), reply)
		})
	}
}

func TestTelemetryLogsGroup_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetTelemetryLogsGroupRequest
		wantErr bool
	}{
		{
			name: "Get TelemetryLogsGroup",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetry-logs-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_TelemetryGroup{
									TelemetryGroup: exampleInvTelemetryLogsGroupResource,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryLogsGroupRequest{
				ResourceId: "telemetry-logs-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get TelemetryLogsGroup with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetry-logs-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryLogsGroupRequest{
				ResourceId: "telemetry-logs-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetTelemetryLogsGroup(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetTelemetryLogsGroup() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetTelemetryLogsGroup() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetTelemetryLogsGroup() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPITelemetryLogsGroupResource, reply)
		})
	}
}

func TestTelemetryLogsGroup_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListTelemetryLogsGroupsRequest
		wantErr bool
	}{
		{
			name: "List TelemetryLogsGroups",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_TelemetryGroup{
											TelemetryGroup: exampleInvTelemetryLogsGroupResource,
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
			req: &restv1.ListTelemetryLogsGroupsRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List TelemetryLogsGroups with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListTelemetryLogsGroupsRequest{
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

			reply, err := server.ListTelemetryLogsGroups(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListTelemetryLogsGroups() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListTelemetryLogsGroups() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListTelemetryLogsGroups() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetTelemetryLogsGroups()) != 1 {
				t.Errorf("ListTelemetryLogsGroups() got %v telemetry logs groups, want 1",
					len(reply.GetTelemetryLogsGroups()))
			}
			compareProtoMessages(t, exampleAPITelemetryLogsGroupResource, reply.GetTelemetryLogsGroups()[0])
		})
	}
}

func TestTelemetryLogsGroup_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteTelemetryLogsGroupRequest
		wantErr bool
	}{
		{
			name: "Delete TelemetryLogsGroup",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "telemetry-logs-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteTelemetryLogsGroupRequest{
				ResourceId: "telemetry-logs-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete TelemetryLogsGroup with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "telemetry-logs-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteTelemetryLogsGroupRequest{
				ResourceId: "telemetry-logs-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteTelemetryLogsGroup(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteTelemetryLogsGroup() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteTelemetryLogsGroup() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteTelemetryLogsGroup() got reply = nil, want non-nil")
				return
			}
		})
	}
}
