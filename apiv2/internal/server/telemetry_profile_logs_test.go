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
	// Example telemetry group resource from API.
	exampleAPITelemetryLogsGroup = &telemetryv1.TelemetryLogsGroupResource{
		ResourceId:           "telemetrygroup-12345678",
		TelemetryLogsGroupId: "telemetrygroup-12345678",
		Name:                 "example-logs-group",
		CollectorKind:        telemetryv1.TelemetryCollectorKind_TELEMETRY_COLLECTOR_KIND_CLUSTER,
		Groups:               []string{"system", "application", "security"},
	}

	// Example telemetry group resource from Inventory.
	exampleInvTelemetryLogsGroup = &inv_telemetryv1.TelemetryGroupResource{
		ResourceId:    "telemetrygroup-12345678",
		Name:          "example-logs-group",
		Groups:        []string{"system", "application", "security"},
		Kind:          inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
		CollectorKind: inv_telemetryv1.CollectorKind_COLLECTOR_KIND_CLUSTER,
	}

	// Example telemetry logs profile resource from API with instance target.
	exampleAPIInstanceLogsProfile = &telemetryv1.TelemetryLogsProfileResource{
		ResourceId:     "telemetryprofile-12345678",
		ProfileId:      "telemetryprofile-12345678", // Alias of ResourceId
		LogLevel:       telemetryv1.SeverityLevel_SEVERITY_LEVEL_INFO,
		TargetInstance: "inst-12345678",
		LogsGroupId:    "telemetrygroup-12345678",
		LogsGroup:      exampleAPITelemetryLogsGroup,
	}

	// Example telemetry logs profile resource from API with site target.
	exampleAPISiteLogsProfile = &telemetryv1.TelemetryLogsProfileResource{
		ResourceId:  "telemetryprofile-23456789",
		ProfileId:   "telemetryprofile-23456789", // Alias of ResourceId
		LogLevel:    telemetryv1.SeverityLevel_SEVERITY_LEVEL_INFO,
		TargetSite:  "site-12345678",
		LogsGroupId: "telemetrygroup-12345678",
		LogsGroup:   exampleAPITelemetryLogsGroup,
	}

	// Example telemetry logs profile resource from API with region target.
	exampleAPIRegionLogsProfile = &telemetryv1.TelemetryLogsProfileResource{
		ResourceId:   "telemetryprofile-34567890",
		ProfileId:    "telemetryprofile-34567890", // Alias of ResourceId
		LogLevel:     telemetryv1.SeverityLevel_SEVERITY_LEVEL_ERROR,
		TargetRegion: "region-12345678",
		LogsGroupId:  "telemetrygroup-12345678",
		LogsGroup:    exampleAPITelemetryLogsGroup,
	}

	// Example telemetry profile resource from Inventory with instance target.
	exampleInvInstanceLogsProfile = &inv_telemetryv1.TelemetryProfile{
		ResourceId: "telemetryprofile-12345678",
		Kind:       inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
		LogLevel:   inv_telemetryv1.SeverityLevel_SEVERITY_LEVEL_INFO,
		Group:      exampleInvTelemetryLogsGroup,
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
	exampleInvSiteLogsProfile = &inv_telemetryv1.TelemetryProfile{
		ResourceId: "telemetryprofile-23456789",
		Kind:       inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
		LogLevel:   inv_telemetryv1.SeverityLevel_SEVERITY_LEVEL_INFO,
		Group:      exampleInvTelemetryLogsGroup,
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
	exampleInvRegionLogsProfile = &inv_telemetryv1.TelemetryProfile{
		ResourceId: "telemetryprofile-34567890",
		Kind:       inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
		LogLevel:   inv_telemetryv1.SeverityLevel_SEVERITY_LEVEL_ERROR,
		Group:      exampleInvTelemetryLogsGroup,
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
func TestTelemetryLogsProfile_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateTelemetryLogsProfileRequest
		wantErr bool
	}{
		{
			name: "Create TelemetryLogsProfile with Instance target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_TelemetryProfile{
								TelemetryProfile: &inv_telemetryv1.TelemetryProfile{
									ResourceId: "telemetryprofile-12345678",
									Kind:       inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
									LogLevel:   inv_telemetryv1.SeverityLevel_SEVERITY_LEVEL_INFO,
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
			req: &restv1.CreateTelemetryLogsProfileRequest{
				TelemetryLogsProfile: &telemetryv1.TelemetryLogsProfileResource{
					LogLevel:       telemetryv1.SeverityLevel_SEVERITY_LEVEL_INFO,
					LogsGroupId:    "telemetrygroup-12345678",
					TargetInstance: "inst-12345678",
				},
			},
			wantErr: false,
		},
		{
			name: "Create TelemetryLogsProfile with Site target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_TelemetryProfile{
								TelemetryProfile: &inv_telemetryv1.TelemetryProfile{
									ResourceId: "telemetryprofile-23456789",
									Kind:       inv_telemetryv1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_LOGS,
									LogLevel:   inv_telemetryv1.SeverityLevel_SEVERITY_LEVEL_INFO,
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
			req: &restv1.CreateTelemetryLogsProfileRequest{
				TelemetryLogsProfile: &telemetryv1.TelemetryLogsProfileResource{
					LogLevel:    telemetryv1.SeverityLevel_SEVERITY_LEVEL_INFO,
					LogsGroupId: "telemetrygroup-12345678",
					TargetSite:  "site-12345678",
				},
			},
			wantErr: false,
		},
		{
			name: "Create TelemetryLogsProfile with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateTelemetryLogsProfileRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateTelemetryLogsProfile(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateTelemetryLogsProfile() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateTelemetryLogsProfile() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateTelemetryLogsProfile() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetTelemetryLogsProfile(), reply)
		})
	}
}

//nolint:funlen // Testing function.
func TestTelemetryLogsProfile_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name       string
		mocks      func() []*mock.Call
		ctx        context.Context
		req        *restv1.GetTelemetryLogsProfileRequest
		wantErr    bool
		example    *telemetryv1.TelemetryLogsProfileResource
		invExample *inv_telemetryv1.TelemetryProfile
	}{
		{
			name: "Get TelemetryLogsProfile with Instance target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetryprofile-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_TelemetryProfile{
									TelemetryProfile: exampleInvInstanceLogsProfile,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryLogsProfileRequest{
				ResourceId: "telemetryprofile-12345678",
			},
			wantErr:    false,
			example:    exampleAPIInstanceLogsProfile,
			invExample: exampleInvInstanceLogsProfile,
		},
		{
			name: "Get TelemetryLogsProfile with Site target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetryprofile-23456789").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_TelemetryProfile{
									TelemetryProfile: exampleInvSiteLogsProfile,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryLogsProfileRequest{
				ResourceId: "telemetryprofile-23456789",
			},
			wantErr:    false,
			example:    exampleAPISiteLogsProfile,
			invExample: exampleInvSiteLogsProfile,
		},
		{
			name: "Get TelemetryLogsProfile with Region target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetryprofile-23456789").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_TelemetryProfile{
									TelemetryProfile: exampleInvRegionLogsProfile,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryLogsProfileRequest{
				ResourceId: "telemetryprofile-23456789",
			},
			wantErr:    false,
			example:    exampleAPIRegionLogsProfile,
			invExample: exampleInvRegionLogsProfile,
		},
		{
			name: "Get TelemetryLogsProfile with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "telemetryprofile-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetTelemetryLogsProfileRequest{
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

			reply, err := server.GetTelemetryLogsProfile(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetTelemetryLogsProfile() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetTelemetryLogsProfile() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetTelemetryLogsProfile() got reply = nil, want non-nil")
				return
			}
			// Note: Since we have multiple example resources based on target type,
			// we use the appropriate one from the test case
			compareProtoMessages(t, tc.example, reply)
		})
	}
}

