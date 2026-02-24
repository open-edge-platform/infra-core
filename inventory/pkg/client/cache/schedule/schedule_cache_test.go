// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package schedule_test

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	schedulev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	sc "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client/cache/schedule"
	inv_testing "github.com/open-edge-platform/infra-core/inventory/v2/pkg/testing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/function"
)

const (
	tenant1 = "11111111-1111-1111-1111-111111111111"
	tenant2 = "22222222-2222-2222-2222-222222222222"
)

var (
	emptyString           = "null"
	timeNow               = time.Now()
	timeNowString         = fmt.Sprint(timeNow.UTC().Unix())
	zeroTimeString        = fmt.Sprint(time.Unix(0, 0).UTC().Unix())
	time1DinFutureString  = fmt.Sprint(timeNow.UTC().Unix() + 86400)
	time1HinFutureInt     = uint64(timeNow.UTC().Unix() + 3600)
	time2HinFutureInt     = uint64(timeNow.UTC().Unix() + 7200)
	time2DinFutureInt     = uint64(timeNow.UTC().Unix() + 172800)
	testString            = "test"
	cronAny               = "*"
	validRepeatedSchedAny = schedulev1.RepeatedScheduleResource{
		CronMinutes:     cronAny,
		CronHours:       cronAny,
		CronDayMonth:    cronAny,
		CronMonth:       cronAny,
		CronDayWeek:     cronAny,
		DurationSeconds: 120,
		TenantId:        tenant1,
	}
	validRepeatedSchedDayOfWeek = schedulev1.RepeatedScheduleResource{
		CronMinutes:     cronAny,
		CronHours:       cronAny,
		CronDayMonth:    cronAny,
		CronMonth:       cronAny,
		CronDayWeek:     fmt.Sprintf("%d", timeNow.Weekday()),
		DurationSeconds: 120,
		TenantId:        tenant1,
	}
	validRepeatedSchedMonth = schedulev1.RepeatedScheduleResource{
		CronMinutes:     cronAny,
		CronHours:       cronAny,
		CronDayMonth:    cronAny,
		CronMonth:       fmt.Sprintf("%d", timeNow.Month()),
		CronDayWeek:     cronAny,
		DurationSeconds: 120,
		TenantId:        tenant1,
	}
	validRepeatedSchedHour = schedulev1.RepeatedScheduleResource{
		CronMinutes:     cronAny,
		CronHours:       fmt.Sprintf("%d", timeNow.Hour()),
		CronDayMonth:    cronAny,
		CronMonth:       cronAny,
		CronDayWeek:     cronAny,
		DurationSeconds: 120,
		TenantId:        tenant1,
	}
	validRepeatedSchedNow = schedulev1.RepeatedScheduleResource{
		CronMinutes:     fmt.Sprintf("%d", timeNow.Minute()),
		CronHours:       fmt.Sprintf("%d", timeNow.Hour()),
		CronDayMonth:    cronAny,
		CronMonth:       fmt.Sprintf("%d", timeNow.Month()),
		CronDayWeek:     cronAny,
		DurationSeconds: 120,
		TenantId:        tenant1,
	}
	// will be scheduled next hour.
	expiredRepeatedSched = schedulev1.RepeatedScheduleResource{
		CronMinutes:     cronAny,
		CronHours:       fmt.Sprintf("%d", (timeNow.Hour()+1)%24),
		CronDayMonth:    cronAny,
		CronMonth:       cronAny,
		CronDayWeek:     cronAny,
		DurationSeconds: 120,
		TenantId:        tenant1,
	}

	// note - all of these are testable relative to 1 day in future.
	validSingleSched1 = schedulev1.SingleScheduleResource{
		StartSeconds: time1HinFutureInt,
		TenantId:     tenant1,
	}
	validSingleSched2 = schedulev1.SingleScheduleResource{
		StartSeconds: time2HinFutureInt,
		TenantId:     tenant1,
	}
	validSingleSched3 = schedulev1.SingleScheduleResource{
		StartSeconds: time1HinFutureInt,
		EndSeconds:   time2DinFutureInt,
		TenantId:     tenant1,
	}
	expiredSingleSched = schedulev1.SingleScheduleResource{
		StartSeconds: time1HinFutureInt,
		EndSeconds:   time2HinFutureInt,
		TenantId:     tenant1,
	}
)

// Define this flag in order to call all tests with the same parameters.
var _ = flag.String(
	"policyBundle",
	"/rego/policy_bundle.tar.gz",
	"Path of policy rego file",
)

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(wd))))

	policyPath := projectRoot + "/out"
	migrationsDir := projectRoot + "/out"

	inv_testing.StartTestingEnvironment(policyPath, "", migrationsDir)
	run := m.Run() // run all tests
	inv_testing.StopTestingEnvironment()

	os.Exit(run)
}

func TestNewScheduleCache(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("Fail to create invclient", func(t *testing.T) {
		scheduleCache, err := sc.NewScheduleCacheClientWithOptions(ctx,
			sc.WithInventoryAddress(""),
		)
		assert.Nil(t, scheduleCache)
		assert.Error(t, err)
	})

	t.Run("Success", func(t *testing.T) {
		scheduleCache, err := sc.NewScheduleCacheClientWithOptions(ctx,
			sc.WithInventoryAddress("bufconn"),
			sc.WithDialOption(
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return inv_testing.BufconnLis.Dial() }),
			),
			sc.WithEnableTracing(false),
			sc.WithRegisterMaxElapsedTime(10*time.Second),
		)
		assert.NotNil(t, scheduleCache)
		assert.NoError(t, err)
		scheduleCache.Stop()
	})

	t.Run("Success - aggregate options", func(t *testing.T) {
		scheduleCache, err := sc.NewScheduleCacheClientWithOptions(ctx,
			sc.WithOptions(sc.Options{
				InventoryAddress: "bufconn",
				DialOptions: []grpc.DialOption{
					grpc.WithContextDialer(
						func(context.Context, string) (net.Conn, error) { return inv_testing.BufconnLis.Dial() },
					),
				},
			}),
		)
		assert.NotNil(t, scheduleCache)
		assert.NoError(t, err)
		scheduleCache.Stop()
	})

	t.Run("Success - load at start", func(t *testing.T) {
		dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
		dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE))
		dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_CLUSTER_UPDATE))

		scheduleCache, err := sc.NewScheduleCacheClientWithOptions(ctx,
			sc.WithInventoryAddress("bufconn"),
			sc.WithDialOption(
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return inv_testing.BufconnLis.Dial() }),
			),
		)
		assert.NotNil(t, scheduleCache)
		assert.NoError(t, err)
		defer scheduleCache.Stop()

		sScheds, _, _, err := scheduleCache.GetSchedules(
			inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE, tenant1, nil, 0, 5)
		require.NoError(t, err)
		assert.Len(t, sScheds, 2)
		rScheds, _, _, err := scheduleCache.GetSchedules(
			inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE, tenant1, nil, 0, 5)
		require.NoError(t, err)
		assert.Len(t, rScheds, 1)
	})
}

