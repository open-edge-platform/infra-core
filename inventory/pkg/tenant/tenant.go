// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"context"
	"regexp"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/policy/rbac"
)

var zlog = logging.GetLogger("TenantInterceptor")

type CtxKey string

const (
	// CtxTenantIDKey key used in context to store tenant ID.
	CtxTenantIDKey = CtxKey("tenantID")
	// roleProjectIDSeparator "_" is used to split projectID (tenantID) from the actual role.
	roleProjectIDSeparator = "_"
	// relaxed uuid regex, replace with strict regex if required.
	uuidPattern = "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"
)

var uuidRegex = regexp.MustCompile(uuidPattern)

// GetExtractTenantIDInterceptor return an interceptor to extract tenant id from JWT roles and provide it in the context.
// The provided expectedRoles are the roles used to extract tenantID from.
// The interceptor returns error only if the tenant id is not found, invalid or missing a JWT.
// This interceptor should run after the AuthN interceptor.
func GetExtractTenantIDInterceptor(expectedRoles []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		interface{}, error,
	) {
		tenantID, err := extractProjectIDFromJWTRoles(ctx, expectedRoles)
		switch {
		case err != nil:
			return nil, err
		case tenantID == "":
			err = errors.Errorfc(
				codes.Unauthenticated,
				"Rejected because missing projectID in JWT roles: rejected=%s",
				info.FullMethod,
			)
			zlog.InfraSec().Err(err).Send()
			return nil, err
		}
		// Add tenant ID to the context only if there is any in the JWT.
		ctx = AddTenantIDToContext(ctx, tenantID)
		return handler(ctx, req)
	}
}

func extractProjectIDFromJWTRoles(ctx context.Context, expectedRoles []string) (string, error) {
	// First try to get activeprojectid from gRPC metadata (set by proxy server)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if projectIDValues := md.Get("activeprojectid"); len(projectIDValues) > 0 {
			projectID := projectIDValues[0]
			if projectID != "" && isValidUUID(projectID) {
				return projectID, nil
			}
		}
	}

	// Fallback to extracting tenant ID from JWT roles
	md, err := rbac.ExtractClaimsFromContext(ctx, false)
	if err != nil {
		return "", err
	}
	roles := md["realm_access/roles"]
	var tenantID string
	for _, role := range roles {
		if containsAny(role, expectedRoles) && strings.Contains(role, roleProjectIDSeparator) {
			// Assumption is that the first UUID before the roleProjectIDSeparator is the project ID.
			roleTID := strings.Split(role, roleProjectIDSeparator)[0]
			if !isValidUUID(roleTID) {
				// Skip invalid UUID, or roles without prefix but that contains a roleProjectIDSeparator.
				continue
			}
			if tenantID == "" {
				tenantID = roleTID
			}
			if tenantID != roleTID {
				return "", errors.Errorfc(codes.Unauthenticated, "Credentials from EN should belong to a single Project!")
			}
		}
	}
	return tenantID, nil
}

func isValidUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}

// AddTenantIDToContext Adds the given tenant ID to the given context. Tenant ID can be retrieved from the context with
// GetTenantIDFromContext function.
func AddTenantIDToContext(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, CtxTenantIDKey, tenantID)
}

// GetTenantIDFromContext Retrieves the string tenant ID from the given context and true, otherwise the empty string is
// returned and false is returned. TenantID can be added to the context with AddTenantIDToContext function.
func GetTenantIDFromContext(ctx context.Context) (string, bool) {
	tenantID, ok := ctx.Value(CtxTenantIDKey).(string)
	return tenantID, ok
}

func containsAny(mainStr string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(mainStr, substr) {
			return true
		}
	}
	return false
}

// GetAgentsRole helper function to get role used by BMAs. It can be used to feed the expected roles of the interceptor
// for most of the RMs.
func GetAgentsRole() []string {
	return []string{
		"en-agent-rw",
	}
}

// GetOnboardingRoles helper function to get role used during onboarding. It can be used to feed the expected roles of
// the interceptor in the OM.
func GetOnboardingRoles() []string {
	return []string{
		"en-agent-rw",
		"en-ob",
	}
}
