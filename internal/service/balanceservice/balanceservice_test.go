package balanceservice

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func NewMock(t *testing.T) (*Service, *MockBalanceRepo, *MockWithdrawalRepo) {
	ctrl := gomock.NewController(t)
	balanceRepo := NewMockBalanceRepo(ctrl)
	withdrawalRepo := NewMockWithdrawalRepo(ctrl)
	service := New(balanceRepo, withdrawalRepo)
	defer ctrl.Finish()
	return service, balanceRepo, withdrawalRepo
}

func TestGetBalance(t *testing.T) {
	service, balanceRepo, _ := NewMock(t)
	tests := []struct {
		name            string
		userID          int
		prepareMock     func()
		expectedBalance *domain.Balance
		expectedError   error
	}{
		{
			name:   "Retrieve balance successfully",
			userID: 1,
			prepareMock: func() {
				balanceRepo.EXPECT().GetUserBalance(gomock.Any(), 1).Return(&domain.Balance{
					UserID:         1,
					CurrentBalance: 100.0,
					WithdrawnTotal: 50.0,
				}, nil)
			},
			expectedBalance: &domain.Balance{
				UserID:         1,
				CurrentBalance: 100.0,
				WithdrawnTotal: 50.0,
			},
			expectedError: nil,
		},
		{
			name:   "Error retrieving balance",
			userID: 1,
			prepareMock: func() {
				balanceRepo.EXPECT().GetUserBalance(gomock.Any(), 1).Return(nil, errors.New("db error"))
			},
			expectedBalance: nil,
			expectedError:   errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepareMock != nil {
				tt.prepareMock()
			}

			balance, err := service.GetBalance(context.Background(), tt.userID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBalance, balance)
			}
		})
	}
}

