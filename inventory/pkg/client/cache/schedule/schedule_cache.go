// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package schedule

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"sync"
	"time"

	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	schedulev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	inverr "github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/collections"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/comparator"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/util/paginator"
)

// This is a special inventory client, that subscribed to schedulesByCacheKey changes, caches the schedule resources locally,
// and exposes internal Go API to other components in order to fetch schedulesByCacheKey given some primitive filters

var (
	logC = logging.GetLogger("ScheduleCache")

	PeriodicCacheRefresh = flag.Duration(
		"schedCachePeriodicRefresh", defaultPeriodicRefresh, "Periodic refresh timeout of the schedule cache")
	inventoryTimeout = flag.Duration(
		"schedCacheInvTimeout", defaultInventoryTimeout, "Schedule cache inventory API calls timeout")
	listAllInventoryTimeout = flag.Duration(
		"schedCacheListAllTimeout",
		defaultListAllTimeout,
		"Timeout used when listing all schedule resources in the schedule cache from Inventory",
	)
)

const (
	BatchSize               = 500
	defaultPeriodicRefresh  = 5 * time.Minute
	defaultInventoryTimeout = 5 * time.Second
	defaultListAllTimeout   = time.Minute

	defaultRegisterMaxElapsedTime = 30 * time.Second

	// eventsWatcherBufSize is the buffer size for the events channel.
	eventsWatcherBufSize = 10
)

type cacheKey struct {
	tenantID, id string
}

func (c cacheKey) String() string {
	return fmt.Sprintf("key[tenantID=%s, resourceID=%s]", c.tenantID, c.id)
}

func ssasCacheKey(ss *schedulev1.SingleScheduleResource) cacheKey {
	return key(ss.GetTenantId(), ss.GetResourceId())
}

func rsasCacheKey(rs *schedulev1.RepeatedScheduleResource) cacheKey {
	return key(rs.GetTenantId(), rs.GetResourceId())
}

func key(tenantID, id string) cacheKey {
	return cacheKey{
		tenantID: tenantID,
		id:       id,
	}
}

// Cache struct used to cache the schedules, multiple maps in order to heave easier access to the schedule resources.
type cache struct {
	schedulesByCacheKey map[cacheKey]*inv_v1.Resource
	lock                sync.Mutex
}

func newCache() *cache {
	return &cache{
		schedulesByCacheKey: make(map[cacheKey]*inv_v1.Resource, 0),
		lock:                sync.Mutex{},
	}
}

func (c *cache) invalidate(key cacheKey) {
	logC.Debug().Msgf("invalidate(%s)", key)
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.schedulesByCacheKey, key)
}

func (c *cache) store(key cacheKey, resource *inv_v1.Resource) {
	logC.Debug().Msgf("store(%s, %v)", key, resource)
	c.lock.Lock()
	defer c.lock.Unlock()
	c.schedulesByCacheKey[key] = resource
}

func (c *cache) get(key cacheKey) *inv_v1.Resource {
	logC.Debug().Msgf("get(%s)", key)
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.schedulesByCacheKey[key]
}

func (c *cache) initialize(schedulesByCacheKey map[cacheKey]*inv_v1.Resource) {
	logC.Debug().Msgf("initialize(%v)", schedulesByCacheKey)
	c.lock.Lock()
	defer c.lock.Unlock()
	c.schedulesByCacheKey = schedulesByCacheKey
}

func (c *cache) getAll() []*inv_v1.Resource {
	logC.Debug().Msgf("getAll()")
	c.lock.Lock()
	defer c.lock.Unlock()
	return maps.Values(c.schedulesByCacheKey)
}

//nolint:revive // use Schedule prefix
type ScheduleCacheClient struct {
	InvClient client.TenantAwareInventoryClient
	cache     *cache

	wg      *sync.WaitGroup
	sigTerm chan bool
}

//nolint:revive // use Schedule prefix
type ScheduleCacheClientConfig struct {
	// Address is the inventory target to connect to.
	Address                string
	RegisterMaxElapsedTime time.Duration
	EnableTracing          bool
}

type Options struct {
	// InventoryAddress is the inventory target to connect to.
	InventoryAddress string
	// EnableTracing is used to enable tracing.
	EnableTracing          bool
	RegisterMaxElapsedTime time.Duration
	DialOptions            []grpc.DialOption
}

