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
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	inv_telemetryv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
)

// Example resources for testing.
var (
	// Example telemetry metrics group resource from API.
	exampleAPITelemetryMetricsGroup = &telemetryv1.TelemetryMetricsGroupResource{
		ResourceId:              "telemetrygroup-12345678",
		TelemetryMetricsGroupId: "telemetrygroup-12345678",
		Name:                    "example-metrics-group",
		CollectorKind:           telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
		Groups:                  []string{"cpu", "memory", "disk", "network"},
	}

	// Example telemetry group resource from Inventory.
	exampleInvTelemetryMetricsGroup = &inv_telemetryv1.TelemetryGroupResource{
		ResourceId:    "telemetrygroup-12345678",
		Name:          "example-metrics-group",
		Groups:        []string{"cpu", "memory", "disk", "network"},
		Kind:          inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
		CollectorKind: inv_telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
	}

	// Example telemetry metrics profile resource from API with instance target.
	exampleAPIInstanceMetricsProfile = &telemetryv1.TelemetryMetricsProfileResource{
		ResourceId:      "telemetryprofile-12345678",
		ProfileId:       "telemetryprofile-12345678", // Alias of ResourceId
		MetricsInterval: 60,                          // 60 seconds
		TargetInstance:  "inst-12345678",
		MetricsGroupId:  "telemetrygroup-12345678",
		MetricsGroup:    exampleAPITelemetryMetricsGroup,
	}

	// Example telemetry metrics profile resource from API with site target.
	exampleAPISiteMetricsProfile = &telemetryv1.TelemetryMetricsProfileResource{
		ResourceId:      "telemetryprofile-23456789",
		ProfileId:       "telemetryprofile-23456789", // Alias of ResourceId
		MetricsInterval: 300,                         // 5 minutes
		TargetSite:      "site-12345678",
		MetricsGroupId:  "telemetrygroup-12345678",
		MetricsGroup:    exampleAPITelemetryMetricsGroup,
	}

	// Example telemetry metrics profile resource from API with region target.
	exampleAPIRegionMetricsProfile = &telemetryv1.TelemetryMetricsProfileResource{
		ResourceId:      "telemetryprofile-34567890",
		ProfileId:       "telemetryprofile-34567890", // Alias of ResourceId
		MetricsInterval: 600,                         // 10 minutes
		TargetRegion:    "region-12345678",
		MetricsGroupId:  "telemetrygroup-12345678",
		MetricsGroup:    exampleAPITelemetryMetricsGroup,
	}

	// Example telemetry profile resource from Inventory with instance target.
	exampleInvInstanceMetricsProfile = &inv_telemetryv1.TelemetryProfile{
		ResourceId:      "telemetryprofile-12345678",
		Kind:            inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
		MetricsInterval: 60, // 60 seconds
		Group:           exampleInvTelemetryMetricsGroup,
		Relation: &inv_telemetryv1.TelemetryProfile_Instance{
			Instance: &inv_computev1.InstanceResource{
				ResourceId: "inst-12345678",
			},
		},
		TenantId:  "tenant-987654",
		CreatedAt: "2025-04-22T10:00:00Z",
		UpdatedAt: "2025-04-22T10:30:00Z",
	}

	// Example telemetry profile resource from Inventory with site target.
	exampleInvSiteMetricsProfile = &inv_telemetryv1.TelemetryProfile{
		ResourceId:      "telemetryprofile-23456789",
		Kind:            inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
		MetricsInterval: 300, // 5 minutes
		Group:           exampleInvTelemetryMetricsGroup,
		Relation: &inv_telemetryv1.TelemetryProfile_Site{
			Site: &inv_locationv1.SiteResource{
				ResourceId: "site-12345678",
			},
		},
		TenantId:  "tenant-987654",
		CreatedAt: "2025-04-22T11:00:00Z",
		UpdatedAt: "2025-04-22T11:30:00Z",
	}

	// Example telemetry profile resource from Inventory with region target.
	exampleInvRegionMetricsProfile = &inv_telemetryv1.TelemetryProfile{
		ResourceId:      "telemetryprofile-34567890",
		Kind:            inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
		MetricsInterval: 600, // 10 minutes
		Group:           exampleInvTelemetryMetricsGroup,
		Relation: &inv_telemetryv1.TelemetryProfile_Region{
			Region: &inv_locationv1.RegionResource{
				ResourceId: "region-12345678",
			},
		},
		TenantId:  "tenant-987654",
		CreatedAt: "2025-04-22T12:00:00Z",
		UpdatedAt: "2025-04-22T12:30:00Z",
	}
)

