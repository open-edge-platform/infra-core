// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0
package proxy_test

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/proxy"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tenant"
)

const (
	// SharedSecretKey environment variable name for shared secret key for signing a token.
	SharedSecretKey = "SHARED_SECRET_KEY"
	secretKey       = "randomSecretKey"
	writeRole       = "im-rw"
	readRole        = "im-r"
)

var (
	allRoles = []string{
		writeRole,
		readRole,
	}
	tenantUUID = uuid.New().String()
)

// To create a request with an authorization header.
func createRequestWithAuthHeader(authScheme, authToken string, method string) *http.Request {
	req := httptest.NewRequest(method, "/", http.NoBody)
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", authScheme, authToken))
	return req
}

// To generates a valid JWT token for testing purposes.
func generateValidJWT(tb testing.TB, tenantID string, rolesSuffix []string) (jwtStr string, err error) {
	tb.Helper()
	roles := make([]string, len(rolesSuffix))
	for i, role := range rolesSuffix {
		roles[i] = tenantID + "_" + role
	}
	claims := &jwt.MapClaims{
		"iss": "https://keycloak.kind.internal/realms/master",
		"exp": time.Now().Add(time.Hour).Unix(),
		"typ": "Bearer",
		"realm_access": map[string]interface{}{
			"roles": roles,
		},
	}
	tb.Setenv(SharedSecretKey, secretKey)
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims)
	jwtStr, err = token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return jwtStr, nil
}

//nolint:funlen // it is a test
func TestAuthenticationAuthorizationInterceptor(t *testing.T) {
	// Set up the environment variable for allowed missing auth clients.
	t.Setenv(proxy.AllowMissingAuthClients, "test-client")
	defer os.Unsetenv(proxy.AllowMissingAuthClients)

	// Set the RBAC policy path to a test file.
	testRbacPolicyPath := "../../rego/authz.rego"
	t.Setenv(proxy.RbacPolicyEnvVar, testRbacPolicyPath)
	defer os.Unsetenv(proxy.RbacPolicyEnvVar)

	jwtStrWithoutTenant, err := generateValidJWT(t, "", allRoles)
	if err != nil {
		t.Errorf("Error signing token: %v", err)
	}
	jwtStrWithTenant, err := generateValidJWT(t, tenantUUID, allRoles)
	if err != nil {
		t.Errorf("Error signing token: %v", err)
	}
	jwtStrInvalid, err := generateValidJWT(t, "abc", allRoles)
	if err != nil {
		t.Errorf("Error signing token: %v", err)
	}
	// Create an Echo instance for testing.
	e := echo.New()

	jwtReadOnly, err := generateValidJWT(t, tenantUUID, []string{readRole})
	require.NoError(t, err, "Error signing token for read-only role")
	jwtReadWrite, err := generateValidJWT(t, tenantUUID, allRoles)
	require.NoError(t, err, "Error signing token for all roles")

	tests := []struct {
		name               string
		request            *http.Request
		expectedStatus     int
		expectedError      string
		addTenantToContext bool
	}{
		{
			name:               "No Authorization header and no allowed missing auth client",
			request:            httptest.NewRequest(http.MethodGet, "/", http.NoBody),
			expectedStatus:     http.StatusUnauthorized,
			expectedError:      "missing Authorization header",
			addTenantToContext: true,
		},
		{
			name:               "Invalid Authorization header format",
			request:            createRequestWithAuthHeader("invalid_format", "token", http.MethodGet),
			expectedStatus:     http.StatusUnauthorized,
			expectedError:      "wrong Authorization header definition",
			addTenantToContext: true,
		},
		{
			name:               "Authorization header with Bearer scheme but invalid JWT token",
			request:            createRequestWithAuthHeader("Bearer", "invalid-token", http.MethodGet),
			expectedStatus:     http.StatusUnauthorized,
			expectedError:      "JWT token is invalid or expired",
			addTenantToContext: true,
		},
		{
			name:               "Authorization header with non-Bearer scheme",
			request:            createRequestWithAuthHeader("Basic", "token", http.MethodGet),
			expectedStatus:     http.StatusUnauthorized,
			expectedError:      "Expecting \"Bearer\" Scheme to be sent",
			addTenantToContext: true,
		},
		{
			name:               "Authorization header with Bearer scheme with valid JWT token and no tenantID",
			request:            createRequestWithAuthHeader("Bearer", jwtStrWithoutTenant, http.MethodGet),
			expectedStatus:     http.StatusUnauthorized,
			expectedError:      "JWT token is valid, but tenantID was not passed in context",
			addTenantToContext: false,
		},
		{
			name:               "Authorization header with Bearer scheme with valid JWT token/tenantID but context invalid",
			request:            createRequestWithAuthHeader("Bearer", jwtStrWithTenant, http.MethodGet),
			expectedStatus:     http.StatusUnauthorized,
			expectedError:      "JWT token is valid, but tenantID was not passed in context",
			addTenantToContext: false,
		},
		{
			name:               "Authorization header with Bearer scheme with valid JWT token and tenantID",
			request:            createRequestWithAuthHeader("Bearer", jwtStrWithTenant, http.MethodGet),
			expectedStatus:     http.StatusOK,
			expectedError:      "JWT token is valid, proceeding with processing",
			addTenantToContext: true,
		},
		{
			name:               "Authorization header with Bearer scheme with valid JWT without tenantID",
			request:            createRequestWithAuthHeader("Bearer", jwtStrInvalid, http.MethodGet),
			expectedStatus:     http.StatusForbidden,
			expectedError:      "JWT token is invalid, no tenantID in JWT roles",
			addTenantToContext: true,
		},
		{
			name:               "Authorization header with Bearer scheme with valid JWT read only roles",
			request:            createRequestWithAuthHeader("Bearer", jwtReadOnly, http.MethodGet),
			expectedStatus:     http.StatusOK,
			addTenantToContext: true,
		},
		{
			name:               "Authorization header with Bearer scheme with JWT read only roles for write operation",
			request:            createRequestWithAuthHeader("Bearer", jwtReadOnly, http.MethodPut),
			expectedStatus:     http.StatusForbidden,
			addTenantToContext: true,
		},
		{
			name:               "Authorization header with Bearer scheme with JWT read write role",
			request:            createRequestWithAuthHeader("Bearer", jwtReadWrite, http.MethodGet),
			expectedStatus:     http.StatusOK,
			addTenantToContext: true,
		},
		{
			name:               "Authorization header with Bearer scheme with JWT read write role",
			request:            createRequestWithAuthHeader("Bearer", jwtReadWrite, http.MethodPut),
			expectedStatus:     http.StatusOK,
			addTenantToContext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a context with the request and recorder.
			c := e.NewContext(tt.request, httptest.NewRecorder())
			// Create a dummy next handler that returns OK status.
			next := func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			}
			if tt.addTenantToContext {
				c.SetRequest(
					c.Request().WithContext(
						tenant.AddTenantIDToContext(c.Request().Context(), tenantUUID),
					),
				)
			}
			// Invoke interceptor.
			handler := proxy.AuthenticationAuthorizationInterceptor(next)
			err := handler(c)
			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
			} else {
				var httpErr *echo.HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.expectedStatus, httpErr.Code, "Expected an HTTP 401 Unauthorized error")
				} else {
					t.Errorf("Expected an echo.HTTPError, got %T", err)
				}
			}
		})
	}
}
