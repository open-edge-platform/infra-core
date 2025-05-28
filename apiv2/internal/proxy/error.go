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

// Custom error structure that includes code as string
type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details []any  `json:"details,omitempty"`
}

// Custom error handler that converts gRPC status code to string
func customErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	// Extract gRPC status
	st := status.Convert(err)

	// Create error response with code as string
	errResp := &errorResponse{
		Code:    st.Code().String(), // Convert gRPC code to string
		Message: st.Message(),
		Details: nil,
	}

	// Set HTTP status based on gRPC code
	httpStatus := runtime.HTTPStatusFromCode(st.Code())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	// Marshal and write error response
	json.NewEncoder(w).Encode(errResp)
}
