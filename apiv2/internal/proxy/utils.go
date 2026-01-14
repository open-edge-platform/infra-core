// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"fmt"
	"strings"

	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/scenario"
)

func BuildAllowedClientList(scenarioName string) (map[string]struct{}, error) {

	allowedServices, ok := scenario.Allowlist[scenarioName]
	if !ok {
		err := fmt.Errorf("unknown scenario %q", scenarioName)
		zlog.Err(err).Msg("unknown scenario")
		return nil, err
	}
	allowed := make(map[string]struct{}, len(allowedServices))
	for _, serviceName := range allowedServices {
		serviceName = strings.TrimSpace(serviceName)
		if serviceName == "" {
			continue
		}
		// validate agains the list of knwon services
		if _, exists := servicesClients[serviceName]; !exists {
			zlog.Warn().Str("service", serviceName).Str("scenario", scenarioName).Msg("allowed service is not registered in proxy servicesClients map")
		} else {
			allowed[serviceName] = struct{}{}
			zlog.Debug().Str("service", serviceName).Str("scenario", scenarioName).Msg("including allowed service")
		}

	}
	return allowed, nil
}
