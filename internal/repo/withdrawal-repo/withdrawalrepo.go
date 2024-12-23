package withdrawalrepo

import (
	"context"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/pg"
	"go.uber.org/zap"
)

type Repository struct {
	db pg.Database
}

func New(db pg.Database) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateWithdrawal(ctx context.Context, withdrawal *domain.Withdrawal) (*domain.Withdrawal, error) {
	query := `
		INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err := r.db.QueryRow(ctx, query, withdrawal.UserID, withdrawal.OrderNumber, withdrawal.Sum, withdrawal.ProcessedAt).Scan(&withdrawal.ID)
	if err != nil {
		zap.L().Error("can't save withdrawal", zap.Error(err))
		return nil, err
	}
	return withdrawal, nil
}

func (r *Repository) GetWithdrawalsByUserID(ctx context.Context, userID int) ([]domain.Withdrawal, error) {
	query := `
        SELECT id, user_id, order_number, sum, processed_at
        FROM withdrawals
        WHERE user_id = $1
        ORDER BY processed_at DESC
    `
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		zap.L().Error("failed to fetch withdrawals", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var withdrawals []domain.Withdrawal
	for rows.Next() {
		var wd domain.Withdrawal
		err := rows.Scan(&wd.ID, &wd.UserID, &wd.OrderNumber, &wd.Sum, &wd.ProcessedAt)
		if err != nil {
			zap.L().Error("failed to scan withdrawal row", zap.Error(err))
			return nil, err
		}
		withdrawals = append(withdrawals, wd)
	}

	return withdrawals, nil
}
