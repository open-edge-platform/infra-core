// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	providerv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/provider/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_server "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
)

// Example resources for testing.
var (
	// Example provider resource from the API.
	exampleAPIProviderResource = &providerv1.ProviderResource{
		ResourceId:     "provider-12345678",
		ProviderKind:   providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
		ProviderVendor: providerv1.ProviderVendor_PROVIDER_VENDOR_UNSPECIFIED,
		Name:           "example-provider",
		ApiEndpoint:    "https://api.aws.example.com",
		ApiCredentials: []string{"access_key", "AKIAIOSFODNN7EXAMPLE", "secret_key", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"},
		Config:         `{"region": "us-west-2", "zone": "us-west-2a"}`,
		ProviderId:     "provider-12345678", // Alias of ResourceId
	}

	// Example provider resource from the Inventory.
	exampleInvProviderResource = &inv_providerv1.ProviderResource{
		ResourceId:     "provider-12345678",
		ProviderKind:   inv_providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
		ProviderVendor: inv_providerv1.ProviderVendor_PROVIDER_VENDOR_UNSPECIFIED,
		Name:           "example-provider",
		ApiEndpoint:    "https://api.aws.example.com",
		ApiCredentials: []string{"access_key", "AKIAIOSFODNN7EXAMPLE", "secret_key", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"},
		Config:         `{"region": "us-west-2", "zone": "us-west-2a"}`,
		TenantId:       "tenant-987654",
		CreatedAt:      "2025-04-22T10:00:00Z",
		UpdatedAt:      "2025-04-22T10:30:00Z",
	}
)

func TestProvider_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateProviderRequest
		wantErr bool
	}{
		{
			name: "Create Provider",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Provider{
								Provider: &inv_providerv1.ProviderResource{
									ResourceId:     "provider-12345678",
									Name:           "example-provider",
									ProviderKind:   inv_providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
									ProviderVendor: inv_providerv1.ProviderVendor_PROVIDER_VENDOR_UNSPECIFIED,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateProviderRequest{
				Provider: &providerv1.ProviderResource{
					Name:           "example-provider",
					ProviderKind:   providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
					ProviderVendor: providerv1.ProviderVendor_PROVIDER_VENDOR_UNSPECIFIED,
				},
			},
			wantErr: false,
		},
		{
			name: "Create Provider with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateProviderRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateProvider(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateProvider() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateProvider() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateProvider() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetProvider(), reply)
		})
	}
}

func TestProvider_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetProviderRequest
		wantErr bool
	}{
		{
			name: "Get Provider",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "provider-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_Provider{
									Provider: exampleInvProviderResource,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetProviderRequest{
				ResourceId: "provider-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get Provider with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "provider-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetProviderRequest{
				ResourceId: "provider-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetProvider(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetProvider() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetProvider() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetProvider() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIProviderResource, reply)
		})
	}
}

func TestProvider_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListProvidersRequest
		wantErr bool
	}{
		{
			name: "List Providers",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_Provider{
											Provider: exampleInvProviderResource,
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
			req: &restv1.ListProvidersRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List Providers with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListProvidersRequest{
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

			reply, err := server.ListProviders(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListProviders() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListProviders() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListProviders() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetProviders()) != 1 {
				t.Errorf("ListProviders() got %v providers, want 1", len(reply.GetProviders()))
			}
			compareProtoMessages(t, exampleAPIProviderResource, reply.GetProviders()[0])
		})
	}
}

func TestProvider_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateProviderRequest
		wantErr bool
	}{
		{
			name: "Update Provider",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "provider-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Provider{
								Provider: exampleInvProviderResource,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateProviderRequest{
				ResourceId: "provider-12345678",
				Provider:   exampleAPIProviderResource,
			},
			wantErr: false,
		},
		{
			name: "Update Provider with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "provider-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateProviderRequest{
				ResourceId: "provider-12345678",
				Provider:   exampleAPIProviderResource,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateProvider(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateProvider() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateProvider() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateProvider() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIProviderResource, reply)
		})
	}
}

func TestProvider_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteProviderRequest
		wantErr bool
	}{
		{
			name: "Delete Provider",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "provider-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteProviderRequest{
				ResourceId: "provider-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete Provider with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "provider-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteProviderRequest{
				ResourceId: "provider-12345678",
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteProvider(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteProvider() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteProvider() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteProvider() got reply = nil, want non-nil")
				return
			}
		})
	}
}
