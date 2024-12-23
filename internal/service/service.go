package service

import (
	"github.com/GlebRadaev/gofermart/internal/handlers/auth"
	"github.com/GlebRadaev/gofermart/internal/handlers/balance"
	"github.com/GlebRadaev/gofermart/internal/handlers/orders"

	pkgauth "github.com/GlebRadaev/gofermart/pkg/auth"

	"github.com/GlebRadaev/gofermart/internal/repo"
	authservice "github.com/GlebRadaev/gofermart/internal/service/authservice"
	balanceservice "github.com/GlebRadaev/gofermart/internal/service/balanceservice"
	orderservice "github.com/GlebRadaev/gofermart/internal/service/orderservice"
)

type Services struct {
	AuthService    auth.Service
	OrderService   orders.Service
	BalanceService balance.Service
}

func New(repo *repo.Repositories) *Services {
	balanceService := balanceservice.New(repo.BalanceRepo, repo.Withdrawal)
	orderService := orderservice.New(repo.OrderRepo)
	authService := authservice.New(repo.UserRepo, balanceService, &pkgauth.HashService{}, &pkgauth.JWTService{})

	return &Services{
		AuthService:    authService,
		OrderService:   orderService,
		BalanceService: balanceService,
	}
}
