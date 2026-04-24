// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package configuration_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	providerv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/flags"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/providerconfiguration"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/tenant-controller/internal/configuration"
)

func TestLoad(t *testing.T) {
	configFile, err := os.CreateTemp("/tmp", "tc-config-*")
	require.NoError(t, err)
	require.NoError(t, configFile.Close())
	defer os.Remove(configFile.Name())

	resources := []*inv_v1.Resource{
		{
			Resource: &inv_v1.Resource_Provider{
				Provider: &providerv1.ProviderResource{
					ProviderKind:   providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
					ProviderVendor: providerv1.ProviderVendor_PROVIDER_VENDOR_UNSPECIFIED,
					Name:           "anyProvider",
					ApiEndpoint:    "anyEndpoint",
					ApiCredentials: nil,
					Config:         `{"enProductKeyIDs":"a,b,c"}`,
					TenantId:       "tid",
				},
			},
		},
		{
			Resource: &inv_v1.Resource_Tenant{
				Tenant: &tenantv1.Tenant{
					ResourceId:       "id",
					CurrentState:     1,
					DesiredState:     1,
					WatcherOsmanager: true,
					TenantId:         "tid",
				},
			},
		},
	}

	err = save(configFile.Name(), resources)
	require.NoError(t, err)

	loadedResources, err := configuration.Load(configFile.Name())
	require.NoError(t, err)
	require.Len(t, loadedResources, 2)

	for idx := range resources {
		areEqual, msg := inv_testing.ProtoEqualOrDiff(loadedResources[idx], resources[idx])
		assert.True(t, areEqual, msg)
	}
}

func TestLoadDefaultConfiguration(t *testing.T) {
	resources, err := configuration.Load("../../configuration/default/resources.json")
	require.NoError(t, err)
	require.NotEmpty(t, resources)

	configFile, err := os.CreateTemp("/tmp", "tc-config-*")
	require.NoError(t, err)
	require.NoError(t, configFile.Close())
	defer os.Remove(configFile.Name())

	err = save(configFile.Name(), resources)
	require.NoError(t, err)

	resources2, err := configuration.Load(configFile.Name())
	require.NoError(t, err)
	require.NotEmpty(t, resources2)

	require.Equal(t, resources, resources2)
}

func TestLoad_failOnMissingFile(t *testing.T) {
	irp, err := configuration.Load("missing-file")
	require.Error(t, err)
	require.ErrorContains(t, err, "no such file")
	require.Nil(t, irp)
}

func TestLoad_failInvalidFormatOfConfigFile(t *testing.T) {
	irp, err := configuration.Load("../../configuration/broken/lenovo/no.json")
	require.Error(t, err)
	require.ErrorContains(t, err, "cannot unmarshal")
	require.Nil(t, irp)
}

func TestLenovoLoad(t *testing.T) {
	configFile, err := os.CreateTemp("/tmp", "tc-config-*")
	require.NoError(t, err)
	require.NoError(t, configFile.Close())
	defer os.Remove(configFile.Name())

	locaConfigs := []configuration.LOCAConfig{
		{
			Name:        "TEST1",
			APIEndpoint: "http://test1.com",
			Username:    "TEST1-USER",
			Password:    "TEST1-PASSWORD",
		},
		{
			Name:        "TEST2",
			APIEndpoint: "http://test2.com",
			Username:    "TEST2-USER",
			Password:    "TEST2-PASSWORD",
		},
	}

	err = saveLOCAConfigs(configFile.Name(), locaConfigs)
	require.NoError(t, err)

	loadedResources, err := configuration.LoadLOCAConfigs(configFile.Name())
	require.NoError(t, err)
	require.Len(t, loadedResources, 2)

	for idx := range locaConfigs {
		areEqual := reflect.DeepEqual(loadedResources[idx], locaConfigs[idx])
		assert.True(t, areEqual)
	}

	emptyConfigFile, err := os.CreateTemp("/tmp", "tc-config-*")
	require.NoError(t, err)
	require.NoError(t, emptyConfigFile.Close())
	defer os.Remove(emptyConfigFile.Name())

	locaConfigs = []configuration.LOCAConfig{}

	err = saveLOCAConfigs(emptyConfigFile.Name(), locaConfigs)
	require.NoError(t, err)

	loadedResources, err = configuration.LoadLOCAConfigs(emptyConfigFile.Name())
	require.NoError(t, err)
	require.Len(t, loadedResources, 0)
}

