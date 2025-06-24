package middleware

import (
	"context"
	"net/http"
	"time"
)

type HTTPMetrics interface {
	IncHTTPRequestsTotal(ctx context.Context, method, path string, status int)
	RecordHTTPRequestDuration(ctx context.Context, method, path string, status int, duration time.Duration)
}

type OpenTelemetryStats struct {
	metrics HTTPMetrics
}

func NewOpenTelemetryStats(metrics HTTPMetrics) *OpenTelemetryStats {
	return &OpenTelemetryStats{
		metrics: metrics,
	}
}

var _ http.ResponseWriter = (*ResponseWriterWrapper)(nil)

type ResponseWriterWrapper struct {
	inner  http.ResponseWriter
	status int
}

func (r *ResponseWriterWrapper) Header() http.Header {
	return r.inner.Header()
}

func (r *ResponseWriterWrapper) Write(bytes []byte) (int, error) {
	return r.inner.Write(bytes)
}

func (r *ResponseWriterWrapper) WriteHeader(statusCode int) {
	r.status = statusCode
	r.inner.WriteHeader(statusCode)
}

func (p *OpenTelemetryStats) CreateHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := ResponseWriterWrapper{
			inner: w,
		}
		next.ServeHTTP(&wrapper, r)
		status := wrapper.status
		if status == 0 {
			status = http.StatusOK
		}
		requestDuration := time.Since(start)

		p.metrics.IncHTTPRequestsTotal(r.Context(), r.Method, r.URL.Path, status)
		p.metrics.RecordHTTPRequestDuration(r.Context(), r.Method, r.URL.Path, status, requestDuration)
	})
}
