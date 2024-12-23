package orders

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
	orderservice "github.com/GlebRadaev/gofermart/internal/service/orderservice"
	"github.com/GlebRadaev/gofermart/pkg/auth"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func NewMock(t *testing.T) (*OrderHandler, *MockService) {
	ctrl := gomock.NewController(t)
	service := NewMockService(ctrl)
	handler := New(service)
	defer ctrl.Finish()
	return handler, service
}

type errorReader struct{}

func (r *errorReader) Read([]byte) (int, error) {
	return 0, errors.New("simulated read error")
}

func TestAddOrderHandler(t *testing.T) {
	handler, service := NewMock(t)

	tests := []struct {
		name          string
		body          string
		prepareMock   func()
		expectedCode  int
		expectedError string
		expectedBody  domain.Order
	}{
		{
			name: "Successful order processing",
			body: "2404815702",
			prepareMock: func() {
				service.EXPECT().
					ProcessOrder(context.WithValue(context.Background(), auth.UserIDKey, 1), 1, "2404815702").
					Return(&domain.Order{
						UserID:      1,
						OrderNumber: "2404815702",
						Status:      "NEW",
						Accrual:     0,
					}, nil)
			},
			expectedCode: http.StatusAccepted,
			expectedBody: domain.Order{
				UserID:      1,
				OrderNumber: "2404815702",
				Status:      "NEW",
				Accrual:     0,
			},
		},
		{
			name:          "Failed to read request body",
			body:          "2404815702",
			prepareMock:   func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Failed to read request body",
		},
		{
			name:          "Invalid request body",
			body:          "",
			prepareMock:   func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Order number is required",
		},
		{
			name:          "Invalid order number",
			body:          "invalidOrder",
			prepareMock:   func() {},
			expectedCode:  http.StatusUnprocessableEntity,
			expectedError: "Invalid order number",
		},
		{
			name: "Order already exists by user",
			body: "2404815702",
			prepareMock: func() {
				service.EXPECT().
					ProcessOrder(context.WithValue(context.Background(), auth.UserIDKey, 1), 1, "2404815702").
					Return(nil, orderservice.ErrOrderAlreadyExistsByUser)
			},
			expectedCode:  http.StatusOK,
			expectedError: "order already exists",
		},
		{
			name: "Order already exists",
			body: "2404815702",
			prepareMock: func() {
				service.EXPECT().
					ProcessOrder(context.WithValue(context.Background(), auth.UserIDKey, 1), 1, "2404815702").
					Return(nil, orderservice.ErrOrderAlreadyExists)
			},
			expectedCode:  http.StatusConflict,
			expectedError: "order already exists",
		},
		{
			name: "Internal server error",
			body: "2404815702",
			prepareMock: func() {
				service.EXPECT().
					ProcessOrder(context.WithValue(context.Background(), auth.UserIDKey, 1), 1, "2404815702").
					Return(nil, errors.New("error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()

			r := httptest.NewRequest(http.MethodPost, "/order", bytes.NewBufferString(tt.body))
			if tt.name == "Failed to read request body" {
				r = httptest.NewRequest(http.MethodPost, "/order", &errorReader{})
			}
			r = r.WithContext(context.WithValue(context.Background(), auth.UserIDKey, 1))
			w := httptest.NewRecorder()

			handler.AddOrder(w, r)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
			if tt.expectedCode == http.StatusAccepted {
				var body domain.Order
				_ = json.NewDecoder(w.Body).Decode(&body)
				assert.Equal(t, tt.expectedBody, body)
			}
		})
	}
}

func TestGetOrdersHandler(t *testing.T) {
	handler, service := NewMock(t)

	tests := []struct {
		name          string
		prepareMock   func()
		expectedCode  int
		expectedError string
		expectedBody  []dto.GetOrdersResponseDTO
	}{
		{
			name: "Successful order retrieval",
			prepareMock: func() {
				service.EXPECT().
					GetOrders(context.WithValue(context.Background(), auth.UserIDKey, 1), 1).
					Return([]domain.Order{
						{
							UserID:      1,
							OrderNumber: "2404815702",
							Status:      "NEW",
							Accrual:     0,
							UploadedAt:  time.Now(),
						},
					}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: []dto.GetOrdersResponseDTO{
				{
					Number:     "2404815702",
					Status:     "NEW",
					Accrual:    0,
					UploadedAt: time.Now().Format(time.RFC3339),
				},
			},
		},
		{
			name: "No orders found",
			prepareMock: func() {
				service.EXPECT().
					GetOrders(context.WithValue(context.Background(), auth.UserIDKey, 1), 1).
					Return([]domain.Order{}, nil)
			}, expectedCode: http.StatusNoContent,
			expectedError: "No data available",
		},
		{
			name: "Internal server error",
			prepareMock: func() {
				service.EXPECT().
					GetOrders(context.WithValue(context.Background(), auth.UserIDKey, 1), 1).
					Return(nil, errors.New("error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()

			r := httptest.NewRequest(http.MethodGet, "/orders", nil)
			r = r.WithContext(context.WithValue(context.Background(), auth.UserIDKey, 1))
			w := httptest.NewRecorder()

			handler.GetOrders(w, r)

			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
			if tt.expectedCode == http.StatusOK {
				var body []dto.GetOrdersResponseDTO
				_ = json.NewDecoder(w.Body).Decode(&body)
				assert.ElementsMatch(t, tt.expectedBody, body)
			}
		})
	}
}