//nolint:funlen //just a test
func TestScheduleCache(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Reduce periodic refresh for easier testing
	periodicCacheRefreshTest := 5 * time.Second
	sc.PeriodicCacheRefresh = &periodicCacheRefreshTest

	scheduleCache, err := sc.NewScheduleCacheClientWithOptions(ctx,
		sc.WithInventoryAddress("bufconn"),
		sc.WithDialOption(
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return inv_testing.BufconnLis.Dial() }),
		),
	)
	assert.NotNil(t, scheduleCache)
	assert.NoError(t, err)
	defer scheduleCache.Stop()

	hT1 := dao.CreateHost(t, tenant1)
	hT2 := dao.CreateHost(t, tenant2)
	ssrT1 := dao.CreateSingleScheduleNoCleanup(
		t, tenant1, inv_testing.SSRTargetHost(hT1), inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	ssrT2 := dao.CreateSingleScheduleNoCleanup(
		t, tenant2, inv_testing.SSRTargetHost(hT2), inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	time.Sleep(1 * time.Second)

	assertSSRInCache(t, scheduleCache, tenant1, ssrT1)
	assertSSRInCache(t, scheduleCache, tenant2, ssrT2)

	// Tenant Isolation guaranteed
	schedT1, _, _, err := scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE, tenant1, nil, 0, 100)
	require.NoError(t, err)
	collections.ForEach[*inv_v1.Resource](schedT1, func(res *inv_v1.Resource) {
		assert.NotEqual(t, ssrT2.GetResourceId(), res.GetSingleschedule().GetResourceId())
	})
	schedT2, _, _, err := scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE, tenant2, nil, 0, 100)
	require.NoError(t, err)
	collections.ForEach[*inv_v1.Resource](schedT2, func(res *inv_v1.Resource) {
		assert.NotEqual(t, ssrT1.GetResourceId(), res.GetSingleschedule().GetResourceId())
	})

	// invalidate so that cache is out of sync with Inv
	scheduleCache.InvalidateCache(tenant1, ssrT1.GetResourceId(), inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED)
	scheduleCache.InvalidateCache(tenant2, ssrT2.GetResourceId(), inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED)

	// wait for periodic refresh
	time.Sleep(10 * time.Second)

	assertSSRInCache(t, scheduleCache, tenant1, ssrT1)
	assertSSRInCache(t, scheduleCache, tenant2, ssrT2)

	rsrT1 := dao.CreateRepeatedScheduleNoCleanup(
		t, tenant1, inv_testing.RSRTargetHost(hT1), inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE))
	rsrT2 := dao.CreateRepeatedScheduleNoCleanup(
		t, tenant2, inv_testing.RSRTargetHost(hT2), inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE))
	time.Sleep(1 * time.Second)

	assertRSRInCache(t, scheduleCache, tenant1, rsrT1)
	assertRSRInCache(t, scheduleCache, tenant2, rsrT2)

	// Tenant Isolation guaranteed
	schedT1, _, _, err = scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE, tenant1, nil, 0, 100)
	require.NoError(t, err)
	collections.ForEach[*inv_v1.Resource](schedT1, func(res *inv_v1.Resource) {
		assert.NotEqual(t, rsrT2.GetResourceId(), res.GetSingleschedule().GetResourceId())
	})
	schedT2, _, _, err = scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE, tenant2, nil, 0, 100)
	require.NoError(t, err)
	collections.ForEach[*inv_v1.Resource](schedT2, func(res *inv_v1.Resource) {
		assert.NotEqual(t, rsrT1.GetResourceId(), res.GetSingleschedule().GetResourceId())
	})

	dao.DeleteResource(t, tenant1, rsrT1.GetResourceId())
	dao.DeleteResource(t, tenant2, rsrT2.GetResourceId())
	hT1ID := hT1.GetResourceId()
	hT2ID := hT2.GetResourceId()

	assert.EventuallyWithT(
		t,
		func(collect *assert.CollectT) {
			schedules, _, _, getErr := scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE,
				tenant1, []*sc.Filters{new(sc.Filters).Add(sc.HasHostID(&hT1ID))}, 0, 5)
			require.NoError(collect, getErr)
			assert.Len(collect, schedules, 0)

			schedules, _, _, getErr = scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE,
				tenant2, []*sc.Filters{new(sc.Filters).Add(sc.HasHostID(&hT2ID))}, 0, 5)
			require.NoError(collect, getErr)
			assert.Len(collect, schedules, 0)
		},
		10*time.Second, // this shall be value of unexported schedule.periodicCacheRefreshSec
		time.Second)

	sScheds, _, _, err := scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE,
		tenant1, []*sc.Filters{new(sc.Filters).Add(sc.HasHostID(&hT1ID))}, 0, 5)
	require.NoError(t, err)
	assert.Len(t, sScheds, 1)

	sScheds, _, _, err = scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE,
		tenant2, []*sc.Filters{new(sc.Filters).Add(sc.HasHostID(&hT2ID))}, 0, 5)
	require.NoError(t, err)
	assert.Len(t, sScheds, 1)

	dao.DeleteResource(t, tenant1, ssrT1.GetResourceId())
	dao.DeleteResource(t, tenant2, ssrT2.GetResourceId())

	assert.EventuallyWithT(
		t,
		func(collect *assert.CollectT) {
			schedules, _, _, getErr := scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE,
				tenant1, []*sc.Filters{new(sc.Filters).Add(sc.HasHostID(&hT1ID))}, 0, 5)
			require.NoError(collect, getErr)
			assert.Len(collect, schedules, 0)

			schedules, _, _, getErr = scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE,
				tenant2, []*sc.Filters{new(sc.Filters).Add(sc.HasHostID(&hT2ID))}, 0, 5)
			require.NoError(collect, getErr)
			assert.Len(collect, schedules, 0)
		},
		10*time.Second, // this shall be value of unexported schedule.periodicCacheRefreshSec
		time.Second)
}

func TestScheduleCacheInvalidate(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	scheduleCache, err := sc.NewScheduleCacheClientWithOptions(ctx,
		sc.WithInventoryAddress("bufconn"),
		sc.WithDialOption(
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return inv_testing.BufconnLis.Dial() }),
		),
	)
	assert.NotNil(t, scheduleCache)
	assert.NoError(t, err)
	defer scheduleCache.Stop()

	hT1 := dao.CreateHost(t, tenant1)
	hT2 := dao.CreateHost(t, tenant2)
	ssrT1 := dao.CreateSingleSchedule(
		t, tenant1, inv_testing.SSRTargetHost(hT1), inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	ssrT2 := dao.CreateSingleSchedule(
		t, tenant2, inv_testing.SSRTargetHost(hT2), inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	time.Sleep(1 * time.Second)

	assertSSRInCache(t, scheduleCache, tenant1, ssrT1)
	assertSSRInCache(t, scheduleCache, tenant2, ssrT2)

	rsrT1 := dao.CreateRepeatedSchedule(
		t, tenant1, inv_testing.RSRTargetHost(hT1), inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE))
	rsrT2 := dao.CreateRepeatedSchedule(
		t, tenant2, inv_testing.RSRTargetHost(hT2), inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_OS_UPDATE))
	time.Sleep(1 * time.Second)

	assertRSRInCache(t, scheduleCache, tenant1, rsrT1)
	assertRSRInCache(t, scheduleCache, tenant2, rsrT2)

	// no resource ID, should be no-op
	scheduleCache.InvalidateCache(tenant1, "", inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED)
	scheduleCache.InvalidateCache(tenant2, "", inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED)
	resIDT1 := rsrT1.GetResourceId()
	resIDT2 := rsrT2.GetResourceId()
	// Nonexistent resID in a tenant should be no-op
	scheduleCache.InvalidateCache(tenant1, resIDT2, inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED)
	scheduleCache.InvalidateCache(tenant2, resIDT1, inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED)
	// no tenant ID, should be no-op
	scheduleCache.InvalidateCache("", resIDT1, inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED)

	assertRSRInCache(t, scheduleCache, tenant1, rsrT1)
	assertRSRInCache(t, scheduleCache, tenant2, rsrT2)

	scheduleCache.InvalidateCache(tenant1, ssrT1.GetResourceId(), inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED)
	scheduleCache.InvalidateCache(tenant2, rsrT2.GetResourceId(), inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED)

	assertRSRInCache(t, scheduleCache, tenant1, rsrT1)
	assertSSRInCache(t, scheduleCache, tenant2, ssrT2)
	assertRSRNotInCache(t, scheduleCache, tenant2, rsrT2.GetResourceId())
	assertSSRNotInCache(t, scheduleCache, tenant1, ssrT1.GetResourceId())

	scheduleCache.InvalidateCache(tenant2, ssrT2.GetResourceId(), inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED)
	scheduleCache.InvalidateCache(tenant1, rsrT1.GetResourceId(), inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED)

	hT1ID := hT1.GetResourceId()
	hT2ID := hT2.GetResourceId()
	sScheds, _, _, err := scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE,
		tenant1, []*sc.Filters{new(sc.Filters).Add(sc.HasHostID(&hT1ID))}, 0, 5)
	require.NoError(t, err)
	assert.Len(t, sScheds, 0)
	rScheds, _, _, err := scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE,
		tenant1, []*sc.Filters{new(sc.Filters).Add(sc.HasHostID(&hT1ID))}, 0, 5)
	require.NoError(t, err)
	assert.Len(t, rScheds, 0)
	sScheds, _, _, err = scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE,
		tenant2, []*sc.Filters{new(sc.Filters).Add(sc.HasHostID(&hT2ID))}, 0, 5)
	require.NoError(t, err)
	assert.Len(t, sScheds, 0)
	rScheds, _, _, err = scheduleCache.GetSchedules(inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE,
		tenant2, []*sc.Filters{new(sc.Filters).Add(sc.HasHostID(&hT2ID))}, 0, 5)
	require.NoError(t, err)
	assert.Len(t, rScheds, 0)
}

