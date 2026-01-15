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
	testTimeout  = time.Duration(120) * time.Second
	jwtToken     = "JWT_TOKEN"
	authKey      = "authorization"
	projectID    = "PROJECT_ID"
	projectIDKey = "ActiveProjectID"
	sleepTime    = 2 * time.Second
)

var (
	apiURL = flag.String("apiurl", "http://localhost:8080", "The URL of the REST API")
	caPath = flag.String("caPath", "", "The path to the CA certificate file of the target cluster")
)

var (
	FilterUUID                 = `uuid = %q`
	FilterSiteID               = `site.resource_id = %q`
	FilterNotHasSite           = "NOT has(site)"
	FilterByMetadata           = `metadata = '%s'`
	FilterByWorkloadMemberID   = `workload_members.resource_id = %q`
	FilterNotHasWorkloadMember = "NOT has(workload_members)"
	FilterHasWorkloadMember    = "has(workload_members)"
	FilterRegionParentID       = `parent_region.resource_id = %q`
	FilterRegionNotHasParent   = "NOT has(parent_region)"
	FilterSiteRegionID         = `region.resource_id = %q`
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
		//nolint:nilerr // If CA cert is not provided, we return an insecure http client
		return &http.Client{}, nil
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM([]byte(caCert))
	if !ok {
		err := fmt.Errorf("failed to parse CA cert into http client")
		return nil, err
	}
	//nolint:gosec // G402: TLS InsecureSkipVerify set to true is not a security issue in tests
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

	client, err := api.NewClientWithResponses(*apiURL, api.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func ListMetadataContains(lst []api.MetadataItem, key, value string) bool {
	for _, v := range lst {
		if v.Key == key && v.Value == value {
			return true
		}
	}

	return false
}

func AddJWTtoTheHeader(_ context.Context, req *http.Request) error {
	// extract token from the environment variable
	jwtTokenStr, ok := os.LookupEnv(jwtToken)
	if !ok {
		return fmt.Errorf("can't find a \"JWT_TOKEN\" variable, please set it in your environment")
	}

	req.Header.Add(authKey, "Bearer "+jwtTokenStr)

	return nil
}

func AddProjectIDtoTheHeader(_ context.Context, req *http.Request) error {
	// extract MT ProjectID from the environment variable
	projectIDStr, ok := os.LookupEnv(projectID)
	if !ok {
		return fmt.Errorf("can't find a \"%s\" variable, please set it in your environment", projectID)
	}

	req.Header.Add(projectIDKey, projectIDStr)

	return nil
}

func hostsContainsID(hosts []api.HostResource, hostID string) bool {
	for _, h := range hosts {
		if *h.ResourceId == hostID {
			return true
		}
	}
	return false
}

func CreateSchedSingle(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	reqSched api.SingleScheduleResource,
) *api.ScheduleServiceCreateSingleScheduleResponse {
	tb.Helper()

	sched, err := apiClient.ScheduleServiceCreateSingleScheduleWithResponse(
		ctx,
		reqSched,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, sched.StatusCode())
	require.NotNil(tb, sched.JSON200, "SingleSchedule creation returned nil JSON200")
	require.NotNil(tb, sched.JSON200.ResourceId, "SingleSchedule creation returned nil ResourceId")
	tb.Cleanup(func() { DeleteSchedSingle(context.Background(), tb, apiClient, *sched.JSON200.ResourceId) })

	return sched
}

func DeleteSchedSingle(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	schedID string,
) {
	tb.Helper()

	schedDel, err := apiClient.ScheduleServiceDeleteSingleScheduleWithResponse(
		ctx,
		schedID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, schedDel.StatusCode())
}

func CreateSchedRepeated(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	reqSched api.RepeatedScheduleResource,
) *api.ScheduleServiceCreateRepeatedScheduleResponse {
	tb.Helper()

	sched, err := apiClient.ScheduleServiceCreateRepeatedScheduleWithResponse(
		ctx,
		reqSched,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, sched.StatusCode())

	require.NotNil(tb, sched.JSON200, "RepeatedSchedule creation returned nil JSON200")
	require.NotNil(tb, sched.JSON200.ResourceId, "RepeatedSchedule creation returned nil ResourceId")
	tb.Cleanup(func() { DeleteSchedRepeated(context.Background(), tb, apiClient, *sched.JSON200.ResourceId) })

	return sched
}

func DeleteSchedRepeated(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	schedID string,
) {
	tb.Helper()

	schedDel, err := apiClient.ScheduleServiceDeleteRepeatedScheduleWithResponse(
		ctx,
		schedID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, schedDel.StatusCode())
}

func CreateRegion(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	regionRequest api.RegionResource,
) *api.RegionServiceCreateRegionResponse {
	tb.Helper()

	region, err := apiClient.RegionServiceCreateRegionWithResponse(ctx, regionRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, region.StatusCode())
	require.NotNil(tb, region.JSON200, "Region creation returned nil JSON200")
	require.NotNil(tb, region.JSON200.ResourceId, "Region creation returned nil ResourceId")

	tb.Cleanup(func() { DeleteRegion(context.Background(), tb, apiClient, *region.JSON200.ResourceId) })
	return region
}

func DeleteRegion(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	regionID string,
) {
	tb.Helper()

	resDelRegion, err := apiClient.RegionServiceDeleteRegionWithResponse(
		ctx,
		regionID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, resDelRegion.StatusCode())
}

func CreateSite(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	siteRequest api.SiteResource,
) *api.SiteServiceCreateSiteResponse {
	tb.Helper()

	site, err := apiClient.SiteServiceCreateSiteWithResponse(ctx, siteRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, site.StatusCode())
	require.NotNil(tb, site.JSON200, "Site creation returned nil JSON200")
	require.NotNil(tb, site.JSON200.ResourceId, "Site creation returned nil ResourceId")

	tb.Cleanup(func() { DeleteSite(context.Background(), tb, apiClient, *site.JSON200.ResourceId) })
	return site
}

func DeleteSite(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	siteID string,
) {
	tb.Helper()

	resDelSite, err := apiClient.SiteServiceDeleteSiteWithResponse(ctx, siteID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, resDelSite.StatusCode())
}

// CreateHost adds a host via the REST APIs, and setup the soft delete upon test cleanup.
func CreateHost(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	hostRequest api.HostResource,
) *api.HostServiceCreateHostResponse {
	tb.Helper()

	host, err := apiClient.HostServiceCreateHostWithResponse(ctx, hostRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, host.StatusCode())
	require.NotNil(tb, host.JSON200, "Host creation returned nil JSON200")

	tb.Cleanup(func() { SoftDeleteHost(context.Background(), tb, apiClient, host.JSON200) })
	return host
}

// SoftDeleteHost
// \unallocate the host if allocated to any site so we free any linked resources (site), and does a soft delete of Host.
// Eventually Host Resource Manager will do the hard delete.
func SoftDeleteHost(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	host *api.HostResource,
) {
	tb.Helper()

	UnallocateHostFromSite(ctx, tb, apiClient, host)
	resDelHost, err := apiClient.HostServiceDeleteHostWithResponse(
		ctx,
		*host.ResourceId,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, resDelHost.StatusCode())
}

// UnallocateHostFromSite: unallocate the given hostId from a site.
func UnallocateHostFromSite(ctx context.Context, tb testing.TB, apiClient *api.ClientWithResponses, hostReq *api.HostResource) {
	tb.Helper()

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
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, res.StatusCode())
}

func AssertInMaintenance(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	hostID *string,
	siteID *string,
	regionID *string,
	timestamp time.Time,
	expectedSchedules int,
	found bool,
) {
	tb.Helper()

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
	require.NoError(tb, err)
	if found {
		assert.Equal(tb, http.StatusOK, sReply.StatusCode())
		length := 0
		if sReply.JSON200.SingleSchedules != nil {
			length += len(sReply.JSON200.SingleSchedules)
		}
		if sReply.JSON200.RepeatedSchedules != nil {
			length += len(sReply.JSON200.RepeatedSchedules)
		}
		assert.Equal(tb, expectedSchedules, length, "Wrong number of schedules")
	} else {
		assert.Equal(tb, http.StatusOK, sReply.StatusCode())
	}
}

func CreateOS(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	reqOS api.OperatingSystemResource,
) *api.OperatingSystemServiceCreateOperatingSystemResponse {
	tb.Helper()

	osCreated, err := apiClient.OperatingSystemServiceCreateOperatingSystemWithResponse(
		ctx,
		reqOS,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, osCreated.StatusCode())
	require.NotNil(tb, osCreated.JSON200, "OS creation returned nil JSON200")
	require.NotNil(tb, osCreated.JSON200.ResourceId, "OS creation returned nil ResourceId")

	tb.Cleanup(func() {
		time.Sleep(sleepTime) // Waits until Instance reconciliation happens
		DeleteOS(context.Background(), tb, apiClient, *osCreated.JSON200.ResourceId)
	})

	return osCreated
}

func DeleteOS(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	osID string,
) {
	tb.Helper()

	osDel, err := apiClient.OperatingSystemServiceDeleteOperatingSystemWithResponse(
		ctx,
		osID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	if osDel.StatusCode() != http.StatusOK {
		tb.Logf("WARNING: Failed to delete OS %s - Status Code: %d", osID, osDel.StatusCode())
		tb.Logf("Response Body: %s", string(osDel.Body))
	}
	assert.Equal(tb, http.StatusOK, osDel.StatusCode())
}

func CreateWorkload(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	reqWorkload api.WorkloadResource,
) *api.WorkloadServiceCreateWorkloadResponse {
	tb.Helper()

	wCreated, err := apiClient.WorkloadServiceCreateWorkloadWithResponse(
		ctx,
		reqWorkload,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, wCreated.StatusCode())

	if wCreated.JSON200 != nil && wCreated.JSON200.ResourceId != nil {
		tb.Cleanup(func() { DeleteWorkload(context.Background(), tb, apiClient, *wCreated.JSON200.ResourceId) })
	}
	return wCreated
}

func DeleteWorkload(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	workloadID string,
) {
	tb.Helper()

	wDel, err := apiClient.WorkloadServiceDeleteWorkloadWithResponse(
		ctx,
		workloadID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, wDel.StatusCode())
}

func CreateWorkloadMember(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	reqMember api.WorkloadMember,
) *api.WorkloadMemberServiceCreateWorkloadMemberResponse {
	tb.Helper()

	mCreated, err := apiClient.WorkloadMemberServiceCreateWorkloadMemberWithResponse(
		ctx,
		reqMember,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, mCreated.StatusCode())

	if mCreated.JSON200 != nil && mCreated.JSON200.ResourceId != nil {
		tb.Cleanup(func() { DeleteWorkloadMember(context.Background(), tb, apiClient, *mCreated.JSON200.ResourceId) })
	}
	return mCreated
}

func DeleteWorkloadMember(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	memberID string,
) {
	tb.Helper()

	mDel, err := apiClient.WorkloadMemberServiceDeleteWorkloadMemberWithResponse(
		ctx,
		memberID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, mDel.StatusCode())
}

func CreateInstance(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	instRequest api.InstanceResource,
) *api.InstanceServiceCreateInstanceResponse {
	tb.Helper()

	createdInstance, err := apiClient.InstanceServiceCreateInstanceWithResponse(
		ctx, instRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, createdInstance.StatusCode())
	assert.NotNil(tb, createdInstance.JSON200)
	assert.NotNil(tb, createdInstance.JSON200.ResourceId)

	tb.Cleanup(func() { DeleteInstance(context.Background(), tb, apiClient, *createdInstance.JSON200.ResourceId) })
	return createdInstance
}

func DeleteInstance(ctx context.Context, tb testing.TB, apiClient *api.ClientWithResponses, instanceID string) {
	tb.Helper()

	resDelInst, err := apiClient.InstanceServiceDeleteInstanceWithResponse(
		ctx, instanceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, resDelInst.StatusCode())
}

func CreateTelemetryLogsGroup(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	request api.TelemetryLogsGroupResource,
) *api.TelemetryLogsGroupServiceCreateTelemetryLogsGroupResponse {
	tb.Helper()

	created, err := apiClient.TelemetryLogsGroupServiceCreateTelemetryLogsGroupWithResponse(
		ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, created.StatusCode())
	require.NotNil(tb, created.JSON200, "TelemetryLogsGroup creation returned nil JSON200")
	require.NotNil(tb, created.JSON200.ResourceId, "TelemetryLogsGroup creation returned nil ResourceId")

	tb.Cleanup(func() {
		DeleteTelemetryLogsGroup(context.Background(), tb, apiClient, *created.JSON200.ResourceId)
	})
	return created
}

func DeleteTelemetryLogsGroup(
	ctx context.Context, tb testing.TB, apiClient *api.ClientWithResponses, id string,
) {
	tb.Helper()

	res, err := apiClient.TelemetryLogsGroupServiceDeleteTelemetryLogsGroupWithResponse(
		ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, res.StatusCode())
}

func CreateTelemetryMetricsGroup(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	request api.TelemetryMetricsGroupResource,
) *api.TelemetryMetricsGroupServiceCreateTelemetryMetricsGroupResponse {
	tb.Helper()

	created, err := apiClient.TelemetryMetricsGroupServiceCreateTelemetryMetricsGroupWithResponse(
		ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, created.StatusCode())
	require.NotNil(tb, created.JSON200, "TelemetryMetricsGroup creation returned nil JSON200")
	require.NotNil(tb, created.JSON200.ResourceId, "TelemetryMetricsGroup creation returned nil ResourceId")

	tb.Cleanup(func() {
		DeleteTelemetryMetricsGroup(context.Background(), tb, apiClient, *created.JSON200.ResourceId)
	})
	return created
}

func DeleteTelemetryMetricsGroup(
	ctx context.Context, tb testing.TB, apiClient *api.ClientWithResponses, id string,
) {
	tb.Helper()

	res, err := apiClient.TelemetryMetricsGroupServiceDeleteTelemetryMetricsGroupWithResponse(
		ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, res.StatusCode())
}

func CreateTelemetryLogsProfile(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	request api.TelemetryLogsProfileResource,
) *api.TelemetryLogsProfileServiceCreateTelemetryLogsProfileResponse {
	tb.Helper()

	created, err := apiClient.TelemetryLogsProfileServiceCreateTelemetryLogsProfileWithResponse(
		ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, created.StatusCode())
	require.NotNil(tb, created.JSON200, "TelemetryLogsProfile creation returned nil JSON200")
	require.NotNil(tb, created.JSON200.ResourceId, "TelemetryLogsProfile creation returned nil ResourceId")

	tb.Cleanup(func() {
		DeleteTelemetryLogsProfile(context.Background(), tb, apiClient, *created.JSON200.ResourceId)
	})
	return created
}

func DeleteTelemetryLogsProfile(
	ctx context.Context, tb testing.TB, apiClient *api.ClientWithResponses, id string,
) {
	tb.Helper()

	res, err := apiClient.TelemetryLogsProfileServiceDeleteTelemetryLogsProfileWithResponse(
		ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, res.StatusCode())
}

func CreateTelemetryMetricsProfile(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	request api.TelemetryMetricsProfileResource,
) *api.TelemetryMetricsProfileServiceCreateTelemetryMetricsProfileResponse {
	tb.Helper()

	created, err := apiClient.TelemetryMetricsProfileServiceCreateTelemetryMetricsProfileWithResponse(
		ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, created.StatusCode())
	require.NotNil(tb, created.JSON200, "TelemetryMetricsProfile creation returned nil JSON200")
	require.NotNil(tb, created.JSON200.ResourceId, "TelemetryMetricsProfile creation returned nil ResourceId")

	tb.Cleanup(func() {
		DeleteTelemetryMetricsProfile(context.Background(), tb, apiClient, *created.JSON200.ResourceId)
	})
	return created
}

func DeleteTelemetryMetricsProfile(
	ctx context.Context, tb testing.TB, apiClient *api.ClientWithResponses, id string,
) {
	tb.Helper()

	res, err := apiClient.TelemetryMetricsProfileServiceDeleteTelemetryMetricsProfileWithResponse(
		ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, res.StatusCode())
}

func CreateProvider(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	reqProvider api.ProviderResource,
) *api.ProviderServiceCreateProviderResponse {
	tb.Helper()

	providerCreated, err := apiClient.ProviderServiceCreateProviderWithResponse(
		ctx,
		reqProvider,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, providerCreated.StatusCode())
	require.NotNil(tb, providerCreated.JSON200, "Provider creation returned nil JSON200")
	require.NotNil(tb, providerCreated.JSON200.ResourceId, "Provider creation returned nil ResourceId")

	tb.Cleanup(func() { DeleteProvider(context.Background(), tb, apiClient, *providerCreated.JSON200.ResourceId) })
	return providerCreated
}

func DeleteProvider(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	providerID string,
) {
	tb.Helper()

	providerDel, err := apiClient.ProviderServiceDeleteProviderWithResponse(
		ctx,
		providerID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, providerDel.StatusCode())
}

// CreateLocalAccount creates a LocalAccount resource and returns the response.
func CreateLocalAccount(ctx context.Context, t *testing.T, apiClient *api.ClientWithResponses,
	request api.LocalAccountResource,
) *api.LocalAccountServiceCreateLocalAccountResponse {
	t.Helper()

	response, err := apiClient.LocalAccountServiceCreateLocalAccountWithResponse(
		ctx,
		request,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode())
	require.NotNil(t, response.JSON200, "LocalAccount creation returned nil JSON200")
	require.NotNil(t, response.JSON200.ResourceId, "LocalAccount creation returned nil ResourceId")

	t.Cleanup(func() {
		DeleteOS(context.Background(), t, apiClient, *response.JSON200.ResourceId)
	})
	return response
}

// DeleteLocalAccount deletes a LocalAccount resource by its ID.
func DeleteLocalAccount(ctx context.Context, t *testing.T,
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
	//nolint:gosec // Ok to use non-cryptographic random number generator for test purposes
	randName := fmt.Sprintf("Test Host %d", rand.Uint32())
	return api.HostResource{
		Name: randName,
		Uuid: &uuidHost,
	}
}

func CreateOsUpdatePolicy(
	ctx context.Context,
	tb testing.TB,
	apiClient *api.ClientWithResponses,
	reqPolicy api.OSUpdatePolicy,
) *api.OSUpdatePolicyCreateOSUpdatePolicyResponse {
	tb.Helper()

	policyCreated, err := apiClient.OSUpdatePolicyCreateOSUpdatePolicyWithResponse(
		ctx,
		reqPolicy,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	assert.Equal(tb, http.StatusOK, policyCreated.StatusCode())
	require.NotNil(tb, policyCreated.JSON200, "OSUpdatePolicy creation returned nil JSON200")
	require.NotNil(tb, policyCreated.JSON200.ResourceId, "OSUpdatePolicy creation returned nil ResourceId")

	tb.Cleanup(func() {
		DeleteOSUpdatePolicy(context.Background(), tb, apiClient, *policyCreated.JSON200.ResourceId)
	})

	return policyCreated
}

func DeleteOSUpdatePolicy(ctx context.Context, tb testing.TB, apiClient *api.ClientWithResponses, policyID string) {
	tb.Helper()

	policyDel, err := apiClient.OSUpdatePolicyDeleteOSUpdatePolicyWithResponse(
		ctx,
		policyID,
		AddJWTtoTheHeader, AddProjectIDtoTheHeader,
	)
	require.NoError(tb, err)
	if policyDel.StatusCode() != http.StatusOK {
		tb.Logf("WARNING: Failed to delete OSUpdatePolicy %s - Status Code: %d", policyID, policyDel.StatusCode())
		tb.Logf("Response Body: %s", string(policyDel.Body))
	}
	assert.Equal(tb, http.StatusOK, policyDel.StatusCode())
}
