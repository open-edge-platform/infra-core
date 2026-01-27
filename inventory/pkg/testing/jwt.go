// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/policy/rbac"
)

const (
	// SharedSecretKey environment variable name for shared secret key for signing a token.
	SharedSecretKey = "SHARED_SECRET_KEY"
	secretKey       = "randomSecretKey"
	readWriteRole   = "im-rw"
	readRole        = "im-r"
	enReadWriteRole = "en-agent-rw"
)

// CreateJWT returns random signing key and JWT token (HS256 encoded) in a string with both roles, read and write.
// Only 1 token can persist in the system (otherwise, env variable holding secret key would be re-written).
func CreateJWT(tb testing.TB, tenantID string) (string, string, error) {
	tb.Helper()
	claims := &jwt.MapClaims{
		"iss": "https://keycloak.kind.internal/realms/master",
		"exp": time.Now().Add(time.Hour).Unix(),
		"typ": "Bearer",
		"realm_access": map[string]interface{}{
			"roles": []string{
				tenantID + "_" + readWriteRole,
				tenantID + "_" + readRole,
			},
		},
	}

	return CreateJWTWithClaims(tb, claims)
}

// CreateENJWT returns random signing key and JWT token (HS256 encoded) in a string with EN's read-write role.
// Only 1 token can persist in the system (otherwise, env variable holding secret key would be re-written).
func CreateENJWT(tb testing.TB, tenantIDs ...string) (string, string, error) {
	tb.Helper()
	roles := make([]string, 2, 2+len(tenantIDs)) //nolint:mnd // Size of default roles
	roles[0] = "default-roles-master"
	roles[1] = "rs-access-r"

	for _, tid := range tenantIDs {
		roles = append(
			roles,
			tid+"_"+enReadWriteRole,
		)
	}

	claims := &jwt.MapClaims{
		"iss": "https://keycloak.kind.internal/realms/master",
		"exp": time.Now().Add(time.Hour).Unix(),
		"typ": "Bearer",
		"realm_access": map[string]interface{}{
			"roles": roles,
		},
	}

	return CreateJWTWithClaims(tb, claims)
}

// CreateJWTWithReadRole returns random signing key and JWT token (HS256 encoded) in a string with only read role.
// Only 1 token can persist in the system (otherwise, env variable holding secret key would be re-written).
func CreateJWTWithReadRole(tb testing.TB, tenantID string) (string, string, error) {
	tb.Helper()
	claims := &jwt.MapClaims{
		"iss": "https://keycloak.kind.internal/realms/master",
		"exp": time.Now().Add(time.Hour).Unix(),
		"typ": "Bearer",
		"realm_access": map[string]interface{}{
			"roles": []string{
				tenantID + "_" + readRole,
			},
		},
	}

	return CreateJWTWithClaims(tb, claims)
}

// CreateJWTWithReadWriteRole returns random signing key and JWT token (HS256 encoded) in a string with read-write role.
// Only 1 token can persist in the system (otherwise, env variable holding secret key would be re-written).
func CreateJWTWithReadWriteRole(tb testing.TB, tenantID string) (string, string, error) {
	tb.Helper()
	claims := &jwt.MapClaims{
		"iss": "https://keycloak.kind.internal/realms/master",
		"exp": time.Now().Add(time.Hour).Unix(),
		"typ": "Bearer",
		"realm_access": map[string]interface{}{
			"roles": []string{
				tenantID + "_" + readWriteRole,
			},
		},
	}

	return CreateJWTWithClaims(tb, claims)
}

// CreateJWTWithClaims returns random signing key and JWT token (HS256 encoded) in a string with defined claims.
func CreateJWTWithClaims(tb testing.TB, claims *jwt.MapClaims) (string, string, error) {
	tb.Helper()

	tb.Setenv(SharedSecretKey, secretKey)
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims)
	jwtStr, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", err
	}
	return secretKey, jwtStr, nil
}

// CreateContextWithJWT can be used only with test clients, which send the request to the server.
func CreateContextWithJWT(tb testing.TB, tenantID string) (context.Context, context.CancelFunc) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ctx = CreateOutgoingContextWithJWT(tb, ctx, tenantID)
	return ctx, cancel
}

// CreateContextWithENJWT can be used only with test clients, which send the request to the server.
func CreateContextWithENJWT(tb testing.TB, tenantIDs ...string) (context.Context, context.CancelFunc) {
	tb.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ctx = CreateOutgoingContextWithENJWT(tb, ctx, tenantIDs...)
	return ctx, cancel
}

//nolint:revive // obsolete: this is testing function, testing env should be first parameter
func CreateIncomingContextWithJWT(tb testing.TB, ctx context.Context, tenantID string) context.Context {
	tb.Helper()

	// adding EN's JWT token to the context
	_, jwtToken, err := CreateJWT(tb, tenantID)
	require.NoError(tb, err)
	return rbac.AddJWTToTheIncomingContext(ctx, jwtToken)
}

//nolint:revive // obsolete: this is testing function, testing env should be first parameter
func CreateIncomingContextWithENJWT(tb testing.TB, ctx context.Context, tenantIDs ...string) context.Context {
	tb.Helper()

	// adding EN's JWT token to the context
	_, jwtToken, err := CreateENJWT(tb, tenantIDs...)
	require.NoError(tb, err)
	return rbac.AddJWTToTheIncomingContext(ctx, jwtToken)
}

//nolint:revive // obsolete: this is testing function, testing env should be first parameter
func CreateOutgoingContextWithENJWT(tb testing.TB, ctx context.Context, tenantIDs ...string) context.Context {
	tb.Helper()

	// adding EN's JWT token to the context
	_, jwtToken, err := CreateENJWT(tb, tenantIDs...)
	require.NoError(tb, err)
	return rbac.AddJWTToTheOutgoingContext(ctx, jwtToken)
}

//nolint:revive // obsolete: this is testing function, testing env should be first parameter
func CreateOutgoingContextWithJWT(tb testing.TB, ctx context.Context, tenantID string) context.Context {
	tb.Helper()

	// adding EN's JWT token to the context
	_, jwtToken, err := CreateJWT(tb, tenantID)
	require.NoError(tb, err)
	return rbac.AddJWTToTheOutgoingContext(ctx, jwtToken)
}