//nolint:funlen // Testing function.
func TestTelemetryLogsProfile_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListTelemetryLogsProfilesRequest
		wantErr bool
	}{
		{
			name: "List TelemetryLogsProfiles",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_TelemetryProfile{
											TelemetryProfile: exampleInvInstanceLogsProfile,
										},
									},
								},
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_TelemetryProfile{
											TelemetryProfile: exampleInvSiteLogsProfile,
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
			req: &restv1.ListTelemetryLogsProfilesRequest{
				PageSize:      10,
				Offset:        0,
				ShowInherited: false,
			},
			wantErr: false,
		},
		{
			name: "List inherited TelemetryLogsProfiles",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("ListInheritedTelemetryProfiles",
						mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(&inventory.ListInheritedTelemetryProfilesResponse{
							TelemetryProfiles: []*inv_telemetryv1.TelemetryProfile{
								exampleInvRegionLogsProfile,
								exampleInvSiteLogsProfile,
							},
							TotalElements: 2,
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListTelemetryLogsProfilesRequest{
				PageSize:      10,
				Offset:        0,
				ShowInherited: true,
				InstanceId:    "inst-12345678",
			},
			wantErr: false,
		},
		{
			name: "List TelemetryLogsProfiles with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListTelemetryLogsProfilesRequest{
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

			reply, err := server.ListTelemetryLogsProfiles(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListTelemetryLogsProfiles() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListTelemetryLogsProfiles() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListTelemetryLogsProfiles() got reply = nil, want non-nil")
				return
			}
		})
	}
}

func TestTelemetryLogsProfile_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateTelemetryLogsProfileRequest
		wantErr bool
	}{
		{
			name: "Update TelemetryLogsProfile",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "telemetryprofile-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_TelemetryProfile{
								TelemetryProfile: exampleInvInstanceLogsProfile,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateTelemetryLogsProfileRequest{
				ResourceId:           "telemetryprofile-12345678",
				TelemetryLogsProfile: exampleAPIInstanceLogsProfile,
			},
			wantErr: false,
		},
		{
			name: "Update TelemetryLogsProfile with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "telemetryprofile-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateTelemetryLogsProfileRequest{
				ResourceId:           "telemetryprofile-12345678",
				TelemetryLogsProfile: exampleAPIInstanceLogsProfile,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateTelemetryLogsProfile(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateTelemetryLogsProfile() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateTelemetryLogsProfile() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateTelemetryLogsProfile() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIInstanceLogsProfile, reply)
		})
	}
}

func TestTelemetryLogsProfile_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteTelemetryLogsProfileRequest
		wantErr bool
	}{
		{
			name: "Delete TelemetryLogsProfile",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "telemetryprofile-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteTelemetryLogsProfileRequest{
				ResourceId: "telemetryprofile-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete TelemetryLogsProfile with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "telemetryprofile-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteTelemetryLogsProfileRequest{
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

			reply, err := server.DeleteTelemetryLogsProfile(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteTelemetryLogsProfile() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteTelemetryLogsProfile() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteTelemetryLogsProfile() got reply = nil, want non-nil")
				return
			}
		})
	}
}
