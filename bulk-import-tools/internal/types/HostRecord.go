// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package types

type HostRecord struct {
	Serial     string
	UUID       string
	OSProfile  string
	Site       string
	Secure     RecordSecure
	RemoteUser string
	// Metadata is a set of key-value pairs (key=value) separated by '&' rather than a
	// JSON string to simplify the input data for the user and to avoid handling commas
	// in the input data, which is a CSV delimiter. Example: cluster-name=test&app-id=testApp
	// The data is decoded to a JSON string before being sent to the server.
	// Example: [{"key":"cluster-name","value":"test"},{"key":"app-id","value":"testApp"}]
	Metadata  string
	Error     string
	RawRecord string
}

type RecordSecure string

const (
	SecureTrue        RecordSecure = "true"
	SecureFalse       RecordSecure = "false"
	SecureUnspecified RecordSecure = ""
)

// StringToRecordSecure converts a string to a RecordSecure enum value.
func StringToRecordSecure(value string) RecordSecure {
	switch value {
	case string(SecureTrue):
		return SecureTrue
	case string(SecureFalse):
		return SecureFalse
	default:
		return SecureUnspecified
	}
}

// RecordSecureToString converts a RecordSecure enum value to a string.
func RecordSecureToString(value RecordSecure) string {
	return string(value)
}
