package balanceservice

import (
	"context"
	"errors"
	"time"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"go.uber.org/zap"
)

type BalanceRepo interface {
	GetUserBalance(ctx context.Context, userID int) (*domain.Balance, error)
	CreateUserBalance(ctx context.Context, userID int) (*domain.Balance, error)
	UpdateUserBalance(ctx context.Context, userID int, balance *domain.Balance) (*domain.Balance, error)
}
type WithdrawalRepo interface {
	CreateWithdrawal(ctx context.Context, withdrawal *domain.Withdrawal) (*domain.Withdrawal, error)
	GetWithdrawalsByUserID(ctx context.Context, userID int) ([]domain.Withdrawal, error)
}

type Service struct {
	balanceRepo    BalanceRepo
	withdrawalRepo WithdrawalRepo
}

func New(balanceRepo BalanceRepo, withdrawalRepo WithdrawalRepo) *Service {
	return &Service{
		balanceRepo:    balanceRepo,
		withdrawalRepo: withdrawalRepo,
	}
}

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

func (s *Service) GetBalance(ctx context.Context, userID int) (*domain.Balance, error) {
	balance, err := s.balanceRepo.GetUserBalance(ctx, userID)
	if err != nil {
		zap.L().Error("failed to get balance", zap.Error(err))
		return nil, err
	}
	return balance, nil
}

func (s *Service) CreateBalance(ctx context.Context, userID int) (*domain.Balance, error) {
	balance, err := s.balanceRepo.CreateUserBalance(ctx, userID)
	if err != nil {
		zap.L().Error("failed to create balance", zap.Error(err))
		return nil, err
	}
	return balance, nil
}

func (s *Service) Withdraw(ctx context.Context, userID int, orderNumber string, sum float64) error {
	balance, err := s.balanceRepo.GetUserBalance(ctx, userID)
	if err != nil {
		zap.L().Error("failed to get balance", zap.Error(err))
		return err
	}

	if balance.CurrentBalance < sum {
		return ErrInsufficientBalance
	}

	withdrawal := &domain.Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		Sum:         sum,
		ProcessedAt: time.Now(),
	}

	_, err = s.withdrawalRepo.CreateWithdrawal(ctx, withdrawal)
	if err != nil {
		zap.L().Error("failed to create withdrawal record", zap.Error(err))
		return err
	}

	balance.CurrentBalance -= sum
	balance.WithdrawnTotal += sum

	if _, err := s.balanceRepo.UpdateUserBalance(ctx, userID, balance); err != nil {
		zap.L().Error("failed to update user balance", zap.Error(err))
		return err
	}

	return nil
}

func (s *Service) GetWithdrawals(ctx context.Context, userID int) ([]domain.Withdrawal, error) {
	withdrawals, err := s.withdrawalRepo.GetWithdrawalsByUserID(ctx, userID)
	if err != nil {
		zap.L().Error("failed to fetch withdrawals", zap.Error(err))
		return nil, err
	}
	return withdrawals, nil
}
