// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package errors_test

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"
	anypb "google.golang.org/protobuf/types/known/anypb"

	"github.com/open-edge-platform/infra-core/inventory/v2/internal/ent"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/errors"
	_ "github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

func TestMain(m *testing.M) {
	_ = flag.String(
		"policyBundle",
		"/rego/policy_bundle.tar.gz",
		"Path of policy rego file",
	)
	run := m.Run() // run all tests
	os.Exit(run)
}

func TestWrap(t *testing.T) {
	testCases := map[string]struct {
		inError error
		outCode codes.Code
		outDesc string
		reason  errors.Reason
		detail  string
	}{
		"ConstraintError": {
			inError: &ent.ConstraintError{},
			outCode: codes.FailedPrecondition,
			outDesc: "constraint failed: ",
			reason:  errors.Reason(codes.FailedPrecondition),
		},
		"NotFoundError": {
			inError: &ent.NotFoundError{},
			outCode: codes.NotFound,
			outDesc: " not found",
			reason:  errors.Reason(codes.NotFound),
		},
		"NotSingularError": {
			inError: &ent.NotSingularError{},
			outCode: codes.Internal,
			outDesc: " not singular",
			reason:  errors.Reason(codes.Internal),
		},
		"NotLoadedError": {
			inError: &ent.NotLoadedError{},
			outCode: codes.Internal,
			outDesc: " edge was not loaded",
			reason:  errors.Reason(codes.Internal),
		},
		"CanceledError": {
			inError: context.Canceled,
			outCode: codes.Canceled,
			outDesc: "context canceled",
			reason:  errors.Reason(codes.Canceled),
			detail:  "context canceled",
		},
		"InternalError": {
			inError: fmt.Errorf("I am an error"),
			outCode: codes.Internal,
			outDesc: "I am an error",
			reason:  errors.Reason(codes.Internal),
			detail:  "I am an error",
		},
		"StatusError": {
			inError: grpc_status.Errorf(codes.Unavailable, "I am not available"),
			outCode: codes.Unavailable,
			outDesc: "I am not available",
			reason:  errors.Reason(codes.Unavailable),
			detail:  "I am not available",
		},
		"NilError": {
			inError: nil,
			outCode: 0,
			outDesc: "",
			reason:  0,
		},
		"InfraError": {
			inError: errors.Errorfc(codes.Unavailable, "I am not available"),
			outCode: codes.Unavailable,
			outDesc: "I am not available",
			reason:  errors.Reason(codes.Unavailable),
			detail:  "I am not available",
		},
		"HostResourceValidationError": {
			inError: &protovalidate.ValidationError{},
			outCode: codes.InvalidArgument,
			outDesc: "validation error:",
			reason:  errors.Reason(codes.InvalidArgument),
			detail:  "validation error:",
		},
		"OkError": {
			inError: grpc_status.Errorf(codes.OK, "I am OK"),
			outCode: 0,
			outDesc: "",
			reason:  0,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			err := errors.Wrap(testCase.inError)

			if testCase.inError == nil && err == nil {
				return
			}

			st := grpc_status.Convert(err)
			// Code validation
			if st.Code() != testCase.outCode {
				t.Errorf("Want Code %s - Got Code %s", testCase.outCode, st.Code())
				return
			}
			// Description validation
			if st.Message() != testCase.outDesc {
				t.Errorf("Want Desc %s - Got Desc %s", testCase.outDesc, st.Message())
				return
			}
			// validate errorInfo
			errorInfo := errors.GetErrorInfo(err)
			if errorInfo == nil {
				t.Errorf("Invalid errorInfo")
				return
			}
			// validate reason
			if errorInfo.Reason != testCase.reason {
				t.Errorf("Want Reason %s - Got Reason %s", testCase.reason, errorInfo.Reason)
			}
			// validate detail
			detail := errors.ErrorToStringWithDetails(err)
			if testCase.detail != "" && !strings.Contains(detail, testCase.detail) {
				t.Errorf("Details %s does not contain %s", detail, testCase.detail)
			}
		})
	}
}

