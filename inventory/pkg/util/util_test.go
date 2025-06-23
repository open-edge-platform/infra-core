// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"flag"
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent/hostresource"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	network_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	ou_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/ou/v1"
	provider_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	schedule_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	telemetry_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/comparator"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/filters"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/paginator"
)

// Define this flag in order to call all tests with the same parameters.
var _ = flag.String(
	"policyBundle",
	"/rego/policy_bundle.tar.gz",
	"Path of policy rego file",
)

func TestResourceKindTransitiveConversion(t *testing.T) {
	for _, val := range inv_v1.ResourceKind_value {
		kind := inv_v1.ResourceKind(val)
		got := util.PrefixToResourceKind(util.ResourceKindToPrefix(kind))
		if !reflect.DeepEqual(got, kind) {
			t.Errorf("PrefixToResourceKind(ResourceKindToPrefix(%v)) != %v", kind, got)
		}
	}
}

func Test_ValidateAndFilterMessage(t *testing.T) {
	testcases := map[string]struct {
		in        *computev1.HostResource
		fieldmask *fieldmaskpb.FieldMask
		filter    bool
		valid     bool
	}{
		"ValidMessageAndFieldmask1": {
			in: &computev1.HostResource{
				CpuCores:     8,
				BmcIp:        "10.11.12.14",
				DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
			},
			fieldmask: &fieldmaskpb.FieldMask{
				Paths: []string{hostresource.FieldBmcIP},
			},
			filter: true,
			valid:  true,
		},
		"ValidMessageAndFieldmask2": {
			in: &computev1.HostResource{
				CpuCores:     8,
				BmcIp:        "10.11.12.14",
				DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
			},
			fieldmask: &fieldmaskpb.FieldMask{
				Paths: []string{hostresource.FieldBmcIP},
			},
			filter: false,
			valid:  true,
		},
		"EmptyFieldmask": {
			in: &computev1.HostResource{
				CpuCores:     8,
				BmcIp:        "10.11.12.14",
				DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
			},
			filter: true,
			valid:  true,
		},
		"EmptyMessage": {
			fieldmask: &fieldmaskpb.FieldMask{
				Paths: []string{hostresource.FieldBmcIP},
			},
			filter: true,
			valid:  true,
		},
		"InvalidFieldmask": {
			in: &computev1.HostResource{
				CpuCores:     8,
				BmcIp:        "10.11.12.14",
				DesiredState: computev1.HostState_HOST_STATE_ONBOARDED,
			},
			fieldmask: &fieldmaskpb.FieldMask{
				Paths: []string{"INVALID_PATH"},
			},
			filter: true,
			valid:  false,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			err := util.ValidateMaskAndFilterMessage(tc.in, tc.fieldmask, tc.filter)
			if tc.valid {
				if err != nil {
					t.Errorf("ValidateMaskAndFilterMessage() errored when was not expecting: %v", err)
				}
			} else if !tc.valid {
				if err == nil {
					t.Errorf("ValidateMaskAndFilterMessage() hasn't generated and error when expected")
				}
			}
		})
	}
}

func Test_GetResourceKindFromMessage(t *testing.T) {
	testcases := map[string]struct {
		in       proto.Message
		expected inv_v1.ResourceKind
		valid    bool
	}{
		"Vm": {
			in:       &computev1.InstanceResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_INSTANCE,
			valid:    true,
		},
		"Host": {
			in:       &computev1.HostResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_HOST,
			valid:    true,
		},
		"HostStorage": {
			in:       &computev1.HoststorageResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_HOSTSTORAGE,
			valid:    true,
		},
		"Hostnic": {
			in:       &computev1.HostnicResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_HOSTNIC,
			valid:    true,
		},
		"NetworkSegment": {
			in:       &network_v1.NetworkSegment{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_NETWORKSEGMENT,
			valid:    true,
		},
		"Netlink": {
			in:       &network_v1.NetlinkResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_NETLINK,
			valid:    true,
		},
		"Endpoint": {
			in:       &network_v1.EndpointResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_ENDPOINT,
			valid:    true,
		},
		"Region": {
			in:       &location_v1.RegionResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_REGION,
			valid:    true,
		},
		"Site": {
			in:       &location_v1.SiteResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_SITE,
			valid:    true,
		},
		"Provider": {
			in:       &provider_v1.ProviderResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_PROVIDER,
			valid:    true,
		},
		"Os": {
			in:       &osv1.OperatingSystemResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_OS,
			valid:    true,
		},
		"SingleSchedule": {
			in:       &schedule_v1.SingleScheduleResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE,
			valid:    true,
		},
		"RepeatedSchedule": {
			in:       &schedule_v1.RepeatedScheduleResource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE,
			valid:    true,
		},
		"Invalid": {
			in:    &anypb.Any{},
			valid: false,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			kind, err := util.GetResourceKindFromMessage(tc.in)
			if tc.valid {
				if err != nil {
					t.Errorf("Got unexpected error: %v", err)
				}
				if kind != tc.expected {
					t.Errorf("Wrong resource prefix: %v", kind)
				}
			} else if err == nil {
				t.Errorf("Expected error, but got none")
			}
		})
	}
}

