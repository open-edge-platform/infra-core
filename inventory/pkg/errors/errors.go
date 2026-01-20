// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

//nolint:revive // Package name conflicts with standard library but provides domain-specific errors
package errors

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"buf.build/go/protovalidate"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
)

// Mapping codes to HTTP statuses.
var errorCodesToStatus = map[codes.Code]int{
	codes.Canceled:           http.StatusNotAcceptable,
	codes.NotFound:           http.StatusNotFound,
	codes.InvalidArgument:    http.StatusUnprocessableEntity,
	codes.DeadlineExceeded:   http.StatusRequestTimeout,
	codes.AlreadyExists:      http.StatusConflict,
	codes.ResourceExhausted:  http.StatusTooManyRequests,
	codes.FailedPrecondition: http.StatusPreconditionFailed,
	codes.OutOfRange:         http.StatusUnprocessableEntity,
	codes.Unimplemented:      http.StatusNotImplemented,
	codes.Unavailable:        http.StatusServiceUnavailable,
	codes.Unauthenticated:    http.StatusUnauthorized,
	codes.PermissionDenied:   http.StatusForbidden,
	codes.Unknown:            http.StatusInternalServerError,
	codes.Aborted:            http.StatusInternalServerError,
	codes.Internal:           http.StatusInternalServerError,
	codes.DataLoss:           http.StatusInternalServerError,
	codes.OK:                 http.StatusOK,
}

// Mapping reasons to gRPC codes.
var reasonToCode = map[Reason]codes.Code{
	Reason_UNKNOWN_CLIENT:        codes.PermissionDenied,
	Reason_OPERATION_IN_PROGRESS: codes.Internal,
}

// Unhandled codes.
const (
	notACode   = "Code("
	NotAnError = "NOT AN ERROR"
	entRegex   = "(?i)ent: "
	pqRegex    = "(?i)pq: "
)

// build creates a new Infra error wrapping a reason, and a
// stacktrace into a gRPC status which is converted ultimately
// into an error.
func build(reason Reason, err error) error {
	code, ok := reasonToCode[reason]
	if !ok {
		//nolint:gosec // Is it safe to assume integer overflow conversion uint32 -> int32?
		code = codes.Code(reason)
		if strings.Contains(code.String(), notACode) {
			code = codes.Unknown
		}
	}
	st := grpc_status.New(code, err.Error())
	ei := &ErrorInfo{
		Reason:     reason,
		Stacktrace: fmt.Sprintf("%+v", err),
	}
	st, err = st.WithDetails(ei)
	if err != nil {
		// If this errored, it will always error
		// here, so better panic so we can figure
		// out why than have this silently passing.
		panic(fmt.Sprintf("Unexpected error attaching details: %v", err))
	}
	return st.Err()
}

func sanitizeError(err error) string {
	// Sanitize errors by masking ent
	r := regexp.MustCompile(entRegex)
	errStr := r.ReplaceAllString(err.Error(), "")
	// Sanitize errors by masking pq
	r = regexp.MustCompile(pqRegex)
	return r.ReplaceAllString(errStr, "")
}

// Wrap wraps the error by adding context details and by carry
// it over a grpc status that can be used to log details of
// errors or to print a generic error to the extern.
//
// err is the error to be wrapped.
func Wrap(err error) error {
	if err != nil {
		errStr := sanitizeError(err)
		// Parse ent errors providing a generic mapping
		switch {
		case ent.IsValidationError(err):
			return build(Reason(codes.InvalidArgument), errors.Errorf("%s", errStr))
		case ent.IsConstraintError(err):
			return build(Reason(codes.FailedPrecondition), errors.Errorf("%s", errStr))
		case ent.IsNotFound(err):
			return build(Reason(codes.NotFound), errors.Errorf("%s", errStr))
		case ent.IsNotSingular(err):
			return build(Reason(codes.Internal), errors.Errorf("%s", errStr))
		case ent.IsNotLoaded(err):
			return build(Reason(codes.Internal), errors.Errorf("%s", errStr))
		}

		e := &protovalidate.ValidationError{}
		if errors.As(err, &e) {
			return build(Reason(codes.InvalidArgument), errors.Errorf("%s", err.Error()))
		}

		// Check if context was canceled
		if errors.Is(err, context.Canceled) {
			return build(Reason(codes.Canceled), errors.Errorf("%s", err.Error()))
		}
		// Check if it is our error and return as it is
		errorInfo := GetErrorInfo(err)
		if errorInfo != nil {
			return err
		}
		// Check if err is a grpc status error
		status, ok := grpc_status.FromError(err)
		if ok {
			//nolint:gosec // Is it safe to assume integer overflow conversion uint32 -> int32?
			return build(Reason(status.Code()), errors.Errorf("%s", status.Message()))
		}
		// Otherwise build using our internal classification
		return build(Reason(codes.Internal), errors.Errorf("%s", err.Error()))
	}
	return err
}

// Errorfc creates an error wrapping a gRPC status. Code is used
// to initialize the gRPC status. The latter that can be
// used to log details of errors or to print a generic error
// to the extern.
//
// code the gRPC code to be used in the gRPC status.
func Errorfc(code codes.Code, format string, args ...interface{}) error {
	if code == codes.OK {
		return nil
	}
	// Add context
	err := errors.Errorf(format, args...)
	return build(Reason(code), err) //nolint:gosec // Is it safe to assume integer overflow conversion uint32 -> int32?
}

// Errorfr creates an error wrapping a gRPC status. Reason
// is translated into gRPC code and included in the error
// as well to provide a more detailed reason due to the caller.
// Status that can be used to log details of errors or to
// print a generic error to the extern.
//
// reason the Reason to be carried over the gRPC status.
func Errorfr(reason Reason, format string, args ...interface{}) error {
	if reason == Reason_OK {
		return nil
	}
	// Add context
	err := errors.Errorf(format, args...)
	return build(reason, err)
}