func TestErrorfr(t *testing.T) {
	testCase := []struct {
		name               string
		inMessage          string
		inReason           errors.Reason
		isReasonFunc       func(err error) bool
		expectedCode       codes.Code
		expectedHTTPStatus int
	}{
		{
			name:               "Reason_UNKNOWN_CLIENT",
			inMessage:          "I am an error",
			inReason:           errors.Reason_UNKNOWN_CLIENT,
			expectedCode:       codes.PermissionDenied,
			expectedHTTPStatus: http.StatusForbidden,
			isReasonFunc:       errors.IsUnKnownClient,
		},
		{
			name:               "Reason_OPERATION_IN_PROGRESS",
			inMessage:          "I am an error",
			inReason:           errors.Reason_OPERATION_IN_PROGRESS,
			expectedCode:       codes.Internal,
			expectedHTTPStatus: http.StatusInternalServerError,
			isReasonFunc:       errors.IsOperationInProgress,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.Errorfr(tt.inReason, "%s", tt.inMessage)

			if err == nil {
				t.Errorf("Error is nil")
			}

			st := grpc_status.Convert(err)
			// Code validation
			if st.Code() != tt.expectedCode {
				t.Errorf("Want Code %s - Got Code %s", tt.expectedCode, st.Code())
				return
			}

			// Description validation
			if st.Message() != tt.inMessage {
				t.Errorf("Want Desc %s - Got Desc %sa", tt.inMessage, st.Message())
				return
			}

			// validate errorInfo
			errorInfo := errors.GetErrorInfo(err)
			if errorInfo == nil {
				t.Errorf("Invalid errorInfo")
				return
			}
			// validate reason
			if errorInfo.Reason != tt.inReason {
				t.Errorf("Want Reason %s - Got Reason %s", tt.inReason, errorInfo.Reason)
			}

			if !tt.isReasonFunc(err) {
				t.Errorf("Want Reason %s - Got Reason %s", tt.inReason, errorInfo.Reason)
			}

			if errors.ErrorToString(err) != st.Message() {
				t.Errorf("Want %s - Got %s", st.Err().Error(), errors.ErrorToString(err))
			}

			if !strings.Contains(errors.ErrorToStringWithDetails(err), "errors.go") {
				t.Errorf("Stacktrace should countain errors.go")
			}

			if errors.ErrorToHTTPStatus(err) != tt.expectedHTTPStatus {
				t.Errorf("Expected HTTP status %d", tt.expectedHTTPStatus)
			}
		})
	}
}

func TestErrorfc(t *testing.T) {
	testCase := struct {
		inMessage string
		inCode    codes.Code
	}{
		inMessage: "I am an error",
		inCode:    codes.PermissionDenied,
	}

	t.Run("TestErrofc", func(t *testing.T) {
		err := errors.Errorfc(testCase.inCode, "%s", testCase.inMessage)

		if err == nil {
			t.Errorf("Error is nil")
		}

		st := grpc_status.Convert(err)
		// Code validation
		if st.Code() != codes.PermissionDenied {
			t.Errorf("Want Code %s - Got Code %s", codes.PermissionDenied, st.Code())
			return
		}
		// Description validation
		if st.Message() != testCase.inMessage {
			t.Errorf("Want Desc %s - Got Desc %s", testCase.inMessage, st.Message())
			return
		}
		// validate errorInfo
		errorInfo := errors.GetErrorInfo(err)
		if errorInfo == nil {
			t.Errorf("Invalid errorInfo")
			return
		}
		// validate reason
		if errorInfo.Reason != errors.Reason(codes.PermissionDenied) {
			t.Errorf("Want Reason %s - Got Reason %s", errors.Reason_UNKNOWN_CLIENT, errorInfo.Reason)
		}

		if errors.ErrorToString(err) != st.Message() {
			t.Errorf("Want %s - Got %s", st.Err().Error(), errors.ErrorToString(err))
		}

		if !strings.Contains(errors.ErrorToStringWithDetails(err), "errors.go") {
			t.Errorf("Stacktrace should countain errors.go")
		}

		if errors.ErrorToHTTPStatus(err) != http.StatusForbidden {
			t.Errorf("Expected forbidden")
		}
	})
	t.Run("NilForOK", func(t *testing.T) {
		err := errors.Errorfc(codes.OK, "")
		assert.Nil(t, err)
	})
	t.Run("InvalidErrorCodeMapsToUnknown", func(t *testing.T) {
		err := errors.Errorfc(codes.Code(1235), "")
		s := grpc_status.Convert(err)
		require.NotNil(t, s)
		assert.Equal(t, codes.Unknown, s.Code())
	})
}

