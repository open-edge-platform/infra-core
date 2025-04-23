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
)

// Example resources for testing
var (
	// Example API repeated schedule resource with host target
	exampleAPIHostRepeatedSchedule = &schedulev1.RepeatedScheduleResource{
		ResourceId:         "schedule-12345678",
		RepeatedScheduleId: "schedule-12345678", // Alias of ResourceId
		Name:               "example-host-schedule",
		ScheduleStatus:     schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
		DurationSeconds:    3600, // 1 hour
		CronMinutes:        "*/15",
		CronHours:          "*",
		CronDayMonth:       "*",
		CronMonth:          "*",
		CronDayWeek:        "1-5",
		TargetHostId:       "host-87654321",
		Relation: &schedulev1.RepeatedScheduleResource_TargetHost{
			TargetHost: exampleAPIHostResource,
		},
	}

	// Example API repeated schedule resource with site target
	exampleAPISiteRepeatedSchedule = &schedulev1.RepeatedScheduleResource{
		ResourceId:         "schedule-23456789",
		RepeatedScheduleId: "schedule-23456789", // Alias of ResourceId
		Name:               "example-site-schedule",
		ScheduleStatus:     schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
		DurationSeconds:    7200, // 2 hours
		CronMinutes:        "0",
		CronHours:          "*/2",
		CronDayMonth:       "*",
		CronMonth:          "*",
		CronDayWeek:        "*",
		TargetSiteId:       "site-12345678",
		// We would typically populate the Relation field here,
		// but for testing we can leave it empty as the test would mock the response
	}

	// Example API repeated schedule resource with region target
	exampleAPIRegionRepeatedSchedule = &schedulev1.RepeatedScheduleResource{
		ResourceId:         "schedule-34567890",
		RepeatedScheduleId: "schedule-34567890", // Alias of ResourceId
		Name:               "example-region-schedule",
		ScheduleStatus:     schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
		DurationSeconds:    86400, // 24 hours
		CronMinutes:        "0",
		CronHours:          "0",
		CronDayMonth:       "1",
		CronMonth:          "*",
		CronDayWeek:        "*",
		TargetRegionId:     "region-12345678",
		// We would typically populate the Relation field here,
		// but for testing we can leave it empty as the test would mock the response
	}

	// Example inventory repeated schedule resource with host target
	exampleInvHostRepeatedSchedule = &inv_schedulev1.RepeatedScheduleResource{
		ResourceId:      "schedule-12345678",
		Name:            "example-host-schedule",
		ScheduleStatus:  inv_schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
		DurationSeconds: 3600, // 1 hour
		CronMinutes:     "*/15",
		CronHours:       "*",
		CronDayMonth:    "*",
		CronMonth:       "*",
		CronDayWeek:     "1-5",
		Relation: &inv_schedulev1.RepeatedScheduleResource_TargetHost{
			TargetHost: &inv_computev1.HostResource{
				ResourceId: "host-87654321",
			},
		},
		TenantId:  "tenant-987654",
		CreatedAt: "2025-04-22T10:00:00Z",
		UpdatedAt: "2025-04-22T10:30:00Z",
	}

	// Example inventory repeated schedule resource with site target
	exampleInvSiteRepeatedSchedule = &inv_schedulev1.RepeatedScheduleResource{
		ResourceId:      "schedule-23456789",
		Name:            "example-site-schedule",
		ScheduleStatus:  inv_schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
		DurationSeconds: 7200, // 2 hours
		CronMinutes:     "0",
		CronHours:       "*/2",
		CronDayMonth:    "*",
		CronMonth:       "*",
		CronDayWeek:     "*",
		Relation: &inv_schedulev1.RepeatedScheduleResource_TargetSite{
			TargetSite: &inv_locationv1.SiteResource{
				ResourceId: "site-12345678",
			},
		},
		TenantId:  "tenant-987654",
		CreatedAt: "2025-04-22T11:00:00Z",
		UpdatedAt: "2025-04-22T11:30:00Z",
	}

	// Example inventory repeated schedule resource with region target
	exampleInvRegionRepeatedSchedule = &inv_schedulev1.RepeatedScheduleResource{
		ResourceId:      "schedule-34567890",
		Name:            "example-region-schedule",
		ScheduleStatus:  inv_schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
		DurationSeconds: 86400, // 24 hours
		CronMinutes:     "0",
		CronHours:       "0",
		CronDayMonth:    "1",
		CronMonth:       "*",
		CronDayWeek:     "*",
		Relation: &inv_schedulev1.RepeatedScheduleResource_TargetRegion{
			TargetRegion: &inv_locationv1.RegionResource{
				ResourceId: "region-12345678",
			},
		},
		TenantId:  "tenant-987654",
		CreatedAt: "2025-04-22T12:00:00Z",
		UpdatedAt: "2025-04-22T12:30:00Z",
	}
)

