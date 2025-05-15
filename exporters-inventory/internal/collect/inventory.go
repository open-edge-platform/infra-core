// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/open-edge-platform/infra-core/exporters-inventory/internal/common"
	"github.com/open-edge-platform/infra-core/exporters-inventory/internal/env"
	"github.com/open-edge-platform/infra-core/exporters-inventory/internal/kpis"
	computev1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/compute/v1"
	inv_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/inventory/v1"
	sched_v1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/schedule/v1"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/client"
	schedule_cache "github.com/open-edge-platform/infra-core/inventory/v2/pkg/client/cache/schedule"
	inv_utils "github.com/open-edge-platform/infra-core/inventory/v2/pkg/util"
)

const (
	logretension            = 600
	batchSize               = 100
	periodicCacheRefreshSec = 60 // sets to refresh the cache every 60 seconds (consider scale)
	totalNumberOFCurlReq    = 5
	defaultTimeout          = 30 * time.Second
	defaultTimeoutRPCCalls  = 5 * time.Second
	// eventsWatcherBufSize is the buffer size for the events channel.
	eventsWatcherBufSize = 10
)

var clientName = "exporter"

// Map to store the TotalProvisioningTime per HostID.
var totalProvisioningTimeMap = map[string]float64{}

// InventoryCollector is the Inventory collector.
// It extracts all the inventory related kpis using the Collect method.
// It implements the Collector interface.
type InventoryCollector struct {
	Name            common.CollectorName
	Address         string
	CollectorClient *InvCollectorClient
	HScheduleCache  *schedule_cache.HScheduleCacheClient
	Cancel          context.CancelFunc
}

func NewInventoryCollector(cfg common.CollectorsConfig) (Collector, error) {
	chanTerm := make(chan bool)
	var wg sync.WaitGroup
	eventsWatcher := make(chan *client.WatchEvents, eventsWatcherBufSize)

	ctx, cancel := context.WithCancel(context.Background())
	invClient, err := newInventoryClient(
		ctx,
		&wg,
		eventsWatcher,
		cfg.Address,
		cfg.CAPath,
		cfg.CertPath,
		cfg.KeyPath,
		cfg.EnableTracing,
	)
	if err != nil {
		cancel()
		return nil, err
	}

	scheduleCache, err := schedule_cache.NewScheduleCacheClientWithOptions(ctx,
		schedule_cache.WithInventoryAddress(cfg.Address),
		schedule_cache.WithEnableTracing(cfg.EnableTracing),
	)
	if err != nil {
		cancel()
		return nil, err
	}
	hScheduleCache, err := schedule_cache.NewHScheduleCacheClient(scheduleCache)
	if err != nil {
		cancel()
		return nil, err
	}

	collectorClient := NewInvCollectorCache(invClient, chanTerm, &wg, eventsWatcher)
	return &InventoryCollector{
		Name:            cfg.Name,
		Address:         cfg.Address,
		Cancel:          cancel,
		CollectorClient: collectorClient,
		HScheduleCache:  hScheduleCache,
	}, nil
}

// Struct used to cache the schedules/hosts, multiple maps in order to heave easier access to the schedule resources.
type invCollectorHostsCache struct {
	Hosts map[string]*inv_v1.Resource
	lock  sync.Mutex
}

type InvCollectorClient struct {
	InvClient client.TenantAwareInventoryClient
	chanTerm  chan bool
	wg        *sync.WaitGroup
	Cache     invCollectorHostsCache
}

