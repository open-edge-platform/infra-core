// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	osv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/os/v1"
	restv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/services/v1"
	inv_server "github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
)

// Example resources for testing.
var (
	// Example OS resource from the API.
	exampleAPIOSResource = &osv1.OperatingSystemResource{
		ResourceId:        "os-12345678",
		Name:              "example-os",
		Architecture:      "x86_64",
		KernelCommand:     "quiet splash nomodeset",
		UpdateSources:     []string{"http://repo.example.com/updates"},
		ImageUrl:          "http://images.example.com/example-os.iso",
		ImageId:           "image-12345678",
		Sha256:            "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		ProfileName:       "server",
		ProfileVersion:    "v1",
		InstalledPackages: "linux-kernel",
		SecurityFeature:   osv1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION,
		OsType:            osv1.OsType_OS_TYPE_MUTABLE,
		OsProvider:        osv1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
		OsResourceID:      "os-12345678", // Alias of ResourceId
		Description:       "example description",
		Metadata:          `{"key1": "value1", "key2": "value2"}`,
		ExistingCvesUrl:   "https://example.com/cves",
		ExistingCves: `[
{
  "cve_id": "CVE-000-000",
  "priority": "critical",
  "affected_packages": [
    "test-package-0.0.0",
    "test-2\test3"
  ],
}]`,
		FixedCvesUrl: "/files/fixed_cves.json",
		FixedCves:    `[{"cve_id":"CVE-000-000"}]`,
	}

	// Example OS resource from the Inventory.
	exampleInvOSResource = &inv_osv1.OperatingSystemResource{
		ResourceId:        "os-12345678",
		Name:              "example-os",
		Architecture:      "x86_64",
		KernelCommand:     "quiet splash nomodeset",
		UpdateSources:     []string{"http://repo.example.com/updates"},
		ImageUrl:          "http://images.example.com/example-os.iso",
		ImageId:           "image-12345678",
		Sha256:            "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		ProfileName:       "server",
		ProfileVersion:    "v1",
		InstalledPackages: "linux-kernel",
		SecurityFeature:   inv_osv1.SecurityFeature_SECURITY_FEATURE_SECURE_BOOT_AND_FULL_DISK_ENCRYPTION,
		OsType:            inv_osv1.OsType_OS_TYPE_MUTABLE,
		OsProvider:        inv_osv1.OsProviderKind_OS_PROVIDER_KIND_INFRA,
		CreatedAt:         "2025-04-22T10:00:00Z",
		UpdatedAt:         "2025-04-22T10:30:00Z",
		TenantId:          "tenant-987654",
		Description:       "example description",
		Metadata:          `{"key1": "value1", "key2": "value2"}`,
		ExistingCvesUrl:   "https://example.com/cves",
		ExistingCves: `[
{
  "cve_id": "CVE-000-000",
  "priority": "critical",
  "affected_packages": [
    "test-package-0.0.0",
    "test-2\test3"
  ],
}]`,
		FixedCvesUrl: "/files/fixed_cves.json",
		FixedCves:    `[{"cve_id":"CVE-000-000"}]`,
	}
)

