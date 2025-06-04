// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store_test

import (
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/store"
	telemetry_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
)

type Metadata struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var (
	// valid metadata key and value.
	Metadata1 = []Metadata{
		{
			Key:   "cluster.orchestration.io/cluster-id",
			Value: "clusterid-1234",
		}, {
			Key:   "cluster-id",
			Value: "clusterid-12345",
		}, {
			Key:   "cluster-id_1-test",
			Value: "clusterid-test",
		}, {
			Key:   "123test-123", // numeric at begin and last.
			Value: "123test-other.symbol_123",
		}, {
			Key:   "0123.test",
			Value: "0123456789",
		}, {
			Key:   "0123",
			Value: "0123",
		}, {
			Key:   "k",
			Value: "v",
		}, {
			Key:   "",
			Value: "",
		}, {
			Key:   "example.com/a",
			Value: "v",
		}, {
			Key:   "example.com/",
			Value: "v",
		}, {
			Key:   "example.com/8",
			Value: "12",
		}, {
			Key:   "example.com/a9_9",
			Value: "12",
		}, {
			Key:   "123_4-test",
			Value: "123test-test_123",
		}, {
			Key:   "test.com/test-123_name.test",
			Value: "123test-other.symbol_123",
		}, {
			Key:   "example.com/2-test_9",
			Value: "123test-other.symbol_123",
		},
	}
	// invalid metadata key with upper case char.
	Metadata2 = []Metadata{
		{
			Key:   "Cluster-id",
			Value: "clusterid-1234",
		},
	}
	// invalid metadata key no prefix.
	Metadata3 = []Metadata{
		{
			Key:   "/cluster-id",
			Value: "clusterid-1234",
		},
	}
	// invalid metadata key with upper case char at end.
	Metadata4 = []Metadata{
		{
			Key:   "cluster-ID",
			Value: "clusterid-1234",
		},
	}
	// invalid metadata value with upper case char at begin.
	Metadata5 = []Metadata{
		{
			Key:   "cluster-id",
			Value: "Clusterid-test",
		},
	}

	// invalid meatadata value length > 63.
	Metadata6 = []Metadata{
		{
			Key:   "cluster-id",
			Value: "invalidvaluelengthinvalidvaluelengthinvalidvaluelengthinvalidval",
		},
	}
	// invalid metadata key( name )length > 63.
	Metadata7 = []Metadata{
		{
			Key:   "cluster.com/invalidkeylengthinvalidkeylengthinvalidkeylengthinvalidkeylength",
			Value: "clusterid-1234",
		},
	}
	// invalid prefix length > 253.
	Metadata8 = []Metadata{
		{
			Key: `invalidprefixlengthinvalidprefixlengthinvalidprefixlengthinvalidprefix
			lengthinvalidprefixlengthinvalidprefixlengthinvalidprefixlengthinvalidprefix
			lengthinvalidprefixlengthinvalidprefixlengthinvalidprefixlengthinva
			lidprefixlengthinvalidprefixlengthinvali/validname`,
			Value: "clusterid-1234",
		},
	}
	// invalid metadata key with prefix upper case char.
	Metadata9 = []Metadata{
		{
			Key:   "Test.com/id",
			Value: "test",
		},
	}
	// invalid metadata key with other symbol at last.
	Metadata10 = []Metadata{
		{
			Key:   "test1234-",
			Value: "test",
		},
	}
	// invalid metadata key with other symbol at begin.
	Metadata11 = []Metadata{
		{
			Key:   "_test1234",
			Value: "test",
		},
	}
	// invalid metadata key name with other symbol at begin.
	Metadata12 = []Metadata{
		{
			Key:   "test.com/-",
			Value: "test",
		},
	}
	// invalid metadata key name othersymbol at last.
	Metadata13 = []Metadata{
		{
			Key:   "test.com/1a_",
			Value: "0123456789",
		},
	} // invalid metadata key name upper case.
	Metadata14 = []Metadata{
		{
			Key:   "test.com/A",
			Value: "0123456789",
		},
	}
)

