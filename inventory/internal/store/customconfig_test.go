// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/customconfigresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
)

var testCloudInitConfig = `#cloud-config
package_update: true
package_upgrade: true

write_files:
    - path: /etc/environment
      content: |
            http_proxy=http://your.proxy.address:port
            https_proxy=https://your.proxy.address:port
            no_proxy=localhost,127.0.0.1

runcmd:
    - source /etc/environment`

//nolint:funlen // length due to test cases
func Test_Create_Get_Delete_CustomConfig(t *testing.T) {
	testcases := map[string]struct {
		in    *computev1.CustomConfigResource
		valid bool
	}{
		"CreateGoodCustomConfig": {
			in: &computev1.CustomConfigResource{
				Name:        "test-custom-config",
				Description: "Test custom config resource",
				Config:      testCloudInitConfig,
			},
			valid: true,
		},
		"CreateBadCustomConfigWithResourceIdSet": {
			in: &computev1.CustomConfigResource{
				ResourceId:  "customconfig-12345678",
				Name:        "test-custom-config",
				Description: "Test invalid custom config resource",
				Config:      testCloudInitConfig,
			},
			valid: false,
		},
		"CreateBadCustomConfigWithInvalidResourceIdSet": {
			in: &computev1.CustomConfigResource{
				ResourceId:  "customconfig-test-12345678",
				Name:        "test-custom-config",
				Description: "Test invalid custom config resource",
				Config:      testCloudInitConfig,
			},
			valid: false,
		},
		"CreateBadCustomConfigWithLongName": {
			in: &computev1.CustomConfigResource{
				Name:        inv_testing.RandomString(2001),
				Description: "Test custom config resource",
				Config:      testCloudInitConfig,
			},
			valid: false,
		},
		"CreateBadCustomConfigWithLongDescription": {
			in: &computev1.CustomConfigResource{
				Name:        "test-custom-config",
				Description: inv_testing.RandomString(257),
				Config:      testCloudInitConfig,
			},
			valid: false,
		},
		"CreateBadCustomConfigWithLongConfig": {
			in: &computev1.CustomConfigResource{
				Name:        "test-custom-config",
				Description: "Test custom config resource",
				Config:      inv_testing.RandomString(16385),
			},
			valid: false,
		},
		"CreateBadCustomConfigWithNoName": {
			in: &computev1.CustomConfigResource{
				Name:        "",
				Description: "Test custom config resource",
				Config:      testCloudInitConfig,
			},
			valid: false,
		},
		"CreateBadCustomConfigWithNoDescription": {
			in: &computev1.CustomConfigResource{
				Name:        "test-custom-config",
				Description: "",
				Config:      testCloudInitConfig,
			},
			valid: true,
		},
		"CreateBadCustomConfigWithNoConfig": {
			in: &computev1.CustomConfigResource{
				Name:        "test-custom-config",
				Description: "Test custom config resource",
				Config:      "",
			},
			valid: false,
		},
		"CreateBadCustomConfigWithMissingName": {
			in: &computev1.CustomConfigResource{
				Description: "Test custom config resource",
				Config:      testCloudInitConfig,
			},
			valid: false,
		},
		"CreateBadCustomConfigWithMissingDescription": {
			in: &computev1.CustomConfigResource{
				Name:   "test-custom-config",
				Config: testCloudInitConfig,
			},
			valid: true,
		},
		"CreateBadCustomConfigWithMissingConfig": {
			in: &computev1.CustomConfigResource{
				Name:        "test-custom-config",
				Description: "Test custom config resource",
			},
			valid: false,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			createresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_CustomConfig{CustomConfig: tc.in},
			}

			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			// create
			cprovResp, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
			customConfigResID := cprovResp.GetCustomConfig().GetResourceId()

			if err != nil {
				if tc.valid {
					t.Errorf("CreateCustomConfig() failed: %s", err)
				}
			} else {
				tc.in.ResourceId = customConfigResID // Update with created resource ID.
				tc.in.CreatedAt = cprovResp.GetCustomConfig().GetCreatedAt()
				tc.in.UpdatedAt = cprovResp.GetCustomConfig().GetUpdatedAt()
				assertSameResource(t, createresreq, cprovResp, nil)
				if !tc.valid {
					t.Errorf("CreateCustomConfig() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				// get non-existent first
				_, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, "customconfig-12345678")
				require.Error(t, err)

				// get
				getresp, err := inv_testing.TestClients[inv_testing.APIClient].Get(ctx, customConfigResID)
				require.NoError(t, err, "CreateCustomConfig() failed")

				// verify data
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.in, getresp.GetResource().GetCustomConfig()); !eq {
					t.Errorf("GetCustomConfig() data not equal: %v", diff)
				}

				// delete non-existent first
				_, err = inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, "customConfig-12345678")
				require.Error(t, err)

				// delete
				_, err = inv_testing.TestClients[inv_testing.RMClient].Delete(
					ctx,
					customConfigResID,
				)
				if err != nil {
					t.Errorf("DeleteCustomConfig() failed %s", err)
				}
			}
		})
	}
}

