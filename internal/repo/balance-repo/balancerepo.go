package balancerepo

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/pg"
	"go.uber.org/zap"
)

type Repository struct {
	db        pg.Database
	txManager pg.TXManager
}

func New(db pg.Database, TxManager pg.TXManager) *Repository {
	return &Repository{
		db:        db,
		txManager: TxManager,
	}
}

func (r *Repository) GetUserBalance(ctx context.Context, userID int) (*domain.Balance, error) {
	query := `
        SELECT id, user_id, current_balance, withdrawn_total
        FROM balances
        WHERE user_id = $1
    `
	row := r.db.QueryRow(ctx, query, userID)
	var balance domain.Balance
	err := row.Scan(&balance.ID, &balance.UserID, &balance.CurrentBalance, &balance.WithdrawnTotal)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		zap.L().Error("failed to get user balance", zap.Error(err))
		return nil, err
	}
	return &balance, nil
}

func (r *Repository) CreateUserBalance(ctx context.Context, userID int) (*domain.Balance, error) {
	query := `
        INSERT INTO balances (user_id, current_balance, withdrawn_total)
        VALUES ($1, 0, 0)
        RETURNING id, user_id, current_balance, withdrawn_total
    `
	row := r.db.QueryRow(ctx, query, userID)
	var balance domain.Balance
	err := row.Scan(&balance.ID, &balance.UserID, &balance.CurrentBalance, &balance.WithdrawnTotal)
	if err != nil {
		zap.L().Error("failed to create user balance", zap.Error(err))
		return nil, err
	}
	return &balance, nil
}

func (r *Repository) UpdateUserBalance(ctx context.Context, userID int, balance *domain.Balance) (*domain.Balance, error) {
	var updatedBalance domain.Balance
	query := `
		UPDATE balances
		SET current_balance = $1, withdrawn_total = $2
		WHERE user_id = $3
		RETURNING id, user_id, current_balance, withdrawn_total
	`
	err := r.txManager.Begin(ctx, func(ctx context.Context) error {
		row := r.db.QueryRow(ctx, query, balance.CurrentBalance, balance.WithdrawnTotal, userID)
		err := row.Scan(&updatedBalance.ID, &updatedBalance.UserID, &updatedBalance.CurrentBalance, &updatedBalance.WithdrawnTotal)
		if err != nil {
			zap.L().Error("failed to update user balance", zap.Error(err))
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &updatedBalance, nil
}