func TestRepeatedSchedule_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateRepeatedScheduleRequest
		wantErr bool
	}{
		{
			name: "Create RepeatedSchedule with host target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Repeatedschedule{
								Repeatedschedule: &inv_schedulev1.RepeatedScheduleResource{
									ResourceId:      "schedule-12345678",
									Name:            "example-host-schedule",
									ScheduleStatus:  inv_schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
									DurationSeconds: 3600,
									CronMinutes:     "*/15",
									Relation: &inv_schedulev1.RepeatedScheduleResource_TargetHost{
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
			req: &restv1.CreateRepeatedScheduleRequest{
				RepeatedSchedule: &schedulev1.RepeatedScheduleResource{
					Name:            "example-host-schedule",
					ScheduleStatus:  schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE,
					DurationSeconds: 3600,
					CronMinutes:     "*/15",
					TargetHostId:    "host-87654321",
				},
			},
			wantErr: false,
		},
		{
			name: "Create RepeatedSchedule with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateRepeatedScheduleRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateRepeatedSchedule(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateRepeatedSchedule() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateRepeatedSchedule() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateRepeatedSchedule() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetRepeatedSchedule(), reply)
		})
	}
}

func TestRepeatedSchedule_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{
		InvClient: mockedClient,
		// InvHCacheClient: mockedClient, // Fixme: This is not used in the test
	}

	cases := []struct {
		name       string
		mocks      func() []*mock.Call
		ctx        context.Context
		req        *restv1.GetRepeatedScheduleRequest
		wantErr    bool
		example    *schedulev1.RepeatedScheduleResource
		invExample *inv_schedulev1.RepeatedScheduleResource
	}{
		{
			name: "Get RepeatedSchedule with host target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetRepeatedSchedule", mock.Anything, "schedule-12345678").
						Return(exampleInvHostRepeatedSchedule, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetRepeatedScheduleRequest{
				ResourceId: "schedule-12345678",
			},
			wantErr:    false,
			example:    exampleAPIHostRepeatedSchedule,
			invExample: exampleInvHostRepeatedSchedule,
		},
		{
			name: "Get RepeatedSchedule with site target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetRepeatedSchedule", mock.Anything, "schedule-23456789").
						Return(exampleInvSiteRepeatedSchedule, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetRepeatedScheduleRequest{
				ResourceId: "schedule-23456789",
			},
			wantErr:    false,
			example:    exampleAPISiteRepeatedSchedule,
			invExample: exampleInvSiteRepeatedSchedule,
		},
		{
			name: "Get RepeatedSchedule with region target",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetRepeatedSchedule", mock.Anything, "schedule-34567890").
						Return(exampleInvRegionRepeatedSchedule, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetRepeatedScheduleRequest{
				ResourceId: "schedule-34567890",
			},
			wantErr:    false,
			example:    exampleAPIRegionRepeatedSchedule,
			invExample: exampleInvRegionRepeatedSchedule,
		},
		{
			name: "Get RepeatedSchedule with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetRepeatedSchedule", mock.Anything, "schedule-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetRepeatedScheduleRequest{
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

			reply, err := server.GetRepeatedSchedule(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetRepeatedSchedule() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetRepeatedSchedule() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetRepeatedSchedule() got reply = nil, want non-nil")
				return
			}
			// Note: Since we have multiple example resources based on target type,
			// we use the appropriate one from the test case
			compareProtoMessages(t, tc.example, reply)
		})
	}
}

