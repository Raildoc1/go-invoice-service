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

	return doInternal(ctx, tx, f)
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

	return doInternal(ctx, tx, f)
}

func doInternal(ctx context.Context, tx *sql.Tx, f func(ctx context.Context, tx *sql.Tx) error) error {
	err := f(ctx, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
