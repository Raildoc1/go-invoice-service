package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"go-invoice-service/common/pkg/logging"
	"go.uber.org/zap"
	"io"
)

func CloseBody(ctx context.Context, body io.ReadCloser, logger *logging.ZapLogger) {
	err := body.Close()
	if err != nil {
		logger.ErrorCtx(ctx, "failed to close body", zap.Error(err))
	}
}

func DecodeJSON[T any](r io.Reader) (T, error) {
	var out T
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&out)
	return out, err
}

func EncodeJSON(w io.Writer, item any) error {
	output, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	_, err = w.Write(output)
	if err != nil {
		return fmt.Errorf("error writing body: %w", err)
	}

	return nil
}