func TestRepeatedSchedule_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{
		InvClient: mockedClient,
		// InvHCacheClient: mockedClient, // Fixme: This is not used in the test
	}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListRepeatedSchedulesRequest
		wantErr bool
	}{
		{
			name: "List RepeatedSchedules",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetRepeatedSchedules", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return([]*inv_schedulev1.RepeatedScheduleResource{
							exampleInvHostRepeatedSchedule,
							exampleInvSiteRepeatedSchedule,
							exampleInvRegionRepeatedSchedule,
						}, false, 3, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListRepeatedSchedulesRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List RepeatedSchedules with host filter",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetRepeatedSchedules", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return([]*inv_schedulev1.RepeatedScheduleResource{
							exampleInvHostRepeatedSchedule,
						}, false, 1, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListRepeatedSchedulesRequest{
				PageSize: 10,
				Offset:   0,
				HostId:   "host-87654321",
			},
			wantErr: false,
		},
		{
			name: "List RepeatedSchedules with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("GetRepeatedSchedules", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
						Return(nil, false, 0, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListRepeatedSchedulesRequest{
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

			reply, err := server.ListRepeatedSchedules(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListRepeatedSchedules() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListRepeatedSchedules() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListRepeatedSchedules() got reply = nil, want non-nil")
				return
			}

			// For host filter, we expect only the host schedule
			if tc.req.GetHostId() != "" {
				if len(reply.GetRepeatedSchedules()) != 1 {
					t.Errorf("ListRepeatedSchedules() got %v schedules, want 1", len(reply.GetRepeatedSchedules()))
				}
				compareProtoMessages(t, exampleAPIHostRepeatedSchedule, reply.GetRepeatedSchedules()[0])
			} else {
				// For no filter, we expect all schedules
				if len(reply.GetRepeatedSchedules()) != 3 {
					t.Errorf("ListRepeatedSchedules() got %v schedules, want 3", len(reply.GetRepeatedSchedules()))
				}
			}
		})
	}
}

func TestRepeatedSchedule_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{
		InvClient: mockedClient,
		// InvHCacheClient: mockedClient, // Fixme: This is not used in the test
	}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateRepeatedScheduleRequest
		wantErr bool
	}{
		{
			name: "Update RepeatedSchedule",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "schedule-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Repeatedschedule{
								Repeatedschedule: exampleInvHostRepeatedSchedule,
							},
						}, nil).Once(),
					mockedClient.On("InvalidateCache", mock.Anything, mock.Anything, mock.Anything).
						Return().Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateRepeatedScheduleRequest{
				ResourceId:       "schedule-12345678",
				RepeatedSchedule: exampleAPIHostRepeatedSchedule,
			},
			wantErr: false,
		},
		{
			name: "Update RepeatedSchedule with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "schedule-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateRepeatedScheduleRequest{
				ResourceId:       "schedule-12345678",
				RepeatedSchedule: exampleAPIHostRepeatedSchedule,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateRepeatedSchedule(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateRepeatedSchedule() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateRepeatedSchedule() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateRepeatedSchedule() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIHostRepeatedSchedule, reply)
		})
	}
}

func TestRepeatedSchedule_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{
		InvClient: mockedClient,
		// InvHCacheClient: mockedClient, // Fixme: This is not used in the test
	}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteRepeatedScheduleRequest
		wantErr bool
	}{
		{
			name: "Delete RepeatedSchedule",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "schedule-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
					mockedClient.On("InvalidateCache", mock.Anything, mock.Anything, mock.Anything).
						Return().Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteRepeatedScheduleRequest{
				ResourceId: "schedule-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete RepeatedSchedule with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "schedule-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteRepeatedScheduleRequest{
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

			reply, err := server.DeleteRepeatedSchedule(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteRepeatedSchedule() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteRepeatedSchedule() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteRepeatedSchedule() got reply = nil, want non-nil")
				return
			}
		})
	}
}
