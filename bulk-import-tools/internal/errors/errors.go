// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"fmt"
)

type ErrorCode int

const (
	ErrNoComment ErrorCode = iota + 1
	ErrOneFieldRequired
	ErrInvalidSN
	ErrInvalidUUID
	ErrInvalidSite
	ErrInvalidOSProfile
	ErrInvalidLocalAccount
	ErrInvalidMetadata
	ErrDuplicateSN
	ErrDuplicateUUID
	ErrPermission
	ErrFileRW
	ErrInternal
	ErrCheckFailed
	ErrFileCreate
	ErrImportFailed
	ErrRegisterFailed
	ErrInstanceFailed
	ErrHostSiteMetadataFailed
	ErrAuthNFailed
	ErrURL
	ErrAlreadyRegistered
	ErrHTTPReq
	ErrOSSecurityMismatch
)

var errorMessages = map[ErrorCode]string{
	ErrNoComment:              "Error not empty",
	ErrOneFieldRequired:       "One of Serial number or UUID required",
	ErrInvalidSN:              "Invalid Serial number",
	ErrInvalidUUID:            "Invalid UUID",
	ErrInvalidSite:            "Invalid Site",
	ErrInvalidOSProfile:       "Invalid OS profile",
	ErrInvalidLocalAccount:    "Invalid Local Account",
	ErrInvalidMetadata:        "Invalid Metadata",
	ErrDuplicateSN:            "Duplicate Serial number",
	ErrDuplicateUUID:          "Duplicate UUID",
	ErrPermission:             "Permission error",
	ErrFileRW:                 "File read/write error",
	ErrInternal:               "Internal error",
	ErrCheckFailed:            "Pre-flight check failed",
	ErrFileCreate:             "File creation error",
	ErrImportFailed:           "Failed to import all hosts",
	ErrRegisterFailed:         "Failed to register host",
	ErrInstanceFailed:         "Failed to create instance",
	ErrHostSiteMetadataFailed: "Failed to allocate site or metadata",
	ErrAuthNFailed:            "Failed to authenticate with server",
	ErrURL:                    "Malformed server URL",
	ErrAlreadyRegistered:      "Host already registered",
	ErrHTTPReq:                "HTTP request error",
	ErrOSSecurityMismatch:     "OS Profile and Security feature mismatch",
}

type CustomError struct {
	Code    ErrorCode
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}

func NewCustomError(code ErrorCode) error {
	msg, ok := errorMessages[code]
	if !ok {
		return fmt.Errorf("unknown error code: %d", code)
	}
	return &CustomError{
		Code:    code,
		Message: msg,
	}
}

func Is(code ErrorCode, err error) bool {
	customErr := new(CustomError)
	if errors.As(err, &customErr) {
		return customErr.Code == code
	}
	return false
}
