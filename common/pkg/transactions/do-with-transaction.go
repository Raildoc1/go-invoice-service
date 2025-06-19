package transactions

import (
	"context"
	"database/sql"
	"fmt"
)

type Manager struct {
	db *sql.DB
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{
		db: db,
	}
}

func (m *Manager) Do(ctx context.Context, f func(ctx context.Context, tx *sql.Tx) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	return f(ctx, tx)
}

func (m *Manager) DoOpts(
	ctx context.Context,
	opts *sql.TxOptions,
	f func(ctx context.Context, tx *sql.Tx) error,
) error {
	tx, err := m.db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	return f(ctx, tx)
}