// Collect implements the collector of the edge infrastructure manager service kpis.
// It uses the function(s) defined in this file to extract the inventory collector
// kpis and return a list of them.
func (col *InventoryCollector) Collect() ([]kpis.KPI, error) {
	// Locks the collector cache to avoid race conditions
	// with reconcileResource method.
	col.CollectorClient.Cache.lock.Lock()
	defer col.CollectorClient.Cache.lock.Unlock()

	collectedKpis := []kpis.KPI{}
	statusKPI := kpis.NewInventoryHostsStatus()
	provisioningStatusKPI := kpis.NewInventoryHostsProvisioning()
	onboardingStatusKPI := kpis.NewInventoryHostsOnboarding()
	updateStatusKPI := kpis.NewInventoryHostsUpdate()
	maintenanceKPI := kpis.NewInventoryHostsSchedule()
	totalProvisioningTimeKPI := kpis.NewInventoryHostsTotalProvisioningTime()
	statusKPI.Status = make(map[string]kpis.HostStatus)
	maintenanceKPI.Status = make(map[string]kpis.HostStatus)
	provisioningStatusKPI.Status = make(map[string]kpis.HostStatus)
	onboardingStatusKPI.Status = make(map[string]kpis.HostStatus)
	updateStatusKPI.Status = make(map[string]kpis.HostStatus)
	totalProvisioningTimeKPI.ProvisioningTime = make(map[string]kpis.HostProvisioningTime)

	var hostHasSched bool
	timeNowString := fmt.Sprint(time.Now().UTC().Unix())
	hosts := col.CollectorClient.getHosts()
	log.Debug().Msg("InventoryCollector collect metrics")
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutRPCCalls)
	defer cancel()
	for _, res := range hosts {
		if res == nil {
			continue
		}
		hostHasSched = false

		host := res.GetHost()
		hostID := host.GetResourceId()
		instanceStatus := host.Instance.GetInstanceStatus()
		tenantID := host.GetTenantId()
		hostSingleScheds := col.loadSingleSchedulesFromCache(
			ctx,
			&hostID,
			&timeNowString,
			&tenantID,
		)
		hostRepeatScheds := col.loadRepeatedSchedulesFromCache(
			ctx,
			&hostID,
			&timeNowString,
			&tenantID,
		)

		if len(hostSingleScheds) > 0 || len(hostRepeatScheds) > 0 {
			hostHasSched = true
		}

		hostStatus := constructHostStruct(hostHasSched, host)
		statusKPI.Status[hostID] = hostStatus
		maintenanceKPI.Status[hostID] = hostStatus
		provisioningStatusKPI.Status[hostID] = hostStatus
		onboardingStatusKPI.Status[hostID] = hostStatus
		updateStatusKPI.Status[hostID] = hostStatus
		if instanceStatus == "Running" {
			hostProvisioningTimeStruct := calculateTotalProvisioningTime(ctx, host)
			// if TotalProvisioningTime is 0 will not send data
			if hostProvisioningTimeStruct.TotalProvisioningTime != 0 {
				totalProvisioningTimeKPI.ProvisioningTime[hostID] = calculateTotalProvisioningTime(ctx, host)
			}
		}
	}

	collectedKpis = append(collectedKpis,
		statusKPI,
		maintenanceKPI,
		provisioningStatusKPI,
		onboardingStatusKPI,
		updateStatusKPI,
	)
	if len(totalProvisioningTimeKPI.ProvisioningTime) > 0 {
		collectedKpis = append(collectedKpis, totalProvisioningTimeKPI)
	}
	return collectedKpis, nil
}

func (col *InventoryCollector) Stop() {
	col.CollectorClient.chanTerm <- true
	// col.Cancel()
}

// newInventoryClient creates a client for the Inventory Service.
func newInventoryClient(
	ctx context.Context,
	wg *sync.WaitGroup,
	eventsWatcher chan *client.WatchEvents,
	address, caPath, certPath, keyPath string,
	enableTracing bool,
) (client.TenantAwareInventoryClient, error) {
	insecureConnection := true

	clientCfg := client.InventoryClientConfig{
		Name:    clientName,
		Address: address,
		SecurityCfg: &client.SecurityConfig{
			Insecure: insecureConnection,
			CaPath:   caPath,
			CertPath: certPath,
			KeyPath:  keyPath,
		},
		Events:                 eventsWatcher,
		EnableRegisterRetry:    true,
		RegisterMaxElapsedTime: defaultTimeout,
		ClientKind:             inv_v1.ClientKind_CLIENT_KIND_API,
		ResourceKinds: []inv_v1.ResourceKind{
			inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE,
			inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE,
			inv_v1.ResourceKind_RESOURCE_KIND_HOST,
		},
		Wg:            wg,
		EnableTracing: enableTracing,
	}

	invClient, err := client.NewTenantAwareInventoryClient(ctx, clientCfg)
	if err != nil {
		return nil, err
	}
	return invClient, nil
}

