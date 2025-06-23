package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"go-invoice-service/common/pkg/timeutils"
	"storage-service/internal/data/postgres/generated/queries"
	"time"
)

var _ queries.DBTX = (*DBTXWithRetry)(nil)

type DBTXWithRetry struct {
	inner         queries.DBTX
	attemptDelays []time.Duration
	onError       func(context.Context, error)
}

func NewDBTXWithRetry(
	inner queries.DBTX,
	attemptDelays []time.Duration,
	onError func(context.Context, error),
) *DBTXWithRetry {
	return &DBTXWithRetry{
		inner:         inner,
		attemptDelays: attemptDelays,
		onError:       onError,
	}
}

func (db *DBTXWithRetry) ExecContext(ctx context.Context, query string, i ...interface{}) (sql.Result, error) {
	return timeutils.RetryRes[sql.Result](
		ctx,
		db.attemptDelays,
		func(ctx context.Context) (sql.Result, error) {
			return db.inner.ExecContext(ctx, query, i...)
		},
		db.needRetry,
		false,
	)
}

func (db *DBTXWithRetry) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return timeutils.RetryRes[*sql.Stmt](
		ctx,
		db.attemptDelays,
		func(ctx context.Context) (*sql.Stmt, error) {
			return db.inner.PrepareContext(ctx, query)
		},
		db.needRetry,
		false,
	)
}

func (db *DBTXWithRetry) QueryContext(ctx context.Context, query string, i ...interface{}) (*sql.Rows, error) {
	return timeutils.RetryRes[*sql.Rows](
		ctx,
		db.attemptDelays,
		func(ctx context.Context) (*sql.Rows, error) {
			return db.inner.QueryContext(ctx, query, i...)
		},
		db.needRetry,
		false,
	)
}

func (db *DBTXWithRetry) QueryRowContext(ctx context.Context, query string, i ...interface{}) *sql.Row {
	row, err := timeutils.RetryRes[*sql.Row](
		ctx,
		db.attemptDelays,
		func(ctx context.Context) (*sql.Row, error) {
			row := db.inner.QueryRowContext(ctx, query, i...)
			return row, row.Err()
		},
		db.needRetry,
		false,
	)

	if err != nil {
		db.onError(ctx, err)
	}

	return row
}

func (db *DBTXWithRetry) needRetry(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	db.onError(ctx, err)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// connection error
		return pgErr.Code[1] == '8'
	}
	return false
}
