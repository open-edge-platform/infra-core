// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package proxy_test

import (
	"encoding/json"
	"testing"

	api "github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetadataItemValidationWorkaround tests MetadataItem functionality
// separately from the OpenAPI validator to work around the kin-openapi oneOf bug.
// This ensures we still validate MetadataItem serialization without relying
// on the problematic oneOf schema validation in the OpenAPI spec.
func TestMetadataItemValidationWorkaround(t *testing.T) {
	t.Run("MetadataItem with label serializes correctly", func(t *testing.T) {
		// Create a host request with metadata similar to utils.Host1Request
		var item api.MetadataItem
		err := item.FromMetadataItem1(api.MetadataItem1{
			Label: api.LabelItem{Key: "examplekey", Value: "examplevalue"},
		})
		require.NoError(t, err)

		hostResource := api.HostResource{
			Name:     utils.Host1Name,
			Uuid:     &utils.Host1UUID1,
			Metadata: &[]api.MetadataItem{item},
		}

		// Test JSON serialization produces correct structure
		jsonBytes, err := json.Marshal(hostResource)
		require.NoError(t, err)

		// Parse back and verify structure
		var parsedHost map[string]interface{}
		err = json.Unmarshal(jsonBytes, &parsedHost)
		require.NoError(t, err)

		// Verify metadata structure
		metadata, ok := parsedHost["metadata"].([]interface{})
		require.True(t, ok, "metadata should be an array")
		require.Len(t, metadata, 1, "should have one metadata item")

		metadataItem, ok := metadata[0].(map[string]interface{})
		require.True(t, ok, "metadata item should be a map")

		// Verify the label structure is correct
		assert.Contains(t, metadataItem, "label", "should contain 'label' property")

		if labelData, exists := metadataItem["label"]; exists {
			labelMap, ok := labelData.(map[string]interface{})
			require.True(t, ok, "label should be a map")
			assert.Equal(t, "examplekey", labelMap["key"])
			assert.Equal(t, "examplevalue", labelMap["value"])
		}
	})

	t.Run("MetadataItem with annotation serializes correctly", func(t *testing.T) {
		// Test annotation type as well
		var item api.MetadataItem
		err := item.FromMetadataItem0(api.MetadataItem0{
			Annotation: api.AnnotationItem{Key: "config.yaml", Value: "some config"},
		})
		require.NoError(t, err)

		// Test JSON serialization
		jsonBytes, err := json.Marshal(item)
		require.NoError(t, err)

		// Parse and verify structure
		var parsedItem map[string]interface{}
		err = json.Unmarshal(jsonBytes, &parsedItem)
		require.NoError(t, err)

		assert.Contains(t, parsedItem, "annotation", "should contain 'annotation' property")

		if annotationData, exists := parsedItem["annotation"]; exists {
			annotationMap, ok := annotationData.(map[string]interface{})
			require.True(t, ok, "annotation should be a map")
			assert.Equal(t, "config.yaml", annotationMap["key"])
			assert.Equal(t, "some config", annotationMap["value"])
		}
	})

	t.Run("Multiple MetadataItems serialize correctly", func(t *testing.T) {
		// Test multiple metadata items like in utils.Host1RequestUpdate
		var labelItem api.MetadataItem
		err := labelItem.FromMetadataItem1(api.MetadataItem1{
			Label: api.LabelItem{Key: "examplekey", Value: "examplevalue"},
		})
		require.NoError(t, err)

		var annotationItem api.MetadataItem
		err = annotationItem.FromMetadataItem0(api.MetadataItem0{
			Annotation: api.AnnotationItem{Key: "config", Value: "data"},
		})
		require.NoError(t, err)

		metadata := []api.MetadataItem{labelItem, annotationItem}

		// Test serialization of multiple items
		jsonBytes, err := json.Marshal(metadata)
		require.NoError(t, err)

		// Parse and verify
		var parsedMetadata []map[string]interface{}
		err = json.Unmarshal(jsonBytes, &parsedMetadata)
		require.NoError(t, err)

		require.Len(t, parsedMetadata, 2)

		// First item should be label
		assert.Contains(t, parsedMetadata[0], "label")
		// Second item should be annotation
		assert.Contains(t, parsedMetadata[1], "annotation")
	})
}
