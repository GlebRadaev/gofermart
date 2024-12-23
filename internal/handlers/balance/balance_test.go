package balance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/dto"
	balanceservice "github.com/GlebRadaev/gofermart/internal/service/balanceservice"
	"github.com/GlebRadaev/gofermart/pkg/auth"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func NewMock(t *testing.T) (*BalanceHandler, *MockService) {
	ctrl := gomock.NewController(t)
	service := NewMockService(ctrl)
	handler := New(service)
	defer ctrl.Finish()
	return handler, service
}

func TestGetBalanceHandler(t *testing.T) {
	handler, service := NewMock(t)
	tests := []struct {
		name          string
		body          string
		prepareMock   func()
		expectedCode  int
		expectedError string
		expectedBody  dto.BalanceResponseDTO
	}{
		{
			name: "Successful retrieval",
			prepareMock: func() {
				service.EXPECT().
					GetBalance(context.WithValue(context.Background(), auth.UserIDKey, 1), 1).
					Return(&domain.Balance{
						CurrentBalance: 100.50,
						WithdrawnTotal: 50.25,
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: dto.BalanceResponseDTO{
				Current:   100.50,
				Withdrawn: 50.25,
			},
		},
		{
			name: "Internal server error",
			prepareMock: func() {
				service.EXPECT().
					GetBalance(context.WithValue(context.Background(), auth.UserIDKey, 1), 1).
					Return(nil, errors.New("error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()
			r := httptest.NewRequest(http.MethodGet, "/balance", nil)
			r = r.WithContext(context.WithValue(context.Background(), auth.UserIDKey, 1))
			w := httptest.NewRecorder()
			handler.GetBalance(w, r)
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedCode == http.StatusOK {
				var body dto.BalanceResponseDTO
				_ = json.NewDecoder(w.Body).Decode(&body)
				assert.Equal(t, tt.expectedBody, body)
			}
		})
	}
}
func TestWithdrawHandler(t *testing.T) {
	handler, service := NewMock(t)

	tests := []struct {
		name          string
		body          string
		prepareMock   func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Successful withdrawal",
			body: `{"order":"2404815702","sum":25.5}`,
			prepareMock: func() {
				service.EXPECT().
					Withdraw(context.WithValue(context.Background(), auth.UserIDKey, 1), 1, "2404815702", 25.5).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Invalid request body",
			body: `{"order":"invalid","sum":invalid}`,
			prepareMock: func() {
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid request body",
		},
		{
			name:          "Invalid order number",
			body:          `{"order":"invalid","sum":25.5}`,
			prepareMock:   func() {},
			expectedCode:  http.StatusUnprocessableEntity,
			expectedError: "Invalid order number",
		},
		{
			name: "Insufficient balance",
			body: `{"order":"2404815702","sum":25.5}`,
			prepareMock: func() {
				service.EXPECT().
					Withdraw(context.WithValue(context.Background(), auth.UserIDKey, 1), 1, "2404815702", 25.5).
					Return(balanceservice.ErrInsufficientBalance)
			},
			expectedCode:  http.StatusPaymentRequired,
			expectedError: "insufficient balance",
		},
		{
			name: "Internal server error",
			body: `{"order":"2404815702","sum":25.5}`,
			prepareMock: func() {
				service.EXPECT().
					Withdraw(context.WithValue(context.Background(), auth.UserIDKey, 1), 1, "2404815702", 25.5).
					Return(errors.New("error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()

			r := httptest.NewRequest(http.MethodPost, "/withdraw", bytes.NewBufferString(tt.body))
			r = r.WithContext(context.WithValue(context.Background(), auth.UserIDKey, 1))
			w := httptest.NewRecorder()

			handler.Withdraw(w, r)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestGetWithdrawalsHandler(t *testing.T) {
	handler, service := NewMock(t)
	time := time.Now()

	tests := []struct {
		name          string
		body          string
		prepareMock   func()
		expectedCode  int
		expectedError string
		expectedBody  []dto.GetWithdrawalsResponseDTO
	}{
		{
			name: "Successful retrieval",
			prepareMock: func() {
				service.EXPECT().GetWithdrawals(context.WithValue(context.Background(), auth.UserIDKey, 1), 1).
					Return([]domain.Withdrawal{
						{
							OrderNumber: "123",
							Sum:         25.5,
							ProcessedAt: time,
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: []dto.GetWithdrawalsResponseDTO{
				{
					Order:       "123",
					Sum:         25.5,
					ProcessedAt: time,
				},
			},
		},
		{
			name: "No withdrawals",
			prepareMock: func() {
				service.EXPECT().GetWithdrawals(context.WithValue(context.Background(), auth.UserIDKey, 1), 1).Return([]domain.Withdrawal{}, nil)
			},
			expectedCode: http.StatusNoContent,
			expectedBody: []dto.GetWithdrawalsResponseDTO{},
		},
		{
			name: "Internal server error",
			prepareMock: func() {
				service.EXPECT().GetWithdrawals(context.WithValue(context.Background(), auth.UserIDKey, 1), 1).Return(nil, errors.New("error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()

			r := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
			r = r.WithContext(context.WithValue(context.Background(), auth.UserIDKey, 1))
			w := httptest.NewRecorder()

			handler.GetWithdrawals(w, r)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedCode == http.StatusOK {
				var body []dto.GetWithdrawalsResponseDTO
				_ = json.NewDecoder(w.Body).Decode(&body)
				assert.Equal(t, len(tt.expectedBody), len(body))
				for i := range tt.expectedBody {
					assert.Equal(t, tt.expectedBody[i].Order, body[i].Order)
					assert.Equal(t, tt.expectedBody[i].Sum, body[i].Sum)
					assert.True(t, tt.expectedBody[i].ProcessedAt.Equal(body[i].ProcessedAt))
				}
			}
		})
	}
}
