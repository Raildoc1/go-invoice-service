package middleware

import (
	"go-invoice-service/common/pkg/logging"
	"net/http"

	"go.uber.org/zap"
)

type Logging struct{}

func NewLogging() *Logging {
	return &Logging{}
}

func (lc *Logging) CreateHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(
			logging.WithContextFields(
				r.Context(),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("remote-addr", r.RemoteAddr),
			),
		)
		next.ServeHTTP(w, r)
	})
}
