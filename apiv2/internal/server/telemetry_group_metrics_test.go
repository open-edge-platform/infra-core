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

// Example resources for testing
var (
	// Example telemetry metrics group resource from the API
	exampleAPITelemetryMetricsGroupResource = &telemetryv1.TelemetryMetricsGroupResource{
		ResourceId:              "telemetry-metrics-12345678",
		TelemetryMetricsGroupId: "telemetry-metrics-12345678", // Alias of ResourceId
		Name:                    "example-metrics-group",
		CollectorKind:           telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
		Groups:                  []string{"cpu", "memory", "disk", "network"},
	}

	// Example telemetry group resource from the Inventory
	// Note: in inventory it's TelemetryGroupResource with Kind=TELEMETRY_RESOURCE_KIND_METRICS
	exampleInvTelemetryMetricsGroupResource = &inv_telemetryv1.TelemetryGroupResource{
		ResourceId:    "telemetry-metrics-12345678",
		Name:          "example-metrics-group",
		Groups:        []string{"cpu", "memory", "disk", "network"},
		Kind:          inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
		CollectorKind: inv_telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
		TenantId:      "tenant-987654",
		CreatedAt:     "2025-04-22T10:00:00Z",
		UpdatedAt:     "2025-04-22T10:30:00Z",
	}
)

func TestTelemetryMetricsGroup_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateTelemetryMetricsGroupRequest
		wantErr bool
	}{
		{
			name: "Create TelemetryMetricsGroup",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_TelemetryGroup{
								TelemetryGroup: &inv_telemetryv1.TelemetryGroupResource{
									ResourceId:    "telemetry-metrics-12345678",
									Name:          "example-metrics-group",
									Kind:          inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
									CollectorKind: inv_telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
									Groups:        []string{"cpu", "memory"},
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateTelemetryMetricsGroupRequest{
				TelemetryMetricsGroup: &telemetryv1.TelemetryMetricsGroupResource{
					Name:          "example-metrics-group",
					CollectorKind: telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
					Groups:        []string{"cpu", "memory"},
				},
			},
			wantErr: false,
		},
		{
			name: "Create TelemetryMetricsGroup with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateTelemetryMetricsGroupRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateTelemetryMetricsGroup(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateTelemetryMetricsGroup() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateTelemetryMetricsGroup() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateTelemetryMetricsGroup() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetTelemetryMetricsGroup(), reply)
		})
	}
}

func TestTelemetryMetricsGroup_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetTelemetryMetricsGroupRequest
		wantErr bool
	}{
		{
			name: "Get TelemetryMetricsGroup",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetry-metrics-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_TelemetryGroup{
									TelemetryGroup: exampleInvTelemetryMetricsGroupResource,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryMetricsGroupRequest{
				ResourceId: "telemetry-metrics-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get TelemetryMetricsGroup with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetry-metrics-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryMetricsGroupRequest{
				ResourceId: "telemetry-metrics-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetTelemetryMetricsGroup(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetTelemetryMetricsGroup() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetTelemetryMetricsGroup() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetTelemetryMetricsGroup() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPITelemetryMetricsGroupResource, reply)
		})
	}
}

func TestTelemetryMetricsGroup_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListTelemetryMetricsGroupsRequest
		wantErr bool
	}{
		{
			name: "List TelemetryMetricsGroups",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_TelemetryGroup{
											TelemetryGroup: exampleInvTelemetryMetricsGroupResource,
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
			req: &restv1.ListTelemetryMetricsGroupsRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List TelemetryMetricsGroups with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListTelemetryMetricsGroupsRequest{
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

			reply, err := server.ListTelemetryMetricsGroups(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListTelemetryMetricsGroups() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListTelemetryMetricsGroups() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListTelemetryMetricsGroups() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetTelemetryMetricsGroups()) != 1 {
				t.Errorf("ListTelemetryMetricsGroups() got %v telemetry metrics groups, want 1",
					len(reply.GetTelemetryMetricsGroups()))
			}
			compareProtoMessages(t, exampleAPITelemetryMetricsGroupResource, reply.GetTelemetryMetricsGroups()[0])
		})
	}
}

func TestTelemetryMetricsGroup_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteTelemetryMetricsGroupRequest
		wantErr bool
	}{
		{
			name: "Delete TelemetryMetricsGroup",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "telemetry-metrics-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteTelemetryMetricsGroupRequest{
				ResourceId: "telemetry-metrics-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete TelemetryMetricsGroup with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "telemetry-metrics-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteTelemetryMetricsGroupRequest{
				ResourceId: "telemetry-metrics-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteTelemetryMetricsGroup(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteTelemetryMetricsGroup() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteTelemetryMetricsGroup() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteTelemetryMetricsGroup() got reply = nil, want non-nil")
				return
			}
		})
	}
}
