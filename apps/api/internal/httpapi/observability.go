package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type traceKey struct{}

var traceparentPattern = regexp.MustCompile(`^00-([0-9a-f]{32})-[0-9a-f]{16}-0[01]$`)

func traceCorrelation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := ""
		if match := traceparentPattern.FindStringSubmatch(r.Header.Get("traceparent")); len(match) == 2 {
			traceID = match[1]
		}
		if traceID == "" {
			raw := make([]byte, 16)
			if _, err := rand.Read(raw); err == nil {
				traceID = hex.EncodeToString(raw)
			}
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), traceKey{}, traceID)))
	})
}
func requestTraceID(ctx context.Context) string {
	value, _ := ctx.Value(traceKey{}).(string)
	return value
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *statusWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(200)
	}
	return w.ResponseWriter.Write(body)
}

type httpMetrics struct {
	requests *prometheus.CounterVec
	latency  *prometheus.HistogramVec
	errors   *prometheus.CounterVec
}

func newHTTPMetrics(reg prometheus.Registerer) *httpMetrics {
	m := &httpMetrics{requests: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "ai_operations_http_requests_total", Help: "HTTP requests."}, []string{"method", "route", "status"}), latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "ai_operations_http_request_duration_seconds", Help: "HTTP request latency."}, []string{"method", "route"}), errors: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "ai_operations_http_errors_total", Help: "HTTP error responses."}, []string{"method", "route", "status"})}
	reg.MustRegister(m.requests, m.latency, m.errors)
	return m
}
func (m *httpMetrics) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		wrapped := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(wrapped, r)
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unmatched"
		}
		status := wrapped.status
		if status == 0 {
			status = 200
		}
		label := strconv.Itoa(status)
		m.requests.WithLabelValues(r.Method, route, label).Inc()
		m.latency.WithLabelValues(r.Method, route).Observe(time.Since(started).Seconds())
		if status >= 400 {
			m.errors.WithLabelValues(r.Method, route, label).Inc()
		}
	})
}
