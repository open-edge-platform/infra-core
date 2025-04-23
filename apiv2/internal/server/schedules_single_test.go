// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	schedulev1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/schedule/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_server "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	inv_schedulev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	schedule_cache "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client/cache/schedule"
)

// Example resources for testing.
var (
	// Example API single schedule resource with host target.
	exampleAPIHostSingleSchedule = &schedulev1.SingleScheduleResource{
		ResourceId:       "schedule-12345678",
		SingleScheduleId: "schedule-12345678", // Alias of ResourceId
		Name:             "example-host-schedule",
		ScheduleStatus:   schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE,
		StartSeconds:     3600, // 1 hour
		EndSeconds:       7200, // 2 hours
		TargetHostId:     "host-87654321",
		Relation: &schedulev1.SingleScheduleResource_TargetHost{
			TargetHost: exampleAPIHostResource,
		},
	}

	// Example API single schedule resource with site target.
	exampleAPISiteSingleSchedule = &schedulev1.SingleScheduleResource{
		ResourceId:       "schedule-23456789",
		SingleScheduleId: "schedule-23456789", // Alias of ResourceId
		Name:             "example-site-schedule",
		ScheduleStatus:   schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE,
		StartSeconds:     43200, // 12 hours
		EndSeconds:       86400, // 24 hours
		TargetSiteId:     "site-12345678",
		// We would typically populate the Relation field here,
		// but for testing we can leave it empty as the test would mock the response.
	}

	// Example API single schedule resource with region target.
	exampleAPIRegionSingleSchedule = &schedulev1.SingleScheduleResource{
		ResourceId:       "schedule-34567890",
		SingleScheduleId: "schedule-34567890", // Alias of ResourceId
		Name:             "example-region-schedule",
		ScheduleStatus:   schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE,
		StartSeconds:     0,
		EndSeconds:       604800, // 7 days
		TargetRegionId:   "region-12345678",
		// We would typically populate the Relation field here,
		// but for testing we can leave it empty as the test would mock the response.
	}

	// Example inventory single schedule resource with host target.
	exampleInvHostSingleSchedule = &inv_schedulev1.SingleScheduleResource{
		ResourceId:     "schedule-12345678",
		Name:           "example-host-schedule",
		ScheduleStatus: inv_schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE,
		StartSeconds:   uint64(3600), // 1 hour
		EndSeconds:     uint64(7200), // 2 hours
		Relation: &inv_schedulev1.SingleScheduleResource_TargetHost{
			TargetHost: &inv_computev1.HostResource{
				ResourceId: "host-87654321",
			},
		},
		TenantId:  "tenant-987654",
		CreatedAt: "2025-04-22T10:00:00Z",
		UpdatedAt: "2025-04-22T10:30:00Z",
	}

	// Example inventory single schedule resource with site target.
	exampleInvSiteSingleSchedule = &inv_schedulev1.SingleScheduleResource{
		ResourceId:     "schedule-23456789",
		Name:           "example-site-schedule",
		ScheduleStatus: inv_schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE,
		StartSeconds:   uint64(43200), // 12 hours
		EndSeconds:     uint64(86400), // 24 hours
		Relation: &inv_schedulev1.SingleScheduleResource_TargetSite{
			TargetSite: &inv_locationv1.SiteResource{
				ResourceId: "site-12345678",
			},
		},
		TenantId:  "tenant-987654",
		CreatedAt: "2025-04-22T11:00:00Z",
		UpdatedAt: "2025-04-22T11:30:00Z",
	}

	// Example inventory single schedule resource with region target.
	exampleInvRegionSingleSchedule = &inv_schedulev1.SingleScheduleResource{
		ResourceId:     "schedule-34567890",
		Name:           "example-region-schedule",
		ScheduleStatus: inv_schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE,
		StartSeconds:   uint64(0),
		EndSeconds:     uint64(604800), // 7 days
		Relation: &inv_schedulev1.SingleScheduleResource_TargetRegion{
			TargetRegion: &inv_locationv1.RegionResource{
				ResourceId: "region-12345678",
			},
		},
		TenantId:  "tenant-987654",
		CreatedAt: "2025-04-22T12:00:00Z",
		UpdatedAt: "2025-04-22T12:30:00Z",
	}
)

func TestSingleSchedule_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateSingleScheduleRequest
		wantErr bool
	}{
		{
			name: "Create SingleSchedule with host target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Singleschedule{
								Singleschedule: &inv_schedulev1.SingleScheduleResource{
									ResourceId:     "schedule-12345678",
									Name:           "example-host-schedule",
									ScheduleStatus: inv_schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE,
									StartSeconds:   uint64(3600),
									EndSeconds:     uint64(7200),
									Relation: &inv_schedulev1.SingleScheduleResource_TargetHost{
										TargetHost: &inv_computev1.HostResource{
											ResourceId: "host-87654321",
										},
									},
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateSingleScheduleRequest{
				SingleSchedule: &schedulev1.SingleScheduleResource{
					Name:           "example-host-schedule",
					ScheduleStatus: schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE,
					StartSeconds:   3600,
					EndSeconds:     7200,
					TargetHostId:   "host-87654321",
				},
			},
			wantErr: false,
		},
		{
			name: "Create SingleSchedule with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateSingleScheduleRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateSingleSchedule(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateSingleSchedule() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateSingleSchedule() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateSingleSchedule() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetSingleSchedule(), reply)
		})
	}
}