// Errorf creates an error wrapping a gRPC status. The latter
// can be used to log details of errors or to print a generic
// error to the extern. Note this will default to an internal error.
func Errorf(format string, args ...interface{}) error {
	// Add context
	err := errors.Errorf(format, args...)
	return build(Reason(codes.Internal), err)
}

// GetErrorInfo is an helper used in the tests.
func GetErrorInfo(err error) *ErrorInfo {
	st := grpc_status.Convert(err)
	for _, detail := range st.Details() {
		if t, ok := detail.(*ErrorInfo); ok {
			return t
		}
	}
	return nil
}

// IsUnKnownClient is a helper function to check if the error
// is UNKNOWN_CLIENT which means a new registration is necessary.
func IsUnKnownClient(err error) bool {
	errorInfo := GetErrorInfo(err)
	if errorInfo != nil && errorInfo.Reason == Reason_UNKNOWN_CLIENT {
		return true
	}
	return false
}

// IsOperationInProgress is a helper function to check if the error is OPERATION_IN_PROGRESS.
func IsOperationInProgress(err error) bool {
	errorInfo := GetErrorInfo(err)
	if errorInfo != nil && errorInfo.Reason == Reason_OPERATION_IN_PROGRESS {
		return true
	}
	return false
}

// IsNotFound is a helper function to check if the error
// is gRPC NOT_FOUND which means the required resource is not found.
func IsNotFound(err error) bool {
	st := grpc_status.Convert(err)
	if st != nil && st.Code() == codes.NotFound {
		return true
	}
	return false
}

// IsCanceled is a helper function to check if the error
// is gRPC CANCELED which means the operation was canceled.
func IsCanceled(err error) bool {
	st := grpc_status.Convert(err)
	if st != nil && st.Code() == codes.Canceled {
		return true
	}
	return false
}

// IsAlreadyExists is a helper function to check if the error
// is gRPC ALREADY_EXISTS which means the required resource already exists.
func IsAlreadyExists(err error) bool {
	st := grpc_status.Convert(err)
	if st != nil && st.Code() == codes.AlreadyExists {
		return true
	}
	return false
}

// IsPermissionDenied is a helper function to check if the error
// is gRPC PERMISSION_DENIED which means the operation is not allowed.
func IsPermissionDenied(err error) bool {
	st := grpc_status.Convert(err)
	if st != nil && st.Code() == codes.PermissionDenied {
		return true
	}
	return false
}

// IsUnauthenticated is a helper function to check if the error
// is gRPC UNAUTHENTICATED which means the client is not authorized to perform operation.
func IsUnauthenticated(err error) bool {
	st := grpc_status.Convert(err)
	if st != nil && st.Code() == codes.Unauthenticated {
		return true
	}
	return false
}

func IsInvalidArgument(err error) bool {
	st := grpc_status.Convert(err)
	if st != nil && st.Code() == codes.InvalidArgument {
		return true
	}
	return false
}

func IsForeignKeyConstraintError(err error) bool {
	st := grpc_status.Convert(err)
	if strings.Contains(st.Message(), "violates foreign key constraint") ||
		strings.Contains(st.Message(), "SQLSTATE 23503") {
		return true
	}

	return false
}

func IsUniqueConstraintError(err error) bool {
	st := grpc_status.Convert(err)
	if strings.Contains(st.Message(), "violates unique constraint") ||
		strings.Contains(st.Message(), "SQLSTATE 23505") {
		return true
	}

	return false
}

func IsSQLError(err error) bool {
	st := grpc_status.Convert(err)
	return strings.Contains(st.Message(), "SQLSTATE")
}

// Consider for the future
// func Append(to, err error) error {
// }

// ErrorToString converts status into string
// without leaking details to the outside.
func ErrorToString(err error) string {
	// not a status -> Unknown
	st := grpc_status.Convert(err)
	if st == nil || st.Code() == codes.OK {
		return NotAnError
	}
	// cut the details and keep only code and reason
	return st.Message()
}

// ErrorToStringWithDetails combines reason and
// stacktrace into an error mesg that can be
// print for debug purposes.
func ErrorToStringWithDetails(err error) string {
	// not a status -> Unknown
	st := grpc_status.Convert(err)
	if st == nil || st.Code() == codes.OK {
		return NotAnError
	}
	errorInfo := GetErrorInfo(err)
	if errorInfo != nil {
		return fmt.Sprintf("%d\n%s", errorInfo.Reason, errorInfo.Stacktrace)
	}
	return st.Message()
}

// ErrorToStatus converts code into a HTTP status.
func ErrorToHTTPStatus(err error) int {
	// not a status -> Unknown
	st := grpc_status.Convert(err)
	// actual conversion
	code := st.Code()
	errorStatus, ok := errorCodesToStatus[code]
	if !ok {
		errorStatus = http.StatusInternalServerError
	}
	return errorStatus
}

// ErrorToSanitizedGrpcError generate error to grpc error code
// and message without details.
func ErrorToSanitizedGrpcError(err error) error {
	status := grpc_status.Convert(err)
	if status == nil || status.Code() == codes.OK {
		return nil
	}
	return grpc_status.Error(status.Code(), status.Message())
}

// GetSanitizeErrorGrpcInterceptor returns the unary server interceptor to sanitize errors.
func GetSanitizeErrorGrpcInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		interface{}, error,
	) {
		resp, err := handler(ctx, req)
		return resp, ErrorToSanitizedGrpcError(err)
	}
}
