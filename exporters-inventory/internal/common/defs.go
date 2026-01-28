// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

// Package common provides shared types and constants used across the exporter.
//
//nolint:revive // Package name "common" is appropriate for shared internal types
package common

type CollectorName string

// Consts define the names of the available collector names.
// It defines those names based on the collectors available
// at the collect package.
var (
	InventoryCollector CollectorName = "INVENTORY"
)