func TestErrorf(t *testing.T) {
	testCase := struct {
		inMessage string
	}{
		inMessage: "I am an error",
	}

	t.Run("TestErrorf", func(t *testing.T) {
		err := errors.Errorf("%s", testCase.inMessage)

		if err == nil {
			t.Errorf("Error is nil")
		}

		st := grpc_status.Convert(err)
		// Code validation
		if st.Code() != codes.Internal {
			t.Errorf("Want Code %s - Got Code %s", codes.Internal, st.Code())
			return
		}
		// Description validation
		if st.Message() != testCase.inMessage {
			t.Errorf("Want Desc %s - Got Desc %s", testCase.inMessage, st.Message())
			return
		}
		// validate errorInfo
		errorInfo := errors.GetErrorInfo(err)
		if errorInfo == nil {
			t.Errorf("Invalid errorInfo")
			return
		}
		// validate reason
		if errorInfo.Reason != errors.Reason(codes.Internal) {
			t.Errorf("Want Reason %s - Got Reason %s", errors.Reason_UNKNOWN_CLIENT, errorInfo.Reason)
		}

		if errors.ErrorToString(err) != st.Message() {
			t.Errorf("Want %s - Got %s", st.Err().Error(), errors.ErrorToString(err))
		}

		if !strings.Contains(errors.ErrorToStringWithDetails(err), "errors.go") {
			t.Errorf("Stacktrace should countain errors.go")
		}

		if errors.ErrorToHTTPStatus(err) != http.StatusInternalServerError {
			t.Errorf("Expected forbidden")
		}
	})
	t.Run("NilForOK", func(t *testing.T) {
		err := errors.Errorfr(errors.Reason_OK, "")
		assert.Nil(t, err)
	})
}

