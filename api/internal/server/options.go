// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/open-edge-platform/infra-core/api/internal/auth"
	api "github.com/open-edge-platform/infra-core/api/pkg/api/v0"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/auditing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/metrics"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tenant"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/tracing"
)

const (
	corsMaxAge               = 600
	rateMaxRequestsPerSecond = 200
	rateExpirePeriod         = 3
	bodyLimitMax             = "100K"
	headerLimitMax           = http.DefaultMaxHeaderBytes
	serverDefaultTimeout     = 15
	serverDefaultIdleTimeout = 60
)

func (m *Manager) setCors(e *echo.Echo) {
	log.InfraSec().Info().Msg("CORS is enabled")
	corsOrigins := strings.Split(m.cfg.RestServer.Cors, ",")
	if len(corsOrigins) > 0 {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: corsOrigins,
			AllowHeaders: []string{
				echo.HeaderAccessControlAllowOrigin,
				echo.HeaderContentType,
				echo.HeaderAuthorization,
				echo.HeaderAccept,
				echo.HeaderReferrerPolicy,
			},
			AllowCredentials: true,
			AllowMethods: []string{
				http.MethodGet,
				http.MethodHead,
				http.MethodPut,
				http.MethodPatch,
				http.MethodPost,
				http.MethodDelete,
				http.MethodOptions,
			},
			MaxAge: corsMaxAge,
		}))
	}
}

func (m *Manager) setEchoDebug(e *echo.Echo) {
	if m.cfg.RestServer.EchoDebug {
		log.InfraSec().Info().Msg("Echo logging is enabled")
		e.Use(middleware.Logger())
	}
}

func (m *Manager) setTracing(e *echo.Echo) {
	if m.cfg.Traces.EnableTracing {
		log.InfraSec().Info().Msg("Tracing is enabled")
		tracing.EnableEchoAutoTracing(e, apiTraceName)
	}
}

// setAuditing enable auditing logs via and echo middleware.
func (m *Manager) setAuditing(e *echo.Echo) {
	if m.cfg.EnableAuditing {
		log.InfraSec().Info().Msg("Auditing is enabled")
		e.Use(auditing.RestEchoMiddleware)
	}
}

func (m *Manager) setAuthentication(e *echo.Echo) {
	if m.cfg.RestServer.Authentication {
		e.Use(auth.AuthenticationAuthorizationInterceptor)
		log.InfraSec().Info().Msg("Authentication and Authorization Interceptors are enabled")
	}
}

func (m *Manager) setTenant(e *echo.Echo) {
	e.Use(tenant.TenantInterceptor)
	log.InfraSec().Info().Msg("Tenant Interceptor is enabled")
}

// setMethodOverride set config that prevents method override in HTTP header.
func (m *Manager) setMethodOverride(e *echo.Echo) {
	log.InfraSec().Info().Msg("MethodOverride nil is enabled")
	e.Pre(middleware.MethodOverrideWithConfig(middleware.MethodOverrideConfig{
		Getter: nil,
	}))
}

// setSecureConfig defines some of the secure config related to HTTP headers.
// Definitions are set based on:
// https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html
// https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html
func (m *Manager) setSecureConfig(e *echo.Echo, excludePrefixes []string) {
	log.InfraSec().Info().Msg("SecureConfig is enabled")
	secureConfig := middleware.SecureConfig{
		XFrameOptions:         "DENY",
		XSSProtection:         "0",
		ContentTypeNosniff:    "nosniff",
		ContentSecurityPolicy: "default-src 'self'; frame-ancestors 'self'; form-action 'self'",
		Skipper: func(c echo.Context) bool {
			for _, excludePrefixes := range excludePrefixes {
				if strings.Contains(c.Request().URL.String(), excludePrefixes) {
					return true
				}
			}
			return false
		},
	}

	e.Use(middleware.SecureWithConfig(secureConfig))
}

// setRateLimiter sets the rate limiter to the server.
func (m *Manager) setRateLimiter(e *echo.Echo) {
	if m.cfg.RestServer.EnableRateLimiter {
		log.InfraSec().Info().Msg("Rate Limiter is enabled")
		config := middleware.RateLimiterConfig{
			Skipper: middleware.DefaultSkipper,
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(
				middleware.RateLimiterMemoryStoreConfig{
					Rate:      rateMaxRequestsPerSecond,
					Burst:     rateMaxRequestsPerSecond,
					ExpiresIn: rateExpirePeriod * time.Minute,
				},
			),
			IdentifierExtractor: func(ctx echo.Context) (string, error) {
				id := ctx.RealIP()
				return id, nil
			},
			ErrorHandler: func(context echo.Context, _ error) error {
				return context.JSON(http.StatusForbidden, nil)
			},
			DenyHandler: func(context echo.Context, _ string, _ error) error {
				return context.JSON(http.StatusTooManyRequests, nil)
			},
		}

		e.Use(middleware.RateLimiterWithConfig(config))
	}
}