//nolint:funlen // Testing function.
func TestTelemetryMetricsProfile_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateTelemetryMetricsProfileRequest
		wantErr bool
	}{
		{
			name: "Create TelemetryMetricsProfile with instance target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_TelemetryProfile{
								TelemetryProfile: &inv_telemetryv1.TelemetryProfile{
									ResourceId:      "telemetryprofile-12345678",
									Kind:            inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
									MetricsInterval: 60,
									Group: &inv_telemetryv1.TelemetryGroupResource{
										ResourceId: "telemetrygroup-12345678",
									},
									Relation: &inv_telemetryv1.TelemetryProfile_Instance{
										Instance: &inv_computev1.InstanceResource{
											ResourceId: "inst-12345678",
										},
									},
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateTelemetryMetricsProfileRequest{
				TelemetryMetricsProfile: &telemetryv1.TelemetryMetricsProfileResource{
					MetricsInterval: 60,
					MetricsGroupId:  "telemetrygroup-12345678",
					TargetInstance:  "inst-12345678",
				},
			},
			wantErr: false,
		},
		{
			name: "Create TelemetryMetricsProfile with site target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_TelemetryProfile{
								TelemetryProfile: &inv_telemetryv1.TelemetryProfile{
									ResourceId:      "telemetryprofile-23456789",
									Kind:            inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
									MetricsInterval: 300,
									Group: &inv_telemetryv1.TelemetryGroupResource{
										ResourceId: "telemetrygroup-12345678",
									},
									Relation: &inv_telemetryv1.TelemetryProfile_Site{
										Site: &inv_locationv1.SiteResource{
											ResourceId: "site-12345678",
										},
									},
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateTelemetryMetricsProfileRequest{
				TelemetryMetricsProfile: &telemetryv1.TelemetryMetricsProfileResource{
					MetricsInterval: 300,
					MetricsGroupId:  "telemetrygroup-12345678",
					TargetSite:      "site-12345678",
				},
			},
			wantErr: false,
		},
		{
			name: "Create TelemetryMetricsProfile with region target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_TelemetryProfile{
								TelemetryProfile: &inv_telemetryv1.TelemetryProfile{
									ResourceId:      "telemetryprofile-23456789",
									Kind:            inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS,
									MetricsInterval: 300,
									Group: &inv_telemetryv1.TelemetryGroupResource{
										ResourceId: "telemetrygroup-12345678",
									},
									Relation: &inv_telemetryv1.TelemetryProfile_Region{
										Region: &inv_locationv1.RegionResource{
											ResourceId: "region-12345678",
										},
									},
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateTelemetryMetricsProfileRequest{
				TelemetryMetricsProfile: &telemetryv1.TelemetryMetricsProfileResource{
					MetricsInterval: 300,
					MetricsGroupId:  "telemetrygroup-12345678",
					TargetRegion:    "region-12345678",
				},
			},
			wantErr: false,
		},
		{
			name: "Create TelemetryMetricsProfile with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateTelemetryMetricsProfileRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateTelemetryMetricsProfile(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateTelemetryMetricsProfile() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateTelemetryMetricsProfile() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateTelemetryMetricsProfile() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetTelemetryMetricsProfile(), reply)
		})
	}
}

//nolint:funlen // Testing function.
func TestTelemetryMetricsProfile_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name       string
		mocks      func() []*mock.Call
		ctx        context.Context
		req        *restv1.GetTelemetryMetricsProfileRequest
		wantErr    bool
		example    *telemetryv1.TelemetryMetricsProfileResource
		invExample *inv_telemetryv1.TelemetryProfile
	}{
		{
			name: "Get TelemetryMetricsProfile with instance target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetryprofile-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_TelemetryProfile{
									TelemetryProfile: exampleInvInstanceMetricsProfile,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryMetricsProfileRequest{
				ResourceId: "telemetryprofile-12345678",
			},
			wantErr:    false,
			example:    exampleAPIInstanceMetricsProfile,
			invExample: exampleInvInstanceMetricsProfile,
		},
		{
			name: "Get TelemetryMetricsProfile with site target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetryprofile-23456789").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_TelemetryProfile{
									TelemetryProfile: exampleInvSiteMetricsProfile,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryMetricsProfileRequest{
				ResourceId: "telemetryprofile-23456789",
			},
			wantErr:    false,
			example:    exampleAPISiteMetricsProfile,
			invExample: exampleInvSiteMetricsProfile,
		},
		{
			name: "Get TelemetryMetricsProfile with region target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetryprofile-23456789").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_TelemetryProfile{
									TelemetryProfile: exampleInvRegionMetricsProfile,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryMetricsProfileRequest{
				ResourceId: "telemetryprofile-23456789",
			},
			wantErr:    false,
			example:    exampleAPIRegionMetricsProfile,
			invExample: exampleInvRegionMetricsProfile,
		},
		{
			name: "Get TelemetryMetricsProfile with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetryprofile-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryMetricsProfileRequest{
				ResourceId: "telemetryprofile-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetTelemetryMetricsProfile(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetTelemetryMetricsProfile() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetTelemetryMetricsProfile() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetTelemetryMetricsProfile() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.example, reply)
		})
	}
}

