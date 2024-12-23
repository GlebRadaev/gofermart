package accrual

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/GlebRadaev/gofermart/internal/config"
	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/service/balanceservice"
	"github.com/GlebRadaev/gofermart/internal/service/orderservice"
	"github.com/GlebRadaev/gofermart/pkg/clients"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func NewMock(t *testing.T) (*Service, *orderservice.MockRepo, *balanceservice.MockBalanceRepo, *clients.MockHTTPClientI) {
	cfg := &config.Config{AccrualAddress: "http://localhost:8081"}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderRepo := orderservice.NewMockRepo(ctrl)
	balanceRepo := balanceservice.NewMockBalanceRepo(ctrl)
	client := clients.NewMockHTTPClientI(ctrl)
	service := New(cfg, orderRepo, balanceRepo, client)
	return service, orderRepo, balanceRepo, client
}

func TestService_Start(t *testing.T) {
	service, _, _, _ := NewMock(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.Start(ctx)
	time.Sleep(20 * time.Millisecond)
	cancel()
}

func TestService_processOrders(t *testing.T) {
	tests := []struct {
		name           string
		mockFindOrders func(ctx context.Context, limit uint32) ([]domain.Order, error)
		mockAddTask    func(ctx context.Context, task func() error) error
		expectedLog    string
		expectedErr    error
		orderCount     int
	}{
		{
			name: "successfully processes orders",
			mockFindOrders: func(ctx context.Context, limit uint32) ([]domain.Order, error) {
				return []domain.Order{
					{OrderNumber: "order1", Status: "NEW", UserID: 1},
					{OrderNumber: "order2", Status: "NEW", UserID: 2},
				}, nil
			},
			mockAddTask: func(ctx context.Context, task func() error) error {
				return nil
			},
			expectedLog: "Processing orders completed successfully",
			expectedErr: nil,
			orderCount:  2,
		},
		{
			name: "fails when finding orders",
			mockFindOrders: func(ctx context.Context, limit uint32) ([]domain.Order, error) {
				return nil, fmt.Errorf("failed to fetch orders for processing")
			},
			mockAddTask: func(ctx context.Context, task func() error) error {
				return nil
			},
			expectedLog: "Failed to fetch orders for processing",
			expectedErr: fmt.Errorf("failed to fetch orders for processing"),
			orderCount:  0,
		},
		{
			name: "error in workerPool AddTask",
			mockFindOrders: func(ctx context.Context, limit uint32) ([]domain.Order, error) {
				return []domain.Order{
					{OrderNumber: "order1", Status: "NEW", UserID: 1},
				}, nil
			},
			mockAddTask: func(ctx context.Context, task func() error) error {
				return fmt.Errorf("failed to add task to worker pool")
			},
			expectedLog: "Error processing orders",
			expectedErr: fmt.Errorf("failed to add task to worker pool"),
			orderCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			orderRepo := orderservice.NewMockRepo(ctrl)
			workerPool := NewMockWorkerPoolI(ctrl)

			orderRepo.EXPECT().
				FindForProcessing(gomock.Any(), gomock.Any()).
				DoAndReturn(tt.mockFindOrders).
				Times(1)
			for i := 0; i < tt.orderCount; i++ {
				workerPool.EXPECT().
					AddTask(gomock.Any(), gomock.Any()).
					DoAndReturn(tt.mockAddTask).
					AnyTimes()
			}

			service := &Service{
				orderRepo:  orderRepo,
				workerPool: workerPool,
				limit:      2,
			}

			logger := zap.NewExample()
			zap.ReplaceGlobals(logger)

			ctx := context.Background()
			service.processOrders(ctx)

			if tt.expectedErr != nil {
				assert.Error(t, tt.expectedErr, tt.expectedErr)
			}
		})
	}
}

