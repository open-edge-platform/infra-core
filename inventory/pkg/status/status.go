// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"fmt"

	statusv1 "github.com/open-edge-platform/infra-core/inventory/v2/pkg/api/status/v1"
)

const (
	DefaultHostStatus               = "Not Connected"
	DefaultOnboardingStatus         = "Not Connected"
	DefaultRegistrationStatus       = "Unknown"
	DefaultInstanceStatus           = "Not Installed"
	DefaultProvisioningStatus       = "Not Provisioned"
	DefaultUpdateStatus             = "No Current Update"
	DefaultTrustedAttestationStatus = "Unknown"
)

type ResourceStatus struct {
	Status          string
	StatusIndicator statusv1.StatusIndication
}

func (rs ResourceStatus) String() string {
	return fmt.Sprintf("ResourceStatus(status=%q, indication=%v)", rs.Status, rs.StatusIndicator)
}

func New(statusMessage string, indication statusv1.StatusIndication) ResourceStatus {
	return ResourceStatus{
		Status:          statusMessage,
		StatusIndicator: indication,
	}
}
