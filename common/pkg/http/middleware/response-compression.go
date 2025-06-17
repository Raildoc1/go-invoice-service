package middleware

import (
	"compress/gzip"
	"go-invoice-service/common/pkg/logging"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type ResponseCompressor struct {
	logger           *logging.ZapLogger
	compressionLevel int
}

func NewResponseCompressor(logger *logging.ZapLogger, compressionLevel int) *ResponseCompressor {
	return &ResponseCompressor{
		logger:           logger,
		compressionLevel: compressionLevel,
	}
}

type gzipWriter struct {
	http.ResponseWriter
	Writer           io.Writer
	uncompressedSize int
}

//nolint:wrapcheck // wrapping unnecessary
func (w *gzipWriter) Write(b []byte) (int, error) {
	w.uncompressedSize += len(b)
	return w.Writer.Write(b)
}

func (rc *ResponseCompressor) CreateHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			rc.logger.DebugCtx(r.Context(), "compression missed")
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, rc.compressionLevel)
		if err != nil {
			rc.logger.ErrorCtx(r.Context(), "Failed to create gzip writer", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(gz *gzip.Writer) {
			err := gz.Close()
			if err != nil {
				rc.logger.ErrorCtx(r.Context(), "Failed to close gzip writer", zap.Error(err))
				return
			}
		}(gz)

		w.Header().Set("Content-Encoding", "gzip")
		wrappedWriter := gzipWriter{ResponseWriter: w, Writer: gz}
		next.ServeHTTP(&wrappedWriter, r)

		rc.logger.DebugCtx(r.Context(), "request compressed",
			zap.Int("Uncompressed size", wrappedWriter.uncompressedSize),
		)
	})
}
