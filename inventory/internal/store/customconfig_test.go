// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