func Test_ScheduleCache_LoadAllFromInventory(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	scheduleCache := sc.NewScheduleCacheClient(
		inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient(),
	)
	nScheds := sc.BatchSize + sc.BatchSize/2
	// Create nScheds single and repeated schedules, keep one to check its value in the cache
	sSchedT1 := dao.CreateSingleSchedule(t, tenant1,
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSchedT1 := dao.CreateRepeatedSchedule(t, tenant1,
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSchedT2 := dao.CreateSingleSchedule(t, tenant2,
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSchedT2 := dao.CreateRepeatedSchedule(t, tenant2,
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	for i := 1; i < nScheds; i++ {
		dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
		dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
		dao.CreateSingleSchedule(t, tenant2, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
		dao.CreateRepeatedSchedule(t, tenant2, inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	}

	scheduleCache.LoadAllSchedulesFromInv()
	assert.Equal(t, nScheds*4, len(scheduleCache.TestGetAllSchedules()))

	assertSSRInCache(t, scheduleCache, tenant1, sSchedT1)
	assertSSRInCache(t, scheduleCache, tenant2, sSchedT2)
	assertRSRInCache(t, scheduleCache, tenant1, rSchedT1)
	assertRSRInCache(t, scheduleCache, tenant2, rSchedT2)
}

func Test_ScheduleCache_GetSingleSchedule(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	scheduleCache := sc.NewScheduleCacheClient(
		inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient(),
	)
	host := dao.CreateHost(t, tenant1)
	site := dao.CreateSite(t, tenant1)
	sSched1 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSched2 := dao.CreateSingleSchedule(
		t, tenant1, inv_testing.SSRTargetHost(host), inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSched2.Relation = &schedulev1.SingleScheduleResource_TargetHost{
		TargetHost: host,
	}
	sSched3 := dao.CreateSingleSchedule(
		t, tenant1, inv_testing.SSRTargetSite(site), inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSched3.Relation = &schedulev1.SingleScheduleResource_TargetSite{
		TargetSite: site,
	}
	scheduleCache.LoadAllSchedulesFromInv()

	testCases := map[string]struct {
		resID string
		valid bool
		exp   *schedulev1.SingleScheduleResource
	}{
		"ValidSched": {
			resID: sSched1.ResourceId,
			valid: true,
			exp:   sSched1,
		},
		"ValidSchedWithHost": {
			resID: sSched2.ResourceId,
			valid: true,
			exp:   sSched2,
		},
		"ValidSchedWithSite": {
			resID: sSched3.ResourceId,
			valid: true,
			exp:   sSched3,
		},
		"EmptyID": {
			resID: "",
			valid: false,
		},
		"InvalidID": {
			resID: "qwe",
			valid: false,
		},
	}

	for tName, tc := range testCases {
		t.Run(tName, func(t *testing.T) {
			res, err := scheduleCache.GetSingleSchedule(tenant1, tc.resID)
			if tc.valid {
				require.NoError(t, err)
				require.NotNil(t, res)
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.exp, res); !eq {
					t.Errorf("wrong single schedule in cache: %v", diff)
				}
			} else {
				require.Error(t, err)
				assert.Equal(t, codes.NotFound, status.Code(err))
			}
		})
	}
}

func Test_ScheduleCache_GetRepeatedSchedule(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	scheduleCache := sc.NewScheduleCacheClient(
		inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient(),
	)
	host := dao.CreateHost(t, tenant1)
	site := dao.CreateSite(t, tenant1)
	rSched1 := dao.CreateRepeatedSchedule(t, tenant1,
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSched2 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRTargetHost(host),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSched2.Relation = &schedulev1.RepeatedScheduleResource_TargetHost{
		TargetHost: host,
	}
	rSched3 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRTargetSite(site),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSched3.Relation = &schedulev1.RepeatedScheduleResource_TargetSite{
		TargetSite: site,
	}
	scheduleCache.LoadAllSchedulesFromInv()

	testCases := map[string]struct {
		resID string
		valid bool
		exp   *schedulev1.RepeatedScheduleResource
	}{
		"ValidSched": {
			resID: rSched1.ResourceId,
			valid: true,
			exp:   rSched1,
		},
		"ValidSchedWithHost": {
			resID: rSched2.ResourceId,
			valid: true,
			exp:   rSched2,
		},
		"ValidSchedWithSite": {
			resID: rSched3.ResourceId,
			valid: true,
			exp:   rSched3,
		},
		"EmptyID": {
			resID: "",
			valid: false,
		},
		"InvalidID": {
			resID: "qwe",
			valid: false,
		},
	}

	for tName, tc := range testCases {
		t.Run(tName, func(t *testing.T) {
			res, err := scheduleCache.GetRepeatedSchedule(tenant1, tc.resID)
			if tc.valid {
				require.NoError(t, err)
				require.NotNil(t, res)
				if eq, diff := inv_testing.ProtoEqualOrDiff(tc.exp, res); !eq {
					t.Errorf("wrong repeated schedule in cache: %v", diff)
				}
			} else {
				require.Error(t, err)
				assert.Equal(t, codes.NotFound, status.Code(err))
			}
		})
	}
}

//nolint:funlen // no need to extract code
func Test_HScheduleCache_Pagination(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	// Verifies that the pagination of resources in the cache works as expected
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Init a schedule cache with the testing client
	scheduleCache := sc.NewScheduleCacheClient(
		inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient(),
	)
	hScheduleCache, err := sc.NewHScheduleCacheClient(scheduleCache)

	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
		dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	}
	scheduleCache.LoadAllSchedulesFromInv()

	sScheds, hasNext, totLen, err := hScheduleCache.GetAllSingleSchedules(ctx, tenant1, 0, 1)
	require.NoError(t, err)
	assert.True(t, hasNext)
	assert.Equal(t, 1, len(sScheds))
	assert.Equal(t, 10, totLen)

	sScheds, hasNext, totLen, err = hScheduleCache.GetAllSingleSchedules(ctx, tenant1, 9, 1)
	require.NoError(t, err)
	assert.False(t, hasNext)
	assert.Equal(t, 1, len(sScheds))
	assert.Equal(t, 10, totLen)

	sScheds, hasNext, totLen, err = hScheduleCache.GetAllSingleSchedules(ctx, tenant1, 0, 10)
	require.NoError(t, err)
	assert.False(t, hasNext)
	assert.Equal(t, 10, len(sScheds))
	assert.Equal(t, 10, totLen)

	sScheds, hasNext, totLen, err = hScheduleCache.GetAllSingleSchedules(ctx, tenant1, 8, 10)
	require.NoError(t, err)
	assert.False(t, hasNext)
	assert.Equal(t, 2, len(sScheds))
	assert.Equal(t, 10, totLen)

	sScheds, hasNext, totLen, err = hScheduleCache.GetAllSingleSchedules(ctx, tenant1, 10, 1)
	require.NoError(t, err)
	assert.False(t, hasNext)
	assert.Equal(t, 0, len(sScheds))
	assert.Equal(t, 10, totLen)

	// pagination + filtering
	dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRStart(time1HinFutureInt), inv_testing.SSREnd(0),
	)
	dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRStart(time1HinFutureInt), inv_testing.SSREnd(0),
	)
	scheduleCache.LoadAllSchedulesFromInv()

	sScheds, hasNext, totLen, err = hScheduleCache.GetSingleSchedules(ctx, tenant1, 0, 10,
		new(sc.Filters).Add(sc.FilterByTS(&time1DinFutureString)))
	require.NoError(t, err)
	assert.False(t, hasNext)
	assert.Equal(t, 2, len(sScheds))
	assert.Equal(t, 2, totLen)

	sScheds, hasNext, totLen, err = hScheduleCache.GetSingleSchedules(ctx, tenant1, 0, 1,
		new(sc.Filters).Add(sc.FilterByTS(&time1DinFutureString)))
	require.NoError(t, err)
	assert.True(t, hasNext)
	assert.Equal(t, 1, len(sScheds))
	assert.Equal(t, 2, totLen)

	var rScheds []*schedulev1.RepeatedScheduleResource
	rScheds, hasNext, totLen, err = hScheduleCache.GetAllRepeatedSchedules(ctx, tenant1, 0, 1)
	require.NoError(t, err)
	assert.True(t, hasNext)
	assert.Equal(t, 1, len(rScheds))
	assert.Equal(t, 10, totLen)

	rScheds, hasNext, totLen, err = hScheduleCache.GetAllRepeatedSchedules(ctx, tenant1, 9, 1)
	require.NoError(t, err)
	assert.False(t, hasNext)
	assert.Equal(t, 1, len(rScheds))
	assert.Equal(t, 10, totLen)

	rScheds, hasNext, totLen, err = hScheduleCache.GetAllRepeatedSchedules(ctx, tenant1, 0, 10)
	require.NoError(t, err)
	assert.False(t, hasNext)
	assert.Equal(t, 10, len(rScheds))
	assert.Equal(t, 10, totLen)

	rScheds, hasNext, totLen, err = hScheduleCache.GetAllRepeatedSchedules(ctx, tenant1, 10, 1)
	require.NoError(t, err)
	assert.False(t, hasNext)
	assert.Equal(t, 0, len(rScheds))
	assert.Equal(t, 10, totLen)

	// pagination + filtering
	dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.RSRDayWeek(fmt.Sprintf("%d", timeNow.Weekday())), inv_testing.RSRDayMonth(cronAny),
		inv_testing.RSRMonth(cronAny), inv_testing.RSRHours(cronAny), inv_testing.RSRMinutes(cronAny),
	)
	dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.RSRDayWeek(fmt.Sprintf("%d", timeNow.Weekday())), inv_testing.RSRDayMonth(cronAny),
		inv_testing.RSRMonth(cronAny), inv_testing.RSRHours(cronAny), inv_testing.RSRMinutes(cronAny),
	)
	scheduleCache.LoadAllSchedulesFromInv()

	rScheds, hasNext, totLen, err = hScheduleCache.GetRepeatedSchedules(ctx, tenant1, 0, 10,
		new(sc.Filters).Add(sc.FilterByTS(&timeNowString)))
	require.NoError(t, err)
	assert.False(t, hasNext)
	assert.Equal(t, 2, len(rScheds))
	assert.Equal(t, 2, totLen)

	rScheds, hasNext, totLen, err = hScheduleCache.GetRepeatedSchedules(ctx, tenant1, 0, 1,
		new(sc.Filters).Add(sc.FilterByTS(&timeNowString)))
	require.NoError(t, err)
	assert.True(t, hasNext)
	assert.Equal(t, 1, len(rScheds))
	assert.Equal(t, 2, totLen)
}

//nolint:funlen //just a test
func Test_HScheduleCache_GetSingleSchedules_WithAutogeneratedFilters(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	// Init a schedule cache with the testing client
	scheduleCache := sc.NewScheduleCacheClient(
		inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient(),
	)
	hScheduleCache, err := sc.NewHScheduleCacheClient(scheduleCache)

	require.NoError(t, err)
	host := dao.CreateHost(t, tenant1)
	site := dao.CreateSite(t, tenant1)
	region := dao.CreateRegion(t, tenant1)
	sSched1 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSched2 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRTargetHost(host))
	sSched2.Relation = &schedulev1.SingleScheduleResource_TargetHost{
		TargetHost: host,
	}
	sSched3 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRTargetHost(host))
	sSched3.Relation = &schedulev1.SingleScheduleResource_TargetHost{
		TargetHost: host,
	}
	sSched4 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRTargetSite(site))
	sSched4.Relation = &schedulev1.SingleScheduleResource_TargetSite{
		TargetSite: site,
	}
	sSched5 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRTargetSite(site))
	sSched5.Relation = &schedulev1.SingleScheduleResource_TargetSite{
		TargetSite: site,
	}
	sSched6 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRRegion(region),
	)
	sSched6.Relation = &schedulev1.SingleScheduleResource_TargetRegion{
		TargetRegion: region,
	}
	sSched7 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRRegion(region),
	)
	sSched7.Relation = &schedulev1.SingleScheduleResource_TargetRegion{
		TargetRegion: region,
	}
	sSched8 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRStart(time1HinFutureInt), inv_testing.SSREnd(0),
	)
	scheduleCache.LoadAllSchedulesFromInv()

	testCases := map[string]struct {
		hostID     *string
		siteID     *string
		regionID   *string
		timestamp  *string
		valid      bool
		exp        []*schedulev1.SingleScheduleResource
		expErrCode codes.Code
	}{
		"All": {
			valid: true,
			exp:   []*schedulev1.SingleScheduleResource{sSched1, sSched2, sSched3, sSched4, sSched5, sSched6, sSched7, sSched8},
		},
		"ListByHost": {
			hostID: &host.ResourceId,
			valid:  true,
			exp:    []*schedulev1.SingleScheduleResource{sSched2, sSched3},
		},
		"ListBySite": {
			siteID: &site.ResourceId,
			valid:  true,
			exp:    []*schedulev1.SingleScheduleResource{sSched4, sSched5},
		},
		"ListByEmptyHostSiteAndRegion": {
			hostID:   &emptyString,
			siteID:   &emptyString,
			regionID: &emptyString,
			exp:      []*schedulev1.SingleScheduleResource{sSched1, sSched8},
			valid:    true,
		},
		"EmptySiteId": {
			siteID: &emptyString,
			exp:    []*schedulev1.SingleScheduleResource{sSched1, sSched2, sSched3, sSched6, sSched7, sSched8},
			valid:  true,
		},
		"EmptyHostId": {
			hostID: &emptyString,
			exp:    []*schedulev1.SingleScheduleResource{sSched1, sSched4, sSched5, sSched6, sSched7, sSched8},
			valid:  true,
		},
		"EmptyRegionId": {
			regionID: &emptyString,
			exp:      []*schedulev1.SingleScheduleResource{sSched1, sSched2, sSched3, sSched4, sSched5, sSched8},
			valid:    true,
		},
		"SetNonEmptyHostAndSite": {
			hostID:     &host.ResourceId,
			siteID:     &site.ResourceId,
			valid:      false,
			expErrCode: codes.InvalidArgument,
		},
		"SetNonEmptyRegionAndSite": {
			regionID:   &region.ResourceId,
			siteID:     &site.ResourceId,
			valid:      false,
			expErrCode: codes.InvalidArgument,
		},
		"TimeSoon": {
			timestamp: &time1DinFutureString,
			valid:     true,
			exp:       []*schedulev1.SingleScheduleResource{sSched8},
		},
	}

	for tName, tc := range testCases {
		t.Run(tName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			filters := new(sc.Filters)
			buildTestFilter(filters, tc.hostID, tc.siteID, tc.regionID)
			filters.Add(sc.FilterByTS(tc.timestamp))

			res, hasNext, totLen, err := hScheduleCache.GetSingleSchedules(ctx, tenant1, 0, 100, filters)
			if !tc.valid {
				require.Error(t, err)
				assert.Equal(t, tc.expErrCode, status.Code(err))
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.False(t, hasNext)
				assert.Equal(t, len(tc.exp), totLen)
				require.Equal(t, len(tc.exp), len(res))
				inv_testing.OrderByResourceID(res)
				inv_testing.OrderByResourceID(tc.exp)
				for i := 0; i < len(tc.exp); i++ {
					if eq, diff := inv_testing.ProtoEqualOrDiff(tc.exp[i], res[i]); !eq {
						t.Errorf("wrong single schedule in cache: %v", diff)
					}
				}
			}
		})
	}
}