func Test_NewInvID(t *testing.T) {
	for _, val := range inv_v1.ResourceKind_value {
		kind := inv_v1.ResourceKind(val)
		id := util.NewInvID(kind)
		assert.True(t, strings.HasPrefix(id, string(util.ResourceKindToPrefix(kind))))
		r := regexp.MustCompile("^[[:lower:]]+-[0-9,a-z]{8}$")
		if !r.MatchString(id) {
			t.Errorf("ID contains unexpected characters! ID=%v", id)
		}
	}
}

func Test_BuildFieldMaskFromMessage(t *testing.T) {
	tests := []struct {
		name       string
		in         *computev1.HostResource
		skipFields []string
		want       []string
		wantErr    bool
	}{
		{
			name: "Success",
			in: &computev1.HostResource{
				CpuCores:    1,
				MemoryBytes: 4 * util.Gigabyte,
				Name:        "Test Resource",
			},
			want:    []string{"name", "memory_bytes", "cpu_cores"},
			wantErr: false,
		},
		{
			name: "SuccessSkipFields",
			in: &computev1.HostResource{
				CpuCores:    1,
				MemoryBytes: 4 * util.Gigabyte,
				Name:        "Test Resource",
			},
			skipFields: []string{"cpu_cores"},
			want:       []string{"name", "memory_bytes"},
			wantErr:    false,
		},
		{
			name: "SuccessSkipFieldsEmpty",
			in: &computev1.HostResource{
				CpuCores:    1,
				MemoryBytes: 4 * util.Gigabyte,
				Name:        "Test Resource",
			},
			skipFields: []string{""},
			want:       []string{"name", "memory_bytes", "cpu_cores"},
			wantErr:    false,
		},
		{
			name: "SuccessSkipFieldsNonExistent",
			in: &computev1.HostResource{
				CpuCores:    1,
				MemoryBytes: 4 * util.Gigabyte,
				Name:        "Test Resource",
			},
			skipFields: []string{"some_field"},
			want:       []string{"name", "memory_bytes", "cpu_cores"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.BuildFieldMaskFromMessage(tt.in, tt.skipFields...)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildFieldMaskFromMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got.Paths, tt.want) {
				t.Errorf("BuildFieldMaskFromMessage() got = %v, want %v", got.Paths, tt.want)
			}
		})
	}
}

func Test_BuildAllFieldMaskFromProto(t *testing.T) {
	tests := []struct {
		name       string
		in         *ou_v1.OuResource
		skipFields []string
		want       []string
		wantErr    bool
	}{
		{
			name: "Success",
			in: &ou_v1.OuResource{
				ResourceId: "ou-1",
				Name:       "some name",
				OuKind:     "some kind",
				ParentOu:   nil,
				Children:   nil,
				Metadata:   "",
			},
			want: []string{
				"resource_id", "name", "ou_kind", "parent_ou", "children", "metadata", "tenant_id", "created_at", "updated_at",
			},
			wantErr: false,
		},
		{
			name: "SuccessSkipFields",
			in: &ou_v1.OuResource{
				ResourceId: "ou-1",
				Name:       "some name",
				OuKind:     "some kind",
				ParentOu:   nil,
				Children:   nil,
				Metadata:   "",
			},
			skipFields: []string{"resource_id", "ou_kind", "metadata", "tenant_id", "created_at"},
			want:       []string{"name", "parent_ou", "children", "updated_at"},
			wantErr:    false,
		},
		{
			name: "SuccessSkipFieldsEmpty",
			in: &ou_v1.OuResource{
				ResourceId: "ou-1",
				Name:       "some name",
				OuKind:     "some kind",
				ParentOu:   nil,
				Children:   nil,
				Metadata:   "",
			},
			skipFields: []string{""},
			want: []string{
				"resource_id", "name", "ou_kind", "parent_ou", "children", "metadata", "tenant_id", "created_at", "updated_at",
			},
			wantErr: false,
		},
		{
			name: "SuccessSkipFieldsNonExistent",
			in: &ou_v1.OuResource{
				ResourceId: "ou-1",
				Name:       "some name",
				OuKind:     "some kind",
				ParentOu:   nil,
				Children:   nil,
				Metadata:   "",
			},
			skipFields: []string{"some_field"},
			want: []string{
				"resource_id", "name", "ou_kind", "parent_ou", "children", "metadata", "tenant_id", "created_at", "updated_at",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.BuildAllFieldMaskFromProto(tt.in, tt.skipFields...)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildAllFieldMaskFromProto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got.Paths, tt.want) {
				t.Errorf("BuildAllFieldMaskFromProto() got = %v, want %v", got.Paths, tt.want)
			}
		})
	}
}

