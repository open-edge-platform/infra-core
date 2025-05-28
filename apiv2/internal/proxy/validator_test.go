// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package proxy_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	server "github.com/open-edge-platform/infra-core/apiv2/v2/internal/proxy"
	api "github.com/open-edge-platform/infra-core/apiv2/v2/pkg/api/v2"
	"github.com/open-edge-platform/infra-core/apiv2/v2/test/utils"
)

// To create a request with an authorization header.
func createRequestWithMethodPathParams(method, path string) *http.Request {
	req := httptest.NewRequest(method, path, http.NoBody)
	return req
}

//nolint:funlen // it is a test
func TestOapiValidatorInterceptor(t *testing.T) {
	// Create an Echo instance for testing.
	e := echo.New()

	siteName := "site"
	siteBody := api.SiteResource{Name: &siteName}
	sitePostRequestValid, err := api.NewSiteServiceCreateSiteRequest("", siteBody)
	assert.NoError(t, err)

	regionID := "region-12345678"
	params := &api.SiteServiceListSitesParams{Filter: &regionID}
	siteGetRequestValid, err := api.NewSiteServiceListSitesRequest("", params)
	assert.NoError(t, err)

	pageSizeWrong := 2000
	paramsWrong := &api.SiteServiceListSitesParams{PageSize: &pageSizeWrong}
	siteGetRequestInvalid, err := api.NewSiteServiceListSitesRequest("", paramsWrong)
	assert.NoError(t, err)

	emptyUUID := ""
	hostParamsWrong := &api.HostServiceListHostsParams{Filter: &emptyUUID}
	hostsGetRequestInvalid, err := api.NewHostServiceListHostsRequest("", hostParamsWrong)
	assert.NoError(t, err)

	hostsPostRequestValid, err := api.NewHostServiceCreateHostRequest("", utils.Host1Request)
	assert.NoError(t, err)

	hostsPostRegisterRequestValid, err := api.NewHostServiceRegisterHostRequest(
		"", utils.HostRegisterAutoOnboard)
	assert.NoError(t, err)

	hostName := "host"
	hostInvalidBodyRequest := api.HostResource{
		Name:        hostName,
		Uuid:        &utils.Host1UUID1,
		CpuTopology: &regionID,
	}
	hostsPostRequestInvalid, err := api.NewHostServiceCreateHostRequest(
		"", hostInvalidBodyRequest)
	assert.NoError(t, err)

	// Enforce the test of required fields in the request body.
	workloadInvalidBodyRequest := api.WorkloadResource{
		// Kind: api.WORKLOADKINDCLUSTER,
	}
	workloadPostRequestInvalidNoName, err := api.NewWorkloadServiceCreateWorkloadRequest(
		"", workloadInvalidBodyRequest)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		request        *http.Request
		expectedStatus int
	}{
		{
			name:           "Wrong requested path",
			request:        httptest.NewRequest(http.MethodGet, "/", http.NoBody),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Wrong requested path",
			request:        createRequestWithMethodPathParams(http.MethodPost, "/home"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid Body Format",
			request:        workloadPostRequestInvalidNoName,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Valid Request/Body Format",
			request: createRequestWithMethodPathParams(http.MethodGet,
				"/edge-infra.orchestrator.apis/v2/hosts"),
			expectedStatus: http.StatusOK,
		},
		{
			name: "Valid Request Path",
			request: createRequestWithMethodPathParams(http.MethodGet,
				"/edge-infra.orchestrator.apis/v2/hosts/summary"),
			expectedStatus: http.StatusOK,
		},
		{
			name: "Valid path prefix with invalid request path",
			request: createRequestWithMethodPathParams(http.MethodGet,
				"/edge-infra.orchestrator.apis/v2/hostss"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Valid path prefix/request",
			request:        sitePostRequestValid,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid path prefix/request",
			request:        siteGetRequestValid,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid params in site GET request - wrong pagesize",
			request:        siteGetRequestInvalid,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid params in host GET request - miss UUID",
			request:        hostsGetRequestInvalid,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Valid path prefix/request - host POST",
			request:        hostsPostRequestValid,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid path prefix/request - host POST",
			request:        hostsPostRequestValid,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid path prefix/request - host register POST",
			request:        hostsPostRegisterRequestValid,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid body request host POST - ready only property",
			request:        hostsPostRequestInvalid,
			expectedStatus: http.StatusBadRequest,
		},
	}

	openAPIDefinition, err := api.GetSwagger()
	require.NoError(t, err, "Failed to get OpenAPI definition")

	for _, s := range openAPIDefinition.Servers {
		s.URL = strings.ReplaceAll(s.URL, "{apiRoot}", "")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a context with the request and recorder.
			c := e.NewContext(tt.request, httptest.NewRecorder())
			// Create a dummy next handler that returns OK status.
			next := func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			}

			// Invoke interceptor.
			validator := server.OapiRequestValidator(openAPIDefinition)
			handler := validator(next)
			err := handler(c)
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
			} else {
				var httpErr *echo.HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.expectedStatus, httpErr.Code, "Expected an HTTP error")
				} else {
					t.Errorf("Expected an echo.HTTPError, got %T", err)
				}
			}
		})
	}
}
