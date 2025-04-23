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

// Example resources for testing
var (
	// Example site resource from the API
	exampleAPISite = &locationv1.SiteResource{
		ResourceId: "site-12345678",
		SiteId:     "site-12345678", // Alias of ResourceId
		Name:       "example-site",
		Region:     exampleAPIRegion, // Using the region from region_test.go
		RegionId:   "region-12345678",
		SiteLat:    377749,
		SiteLng:    -1224194,
		Metadata: []*commonv1.MetadataItem{
			{Key: "environment", Value: "production"},
			{Key: "location", Value: "datacenter-1"},
		},
		InheritedMetadata: []*commonv1.MetadataItem{
			{Key: "org", Value: "engineering"},
		},
	}

	// Example site resource from the Inventory
	exampleInvSite = &inv_locationv1.SiteResource{
		ResourceId: "site-12345678",
		Name:       "example-site",
		SiteLat:    377749,
		SiteLng:    -1224194,
		Metadata:   `[{"key":"environment","value":"production"},{"key":"location","value":"datacenter-1"}]`,
		Region:     exampleInvRegion, // Using the region from region_test.go
		TenantId:   "tenant-987654",
		CreatedAt:  "2025-04-22T10:00:00Z",
		UpdatedAt:  "2025-04-22T10:30:00Z",
	}
)

func TestSite_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateSiteRequest
		wantErr bool
	}{
		{
			name: "Create Site",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Site{
								Site: &inv_locationv1.SiteResource{
									ResourceId: "site-12345678",
									Name:       "example-site",
									SiteLat:    377749,
									SiteLng:    -1224194,
									Region: &inv_locationv1.RegionResource{
										ResourceId: "region-12345678",
									},
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateSiteRequest{
				Site: &locationv1.SiteResource{
					Name:     "example-site",
					SiteLat:  377749,
					SiteLng:  -1224194,
					RegionId: "region-12345678",
				},
			},
			wantErr: false,
		},
		{
			name: "Create Site with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateSiteRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateSite(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateSite() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateSite() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateSite() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetSite(), reply)
		})
	}
}

func TestSite_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetSiteRequest
		wantErr bool
	}{
		{
			name: "Get Site",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "site-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_Site{
									Site: exampleInvSite,
								},
							},
							RenderedMetadata: &inventory.GetResourceResponse_ResourceMetadata{
								PhyMetadata: `[{"key":"environment","value":"production"},{"key":"location","value":"datacenter-1"}]`,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetSiteRequest{
				ResourceId: "site-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get Site with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "site-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetSiteRequest{
				ResourceId: "site-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetSite(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetSite() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetSite() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetSite() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPISite, reply)
		})
	}
}

func TestSite_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListSitesRequest
		wantErr bool
	}{
		{
			name: "List Sites",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_Site{
											Site: exampleInvSite,
										},
									},
									RenderedMetadata: &inventory.GetResourceResponse_ResourceMetadata{
										PhyMetadata: `[{"key":"environment","value":"production"},{"key":"location","value":"datacenter-1"}]`,
									},
								},
							},
							TotalElements: 1,
							HasNext:       false,
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListSitesRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List Sites with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListSitesRequest{
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

			reply, err := server.ListSites(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListSites() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListSites() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListSites() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetSites()) != 1 {
				t.Errorf("ListSites() got %v sites, want 1", len(reply.GetSites()))
			}
			compareProtoMessages(t, exampleAPISite, reply.GetSites()[0])
		})
	}
}

func TestSite_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateSiteRequest
		wantErr bool
	}{
		{
			name: "Update Site",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "site-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Site{
								Site: exampleInvSite,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateSiteRequest{
				ResourceId: "site-12345678",
				Site:       exampleAPISite,
			},
			wantErr: false,
		},
		{
			name: "Update Site with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "site-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateSiteRequest{
				ResourceId: "site-12345678",
				Site:       exampleAPISite,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateSite(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateSite() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateSite() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateSite() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPISite, reply)
		})
	}
}

func TestSite_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteSiteRequest
		wantErr bool
	}{
		{
			name: "Delete Site",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "site-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteSiteRequest{
				ResourceId: "site-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete Site with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "site-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteSiteRequest{
				ResourceId: "site-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteSite(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteSite() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteSite() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteSite() got reply = nil, want non-nil")
				return
			}
		})
	}
}