// setLimits sets the max size of a request body
// and the max size of header bytes.
func (m *Manager) setLimits(e *echo.Echo) {
	log.InfraSec().Info().Msg("Header/Body Limits are enabled")
	e.Use(middleware.BodyLimit(bodyLimitMax))
	e.Server.MaxHeaderBytes = headerLimitMax
}

// setTimeout sets the timeout of a request.
func (m *Manager) setTimeout(e *echo.Echo) {
	log.InfraSec().Info().Msg("Timeout is enabled")
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		ErrorMessage: "request timeout",
		Timeout:      serverDefaultTimeout * time.Second,
	}))
	e.Server.ReadTimeout = serverDefaultTimeout * time.Second
	e.Server.WriteTimeout = serverDefaultTimeout * time.Second
	e.Server.IdleTimeout = serverDefaultIdleTimeout * time.Second
}

// UnicodePrintableCharsChecker checks if the request body contains
// just unicode characters and returns error if it finds any non unicode characters.
func UnicodePrintableCharsChecker(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Body != nil {
			bodyBytes, err := io.ReadAll(c.Request().Body)
			if err != nil {
				log.InfraSec().InfraErr(err).Msg("request body parse io error")
				return &echo.HTTPError{
					Code:    http.StatusInternalServerError,
					Message: "request body parse error",
				}
			}
			nextReader := io.NopCloser(bytes.NewReader(bodyBytes))
			currentReader := bytes.NewReader(bodyBytes)

			for {
				r, _, err := currentReader.ReadRune()
				if err == io.EOF {
					break
				}

				if err != nil {
					log.InfraSec().InfraErr(err).Msg("request body parse error")
					return &echo.HTTPError{
						Code:    http.StatusBadRequest,
						Message: "request body parse error",
					}
				}

				if !unicode.IsPrint(r) && !unicode.IsSpace(r) {
					err := fmt.Errorf("request body contains non printable characters")
					log.InfraSec().InfraErr(err).Msg("")
					return &echo.HTTPError{
						Code:    http.StatusBadRequest,
						Message: err.Error(),
					}
				}
			}
			c.Request().Body = nextReader
		}
		return next(c)
	}
}

func UnicodePrintableCharsCheckerMiddleware() echo.MiddlewareFunc {
	return UnicodePrintableCharsChecker
}

func (m *Manager) setUnicodeChecker(e *echo.Echo) {
	log.InfraSec().Info().Msg("UnicodeChecker is enabled")
	e.Use(UnicodePrintableCharsChecker)
}

func (m *Manager) setOapiValidator(e *echo.Echo) {
	log.InfraSec().Info().Msg("OpenAPI Validator is enabled")
	openAPIDefinition, err := api.GetSwagger()
	if err != nil {
		log.InfraSec().InfraErr(err).Msgf("OpenAPI Validator failed to load OpenAPI definition")
	}

	for _, s := range openAPIDefinition.Servers {
		log.Info().Str("url", s.URL).Msgf("Servers")
		s.URL = strings.ReplaceAll(s.URL, "{apiRoot}", "")
	}

	e.Use(OapiRequestValidator(openAPIDefinition))
}

// setOptions sets all options to echo.Echo defined in this file.
func (m *Manager) setOptions(e *echo.Echo) {
	log.InfraSec().Info().Msg("Setting web server options")
	// NOTE the CORS middleware has to be the first one
	// if not OPTIONS pre-flights are denied by the OapiRequestValidator middleware
	m.setCors(e)
	m.setUnicodeChecker(e)
	m.setEchoDebug(e)
	m.setTracing(e)
	m.setTenant(e)
	m.setAuthentication(e)
	m.setAuditing(e) // TODO ITEP-2566 move before authentication
	m.setMethodOverride(e)
	m.setRateLimiter(e)
	m.setLimits(e)
	m.setTimeout(e)
	m.setSecureConfig(e, []string{})
	m.setOapiValidator(e)
	e.HideBanner = true
	e.HidePort = true
}

func (m *Manager) enableMetrics(e *echo.Echo) {
	e.Use(echoprometheus.NewMiddleware("api"))
	go func() {
		metricsSrv := echo.New()
		prometheus.MustRegister(metrics.GetClientMetricsWithLatency())
		metricsSrv.GET("/metrics", echoprometheus.NewHandler())
		if metricsErr := metricsSrv.Start(m.cfg.RestServer.MetricsAddress); metricsErr != nil &&
			!errors.Is(metricsErr, http.ErrServerClosed) {
			log.Fatal().Err(metricsErr).Msg("failed to start metrics port")
		}
	}()
}
