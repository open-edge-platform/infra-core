// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

var log = logging.GetLogger("tests")

var (
	emptyString      = ""
	emptyStringWrong = " "
)

const (
	testTimeout              = time.Duration(120) * time.Second
	jwtToken                 = "JWT_TOKEN"
	authKey                  = "authorization"
	projectID                = "PROJECT_ID"
	projectIDKey             = "ActiveProjectID"
	userAgent                = "User-Agent"
	ecmServiceName           = "ecm-api"
	observabilityServiceName = "common-metric-query-metrics"
)

var (
	apiUrl = flag.String("apiurl", "http://localhost:8080", "The URL of the REST API")
	caPath = flag.String("caPath", "", "The path to the CA certificate file of the target cluster")
)

var (
	FilterUUID                 = `uuid = %q`
	FilterSiteId               = `site.resource_id = %q`
	FilterNotHasSite           = "NOT has(site)"
	FilterByMetadata           = `metadata = '%s'`
	FilterByWorkloadMemberId   = `workload_members.resource_id = %q`
	FilterNotHasWorkloadMember = "NOT has(workload_members)"
	FilterHasWorkloadMember    = "has(workload_members)"
	FilterRegionParentId       = `parent_region.resource_id = %q`
	FilterRegionNotHasParent   = "NOT has(parent_region)"
	FilterSiteRegionId         = `region.resource_id = %q`
	FilterSiteNotHasRegion     = "NOT has(region)"
)

func LoadFile(filePath string) (string, error) {
	dirFile, err := filepath.Abs(filePath)
	if err != nil {
		log.Err(err).Msgf("failed LoadFile, filepath unexistent %s", filePath)
		return "", err
	}

	dataBytes, err := os.ReadFile(dirFile)
	if err != nil {
		log.Err(err).Msgf("failed to read file %s", dirFile)
		return "", err
	}

	dataStr := string(dataBytes)
	return dataStr, nil
}

