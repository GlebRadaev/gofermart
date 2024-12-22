package orderservice

import (
	"context"
	"errors"
	"time"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"go.uber.org/zap"
)

type Repo interface {
	FindByOrderNumber(ctx context.Context, orderNumber string) (*domain.Order, error)
	Save(ctx context.Context, order *domain.Order) error
	FindOrdersByUserID(ctx context.Context, userID int) ([]domain.Order, error)
	FindForProcessing(ctx context.Context, limit uint32) ([]domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
}
type Service struct {
	repo Repo
}

func New(repo Repo) *Service {
	return &Service{
		repo: repo,
	}
}

const (
	// NewOrderStatus новый заказ;
	NewOrderStatus string = "NEW"
	// RegisteredOrderStatus заказ зарегистрирован, но начисление не рассчитано;
	RegisteredOrderStatus string = "REGISTERED"
	// InvalidOrderStatus заказ не принят к расчёту, и вознаграждение не будет начислено;
	InvalidOrderStatus string = "INVALID"
	// ProcessingOrderStatus расчёт начисления в процессе;
	ProcessingOrderStatus string = "PROCESSING"
	// ProcessedOrderStatus расчёт начисления окончен;
	ProcessedOrderStatus string = "PROCESSED"
)

var (
	ErrOrderAlreadyExistsByUser = errors.New("order already exists by user")
	ErrOrderAlreadyExists       = errors.New("order already exists")
)

func (s *Service) ProcessOrder(ctx context.Context, userID int, orderNumber string) (*domain.Order, error) {
	existingOrder, err := s.repo.FindByOrderNumber(ctx, orderNumber)
	if err != nil {
		return nil, err
	}
	if existingOrder != nil {
		if existingOrder.UserID == userID {
			zap.L().Info("order already exists by user", zap.String("order_number", orderNumber))
			return nil, ErrOrderAlreadyExistsByUser
		}
		zap.L().Info("order already exists", zap.String("order_number", orderNumber))
		return nil, ErrOrderAlreadyExists
	}

	order := &domain.Order{
		UserID:      userID,
		OrderNumber: orderNumber,
		Accrual:     0,
		Status:      NewOrderStatus,
		UploadedAt:  time.Now(),
	}

	err = s.repo.Save(ctx, order)
	if err != nil {
		zap.L().Error("can't save order: ", zap.Error(err))
		return nil, err
	}

	return order, nil
}

func (s *Service) GetOrders(ctx context.Context, userID int) ([]domain.Order, error) {
	orders, err := s.repo.FindOrdersByUserID(ctx, userID)
	if err != nil {
		zap.L().Error("failed to get orders", zap.Error(err))
		return nil, err
	}
	filteredOrders := make([]domain.Order, 0)
	for _, order := range orders {
		if order.Status != RegisteredOrderStatus {
			filteredOrders = append(filteredOrders, order)
		}
	}
	if len(filteredOrders) == 0 {
		return nil, nil
	}

	return filteredOrders, nil
}
