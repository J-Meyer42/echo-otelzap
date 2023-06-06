package echozap

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Skipper func(c echo.Context) bool

	// ZapLoggerConfig defines the config for ZapLogger middleware
	ZapLoggerConfig struct {
		Skipper         Skipper
		CustomFieldFunc func(c echo.Context) []zapcore.Field
	}
)

var (
	// DefaultZapLoggerConfig is the default ZapLogger middleware config.
	DefaultZapLoggerConfig = ZapLoggerConfig{
		Skipper: DefaultSkipper,
	}
)

// DefaultSkipper returns false which processes the middleware
func DefaultSkipper(echo.Context) bool {
	return false
}

// ZapLogger is a middleware and zap to provide an "access log" like logging for each request.
func ZapLogger(log *zap.Logger) echo.MiddlewareFunc {
	return ZapLoggerWithConfig(log, DefaultZapLoggerConfig)
}

// OtelZapLogger is a middleware and zap to provide an "access log" like logging and opentelemetry support for each request.
func OtelZapLogger(log *otelzap.Logger) echo.MiddlewareFunc {
	return OtelZapLoggerWithConfig(log, DefaultZapLoggerConfig)
}

// ZapLoggerWithConfig is a middleware (with configuration) and zap to provide an "access log" like logging for each request.
func ZapLoggerWithConfig(log *zap.Logger, config ZapLoggerConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			fields := []zapcore.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("latency", time.Since(start).String()),
				zap.String("host", req.Host),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
				zap.String("user_agent", req.UserAgent()),
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			fields = append(fields, zap.String("request_id", id))

			// Append custom logger fields if provided
			if config.CustomFieldFunc != nil {
				fields = append(fields, config.CustomFieldFunc(c)...)
			}

			n := res.Status
			switch {
			case n >= 500:
				log.With(zap.Error(err)).Error("Server error", fields...)
			case n >= 400:
				log.With(zap.Error(err)).Warn("Client error", fields...)
			case n >= 300:
				log.Info("Redirection", fields...)
			default:
				log.Info("Success", fields...)
			}

			return nil
		}
	}
}

// OtelZapLoggerWithConfig is a middleware (with configuration) and otelzap to provide an "access log" like logging and opentelemetry support for each request.
func OtelZapLoggerWithConfig(log *otelzap.Logger, config ZapLoggerConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		// Defaults
		if config.Skipper == nil {
			config.Skipper = DefaultZapLoggerConfig.Skipper
		}

		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()

			fields := []zapcore.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("latency", time.Since(start).String()),
				zap.String("host", req.Host),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
				zap.String("user_agent", req.UserAgent()),
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			fields = append(fields, zap.String("request_id", id))

			// Append custom logger fields if provided
			if config.CustomFieldFunc != nil {
				fields = append(fields, config.CustomFieldFunc(c)...)
			}

			n := res.Status
			switch {
			case n >= 500:
				log.With(zap.Error(err)).Error("Server error", fields...)
			case n >= 400:
				log.With(zap.Error(err)).Warn("Client error", fields...)
			case n >= 300:
				log.Info("Redirection", fields...)
			default:
				log.Info("Success", fields...)
			}

			return nil
		}
	}
}