func GetClientWithCA(caPath string) (*http.Client, error) {
	caCert, err := LoadFile(caPath)
	if err != nil {
		log.Warn().Msg("CA cert not provided, using httpclient insecure client")
		return &http.Client{}, nil
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM([]byte(caCert))
	if !ok {
		err := fmt.Errorf("failed to parse CA cert into http client")
		return nil, err
	}
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	return &http.Client{
		Transport: transport,
	}, nil
}

func GetAPIClient() (*api.ClientWithResponses, error) {
	httpClient, err := GetClientWithCA(*caPath)
	if err != nil {
		return nil, err
	}

	client, err := api.NewClientWithResponses(*apiUrl, api.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func ListStringContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func ListMetadataContains(lst []api.MetadataItem, key, value string) bool {
	for _, v := range lst {
		if v.Key == key && v.Value == value {
			return true
		}
	}

	return false
}

func AddJWTtoTheHeader(ctx context.Context, req *http.Request) error {
	// extract token from the environment variable
	jwtTokenStr, ok := os.LookupEnv(jwtToken)
	if !ok {
		return fmt.Errorf("can't find a \"JWT_TOKEN\" variable, please set it in your environment")
	}

	req.Header.Add(authKey, "Bearer "+jwtTokenStr)

	return nil
}

func AddProjectIDtoTheHeader(ctx context.Context, req *http.Request) error {
	// extract MT ProjectID from the environment variable
	projectIDStr, ok := os.LookupEnv(projectID)
	if !ok {
		return fmt.Errorf("can't find a \"%s\" variable, please set it in your environment", projectID)
	}

	req.Header.Add(projectIDKey, projectIDStr)

	return nil
}

func hostsContainsId(hosts []api.HostResource, hostID string) bool {
	for _, h := range hosts {
		if *h.ResourceId == hostID {
			return true
		}
	}
	return false
}

func CreateSchedSingle(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	reqSched api.SingleScheduleResource,
) *api.ScheduleServiceCreateSingleScheduleResponse {
	t.Helper()

	sched, err := apiClient.ScheduleServiceCreateSingleScheduleWithResponse(
		ctx,
		reqSched,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, sched.StatusCode())

	t.Cleanup(func() { DeleteSchedSingle(t, context.Background(), apiClient, *sched.JSON200.ResourceId) })
	return sched
}

func DeleteSchedSingle(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	schedID string,
) {
	t.Helper()

	schedDel, err := apiClient.ScheduleServiceDeleteSingleScheduleWithResponse(
		ctx,
		schedID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, schedDel.StatusCode())
}

func CreateSchedRepeated(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	reqSched api.RepeatedScheduleResource,
) *api.ScheduleServiceCreateRepeatedScheduleResponse {
	t.Helper()

	sched, err := apiClient.ScheduleServiceCreateRepeatedScheduleWithResponse(
		ctx,
		reqSched,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, sched.StatusCode())

	t.Cleanup(func() { DeleteSchedRepeated(t, context.Background(), apiClient, *sched.JSON200.ResourceId) })
	return sched
}

func DeleteSchedRepeated(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	schedID string,
) {
	t.Helper()

	schedDel, err := apiClient.ScheduleServiceDeleteRepeatedScheduleWithResponse(
		ctx,
		schedID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, schedDel.StatusCode())
}

func CreateRegion(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	regionRequest api.RegionResource,
) *api.RegionServiceCreateRegionResponse {
	t.Helper()

	region, err := apiClient.RegionServiceCreateRegionWithResponse(ctx, regionRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, region.StatusCode())

	t.Cleanup(func() { DeleteRegion(t, context.Background(), apiClient, *region.JSON200.ResourceId) })
	return region
}

func DeleteRegion(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	regionID string,
) {
	t.Helper()

	resDelRegion, err := apiClient.RegionServiceDeleteRegionWithResponse(
		ctx,
		regionID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resDelRegion.StatusCode())
}

func CreateSite(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	siteRequest api.SiteResource,
) *api.SiteServiceCreateSiteResponse {
	t.Helper()

	site, err := apiClient.SiteServiceCreateSiteWithResponse(ctx, siteRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, site.StatusCode())

	t.Cleanup(func() { DeleteSite(t, context.Background(), apiClient, *site.JSON200.ResourceId) })
	return site
}

func DeleteSite(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	siteID string,
) {
	t.Helper()

	resDelSite, err := apiClient.SiteServiceDeleteSiteWithResponse(ctx, siteID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resDelSite.StatusCode())
}

// CreateHost adds a host via the REST APIs, and setup the soft delete upon test cleanup.
func CreateHost(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	hostRequest api.HostResource,
) *api.HostServiceCreateHostResponse {
	t.Helper()

	host, err := apiClient.HostServiceCreateHostWithResponse(ctx, hostRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, host.StatusCode())
	t.Cleanup(func() { SoftDeleteHost(t, context.Background(), apiClient, host.JSON200) })
	return host
}

// SoftDeleteHost: unallocate the host if allocated to any site so we free any linked resources (site), and does a soft delete of Host.
// Eventually Host Resource Manager will do the hard delete.
func SoftDeleteHost(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	host *api.HostResource,
) {
	t.Helper()

	UnallocateHostFromSite(t, ctx, apiClient, host)
	resDelHost, err := apiClient.HostServiceDeleteHostWithResponse(
		ctx,
		*host.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resDelHost.StatusCode())
}

// UnallocateHostFromSite: unallocate the given hostId from a site.
func UnallocateHostFromSite(t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, hostReq *api.HostResource) {
	t.Helper()

	hostUp := api.HostResource{
		Name:   hostReq.Name,
		SiteId: &emptyString,
	}
	res, err := apiClient.HostServiceUpdateHostWithResponse(
		ctx,
		*hostReq.ResourceId,
		hostUp,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
}

func AssertInMaintenance(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	hostID *string,
	siteID *string,
	regionID *string,
	timestamp time.Time,
	expectedSchedules int,
	found bool,
) {
	t.Helper()

	timestampString := fmt.Sprint(timestamp.UTC().Unix())
	sReply, err := apiClient.ScheduleServiceListSchedulesWithResponse(
		ctx,
		&api.ScheduleServiceListSchedulesParams{
			HostId:    hostID,
			SiteId:    siteID,
			RegionId:  regionID,
			UnixEpoch: &timestampString,
		},
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	if found {
		assert.Equal(t, http.StatusOK, sReply.StatusCode())
		length := 0
		if sReply.JSON200.SingleSchedules != nil {
			length += len(sReply.JSON200.SingleSchedules)
		}
		if sReply.JSON200.RepeatedSchedules != nil {
			length += len(sReply.JSON200.RepeatedSchedules)
		}
		assert.Equal(t, expectedSchedules, length, "Wrong number of schedules")
	} else {
		assert.Equal(t, http.StatusOK, sReply.StatusCode())
	}
}

func CreateOS(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	reqOS api.OperatingSystemResource,
) *api.OperatingSystemServiceCreateOperatingSystemResponse {
	t.Helper()

	osCreated, err := apiClient.OperatingSystemServiceCreateOperatingSystemWithResponse(
		ctx,
		reqOS,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, osCreated.StatusCode())

	t.Cleanup(func() {
		time.Sleep(2 * time.Second) // Waits until Instance reconciliation happens
		DeleteOS(t, context.Background(), apiClient, *osCreated.JSON200.ResourceId)
	})

	return osCreated
}

func DeleteOS(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	osID string,
) {
	t.Helper()

	osDel, err := apiClient.OperatingSystemServiceDeleteOperatingSystemWithResponse(
		ctx,
		osID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, osDel.StatusCode())
}

func CreateWorkload(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	reqWorkload api.WorkloadResource,
) *api.WorkloadServiceCreateWorkloadResponse {
	t.Helper()

	wCreated, err := apiClient.WorkloadServiceCreateWorkloadWithResponse(
		ctx,
		reqWorkload,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, wCreated.StatusCode())

	t.Cleanup(func() { DeleteWorkload(t, context.Background(), apiClient, *wCreated.JSON200.ResourceId) })
	return wCreated
}

func DeleteWorkload(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	workloadID string,
) {
	t.Helper()

	wDel, err := apiClient.WorkloadServiceDeleteWorkloadWithResponse(
		ctx,
		workloadID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, wDel.StatusCode())
}

func CreateWorkloadMember(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	reqMember api.WorkloadMember,
) *api.WorkloadMemberServiceCreateWorkloadMemberResponse {
	t.Helper()

	mCreated, err := apiClient.WorkloadMemberServiceCreateWorkloadMemberWithResponse(
		ctx,
		reqMember,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, mCreated.StatusCode())

	t.Cleanup(func() { DeleteWorkloadMember(t, context.Background(), apiClient, *mCreated.JSON200.ResourceId) })
	return mCreated
}

func DeleteWorkloadMember(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	memberID string,
) {
	t.Helper()

	mDel, err := apiClient.WorkloadMemberServiceDeleteWorkloadMemberWithResponse(
		ctx,
		memberID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, mDel.StatusCode())
}

func CreateInstance(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	instRequest api.InstanceResource,
) *api.InstanceServiceCreateInstanceResponse {
	t.Helper()

	createdInstance, err := apiClient.InstanceServiceCreateInstanceWithResponse(ctx, instRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, createdInstance.StatusCode())

	t.Cleanup(func() { DeleteInstance(t, context.Background(), apiClient, *createdInstance.JSON200.ResourceId) })
	return createdInstance
}

func DeleteInstance(t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, instanceID string) {
	t.Helper()

	resDelInst, err := apiClient.InstanceServiceDeleteInstanceWithResponse(ctx, instanceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resDelInst.StatusCode())
}

func CreateTelemetryLogsGroup(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	request api.TelemetryLogsGroupResource,
) *api.TelemetryLogsGroupServiceCreateTelemetryLogsGroupResponse {
	t.Helper()

	created, err := apiClient.TelemetryLogsGroupServiceCreateTelemetryLogsGroupWithResponse(ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, created.StatusCode())

	t.Cleanup(func() {
		DeleteTelemetryLogsGroup(t, context.Background(), apiClient, *created.JSON200.ResourceId)
	})
	return created
}

func DeleteTelemetryLogsGroup(
	t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, id string,
) {
	t.Helper()

	res, err := apiClient.TelemetryLogsGroupServiceDeleteTelemetryLogsGroupWithResponse(ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
}

func CreateTelemetryMetricsGroup(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	request api.TelemetryMetricsGroupResource,
) *api.TelemetryMetricsGroupServiceCreateTelemetryMetricsGroupResponse {
	t.Helper()

	created, err := apiClient.TelemetryMetricsGroupServiceCreateTelemetryMetricsGroupWithResponse(ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, created.StatusCode())

	t.Cleanup(func() {
		DeleteTelemetryMetricsGroup(t, context.Background(), apiClient, *created.JSON200.ResourceId)
	})
	return created
}

func DeleteTelemetryMetricsGroup(
	t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, id string,
) {
	t.Helper()

	res, err := apiClient.TelemetryMetricsGroupServiceDeleteTelemetryMetricsGroupWithResponse(ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
}

func CreateTelemetryLogsProfile(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	request api.TelemetryLogsProfileResource,
) *api.TelemetryLogsProfileServiceCreateTelemetryLogsProfileResponse {
	t.Helper()

	created, err := apiClient.TelemetryLogsProfileServiceCreateTelemetryLogsProfileWithResponse(ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, created.StatusCode())

	t.Cleanup(func() {
		DeleteTelemetryLogsProfile(t, context.Background(), apiClient, *created.JSON200.ResourceId)
	})
	return created
}

func DeleteTelemetryLogsProfile(
	t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, id string,
) {
	t.Helper()

	res, err := apiClient.TelemetryLogsProfileServiceDeleteTelemetryLogsProfileWithResponse(ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
}

func CreateTelemetryMetricsProfile(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	request api.TelemetryMetricsProfileResource,
) *api.TelemetryMetricsProfileServiceCreateTelemetryMetricsProfileResponse {
	t.Helper()

	created, err := apiClient.TelemetryMetricsProfileServiceCreateTelemetryMetricsProfileWithResponse(ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, created.StatusCode())

	t.Cleanup(func() {
		DeleteTelemetryMetricsProfile(t, context.Background(), apiClient, *created.JSON200.ResourceId)
	})
	return created
}

func DeleteTelemetryMetricsProfile(
	t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, id string,
) {
	t.Helper()

	res, err := apiClient.TelemetryMetricsProfileServiceDeleteTelemetryMetricsProfileWithResponse(ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode())
}

func CreateProvider(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	reqProvider api.ProviderResource,
) *api.ProviderServiceCreateProviderResponse {
	t.Helper()

	providerCreated, err := apiClient.ProviderServiceCreateProviderWithResponse(
		ctx,
		reqProvider,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, providerCreated.StatusCode())

	t.Cleanup(func() { DeleteProvider(t, context.Background(), apiClient, *providerCreated.JSON200.ResourceId) })
	return providerCreated
}

func DeleteProvider(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	providerID string,
) {
	t.Helper()

	providerDel, err := apiClient.ProviderServiceDeleteProviderWithResponse(
		ctx,
		providerID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, providerDel.StatusCode())
}

// CreateLocalAccount creates a LocalAccount resource and returns the response.
func CreateLocalAccount(t *testing.T, ctx context.Context, apiClient *api.ClientWithResponses,
	request api.LocalAccountResource,
) *api.LocalAccountServiceCreateLocalAccountResponse {
	t.Helper()

	response, err := apiClient.LocalAccountServiceCreateLocalAccountWithResponse(
		ctx,
		request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.NotNil(t, response.JSON200)
	require.NotNil(t, response.JSON200.ResourceId)

	return response
}

// DeleteLocalAccount deletes a LocalAccount resource by its ID.
func DeleteLocalAccount(t *testing.T, ctx context.Context,
	apiClient *api.ClientWithResponses, resourceID string,
) {
	t.Helper()

	response, err := apiClient.LocalAccountServiceDeleteLocalAccountWithResponse(
		ctx,
		resourceID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, response.StatusCode())
}

func GetHostRequestWithRandomUUID() api.HostResource {
	uuidHost := uuid.New().String()
	randName := fmt.Sprintf("Test Host %d", rand.Uint32())
	return api.HostResource{
		Name: randName,
		Uuid: &uuidHost,
	}
}
