// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	customconfigv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/customconfig/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_server "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	inv_computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
)

var exampleInvCustomConfig = &inv_computev1.CustomConfigResource{
	ResourceId:  "customconfig-1234",
	Name:        "example-config",
	Description: "Example description",
	Config:      "config-content",
	CreatedAt:   "2025-06-01T00:00:00Z",
	UpdatedAt:   "2025-06-02T00:00:00Z",
}

var exampleAPICustomConfig = &customconfigv1.CustomConfigResource{
	ResourceId:    "customconfig-1234",
	Name:          "example-config",
	Description:   "Example description",
	ConfigContent: "config-content",
}

func TestCustomConfig_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateCustomConfigRequest
		wantErr bool
	}{
		{
			name: "Create CustomConfig",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_CustomConfig{
								CustomConfig: exampleInvCustomConfig,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateCustomConfigRequest{
				CustomConfig: exampleAPICustomConfig,
			},
			wantErr: false,
		},
		{
			name: "Create CustomConfig with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("create error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateCustomConfigRequest{},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			resp, err := server.CreateCustomConfig(tc.ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				compareProtoMessages(t, tc.req.GetCustomConfig(), resp)
			}
		})
	}
}

func TestCustomConfig_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetCustomConfigRequest
		wantErr bool
	}{
		{
			name: "Get CustomConfig",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "customconfig-1234").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_CustomConfig{
									CustomConfig: exampleInvCustomConfig,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetCustomConfigRequest{
				ResourceId: "customconfig-1234",
			},
			wantErr: false,
		},
		{
			name: "Get CustomConfig with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "customconfig-1234").
						Return(nil, errors.New("get error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetCustomConfigRequest{
				ResourceId: "customconfig-1234",
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			resp, err := server.GetCustomConfig(tc.ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				compareProtoMessages(t, exampleAPICustomConfig, resp)
			}
		})
	}
}

func TestCustomConfig_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListCustomConfigsRequest
		wantErr bool
	}{
		{
			name: "List CustomConfigs",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_CustomConfig{
											CustomConfig: exampleInvCustomConfig,
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
			req: &restv1.ListCustomConfigsRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List CustomConfigs with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("list error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListCustomConfigsRequest{
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
			resp, err := server.ListCustomConfigs(tc.ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, 1, len(resp.CustomConfigs))
				compareProtoMessages(t, exampleAPICustomConfig, resp.CustomConfigs[0])
			}
		})
	}
}

func TestCustomConfig_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteCustomConfigRequest
		wantErr bool
	}{
		{
			name: "Delete CustomConfig",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "customconfig-1234").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteCustomConfigRequest{
				ResourceId: "customconfig-1234",
			},
			wantErr: false,
		},
		{
			name: "Delete CustomConfig with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "customconfig-1234").
						Return(nil, errors.New("delete error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteCustomConfigRequest{
				ResourceId: "customconfig-1234",
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}
			resp, err := server.DeleteCustomConfig(tc.ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}
