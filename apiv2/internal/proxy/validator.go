// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type Echokeys string

const (
	EchoContextKey Echokeys = "oapi-codegen/echo-context"
	UserDataKey    Echokeys = "oapi-codegen/user-data"
)

// OapiRequestValidator creates a validator from a swagger object.
func OapiRequestValidator(swagger *openapi3.T) echo.MiddlewareFunc {
	return OapiRequestValidatorWithOptions(swagger, &Options{SilenceServersWarning: true})
}

// ErrorHandler is called when there is an error in validation.
type ErrorHandler func(c echo.Context, err *echo.HTTPError) error

// MultiErrorHandler is called when oapi returns a MultiError type.
type MultiErrorHandler func(openapi3.MultiError) *echo.HTTPError

// Options to customize request validation. These are passed through to
// openapi3filter.
type Options struct {
	ErrorHandler          ErrorHandler
	Options               openapi3filter.Options
	ParamDecoder          openapi3filter.ContentParameterDecoder
	UserData              interface{}
	Skipper               echomiddleware.Skipper
	MultiErrorHandler     MultiErrorHandler
	SilenceServersWarning bool
}

// OapiRequestValidatorWithOptions creates a validator from a swagger object, with validation options.
func OapiRequestValidatorWithOptions(swagger *openapi3.T, options *Options) echo.MiddlewareFunc {
	if swagger.Servers != nil && (options == nil || !options.SilenceServersWarning) {
		zlog.InfraSec().Warn().Msg("OapiRequestValidatorWithOptions called with an OpenAPI spec that has `Servers` set." +
			"This may lead to an HTTP 400 with `no matching operation was found` when sending a valid request, " +
			"as the validator performs `Host` header validation. If you're expecting `Host` header validation, " +
			"you can silence this warning by setting `Options.SilenceServersWarning = true`.")
	}

	router, err := gorillamux.NewRouter(swagger)
	if err != nil {
		panic(err)
	}

	skipper := getSkipperFromOptions(options)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			err := ValidateRequestFromContext(c, router, options)
			if err != nil {
				if options != nil && options.ErrorHandler != nil {
					return options.ErrorHandler(c, err)
				}
				return err
			}
			return next(c)
		}
	}
}

func parseValidateErr(err error, options *Options) *echo.HTTPError {
	if err != nil {
		me := openapi3.MultiError{}
		if errors.As(err, &me) {
			errFunc := getMultiErrorHandlerFromOptions(options)
			return errFunc(me)
		}
		var reqErr *openapi3filter.RequestError
		var secErr *openapi3filter.SecurityRequirementsError
		switch {
		case errors.As(err, &reqErr):
			// We've got a bad request
			// Split up the verbose error by lines and return the first one
			// openapi errors seem to be multi-line with a decent message on the first
			errorLines := strings.Split(reqErr.Error(), "\n")
			return &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  errorLines[0],
				Internal: err,
			}
		case errors.As(err, &secErr):
			for _, err := range secErr.Errors {
				var httpErr *echo.HTTPError
				if errors.As(err, &httpErr) {
					return httpErr
				}
			}
			return &echo.HTTPError{
				Code:     http.StatusForbidden,
				Message:  secErr.Error(),
				Internal: err,
			}
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  fmt.Sprintf("error validating request: %s", err),
				Internal: err,
			}
		}
	}
	return nil
}

// ValidateRequestFromContext is called from the middleware above and actually does the work
// of validating a request.
func ValidateRequestFromContext(ctx echo.Context, router routers.Router, options *Options) *echo.HTTPError {
	req := ctx.Request()
	route, pathParams, err := router.FindRoute(req)
	// We failed to find a matching route for the request.
	var routeErr *routers.RouteError
	if err != nil {
		switch {
		case errors.As(err, &routeErr):
			// We've got a bad request, the path requested doesn't match
			// either server, or path, or something.
			return echo.NewHTTPError(http.StatusNotFound, routeErr.Reason)
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return echo.NewHTTPError(http.StatusInternalServerError,
				fmt.Sprintf("error validating route: %s", err.Error()))
		}
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	// Pass the Echo context into the request validator, so that any callbacks
	// which it invokes make it available.
	requestContext := context.WithValue(context.Background(), EchoContextKey, ctx)

	if options != nil {
		validationInput.Options = &options.Options
		validationInput.ParamDecoder = options.ParamDecoder
		requestContext = context.WithValue(requestContext, UserDataKey, options.UserData)
	}

	err = openapi3filter.ValidateRequest(requestContext, validationInput)
	return parseValidateErr(err, options)
}

// attempt to get the skipper from the options whether it is set or not.
func getSkipperFromOptions(options *Options) echomiddleware.Skipper {
	if options == nil {
		return echomiddleware.DefaultSkipper
	}

	if options.Skipper == nil {
		return echomiddleware.DefaultSkipper
	}

	return options.Skipper
}

// attempt to get the MultiErrorHandler from the options. If it is not set,
// return a default handler.
func getMultiErrorHandlerFromOptions(options *Options) MultiErrorHandler {
	if options == nil {
		return defaultMultiErrorHandler
	}

	if options.MultiErrorHandler == nil {
		return defaultMultiErrorHandler
	}

	return options.MultiErrorHandler
}

// defaultMultiErrorHandler returns a StatusBadRequest (400) and a list
// of all of the errors. This method is called if there are no other
// methods defined on the options.
func defaultMultiErrorHandler(me openapi3.MultiError) *echo.HTTPError {
	return &echo.HTTPError{
		Code:     http.StatusBadRequest,
		Message:  me.Error(),
		Internal: me,
	}
}
