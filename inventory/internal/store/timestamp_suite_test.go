// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/store"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	location_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/location/v1"
	network_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/network/v1"
	osv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/os/v1"
	ou_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/ou/v1"
	provider_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/provider/v1"
	remoteaccessv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/remoteaccess/v1"
	schedule_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	telemetry_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/telemetry/v1"
	tenantv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/tenant/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

type tsSuiteTestRes interface {
	GetResourceId() string
	GetTenantId() string
	GetCreatedAt() string
	GetUpdatedAt() string
}

func CreateTimestampSuite(
	t *testing.T,
	createRes func(*inv_testing.InvResourceDAO) (resource tsSuiteTestRes),
	emptyRes func() proto.Message,
	updateRes func() (proto.Message, []string),
) *TimestampTestSuite {
	t.Helper()
	return &TimestampTestSuite{
		createResource: createRes,
		emptyResource:  emptyRes,
		updateResource: updateRes,
	}
}

type TimestampTestSuite struct {
	suite.Suite
	apiClient         client.TenantAwareInventoryClient
	createResource    func(*inv_testing.InvResourceDAO) (resource tsSuiteTestRes)
	emptyResource     func() proto.Message
	updateResource    func() (proto.Message, []string)
	dao               *inv_testing.InvResourceDAO
	creationTime      time.Time
	creationCreatedAt time.Time
	creationUpdatedAt time.Time
	tenantID          string
	resourceID        string
}

func (tts *TimestampTestSuite) SetupSuite() {
	tts.dao = inv_testing.NewInvResourceDAOOrFail(tts.T())
	tts.apiClient = inv_testing.GetClient(tts.T(), inv_testing.APIClient).GetTenantAwareInventoryClient()

	// Create resource in setup, to have resource during update test.
	tts.creationTime = time.Now()
	time.Sleep(10 * time.Millisecond)

	res := tts.createResource(tts.dao)

	tts.resourceID = res.GetResourceId()
	tts.tenantID = res.GetTenantId()
	err := error(nil)
	tts.creationCreatedAt, err = time.Parse(store.ISO8601Format, res.GetCreatedAt())
	tts.Require().NoError(err)
	tts.creationUpdatedAt, err = time.Parse(store.ISO8601Format, res.GetUpdatedAt())
	tts.Require().NoError(err)
}

func (tts *TimestampTestSuite) TestCreationTimestamps() {
	tts.Require().NotEmpty(tts.creationTime)
	tts.Require().NotEmpty(tts.creationCreatedAt)
	tts.Require().NotEmpty(tts.creationUpdatedAt)

	tts.Assert().True(tts.creationCreatedAt.After(tts.creationTime))
	tts.Assert().True(tts.creationUpdatedAt.After(tts.creationTime))
	tts.Assert().Zero(tts.creationCreatedAt.Compare(tts.creationUpdatedAt))
}

