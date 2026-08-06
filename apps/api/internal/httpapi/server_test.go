package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

type checkFunc func(context.Context) error

func (f checkFunc) Ping(ctx context.Context) error { return f(ctx) }

func handler(dbErr, redisErr error) http.Handler {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return NewHandler(logger, Dependencies{
		Database: checkFunc(func(context.Context) error { return dbErr }),
		Redis:    checkFunc(func(context.Context) error { return redisErr }),
	}, prometheus.NewRegistry())
}

func TestHealthEndpoints(t *testing.T) {
	for _, test := range []struct {
		path string
		want int
	}{{"/health/live", 200}, {"/health/ready", 200}, {"/missing", 404}} {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler(nil, nil).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != test.want {
				t.Fatalf("status = %d, want %d", recorder.Code, test.want)
			}
			if recorder.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("content type = %q", recorder.Header().Get("Content-Type"))
			}
		})
	}
}

func TestReadinessFailsSafely(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler(errors.New("database password must not leak"), nil).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "password") {
		t.Fatal("response leaked dependency error")
	}
}

func TestErrorEnvelopeIncludesRequestID(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler(nil, nil).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/missing", nil))
	body := recorder.Body.String()
	if !strings.Contains(body, `"code":"not_found"`) || !strings.Contains(body, `"request_id":"`) {
		t.Fatalf("unexpected error envelope: %s", body)
	}
}

func TestMetricsUseBoundedRouteLabels(t *testing.T) {
	registry := prometheus.NewRegistry()
	api := NewHandler(slog.New(slog.NewJSONHandler(io.Discard, nil)), Dependencies{Database: checkFunc(func(context.Context) error { return nil }), Redis: checkFunc(func(context.Context) error { return nil })}, registry)
	api.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/unknown/12345", nil))
	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		for _, metric := range family.Metric {
			for _, label := range metric.Label {
				if label.GetName() == "route" && label.GetValue() != "unmatched" {
					t.Fatalf("unbounded route label %q", label.GetValue())
				}
			}
		}
	}
}

func TestTraceparentCorrelation(t *testing.T) {
	var captured string
	next := traceCorrelation(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { captured = requestTraceID(r.Context()) }))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	next.ServeHTTP(httptest.NewRecorder(), request)
	if captured != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("trace id=%q", captured)
	}
}
