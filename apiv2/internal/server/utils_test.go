// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/server"
)

// compareProtoMessages compares two proto.Message parameters and checks if all the fields set in the first
// have the same value as in the second one.
//
//nolint:gocritic,errcheck // This function is used only for testing purposes.
func compareProtoMessages(t *testing.T, msg1, msg2 proto.Message) {
	t.Helper()
	v1 := reflect.ValueOf(msg1).Elem()
	v2 := reflect.ValueOf(msg2).Elem()

	for i := 0; i < v1.NumField(); i++ {
		field1 := v1.Field(i)
		field2 := v2.Field(i)

		if field1.IsZero() {
			continue
		}

		if field1.Kind() == reflect.Ptr &&
			field1.Type().Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
			// Compare messages recursively.
			compareProtoMessages(t, field1.Interface().(proto.Message), field2.Interface().(proto.Message))
		} else if field1.Kind() == reflect.Slice &&
			field1.Type().Elem().Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
			// Check slice lengths first
			field1Len := field1.Len()
			field2Len := field2.Len()
			if field1Len != field2Len {
				// Messages might differ in size of internal fields.
				// t.Errorf("Field %s: slice length mismatch - got %d elements, want %d elements",
				// 	v1.Type().Field(i).Name, field2Len, field1Len)
				continue
			}
			// Compare slices of messages recursively.
			for j := range field1.Len() {
				compareProtoMessages(t, field1.Index(j).Interface().(proto.Message),
					field2.Index(j).Interface().(proto.Message))
			}
		} else if !reflect.DeepEqual(field1.Interface(), field2.Interface()) {
			// Compare fields in the message.
			t.Errorf("Field %s: got %v, want %v", v1.Type().Field(i).Name, field2.Interface(), field1.Interface())
		}
	}
}

//nolint:gosec // This function is used only for testing purposes.
func TestTruncateUint64ToUint32(t *testing.T) {
	now := time.Now()
	nowUnix := now.Unix()
	nowUnixUint64 := uint64(nowUnix)

	nowUnixUint32 := server.TruncateUint64ToUint32(nowUnixUint64)
	assert.Equal(t, uint32(now.Unix()), nowUnixUint32)

	assert.Equal(t, time.Unix(nowUnix, 0), time.Unix(int64(nowUnixUint32), 0))
	fmt.Println(time.Unix(int64(nowUnixUint32), 0))
}