//nolint:cyclop,funlen // length due to test cases
func Test_FilterCustomConfig(t *testing.T) {
	customConfig1 := inv_testing.CreateCustomConfig(t,
		"test-custom-config-1",
		"Test custom config resource 1",
		testCloudInitConfig,
	)

	createresreq2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_CustomConfig{
			CustomConfig: &computev1.CustomConfigResource{
				Name:        "test-custom-config-2",
				Description: "Test custom config resource 2",
				Config:      testCloudInitConfig,
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	customConfig2, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq2)
	require.NoError(t, err)
	// Get the resource ID for the created CustomConfig2
	customConfig2ResID := inv_testing.GetResourceIDOrFail(t, customConfig2)
	// Clean up the resource after the test
	t.Cleanup(func() { inv_testing.DeleteResource(t, customConfig2ResID) })

	expCustomConfig1 := customConfig1
	expCustomConfig2 := customConfig2.GetCustomConfig()
	expCustomConfig2.ResourceId = customConfig2ResID

	testcases := map[string]struct {
		in        *inv_v1.ResourceFilter
		resources []*computev1.CustomConfigResource
		valid     bool
	}{
		"NoFilter": {
			in:        &inv_v1.ResourceFilter{},
			resources: []*computev1.CustomConfigResource{expCustomConfig1, expCustomConfig2},
			valid:     true,
		},
		"NoFilterOrderByResourceID": {
			in: &inv_v1.ResourceFilter{
				OrderBy: customconfigresource.FieldResourceID,
			},
			resources: []*computev1.CustomConfigResource{expCustomConfig1, expCustomConfig2},
			valid:     true,
		},
		"FilterByName": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, customconfigresource.FieldName, expCustomConfig2.Name),
			},
			resources: []*computev1.CustomConfigResource{expCustomConfig2},
			valid:     true,
		},
		"FilterByDescription": {
			in: &inv_v1.ResourceFilter{
				Filter: fmt.Sprintf(`%s = %q`, customconfigresource.FieldDescription, expCustomConfig2.Description),
			},
			resources: []*computev1.CustomConfigResource{expCustomConfig2},
			valid:     true,
		},
		"FilterWithOffsetLimit1": {
			in: &inv_v1.ResourceFilter{
				Offset: 5,
				Limit:  0,
			},
			valid: true,
		},
		"FilterWithOffsetLimit2": {
			in: &inv_v1.ResourceFilter{
				Offset: 0,
				Limit:  5,
			},
			resources: []*computev1.CustomConfigResource{expCustomConfig1, expCustomConfig2},
			valid:     true,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			tc.in.Resource = &inv_v1.Resource{Resource: &inv_v1.Resource_CustomConfig{}}
			findres, err := inv_testing.TestClients[inv_testing.APIClient].Find(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("FilterCustomConfig() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("FilterCustomConfig() succeeded but should have failed")
				}
			}

			// only get/delete if valid test with non-zero returned response and hasn't failed, otherwise may segfault
			if !t.Failed() && tc.valid {
				if len(findres.Resources) != len(tc.resources) {
					t.Errorf("Expected to obtain %d Resource IDs, but obtained back %d Resource IDs",
						len(tc.resources), len(findres.Resources))
				}

				resIDs := inv_testing.GetSortedResourceIDSlice(tc.resources)
				inv_testing.SortHasResourceIDAndTenantID(findres.Resources)

				if !reflect.DeepEqual(resIDs, findres.Resources) {
					t.Errorf(
						"FilterCustomConfig() failed - want: %s, got: %s",
						resIDs,
						findres.Resources,
					)
				}
			}

			listres, err := inv_testing.TestClients[inv_testing.APIClient].List(ctx, tc.in)

			if err != nil {
				if tc.valid {
					t.Errorf("ListCustomConfig() failed: %s", err)
				}
			} else {
				if !tc.valid {
					t.Errorf("ListCustomConfig() succeeded but should have failed")
				}
			}

			// only get/delete if valid test and hasn't failed otherwise may segfault
			if !t.Failed() && tc.valid {
				resources := make([]*computev1.CustomConfigResource, 0, len(listres.Resources))
				for _, r := range listres.Resources {
					resources = append(resources, r.GetResource().GetCustomConfig())
				}
				inv_testing.OrderByResourceID(resources)
				inv_testing.OrderByResourceID(tc.resources)
				for i, expected := range tc.resources {
					if eq, diff := inv_testing.ProtoEqualOrDiff(expected, resources[i]); !eq {
						t.Errorf("ListCustomConfig() data not equal: %v", diff)
					}
				}
			}
		})
	}
}