type Option func(*Options)

// WithInventoryAddress sets the Inventory Address.
func WithInventoryAddress(invAddr string) Option {
	return func(options *Options) {
		options.InventoryAddress = invAddr
	}
}

// WithEnableTracing enable tracing.
func WithEnableTracing(enableTracing bool) Option {
	return func(options *Options) {
		options.EnableTracing = enableTracing
	}
}

func WithRegisterMaxElapsedTime(maxElapsedTime time.Duration) Option {
	return func(options *Options) {
		options.RegisterMaxElapsedTime = maxElapsedTime
	}
}

func WithDialOption(dialOption grpc.DialOption) Option {
	return func(options *Options) {
		options.DialOptions = append(options.DialOptions, dialOption)
	}
}

// WithOptions sets the Inventory client options.
func WithOptions(options Options) Option {
	return func(opts *Options) {
		*opts = options
	}
}

// NewScheduleCacheClientWithOptions creates a client for the Inventory Service.
func NewScheduleCacheClientWithOptions(
	ctx context.Context,
	opts ...Option,
) (*ScheduleCacheClient, error) {
	var wg sync.WaitGroup
	eventsWatcher := make(chan *client.WatchEvents, eventsWatcherBufSize)

	var options Options
	options.RegisterMaxElapsedTime = defaultRegisterMaxElapsedTime
	for _, opt := range opts {
		opt(&options)
	}

	clientCfg := client.InventoryClientConfig{
		Name:    "schedule_cache",
		Address: options.InventoryAddress,
		SecurityCfg: &client.SecurityConfig{
			// TODO: support secured connection
			Insecure: true,
		},
		Events:                    eventsWatcher,
		EnableRegisterRetry:       false,
		AbortOnUnknownClientError: true,
		ClientKind:                inv_v1.ClientKind_CLIENT_KIND_API,
		ResourceKinds: []inv_v1.ResourceKind{
			inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE,
			inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE,
		},
		Wg:            &wg,
		EnableTracing: options.EnableTracing,
		DialOptions:   options.DialOptions,
	}

	invClient, err := client.NewTenantAwareInventoryClient(ctx, clientCfg)
	if err != nil {
		return nil, err
	}

	invHandler := NewScheduleCacheClient(invClient)

	// blocking load the first time
	invHandler.LoadAllSchedulesFromInv()

	// Schedule periodic job to update the cache every periodicCacheRefreshSec seconds
	ticker := time.NewTicker(*PeriodicCacheRefresh)

	invHandler.wg.Add(1)
	go func() {
		for {
			select {
			case ev, ok := <-eventsWatcher:
				if ok {
					invHandler.manageEvent(ev.Event)
				} else {
					// If eventsWatcher channel is closed, stream ended
					ticker.Stop()
					logC.InfraSec().Fatal().Msg("gRPC stream with inventory closed")
				}
			case <-ticker.C:
				// Reconcile state of the cache
				invHandler.LoadAllSchedulesFromInv()
			case <-invHandler.sigTerm:
				// Stop the ticker and signal done
				// No other events will be processed
				ticker.Stop()
				invHandler.wg.Done()
				return
			}
		}
	}()

	return invHandler, err
}

// NewScheduleCacheClient creates a client that wraps an existing Inventory client. Mainly for testing.
func NewScheduleCacheClient(invClient client.TenantAwareInventoryClient) *ScheduleCacheClient {
	return &ScheduleCacheClient{
		InvClient: invClient,
		cache:     newCache(),
		wg:        &sync.WaitGroup{},
		sigTerm:   make(chan bool),
	}
}

func (sc *ScheduleCacheClient) Stop() {
	close(sc.sigTerm)
	sc.wg.Wait()
	logC.Info().Msg("Schedule cache client stopped")
}

func (sc *ScheduleCacheClient) manageEvent(event *inv_v1.SubscribeEventsResponse) {
	tenantID, resourceID, err := util.GetResourceKeyFromResource(event.GetResource())
	if err != nil {
		logC.Err(err).Msgf("this should never happen, tenant ID and resource ID should always be set in the resource")
		return
	}
	evKey := key(tenantID, resourceID)
	logC.Debug().Msgf("Got event: eventKind=%s, %s", event.EventKind, evKey)
	kind, err := util.GetResourceKindFromResourceID(resourceID)
	if err != nil {
		return
	}
	if kind != inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE && kind != inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE {
		logC.InfraError("Unexpected resource kind in manageEvent: eventKind=%s", kind)
		return
	}
	sc.reconcileResource(evKey, event.EventKind)
}

