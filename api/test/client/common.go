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

	"github.com/open-edge-platform/infra-core/api/pkg/api/v0"
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
	authKey                  = "Authorization"
	projectID                = "PROJECT_ID"
	projectIDKey             = "ActiveProjectID"
	userAgent                = "User-Agent"
	ecmServiceName           = "ecm-api"
	observabilityServiceName = "common-metric-query-metrics"
)

var (
	apiUrl = flag.String("apiurl", "http://localhost:8080/edge-infra.orchestrator.apis/v1", "The URL of the edge infrastructure manager REST API")
	caPath = flag.String("caPath", "", "The path to the CA certificate file of the target cluster")
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

func ListMetadataContains(lst api.Metadata, key, value string) bool {
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
		return fmt.Errorf("can't find a \"%s\" variable, please set it in your environment", jwtToken)
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

func hostsContainsId(hosts []api.Host, hostID string) bool {
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
	reqSched api.SingleSchedule,
) *api.PostSchedulesSingleResponse {
	t.Helper()

	sched, err := apiClient.PostSchedulesSingleWithResponse(
		ctx,
		reqSched,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, sched.StatusCode())

	t.Cleanup(func() { DeleteSchedSingle(t, context.Background(), apiClient, *sched.JSON201.SingleScheduleID) })
	return sched
}

func DeleteSchedSingle(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	schedID string,
) {
	t.Helper()

	schedDel, err := apiClient.DeleteSchedulesSingleSingleScheduleIDWithResponse(
		ctx,
		schedID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, schedDel.StatusCode())
}

func CreateSchedRepeated(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	reqSched api.RepeatedSchedule,
) *api.PostSchedulesRepeatedResponse {
	t.Helper()

	sched, err := apiClient.PostSchedulesRepeatedWithResponse(
		ctx,
		reqSched,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, sched.StatusCode())

	t.Cleanup(func() { DeleteSchedRepeated(t, context.Background(), apiClient, *sched.JSON201.RepeatedScheduleID) })
	return sched
}

func DeleteSchedRepeated(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	schedID string,
) {
	t.Helper()

	schedDel, err := apiClient.DeleteSchedulesRepeatedRepeatedScheduleIDWithResponse(
		ctx,
		schedID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, schedDel.StatusCode())
}

func CreateRegion(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	regionRequest api.Region,
) *api.PostRegionsResponse {
	t.Helper()

	region, err := apiClient.PostRegionsWithResponse(ctx, regionRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, region.StatusCode())

	t.Cleanup(func() { DeleteRegion(t, context.Background(), apiClient, *region.JSON201.RegionID) })
	return region
}

func DeleteRegion(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	regionID string,
) {
	t.Helper()

	resDelRegion, err := apiClient.DeleteRegionsRegionIDWithResponse(
		ctx,
		regionID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resDelRegion.StatusCode())
}

func CreateSite(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	siteRequest api.Site,
) *api.PostSitesResponse {
	t.Helper()

	site, err := apiClient.PostSitesWithResponse(ctx, siteRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, site.StatusCode())

	t.Cleanup(func() { DeleteSite(t, context.Background(), apiClient, *site.JSON201.SiteID) })
	return site
}

func DeleteSite(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	siteID string,
) {
	t.Helper()

	resDelSite, err := apiClient.DeleteSitesSiteIDWithResponse(ctx, siteID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resDelSite.StatusCode())
}

func CreateOu(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	ouRequest api.OU,
) *api.PostOusResponse {
	t.Helper()

	createdOu, err := apiClient.PostOusWithResponse(ctx, ouRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createdOu.StatusCode())

	t.Cleanup(func() { DeleteOu(t, context.Background(), apiClient, *createdOu.JSON201.OuID) })
	return createdOu
}

func DeleteOu(t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, ouID string) {
	t.Helper()

	resDelOu, err := apiClient.DeleteOusOuIDWithResponse(ctx, ouID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resDelOu.StatusCode())
}

// CreateHost adds a host via the REST APIs, and setup the soft delete upon test cleanup.
func CreateHost(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	hostRequest api.Host,
) *api.PostComputeHostsResponse {
	t.Helper()

	host, err := apiClient.PostComputeHostsWithResponse(ctx, hostRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, host.StatusCode())

	hostID := *host.JSON201.ResourceId
	t.Cleanup(func() { SoftDeleteHost(t, context.Background(), apiClient, hostID) })
	return host
}

// SoftDeleteHost: unallocate the host if allocated to any site so we free any linked resources (site), and does a soft delete of Host.
// Eventually Host Resource Manager will do the hard delete.
func SoftDeleteHost(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	hostID string,
) {
	t.Helper()

	UnallocateHostFromSite(t, ctx, apiClient, hostID)
	resDelHost, err := apiClient.DeleteComputeHostsHostIDWithResponse(
		ctx,
		hostID,
		api.DeleteComputeHostsHostIDJSONRequestBody{},
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resDelHost.StatusCode())
}

// UnallocateHostFromSite: unallocate the given hostId from a site.
func UnallocateHostFromSite(t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, hostId string) {
	t.Helper()

	hostRequestPatch := api.Host{SiteId: &emptyString}
	res, err := apiClient.PatchComputeHostsHostIDWithResponse(
		ctx,
		hostId,
		hostRequestPatch,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
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
	sReply, err := apiClient.GetSchedulesWithResponse(
		ctx,
		&api.GetSchedulesParams{
			HostID:    hostID,
			SiteID:    siteID,
			RegionID:  regionID,
			UnixEpoch: &timestampString,
		},
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	if found {
		assert.Equal(t, http.StatusOK, sReply.StatusCode())
		length := 0
		if sReply.JSON200.SingleSchedules != nil {
			length += len(*sReply.JSON200.SingleSchedules)
		}
		if sReply.JSON200.RepeatedSchedules != nil {
			length += len(*sReply.JSON200.RepeatedSchedules)
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
) *api.PostOSResourcesResponse {
	t.Helper()

	osCreated, err := apiClient.PostOSResourcesWithResponse(
		ctx,
		reqOS,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, osCreated.StatusCode())

	t.Cleanup(func() {
		time.Sleep(2 * time.Second) // Waits until Instance reconciliation happens
		DeleteOS(t, context.Background(), apiClient, *osCreated.JSON201.OsResourceID)
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

	osDel, err := apiClient.DeleteOSResourcesOSResourceIDWithResponse(
		ctx,
		osID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, osDel.StatusCode())
}

func CreateWorkload(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	reqWorkload api.Workload,
) *api.PostWorkloadsResponse {
	t.Helper()

	wCreated, err := apiClient.PostWorkloadsWithResponse(
		ctx,
		reqWorkload,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, wCreated.StatusCode())

	t.Cleanup(func() { DeleteWorkload(t, context.Background(), apiClient, *wCreated.JSON201.WorkloadId) })
	return wCreated
}

func DeleteWorkload(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	workloadID string,
) {
	t.Helper()

	wDel, err := apiClient.DeleteWorkloadsWorkloadIDWithResponse(
		ctx,
		workloadID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, wDel.StatusCode())
}

func CreateWorkloadMember(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	reqMember api.WorkloadMember,
) *api.PostWorkloadMembersResponse {
	t.Helper()

	mCreated, err := apiClient.PostWorkloadMembersWithResponse(
		ctx,
		reqMember,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, mCreated.StatusCode())

	t.Cleanup(func() { DeleteWorkloadMember(t, context.Background(), apiClient, *mCreated.JSON201.WorkloadMemberId) })
	return mCreated
}

func DeleteWorkloadMember(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	memberID string,
) {
	t.Helper()

	mDel, err := apiClient.DeleteWorkloadMembersWorkloadMemberIDWithResponse(
		ctx,
		memberID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, mDel.StatusCode())
}

func CreateInstance(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	instRequest api.Instance,
) *api.PostInstancesResponse {
	t.Helper()

	createdInstance, err := apiClient.PostInstancesWithResponse(ctx, instRequest, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createdInstance.StatusCode())

	t.Cleanup(func() { DeleteInstance(t, context.Background(), apiClient, *createdInstance.JSON201.InstanceID) })
	return createdInstance
}

func DeleteInstance(t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, instanceID string) {
	t.Helper()

	resDelInst, err := apiClient.DeleteInstancesInstanceIDWithResponse(ctx, instanceID, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resDelInst.StatusCode())
}

func CreateTelemetryLogsGroup(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	request api.TelemetryLogsGroup,
) *api.PostTelemetryGroupsLogsResponse {
	t.Helper()

	created, err := apiClient.PostTelemetryGroupsLogsWithResponse(ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, created.StatusCode())

	t.Cleanup(func() {
		DeleteTelemetryLogsGroup(t, context.Background(), apiClient, *created.JSON201.TelemetryLogsGroupId)
	})
	return created
}

func DeleteTelemetryLogsGroup(
	t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, id string,
) {
	t.Helper()

	res, err := apiClient.DeleteTelemetryGroupsLogsTelemetryLogsGroupIdWithResponse(ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode())
}

func CreateTelemetryMetricsGroup(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	request api.TelemetryMetricsGroup,
) *api.PostTelemetryGroupsMetricsResponse {
	t.Helper()

	created, err := apiClient.PostTelemetryGroupsMetricsWithResponse(ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, created.StatusCode())

	t.Cleanup(func() {
		DeleteTelemetryMetricsGroup(t, context.Background(), apiClient, *created.JSON201.TelemetryMetricsGroupId)
	})
	return created
}

func DeleteTelemetryMetricsGroup(
	t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, id string,
) {
	t.Helper()

	res, err := apiClient.DeleteTelemetryGroupsMetricsTelemetryMetricsGroupIdWithResponse(ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode())
}

func CreateTelemetryLogsProfile(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	request api.TelemetryLogsProfile,
) *api.PostTelemetryProfilesLogsResponse {
	t.Helper()

	created, err := apiClient.PostTelemetryProfilesLogsWithResponse(ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, created.StatusCode())

	t.Cleanup(func() {
		DeleteTelemetryLogsProfile(t, context.Background(), apiClient, *created.JSON201.ProfileId)
	})
	return created
}

func DeleteTelemetryLogsProfile(
	t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, id string,
) {
	t.Helper()

	res, err := apiClient.DeleteTelemetryProfilesLogsTelemetryLogsProfileIdWithResponse(ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode())
}

func CreateTelemetryMetricsProfile(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	request api.TelemetryMetricsProfile,
) *api.PostTelemetryProfilesMetricsResponse {
	t.Helper()

	created, err := apiClient.PostTelemetryProfilesMetricsWithResponse(ctx, request, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, created.StatusCode())

	t.Cleanup(func() {
		DeleteTelemetryMetricsProfile(t, context.Background(), apiClient, *created.JSON201.ProfileId)
	})
	return created
}

func DeleteTelemetryMetricsProfile(
	t testing.TB, ctx context.Context, apiClient *api.ClientWithResponses, id string,
) {
	t.Helper()

	res, err := apiClient.DeleteTelemetryProfilesMetricsTelemetryMetricsProfileIdWithResponse(ctx, id, AddJWTtoTheHeader, AddProjectIDtoTheHeader)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode())
}

func CreateProvider(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	reqProvider api.Provider,
) *api.PostProvidersResponse {
	t.Helper()

	providerCreated, err := apiClient.PostProvidersWithResponse(
		ctx,
		reqProvider,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, providerCreated.StatusCode())

	t.Cleanup(func() { DeleteProvider(t, context.Background(), apiClient, *providerCreated.JSON201.ProviderID) })
	return providerCreated
}

func DeleteProvider(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	providerID string,
) {
	t.Helper()

	providerDel, err := apiClient.DeleteProvidersProviderIDWithResponse(
		ctx,
		providerID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, providerDel.StatusCode())
}

// This utility function mimics an addition of a required information to the HTTP Header.
func addEcmUserAgentToTheHeader(_ context.Context, req *http.Request) error {
	req.Header.Add(userAgent, ecmServiceName)
	return nil
}

func addRequestEditors(client *api.Client) error {
	client.RequestEditors = append(client.RequestEditors, authObsUserAgent)
	return nil
}

func authObsUserAgent(_ context.Context, req *http.Request) error {
	req.Header.Set(userAgent, observabilityServiceName)
	return nil
}

func GetHostRequestWithRandomUUID() api.Host {
	uuidHost := uuid.New()
	return api.Host{
		Name: fmt.Sprintf("Test Host %d", rand.Uint32()),
		Uuid: &uuidHost,
	}
}

func CreateLocalAccount(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	reqLocalAccount api.LocalAccount,
) *api.PostLocalAccountsResponse {
	t.Helper()

	localAccountCreated, err := apiClient.PostLocalAccountsWithResponse(
		ctx,
		reqLocalAccount,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, localAccountCreated.StatusCode())

	t.Cleanup(func() { DeleteLocalAccount(t, context.Background(), apiClient, *localAccountCreated.JSON201.LocalAccountID) })
	return localAccountCreated
}

func DeleteLocalAccount(
	t testing.TB,
	ctx context.Context,
	apiClient *api.ClientWithResponses,
	localAccountID string,
) {
	t.Helper()

	localAccountDel, err := apiClient.DeleteLocalAccountsLocalAccountIDWithResponse(
		ctx,
		localAccountID,
		AddJWTtoTheHeader,
		AddProjectIDtoTheHeader,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, localAccountDel.StatusCode())
}
