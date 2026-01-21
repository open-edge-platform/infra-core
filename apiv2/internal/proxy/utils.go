// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/common"
)

func BuildAllowedClientList(scenarioName string, allowlist map[string][]string) (map[string]struct{}, error) {
	// Convert servicesClients map to map[string]interface{} to be used by the common function
	knownServices := make(map[string]interface{}, len(servicesClients))
	for key := range servicesClients {
		knownServices[key] = nil
	}

	allowed, unknown, err := common.BuildAllowedHandlersList(scenarioName, allowlist, knownServices)
	if err != nil {
		zlog.Err(err).Msg("unknown scenario")
		return nil, err
	}

	// Log unknown services
	for _, serviceName := range unknown {
		zlog.Warn().Str("scenario", scenarioName).Msgf("allowed client %s is not among known clients", serviceName)
	}

	// Log registered services
	for serviceName := range allowed {
		zlog.Debug().Str("scenario", scenarioName).Msgf("including client %s in the allowed clients list", serviceName)
	}

	return allowed, nil
}