func TestLenovoConversion(t *testing.T) {
	*flags.FlagDisableCredentialsManagement = true

	configFile, err := os.CreateTemp("/tmp", "tc-config-*")
	require.NoError(t, err)
	require.NoError(t, configFile.Close())
	defer os.Remove(configFile.Name())

	locaConfigs := []configuration.LOCAConfig{
		{
			Name:        "TEST1",
			APIEndpoint: "http://test1.com",
			Username:    "TEST1-USER",
			Password:    "TEST1-PASSWORD",
			DNSDomain:   "kind.internal",
			InstanceTpl: "intel{{#}}",
		},
		{
			Name:        "TEST2",
			APIEndpoint: "http://test2.com",
			Username:    "TEST2-USER",
			Password:    "TEST2-PASSWORD",
			DNSDomain:   "kind.internal",
			InstanceTpl: "intel{{#}}",
		},
	}

	err = saveLOCAConfigs(configFile.Name(), locaConfigs)
	require.NoError(t, err)

	resources := []*inv_v1.Resource{
		{
			Resource: &inv_v1.Resource_Provider{
				Provider: &providerv1.ProviderResource{
					ProviderKind:   providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
					ProviderVendor: providerv1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LOCA,
					Name:           "TEST1",
					ApiEndpoint:    "http://test1.com",
					ApiCredentials: []string{
						"TEST1-secret",
					},
					Config: `{"instance_tpl":"intel{{#}}","dns_domain":"kind.internal"}`,
				},
			},
		},
		{
			Resource: &inv_v1.Resource_Provider{
				Provider: &providerv1.ProviderResource{
					ProviderKind:   providerv1.ProviderKind_PROVIDER_KIND_BAREMETAL,
					ProviderVendor: providerv1.ProviderVendor_PROVIDER_VENDOR_LENOVO_LOCA,
					Name:           "TEST2",
					ApiEndpoint:    "http://test2.com",
					ApiCredentials: []string{
						"TEST2-secret",
					},
					Config: `{"instance_tpl":"intel{{#}}","dns_domain":"kind.internal"}`,
				},
			},
		},
	}

	resourcesProvider, err := configuration.NewLenovoInitResourcesDefinitionLoader(configFile.Name())
	require.NoError(t, err)

	resourcesLoaded := resourcesProvider.Get()
	require.Len(t, resourcesLoaded, 2)

	for idx := range resources {
		areEqual, msg := inv_testing.ProtoEqualOrDiff(resourcesLoaded[idx], resources[idx])
		assert.True(t, areEqual, msg)

		config := &providerconfiguration.LOCAProviderConfig{}
		err := json.Unmarshal([]byte(resourcesLoaded[idx].GetProvider().GetConfig()), config)
		require.NoError(t, err)

		assert.Equal(t, "intel{{#}}", config.InstanceTpl)
		assert.Equal(t, "kind.internal", config.DNSDomain)
	}
}

// save - serializes given resources into json format.
func save(path string, resources []*inv_v1.Resource) error {
	rawMessages := make([]json.RawMessage, 0)

	for _, resource := range resources {
		data, err := protojson.Marshal(resource)
		if err != nil {
			return err
		}
		rawMessages = append(rawMessages, data)
	}

	targetFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	enc := json.NewEncoder(targetFile)
	enc.SetIndent("", "  ")
	return enc.Encode(rawMessages)
}

// Save - serializes given resources into json format.
func saveLOCAConfigs(path string, locaConfigs []configuration.LOCAConfig) error {
	targetFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	enc := json.NewEncoder(targetFile)
	enc.SetIndent("", "  ")
	return enc.Encode(locaConfigs)
}