func TestIsNotFound(t *testing.T) {
	testCases := map[string]struct {
		inError  error
		expected bool
	}{
		"grpcNotFound": {
			inError:  grpc_status.Error(codes.NotFound, "error"),
			expected: true,
		},
		"wrappedGrpcNotFound1": {
			inError:  errors.Wrap(grpc_status.Error(codes.NotFound, "error")),
			expected: true,
		},
		"wrappedGrpcNotFound2": {
			inError:  errors.Errorfc(codes.NotFound, "error"),
			expected: true,
		},
		"nilError": {
			inError:  nil,
			expected: false,
		},
		"invalidError": {
			inError:  errors.Errorf("error"),
			expected: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := errors.IsNotFound(testCase.inError)
			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestIsAlreadyExists(t *testing.T) {
	testCases := map[string]struct {
		inError  error
		expected bool
	}{
		"grpcAlreadyExists": {
			inError:  grpc_status.Error(codes.AlreadyExists, "error"),
			expected: true,
		},
		"wrappedGrpcAlreadyExists1": {
			inError:  errors.Wrap(grpc_status.Error(codes.AlreadyExists, "error")),
			expected: true,
		},
		"wrappedGrpcAlreadyExists2": {
			inError:  errors.Errorfc(codes.AlreadyExists, "error"),
			expected: true,
		},
		"nilError": {
			inError:  nil,
			expected: false,
		},
		"invalidError": {
			inError:  errors.Errorf("error"),
			expected: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := errors.IsAlreadyExists(testCase.inError)
			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestIsPermissionDenied(t *testing.T) {
	testCases := map[string]struct {
		inError  error
		expected bool
	}{
		"grpcPermissionDenied": {
			inError:  grpc_status.Error(codes.PermissionDenied, "error"),
			expected: true,
		},
		"wrappedGrpcPermissionDenied1": {
			inError:  errors.Wrap(grpc_status.Error(codes.PermissionDenied, "error")),
			expected: true,
		},
		"wrappedGrpcPermissionDenied2": {
			inError:  errors.Errorfc(codes.PermissionDenied, "error"),
			expected: true,
		},
		"nilError": {
			inError:  nil,
			expected: false,
		},
		"invalidError": {
			inError:  errors.Errorf("error"),
			expected: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := errors.IsPermissionDenied(testCase.inError)
			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestIsUnauthenticated(t *testing.T) {
	testCases := map[string]struct {
		inError  error
		expected bool
	}{
		"grpcUnauthenticated": {
			inError:  grpc_status.Error(codes.Unauthenticated, "error"),
			expected: true,
		},
		"wrappedGrpcUnauthenticated1": {
			inError:  errors.Wrap(grpc_status.Error(codes.Unauthenticated, "error")),
			expected: true,
		},
		"wrappedGrpcUnauthenticated2": {
			inError:  errors.Errorfc(codes.Unauthenticated, "error"),
			expected: true,
		},
		"nilError": {
			inError:  nil,
			expected: false,
		},
		"invalidError": {
			inError:  errors.Errorf("error"),
			expected: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := errors.IsUnauthenticated(testCase.inError)
			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestErrorToString(t *testing.T) {
	testCases := map[string]struct {
		inError  error
		expected string
	}{
		"nilError": {
			inError:  nil,
			expected: errors.NotAnError,
		},
		"okStatus": {
			inError:  grpc_status.Error(codes.OK, "OK"),
			expected: errors.NotAnError,
		},
		"okError": {
			inError:  errors.Errorfc(codes.OK, "OK"),
			expected: errors.NotAnError,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := errors.ErrorToString(testCase.inError)
			assert.Equal(t, testCase.expected, res)
		})
	}
}

func TestErrorToStringWithDetails(t *testing.T) {
	testCases := map[string]struct {
		inError  error
		expected string
	}{
		"nilError": {
			inError:  nil,
			expected: errors.NotAnError,
		},
		"okStatus": {
			inError:  grpc_status.Error(codes.OK, "OK"),
			expected: errors.NotAnError,
		},
		"okError": {
			inError:  errors.Errorfc(codes.OK, "OK"),
			expected: errors.NotAnError,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := errors.ErrorToStringWithDetails(testCase.inError)
			assert.Equal(t, testCase.expected, res)
		})
	}
}

func ExampleWrap() {
	errors.Wrap(context.Canceled)
}

func ExampleErrorf() {
	errors.Errorf("I am an %s", "error")
}

func ExampleErrorfc() {
	errors.Errorfc(codes.DeadlineExceeded, "I am an %s", "error")
}

func ExampleErrorfr() {
	errors.Errorfr(errors.Reason_UNKNOWN_CLIENT, "I am an %s", "error")
}

func ExampleIsUnKnownClient() {
	err := errors.Errorfr(errors.Reason_UNKNOWN_CLIENT, "I am a unknown client error")
	if errors.IsUnKnownClient(err) {
		fmt.Print("error is not unknown client")
	}
}

func ExampleIsNotFound() {
	err := errors.Errorfc(codes.NotFound, "I am a not found error")
	if errors.IsNotFound(err) {
		fmt.Print("error is not found")
	}
}

func ExampleIsCanceled() {
	err := errors.Errorfc(codes.Canceled, "I was canceled")
	if errors.IsCanceled(err) {
		fmt.Print("error is canceled")
	}
}

func ExampleErrorToStringWithDetails() {
	fmt.Printf("Detail %s", errors.ErrorToStringWithDetails(context.Canceled))
	// Output: Detail context canceled
}

func ExampleErrorToString() {
	fmt.Printf("Msg %s", errors.ErrorToString(context.Canceled))
	// Output: Msg context canceled
}

func ExampleErrorToHTTPStatus() {
	err := errors.Wrap(context.Canceled)
	fmt.Printf("Status %d", errors.ErrorToHTTPStatus(err))
	// Output: Status 406
}

func TestErrorToSanitizedGrpcError(t *testing.T) {
	testCases := map[string]struct {
		inError      error
		expectedDesc string
	}{
		"grpcError": {
			inError:      grpc_status.Error(codes.InvalidArgument, "invalid id"),
			expectedDesc: "invalid id",
		},
		"nilError": {
			inError:      nil,
			expectedDesc: "",
		},
	}
	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			res := errors.ErrorToSanitizedGrpcError(testCase.inError)
			if res == nil {
				assert.Equal(t, testCase.expectedDesc, "")
			} else {
				desc := strings.Split(res.Error(), "desc = ")
				assert.Equal(t, testCase.expectedDesc, desc[1])
			}
		})
	}
}

func TestErrorReasonToEnum(t *testing.T) {
	r := errors.Reason_OK
	rNew := r.Enum()
	assert.Equal(t, r, *rNew)
}

func TestErrorReasonToString(t *testing.T) {
	r := errors.Reason_OK
	rStr := r.String()
	assert.Equal(t, "OK", rStr)
}

func TestErrorInfoToString(t *testing.T) {
	eInfo := errors.ErrorInfo{Reason: errors.Reason_UNKNOWN_CLIENT}
	rStr := eInfo.String()
	assert.Equal(t, "reason:UNKNOWN_CLIENT", rStr)
}

func TestErrorInfoToGetReason(t *testing.T) {
	eInfo := errors.ErrorInfo{Reason: errors.Reason_UNKNOWN_CLIENT}
	reason := eInfo.GetReason()
	assert.Equal(t, "UNKNOWN_CLIENT", reason.String())

	var pInfo *errors.ErrorInfo
	preason := pInfo.GetReason()
	assert.Equal(t, errors.Reason_OK, preason)
}

func TestErrorInfoToGetStackTrace(t *testing.T) {
	eInfo := errors.ErrorInfo{Stacktrace: "test"}
	sTrace := eInfo.GetStacktrace()
	assert.Equal(t, "test", sTrace)

	var pInfo *errors.ErrorInfo
	pTrace := pInfo.GetStacktrace()
	assert.Equal(t, "", pTrace)
}

func TestErrorInfoToGetDetails(t *testing.T) {
	eInfo := errors.ErrorInfo{
		Details: []*anypb.Any{
			{
				Value: []byte("testDetails"),
			},
		},
	}
	sDetails := eInfo.GetDetails()
	assert.Equal(t, "testDetails", string(sDetails[0].Value))

	var pInfo *errors.ErrorInfo
	pDetails := pInfo.GetDetails()
	assert.Nil(t, pDetails)
}
