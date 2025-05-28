// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
)

// Custom error structure that includes code as string.
type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details []any  `json:"details,omitempty"`
}

// Custom error handler that converts gRPC status code to string.
func customErrorHandler(_ context.Context, _ *runtime.ServeMux,
	_ runtime.Marshaler, w http.ResponseWriter, _ *http.Request,
	err error,
) {
	// Extract gRPC status.
	st := status.Convert(err)

	// Create error response with code as string.
	errResp := &errorResponse{
		Code:    st.Code().String(), // Convert gRPC code to string.
		Message: st.Message(),
		Details: nil,
	}

	// Set HTTP status based on gRPC code
	httpStatus := runtime.HTTPStatusFromCode(st.Code())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	// Marshal and write error response
	err = json.NewEncoder(w).Encode(errResp)
	if err != nil {
		// If encoding fails, log the error but do not write to response
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
