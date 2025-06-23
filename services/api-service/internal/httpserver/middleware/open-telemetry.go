package middleware

import (
	"go-invoice-service/api-service/internal/metrics"
	"net/http"
	"time"
)

type OpenTelemetryStats struct{}

func NewOpenTelemetryStats() *OpenTelemetryStats {
	return &OpenTelemetryStats{}
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

		metrics.IncHTTPRequestsTotal(r.Context(), r.Method, r.URL.Path, status)
		metrics.RecordHTTPRequestDuration(r.Context(), r.Method, r.URL.Path, status, requestDuration)
	})
}
