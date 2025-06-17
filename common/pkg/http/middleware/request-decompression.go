package middleware

import (
	"compress/gzip"
	"fmt"
	"go-invoice-service/common/pkg/logging"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type RequestDecompressor struct {
	logger *logging.ZapLogger
}

func NewRequestDecompressor(logger *logging.ZapLogger) *RequestDecompressor {
	return &RequestDecompressor{
		logger: logger,
	}
}

type compressReader struct {
	originalReader      io.ReadCloser
	decompressingReader *gzip.Reader
}

func newCompressReader(reader io.ReadCloser) (*compressReader, error) {
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	return &compressReader{
		originalReader:      reader,
		decompressingReader: gzipReader,
	}, nil
}

//nolint:wrapcheck // wrapping unnecessary
func (r *compressReader) Read(p []byte) (n int, err error) {
	return r.decompressingReader.Read(p)
}

//nolint:wrapcheck // wrapping unnecessary
func (r *compressReader) Close() error {
	if err := r.originalReader.Close(); err != nil {
		return err
	}
	return r.decompressingReader.Close()
}

func (rd *RequestDecompressor) CreateHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			next.ServeHTTP(w, r)
			return
		}

		decompressingReader, err := newCompressReader(r.Body)

		if err != nil {
			rd.logger.ErrorCtx(r.Context(), "failed to decompress", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = decompressingReader

		next.ServeHTTP(w, r)
	})
}
