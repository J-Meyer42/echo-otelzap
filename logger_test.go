package echozap

import (
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapLogger(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	}

	obs, logs := observer.New(zap.DebugLevel)

	logger := zap.New(obs)

	err := ZapLogger(logger)(h)(c)

	assert.Nil(t, err)

	logFields := logs.AllUntimed()[0].ContextMap()

	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, int64(200), logFields["status"])
	assert.NotNil(t, logFields["latency"])
	assert.Equal(t, "GET /something", logFields["request"])
	assert.NotNil(t, logFields["host"])
	assert.NotNil(t, logFields["size"])
}

func TestZapLoggerWithConfig(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	}

	obs, logs := observer.New(zap.DebugLevel)

	logger := zap.New(obs)

	err := ZapLoggerWithConfig(logger, ZapLoggerConfig{
		Skipper: func(ctx echo.Context) bool {
			return strings.Contains(ctx.Request().URL.Path, "/something")
		},
	})(h)(c)

	assert.Nil(t, err)

	assert.Equal(t, 0, logs.Len())
}

func TestOtelZapLogger(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	}

	obs, logs := observer.New(zap.DebugLevel)

	logger := zap.New(obs)

	otelLogger := otelzap.New(logger)

	err := OtelZapLogger(otelLogger)(h)(c)

	assert.Nil(t, err)

	logFields := logs.AllUntimed()[0].ContextMap()

	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, int64(200), logFields["status"])
	assert.NotNil(t, logFields["latency"])
	assert.Equal(t, "GET /something", logFields["request"])
	assert.NotNil(t, logFields["host"])
	assert.NotNil(t, logFields["size"])
}

func TestOtelZapLoggerWithConfig(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	}

	obs, logs := observer.New(zap.DebugLevel)

	logger := zap.New(obs)

	otelLogger := otelzap.New(logger)

	err := OtelZapLoggerWithConfig(otelLogger, ZapLoggerConfig{
		Skipper: func(ctx echo.Context) bool {
			return strings.Contains(ctx.Request().URL.Path, "/something")
		},
	})(h)(c)

	assert.Nil(t, err)

	assert.Equal(t, 0, logs.Len())
}

func TestOtelCustomLoggerFields(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	}

	obs, logs := observer.New(zap.DebugLevel)

	logger := zap.New(obs)

	otelLogger := otelzap.New(logger)

	// Define your custom logger fields generator function
	customFieldFunc := func(c echo.Context) []zapcore.Field {
		// Generate and return custom logger fields based on the context
		// ...
		return []zapcore.Field{
			zap.String("custom_field1", "value1"),
			zap.String("custom_field2", "value2"),
		}
	}

	// Configure the ZapLogger middleware with custom fields
	config := ZapLoggerConfig{
		CustomFieldFunc: customFieldFunc,
	}

	err := OtelZapLoggerWithConfig(otelLogger, config)(h)(c)

	assert.Nil(t, err)

	logFields := logs.AllUntimed()[0].ContextMap()

	assert.Equal(t, 1, logs.Len())
	assert.Equal(t, "value1", logFields["custom_field1"])
	assert.NotNil(t, logFields["custom_field2"])
}