// NewInvCollectorCache uses the inventory client to create a cache for the Inventory collector.
func NewInvCollectorCache(
	invClient client.TenantAwareInventoryClient,
	chanTerm chan bool,
	wg *sync.WaitGroup,
	eventsWatcher chan *client.WatchEvents,
) *InvCollectorClient {
	invHandler := &InvCollectorClient{
		InvClient: invClient,
		chanTerm:  chanTerm,
		wg:        wg,
		Cache: invCollectorHostsCache{
			Hosts: make(map[string]*inv_v1.Resource, 0),
			lock:  sync.Mutex{},
		},
	}
	// blocking load the first time
	invHandler.LoadAllHostsFromInv()
	// starts to listen to events
	invHandler.watchEvents(eventsWatcher)
	return invHandler
}

func (sc *InvCollectorClient) watchEvents(eventsWatcher chan *client.WatchEvents) {
	// Schedule periodic job to update the cache every periodicCacheRefreshSec seconds
	ticker := time.NewTicker(periodicCacheRefreshSec * time.Second)

	go func() {
		for {
			select {
			case ev, ok := <-eventsWatcher:
				if ok {
					sc.manageEvent(ev.Event)
				} else {
					// If eventsWatcher channel is closed, stream ended.
					ticker.Stop()
					log.InfraSec().Fatal().Msg("gRPC stream with inventory closed")
				}
			case <-ticker.C:
				// Reconcile state of the cache, schedules and hosts.
				sc.LoadAllHostsFromInv()
			case <-sc.chanTerm:
				return
			}
		}
	}()
	log.Info().Msgf("watchEvents started")
}

func (sc *InvCollectorClient) manageEvent(event *inv_v1.SubscribeEventsResponse) {
	log.Debug().Msgf("Got event: eventKind=%s, resourceID=%s", event.EventKind, event.ResourceId)
	kind, err := inv_utils.GetResourceKindFromResourceID(event.ResourceId)
	if err != nil {
		return
	}
	if kind != inv_v1.ResourceKind_RESOURCE_KIND_SINGLESCHEDULE &&
		kind != inv_v1.ResourceKind_RESOURCE_KIND_REPEATEDSCHEDULE &&
		kind != inv_v1.ResourceKind_RESOURCE_KIND_HOST {
		log.InfraSec().InfraError("Unexpected resource kind in manageEvent: eventKind=%s", kind)
		return
	}
	tID, rID, err := inv_utils.GetResourceKeyFromResource(event.GetResource())
	if err != nil {
		log.InfraSec().Err(err).Msgf("Got resource without tenantID and/or resourceID: %v", event.GetResource())
	}
	sc.reconcileResource(tID, rID, event.EventKind)
}

func (sc *InvCollectorClient) getResource(tenantID, resourceID string) (*inv_v1.Resource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := sc.InvClient.Get(ctx, tenantID, resourceID)
	if err != nil {
		return nil, err
	}
	return res.GetResource(), nil
}

func (sc *InvCollectorClient) reconcileResource(tenantID, resourceID string, evKind inv_v1.SubscribeEventsResponse_EventKind) {
	sc.Cache.lock.Lock()
	defer sc.Cache.lock.Unlock()

	if evKind == inv_v1.SubscribeEventsResponse_EVENT_KIND_DELETED {
		// Do not query inventory when deleting, resource won't be there

		resourceKind, err := inv_utils.GetResourceKindFromResourceID(resourceID)
		if err != nil {
			log.InfraSec().InfraErr(err).
				Msgf("Failed to reconcile resource unknown kind: tenantID=%s, resourceID=%s", tenantID, resourceID)
			return
		}

		if resourceKind == inv_v1.ResourceKind_RESOURCE_KIND_HOST {
			delete(sc.Cache.Hosts, resourceID)
		}

		return
	}

	// Not a DELETE
	res, err := sc.getResource(tenantID, resourceID)
	if err != nil {
		log.InfraSec().InfraErr(err).Msgf("Failed to reconcile resource: tenantID=%s, resourceID=%s", tenantID, resourceID)
		return
	}

	switch evKind {
	case inv_v1.SubscribeEventsResponse_EVENT_KIND_CREATED,
		inv_v1.SubscribeEventsResponse_EVENT_KIND_UPDATED:
		if inv_utils.GetResourceKindFromResource(res) == inv_v1.ResourceKind_RESOURCE_KIND_HOST {
			sc.Cache.Hosts[resourceID] = res
		}
	default:
		log.InfraError("Unexpected event while reconciling resource: tenantID=%s, resourceID=%s, evKind=%s",
			tenantID, resourceID, evKind)
	}
}

