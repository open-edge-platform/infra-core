// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Tests removed: datamodel package is replaced by internal/tenancy. See tenancy/tenancy-hook.go.
package datamodel

import (
	"os"
	"testing"

	// Import logging to register globalLogLevel flag used by the test runner.
	_ "github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