//nolint:funlen //just a test
func Test_HScheduleCache_GetSingleSchedules_WithFilters(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	// Init a schedule cache with the testing client
	scheduleCache := sc.NewScheduleCacheClient(
		inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient(),
	)
	hScheduleCache, err := sc.NewHScheduleCacheClient(scheduleCache)

	require.NoError(t, err)
	host := dao.CreateHost(t, tenant1)
	site := dao.CreateSite(t, tenant1)
	region := dao.CreateRegion(t, tenant1)
	sSched1 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSched2 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRTargetHost(host))
	sSched2.Relation = &schedulev1.SingleScheduleResource_TargetHost{
		TargetHost: host,
	}
	sSched3 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRTargetHost(host),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSched3.Relation = &schedulev1.SingleScheduleResource_TargetHost{
		TargetHost: host,
	}
	sSched4 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRTargetSite(site),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSched4.Relation = &schedulev1.SingleScheduleResource_TargetSite{
		TargetSite: site,
	}
	sSched5 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRTargetSite(site),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSched5.Relation = &schedulev1.SingleScheduleResource_TargetSite{
		TargetSite: site,
	}
	sSched6 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRRegion(region),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSched6.Relation = &schedulev1.SingleScheduleResource_TargetRegion{
		TargetRegion: region,
	}
	sSched7 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRRegion(region),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSched7.Relation = &schedulev1.SingleScheduleResource_TargetRegion{
		TargetRegion: region,
	}
	sSched8 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRStart(time1HinFutureInt), inv_testing.SSREnd(0),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	scheduleCache.LoadAllSchedulesFromInv()

	testCases := map[string]struct {
		filters    *sc.Filters
		timestamp  *string
		valid      bool
		exp        []*schedulev1.SingleScheduleResource
		expErrCode codes.Code
	}{
		"All#1": {
			filters: sc.DefaultFilter,
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSched1, sSched2, sSched3, sSched4, sSched5, sSched6, sSched7, sSched8},
		},
		"All#2": {
			valid: true,
			exp:   []*schedulev1.SingleScheduleResource{sSched1, sSched2, sSched3, sSched4, sSched5, sSched6, sSched7, sSched8},
		},
		"All#3": {
			filters: new(sc.Filters),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSched1, sSched2, sSched3, sSched4, sSched5, sSched6, sSched7, sSched8},
		},
		"ListByHost": {
			filters: new(sc.Filters).Add(sc.HasHostID(&host.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSched2, sSched3},
		},
		"ListBySite": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&site.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSched4, sSched5},
		},
		"ListByEmptyHostSiteAndRegion": {
			filters: new(sc.Filters).
				Add(sc.HasHostID(&emptyString)).
				Add(sc.HasRegionID(&emptyString)).
				Add(sc.HasSiteID(&emptyString)),

			exp:   []*schedulev1.SingleScheduleResource{sSched1, sSched8},
			valid: true,
		},
		"EmptySiteId": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&emptyString)),
			exp:     []*schedulev1.SingleScheduleResource{sSched1, sSched2, sSched3, sSched6, sSched7, sSched8},
			valid:   true,
		},
		"EmptyHostId": {
			filters: new(sc.Filters).Add(sc.HasHostID(&emptyString)),
			exp:     []*schedulev1.SingleScheduleResource{sSched1, sSched4, sSched5, sSched6, sSched7, sSched8},
			valid:   true,
		},
		"EmptyRegionId": {
			filters: new(sc.Filters).Add(sc.HasRegionID(&emptyString)),
			exp:     []*schedulev1.SingleScheduleResource{sSched1, sSched2, sSched3, sSched4, sSched5, sSched8},
			valid:   true,
		},
		"SetNonEmptyHostAndSite": {
			filters: new(sc.Filters).
				Add(sc.HasHostID(&host.ResourceId)).
				Add(sc.HasSiteID(&site.ResourceId)),
			valid:      false,
			expErrCode: codes.InvalidArgument,
		},
		"SetNonEmptyRegionAndSite": {
			filters: new(sc.Filters).
				Add(sc.HasRegionID(&region.ResourceId)).
				Add(sc.HasSiteID(&site.ResourceId)),
			valid:      false,
			expErrCode: codes.InvalidArgument,
		},
		"TimeSoon": {
			filters: new(sc.Filters).Add(sc.FilterByTS(&time1DinFutureString)),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSched8},
		},
	}

	for tName, tc := range testCases {
		t.Run(tName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			res, hasNext, totLen, err := hScheduleCache.GetSingleSchedules(ctx, tenant1, 0, 100, tc.filters)
			if !tc.valid {
				require.Error(t, err)
				assert.Equal(t, tc.expErrCode, status.Code(err))
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.False(t, hasNext)
				assert.Equal(t, len(tc.exp), totLen)
				require.Equal(t, len(tc.exp), len(res))
				inv_testing.OrderByResourceID(res)
				inv_testing.OrderByResourceID(tc.exp)
				for i := 0; i < len(tc.exp); i++ {
					if eq, diff := inv_testing.ProtoEqualOrDiff(tc.exp[i], res[i]); !eq {
						t.Errorf("wrong single schedule in cache: %v", diff)
					}
				}
			}
		})
	}
}

