// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	api "github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	apitestclient "github.com/open-edge-platform/infra-core/apiv2/v2/test/client"
	apitestutils "github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
	"github.com/open-edge-platform/infra-core/exporters-inventory/test/utils"
)

const testTimeout = time.Duration(120) * time.Second

var (
	apiURL                = flag.String("apiURL", "http://localhost:8080/edge-infra.orchestrator.apis/v1", "The URL of the edge infrastructure manager REST API")
	promURL               = flag.String("promURL", "http://localhost:9101/metrics", "The URL of the Exporter Prometheus REST API")
	metricNameStatus      = "edge_host_status"
	metricNameMaintenance = "edge_host_schedule"
)

// TestExporter_HTTP uses an inventory client to add resources, Host and Schedules,
// validates the maintenance status of the hosts, and retrieves the metrics from
// the exporter HTTP web interface, and validates them against the inventory information.
func TestExporter_HTTP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	apiClient, err := api.NewClientWithResponses(*apiURL)
	require.NoError(t, err)

	region1 := apitestclient.CreateRegion(ctx, t, apiClient, apitestutils.Region1Request)
	apitestutils.Site1Request.RegionId = region1.JSON200.ResourceId
	site2 := apitestclient.CreateSite(ctx, t, apiClient, apitestutils.Site1Request)
	apitestutils.Site1Request.RegionId = nil

	site1 := apitestclient.CreateSite(ctx, t, apiClient, apitestutils.Site1Request)

	hostReq1 := apitestclient.GetHostRequestWithRandomUUID()
	hostReq1.SiteId = site1.JSON200.ResourceId
	host1 := apitestclient.CreateHost(ctx, t, apiClient, hostReq1)

	host2 := apitestclient.CreateHost(ctx, t, apiClient, apitestclient.GetHostRequestWithRandomUUID())
	host3 := apitestclient.CreateHost(ctx, t, apiClient, apitestclient.GetHostRequestWithRandomUUID())

	hostReq4 := apitestclient.GetHostRequestWithRandomUUID()
	hostReq4.SiteId = site2.JSON200.ResourceId
	host4 := apitestclient.CreateHost(ctx, t, apiClient, hostReq4)

	apitestutils.SingleScheduleAlwaysRequest.TargetRegionId = region1.JSON200.ResourceId
	apitestclient.CreateSchedSingle(ctx, t, apiClient, apitestutils.SingleScheduleAlwaysRequest)
	apitestutils.SingleScheduleAlwaysRequest.TargetRegionId = nil

	apitestutils.SingleScheduleAlwaysRequest.TargetSiteId = site1.JSON200.ResourceId
	apitestclient.CreateSchedSingle(ctx, t, apiClient, apitestutils.SingleScheduleAlwaysRequest)
	apitestutils.SingleScheduleAlwaysRequest.TargetSiteId = nil

	apitestutils.SingleScheduleAlwaysRequest.TargetHostId = host2.JSON200.ResourceId
	apitestclient.CreateSchedSingle(ctx, t, apiClient, apitestutils.SingleScheduleAlwaysRequest)
	apitestutils.SingleScheduleAlwaysRequest.TargetHostId = nil

	apitestutils.SingleScheduleNever.TargetHostId = host3.JSON200.ResourceId
	apitestclient.CreateSchedSingle(ctx, t, apiClient, apitestutils.SingleScheduleNever)
	apitestutils.SingleScheduleNever.TargetHostId = nil

	timestamp := time.Now()

	// Host1 should be in maintenance (it's in Site1, and we have maintenance window for Site1)
	apitestclient.AssertInMaintenance(ctx, t, apiClient, host1.JSON200.ResourceId, nil, nil, timestamp, 1, true)
	apitestclient.AssertInMaintenance(ctx, t, apiClient, nil, site1.JSON200.ResourceId, nil, timestamp, 1, true)

	// Host2 should be in maintenance (it's directly in maintenance)
	apitestclient.AssertInMaintenance(ctx, t, apiClient, host2.JSON200.ResourceId, nil, nil, timestamp, 1, true)

	// Host3 should not be in maintenance
	apitestclient.AssertInMaintenance(ctx, t, apiClient, host3.JSON200.ResourceId, nil, nil, timestamp, 0, false)

	// Host4 should be in maintenance because of maintenance window of Region1
	apitestclient.AssertInMaintenance(ctx, t, apiClient, host4.JSON200.ResourceId, nil, nil, timestamp, 1, true)
	apitestclient.AssertInMaintenance(ctx, t, apiClient, nil, nil, region1.JSON200.ResourceId, timestamp, 1, true)

	time.Sleep(20 * time.Second)
	metricsText, err := GetMetricsHTTP(*promURL)
	assert.NoError(t, err)

	metricsFormated, err := utils.ParsePrometheusTextMetrics(metricsText)
	assert.NoError(t, err)

	// Validate host 1 - has status and in maintenance (due to site1 maintenance).
	host1Labels := map[string]string{
		"hostID": *host1.JSON200.ResourceId,
	}
	host1ValueStatus := 1
	host1ValueMaintenance := 1

	ackHost1Status := utils.ValidateMetrics(metricsFormated, metricNameStatus, host1Labels, float64(host1ValueStatus))
	assert.True(t, ackHost1Status)
	ackHost1Maintenance := utils.ValidateMetrics(metricsFormated, metricNameMaintenance, host1Labels, float64(host1ValueMaintenance))
	assert.True(t, ackHost1Maintenance)

	// Validate host 2 - has status and in maintenance (directly scheduled).
	host2Labels := map[string]string{
		"hostID": *host2.JSON200.ResourceId,
	}
	host2ValueStatus := 1
	host2ValueMaintenance := 1

	ackHost2Status := utils.ValidateMetrics(metricsFormated, metricNameStatus, host2Labels, float64(host2ValueStatus))
	assert.True(t, ackHost2Status)
	ackHost2Maintenance := utils.ValidateMetrics(metricsFormated, metricNameMaintenance, host2Labels, float64(host2ValueMaintenance))
	assert.True(t, ackHost2Maintenance)

	// Validate host 3 - has status and not in maintenance (no schedules).
	host3Labels := map[string]string{
		"hostID": *host3.JSON200.ResourceId,
	}
	host3ValueStatus := 1
	host3ValueMaintenance := 0

	ackHost3Status := utils.ValidateMetrics(metricsFormated, metricNameStatus, host3Labels, float64(host3ValueStatus))
	assert.True(t, ackHost3Status)
	ackHost3Maintenance := utils.ValidateMetrics(metricsFormated, metricNameMaintenance, host3Labels, float64(host3ValueMaintenance))
	assert.True(t, ackHost3Maintenance)

	// Validate host 4 - has status and in maintenance (due to region1 maintenance).
	host4Labels := map[string]string{
		"hostID": *host4.JSON200.ResourceId,
	}
	host4ValueStatus := 1
	host4ValueMaintenance := 1

	ackHost4Status := utils.ValidateMetrics(metricsFormated, metricNameStatus, host4Labels, float64(host4ValueStatus))
	assert.True(t, ackHost4Status)
	ackHost4Maintenance := utils.ValidateMetrics(metricsFormated, metricNameMaintenance, host4Labels, float64(host4ValueMaintenance))
	assert.True(t, ackHost4Maintenance)
}
