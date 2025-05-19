// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package invhandlers

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	"github.com/modern-go/reflect2"
	"github.com/viant/xunsafe"
	"google.golang.org/grpc/codes"

	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	"github.com/open-edge-platform/infra-core/api/pkg/utils"
	inventory "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

const (
	emptyNullCase     = "null"
	emptyCase         = ""
	ISO8601TimeFormat = "2006-01-02T15:04:05.999Z"
)

func marshalMetadata(bodyMetadata *api.Metadata) (string, error) {
	var metadata string

	if bodyMetadata != nil {
		metadataBytes, metaErr := json.Marshal(bodyMetadata)
		if metaErr != nil {
			log.InfraErr(metaErr).Msg("marshal metadata error")
			err := errors.Wrap(metaErr)
			return "", err
		}
		metadata = string(metadataBytes)
	}
	return metadata, nil
}

func unmarshalMetadata(metadataIn string) (*api.Metadata, error) {
	metadataOut := &api.Metadata{}
	if metadataIn != "" {
		metaErr := json.Unmarshal([]byte(metadataIn), metadataOut)
		if metaErr != nil {
			log.InfraErr(metaErr).Msg("unmarshal metadata error")
			err := errors.Wrap(metaErr)
			return nil, err
		}
	}
	return metadataOut, nil
}

// getProtoFieldListFromOpenapiPointer get the field mask from the given openapi object pointer and the given map from
// openAPI fields and Proto field names. The openapi parameter must be a pointer to a struct.
func getProtoFieldListFromOpenapiPointer(
	oapiPointer interface{},
	openAPIToProtoFieldMap map[string]string,
) []string {
	var fieldList []string
	rt := reflect.TypeOf(oapiPointer).Elem()
	rv := reflect.ValueOf(oapiPointer).Elem()
	for i := 0; i < rt.NumField(); i++ {
		field := xunsafe.FieldByIndex(rt, i)
		if !reflect2.IsNil(rv.FieldByName(field.Name).Interface()) {
			v := strings.Split(field.Tag.Get("json"), ",")[0]
			// Set only fields that are actually writable
			if protoFieldName, ok := openAPIToProtoFieldMap[v]; ok {
				fieldList = append(fieldList, protoFieldName)
			}
		}
	}
	return fieldList
}

// getProtoFieldListFromOpenapiPointer get the field mask from the given openapi object (non-pointer) and the given map from
// openAPI fields and Proto field names. The openapi parameter must be directly a struct, not a pointer.
func getProtoFieldListFromOpenapiValue(
	oapiValue interface{},
	openAPIToProtoFieldMap map[string]string,
) []string {
	var fieldList []string
	rt := reflect.TypeOf(oapiValue)
	rv := reflect.ValueOf(oapiValue)
	for i := 0; i < rt.NumField(); i++ {
		field := xunsafe.FieldByIndex(rt, i)
		if !reflect2.IsNil(rv.FieldByName(field.Name).Interface()) {
			v := strings.Split(field.Tag.Get("json"), ",")[0]
			// Set only fields that are actually writable
			if protoFieldName, ok := openAPIToProtoFieldMap[v]; ok {
				fieldList = append(fieldList, protoFieldName)
			}
		}
	}
	return fieldList
}

func isUnset(resourceID *string) bool {
	return resourceID == nil || *resourceID == ""
}

func isSet(resourceID *string) bool {
	return !isUnset(resourceID)
}

// parsePagination parses the pagination fields converting them to limit and offset for the inventory APIs.
func parsePagination(pageSize, off *int) (limit, offset uint32, err error) {
	limit, offset = 20, 0
	if pageSize != nil {
		// We know by design that this cast should never fail, pageSize is limited by the API definition
		limit, err = utils.SafeIntToUint32(*pageSize)
		if err != nil {
			log.InfraErr(err).Msg("error when converting pagination limit/pagesize")
			return 0, 0, err
		}
	}
	if off != nil && pageSize != nil {
		offset, err = util.IntToUint32(*off)
		if err != nil {
			log.InfraErr(err).Msg("error when converting pagination index")
			return 0, 0, err
		}
	}
	return limit, offset, nil
}

func castToInventoryResource(message interface{}) (*inventory.Resource, error) {
	castedMsg, ok := message.(*inventory.Resource)
	if !ok {
		return nil, errors.Errorfc(codes.InvalidArgument, "Expected to obtain inventory.Resource on input, got %T", message)
	}
	return castedMsg, nil
}

func getPtr[T any](v T) *T {
	return &v
}

func GrpcToOpenAPIStatusIndicator(grpcIndicator statusv1.StatusIndication) *api.StatusIndicator {
	indicatorMap := map[statusv1.StatusIndication]api.StatusIndicator{
		statusv1.StatusIndication_STATUS_INDICATION_UNSPECIFIED: api.STATUSINDICATIONUNSPECIFIED,
		statusv1.StatusIndication_STATUS_INDICATION_ERROR:       api.STATUSINDICATIONERROR,
		statusv1.StatusIndication_STATUS_INDICATION_IN_PROGRESS: api.STATUSINDICATIONINPROGRESS,
		statusv1.StatusIndication_STATUS_INDICATION_IDLE:        api.STATUSINDICATIONIDLE,
	}

	apiStatusIndicator, has := indicatorMap[grpcIndicator]
	if !has {
		apiStatusIndicator = indicatorMap[statusv1.StatusIndication_STATUS_INDICATION_UNSPECIFIED]
	}

	return &apiStatusIndicator
}

type withCreatedAtUpdatedAtInvRes interface {
	GetCreatedAt() string
	GetUpdatedAt() string
}

func GrpcToOpenAPITimestamps(obj withCreatedAtUpdatedAtInvRes) *api.Timestamps {
	if obj == nil {
		return nil
	}
	createdAt, err := time.Parse(ISO8601TimeFormat, obj.GetCreatedAt())
	if err != nil {
		// In case of error, just log and set time to 0.
		log.Err(err).Msg("error when parsing createdAt timestamp, continuing")
		createdAt = time.Unix(0, 0)
	}
	updatedAt, err := time.Parse(ISO8601TimeFormat, obj.GetUpdatedAt())
	if err != nil {
		// In case of error, just log and set time to 0.
		log.Err(err).Msg("error when parsing updatedAt timestamp, continuing")
		updatedAt = time.Unix(0, 0)
	}
	return &api.Timestamps{
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}
}