func TestCreateBalance(t *testing.T) {
	service, balanceRepo, _ := NewMock(t)

	tests := []struct {
		name           string
		userID         int
		prepareMock    func()
		expectedError  error
		expectedResult *domain.Balance
	}{
		{
			name:   "Successful balance creation",
			userID: 1,
			prepareMock: func() {
				balanceRepo.EXPECT().CreateUserBalance(gomock.Any(), 1).Return(&domain.Balance{
					UserID:         1,
					CurrentBalance: 0.0,
					WithdrawnTotal: 0.0,
				}, nil)
			},
			expectedError: nil,
			expectedResult: &domain.Balance{
				UserID:         1,
				CurrentBalance: 0.0,
				WithdrawnTotal: 0.0,
			},
		},
		{
			name:   "Failed balance creation",
			userID: 1,
			prepareMock: func() {
				balanceRepo.EXPECT().CreateUserBalance(gomock.Any(), 1).Return(nil, errors.New("failed to create balance"))
			},
			expectedError:  errors.New("failed to create balance"),
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepareMock != nil {
				tt.prepareMock()
			}

			result, err := service.CreateBalance(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestWithdraw(t *testing.T) {
	service, balanceRepo, withdrawalRepo := NewMock(t)
	tests := []struct {
		name          string
		userID        int
		orderNumber   string
		sum           float64
		prepareMock   func()
		expectedError error
	}{
		{
			name:        "Successful withdrawal",
			userID:      1,
			orderNumber: "12345",
			sum:         50.0,
			prepareMock: func() {
				balanceRepo.EXPECT().GetUserBalance(gomock.Any(), 1).Return(&domain.Balance{
					UserID:         1,
					CurrentBalance: 100.0,
					WithdrawnTotal: 20.0,
				}, nil)
				withdrawalRepo.EXPECT().CreateWithdrawal(gomock.Any(), gomock.Any()).Return(&domain.Withdrawal{}, nil)
				balanceRepo.EXPECT().UpdateUserBalance(gomock.Any(), 1, &domain.Balance{
					UserID:         1,
					CurrentBalance: 50.0,
					WithdrawnTotal: 70.0,
				}).Return(nil, nil)
			},
			expectedError: nil,
		},
		{
			name:        "Insufficient balance",
			userID:      1,
			orderNumber: "12345",
			sum:         150.0,
			prepareMock: func() {
				balanceRepo.EXPECT().GetUserBalance(gomock.Any(), 1).Return(&domain.Balance{
					UserID:         1,
					CurrentBalance: 100.0,
					WithdrawnTotal: 20.0,
				}, nil)
			},
			expectedError: ErrInsufficientBalance,
		},
		{
			name:        "Error getting user balance",
			userID:      1,
			orderNumber: "12345",
			sum:         50.0,
			prepareMock: func() {
				balanceRepo.EXPECT().GetUserBalance(gomock.Any(), 1).Return(nil, fmt.Errorf("failed to get balance"))
			},
			expectedError: fmt.Errorf("failed to get balance"),
		},
		{
			name:        "Error creating withdrawal",
			userID:      1,
			orderNumber: "12345",
			sum:         50.0,
			prepareMock: func() {
				balanceRepo.EXPECT().GetUserBalance(gomock.Any(), 1).Return(&domain.Balance{
					UserID:         1,
					CurrentBalance: 100.0,
					WithdrawnTotal: 20.0,
				}, nil)
				withdrawalRepo.EXPECT().CreateWithdrawal(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("failed to create withdrawal record"))
			},
			expectedError: fmt.Errorf("failed to create withdrawal record"),
		},
		{
			name:        "Error updating user balance",
			userID:      1,
			orderNumber: "12345",
			sum:         50.0,
			prepareMock: func() {
				balanceRepo.EXPECT().GetUserBalance(gomock.Any(), 1).Return(&domain.Balance{
					UserID:         1,
					CurrentBalance: 100.0,
					WithdrawnTotal: 20.0,
				}, nil)
				withdrawalRepo.EXPECT().CreateWithdrawal(gomock.Any(), gomock.Any()).Return(&domain.Withdrawal{}, nil)
				balanceRepo.EXPECT().UpdateUserBalance(gomock.Any(), 1, &domain.Balance{
					UserID:         1,
					CurrentBalance: 50.0,
					WithdrawnTotal: 70.0,
				}).Return(nil, fmt.Errorf("failed to update user balance"))
			},
			expectedError: fmt.Errorf("failed to update user balance"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepareMock != nil {
				tt.prepareMock()
			}

			err := service.Withdraw(context.Background(), tt.userID, tt.orderNumber, tt.sum)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetWithdrawals(t *testing.T) {
	service, _, withdrawalRepo := NewMock(t)
	time := time.Now()
	tests := []struct {
		name                string
		userID              int
		prepareMock         func()
		expectedWithdrawals []domain.Withdrawal
		expectedError       error
	}{
		{
			name:   "Retrieve withdrawals successfully",
			userID: 1,
			prepareMock: func() {
				withdrawalRepo.EXPECT().GetWithdrawalsByUserID(gomock.Any(), 1).Return([]domain.Withdrawal{
					{
						UserID:      1,
						OrderNumber: "12345",
						Sum:         50.0,
						ProcessedAt: time,
					},
				}, nil)
			},
			expectedWithdrawals: []domain.Withdrawal{
				{
					UserID:      1,
					OrderNumber: "12345",
					Sum:         50.0,
					ProcessedAt: time,
				},
			},
			expectedError: nil,
		},
		{
			name:   "Error retrieving withdrawals",
			userID: 1,
			prepareMock: func() {
				withdrawalRepo.EXPECT().GetWithdrawalsByUserID(gomock.Any(), 1).Return(nil, errors.New("db error"))
			},
			expectedWithdrawals: nil,
			expectedError:       errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepareMock != nil {
				tt.prepareMock()
			}

			withdrawals, err := service.GetWithdrawals(context.Background(), tt.userID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedWithdrawals, withdrawals)
			}
		})
	}
}
