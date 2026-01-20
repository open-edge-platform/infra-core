// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"strings"
)

// BuildAllowedHandlersList builds a map of allowed services based on the scenario
func BuildAllowedHandlersList(scenarioName string, allowlist map[string][]string, knownServices map[string]interface{}) (map[string]struct{}, []string, error) {
	allowedServices, ok := allowlist[scenarioName]
	if !ok {
		err := fmt.Errorf("unknown scenario %q", scenarioName)
		return nil, nil, err
	}

	allowed := make(map[string]struct{}, len(allowedServices))
	unknown := []string{}
	for _, serviceName := range allowedServices {
		serviceName = strings.TrimSpace(serviceName)
		if serviceName == "" {
			continue
		}

		// validate against the list of known services
		if _, exists := knownServices[serviceName]; !exists {
			unknown = append(unknown, serviceName)
		} else {
			allowed[serviceName] = struct{}{}
		}
	}

	return allowed, unknown, nil
}
