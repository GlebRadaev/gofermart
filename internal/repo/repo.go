package repo

import (
	"github.com/GlebRadaev/gofermart/internal/pg"
	balancerepo "github.com/GlebRadaev/gofermart/internal/repo/balance-repo"
	orderrepo "github.com/GlebRadaev/gofermart/internal/repo/order-repo"
	userrepo "github.com/GlebRadaev/gofermart/internal/repo/user-repo"
	withdrawalrepo "github.com/GlebRadaev/gofermart/internal/repo/withdrawal-repo"
	"github.com/GlebRadaev/gofermart/internal/service/authservice"
	"github.com/GlebRadaev/gofermart/internal/service/balanceservice"
	"github.com/GlebRadaev/gofermart/internal/service/orderservice"
)

type Repositories struct {
	UserRepo    authservice.Repo
	OrderRepo   orderservice.Repo
	BalanceRepo balanceservice.BalanceRepo
	Withdrawal  balanceservice.WithdrawalRepo
}

func New(conn pg.Database, txManager pg.TXManager) *Repositories {
	userRepo := userrepo.New(conn)
	orderRepo := orderrepo.New(conn, txManager)
	balanceRepo := balancerepo.New(conn, txManager)
	withdrawalRepo := withdrawalrepo.New(conn)

	return &Repositories{
		UserRepo:    userRepo,
		OrderRepo:   orderRepo,
		BalanceRepo: balanceRepo,
		Withdrawal:  withdrawalRepo,
	}
}