func (col *InventoryCollector) loadSingleSchedulesFromCache(
	ctx context.Context,
	hostID, timestampString, tenantID *string,
) []*sched_v1.SingleScheduleResource {
	allSingleSchedules := make([]*sched_v1.SingleScheduleResource, 0)
	hasNext := true
	offset := 0
	limit := batchSize

	for hasNext {
		filters := new(schedule_cache.Filters).
			Add(schedule_cache.HasHostID(hostID)).
			Add(schedule_cache.FilterByTS(timestampString))
		sScheds, respHasNext, _, err := col.HScheduleCache.GetSingleSchedules(ctx, *tenantID, offset, limit, filters)
		if err != nil {
			log.InfraErr(err).Msg("Failed to get single schedules from inventory.")
			break
		}

		if len(sScheds) == 0 {
			log.Debug().Msg("No more single schedules in Inventory.")
			break
		}

		allSingleSchedules = append(allSingleSchedules, sScheds...)

		hasNext = respHasNext
		limit += batchSize
		offset += batchSize
	}

	return allSingleSchedules
}

func (col *InventoryCollector) loadRepeatedSchedulesFromCache(
	ctx context.Context,
	hostID, timestampString, tenantID *string,
) []*sched_v1.RepeatedScheduleResource {
	allRepeatedSchedules := make([]*sched_v1.RepeatedScheduleResource, 0)
	hasNext := true
	offset := 0
	limit := batchSize

	for hasNext {
		filters := new(schedule_cache.Filters).
			Add(schedule_cache.HasHostID(hostID)).
			Add(schedule_cache.FilterByTS(timestampString))
		rScheds, respHasNext, _, err := col.HScheduleCache.GetRepeatedSchedules(ctx, *tenantID, offset, limit, filters)
		if err != nil {
			log.InfraErr(err).Msg("Failed to get repeated schedules from inventory.")
			break
		}

		if len(rScheds) == 0 {
			log.Debug().Msg("No more reeated schedules in Inventory.")
			break
		}

		allRepeatedSchedules = append(allRepeatedSchedules, rScheds...)

		hasNext = respHasNext
		limit += batchSize
		offset += batchSize
	}

	return allRepeatedSchedules
}

func loadResourceFromInv(
	ctx context.Context,
	invClient client.TenantAwareInventoryClient,
	resourceType inv_v1.ResourceKind,
) map[string]*inv_v1.Resource {
	resources := make(map[string]*inv_v1.Resource, 0)
	resFilter, err := inv_utils.GetResourceFromKind(resourceType)
	if err != nil {
		log.InfraErr(err).Msgf("Failed to get %v resources from inventory.", resourceType)
		return resources
	}
	// Load all across tenants.
	filterRequest := inv_v1.ResourceFilter{
		Resource: resFilter,
		Filter:   "", // Empty filter, get all resources for the given type
		Limit:    batchSize,
		Offset:   0,
	}
	hasNext := true
	for hasNext {
		listResponse, err := invClient.List(ctx, &filterRequest)
		if err != nil {
			log.InfraErr(err).Msgf("Failed to %v resources from inventory.", resources)
			break
		}
		if len(listResponse.GetResources()) == 0 {
			log.Debug().Msgf("No more %v resources in Inventory.", resourceType)
			break
		}
		var wrongResourceType inv_v1.ResourceKind
		for _, res := range listResponse.GetResources() {
			wrongResourceType = inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED
			switch res.GetResource().GetResource().(type) {
			case *inv_v1.Resource_Host:
				if resourceType == inv_v1.ResourceKind_RESOURCE_KIND_HOST {
					resources[res.GetResource().GetHost().ResourceId] = res.GetResource()
				} else {
					wrongResourceType = inv_v1.ResourceKind_RESOURCE_KIND_HOST
				}
			default:
				log.InfraSec().InfraError("Unsupported resource type: %v", resourceType)
			}
			if wrongResourceType != inv_v1.ResourceKind_RESOURCE_KIND_UNSPECIFIED {
				// We should never reach this point
				log.InfraSec().InfraError("Got wrong resource type: expected=%v, got=%v", resourceType, wrongResourceType)
			}
		}
		hasNext = listResponse.HasNext
		filterRequest.Limit += batchSize
		filterRequest.Offset += batchSize
	}
	return resources
}

