// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	commonv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/common/v1"
	locationv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/location/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_server "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_locationv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
)

// Example resources for testing.
var (
	// Example parent region resources.
	exampleParentAPIRegion = &locationv1.RegionResource{
		ResourceId:        "region-12345679",
		RegionID:          "region-12345679", // Alias of ResourceId
		Name:              "parent-region",
		Metadata:          []*commonv1.MetadataItem{},
		InheritedMetadata: []*commonv1.MetadataItem{},
	}

	exampleParentInvRegion = &inv_locationv1.RegionResource{
		ResourceId: "region-12345679",
		Name:       "parent-region",
		Metadata:   `[]`,
	}

	// Example region resources.
	exampleAPIRegion = &locationv1.RegionResource{
		ResourceId:        "region-12345678",
		RegionID:          "region-12345678", // Alias of ResourceId
		Name:              "example-region",
		ParentRegion:      exampleParentAPIRegion,
		ParentId:          "region-12345679",
		Metadata:          []*commonv1.MetadataItem{{Key: "env", Value: "test"}},
		InheritedMetadata: []*commonv1.MetadataItem{{Key: "org", Value: "engineering"}},
		TotalSites:        0,
	}

	exampleInvRegion = &inv_locationv1.RegionResource{
		ResourceId:   "region-12345678",
		Name:         "example-region",
		Metadata:     `[{"key":"env","value":"test"}]`,
		ParentRegion: exampleParentInvRegion,
		TenantId:     "tenant-987654",
		CreatedAt:    "2025-04-22T10:00:00Z",
		UpdatedAt:    "2025-04-22T10:30:00Z",
	}
)

func TestRegion_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateRegionRequest
		wantErr bool
	}{
		{
			name: "Create Region",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Region{
								Region: &inv_locationv1.RegionResource{
									ResourceId: "region-12345678",
									Name:       "example-region",
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateRegionRequest{
				Region: &locationv1.RegionResource{
					Name: "example-region",
				},
			},
			wantErr: false,
		},
		{
			name: "Create Region with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateRegionRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateRegion(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateRegion() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateRegion() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateRegion() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetRegion(), reply)
		})
	}
}

func TestRegion_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetRegionRequest
		wantErr bool
	}{
		{
			name: "Get Region",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "region-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_Region{
									Region: exampleInvRegion,
								},
							},
							RenderedMetadata: &inventory.GetResourceResponse_ResourceMetadata{
								PhyMetadata: `[{"key":"org","value":"engineering"}]`,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetRegionRequest{
				ResourceId: "region-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get Region with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "region-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetRegionRequest{
				ResourceId: "region-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetRegion(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetRegion() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetRegion() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetRegion() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIRegion, reply)
		})
	}
}

func TestRegion_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListRegionsRequest
		wantErr bool
	}{
		{
			name: "List Regions",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_Region{
											Region: exampleInvRegion,
										},
									},
									RenderedMetadata: &inventory.GetResourceResponse_ResourceMetadata{
										PhyMetadata: `[{"key":"org","value":"engineering"}]`,
									},
								},
							},
							TotalElements: 1,
							HasNext:       false,
						}, nil).Once(),
					mockedClient.On("GetSitesPerRegion", mock.Anything, mock.Anything).
						Return(&inventory.GetSitesPerRegionResponse{
							Regions: []*inventory.GetSitesPerRegionResponse_Node{
								{
									ResourceId: "region-12345678",
									ChildSites: 0,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListRegionsRequest{
				PageSize:       10,
				Offset:         0,
				ShowTotalSites: true,
			},
			wantErr: false,
		},
		{
			name: "List Regions with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListRegionsRequest{
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

			reply, err := server.ListRegions(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListRegions() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListRegions() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListRegions() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetRegions()) != 1 {
				t.Errorf("ListRegions() got %v regions, want 1", len(reply.GetRegions()))
			}
			compareProtoMessages(t, exampleAPIRegion, reply.GetRegions()[0])
		})
	}
}

func TestRegion_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateRegionRequest
		wantErr bool
	}{
		{
			name: "Update Region",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "region-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Region{
								Region: exampleInvRegion,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateRegionRequest{
				ResourceId: "region-12345678",
				Region:     exampleAPIRegion,
			},
			wantErr: false,
		},
		{
			name: "Update Region with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "region-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateRegionRequest{
				ResourceId: "region-12345678",
				Region:     exampleAPIRegion,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateRegion(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateRegion() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateRegion() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateRegion() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIRegion, reply)
		})
	}
}

func TestRegion_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteRegionRequest
		wantErr bool
	}{
		{
			name: "Delete Region",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "region-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteRegionRequest{
				ResourceId: "region-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete Region with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "region-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteRegionRequest{
				ResourceId: "region-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteRegion(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteRegion() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteRegion() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteRegion() got reply = nil, want non-nil")
				return
			}
		})
	}
}