func TestService_handleOrder(t *testing.T) {
	testCases := []struct {
		name            string
		order           domain.Order
		httpStatus      int
		responseBody    string
		expectedStatus  string
		expectedAccrual float64
		updateError     error
		balanceError    error
		expectedError   string
		cancelContext   bool
		retryError      error
		retryHeaders    http.Header
	}{
		{
			name:            "Successful processing - NEW",
			order:           domain.Order{OrderNumber: "123", Status: "NEW", UserID: 1},
			httpStatus:      http.StatusOK,
			responseBody:    `{"order":"123","status":"NEW","accrual":0}`,
			expectedStatus:  "NEW",
			expectedAccrual: 0,
		},
		{
			name:            "Successful processing - PROCESSED",
			order:           domain.Order{OrderNumber: "124", Status: "NEW", UserID: 1},
			httpStatus:      http.StatusOK,
			responseBody:    `{"order":"124","status":"PROCESSED","accrual":100.0}`,
			expectedStatus:  "PROCESSED",
			expectedAccrual: 100.0,
		},
		{
			name:          "Context canceled",
			order:         domain.Order{OrderNumber: "130", Status: "NEW", UserID: 1},
			httpStatus:    http.StatusOK,
			responseBody:  `{"order":"130","status":"NEW","accrual":0}`,
			expectedError: context.Canceled.Error(),
			cancelContext: true,
		},
		{
			name:          "Failed processing after retries",
			order:         domain.Order{OrderNumber: "127", Status: "NEW", UserID: 1},
			httpStatus:    http.StatusInternalServerError,
			responseBody:  "",
			expectedError: "failed to process order 127 after 3 retries: server error",
			retryError:    errors.New("server error"),
		},
		{
			name:          "Order not found after retries",
			order:         domain.Order{OrderNumber: "128", Status: "NEW", UserID: 1},
			httpStatus:    http.StatusNoContent,
			responseBody:  "",
			expectedError: "failed to process not found order 128 after 3 retries",
		},
		{
			name:          "Unexpected status code",
			order:         domain.Order{OrderNumber: "128", Status: "NEW", UserID: 1},
			httpStatus:    http.StatusTeapot,
			responseBody:  "",
			expectedError: "unexpected status code",
		},
		{
			name:          "Rate limit handling",
			order:         domain.Order{OrderNumber: "128", Status: "NEW", UserID: 1},
			httpStatus:    http.StatusTooManyRequests,
			responseBody:  "",
			expectedError: "",
			retryError:    nil,
			retryHeaders:  http.Header{"Retry-After": []string{"1"}},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			service, orderRepo, balanceRepo, client := NewMock(t)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if tt.cancelContext {
				cancel()
			}
			if tt.retryError != nil {
				client.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Return(tt.httpStatus, []byte(tt.responseBody), http.Header{}, tt.retryError).Times(3)
			} else {
				if tt.retryHeaders != nil {
					client.EXPECT().
						Get(gomock.Any(), gomock.Any()).
						Return(tt.httpStatus, []byte(tt.responseBody), tt.retryHeaders, nil).Times(1)
				} else {
					client.EXPECT().
						Get(gomock.Any(), gomock.Any()).
						Return(tt.httpStatus, []byte(tt.responseBody), http.Header{}, nil).
						Times(3)

				}
			}

			orderRepo.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, order *domain.Order) error {
					assert.Equal(t, tt.expectedStatus, order.Status)
					assert.Equal(t, tt.expectedAccrual, order.Accrual)
					assert.Equal(t, tt.order.OrderNumber, order.OrderNumber)
					return tt.updateError
				}).
				Times(1)
			if tt.expectedStatus == "PROCESSED" {
				balanceRepo.EXPECT().GetUserBalance(gomock.Any(), tt.order.UserID).Return(&domain.Balance{CurrentBalance: 0}, nil).Times(1)
				balanceRepo.EXPECT().
					UpdateUserBalance(gomock.Any(), tt.order.UserID, gomock.Any()).
					DoAndReturn(func(ctx context.Context, userID int, balance *domain.Balance) (*domain.Balance, error) {
						assert.Equal(t, tt.order.UserID, userID)
						assert.InDelta(t, tt.expectedAccrual, balance.CurrentBalance, 0.001)
						return &domain.Balance{CurrentBalance: tt.expectedAccrual}, tt.balanceError
					}).
					Times(1)
			}

			err := service.handleOrder(ctx, tt.order)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ProcessAccrual(t *testing.T) {
	service, orderRepo, balanceRepo, _ := NewMock(t)

	testCases := []struct {
		name            string
		order           domain.Order
		respBody        []byte
		updateErr       error
		balanceErr      error
		expectErr       bool
		expectedStatus  string
		expectedAccrual float64
	}{
		{
			name: "Successful processing - PROCESSED",
			order: domain.Order{
				OrderNumber: "123",
				UserID:      1,
				Status:      "NEW",
			},
			respBody:        []byte(`{"order":"123","status":"PROCESSED","accrual":100.5}`),
			updateErr:       nil,
			balanceErr:      nil,
			expectErr:       false,
			expectedStatus:  "PROCESSED",
			expectedAccrual: 100.5,
		},
		{
			name: "Successful processing - INVALID",
			order: domain.Order{
				OrderNumber: "456",
				UserID:      2,
				Status:      "NEW",
			},
			respBody:        []byte(`{"order":"456","status":"INVALID"}`),
			updateErr:       nil,
			balanceErr:      nil,
			expectErr:       false,
			expectedStatus:  "INVALID",
			expectedAccrual: 0,
		},
		{
			name: "Error updating order",
			order: domain.Order{
				OrderNumber: "789",
				UserID:      3,
				Status:      "NEW",
			},
			respBody:        []byte(`{"order":"789","status":"PROCESSED","accrual":50.0}`),
			updateErr:       errors.New("update error"),
			balanceErr:      nil,
			expectErr:       true,
			expectedStatus:  "PROCESSED",
			expectedAccrual: 50.0,
		},
		{
			name: "Error updating balance",
			order: domain.Order{
				OrderNumber: "101",
				UserID:      4,
				Status:      "NEW",
			},
			respBody:        []byte(`{"order":"101","status":"PROCESSED","accrual":75.5}`),
			updateErr:       nil,
			balanceErr:      errors.New("balance update error"),
			expectErr:       true,
			expectedStatus:  "PROCESSED",
			expectedAccrual: 75.5,
		},
		{
			name: "Error parsing response body",
			order: domain.Order{
				OrderNumber: "123",
				UserID:      1,
				Status:      "NEW",
			},
			respBody:        []byte(`{invalid json}`),
			updateErr:       nil,
			balanceErr:      nil,
			expectErr:       true,
			expectedStatus:  "NEW",
			expectedAccrual: 0,
		},
		{
			name: "Error: Order number mismatch",
			order: domain.Order{
				OrderNumber: "123",
				UserID:      1,
				Status:      "NEW",
			},
			respBody:        []byte(`{"order":"456","status":"PROCESSED","accrual":100.5}`),
			updateErr:       nil,
			balanceErr:      nil,
			expectErr:       true,
			expectedStatus:  "NEW",
			expectedAccrual: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectedStatus == "PROCESSED" && tc.expectedAccrual > 0 {
				balanceRepo.EXPECT().GetUserBalance(gomock.Any(), tc.order.UserID).Return(&domain.Balance{}, nil)
				balanceRepo.EXPECT().UpdateUserBalance(gomock.Any(), tc.order.UserID, gomock.Any()).Return(nil, tc.balanceErr)
			}

			orderRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, order *domain.Order) error {
				assert.Equal(t, tc.expectedStatus, order.Status)
				assert.Equal(t, tc.expectedAccrual, order.Accrual)
				return tc.updateErr
			})

			err := service.processAccrual(context.Background(), tc.order, tc.respBody)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestService_UpdateBalance(t *testing.T) {
	service, _, balanceRepo, _ := NewMock(t)

	testCases := []struct {
		name            string
		userID          int
		accrual         float64
		initialBalance  *domain.Balance
		fetchErr        error
		updateErr       error
		expectErr       bool
		expectedErrMsg  string
		expectedBalance float64
	}{
		{
			name:            "Successful balance update",
			userID:          1,
			accrual:         10.0,
			initialBalance:  &domain.Balance{CurrentBalance: 20.0},
			fetchErr:        nil,
			updateErr:       nil,
			expectErr:       false,
			expectedErrMsg:  "",
			expectedBalance: 30.0,
		},
		{
			name:           "Error fetching balance",
			userID:         1,
			accrual:        10.0,
			initialBalance: nil,
			fetchErr:       errors.New("database error"),
			updateErr:      nil,
			expectErr:      true,
			expectedErrMsg: "database error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			balanceRepo.EXPECT().GetUserBalance(gomock.Any(), tc.userID).DoAndReturn(func(ctx context.Context, id int) (*domain.Balance, error) {
				return tc.initialBalance, tc.fetchErr
			})

			if tc.fetchErr == nil {
				balanceRepo.EXPECT().UpdateUserBalance(gomock.Any(), tc.userID, gomock.Any()).DoAndReturn(func(ctx context.Context, id int, updatedBalance *domain.Balance) (interface{}, error) {
					if tc.updateErr != nil {
						return nil, tc.updateErr
					}
					assert.Equal(t, tc.userID, id)
					assert.Equal(t, tc.expectedBalance, updatedBalance.CurrentBalance)
					return nil, nil
				}).AnyTimes()
			}

			err := service.UpdateBalance(context.Background(), tc.userID, tc.accrual)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErrMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_handleRateLimit(t *testing.T) {
	service, _, _, _ := NewMock(t)

	order := domain.Order{OrderNumber: "123"}
	attempt := 1

	headers := http.Header{}
	headers.Set("Retry-After", "1")

	start := time.Now()
	err := service.handleRateLimit(order, headers, attempt)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, elapsed, 1*time.Second)
	assert.LessOrEqual(t, elapsed, 2*time.Second)

	headers = http.Header{}
	start = time.Now()
	err = service.handleRateLimit(order, headers, attempt)
	elapsed = time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, elapsed, retryInterval*time.Duration(attempt))
	assert.LessOrEqual(t, elapsed, retryInterval*time.Duration(attempt)+time.Second)
}
