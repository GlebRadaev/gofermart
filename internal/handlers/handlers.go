package handlers

import (
	"net/http"

	_ "github.com/GlebRadaev/gofermart/docs"
	authhandlers "github.com/GlebRadaev/gofermart/internal/handlers/auth"
	balancehandlers "github.com/GlebRadaev/gofermart/internal/handlers/balance"
	ordershandlers "github.com/GlebRadaev/gofermart/internal/handlers/orders"
	"github.com/GlebRadaev/gofermart/internal/service"
	"github.com/GlebRadaev/gofermart/pkg/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type AuthHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
}

type OrderHandler interface {
	AddOrder(w http.ResponseWriter, r *http.Request)
	GetOrders(w http.ResponseWriter, r *http.Request)
}

type BalanceHandler interface {
	GetBalance(w http.ResponseWriter, r *http.Request)
	Withdraw(w http.ResponseWriter, r *http.Request)
	GetWithdrawals(w http.ResponseWriter, r *http.Request)
}

type Handlers struct {
	AuthHandler    AuthHandler
	OrderHandler   OrderHandler
	BalanceHandler BalanceHandler
}

func New(s *service.Services) *Handlers {
	return &Handlers{
		AuthHandler:    authhandlers.New(s.AuthService),
		OrderHandler:   ordershandlers.New(s.OrderService),
		BalanceHandler: balancehandlers.New(s.BalanceService),
	}
}

func (h *Handlers) InitRoutes(r chi.Router) chi.Router {
	r.Use(
		middleware.RealIP,
		middleware.Recoverer,
		middleware.Logger,
	)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	))
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", h.AuthHandler.Register)
		r.Post("/login", h.AuthHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(auth.AuthMiddleware)
			r.Route("/orders", func(r chi.Router) {
				r.Post("/", h.OrderHandler.AddOrder)
				r.Get("/", h.OrderHandler.GetOrders)
			})
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", h.BalanceHandler.GetBalance)
				r.Post("/withdraw", h.BalanceHandler.Withdraw)
			})
			r.Get("/withdrawals", h.BalanceHandler.GetWithdrawals)
		})
	})

	return r
}