func (tts *TimestampTestSuite) TestUpdateTimestamps() {
	var upRes proto.Message
	var fields []string

	if tts.updateResource == nil {
		upRes = proto.Clone(tts.emptyResource())
		// Here we use reflection to set the created_at field in the upRes
		upResProto := upRes.ProtoReflect()
		var fieldName string
		switch upResProto.Descriptor().Name() {
		// Special case for host sub resources, where there is no names.
		// Other special cases are handled via the updateResource func.
		case "HostgpuResource", "HostnicResource", "HoststorageResource", "HostusbResource":
			fieldName = "device_name"
		default:
			fieldName = "name"
		}
		field := upResProto.Descriptor().Fields().ByName(protoreflect.Name(fieldName))

		upResProto.Set(field, protoreflect.ValueOfString("TEST"))
		fields = append(fields, fieldName)
	} else {
		upRes, fields = tts.updateResource()
	}
	updateRes, err := util.WrapResource(upRes)
	tts.Require().NoError(err)

	updatedResource, err := tts.apiClient.Update(
		context.TODO(),
		tts.tenantID,
		tts.resourceID,
		&fieldmaskpb.FieldMask{Paths: fields},
		updateRes,
	)
	tts.Require().NoError(err)

	// Here we use reflection to get the created_at and updated_at fields in the uupRes
	unwrappedUpRes, err := util.UnwrapResource[proto.Message](updatedResource)
	tts.Require().NoError(err)
	updatedResProto := unwrappedUpRes.ProtoReflect()
	fieldCreatedAt := updatedResProto.Descriptor().Fields().ByName("created_at")
	fieldUpdatedAt := updatedResProto.Descriptor().Fields().ByName("updated_at")
	createdAt := updatedResProto.Get(fieldCreatedAt).String()
	updatedAt := updatedResProto.Get(fieldUpdatedAt).String()

	tts.Require().NotEmpty(createdAt)
	tts.Require().NotEmpty(updatedAt)

	updatedAtParsed, err := time.Parse(store.ISO8601Format, updatedAt)
	tts.Require().NoError(err)
	createdAtParsed, err := time.Parse(store.ISO8601Format, createdAt)
	tts.Require().NoError(err)

	tts.Assert().Zero(createdAtParsed.Compare(tts.creationCreatedAt))
	tts.Assert().NotEqual(tts.creationUpdatedAt, updatedAt)
	tts.Assert().True(updatedAtParsed.Compare(tts.creationCreatedAt) > 0)
}

func (tts *TimestampTestSuite) TestFailUpdateUpdatedAtTimestamp() {
	var otherFields []string
	var upRes proto.Message
	if tts.updateResource == nil {
		upRes = proto.Clone(tts.emptyResource())
	} else {
		// If we have updateResource use it as base.
		upRes, otherFields = tts.updateResource()
	}

	upResProto := upRes.ProtoReflect()
	fieldCreatedAt := upResProto.Descriptor().Fields().ByName("created_at")
	fieldUpdatedAt := upResProto.Descriptor().Fields().ByName("updated_at")
	upResProto.Set(fieldCreatedAt, protoreflect.ValueOfString("2006-01-02T15:04:05.999Z"))
	upResProto.Set(fieldUpdatedAt, protoreflect.ValueOfString("2006-01-02T15:04:05.999Z"))
	wrappedUpRes, err := util.WrapResource(upRes)
	tts.Require().NoError(err)

	_, err = tts.apiClient.Update(
		context.TODO(),
		tts.tenantID,
		tts.resourceID,
		&fieldmaskpb.FieldMask{Paths: append(otherFields, "created_at")},
		wrappedUpRes,
	)
	tts.Require().Error(err, "error shall be thrown when updateRequest.resource.createdAt != ''")
	tts.Require().Equalf(codes.InvalidArgument, status.Code(err),
		"invalidArgument error is expected whilst returned is: %v", err)
	tts.Require().ErrorContains(err, "field created_at is immutable")

	_, err = tts.apiClient.Update(
		context.TODO(),
		tts.tenantID,
		tts.resourceID,
		&fieldmaskpb.FieldMask{Paths: append(otherFields, "updated_at")},
		wrappedUpRes,
	)
	tts.Require().Error(err, "error shall be thrown when updateRequest.resource.updatedAt != ''")
	tts.Require().Equalf(codes.InvalidArgument, status.Code(err),
		"invalidArgument error is expected whilst returned is: %v", err)
	tts.Require().ErrorContains(err, "field updated_at is immutable")
}

