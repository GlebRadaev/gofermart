package orderrepo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/pg"
	"go.uber.org/zap"
)

type Repository struct {
	db        pg.Database
	txManager pg.TXManager
}

func New(db pg.Database, txManager pg.TXManager) *Repository {
	return &Repository{
		db:        db,
		txManager: txManager,
	}
}

func (r *Repository) FindByOrderNumber(ctx context.Context, orderNumber string) (*domain.Order, error) {
	query := `
        SELECT *
        FROM orders
        WHERE order_number = $1
    `
	row := r.db.QueryRow(ctx, query, orderNumber)

	var order domain.Order
	err := row.Scan(&order.ID, &order.UserID, &order.OrderNumber, &order.Status, &order.Accrual, &order.UploadedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		zap.L().Error("can't find order", zap.Error(err))
		return nil, err
	}
	return &order, nil
}

func (r *Repository) FindOrdersByUserID(ctx context.Context, userID int) ([]domain.Order, error) {
	query := `
        SELECT *
        FROM orders
        WHERE user_id = $1
        ORDER BY uploaded_at DESC
    `
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		zap.L().Error("can't get orders", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.OrderNumber, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			zap.L().Error("can't scan order row", zap.Error(err))
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *Repository) Save(ctx context.Context, order *domain.Order) error {
	query := `
        INSERT INTO orders (user_id, order_number, status, accrual, uploaded_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	err := r.txManager.Begin(ctx, func(ctx context.Context) error {
		_, err := r.db.Exec(ctx, query, order.UserID, order.OrderNumber, order.Status, order.Accrual, order.UploadedAt)
		if err != nil {
			zap.L().Error("can't save order", zap.Error(err))
			return err
		}
		return nil

	})
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, order *domain.Order) error {
	query := `
        UPDATE orders
        SET status = $1, accrual = $2
        WHERE id = $3
    `
	err := r.txManager.Begin(ctx, func(ctx context.Context) error {
		_, err := r.db.Exec(ctx, query, order.Status, order.Accrual, order.ID)
		if err != nil {
			zap.L().Error("failed to update order", zap.Error(err))
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) FindForProcessing(ctx context.Context, limit uint32) ([]domain.Order, error) {
	query := `
        SELECT *
        FROM orders
		WHERE status = 'NEW' OR status = 'PROCESSING'
        ORDER BY uploaded_at ASC
		LIMIT $1
    `
	rows, err := r.db.Query(ctx, query, int(limit))
	if err != nil {
		zap.L().Error("can't get orders for processing", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.OrderNumber, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			zap.L().Error("can't scan order row for processing", zap.Error(err))
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
