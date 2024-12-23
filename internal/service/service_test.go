package service

import (
	"testing"

	"github.com/GlebRadaev/gofermart/internal/repo"
	"github.com/GlebRadaev/gofermart/internal/service/authservice"
	"github.com/GlebRadaev/gofermart/internal/service/balanceservice"
	"github.com/GlebRadaev/gofermart/internal/service/orderservice"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := authservice.NewMockRepo(ctrl)
	mockOrderRepo := orderservice.NewMockRepo(ctrl)
	mockBalanceRepo := balanceservice.NewMockBalanceRepo(ctrl)
	mockWithdrawalRepo := balanceservice.NewMockWithdrawalRepo(ctrl)

	repos := &repo.Repositories{
		UserRepo:    mockUserRepo,
		OrderRepo:   mockOrderRepo,
		BalanceRepo: mockBalanceRepo,
		Withdrawal:  mockWithdrawalRepo,
	}

	services := New(repos)

	assert.NotNil(t, services.AuthService)
	assert.NotNil(t, services.OrderService)
	assert.NotNil(t, services.BalanceService)
}