func (sc *InvCollectorClient) LoadAllHostsFromInv() {
	log.Debug().Msgf("LoadAllHostsFromInv")
	// TODO Current reconciliation is dumb, clean everything in the local cache and re-build it
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutRPCCalls)
	defer cancel()

	// Load all host across tenants.
	hosts := loadResourceFromInv(ctx, sc.InvClient, inv_v1.ResourceKind_RESOURCE_KIND_HOST)

	sc.Cache.lock.Lock()
	defer sc.Cache.lock.Unlock()
	sc.Cache.Hosts = hosts
}

// getHosts return a cloned copy of each one of the host
// resources in the cache.
// The clone of the messages is done to avoid race conditions
// where getHosts is used, in case access to host info is needed.
func (sc *InvCollectorClient) getHosts() []*inv_v1.Resource {
	hosts := make([]*inv_v1.Resource, 0)
	for _, host := range sc.Cache.Hosts {
		hosts = append(hosts, host)
	}
	return hosts
}

// construct Host Struct for the host.
func constructHostStruct(hostHasSched bool, host *computev1.HostResource) kpis.HostStatus {
	instance := host.GetInstance()
	provisionStatus := ""
	if instance != nil {
		provisionStatus = instance.GetProvisioningStatus()
	}
	updateStatus := ""
	if instance != nil {
		updateStatus = instance.GetUpdateStatus()
	}

	return kpis.HostStatus{
		HostID:             host.GetResourceId(),
		DeviceGUID:         host.GetUuid(),
		Name:               host.GetName(),
		Hostname:           host.GetHostname(),
		Serial:             host.GetSerialNumber(),
		TenantID:           host.GetTenantId(),
		Status:             host.GetHostStatus(),
		ProvisioningStatus: provisionStatus,
		OnboardingStatus:   host.GetOnboardingStatus(),
		UpdateStatus:       updateStatus,
		HasSchedule:        hostHasSched,
	}
}

// calculate Total Provisioning Time for the host.
func calculateTotalProvisioningTime(ctx context.Context, host *computev1.HostResource) kpis.HostProvisioningTime {
	var totalProvisioningTime float64
	hostID := host.GetResourceId()
	if totalProvisioningTimevalue, ok := totalProvisioningTimeMap[hostID]; ok {
		totalProvisioningTime = totalProvisioningTimevalue
	} else {
		totalProvisioningTime = fetchTotalProvisioningTime(ctx, host.GetUuid(), hostID, host.GetTenantId())
		if totalProvisioningTime != 0 {
			totalProvisioningTimeMap[hostID] = totalProvisioningTime
		}
	}
	return kpis.HostProvisioningTime{
		HostID:                hostID,
		DeviceGUID:            host.GetUuid(),
		Name:                  host.GetName(),
		Hostname:              host.GetHostname(),
		Serial:                host.GetSerialNumber(),
		TenantID:              host.GetTenantId(),
		TotalProvisioningTime: totalProvisioningTime,
	}
}

