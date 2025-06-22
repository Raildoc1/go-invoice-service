package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
	"go-invoice-service/api-service/internal/metrics"
	"net/http"
	"strconv"
	"time"
)

type Prometheus struct{}

func NewPrometheus() *Prometheus {
	return &Prometheus{}
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

func (p *Prometheus) CreateHandler(next http.Handler) http.Handler {
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

		labels := prometheus.Labels{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": strconv.Itoa(status),
		}

		metrics.HttpRequestsTotal.With(labels).Inc()
		metrics.HttpRequestDuration.With(labels).Observe(requestDuration.Seconds())
	})
}