func (sc *ScheduleCacheClient) getResource(key cacheKey) (*inv_v1.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), *inventoryTimeout)
	defer cancel()
	res, err := sc.InvClient.Get(ctx, key.tenantID, key.id)
	if err != nil {
		return nil, err
	}
	return res.GetResource(), nil
}

func (sc *ScheduleCacheClient) Invalidate(key cacheKey) {
	sc.cache.invalidate(key)
}

func (sc *ScheduleCacheClient) reconcileResource(key cacheKey, evKind inv_v1.SubscribeEventsResponse_EventKind) {
	if evKind == inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED {
		// Do not query inventory when deleting, resource won't be there
		sc.Invalidate(key)
		return
	}

	// Not a DELETE
	res, err := sc.getResource(key)
	if err != nil {
		logC.InfraErr(err).Msgf("Failed to reconcile resource: %s", key)
		return
	}

	switch evKind {
	case inv_v1.SubscribeEventsResponse_EVENT_KIND_CREATED, inv_v1.SubscribeEventsResponse_EVENT_KIND_UPDATED:
		sc.cache.store(key, res)
	default:
		logC.InfraError(
			"Unexpected event while reconciling resource: %s, evKind=%s", key, evKind)
	}
}

func getResourceFromType(resourceType inv_v1.ResourceKind) *inv_v1.Resource {
	switch resourceType {
	case inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE:
		return &inv_v1.Resource{Resource: &inv_v1.Resource_Singleschedule{}}
	case inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE:
		return &inv_v1.Resource{Resource: &inv_v1.Resource_Repeatedschedule{}}
	// TODO: implement for other resources
	default:
		return nil
	}
}

//nolint:cyclop // cyclomatic complexity is 11
func loadResourceFromInv(
	ctx context.Context,
	invClient client.TenantAwareInventoryClient,
	resourceType inv_v1.ResourceKind,
) map[cacheKey]*inv_v1.Resource {
	filterRequest := inv_v1.ResourceFilter{
		Resource: getResourceFromType(resourceType),
		Filter:   "", // Empty filter, get all resources for the given type
		Limit:    BatchSize,
		Offset:   0,
	}
	resources := make(map[cacheKey]*inv_v1.Resource)
	hasNext := true
	for hasNext {
		listResponse, err := invClient.List(ctx, &filterRequest)
		if inverr.IsNotFound(err) {
			logC.Debug().Msgf("No more %v resources in inventory.", resourceType)
			break
		}
		if err != nil {
			logC.InfraErr(err).Msgf("Failed to %v resources from inventory.", resources)
			break
		}
		var wrongResourceType inv_v1.ResourceKind
		for _, res := range listResponse.GetResources() {
			wrongResourceType = inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED
			switch res.GetResource().GetResource().(type) {
			case *inv_v1.Resource_Singleschedule:
				if resourceType == inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE {
					resources[ssasCacheKey(res.GetResource().GetSingleschedule())] = res.GetResource()
				} else {
					wrongResourceType = inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE
				}
			case *inv_v1.Resource_Repeatedschedule:
				if resourceType == inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE {
					resources[rsasCacheKey(res.GetResource().GetRepeatedschedule())] = res.GetResource()
				} else {
					wrongResourceType = inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE
				}
			// TODO: implement for other resources
			default:
				logC.InfraError("Unsupported resource type: %v", resourceType)
			}
			if wrongResourceType != inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED {
				// We should never reach this point
				logC.InfraError("Got wrong resource type: expected=%v, got=%v", resourceType, wrongResourceType)
			}
		}
		hasNext = listResponse.HasNext
		filterRequest.Offset += BatchSize
	}
	return resources
}

func (sc *ScheduleCacheClient) LoadAllSchedulesFromInv() {
	// TODO Current reconciliation is dumb, clean everything in the local cache and re-build it
	ctx, cancel := context.WithTimeout(context.Background(), *listAllInventoryTimeout)
	defer cancel()

	singleSchedules := loadResourceFromInv(ctx, sc.InvClient, inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE)
	repeatedSchedules := loadResourceFromInv(ctx, sc.InvClient, inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE)
	schedules := make(map[cacheKey]*inv_v1.Resource, len(singleSchedules)+len(repeatedSchedules))
	for k, v := range singleSchedules {
		schedules[k] = v
	}
	for k, v := range repeatedSchedules {
		schedules[k] = v
	}

	sc.cache.initialize(schedules)
}