//nolint:funlen // Testing function.
func TestTelemetryMetricsProfile_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListTelemetryMetricsProfilesRequest
		wantErr bool
	}{
		{
			name: "List TelemetryMetricsProfiles",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_TelemetryProfile{
											TelemetryProfile: exampleInvInstanceMetricsProfile,
										},
									},
								},
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_TelemetryProfile{
											TelemetryProfile: exampleInvSiteMetricsProfile,
										},
									},
								},
							},
							TotalElements: 2,
							HasNext:       false,
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListTelemetryMetricsProfilesRequest{
				PageSize:      10,
				Offset:        0,
				ShowInherited: false,
			},
			wantErr: false,
		},
		{
			name: "List inherited TelemetryMetricsProfiles",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("ListInheritedTelemetryProfiles",
						mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(&inventory.ListInheritedTelemetryProfilesResponse{
							TelemetryProfiles: []*inv_telemetryv1.TelemetryProfile{
								exampleInvRegionMetricsProfile,
								exampleInvSiteMetricsProfile,
							},
							TotalElements: 2,
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListTelemetryMetricsProfilesRequest{
				PageSize:      10,
				Offset:        0,
				ShowInherited: true,
				InstanceId:    "inst-12345678",
			},
			wantErr: false,
		},
		{
			name: "List TelemetryMetricsProfiles with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListTelemetryMetricsProfilesRequest{
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

			reply, err := server.ListTelemetryMetricsProfiles(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListTelemetryMetricsProfiles() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListTelemetryMetricsProfiles() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListTelemetryMetricsProfiles() got reply = nil, want non-nil")
				return
			}
		})
	}
}

func TestTelemetryMetricsProfile_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateTelemetryMetricsProfileRequest
		wantErr bool
	}{
		{
			name: "Update TelemetryMetricsProfile",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "telemetryprofile-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_TelemetryProfile{
								TelemetryProfile: exampleInvInstanceMetricsProfile,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateTelemetryMetricsProfileRequest{
				ResourceId:              "telemetryprofile-12345678",
				TelemetryMetricsProfile: exampleAPIInstanceMetricsProfile,
			},
			wantErr: false,
		},
		{
			name: "Update TelemetryMetricsProfile with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "telemetryprofile-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateTelemetryMetricsProfileRequest{
				ResourceId:              "telemetryprofile-12345678",
				TelemetryMetricsProfile: exampleAPIInstanceMetricsProfile,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateTelemetryMetricsProfile(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateTelemetryMetricsProfile() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateTelemetryMetricsProfile() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateTelemetryMetricsProfile() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIInstanceMetricsProfile, reply)
		})
	}
}

func TestTelemetryMetricsProfile_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteTelemetryMetricsProfileRequest
		wantErr bool
	}{
		{
			name: "Delete TelemetryMetricsProfile",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "telemetryprofile-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteTelemetryMetricsProfileRequest{
				ResourceId: "telemetryprofile-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete TelemetryMetricsProfile with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "telemetryprofile-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteTelemetryMetricsProfileRequest{
				ResourceId: "telemetryprofile-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteTelemetryMetricsProfile(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteTelemetryMetricsProfile() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteTelemetryMetricsProfile() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteTelemetryMetricsProfile() got reply = nil, want non-nil")
				return
			}
		})
	}
}
