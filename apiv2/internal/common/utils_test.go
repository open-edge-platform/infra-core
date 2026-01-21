// SPDX-FileCopyrightText: (C) 2026 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	"testing"

	"github.com/open-edge-platform/infra-core/apiv2/v2/internal/common"
)

func TestBuildAllowedHandlersList(t *testing.T) {
	tests := []struct {
		name          string
		scenarioName  string
		allowlist     map[string][]string
		knownServices map[string]interface{}
		wantLen       int
		wantErr       bool
		wantServices  []string
	}{
		{
			name:         "valid scenario with services",
			scenarioName: "test-scenario-full",
			allowlist: map[string][]string{
				"test-scenario-full": {"HostService", "LocationService", "ProviderService"},
			},
			knownServices: map[string]interface{}{
				"HostService":     nil,
				"LocationService": nil,
				"ProviderService": nil,
			},
			wantLen:      3,
			wantErr:      false,
			wantServices: []string{"HostService", "LocationService", "ProviderService"},
		},
		{
			name:         "unknown scenario",
			scenarioName: "test-scenario-unknown",
			allowlist: map[string][]string{
				"test-scenario-full": {"HostService"},
			},
			knownServices: map[string]interface{}{
				"HostService": nil,
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:         "scenario with empty service names",
			scenarioName: "test-scenario-empty",
			allowlist: map[string][]string{
				"test-scenario-empty": {"HostService", "  ", "", "LocationService"},
			},
			knownServices: map[string]interface{}{
				"HostService":     nil,
				"LocationService": nil,
			},
			wantLen:      2,
			wantErr:      false,
			wantServices: []string{"HostService", "LocationService"},
		},
		{
			name:         "scenario with no services",
			scenarioName: "test-scenario-minimal",
			allowlist: map[string][]string{
				"test-scenario-minimal": {},
			},
			knownServices: map[string]interface{}{},
			wantLen:       0,
			wantErr:       false,
		},
		{
			name:         "scenario with whitespace in service names",
			scenarioName: "test-scenario-whitespace",
			allowlist: map[string][]string{
				"test-scenario-whitespace": {"  HostService  ", "LocationService "},
			},
			knownServices: map[string]interface{}{
				"HostService":     nil,
				"LocationService": nil,
			},
			wantLen:      2,
			wantErr:      false,
			wantServices: []string{"HostService", "LocationService"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := common.BuildAllowedHandlersList(tt.scenarioName, tt.allowlist, tt.knownServices)

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildAllowedHandlersList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if len(got) != tt.wantLen {
				t.Errorf("BuildAllowedHandlersList() got map length %d, want %d", len(got), tt.wantLen)
			}

			for _, service := range tt.wantServices {
				if _, exists := got[service]; !exists {
					t.Errorf("BuildAllowedHandlersList() missing service %q in result", service)
				}
			}
		})
	}
}

func TestBuildAllowedHandlersList_UnregisteredService(t *testing.T) {
	allowlist := map[string][]string{
		"test-scenario-unregistered": {"HostService", "UnknownService"},
	}

	knownServices := map[string]interface{}{
		"HostService": nil,
		// UnknownService is not in the knownServices map
	}

	list, _, err := common.BuildAllowedHandlersList("test-scenario-unregistered", allowlist, knownServices)
	if err != nil {
		t.Errorf("BuildAllowedHandlersList() unexpected error: %v", err)
	}

	if len(list) != 1 {
		t.Errorf("BuildAllowedHandlersList() got map length %d, want 1", len(list))
	}

	// Only known services should be in the allowed map
	if _, exists := list["HostService"]; !exists {
		t.Error("BuildAllowedHandlersList() missing HostService in result")
	}
	if _, exists := list["UnknownService"]; exists {
		t.Error("BuildAllowedHandlersList() should not include UnknownService in result")
	}
}