// Invoke the test suite for all resources available in inventory.
//
//nolint:funlen // Matrix style with in-function definition of tests specs.
func TestTimestamp(t *testing.T) {
	resourceTests := []struct {
		name       string
		createFunc func(*inv_testing.InvResourceDAO) tsSuiteTestRes
		emptyFunc  func() proto.Message
		updateFunc func() (proto.Message, []string)
	}{
		{
			name: "Host",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateHost(t, tenantIDOne)
			},
			emptyFunc: func() proto.Message {
				return &computev1.HostResource{}
			},
		},
		{
			name: "HostGPU",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateHostGPU(t, tenantIDOne, dao.CreateHost(t, tenantIDOne))
			},
			emptyFunc: func() proto.Message {
				return &computev1.HostgpuResource{}
			},
		},
		{
			name: "HostNic",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateHostNic(t, tenantIDOne, dao.CreateHost(t, tenantIDOne))
			},
			emptyFunc: func() proto.Message {
				return &computev1.HostnicResource{}
			},
		},
		{
			name: "HostStorage",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateHostStorage(t, tenantIDOne, dao.CreateHost(t, tenantIDOne))
			},
			emptyFunc: func() proto.Message {
				return &computev1.HoststorageResource{}
			},
		},
		{
			name: "HostUSB",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateHostUsb(t, tenantIDOne, dao.CreateHost(t, tenantIDOne))
			},
			emptyFunc: func() proto.Message {
				return &computev1.HostusbResource{}
			},
		},
		{
			name: "Instance",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateInstance(t, tenantIDOne, nil, dao.CreateOs(t, tenantIDOne))
			},
			emptyFunc: func() proto.Message {
				return &computev1.InstanceResource{}
			},
		},
		{
			name: "Workload",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateWorkload(t, tenantIDOne)
			},
			emptyFunc: func() proto.Message {
				return &computev1.WorkloadResource{}
			},
		},
		{
			name: "Region",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateRegion(t, tenantIDOne)
			},
			emptyFunc: func() proto.Message {
				return &location_v1.RegionResource{}
			},
		},
		{
			name: "Site",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateSite(t, tenantIDOne)
			},
			emptyFunc: func() proto.Message {
				return &location_v1.SiteResource{}
			},
		},
		{
			name: "Endpoint",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateEndpoint(t, tenantIDOne, nil)
			},
			emptyFunc: func() proto.Message {
				return &network_v1.EndpointResource{}
			},
		},
		{
			name: "Netlink",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateNetLink(t, tenantIDOne, true)
			},
			emptyFunc: func() proto.Message {
				return &network_v1.NetlinkResource{}
			},
		},
		{
			name: "NetworkSegment",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateNetworkSegment(t, tenantIDOne, "", dao.CreateSite(t, tenantIDOne), 12, true)
			},
			emptyFunc: func() proto.Message {
				return &network_v1.NetworkSegment{}
			},
		},
		{
			name: "IPAddress",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				host := dao.CreateHost(t, tenantIDOne)
				hostNic := dao.CreateHostNic(t, tenantIDOne, host)
				return dao.CreateIPAddress(t, tenantIDOne, hostNic, true)
			},
			emptyFunc: func() proto.Message {
				return &network_v1.IPAddressResource{}
			},
			updateFunc: func() (proto.Message, []string) {
				return &network_v1.IPAddressResource{
						StatusDetail: "TEST",
					},
					[]string{network_v1.IPAddressResourceFieldStatusDetail}
			},
		},
		{
			name: "Os",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateOs(t, tenantIDOne)
			},
			emptyFunc: func() proto.Message {
				return &osv1.OperatingSystemResource{}
			},
		},
		{
			name: "Ou",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateOu(t, tenantIDOne)
			},
			emptyFunc: func() proto.Message {
				return &ou_v1.OuResource{}
			},
		},
		{
			name: "Provider",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateProvider(t, tenantIDOne, "providerTest",
					inv_testing.ProviderKind(provider_v1.ProviderKind_PROVIDER_KIND_BAREMETAL))
			},
			emptyFunc: func() proto.Message {
				return &provider_v1.ProviderResource{}
			},
		},
		{
			name: "RemoteAccessConfiguration",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateRemoteAccessConfiguration(t, tenantIDOne)
			},
			emptyFunc: func() proto.Message {
				return &remoteaccessv1.RemoteAccessConfiguration{}
			},
			updateFunc: func() (proto.Message, []string) {
				return &remoteaccessv1.RemoteAccessConfiguration{
						User: "TEST",
					},
					[]string{remoteaccessv1.RemoteAccessConfigurationFieldUser}
			},
		},
		{
			name: "SingleSchedule",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateSingleSchedule(t, tenantIDOne)
			},
			emptyFunc: func() proto.Message {
				return &schedule_v1.SingleScheduleResource{}
			},
		},
		{
			name: "RepeatedSchedule",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateRepeatedSchedule(t, tenantIDOne)
			},
			emptyFunc: func() proto.Message {
				return &schedule_v1.RepeatedScheduleResource{}
			},
			updateFunc: func() (proto.Message, []string) {
				return &schedule_v1.RepeatedScheduleResource{
						CronMinutes:  "*",
						CronHours:    "*",
						CronDayMonth: "*",
						CronMonth:    "*",
						CronDayWeek:  "*",
					},
					[]string{
						schedule_v1.RepeatedScheduleResourceFieldCronMinutes,
						schedule_v1.RepeatedScheduleResourceFieldCronHours,
						schedule_v1.RepeatedScheduleResourceFieldCronDayMonth,
						schedule_v1.RepeatedScheduleResourceFieldCronMonth,
						schedule_v1.RepeatedScheduleResourceFieldCronDayWeek,
					}
			},
		},
		{
			name: "TelemetryGroup",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateTelemetryGroupLogs(t, tenantIDOne, true)
			},
			emptyFunc: func() proto.Message {
				return &telemetry_v1.TelemetryGroupResource{}
			},
		},
		{
			name: "TelemetryProfile",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				site := dao.CreateSite(t, tenantIDOne)
				tpGroup := dao.CreateTelemetryGroupLogs(t, tenantIDOne, true)
				return dao.CreateTelemetryProfile(t, tenantIDOne, inv_testing.TelemetryProfileTarget(site), tpGroup, true)
			},
			emptyFunc: func() proto.Message {
				return &telemetry_v1.TelemetryProfile{}
			},
			updateFunc: func() (proto.Message, []string) {
				return &telemetry_v1.TelemetryProfile{
						LogLevel: telemetry_v1.SeverityLevel_SEVERITY_LEVEL_CRITICAL,
					},
					[]string{telemetry_v1.TelemetryProfileFieldLogLevel}
			},
		},
		{
			name: "Tenant",
			createFunc: func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateTenantWithOpts(t, tenantIDOne, true,
					inv_testing.TenantDesiredState(tenantv1.TenantState_TENANT_STATE_CREATED))
			},
			emptyFunc: func() proto.Message {
				return &tenantv1.Tenant{}
			},
			updateFunc: func() (proto.Message, []string) {
				return &tenantv1.Tenant{
						WatcherOsmanager: true,
					},
					[]string{tenantv1.TenantFieldWatcherOsmanager}
			},
		},
	}

	for _, rt := range resourceTests {
		t.Run(rt.name, func(t *testing.T) {
			suite.Run(t, CreateTimestampSuite(t, rt.createFunc, rt.emptyFunc, rt.updateFunc))
		})
	}

	t.Run("WorkloadMember", func(t *testing.T) {
		dao := inv_testing.NewInvResourceDAOOrFail(t)
		instance1 := dao.CreateInstance(t, tenantIDOne, nil, dao.CreateOs(t, tenantIDOne))
		instance2 := dao.CreateInstance(t, tenantIDOne, nil, dao.CreateOs(t, tenantIDOne))
		suite.Run(t, CreateTimestampSuite(t,
			func(dao *inv_testing.InvResourceDAO) tsSuiteTestRes {
				return dao.CreateWorkloadMember(t, tenantIDOne, dao.CreateWorkload(t, tenantIDOne), instance1)
			},
			func() proto.Message {
				return &computev1.WorkloadMember{}
			},
			func() (proto.Message, []string) {
				return &computev1.WorkloadMember{
						Instance: instance2,
					},
					[]string{computev1.WorkloadMemberEdgeInstance}
			},
		))
	})
}
