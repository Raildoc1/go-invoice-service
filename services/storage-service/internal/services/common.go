package services

import (
	"context"
	"database/sql"
)

type TransactionsManager interface {
	Do(ctx context.Context, f func(ctx context.Context, tx *sql.Tx) error) error
}