func fetchTotalProvisioningTime(ctx context.Context, deviceGUID, hostID, projectID string) float64 {
	log.Debug().Msgf("fetchTotalProvisioningTime for deviceGUID: %s, hostID: %s,projectID: %s", deviceGUID, hostID, projectID)

	timeA2_2, timeA3_13_1, timeA3_13_2, timeA3_13_3, timeA3_13_4 := startAndEndTimeKPI(ctx, deviceGUID, hostID, projectID)
	var totalProvisioningTime float64

	var startA2_2, endA3_13 int64
	// Finding start time from a2_2 file
	if timeA2_2 != "" {
		var err error
		startA2_2, err = strconv.ParseInt(timeA2_2, 10, 64)
		if err != nil {
			log.Error().Err(err).Msg("Error calculating start_A2_2 ")
			return totalProvisioningTime
		}
		log.Debug().Msgf("start_A2_2 :%d ", startA2_2)
	} else {
		log.Error().Msg("Not able to find start_A2_2 time or end A3_13 time")
		return totalProvisioningTime
	}

	// Finding Endtime from a3_13 file
	if timeA3_13_1 != "" || timeA3_13_2 != "" || timeA3_13_3 != "" || timeA3_13_4 != "" {
		endA3_13 = maxTime(timeA3_13_1, timeA3_13_2, timeA3_13_3, timeA3_13_4)
		log.Debug().Msgf("end_A3_13 :%d ", endA3_13)
	} else {
		log.Error().Msg("Not able to find start_A2_2 time or end A3_13 time")
		return totalProvisioningTime
	}

	totalProvisioningTime = math.Abs(float64(endA3_13 - startA2_2))
	log.Debug().Msgf("KPI Total Time: %f Seconds\n", totalProvisioningTime)
	return totalProvisioningTime
}

func startAndEndTimeKPI(
	ctx context.Context,
	deviceGUID, hostID, projectID string,
) (timeA2_2 string, timeA3_13_1 string,
	timeA3_13_2 string, timeA3_13_3 string,
	timeA3_13_4 string,
) {
	// Getting data for last 10 hours
	startTime := time.Now().Add(-time.Duration(logretension) * time.Minute).UnixNano()
	endTime := time.Now().UnixNano()

	// query strings for searching starttime in file uOS_bootkitLogs -- A2_2
	queryA2_2 := `{file_type="uOS_bootkitLogs", host_guid="` + deviceGUID + `"} |= "s_netconf_start"`
	patternA2_2 := `.*s_netconf_start":(\d+)`

	// query strings for searching endtime in file ClusterAgent -- A3_13_1
	queryA3_13_1 := []string{
		`{file_type="ClusterAgent", host_guid="` + deviceGUID,
		`"} |= "Edge Cluster Manager response" |= "` + hostID + `"`,
	}
	patternA3_13_1 := `ClusterAgent.*time.*?(20.*Z).*Edge Cluster Manager response:`

	// query strings for searching endtime in file PlatformUpdateAgent -- A3_13_2
	queryA3_13_2 := []string{
		`{file_type="PlatformUpdateAgent", host_guid="` + deviceGUID,
		`"} |= "PlatformUpdateStatusRequest sent successfully" |= "` + hostID + `"`,
	}
	patternA3_13_2 := `PlatformUpdateAgent.*time.*?(20.*Z).*PlatformUpdateStatusRequest sent successfully`

	// query strings for searching endtime in file Platform_Telemetry_Agent -- A3_13_3
	queryA3_13_3 := []string{
		`{file_type="Platform_Telemetry_Agent", host_guid="` + deviceGUID,
		`"} |= "from Telemetry Manager" |= "` + hostID + `"`,
	}
	patternA3_13_3 := `Platform_Telemetry_Agent.*time.*?(20.*Z).*from Telemetry Manager`

	// query strings for searching endtime in file HardwareAgent -- A3_13_4
	queryA3_13_4 := []string{
		`{file_type="HardwareAgent", host_guid="` + deviceGUID,
		`"} |= "UpdateHostSystemInfoByGUIDRequest sent successfully" |= "` + hostID + `"`,
	}
	patternA3_13_4 := `HardwareAgent.*time.*?(20.*Z).*UpdateHostSystemInfoByGUIDRequest sent successfully`

	wg := sync.WaitGroup{}
	wg.Add(totalNumberOFCurlReq)
	go executeCurl(ctx, &timeA2_2, projectID, queryA2_2, startTime, endTime, &wg, patternA2_2, true)
	go executeCurl(ctx, &timeA3_13_1, projectID, queryA3_13_1[0]+queryA3_13_1[1], startTime,
		endTime, &wg, patternA3_13_1, false)
	go executeCurl(ctx, &timeA3_13_2, projectID, queryA3_13_2[0]+queryA3_13_2[1], startTime,
		endTime, &wg, patternA3_13_2, false)
	go executeCurl(ctx, &timeA3_13_3, projectID, queryA3_13_3[0]+queryA3_13_3[1], startTime,
		endTime, &wg, patternA3_13_3, false)
	go executeCurl(ctx, &timeA3_13_4, projectID, queryA3_13_4[0]+queryA3_13_4[1], startTime,
		endTime, &wg, patternA3_13_4, false)
	wg.Wait()

	return timeA2_2, timeA3_13_1, timeA3_13_2, timeA3_13_3, timeA3_13_4
}

