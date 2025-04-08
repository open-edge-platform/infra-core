// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/api/internal/server"
	api "github.com/open-edge-platform/infra-core/api/pkg/api/v0"
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
	body := api.Site{Name: &siteName}
	sitePostRequestValid, err := api.NewPostSitesRequest("/edge-infra.orchestrator.apis/v1/", body)
	assert.NoError(t, err)

	regionID := "region-12345678"
	params := &api.GetSitesParams{RegionID: &regionID}
	siteGetRequestValid, err := api.NewGetSitesRequest("/edge-infra.orchestrator.apis/v1/", params)
	assert.NoError(t, err)

	regionIDWrong := "region-1234567"
	paramsWrong := &api.GetSitesParams{RegionID: &regionIDWrong}
	siteGetRequestInvalid, err := api.NewGetSitesRequest("/edge-infra.orchestrator.apis/v1/", paramsWrong)
	assert.NoError(t, err)

	emptyUUID := ""
	hostParamsWrong := &api.GetComputeHostsParams{Uuid: &emptyUUID}
	hostsGetRequestInvalid, err := api.NewGetComputeHostsRequest("/edge-infra.orchestrator.apis/v1/", hostParamsWrong)
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
			name: "Invalid Body Format",
			request: createRequestWithMethodPathParams(http.MethodPost,
				"/edge-infra.orchestrator.apis/v1/compute/hosts"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Valid Request/Body Format",
			request: createRequestWithMethodPathParams(http.MethodGet,
				"/edge-infra.orchestrator.apis/v1/compute/hosts"),
			expectedStatus: http.StatusOK,
		},
		{
			name: "Valid Request Path",
			request: createRequestWithMethodPathParams(http.MethodGet,
				"/edge-infra.orchestrator.apis/v1/compute/hosts/summary"),
			expectedStatus: http.StatusOK,
		},
		{
			name: "Valid path prefix with invalid request path",
			request: createRequestWithMethodPathParams(http.MethodGet,
				"/edge-infra.orchestrator.apis/v1/compute/hosts/summaries"),
			expectedStatus: http.StatusBadRequest,
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
			name:           "Invalid params in site GET request - wrong region ID",
			request:        siteGetRequestInvalid,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid params in host GET request - miss UUID",
			request:        hostsGetRequestInvalid,
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