//nolint:funlen // it's a test
func Test_HScheduleCache_GetRepeatedSchedulesFilters(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	// Init a schedule cache with the testing client
	scheduleCache := sc.NewScheduleCacheClient(
		inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient(),
	)
	hScheduleCache, err := sc.NewHScheduleCacheClient(scheduleCache)

	require.NoError(t, err)
	host := dao.CreateHost(t, tenant1)
	site := dao.CreateSite(t, tenant1)
	region := dao.CreateRegion(t, tenant1)
	rSched1 := dao.CreateRepeatedSchedule(t, tenant1,
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSched2 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRTargetHost(host),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSched2.Relation = &schedulev1.RepeatedScheduleResource_TargetHost{
		TargetHost: host,
	}
	rSched3 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRTargetHost(host),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSched3.Relation = &schedulev1.RepeatedScheduleResource_TargetHost{
		TargetHost: host,
	}
	rSched4 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRTargetSite(site),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSched4.Relation = &schedulev1.RepeatedScheduleResource_TargetSite{
		TargetSite: site,
	}
	rSched5 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRTargetSite(site),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSched5.Relation = &schedulev1.RepeatedScheduleResource_TargetSite{
		TargetSite: site,
	}
	rSched6 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRRegion(region),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSched6.Relation = &schedulev1.RepeatedScheduleResource_TargetRegion{
		TargetRegion: region,
	}
	rSched7 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRRegion(region),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSched7.Relation = &schedulev1.RepeatedScheduleResource_TargetRegion{
		TargetRegion: region,
	}
	rSched8 := dao.CreateRepeatedSchedule(t, tenant1,
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.RSRDayWeek(fmt.Sprintf("%d", timeNow.Weekday())), inv_testing.RSRDayMonth(cronAny),
		inv_testing.RSRMonth(cronAny), inv_testing.RSRHours(cronAny), inv_testing.RSRMinutes(cronAny),
	)

	scheduleCache.LoadAllSchedulesFromInv()

	testCases := map[string]struct {
		hostID     *string
		siteID     *string
		regionID   *string
		timestamp  *string
		valid      bool
		exp        []*schedulev1.RepeatedScheduleResource
		expErrCode codes.Code
	}{
		// TODO: this behavior is a bit weird, I would expect that by passing nil to host and site I get only the schedule
		//  without any host or site associated.
		"All": {
			valid: true,
			exp:   []*schedulev1.RepeatedScheduleResource{rSched1, rSched2, rSched3, rSched4, rSched5, rSched6, rSched7, rSched8},
		},
		"ListByHost": {
			hostID: &host.ResourceId,
			valid:  true,
			exp:    []*schedulev1.RepeatedScheduleResource{rSched2, rSched3},
		},
		"ListBySite": {
			siteID: &site.ResourceId,
			valid:  true,
			exp:    []*schedulev1.RepeatedScheduleResource{rSched4, rSched5},
		},
		"ListByRegion": {
			regionID: &region.ResourceId,
			valid:    true,
			exp:      []*schedulev1.RepeatedScheduleResource{rSched6, rSched7},
		},
		"ListByEmptyHostSiteAndRegion": {
			hostID:   &emptyString,
			siteID:   &emptyString,
			regionID: &emptyString,
			exp:      []*schedulev1.RepeatedScheduleResource{rSched1, rSched8},
			valid:    true,
		},
		"EmptySiteId": {
			siteID: &emptyString,
			valid:  true,
			exp:    []*schedulev1.RepeatedScheduleResource{rSched1, rSched2, rSched3, rSched6, rSched7, rSched8},
		},
		"EmptyHostId": {
			hostID: &emptyString,
			valid:  true,
			exp:    []*schedulev1.RepeatedScheduleResource{rSched1, rSched4, rSched5, rSched6, rSched7, rSched8},
		},
		"EmptyRegionId": {
			regionID: &emptyString,
			valid:    true,
			exp:      []*schedulev1.RepeatedScheduleResource{rSched1, rSched2, rSched3, rSched4, rSched5, rSched8},
		},
		"SetNonEmptyHostAndSite": {
			hostID:     &host.ResourceId,
			siteID:     &site.ResourceId,
			valid:      false,
			expErrCode: codes.InvalidArgument,
		},
		"SetNonEmptyHostAndRegion": {
			hostID:     &host.ResourceId,
			regionID:   &region.ResourceId,
			valid:      false,
			expErrCode: codes.InvalidArgument,
		},
		"TimeNow": {
			timestamp: &timeNowString,
			valid:     true,
			exp:       []*schedulev1.RepeatedScheduleResource{rSched8},
		},
	}

	for tName, tc := range testCases {
		t.Run(tName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			filters := new(sc.Filters)
			buildTestFilter(filters, tc.hostID, tc.siteID, tc.regionID)
			filters.Add(sc.FilterByTS(tc.timestamp))

			res, hasNext, totLen, err := hScheduleCache.GetRepeatedSchedules(ctx, tenant1, 0, 100, filters)
			if !tc.valid {
				require.Error(t, err)
				assert.Equal(t, tc.expErrCode, status.Code(err))
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.False(t, hasNext)
				assert.Equal(t, len(tc.exp), totLen)
				require.Equal(t, len(tc.exp), len(res))
				inv_testing.OrderByResourceID(res)
				inv_testing.OrderByResourceID(tc.exp)
				for i := 0; i < len(tc.exp); i++ {
					if eq, diff := inv_testing.ProtoEqualOrDiff(tc.exp[i], res[i]); !eq {
						t.Errorf("wrong repeated schedule in cache: %v", diff)
					}
				}
			}
		})
	}
}