func TestDeleteResources_CustomConfig(t *testing.T) {
	suite.Run(t, &struct{ hardDeleteAllResourcesSuite }{
		hardDeleteAllResourcesSuite: hardDeleteAllResourcesSuite{
			createModel: func(dao *inv_testing.InvResourceDAO) (string, int) {
				tenantID := uuid.NewString()
				return tenantID, len([]any{
					dao.CreateCustomConfigNoCleanup(t, tenantID, "test-custom-config-1",
						"Test custom config 1", testCloudInitConfig),
					dao.CreateCustomConfigNoCleanup(t, tenantID, "test-custom-config-2",
						"Test custom config 2", testCloudInitConfig),
					dao.CreateCustomConfigNoCleanup(t, tenantID, "test-custom-config-3",
						"Test custom config 3", testCloudInitConfig),
				})
			},
			resourceKind: inv_v1.ResourceKind_RESOURCE_KIND_CUSTOMCONFIG,
		},
	})
}

func Test_Unique_CustomConfig_On_Create(t *testing.T) {
	t.Run("CustomConfig_Instance", func(t *testing.T) {
		createresreq := &inv_v1.Resource{
			Resource: &inv_v1.Resource_CustomConfig{
				CustomConfig: &computev1.CustomConfigResource{
					Name:        "test-custom-config",
					Description: "Test custom config resource",
					Config:      testCloudInitConfig,
				},
			},
		}

		// build a context for gRPC
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// create
		_, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
		require.NoError(t, err, "CreateCustomConfig() should Not fail")
		// create another custom config with same name
		_, err = inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
		require.Error(t, err, "CreateCustomConfig() should fail")
		// create another custom config with same name second time
		_, err = inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
		require.Error(t, err, "CreateCustomConfig() should fail")
	})
}

func Test_UpdateCustomConfig(t *testing.T) {
	createresreq := &inv_v1.Resource{
		Resource: &inv_v1.Resource_CustomConfig{
			CustomConfig: &computev1.CustomConfigResource{
				Name:        "test-custom-config-1",
				Description: "Test custom config resource 1",
				Config:      testCloudInitConfig,
			},
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	customConfig, err := inv_testing.TestClients[inv_testing.APIClient].Create(ctx, createresreq)
	require.NoError(t, err)

	// Get the resource ID for CustomConfig
	customConfigResID := inv_testing.GetResourceIDOrFail(t, customConfig)
	// Clean up the resource after the test
	t.Cleanup(func() { inv_testing.DeleteResource(t, customConfigResID) })

	testcases := map[string]struct {
		in         *computev1.CustomConfigResource
		resourceID string
		fieldMask  *fieldmaskpb.FieldMask
		valid      bool
	}{
		"UpdateCustomConfigName": {
			in: &computev1.CustomConfigResource{
				Name:        "test-custom-config-2",
				Description: "Test custom config resource 1",
				Config:      testCloudInitConfig,
			},
			resourceID: customConfigResID,
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					customconfigresource.FieldName,
				},
			},
			valid: false,
		},
	}

	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			updateresreq := &inv_v1.Resource{
				Resource: &inv_v1.Resource_CustomConfig{CustomConfig: tc.in},
			}
			// build a context for gRPC
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			upRes, err := inv_testing.TestClients[inv_testing.RMClient].Update(
				ctx,
				tc.resourceID,
				tc.fieldMask,
				updateresreq,
			)
			if !tc.valid {
				require.Errorf(t, err, "UpdateResource() succeeded but should have failed")
				assert.Nil(t, upRes)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, upRes)
		})
	}
}

func Test_StrongRelations_On_Delete_CustomConfig(t *testing.T) {
	t.Run("CustomConfig_Instance", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		os := inv_testing.CreateOs(t)
		host := inv_testing.CreateHost(t, nil, nil)
		customConfig := inv_testing.CreateCustomConfig(t,
			"test-custom-config",
			"Test custom config resource",
			testCloudInitConfig,
		)

		// Create slice of custom config
		customconfigSlice := []*computev1.CustomConfigResource{customConfig}

		_ = inv_testing.CreateInstanceWithCustomConfig(t, host, os, customconfigSlice)

		_, err := inv_testing.TestClients[inv_testing.APIClient].Delete(ctx, customConfig.ResourceId)

		require.Error(t, err, "DeleteCustomConfig() should fail")
	})
}