func Test_ValidateMetadata(t *testing.T) {
	testcases := map[string]struct {
		in    []Metadata
		valid bool
	}{
		"ValidMetadatakeyAndValue":                         {in: Metadata1, valid: true},
		"InValidMetadatakeyWithUppercaseChar":              {in: Metadata2, valid: false},
		"InValidMetadatakeyNameNoPrefix":                   {in: Metadata3, valid: false},
		"InValidMetadatakeyWithUppercaseLastChar":          {in: Metadata4, valid: false},
		"InValidMetadataValueWithUppercaseChar":            {in: Metadata5, valid: false},
		"InValidMetadataValueLength":                       {in: Metadata6, valid: false},
		"InValidMetadataKeyNameLength":                     {in: Metadata7, valid: false},
		"InValidMetadataKeyPrefixLength":                   {in: Metadata8, valid: false},
		"InValidMetadataKeyPrefixUppercaseChar":            {in: Metadata9, valid: false},
		"InValidMetadataKeyNameOtherSymbolLast":            {in: Metadata10, valid: false},
		"InValidMetadataKeyNameOtherSymbolBegin":           {in: Metadata11, valid: false},
		"InValidMetadataKeyNameOtherSymbolBeginwithPrefix": {in: Metadata12, valid: false},
		"InValidMetadataKeyNameOtherSymbolLastwithPrefix":  {in: Metadata13, valid: false},
		"InValidMetadataKeyNameUpperCasewithPrefix":        {in: Metadata14, valid: false},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			bytes, err := json.Marshal(tc.in)
			if err != nil {
				t.Errorf("Error while marshaling the metadata  %s", err)
			}
			_, err = store.ValidateMetadata(string(bytes))
			if err != nil {
				if !tc.valid {
					assert.Error(t, err)
				}
			}
		})
	}
}

func Test_EmptyEnumStateMap(t *testing.T) {
	v, err := store.EmptyEnumStateMap("invalid_input",
		int32(telemetry_v1.TelemetryResourceKind_TELEMETRY_RESOURCE_KIND_METRICS))
	assert.Error(t, err)
	assert.Nil(t, v)
}

func Test_ValidateOSMetadata(t *testing.T) {
	testcases := map[string]struct {
		in    string
		valid bool
	}{
		"ValidMetadatakeyAndValue":                         {in: helperMetadataToJSONPlain(t, Metadata1), valid: true},
		"InValidMetadatakeyWithUppercaseChar":              {in: helperMetadataToJSONPlain(t, Metadata2), valid: false},
		"InValidMetadatakeyNameNoPrefix":                   {in: helperMetadataToJSONPlain(t, Metadata3), valid: false},
		"InValidMetadatakeyWithUppercaseLastChar":          {in: helperMetadataToJSONPlain(t, Metadata4), valid: false},
		"InValidMetadataValueWithUppercaseChar":            {in: helperMetadataToJSONPlain(t, Metadata5), valid: false},
		"InValidMetadataValueLength":                       {in: helperMetadataToJSONPlain(t, Metadata6), valid: false},
		"InValidMetadataKeyNameLength":                     {in: helperMetadataToJSONPlain(t, Metadata7), valid: false},
		"InValidMetadataKeyPrefixLength":                   {in: helperMetadataToJSONPlain(t, Metadata8), valid: false},
		"InValidMetadataKeyPrefixUppercaseChar":            {in: helperMetadataToJSONPlain(t, Metadata9), valid: false},
		"InValidMetadataKeyNameOtherSymbolLast":            {in: helperMetadataToJSONPlain(t, Metadata10), valid: false},
		"InValidMetadataKeyNameOtherSymbolBegin":           {in: helperMetadataToJSONPlain(t, Metadata11), valid: false},
		"InValidMetadataKeyNameOtherSymbolBeginwithPrefix": {in: helperMetadataToJSONPlain(t, Metadata12), valid: false},
		"InValidMetadataKeyNameOtherSymbolLastwithPrefix":  {in: helperMetadataToJSONPlain(t, Metadata13), valid: false},
		"InValidMetadataKeyNameUpperCasewithPrefix":        {in: helperMetadataToJSONPlain(t, Metadata14), valid: false},
		"InvalidMetadataDuplicateKeys": {
			in:    `{"key1":"value1","key2":"value2","key1":"value3"}`,
			valid: false,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			bytes, err := json.Marshal(tc.in)
			if err != nil {
				t.Errorf("Error while marshaling the metadata  %s", err)
			}
			_, err = store.ValidateOSMetadata(string(bytes))
			if err != nil {
				if !tc.valid {
					assert.Error(t, err)
				}
			}
		})
	}
}

func helperMetadataToJSONPlain(t *testing.T, metadata []Metadata) string {
	t.Helper()

	result := make(map[string]string)
	for _, m := range metadata {
		result[m.Key] = m.Value
	}
	val, err := json.Marshal(result)
	require.NoError(t, err, "Failed to marshal metadata to JSON")
	return string(val)
}