//nolint:funlen // TestSingleSchedule_Get tests the GetSingleSchedule method of the InventorygRPCServer.
func TestSingleSchedule_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{
		InvClient:       mockedClient,
		InvHCacheClient: &schedule_cache.HScheduleCacheClient{},
	}

	cases := []struct {
		name       string
		mocks      func() []*mock.Call
		ctx        context.Context
		req        *restv1.GetSingleScheduleRequest
		wantErr    bool
		example    *schedulev1.SingleScheduleResource
		invExample *inv_schedulev1.SingleScheduleResource
	}{
		{
			name: "Get SingleSchedule with host target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetSingleSchedule", mock.Anything, "schedule-12345678").
						Return(exampleInvHostSingleSchedule, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetSingleScheduleRequest{
				ResourceId: "schedule-12345678",
			},
			wantErr:    false,
			example:    exampleAPIHostSingleSchedule,
			invExample: exampleInvHostSingleSchedule,
		},
		{
			name: "Get SingleSchedule with site target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetSingleSchedule", mock.Anything, "schedule-23456789").
						Return(exampleInvSiteSingleSchedule, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetSingleScheduleRequest{
				ResourceId: "schedule-23456789",
			},
			wantErr:    false,
			example:    exampleAPISiteSingleSchedule,
			invExample: exampleInvSiteSingleSchedule,
		},
		{
			name: "Get SingleSchedule with region target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetSingleSchedule", mock.Anything, "schedule-34567890").
						Return(exampleInvRegionSingleSchedule, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetSingleScheduleRequest{
				ResourceId: "schedule-34567890",
			},
			wantErr:    false,
			example:    exampleAPIRegionSingleSchedule,
			invExample: exampleInvRegionSingleSchedule,
		},
		{
			name: "Get SingleSchedule with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetSingleSchedule", mock.Anything, "schedule-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetSingleScheduleRequest{
				ResourceId: "schedule-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetSingleSchedule(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetSingleSchedule() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetSingleSchedule() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetSingleSchedule() got reply = nil, want non-nil")
				return
			}
			// Note: Since we have multiple example resources based on target type,
			// we use the appropriate one from the test case
			compareProtoMessages(t, tc.example, reply)
		})
	}
}

func TestSingleSchedule_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{
		InvClient: mockedClient,
	}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListSingleSchedulesRequest
		wantErr bool
	}{
		{
			name: "List SingleSchedules",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetSingleSchedules",
						mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return([]*inv_schedulev1.SingleScheduleResource{
							exampleInvHostSingleSchedule,
							exampleInvSiteSingleSchedule,
							exampleInvRegionSingleSchedule,
						}, false, 3, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListSingleSchedulesRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List SingleSchedules with host filter",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetSingleSchedules",
						mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return([]*inv_schedulev1.SingleScheduleResource{
							exampleInvHostSingleSchedule,
						}, false, 1, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListSingleSchedulesRequest{
				PageSize: 10,
				Offset:   0,
				HostId:   "host-87654321",
			},
			wantErr: false,
		},
		{
			name: "List SingleSchedules with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetSingleSchedules",
						mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(nil, false, 0, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListSingleSchedulesRequest{
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

			reply, err := server.ListSingleSchedules(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListSingleSchedules() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListSingleSchedules() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListSingleSchedules() got reply = nil, want non-nil")
				return
			}

			// For host filter, we expect only the host schedule
			if tc.req.GetHostId() != "" {
				if len(reply.GetSingleSchedules()) != 1 {
					t.Errorf("ListSingleSchedules() got %v schedules, want 1", len(reply.GetSingleSchedules()))
				}
				compareProtoMessages(t, exampleAPIHostSingleSchedule, reply.GetSingleSchedules()[0])
			} else if len(reply.GetSingleSchedules()) != 3 {
				// For no filter, we expect all schedules
				t.Errorf("ListSingleSchedules() got %v schedules, want 3", len(reply.GetSingleSchedules()))
			}
		})
	}
}

func TestSingleSchedule_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{
		InvClient: mockedClient,
	}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateSingleScheduleRequest
		wantErr bool
	}{
		{
			name: "Update SingleSchedule",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "schedule-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Singleschedule{
								Singleschedule: exampleInvHostSingleSchedule,
							},
						}, nil).Once(),
					mockedClient.On("InvalidateCache", mock.Anything, mock.Anything, mock.Anything).
						Return().Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateSingleScheduleRequest{
				ResourceId:     "schedule-12345678",
				SingleSchedule: exampleAPIHostSingleSchedule,
			},
			wantErr: false,
		},
		{
			name: "Update SingleSchedule with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "schedule-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateSingleScheduleRequest{
				ResourceId:     "schedule-12345678",
				SingleSchedule: exampleAPIHostSingleSchedule,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateSingleSchedule(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateSingleSchedule() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateSingleSchedule() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateSingleSchedule() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIHostSingleSchedule, reply)
		})
	}
}

func TestSingleSchedule_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{
		InvClient: mockedClient,
	}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteSingleScheduleRequest
		wantErr bool
	}{
		{
			name: "Delete SingleSchedule",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "schedule-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
					mockedClient.On("InvalidateCache", mock.Anything, mock.Anything, mock.Anything).
						Return().Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteSingleScheduleRequest{
				ResourceId: "schedule-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete SingleSchedule with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "schedule-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteSingleScheduleRequest{
				ResourceId: "schedule-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteSingleSchedule(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteSingleSchedule() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteSingleSchedule() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteSingleSchedule() got reply = nil, want non-nil")
				return
			}
		})
	}
}
