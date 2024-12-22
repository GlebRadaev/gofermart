package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	accrual "github.com/GlebRadaev/gofermart/internal/accrual"
	"github.com/GlebRadaev/gofermart/internal/config"
	"github.com/GlebRadaev/gofermart/internal/handlers"
	"github.com/GlebRadaev/gofermart/internal/pg"
	"github.com/GlebRadaev/gofermart/internal/repo"
	"github.com/GlebRadaev/gofermart/internal/service"
	"github.com/GlebRadaev/gofermart/pkg/clients"
	"github.com/GlebRadaev/gofermart/pkg/logger"
)

type ApplicationI interface {
	Start(ctx context.Context) error
	Wait(ctx context.Context, cancel context.CancelFunc) error
}

type Application struct {
	cfg  *config.Config
	api  *handlers.Handlers
	srv  *service.Services
	repo *repo.Repositories
	ext  *accrual.Service

	errCh chan error
	wg    sync.WaitGroup
	ready bool
}

func New() *Application {
	return &Application{
		errCh: make(chan error),
	}
}

func (a *Application) Start(ctx context.Context) error {
	cfg := config.New()

	err := logger.InitLogger(cfg)
	if err != nil {
		return fmt.Errorf("can't init logger: %w", err)
	}

	pool, err := getPgxpool(ctx, cfg)
	if err != nil {
		zap.L().Error("build pgx pool failed: ", zap.Error(err))
		return fmt.Errorf("can't build pgx pool: %w", err)
	}
	if err := pg.RunMigrations(pool); err != nil {
		zap.L().Error("migrations failed: ", zap.Error(err))
		return fmt.Errorf("can't run migrations: %w", err)
	}
	txManager := pg.NewTXManager(pool)

	conn := pg.New(pool)
	a.cfg = cfg
	a.repo = repo.New(conn, txManager)
	a.srv = service.New(a.repo)
	a.api = handlers.New(a.srv)
	a.ext = accrual.New(cfg, a.repo.OrderRepo, a.repo.BalanceRepo, clients.NewHTTPClient())

	if err = a.startHTTPServer(ctx); err != nil {
		return fmt.Errorf("can't start http server: %w", err)
	}

	a.startAccrualServer(ctx)

	a.ready = true
	zap.L().Info("all systems started successfully")
	return nil
}

func getPgxpool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	cfgpool, err := pgxpool.ParseConfig(cfg.Database)
	if err != nil {
		return nil, err
	}
	dbpool, err := pgxpool.NewWithConfig(ctx, cfgpool)
	if err != nil {
		return nil, err
	}
	if err = dbpool.Ping(ctx); err != nil {
		return nil, err
	}
	return dbpool, nil
}

func (a *Application) startHTTPServer(ctx context.Context) error {
	router := chi.NewRouter()
	a.api.InitRoutes(router)
	server := http.Server{
		Addr:    a.cfg.Address,
		Handler: router,
	}
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		<-ctx.Done()

		sCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(sCtx)
	}()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		zap.L().Info("starting http server on port", zap.String("port", a.cfg.Address))
		if err := server.ListenAndServe(); err != nil {
			a.errCh <- fmt.Errorf("http server exited with error: %w", err)
		}
	}()

	return nil
}

func (a *Application) startAccrualServer(ctx context.Context) error {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.ext.Start(ctx)
	}()

	return nil
}

func (a *Application) Wait(ctx context.Context, cancel context.CancelFunc) error {
	var appErr error

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		for err := range a.errCh {
			cancel()
			zap.L().Error(err.Error())
			appErr = err
		}
	}()

	<-ctx.Done()
	a.wg.Wait()
	close(a.errCh)
	wg.Wait()

	return appErr
}