func Test_GetResourceKind(t *testing.T) {
	testcases := map[string]struct {
		in       *inv_v1.Resource
		expected inv_v1.ResourceKind
	}{
		"Host": {
			in: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Host{},
			},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_HOST,
		},
		"Workload": {
			in: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Workload{},
			},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_WORKLOAD,
		},
		"Invalid1": {
			in:       &inv_v1.Resource{},
			expected: inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED,
		},
		"Invalid2": {
			in:       nil,
			expected: inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED,
		},
	}
	for tcName, tc := range testcases {
		t.Run(tcName, func(t *testing.T) {
			kind := util.GetResourceKindFromResource(tc.in)
			assert.Equal(t, tc.expected, kind)
		})
	}
}

func TestIntToUint32(t *testing.T) {
	tests := []struct {
		name    string
		i       int
		want    uint32
		wantErr bool
	}{
		{name: "Negative", i: -1, want: 0, wantErr: true},
		{name: "Zero", i: 0, want: 0, wantErr: false},
		{name: "One", i: 1, want: 1, wantErr: false},
		{name: "MaxInt", i: math.MaxInt, want: 0, wantErr: true},
		{name: "MaxUint32", i: math.MaxUint32, want: math.MaxUint32, wantErr: false},
		{name: "MaxUint32+1", i: math.MaxUint32 + 1, want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.IntToUint32(tt.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("IntToUint32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IntToUint32() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint64ToUint32(t *testing.T) {
	tests := []struct {
		name    string
		i       uint64
		want    uint32
		wantErr bool
	}{
		{name: "Zero", i: 0, want: 0, wantErr: false},
		{name: "One", i: 1, want: 1, wantErr: false},
		{name: "MaxUint32", i: math.MaxUint32, want: math.MaxUint32, wantErr: false},
		{name: "MaxUint32+1", i: math.MaxUint32 + 1, want: 0, wantErr: true},
		{name: "MaxUint64", i: math.MaxUint64, want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.Uint64ToUint32(tt.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint64ToUint32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Uint64ToUint32() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntToInt32(t *testing.T) {
	tests := []struct {
		name    string
		i       int
		want    int32
		wantErr bool
	}{
		{name: "Negative", i: -1, want: -1, wantErr: false},
		{name: "Zero", i: 0, want: 0, wantErr: false},
		{name: "One", i: 1, want: 1, wantErr: false},
		{name: "MaxInt", i: math.MaxInt, want: 0, wantErr: true},
		{name: "MinInt", i: math.MinInt, want: 0, wantErr: true},
		{name: "MaxInt32", i: math.MaxInt32, want: math.MaxInt32, wantErr: false},
		{name: "MinInt32", i: math.MinInt32, want: math.MinInt32, wantErr: false},
		{name: "MaxInt32+1", i: math.MaxInt32 + 1, want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.IntToInt32(tt.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("IntToInt32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IntToInt32() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUint32ToInt(t *testing.T) {
	tests := []struct {
		name    string
		i       uint32
		want    int
		wantErr bool
	}{
		{name: "Zero", i: 0, want: 0, wantErr: false},
		{name: "One", i: 1, want: 1, wantErr: false},
		{name: "MaxUint32", i: math.MaxUint32, want: 0, wantErr: true},
		{name: "MaxInt32", i: math.MaxInt32, want: math.MaxInt32, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.Uint32ToInt(tt.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uint32ToInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Uint32ToInt() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt64ToInt32(t *testing.T) {
	tests := []struct {
		name    string
		i       int64
		want    int32
		wantErr bool
	}{
		{name: "Zero", i: 0, want: 0, wantErr: false},
		{name: "One", i: 1, want: 1, wantErr: false},
		{name: "MaxInt32", i: math.MaxInt32, want: math.MaxInt32, wantErr: false},
		{name: "MaxInt32+1", i: math.MaxInt32 + 1, want: 0, wantErr: true},
		{name: "MaxInt64", i: math.MaxInt64, want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.Int64ToInt32(tt.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("Int64ToInt32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Int64ToInt32() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMulInt64(t *testing.T) {
	tests := []struct {
		name    string
		left    int64
		right   int64
		want    int64
		wantErr bool
	}{
		{name: "pos*pos", left: 10, right: 20, want: 200, wantErr: false},
		{name: "pos*neg", left: 10, right: -20, want: -200, wantErr: false},
		{name: "neg*pos", left: -10, right: 20, want: -200, wantErr: false},
		{name: "neg*neg", left: -10, right: -20, want: 200, wantErr: false},
		{name: "min*zero", left: math.MinInt64, right: 0, wantErr: false},
		{name: "min*one", left: math.MinInt64, right: 1, want: math.MinInt64, wantErr: false},
		{name: "min*two", left: math.MinInt64, right: 2, wantErr: true},
		{name: "one*min", left: 1, right: math.MinInt64, want: math.MinInt64, wantErr: false},
		{name: "max*zero", left: math.MaxInt64, right: 0, want: 0, wantErr: false},
		{name: "max*one", left: math.MaxInt64, right: 1, want: math.MaxInt64, wantErr: false},
		{name: "max*two", left: math.MaxInt64, right: 2, wantErr: true},
		{name: "overflow2", left: math.MaxInt64, right: -2, wantErr: true},
		{name: "overflow4", left: math.MinInt64, right: -2, wantErr: true},
		{name: "overflowMinMax", left: math.MinInt64, right: math.MaxInt64, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.MulInt64(tt.left, tt.right)
			if (err != nil) != tt.wantErr {
				t.Errorf("MulInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MulInt64() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBEnv_String(t *testing.T) {
	env := util.DBEnv{
		Host:     "foohost",
		Port:     "1234",
		Database: "",
		User:     "",
		Pass:     "secret",
		SslMode:  "",
	}
	tests := []struct {
		name string
		s    string
	}{
		{name: "String()", s: env.String()},
		{name: "Sprintf(%v)", s: fmt.Sprintf("<nolint> %v", env)}, // avoid linter error.
		{name: "Sprintf(%+v)", s: fmt.Sprintf("%+v", env)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tt.s, "foohost")
			assert.Contains(t, tt.s, "1234")
			assert.NotContains(t, tt.s, "secret")
		})
	}
}

func TestGetResourceKindFromResourceID(t *testing.T) {
	tests := []struct {
		name    string
		resID   string
		want    inv_v1.ResourceKind
		wantErr bool
	}{
		{name: "ValidHostID", resID: "host-1234567", want: inv_v1.ResourceKind_RESOURCE_KIND_HOST, wantErr: false},
		{name: "InvalidResourceType", resID: "foo-1234567", want: inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED, wantErr: true},
		{name: "Empty", resID: "", want: inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.GetResourceKindFromResourceID(tt.resID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResourceKindFromResourceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "GetResourceKindFromResourceID(%v)", tt.resID)
		})
	}
}

func TestLookupDBEnv(t *testing.T) {
	setupValidDBEnv := func(t *testing.T) {
		t.Helper()
		t.Setenv("PGHOST", "foohost")
		t.Setenv("PGPORT", "4321")
		t.Setenv("PGDATABASE", "my_database_1")
		t.Setenv("PGUSER", "some_user")
		t.Setenv("PGPASSWORD", "pass1234")
		t.Setenv("PGSSLMODE", "disable")
		t.Setenv("PGHOST_RO", "foohost2")
		t.Setenv("PGPORT_RO", "4321")
		t.Setenv("PGDATABASE_RO", "my_database_1")
		t.Setenv("PGUSER_RO", "some_user")
		t.Setenv("PGPASSWORD_RO", "pass1234")
	}
	t.Run("MissingEnvVar", func(t *testing.T) {
		for _, envKey := range []string{
			"PGHOST",
			"PGPORT",
			"PGDATABASE",
			"PGUSER",
			"PGPASSWORD",
			"PGSSLMODE",
			"PGPORT_RO",
			"PGDATABASE_RO",
			"PGUSER_RO",
			"PGPASSWORD_RO",
		} {
			setupValidDBEnv(t)
			require.NoError(t, os.Unsetenv(envKey))
			_, _, err := util.LookupDBEnv()
			assert.Error(t, err)
		}
	})
	t.Run("MissingOtherROEnvVar", func(t *testing.T) {
		for _, envKey := range []string{
			"PGPORT_RO",
			"PGDATABASE_RO",
			"PGUSER_RO",
			"PGPASSWORD_RO",
		} {
			setupValidDBEnv(t)
			require.NoError(t, os.Unsetenv(envKey))
			_, _, err := util.LookupDBEnv()
			assert.Error(t, err)
		}
	})
	t.Run("MissingRO", func(t *testing.T) {
		// If RO host is missing, the RO DB URL won't be populated
		setupValidDBEnv(t)
		require.NoError(t, os.Unsetenv("PGHOST_RO"))
		got, gotRo, err := util.LookupDBEnv()
		require.NoError(t, err)
		assert.Equal(t, "foohost", got.Host)
		assert.Empty(t, gotRo)
	})
	t.Run("ValidEnv", func(t *testing.T) {
		setupValidDBEnv(t)
		got, gotRo, err := util.LookupDBEnv()
		require.NoError(t, err)
		require.Equal(t, "foohost", got.Host)
		require.Equal(t, "foohost2", gotRo.Host)
	})
}

func TestGetResourceFromType(t *testing.T) {
	t.Run("TestSomeKind", func(t *testing.T) {
		ipAddrRes, err := util.GetResourceFromKind(inv_v1.ResourceKind_RESOURCE_KIND_IPADDRESS)
		require.NoErrorf(t, err, errors.ErrorToStringWithDetails(err))
		switch tp := ipAddrRes.GetResource().(type) {
		case *inv_v1.Resource_Ipaddress:
		default:
			assert.Fail(t, "Wrong resource type", "got %#v but expected IPAddress", tp)
		}

		siteRes, err := util.GetResourceFromKind(inv_v1.ResourceKind_RESOURCE_KIND_SITE)
		require.NoErrorf(t, err, errors.ErrorToStringWithDetails(err))
		switch tp := siteRes.GetResource().(type) {
		case *inv_v1.Resource_Site:
		default:
			assert.Fail(t, "Wrong resource type", "got %#v but expected Site", tp)
		}
	})

	t.Run("TestAllResKind", func(t *testing.T) {
		for kindName, val := range inv_v1.ResourceKind_value {
			kind := inv_v1.ResourceKind(val)
			res, err := util.GetResourceFromKind(inv_v1.ResourceKind(val))
			if kind != inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED {
				require.NoError(t, err, errors.ErrorToStringWithDetails(err))
				assert.NotNilf(t, res, "%s doesn't translate to a Resource", kindName)
			} else {
				assert.Error(t, err)
				assert.Nil(t, res)
			}
		}
	})
}

func TestWrapResource(t *testing.T) {
	for _, val := range inv_v1.ResourceKind_value {
		kind := inv_v1.ResourceKind(val)
		if kind == inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED {
			continue
		}
		res, err := util.GetResourceFromKind(kind)
		require.NoError(t, err)

		un, err := util.UnwrapResource[proto.Message](res)
		require.NoError(t, err)

		wrapped, err := util.WrapResource(un)
		require.NoError(t, err)

		require.Equal(t, res, wrapped)
	}
}

func TestUnwrapResource(t *testing.T) {
	testCases := map[string]struct {
		in    *inv_v1.Resource
		expPM proto.Message
		valid bool
	}{
		"EmptyResource": {
			in:    &inv_v1.Resource{},
			valid: false,
		},
		"TestRegion": {
			in: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Region{
					Region: &location_v1.RegionResource{
						Name:         "TEST",
						ParentRegion: &location_v1.RegionResource{},
					},
				},
			},
			expPM: &location_v1.RegionResource{
				Name:         "TEST",
				ParentRegion: &location_v1.RegionResource{},
			},
			valid: true,
		},
		"TestIpaddress": {
			in: &inv_v1.Resource{
				Resource: &inv_v1.Resource_Ipaddress{
					Ipaddress: &network_v1.IPAddressResource{
						Address: "10.0.0.1/24",
					},
				},
			},
			expPM: &network_v1.IPAddressResource{
				Address: "10.0.0.1/24",
			},
			valid: true,
		},
	}
	for tName, tc := range testCases {
		t.Run(tName, func(t *testing.T) {
			actual, err := util.UnwrapResource[proto.Message](tc.in)
			if !tc.valid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.expPM, actual); !eq {
					t.Errorf("Data not equal: %v", diff)
				}
			}
		})
	}

	for kindName, val := range inv_v1.ResourceKind_value {
		kind := inv_v1.ResourceKind(val)
		if kind == inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED {
			continue
		}
		res, err := util.GetResourceFromKind(inv_v1.ResourceKind(val))
		t.Run("TestAllResKind_"+kindName, func(t *testing.T) {
			require.NoError(t, err)
			actual, err := util.UnwrapResource[proto.Message](res)
			require.NoError(t, err)
			expected, err := util.GetSetResource(res)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})
	}
}

type testCaseGetSpecificResourceList[T proto.Message] struct {
	in    []*inv_v1.Resource
	exp   []T
	valid bool
}

func Test_GetSpecificResourceList(t *testing.T) {
	// Invalid
	t.Run("InvalidMultipleResourcesTypes", func(t *testing.T) {
		tc := testCaseGetSpecificResourceList[*location_v1.RegionResource]{
			in: []*inv_v1.Resource{
				{Resource: &inv_v1.Resource_Region{}}, {Resource: &inv_v1.Resource_Instance{}},
			},
			valid: false,
		}
		runGetSpecificResourceList[*location_v1.RegionResource](t, tc)
	})
	t.Run("InvalidWrongConcreteType", func(t *testing.T) {
		tc := testCaseGetSpecificResourceList[*location_v1.RegionResource]{
			in:    []*inv_v1.Resource{{Resource: &inv_v1.Resource_Instance{}}, {Resource: &inv_v1.Resource_Instance{}}},
			valid: false,
		}
		runGetSpecificResourceList[*location_v1.RegionResource](t, tc)
	})

	// Valid
	t.Run("RegionNil", func(t *testing.T) {
		tc := testCaseGetSpecificResourceList[*location_v1.RegionResource]{
			in: []*inv_v1.Resource{
				{Resource: &inv_v1.Resource_Region{}},
				{Resource: &inv_v1.Resource_Region{}},
			},
			exp:   []*location_v1.RegionResource{nil, nil},
			valid: true,
		}
		runGetSpecificResourceList[*location_v1.RegionResource](t, tc)
	})
	t.Run("RegionValid", func(t *testing.T) {
		region1 := &location_v1.RegionResource{ResourceId: "region-12345678"}
		region2 := &location_v1.RegionResource{ResourceId: "region-12345678"}
		tc := testCaseGetSpecificResourceList[*location_v1.RegionResource]{
			in: []*inv_v1.Resource{
				{Resource: &inv_v1.Resource_Region{Region: region1}},
				{Resource: &inv_v1.Resource_Region{Region: region2}},
			},
			exp:   []*location_v1.RegionResource{region1, region2},
			valid: true,
		}
		runGetSpecificResourceList[*location_v1.RegionResource](t, tc)
	})
	t.Run("SingleSchedule", func(t *testing.T) {
		sSched1 := &schedule_v1.SingleScheduleResource{ResourceId: "singlesche-12345678"}
		sSched2 := &schedule_v1.SingleScheduleResource{ResourceId: "singlesche-12345678"}
		tc := testCaseGetSpecificResourceList[*schedule_v1.SingleScheduleResource]{
			in: []*inv_v1.Resource{
				{Resource: &inv_v1.Resource_Singleschedule{Singleschedule: sSched1}},
				{Resource: &inv_v1.Resource_Singleschedule{Singleschedule: sSched2}},
			},
			exp:   []*schedule_v1.SingleScheduleResource{sSched1, sSched2},
			valid: true,
		}
		runGetSpecificResourceList[*schedule_v1.SingleScheduleResource](t, tc)
	})
}

func runGetSpecificResourceList[T proto.Message](t *testing.T, tc testCaseGetSpecificResourceList[T]) {
	t.Helper()

	resources, err := util.GetSpecificResourceList[T](tc.in)
	if !tc.valid {
		assert.Error(t, err)
		assert.Nil(t, resources)
	} else {
		require.NoErrorf(t, err, errors.ErrorToStringWithDetails(err))
		require.Len(t, resources, len(tc.exp))
		assert.Equal(t, tc.exp, resources)
	}
}

func Test_BuildFieldMaskFromFields(t *testing.T) {
	t.Run("ValidStrings", func(t *testing.T) {
		v1 := "value1"
		v2 := "VALUE2"
		v3 := "123.45$%^"
		actual := util.BuildNestedFieldMaskFromFields(v1, v2, v3)
		assert.Equal(t, "value1.VALUE2.123.45$%^", actual)
	})
	t.Run("Empty", func(t *testing.T) {
		actual := util.BuildNestedFieldMaskFromFields()
		assert.Empty(t, actual)
	})
	t.Run("Single", func(t *testing.T) {
		value := "test"
		actual := util.BuildNestedFieldMaskFromFields(value)
		assert.Equal(t, value, actual)
	})
}

func TestGetResourceIdFromResource(t *testing.T) {
	for _, val := range inv_v1.ResourceKind_value {
		kind := inv_v1.ResourceKind(val)
		if kind == inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED {
			continue
		}

		res, err := util.GetResourceFromKind(kind)
		require.NoError(t, err)

		prefix := util.ResourceKindToPrefix(kind)
		require.NotEqual(t, prefix, util.ResourcePrefixUnspecified)

		// Set res ID?
		t.Run(string(prefix), func(t *testing.T) {
			resID, err := util.GetResourceIDFromResource(res)
			assert.NoError(t, err)
			assert.Equalf(t, "", resID, "GetResourceIdFromResource(%v)", res)
		})
	}
}

func TestCheckListOutputIsSingular(t *testing.T) {
	newResource := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Instance{
			Instance: &computev1.InstanceResource{
				ResourceId: "inst-1234567",
			},
		},
	}
	resources := make([]*inv_v1.Resource, 0)

	// Should return a NotFound error - empty list of resources
	err1 := util.CheckListOutputIsSingular(resources)
	require.Error(t, err1)
	assert.Equal(t, grpc_status.Convert(err1).Code(), codes.NotFound)

	resources = append(resources, newResource)
	// Should NOT return an error - list contains exactly one resource
	err2 := util.CheckListOutputIsSingular(resources)
	require.NoError(t, err2)

	resources = append(resources, newResource)
	// Should return an Internal error - list contains more than one resource
	err3 := util.CheckListOutputIsSingular(resources)
	require.Error(t, err3)
	assert.Equal(t, grpc_status.Convert(err3).Code(), codes.Internal)
}

func Test_ConvertRawUUIDToInventoryUUID(t *testing.T) {
	uuidIncorrect := "57ED598C4B9411EE806C3A7C7693AAC3"

	convertedUUID, err := util.ConvertRawUUIDToInventoryUUID(uuidIncorrect)
	require.NoError(t, err)
	assert.Equal(t, convertedUUID, "57ed598c-4b94-11ee-806c-3a7c7693aac3")

	_, err = util.ConvertRawUUIDToInventoryUUID("bla-bla")
	require.Error(t, err)
}

func Test_ConvertInventoryUUIDToRawUUID(t *testing.T) {
	uuidCorrect := "071995e7-264a-4adb-9669-5eff59951bf1"

	convertedIncorrectUUID := util.ConvertInventoryUUIDToLenovoUUID(uuidCorrect)
	assert.Equal(t, convertedIncorrectUUID, "071995E7264A4ADB96695EFF59951BF1")
}

func Test_ConvertBothWays(t *testing.T) {
	uuidIncorrect := "57ED598C4B9411EE806C3A7C7693AAC3"

	convertedUUID, err := util.ConvertRawUUIDToInventoryUUID(uuidIncorrect)
	require.NoError(t, err)
	assert.Equal(t, convertedUUID, "57ed598c-4b94-11ee-806c-3a7c7693aac3")

	uuidIncorrectBack := util.ConvertInventoryUUIDToLenovoUUID(uuidIncorrect)
	assert.Equal(t, uuidIncorrect, uuidIncorrectBack)

	// now let's do the same but with correc UUID
	uuidCorrect := "071995e7-264a-4adb-9669-5eff59951bf1"

	convertedIncorrectUUID := util.ConvertInventoryUUIDToLenovoUUID(uuidCorrect)
	assert.Equal(t, convertedIncorrectUUID, "071995E7264A4ADB96695EFF59951BF1")

	uuidCorrectBack, err := util.ConvertRawUUIDToInventoryUUID(convertedIncorrectUUID)
	require.NoError(t, err)
	assert.Equal(t, uuidCorrect, uuidCorrectBack)
}

func TestFilterBuilder(t *testing.T) {
	tcs := []struct {
		actual   string
		expected string
	}{
		{
			// evaluate single clause
			actual:   filters.ValEq("a", "b")(),
			expected: `a = "b"`,
		},
		{
			// evaluate single clause
			actual:   filters.ValEq("a", 100)(),
			expected: `a = 100`,
		},
		{
			actual:   filters.NewBuilder().And(filters.ValEq("a", "b")).Build(),
			expected: `a = "b"`,
		},
		{
			actual:   filters.NewBuilderWith(filters.ValEq("a", "b")).Build(),
			expected: `a = "b"`,
		},
		{
			actual:   filters.NewBuilderWith(filters.ValDotValEq("a", "b", "c")).Build(),
			expected: `a.b = "c"`,
		},
		{
			actual:   filters.NewBuilderWith(filters.ValDotValDotValEq("a", "b", "c", "d")).Build(),
			expected: `a.b.c = "d"`,
		},
		{
			actual:   filters.NewBuilderWith(filters.ValDotValEq("a", "b", "c")).And(filters.NotHas("d")).Build(),
			expected: `a.b = "c" AND NOT has(d)`,
		},
		{
			actual:   filters.NewBuilderWith(filters.ValDotValEq("a", "b", "c")).Or(filters.NotHas("d")).Build(),
			expected: `a.b = "c" OR NOT has(d)`,
		},
		{
			actual: filters.NewBuilderWith(filters.ValEq("a", "b")).
				And(filters.ValDotValEq("c", "d", "e")).Or(filters.NotHas("f")).Build(),
			expected: `a = "b" AND c.d = "e" OR NOT has(f)`,
		},
		{
			actual:   filters.NewBuilder().And(filters.ValEq("a", "b")).And(filters.ValEq("c", "d")).Build(),
			expected: `a = "b" AND c = "d"`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.actual)
		})
	}
}

func TestPaginatorWith(t *testing.T) {
	type tc struct {
		input         []int
		offset, limit int
		expRes        []int
		expNext       bool
		expTotal      int
	}

	tcs := []tc{
		{
			input:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			offset:   0,
			limit:    2,
			expRes:   []int{0, 1},
			expNext:  true,
			expTotal: 10,
		},
		{
			input:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			offset:   2,
			limit:    2,
			expRes:   []int{2, 3},
			expNext:  true,
			expTotal: 10,
		},
		{
			input:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			offset:   0,
			limit:    10,
			expRes:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			expNext:  false,
			expTotal: 10,
		},
		{
			input:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			offset:   9,
			limit:    10,
			expRes:   []int{9},
			expNext:  false,
			expTotal: 10,
		},
		{
			input:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			offset:   10,
			limit:    10,
			expRes:   []int{},
			expNext:  false,
			expTotal: 10,
		},
	}

	for _, test := range tcs {
		t.Run(fmt.Sprintf("in: %v, o: %d, l: %d", test.input, test.offset, test.limit), func(t *testing.T) {
			sut := paginator.NewPaginator[int](test.offset, test.limit)
			res, next, total := sut.Apply(test.input)
			require.Equal(t, test.expRes, res)
			require.Equal(t, test.expNext, next)
			require.Equal(t, test.expTotal, total)
		})
	}
}

func TestResourceIDComparator(t *testing.T) {
	rs1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Repeatedschedule{Repeatedschedule: &schedule_v1.RepeatedScheduleResource{ResourceId: "rs1"}},
	}
	rs2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Repeatedschedule{Repeatedschedule: &schedule_v1.RepeatedScheduleResource{ResourceId: "rs2"}},
	}
	rs3 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Repeatedschedule{Repeatedschedule: &schedule_v1.RepeatedScheduleResource{ResourceId: "rs3"}},
	}
	ss1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Singleschedule{Singleschedule: &schedule_v1.SingleScheduleResource{ResourceId: "ss1"}},
	}
	ss2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Singleschedule{Singleschedule: &schedule_v1.SingleScheduleResource{ResourceId: "ss2"}},
	}
	ss3 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_Singleschedule{Singleschedule: &schedule_v1.SingleScheduleResource{ResourceId: "ss3"}},
	}
	tg1 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_TelemetryGroup{TelemetryGroup: &telemetry_v1.TelemetryGroupResource{ResourceId: "tg1"}},
	}
	tg2 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_TelemetryGroup{TelemetryGroup: &telemetry_v1.TelemetryGroupResource{ResourceId: "tg2"}},
	}
	tg3 := &inv_v1.Resource{
		Resource: &inv_v1.Resource_TelemetryGroup{TelemetryGroup: &telemetry_v1.TelemetryGroupResource{ResourceId: "tg3"}},
	}
	nothing := &inv_v1.Resource{}
	var justANil *inv_v1.Resource

	tests := []struct {
		name     string
		input    []*inv_v1.Resource
		expected []*inv_v1.Resource
	}{
		{
			name:     "sort repeated schedules",
			input:    []*inv_v1.Resource{rs3, rs2, rs1},
			expected: []*inv_v1.Resource{rs1, rs2, rs3},
		},
		{
			name:     "sort single schedules",
			input:    []*inv_v1.Resource{ss3, ss2, ss1},
			expected: []*inv_v1.Resource{ss1, ss2, ss3},
		},
		{
			name:     "sort telemetry groups",
			input:    []*inv_v1.Resource{tg3, tg2, tg1},
			expected: []*inv_v1.Resource{tg1, tg2, tg3},
		},
		{
			name:     "sort single and repeated schedules",
			input:    []*inv_v1.Resource{rs3, ss2, ss3, rs1},
			expected: []*inv_v1.Resource{rs1, rs3, ss2, ss3},
		},
		{
			name:     "sort different resources",
			input:    []*inv_v1.Resource{tg2, ss1, tg1, rs3, rs2, tg3, ss3, rs1, ss2},
			expected: []*inv_v1.Resource{rs1, rs2, rs3, ss1, ss2, ss3, tg1, tg2, tg3},
		},
		{
			name:     "sort nothings",
			input:    []*inv_v1.Resource{nothing, nothing, nothing},
			expected: []*inv_v1.Resource{nothing, nothing, nothing},
		},
		{
			name:     "sort somethings and nothings",
			input:    []*inv_v1.Resource{ss3, nothing, rs1},
			expected: []*inv_v1.Resource{rs1, ss3, nothing},
		},
		{
			name:     "sort nils",
			input:    []*inv_v1.Resource{justANil, justANil},
			expected: []*inv_v1.Resource{justANil, justANil},
		},
		{
			name:     "sort somethings and nils",
			input:    []*inv_v1.Resource{ss3, justANil, rs1},
			expected: []*inv_v1.Resource{rs1, ss3, justANil},
		},
		{
			name:     "sort somethings and nothing and a nil",
			input:    []*inv_v1.Resource{nothing, ss3, justANil, rs1},
			expected: []*inv_v1.Resource{rs1, ss3, nothing, justANil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedOrder := toString(tt.expected)
			sort.Slice(comparator.ResourceIDAscComparator(tt.input))
			providedOrder := toString(tt.input)
			assert.Equal(t, expectedOrder, providedOrder,
				"returned order:\n\t (%s)\nis other than expected:\n\t(%s)", providedOrder, expectedOrder)
		})
	}
}

func toString(array []*inv_v1.Resource) string {
	return strings.Join(collections.MapSlice[*inv_v1.Resource, string](array, func(r *inv_v1.Resource) string {
		return r.String()
	}), "\n\t")
}

func TestGetResourceKeyFromResource(t *testing.T) {
	for _, val := range inv_v1.ResourceKind_value {
		kind := inv_v1.ResourceKind(val)
		if kind == inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED {
			continue
		}

		res, err := util.GetResourceFromKind(kind)
		require.NoError(t, err)

		tenantID, resourceID, err := util.GetResourceKeyFromResource(res)
		require.NoError(t, err)
		assert.Equal(t, "", tenantID)
		assert.Equal(t, "", resourceID)
	}
}

func TestConcatMapValuesSorted(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected string
	}{
		{
			name:     "NilMap",
			input:    nil,
			expected: "",
		},
		{
			name:     "EmptyMap",
			input:    map[string]string{},
			expected: "",
		},
		{
			name: "SingleKeyValue",
			input: map[string]string{
				"a": "foo",
			},
			expected: "\"foo\"",
		},
		{
			name: "SingleKeyEmptyValue",
			input: map[string]string{
				"a": "",
			},
			expected: "",
		},
		{
			name: "MultipleKeysSorted",
			input: map[string]string{
				"b": "bar",
				"a": "foo",
				"c": "baz",
			},
			expected: "\"foo\\x1fbar\\x1fbaz\"",
		},
		{
			name: "KeysWithEmptyValue",
			input: map[string]string{
				"a": "",
				"b": "bar",
			},
			expected: "\"bar\"",
		},
		{
			name: "KeysWithInterspersedEmptyValue",
			input: map[string]string{
				"a": "foo",
				"b": "",
				"c": "baz",
			},
			expected: "\"foo\\x1fbaz\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := util.ConcatMapValuesSorted(tt.input, "\x1f")
			if got != tt.expected {
				t.Errorf("ConcatMapValuesSorted() = %q, want %q", got, tt.expected)
			}
		})
	}
}
