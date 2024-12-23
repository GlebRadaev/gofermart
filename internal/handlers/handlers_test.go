package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/GlebRadaev/gofermart/docs"
	"github.com/GlebRadaev/gofermart/internal/handlers/auth"
	"github.com/GlebRadaev/gofermart/internal/handlers/balance"
	"github.com/GlebRadaev/gofermart/internal/handlers/orders"
	"github.com/GlebRadaev/gofermart/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	services := &service.Services{
		AuthService:    auth.NewMockService(ctrl),
		OrderService:   orders.NewMockService(ctrl),
		BalanceService: balance.NewMockService(ctrl),
	}

	h := New(services)
	assert.NotNil(t, h, "Handlers should not be nil")
}

func TestInitRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthHandler := NewMockAuthHandler(ctrl)
	mockOrderHandler := NewMockOrderHandler(ctrl)
	mockBalanceHandler := NewMockBalanceHandler(ctrl)

	mockAuthHandler.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes()
	mockAuthHandler.EXPECT().Login(gomock.Any(), gomock.Any()).AnyTimes()
	mockOrderHandler.EXPECT().AddOrder(gomock.Any(), gomock.Any()).AnyTimes()
	mockOrderHandler.EXPECT().GetOrders(gomock.Any(), gomock.Any()).AnyTimes()
	mockBalanceHandler.EXPECT().GetBalance(gomock.Any(), gomock.Any()).AnyTimes()
	mockBalanceHandler.EXPECT().Withdraw(gomock.Any(), gomock.Any()).AnyTimes()
	mockBalanceHandler.EXPECT().GetWithdrawals(gomock.Any(), gomock.Any()).AnyTimes()

	h := &Handlers{
		AuthHandler:    mockAuthHandler,
		OrderHandler:   mockOrderHandler,
		BalanceHandler: mockBalanceHandler,
	}

	router := chi.NewRouter()
	h.InitRoutes(router)

	tests := []struct {
		method string
		url    string
		status int
	}{
		{"POST", "/api/user/register", http.StatusOK},
		{"POST", "/api/user/login", http.StatusOK},
		{"POST", "/api/user/orders", http.StatusUnauthorized},
		{"GET", "/api/user/orders", http.StatusUnauthorized},
		{"GET", "/api/user/balance", http.StatusUnauthorized},
		{"POST", "/api/user/balance/withdraw", http.StatusUnauthorized},
		{"GET", "/api/user/withdrawals", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.url, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.status, rec.Code)
		})
	}
}