func TestOperatingSystem_Create(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.CreateOperatingSystemRequest
		wantErr bool
	}{
		{
			name: "Create OperatingSystem",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Os{
								Os: &inv_osv1.OperatingSystemResource{
									ResourceId: "os-12345678",
									Name:       "example-os",
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.CreateOperatingSystemRequest{
				Os: &osv1.OperatingSystemResource{
					Name: "example-os",
				},
			},
			wantErr: false,
		},
		{
			name: "Create OperatingSystem with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Create", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx:     context.Background(),
			req:     &restv1.CreateOperatingSystemRequest{},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.CreateOperatingSystem(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("CreateOperatingSystem() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateOperatingSystem() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("CreateOperatingSystem() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, tc.req.GetOs(), reply)
		})
	}
}

func TestOperatingSystem_Get(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.GetOperatingSystemRequest
		wantErr bool
	}{
		{
			name: "Get OperatingSystem",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "os-12345678").
						Return(&inventory.GetResourceResponse{
							Resource: &inventory.Resource{
								Resource: &inventory.Resource_Os{
									Os: exampleInvOSResource,
								},
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetOperatingSystemRequest{
				ResourceId: "os-12345678",
			},
			wantErr: false,
		},
		{
			name: "Get OperatingSystem with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Get", mock.Anything, "os-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.GetOperatingSystemRequest{
				ResourceId: "os-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.GetOperatingSystem(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetOperatingSystem() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("GetOperatingSystem() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("GetOperatingSystem() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIOSResource, reply)
		})
	}
}

func TestOperatingSystem_List(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.ListOperatingSystemsRequest
		wantErr bool
	}{
		{
			name: "List OperatingSystems",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(&inventory.ListResourcesResponse{
							Resources: []*inventory.GetResourceResponse{
								{
									Resource: &inventory.Resource{
										Resource: &inventory.Resource_Os{
											Os: exampleInvOSResource,
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
			req: &restv1.ListOperatingSystemsRequest{
				PageSize: 10,
				Offset:   0,
			},
			wantErr: false,
		},
		{
			name: "List OperatingSystems with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("List", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.ListOperatingSystemsRequest{
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

			reply, err := server.ListOperatingSystems(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ListOperatingSystems() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("ListOperatingSystems() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("ListOperatingSystems() got reply = nil, want non-nil")
				return
			}
			if len(reply.GetOperatingSystemResources()) != 1 {
				t.Errorf("ListOperatingSystems() got %v operating systems, want 1", len(reply.GetOperatingSystemResources()))
			}
			compareProtoMessages(t, exampleAPIOSResource, reply.GetOperatingSystemResources()[0])
		})
	}
}

func TestOperatingSystem_Update(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.UpdateOperatingSystemRequest
		wantErr bool
	}{
		{
			name: "Update OperatingSystem",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "os-12345678", mock.Anything, mock.Anything).
						Return(&inventory.Resource{
							Resource: &inventory.Resource_Os{
								Os: exampleInvOSResource,
							},
						}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateOperatingSystemRequest{
				ResourceId: "os-12345678",
				Os:         exampleAPIOSResource,
			},
			wantErr: false,
		},
		{
			name: "Update OperatingSystem with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Update", mock.Anything, "os-12345678", mock.Anything, mock.Anything).
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.UpdateOperatingSystemRequest{
				ResourceId: "os-12345678",
				Os:         exampleAPIOSResource,
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.UpdateOperatingSystem(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UpdateOperatingSystem() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateOperatingSystem() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("UpdateOperatingSystem() got reply = nil, want non-nil")
				return
			}
			compareProtoMessages(t, exampleAPIOSResource, reply)
		})
	}
}

func TestOperatingSystem_Delete(t *testing.T) {
	mockedClient := newMockedInventoryTestClient()
	server := inv_server.InventorygRPCServer{InvClient: mockedClient}

	cases := []struct {
		name    string
		mocks   func() []*mock.Call
		ctx     context.Context
		req     *restv1.DeleteOperatingSystemRequest
		wantErr bool
	}{
		{
			name: "Delete OperatingSystem",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "os-12345678").
						Return(&inventory.DeleteResourceResponse{}, nil).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteOperatingSystemRequest{
				ResourceId: "os-12345678",
			},
			wantErr: false,
		},
		{
			name: "Delete OperatingSystem with error",
			mocks: func() []*mock.Call {
				return []*mock.Call{
					mockedClient.On("Delete", mock.Anything, "os-12345678").
						Return(nil, errors.New("error")).Once(),
				}
			},
			ctx: context.Background(),
			req: &restv1.DeleteOperatingSystemRequest{
				ResourceId: "os-12345678",
			},
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mocks != nil {
				tc.mocks()
			}

			reply, err := server.DeleteOperatingSystem(tc.ctx, tc.req)
			if tc.wantErr {
				if err == nil {
					t.Errorf("DeleteOperatingSystem() got err = nil, want err")
				}
				return
			}
			if err != nil {
				t.Errorf("DeleteOperatingSystem() got err = %v, want nil", err)
				return
			}
			if reply == nil {
				t.Errorf("DeleteOperatingSystem() got reply = nil, want non-nil")
				return
			}
		})
	}
}
