// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"math"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/open-edge-platform/infra-core/apiv2/v2/internal/pbapi/resources/common/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
)

const ISO8601TimeFormat = "2006-01-02T15:04:05.999Z"

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

// SafeInt32ToUint32 converts an int32 to uint32 safely.
func SafeInt32ToUint32(n int32) (uint32, error) {
	if n < 0 {
		return 0, errors.Errorfc(codes.InvalidArgument, "cannot convert a negative int32 to uint32")
	}
	res := int(n)
	if res > math.MaxUint32 && int32(res) != n { //nolint:gosec // no risk of overflow
		return 0, errors.Errorfc(codes.InvalidArgument, "int exceeds uint32 max limit")
	}
	return uint32(n), nil
}

// SafeUint32ToUint64 converts an uint32 to uint64 safely.
func SafeUint32ToUint64(n uint32) (uint64, error) {
	res := uint64(n)
	if uint32(res) != n { //nolint:gocritic,gosec // no risk of overflow
		return 0, errors.Errorfc(codes.InvalidArgument, "uint32 wrongly converted to uint64")
	}
	return res, nil
}

// TruncateUint64ToUint32 truncates the lower bits of a uint64 to fit into a uint32.
func TruncateUint64ToUint32(value uint64) uint32 {
	return uint32(value & math.MaxUint32) //nolint:gosec // Mask the lower 32 bits.
}

// SafeUint32Toint32 safely converts a uint32 to a int32.
func SafeUint32Toint32(value uint32) (int32, error) {
	if value > math.MaxInt32 {
		return 0, errors.Errorfc(codes.InvalidArgument, "value exceeds int32 range")
	}
	return int32(value), nil
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

func parseFielmask(
	message proto.Message,
	fieldMask *fieldmaskpb.FieldMask,
	fieldsMap map[string]string,
) (*fieldmaskpb.FieldMask, error) {
	if len(fieldMask.GetPaths()) == 0 {
		return fieldMask, nil
	}

	fieldMaskPaths := make([]string, 0, len(fieldMask.Paths))
	for _, path := range fieldMask.GetPaths() {
		fieldName, ok := fieldsMap[path]
		if !ok {
			zlog.Warn().Msgf("Field %s not found in fields map", path)
		} else {
			fieldMaskPaths = append(fieldMaskPaths, fieldName)
		}
	}

	fieldmaskParsed, err := fieldmaskpb.New(message, fieldMaskPaths...)
	if err != nil {
		zlog.InfraErr(err).Msgf("failed to parse fieldmask %s", fieldMask.String())
		return nil, err
	}
	zlog.Debug().Msgf("Fieldmask %s from fields map %s", fieldmaskParsed.String(), fieldMaskPaths)
	return fieldmaskParsed, nil
}

type withCreatedAtUpdatedAtInvRes interface {
	GetCreatedAt() string
	GetUpdatedAt() string
}

func GrpcToOpenAPITimestamps(obj withCreatedAtUpdatedAtInvRes) *commonv1.Timestamps {
	if obj == nil {
		return nil
	}
	createdAt, err := time.Parse(ISO8601TimeFormat, obj.GetCreatedAt())
	if err != nil {
		// In case of error, just log and set time to 0.
		zlog.Err(err).Msg("error when parsing createdAt timestamp, continuing")
		createdAt = time.Unix(0, 0)
	}
	updatedAt, err := time.Parse(ISO8601TimeFormat, obj.GetUpdatedAt())
	if err != nil {
		// In case of error, just log and set time to 0.
		zlog.Err(err).Msg("error when parsing updatedAt timestamp, continuing")
		updatedAt = time.Unix(0, 0)
	}
	return &commonv1.Timestamps{
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}
}
