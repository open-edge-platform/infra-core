// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"errors"
	"math"

	commonv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/common/v1"
)

func fromInvMetadata(metadata string) ([]*commonv1.MetadataItem, error) {
	var apiMetadata []*commonv1.MetadataItem
	if metadata != "" {
		err := json.Unmarshal([]byte(metadata), &apiMetadata)
		if err != nil {
			zlog.InfraErr(err).Msgf("failed to unmarshal metadata: %s", metadata)
			return nil, err
		}
	}
	return apiMetadata, nil
}

func toInvMetadata(apiMetadata []*commonv1.MetadataItem) (string, error) {
	var invMetadata string
	if apiMetadata != nil {
		invMetadataBytes, err := json.Marshal(apiMetadata)
		if err != nil {
			zlog.InfraErr(err).Msgf("failed to marshal metadata: %v", apiMetadata)
			return "", err
		}
		invMetadata = string(invMetadataBytes)
	}
	return invMetadata, nil
}

// SafeIntToUint32 converts an int to uint32 safely.
func SafeIntToUint32(n int) (uint32, error) {
	if n < 0 {
		return 0, errors.New("cannot convert a negative int to uint32")
	}
	if n > math.MaxUint32 {
		return 0, errors.New("int exceeds uint32 max limit")
	}
	return uint32(n), nil
}

// SafeInt32ToUint32 converts an int32 to uint32 safely.
func SafeInt32ToUint32(n int32) (uint32, error) {
	if n < 0 {
		return 0, errors.New("cannot convert a negative int32 to uint32")
	}
	res := int(n)
	if res > math.MaxUint32 && int32(res) != n { //nolint:gosec // no risk of overflow
		return 0, errors.New("int exceeds uint32 max limit")
	}
	return uint32(n), nil
}

// SafeIntToInt32 converts an int to int32 safely.
func SafeIntToInt32(n int) (int32, error) {
	if n < 0 {
		return 0, errors.New("cannot convert a negative int to uint32")
	}
	if n > math.MaxInt32 {
		return 0, errors.New("int exceeds uint32 max limit")
	}
	return int32(n), nil
}

// SafeUint64ToUint32 safely converts a uint64 to a uint32.
func SafeUint64ToUint32(value uint64) (uint32, error) {
	if value > math.MaxUint32 {
		return 0, errors.New("value exceeds uint32 range")
	}
	return uint32(value), nil
}

func isUnset(resourceID *string) bool {
	return resourceID == nil || *resourceID == ""
}

func isSet(resourceID *string) bool {
	return !isUnset(resourceID)
}

// parsePagination parses the pagination fields converting them to limit and offset for the inventory APIs.
func parsePagination(pageSize, off int32) (limit, offset uint32, err error) {
	// We know by design that this cast should never fail, pageSize is limited by the API definition
	limit, err = SafeInt32ToUint32(pageSize)
	if err != nil {
		zlog.InfraErr(err).Msg("error when converting pagination limit/pagesize")
		return 0, 0, err
	}

	offset, err = SafeInt32ToUint32(off)
	if err != nil {
		zlog.InfraErr(err).Msg("error when converting pagination index")
		return 0, 0, err
	}

	return limit, offset, nil
}