// Test does not support duplicate in the expected results.
//
//nolint:funlen // it's a test
func Test_FilterByTS_Single_Sched(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	scheduleCache := sc.NewScheduleCacheClient(
		inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient(),
	)
	hScheduleCache, err := sc.NewHScheduleCacheClient(scheduleCache)
	require.NoError(t, err)

	testcases := map[string]struct {
		timeWant       *string
		schedResources []*schedulev1.SingleScheduleResource
		expect         []*schedulev1.SingleScheduleResource
		fail           bool
	}{
		"oneSched": {
			&time1DinFutureString,
			[]*schedulev1.SingleScheduleResource{
				&validSingleSched1,
			},
			[]*schedulev1.SingleScheduleResource{
				&validSingleSched1,
			},
			false,
		},
		"validScheds": {
			&time1DinFutureString,
			[]*schedulev1.SingleScheduleResource{
				&validSingleSched1,
				&validSingleSched2,
				&validSingleSched3,
			},
			[]*schedulev1.SingleScheduleResource{
				&validSingleSched1,
				&validSingleSched2,
				&validSingleSched3,
			},
			false,
		},
		"expiredSched": {
			&time1DinFutureString,
			[]*schedulev1.SingleScheduleResource{
				&expiredSingleSched,
			},
			[]*schedulev1.SingleScheduleResource{},
			false,
		},
		"mixedScheds": {
			&time1DinFutureString,
			[]*schedulev1.SingleScheduleResource{
				&validSingleSched1,
				&expiredSingleSched,
			},
			[]*schedulev1.SingleScheduleResource{
				&validSingleSched1,
			},
			false,
		},
		"noTimeScheds": {
			nil,
			[]*schedulev1.SingleScheduleResource{
				&validSingleSched1,
				&expiredSingleSched,
			},
			[]*schedulev1.SingleScheduleResource{
				&validSingleSched1,
				&expiredSingleSched,
			},
			false,
		},
		"invalidTimestamp": {
			&testString,
			[]*schedulev1.SingleScheduleResource{
				&validSingleSched1,
				&expiredSingleSched,
			},
			nil,
			true,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			for _, sch := range tc.schedResources {
				dao.CreateSingleSchedule(t, tenant1,
					inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_UNSPECIFIED),
					inv_testing.SSRStart(sch.StartSeconds), inv_testing.SSREnd(sch.EndSeconds),
				)
			}
			scheduleCache.LoadAllSchedulesFromInv()
			filters := new(sc.Filters).
				Add(sc.FilterByTS(tc.timeWant))
			res, hasNext, totLen, err := hScheduleCache.GetSingleSchedules(ctx, tenant1, 0, 100, filters)

			if !tc.fail {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.False(t, hasNext)
				assert.Equal(t, len(tc.expect), totLen)
				require.Equal(t, len(tc.expect), len(res))

				for i := 0; i < len(res); i++ {
					found := false
					res[i].ResourceId = ""
					res[i].Name = ""
					res[i].CreatedAt = ""
					res[i].UpdatedAt = ""
					for j := 0; j < len(tc.expect); j++ {
						if eq, _ := inv_testing.ProtoEqualOrDiff(tc.expect[j], res[i]); eq {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("single schedule not found in cache: %v", res[i])
					}
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}

// Test does not support duplicate in the expected results.
//
//nolint:funlen // it's a test
func Test_FilterByTS_Repeat_Sched(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	scheduleCache := sc.NewScheduleCacheClient(
		inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient(),
	)
	hScheduleCache, err := sc.NewHScheduleCacheClient(scheduleCache)
	require.NoError(t, err)

	testcases := map[string]struct {
		timeWant       *string
		schedResources []*schedulev1.RepeatedScheduleResource
		expect         []*schedulev1.RepeatedScheduleResource
		fail           bool
	}{
		"validScheds": {
			&timeNowString,
			[]*schedulev1.RepeatedScheduleResource{
				&validRepeatedSchedAny,
				&validRepeatedSchedDayOfWeek,
				&validRepeatedSchedMonth,
				&validRepeatedSchedHour,
				&validRepeatedSchedNow,
			},
			[]*schedulev1.RepeatedScheduleResource{
				&validRepeatedSchedAny,
				&validRepeatedSchedDayOfWeek,
				&validRepeatedSchedMonth,
				&validRepeatedSchedHour,
				&validRepeatedSchedNow,
			},
			false,
		},
		"expiredSched": {
			&timeNowString,
			[]*schedulev1.RepeatedScheduleResource{
				&expiredRepeatedSched,
			},
			[]*schedulev1.RepeatedScheduleResource{},
			false,
		},
		"mixedScheds": {
			&timeNowString,
			[]*schedulev1.RepeatedScheduleResource{
				&validRepeatedSchedAny,
				&expiredRepeatedSched,
			},
			[]*schedulev1.RepeatedScheduleResource{
				&validRepeatedSchedAny,
			},
			false,
		},
		"noTimeScheds": {
			nil,
			[]*schedulev1.RepeatedScheduleResource{
				&validRepeatedSchedAny,
				&expiredRepeatedSched,
			},
			[]*schedulev1.RepeatedScheduleResource{
				&validRepeatedSchedAny,
				&expiredRepeatedSched,
			},
			false,
		},
		"invalidTimestamp": {
			&testString,
			[]*schedulev1.RepeatedScheduleResource{
				&validRepeatedSchedAny,
				&expiredRepeatedSched,
			},
			nil,
			true,
		},
	}
	for tcname, tc := range testcases {
		t.Run(tcname, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			for _, sch := range tc.schedResources {
				dao.CreateRepeatedSchedule(t, tenant1,
					inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_UNSPECIFIED),
					inv_testing.RSRDuration(sch.DurationSeconds), inv_testing.RSRDayWeek(sch.CronDayWeek),
					inv_testing.RSRDayMonth(sch.CronDayMonth), inv_testing.RSRMonth(sch.CronMonth),
					inv_testing.RSRHours(sch.CronHours), inv_testing.RSRMinutes(sch.CronMinutes),
				)
			}
			scheduleCache.LoadAllSchedulesFromInv()
			filters := new(sc.Filters).
				Add(sc.FilterByTS(tc.timeWant))
			res, hasNext, totLen, err := hScheduleCache.GetRepeatedSchedules(ctx, tenant1, 0, 100, filters)

			if !tc.fail {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.False(t, hasNext)
				assert.Equal(t, len(tc.expect), totLen)
				require.Equalf(t, len(tc.expect), len(res),
					"TIMEWANT: %v, EXPECTED: %v, CURRENT: %v", tc.timeWant, tc.expect, res)

				for i := 0; i < len(res); i++ {
					found := false
					res[i].ResourceId = ""
					res[i].Name = ""
					res[i].CreatedAt = ""
					res[i].UpdatedAt = ""
					for j := 0; j < len(tc.expect); j++ {
						if eq, _ := inv_testing.ProtoEqualOrDiff(tc.expect[j], res[i]); eq {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("repeated schedule not found in cache: %v", res[i])
					}
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}

//nolint:funlen // no need to extract code
func Test_HScheduleCache_GetSingleSchedules_WithHierarchy(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	// Init a schedule cache with the testing client
	scheduleCache := sc.NewScheduleCacheClient(
		inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient(),
	)
	hScheduleCache, err := sc.NewHScheduleCacheClient(scheduleCache)

	require.NoError(t, err)
	region1 := dao.CreateRegion(t, tenant1)
	region2 := dao.CreateRegion(t, tenant1)
	region3 := dao.CreateRegion(t, tenant1, inv_testing.RegionParentRegion(region2))
	site1 := dao.CreateSite(t, tenant1, inv_testing.SiteRegion(region1))
	site2 := dao.CreateSite(t, tenant1)
	site3 := dao.CreateSite(t, tenant1, inv_testing.SiteRegion(region3))
	host1 := dao.CreateHost(t, tenant1, inv_testing.HostSite(site1))
	host2 := dao.CreateHost(t, tenant1)

	sSchedR1 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRRegion(region1),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSchedR1.Relation = &schedulev1.SingleScheduleResource_TargetRegion{
		TargetRegion: region1,
	}
	sSchedR1_1 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRRegion(region1),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRStart(time1HinFutureInt), inv_testing.SSREnd(0),
	)
	sSchedR1_1.Relation = &schedulev1.SingleScheduleResource_TargetRegion{
		TargetRegion: region1,
	}
	sSchedR2 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRRegion(region2),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
	)
	sSchedR2.Relation = &schedulev1.SingleScheduleResource_TargetRegion{
		TargetRegion: region2,
	}
	sSchedR3 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRRegion(region3),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSchedR3.Relation = &schedulev1.SingleScheduleResource_TargetRegion{
		TargetRegion: region3,
	}

	sSchedS1 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRTargetSite(site1),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSchedS1.Relation = &schedulev1.SingleScheduleResource_TargetSite{
		TargetSite: site1,
	}
	sSchedS1_1 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRTargetSite(site1),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRStart(time1HinFutureInt), inv_testing.SSREnd(0))
	sSchedS1_1.Relation = &schedulev1.SingleScheduleResource_TargetSite{
		TargetSite: site1,
	}
	sSchedS2 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRTargetSite(site2),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSchedS2.Relation = &schedulev1.SingleScheduleResource_TargetSite{
		TargetSite: site2,
	}

	sSchedH1 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRTargetHost(host1),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSchedH1.Relation = &schedulev1.SingleScheduleResource_TargetHost{
		TargetHost: host1,
	}
	sSchedH1_1 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRTargetHost(host1),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.SSRStart(time1HinFutureInt), inv_testing.SSREnd(0))
	sSchedH1_1.Relation = &schedulev1.SingleScheduleResource_TargetHost{
		TargetHost: host1,
	}
	sSchedH2 := dao.CreateSingleSchedule(t, tenant1, inv_testing.SSRTargetHost(host2),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSchedH2.Relation = &schedulev1.SingleScheduleResource_TargetHost{
		TargetHost: host2,
	}

	// Add hierarchy for T2 just to ensure we guarantee isolation
	region1T2 := dao.CreateRegion(t, tenant2)
	region2T2 := dao.CreateRegion(t, tenant2, inv_testing.RegionParentRegion(region1T2))
	site1T2 := dao.CreateSite(t, tenant2, inv_testing.SiteRegion(region2T2))
	host1T2 := dao.CreateHost(t, tenant2, inv_testing.HostSite(site1T2))

	sSchedHT2 := dao.CreateSingleSchedule(t, tenant2, inv_testing.SSRRegion(region1T2),
		inv_testing.SSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	sSchedHT2.Relation = &schedulev1.SingleScheduleResource_TargetRegion{
		TargetRegion: region1T2,
	}

	scheduleCache.LoadAllSchedulesFromInv()

	testCases := map[string]struct {
		filters    *sc.Filters
		valid      bool
		exp        []*schedulev1.SingleScheduleResource
		expErrCode codes.Code
	}{
		"host1": {
			filters: new(sc.Filters).Add(sc.HasHostID(&host1.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSchedH1, sSchedH1_1, sSchedS1, sSchedS1_1, sSchedR1, sSchedR1_1},
		},
		"host1_ts": {
			filters: new(sc.Filters).Add(sc.HasHostID(&host1.ResourceId)).
				Add(sc.FilterByTS(&time1DinFutureString)),
			valid: true,
			exp:   []*schedulev1.SingleScheduleResource{sSchedH1_1, sSchedS1_1, sSchedR1_1},
		},
		"host2": {
			filters: new(sc.Filters).Add(sc.HasHostID(&host2.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSchedH2},
		},
		"site1": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&site1.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSchedS1, sSchedS1_1, sSchedR1, sSchedR1_1},
		},
		"site1_ts": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&site1.ResourceId)).
				Add(sc.FilterByTS(&time1DinFutureString)),
			valid: true,
			exp:   []*schedulev1.SingleScheduleResource{sSchedS1_1, sSchedR1_1},
		},
		"site2": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&site2.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSchedS2},
		},
		"site2_ts": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&site2.ResourceId)).
				Add(sc.FilterByTS(&time1DinFutureString)),
			valid: true,
			exp:   []*schedulev1.SingleScheduleResource{},
		},
		"site3": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&site3.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSchedR3, sSchedR2},
		},
		"region1": {
			filters: new(sc.Filters).Add(sc.HasRegionID(&region1.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSchedR1, sSchedR1_1},
		},
		"region1_ts": {
			filters: new(sc.Filters).Add(sc.HasRegionID(&region1.ResourceId)).
				Add(sc.FilterByTS(&time1DinFutureString)),
			valid: true,
			exp:   []*schedulev1.SingleScheduleResource{sSchedR1_1},
		},
		"region2": {
			filters: new(sc.Filters).Add(sc.HasRegionID(&region2.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSchedR2},
		},
		"region3": {
			filters: new(sc.Filters).Add(sc.HasRegionID(&region3.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.SingleScheduleResource{sSchedR3, sSchedR2},
		},
		"region1_ts_no_match": {
			filters: new(sc.Filters).Add(sc.HasRegionID(&region1.ResourceId)).
				Add(sc.FilterByTS(&zeroTimeString)),
			valid: true,
			exp:   []*schedulev1.SingleScheduleResource{},
		},
	}

	for tName, tc := range testCases {
		t.Run(tName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			res, hasNext, totLen, sErr := hScheduleCache.GetSingleSchedules(ctx, tenant1, 0, 100, tc.filters)
			if !tc.valid {
				require.Error(t, sErr)
				assert.Equal(t, tc.expErrCode, status.Code(sErr))
			} else {
				require.NoError(t, sErr)
				require.NotNil(t, res)
				assert.False(t, hasNext)
				assert.Equal(t, len(tc.exp), totLen)
				require.Equal(t, len(tc.exp), len(res))
				inv_testing.OrderByResourceID(res)
				inv_testing.OrderByResourceID(tc.exp)
				for i := 0; i < len(tc.exp); i++ {
					if eq, diff := inv_testing.ProtoEqualOrDiff(tc.exp[i], res[i]); !eq {
						t.Errorf("wrong single schedule in cache: %v", diff)
					}
				}
			}
		})
	}

	// Ensure Tenant2 schedules are there
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, hasNext, totLen, err := hScheduleCache.GetSingleSchedules(
		ctx, tenant2, 0, 100, new(sc.Filters).Add(sc.HasRegionID(&host1T2.ResourceId)))
	require.NoError(t, err)
	assert.Equal(t, 1, totLen)
	assert.False(t, hasNext)
	require.Len(t, res, 1)
	if eq, diff := inv_testing.ProtoEqualOrDiff(res[0], sSchedHT2); !eq {
		t.Errorf("wrong single schedule in cache: %v", diff)
	}
}

//nolint:funlen // no need to extract code
func Test_HScheduleCache_GetRepeatedSchedules_WithHierarchy(t *testing.T) {
	dao := inv_testing.NewInvResourceDAOOrFail(t)
	// Init a schedule cache with the testing client
	scheduleCache := sc.NewScheduleCacheClient(
		inv_testing.TestClients[inv_testing.APIClient].GetTenantAwareInventoryClient(),
	)
	hScheduleCache, err := sc.NewHScheduleCacheClient(scheduleCache)
	require.NoError(t, err)

	optsMatchingNow := []inv_testing.Opt[schedulev1.RepeatedScheduleResource]{
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE),
		inv_testing.RSRDayWeek(fmt.Sprintf("%d", timeNow.Weekday())),
		inv_testing.RSRDayMonth(cronAny),
		inv_testing.RSRMonth(cronAny),
		inv_testing.RSRHours(cronAny),
		inv_testing.RSRMinutes(cronAny),
		nil,
	}

	region1 := dao.CreateRegion(t, tenant1)
	region2 := dao.CreateRegion(t, tenant1)
	region3 := dao.CreateRegion(t, tenant1, inv_testing.RegionParentRegion(region2))
	site1 := dao.CreateSite(t, tenant1, inv_testing.SiteRegion(region1))
	site2 := dao.CreateSite(t, tenant1)
	site3 := dao.CreateSite(t, tenant1, inv_testing.SiteRegion(region3))
	host1 := dao.CreateHost(t, tenant1, inv_testing.HostSite(site1))
	host2 := dao.CreateHost(t, tenant1)

	rSchedR1 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRRegion(region1),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSchedR1.Relation = &schedulev1.RepeatedScheduleResource_TargetRegion{
		TargetRegion: region1,
	}
	optsMatchingNow[len(optsMatchingNow)-1] = inv_testing.RSRRegion(region1)
	rSchedR1_1 := dao.CreateRepeatedSchedule(t, tenant1, optsMatchingNow...)
	rSchedR1_1.Relation = &schedulev1.RepeatedScheduleResource_TargetRegion{
		TargetRegion: region1,
	}
	rSchedR2 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRRegion(region2),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSchedR2.Relation = &schedulev1.RepeatedScheduleResource_TargetRegion{
		TargetRegion: region2,
	}
	optsMatchingNow[len(optsMatchingNow)-1] = inv_testing.RSRRegion(region3)
	rSchedR3 := dao.CreateRepeatedSchedule(t, tenant1, optsMatchingNow...)
	rSchedR3.Relation = &schedulev1.RepeatedScheduleResource_TargetRegion{
		TargetRegion: region3,
	}

	optsMatchingNow = optsMatchingNow[:len(optsMatchingNow)-1]
	rSchedS1 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRTargetSite(site1),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSchedS1.Relation = &schedulev1.RepeatedScheduleResource_TargetSite{
		TargetSite: site1,
	}
	optsMatchingNow = append(optsMatchingNow, inv_testing.RSRTargetSite(site1))
	rSchedS1_1 := dao.CreateRepeatedSchedule(t, tenant1, optsMatchingNow...)
	optsMatchingNow = optsMatchingNow[:len(optsMatchingNow)-1]
	rSchedS1_1.Relation = &schedulev1.RepeatedScheduleResource_TargetSite{
		TargetSite: site1,
	}
	rSchedS2 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRTargetSite(site2),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSchedS2.Relation = &schedulev1.RepeatedScheduleResource_TargetSite{
		TargetSite: site2,
	}

	rSchedH1 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRTargetHost(host1),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSchedH1.Relation = &schedulev1.RepeatedScheduleResource_TargetHost{
		TargetHost: host1,
	}
	optsMatchingNow = append(optsMatchingNow, inv_testing.RSRTargetHost(host1))
	rSchedH1_1 := dao.CreateRepeatedSchedule(t, tenant1, optsMatchingNow...)
	rSchedH1_1.Relation = &schedulev1.RepeatedScheduleResource_TargetHost{
		TargetHost: host1,
	}
	rSchedH2 := dao.CreateRepeatedSchedule(t, tenant1, inv_testing.RSRTargetHost(host2),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSchedH2.Relation = &schedulev1.RepeatedScheduleResource_TargetHost{
		TargetHost: host2,
	}

	// Add hierarchy for T2 just to ensure we guarantee isolation
	region1T2 := dao.CreateRegion(t, tenant2)
	region2T2 := dao.CreateRegion(t, tenant2, inv_testing.RegionParentRegion(region1T2))
	site1T2 := dao.CreateSite(t, tenant2, inv_testing.SiteRegion(region2T2))
	host1T2 := dao.CreateHost(t, tenant2, inv_testing.HostSite(site1T2))

	rSchedHT2 := dao.CreateRepeatedSchedule(t, tenant2, inv_testing.RSRRegion(region1T2),
		inv_testing.RSRStatus(schedulev1.ScheduleStatus_SCHEDULE_STATUS_MAINTENANCE))
	rSchedHT2.Relation = &schedulev1.RepeatedScheduleResource_TargetRegion{
		TargetRegion: region1T2,
	}

	scheduleCache.LoadAllSchedulesFromInv()

	testCases := map[string]struct {
		filters    *sc.Filters
		timestamp  *string
		valid      bool
		exp        []*schedulev1.RepeatedScheduleResource
		expErrCode codes.Code
	}{
		"host1": {
			filters: new(sc.Filters).Add(sc.HasHostID(&host1.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.RepeatedScheduleResource{rSchedH1, rSchedH1_1, rSchedS1, rSchedS1_1, rSchedR1, rSchedR1_1},
		},
		"host1_ts": {
			filters: new(sc.Filters).Add(sc.HasHostID(&host1.ResourceId)).
				Add(sc.FilterByTS(&timeNowString)),
			valid: true,
			exp:   []*schedulev1.RepeatedScheduleResource{rSchedH1_1, rSchedS1_1, rSchedR1_1},
		},
		"host2": {
			filters: new(sc.Filters).Add(sc.HasHostID(&host2.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.RepeatedScheduleResource{rSchedH2},
		},
		"site1": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&site1.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.RepeatedScheduleResource{rSchedS1, rSchedS1_1, rSchedR1, rSchedR1_1},
		},
		"site1_ts": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&site1.ResourceId)).
				Add(sc.FilterByTS(&timeNowString)),
			valid: true,
			exp:   []*schedulev1.RepeatedScheduleResource{rSchedS1_1, rSchedR1_1},
		},
		"site2": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&site2.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.RepeatedScheduleResource{rSchedS2},
		},
		"site2_ts": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&site2.ResourceId)).
				Add(sc.FilterByTS(&timeNowString)),
			valid: true,
			exp:   []*schedulev1.RepeatedScheduleResource{},
		},
		"site3": {
			filters: new(sc.Filters).Add(sc.HasSiteID(&site3.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.RepeatedScheduleResource{rSchedR3, rSchedR2},
		},
		"region1": {
			filters: new(sc.Filters).Add(sc.HasRegionID(&region1.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.RepeatedScheduleResource{rSchedR1, rSchedR1_1},
		},
		"region1_ts": {
			filters: new(sc.Filters).Add(sc.HasRegionID(&region1.ResourceId)).
				Add(sc.FilterByTS(&timeNowString)),
			valid: true,
			exp:   []*schedulev1.RepeatedScheduleResource{rSchedR1_1},
		},
		"region2": {
			filters: new(sc.Filters).Add(sc.HasRegionID(&region2.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.RepeatedScheduleResource{rSchedR2},
		},
		"region3": {
			filters: new(sc.Filters).Add(sc.HasRegionID(&region3.ResourceId)),
			valid:   true,
			exp:     []*schedulev1.RepeatedScheduleResource{rSchedR3, rSchedR2},
		},
	}

	for tName, tc := range testCases {
		t.Run(tName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			res, hasNext, totLen, rsErr := hScheduleCache.GetRepeatedSchedules(ctx, tenant1, 0, 100, tc.filters)
			if !tc.valid {
				require.Error(t, rsErr)
				assert.Equal(t, tc.expErrCode, status.Code(rsErr))
			} else {
				require.NoError(t, rsErr)
				require.NotNil(t, res)
				assert.False(t, hasNext)
				assert.Equal(t, len(tc.exp), totLen)
				require.Equal(t, len(tc.exp), len(res))
				inv_testing.OrderByResourceID(res)
				inv_testing.OrderByResourceID(tc.exp)
				for i := 0; i < len(tc.exp); i++ {
					if eq, diff := inv_testing.ProtoEqualOrDiff(tc.exp[i], res[i]); !eq {
						t.Errorf("wrong single schedule in cache: %v", diff)
					}
				}
			}
		})
	}

	// Ensure Tenant2 schedules are there
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, hasNext, totLen, err := hScheduleCache.GetRepeatedSchedules(
		ctx, tenant2, 0, 100, new(sc.Filters).Add(sc.HasRegionID(&host1T2.ResourceId)))
	require.NoError(t, err)
	assert.Equal(t, 1, totLen)
	assert.False(t, hasNext)
	require.Len(t, res, 1)
	if eq, diff := inv_testing.ProtoEqualOrDiff(res[0], rSchedHT2); !eq {
		t.Errorf("wrong repeated schedule in cache: %v", diff)
	}
}

func Test_NewStandardFilter(t *testing.T) {
	newFilter := sc.NewStandardFilter(nil, "")
	assert.Nil(t, newFilter.GetFilterFunc())
	assert.Equal(t, "", newFilter.GetDescription())

	newFilter = sc.NewStandardFilter(func(_ *inv_v1.Resource) bool {
		return true
	}, "test")
	assert.NotNil(t, newFilter.GetFilterFunc())
	assert.Equal(t, "test", newFilter.GetDescription())
}

func buildTestFilter(filters *sc.Filters, host, site, region *string) {
	if function.IsEmptyNullCase(host) {
		filters.Add(sc.HasNoHost())
	}
	if function.IsNotEmptyNullCase(host) {
		filters.Add(sc.HasHostID(host))
	}

	if function.IsEmptyNullCase(site) {
		filters.Add(sc.HasNoSite())
	}
	if function.IsNotEmptyNullCase(site) {
		filters.Add(sc.HasSiteID(site))
	}

	if function.IsEmptyNullCase(region) {
		filters.Add(sc.HasNoRegion())
	}
	if function.IsNotEmptyNullCase(region) {
		filters.Add(sc.HasRegionID(region))
	}
}

func assertRSRInCache(
	t *testing.T, schedCache *sc.ScheduleCacheClient, tID string, rsr *schedulev1.RepeatedScheduleResource,
) {
	t.Helper()
	cachedRsr, err := schedCache.GetRepeatedSchedule(tID, rsr.GetResourceId())
	require.NoError(t, err)
	assert.Equal(t, rsr.GetName(), cachedRsr.GetName())
}

func assertSSRInCache(
	t *testing.T, schedCache *sc.ScheduleCacheClient, tID string, ssr *schedulev1.SingleScheduleResource,
) {
	t.Helper()
	cachedSsr, err := schedCache.GetSingleSchedule(tID, ssr.GetResourceId())
	require.NoError(t, err)
	assert.Equal(t, ssr.GetName(), cachedSsr.GetName())
}

func assertRSRNotInCache(
	t *testing.T, schedCache *sc.ScheduleCacheClient, tID, rsrID string,
) {
	t.Helper()
	cached, err := schedCache.GetRepeatedSchedule(tID, rsrID)
	require.Error(t, err)
	assert.Nil(t, cached)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func assertSSRNotInCache(
	t *testing.T, schedCache *sc.ScheduleCacheClient, tID, ssrID string,
) {
	t.Helper()
	cached, err := schedCache.GetSingleSchedule(tID, ssrID)
	require.Error(t, err)
	assert.Nil(t, cached)
	assert.Equal(t, codes.NotFound, status.Code(err))
}