func (sc *ScheduleCacheClient) getAllSchedules(resKind inv_v1.ResourceKind, tenantID string) []*inv_v1.Resource {
	matchedSchedules := make([]*inv_v1.Resource, 0)
	for _, res := range sc.cache.getAll() {
		switch {
		case resKind == inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE &&
			res.GetSingleschedule() != nil && res.GetSingleschedule().GetTenantId() == tenantID:
			matchedSchedules = append(matchedSchedules, res)
		case resKind == inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE &&
			res.GetRepeatedschedule() != nil && res.GetRepeatedschedule().GetTenantId() == tenantID:
			matchedSchedules = append(matchedSchedules, res)
		default:
			// no-op, skip
		}
	}
	return matchedSchedules
}

// GetSchedules returns all schedules filtered by the provided filters and pagination info.
func (sc *ScheduleCacheClient) GetSchedules(resKind inv_v1.ResourceKind, tenantID string, filters []*Filters, offset, limit int) (
	res []*inv_v1.Resource,
	next bool,
	total int,
	err error,
) {
	// make sure Filters are specified, if not apply DefaultFilter(getALL)
	if filters == nil {
		filters = []*Filters{DefaultFilter}
	}

	all := sc.getAllSchedules(resKind, tenantID)
	var filtered []*inv_v1.Resource
	for _, filter := range filters {
		// if filter is not specified, apply DefaultFilter(getALL)
		if filter == nil {
			filter = DefaultFilter
		}
		if err := filter.Validate(); err != nil {
			return nil, false, 0, err
		}
		filtered = append(filtered, collections.Filter(all, filter.Evaluate)...)
	}
	sort.Slice(comparator.ResourceIDAscComparator(filtered))
	res, next, total = paginator.NewPaginator[*inv_v1.Resource](offset, limit).Apply(filtered)
	return res, next, total, nil
}

func (sc *ScheduleCacheClient) getSchedule(resKind inv_v1.ResourceKind, key cacheKey) (*inv_v1.Resource, error) {
	if res := sc.cache.get(key); res != nil {
		switch resKind {
		case inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE:
			if res.GetSingleschedule() != nil {
				return res, nil
			}
		case inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE:
			if res.GetRepeatedschedule() != nil {
				return res, nil
			}
		default:
			// no-op
		}
	}
	logC.Debug().Msgf("Schedule cache miss: %s", key)
	return nil, inverr.Errorfc(codes.NotFound, "Resource not found in cache")
}

func (sc *ScheduleCacheClient) GetSingleSchedule(tenantID, resourceID string) (*schedulev1.SingleScheduleResource, error) {
	ssKey := key(tenantID, resourceID)
	logC.Debug().Msgf("Get single schedule from cache: %s", ssKey)
	res, err := sc.getSchedule(inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE, ssKey)
	if err != nil {
		return nil, err
	}
	return res.GetSingleschedule(), nil
}

func (sc *ScheduleCacheClient) GetRepeatedSchedule(tenantID, resourceID string) (*schedulev1.RepeatedScheduleResource, error) {
	rsKey := key(tenantID, resourceID)
	logC.Debug().Msgf("Get repeated schedule from cache: %s", rsKey)
	res, err := sc.getSchedule(inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE, rsKey)
	if err != nil {
		return nil, err
	}
	return res.GetRepeatedschedule(), nil
}

func (sc *ScheduleCacheClient) InvalidateCache(
	tenantID, resourceID string, invalidateKind inv_v1.SubscribeEventsResponse_EventKind,
) {
	invKey := key(tenantID, resourceID)
	logC.Debug().Msgf("Invalidate cache: eventKind=%s, %s", invalidateKind, invKey)
	sc.reconcileResource(invKey, invalidateKind)
}

// TestGetAllSchedules gets all the schedule without tenant distinction. Used for testing purposes only.
func (sc *ScheduleCacheClient) TestGetAllSchedules() []*inv_v1.Resource {
	return sc.cache.getAll()
}