func executeCurl(
	ctx context.Context, output *string,
	projectID, query string,
	startTime, endTime int64,
	wg *sync.WaitGroup, pattern string,
	lastMatch bool, // set to true to use the last matching string if multiple exist, otherwise the first match will be used.
) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	defer wg.Done()
	baseURL := "http://" + env.ENLokiURL + "/loki/api/v1/query_range"
	headers := map[string]string{
		"Accept":        "application/json",
		"X-Scope-OrgID": projectID,
	}

	params := url.Values{}
	params.Add("direction", "forward")
	params.Add("start", strconv.FormatInt(startTime, 10))
	params.Add("end", strconv.FormatInt(endTime, 10))
	params.Add("query", query)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, http.NoBody)
	if err != nil {
		log.Error().Err(err).Msg("error creating request")
		return
	}

	// Add headers
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Add query parameters to the URL
	req.URL.RawQuery = params.Encode()

	// Make the request
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Error().Err(err).Msg("error making request")
		return
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		log.Error().Msgf("http request failed with status: %s", resp.Status)
		return
	}

	// Parse the JSON response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error().Err(err).Msg("error decoding JSON response")
		return
	}

	findTimeFromResponse(result, output, pattern, lastMatch)
}

func findTimeFromResponse(result map[string]interface{}, output *string, pattern string, lastMatch bool) {
	// Extract and print the desired values
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		log.Error().Msg("error parsing data field")
		return
	}

	results, ok := data["result"].([]interface{})
	if !ok {
		log.Error().Msg("error parsing result field")
		return
	}

	var builder strings.Builder
	for _, r := range results {
		values, ok := r.(map[string]interface{})["values"].([]interface{})
		if !ok {
			log.Error().Msg("error parsing values field")
			return
		}

		for _, v := range values {
			valuePair, ok := v.([]interface{})
			if !ok || len(valuePair) != 2 {
				log.Error().Msg("error parsing value pair")
				return
			}

			valuePairString, ok := valuePair[1].(string)
			if !ok {
				log.Error().Msg("error parsing value pair string")
				return
			}
			builder.WriteString(valuePairString)
			builder.WriteString("\n")
		}
	}

	// Remove slash and back slash from the string
	outputString := strings.ReplaceAll(builder.String(), "/", "")
	outputString = strings.ReplaceAll(outputString, "\\", "")
	findFirstOrLastMatch(outputString, pattern, output, lastMatch)
}

func findFirstOrLastMatch(outputString, pattern string, output *string, lastMatch bool) {
	if lastMatch {
		// fetching the last time from the response
		outputmatch := regexp.MustCompile(pattern).FindAllStringSubmatch(outputString, -1)

		if len(outputmatch) > 0 {
			*output = outputmatch[len(outputmatch)-1][1]
		}
	} else {
		// fetching the first time from the response
		outputmatch := regexp.MustCompile(pattern).FindStringSubmatch(outputString)

		if len(outputmatch) > 0 {
			*output = outputmatch[1]
		}
	}
}

func maxTime(dateStrings ...string) int64 {
	var maxTime time.Time
	for _, dateStr := range dateStrings {
		// Parse the date string using the RFC3339 layout
		if dateStr == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			log.Error().Err(err).Msgf("Error calculating end_A3_13 %s", dateStr)
			return 0
		}
		// Compare the parsed time with the current max time
		if t.After(maxTime) {
			maxTime = t
		}
	}
	return maxTime.Unix()
}
