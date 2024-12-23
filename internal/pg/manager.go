package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type TXManager interface {
	Begin(ctx context.Context, fn TransactionalFn) error
}

type Manager struct {
	db Database
}

func NewTXManager(db Database) *Manager {
	return &Manager{db: db}
}

type TransactionalFn func(ctx context.Context) (err error)

func (mgr *Manager) Begin(ctx context.Context, fn TransactionalFn) (err error) {
	_, ok := TXFrom(ctx)
	if ok {
		return fn(ctx)
	}

	tx, err := mgr.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pg: can't begin tx: %w", err)
	}
	ctx = With(ctx, WithTransaction(tx))
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				err = fmt.Errorf("pg: can't rollback tx: %w, original error: %w", rollbackErr, err)
			}
		}
	}()
	err = fn(ctx)
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("pg: can't commit tx: %w", err)
	}
	return
}
