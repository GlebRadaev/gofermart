package accrual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GlebRadaev/gofermart/internal/config"
	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/service/balanceservice"
	"github.com/GlebRadaev/gofermart/internal/service/orderservice"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/GlebRadaev/gofermart/pkg/clients"
)

const (
	maxRetries    = 3
	retryInterval = time.Second * 1
)

var processingOrders sync.Map

type Response struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type Service struct {
	url            string
	orderRepo      orderservice.Repo
	balanceRepo    balanceservice.BalanceRepo
	client         clients.HTTPClientI
	limit          uint32
	workerPool     WorkerPoolI
	updateInterval time.Duration
}

func New(cfg *config.Config, orderRepo orderservice.Repo, balanceRepo balanceservice.BalanceRepo, client clients.HTTPClientI) *Service {
	return &Service{
		url:            cfg.AccrualAddress,
		orderRepo:      orderRepo,
		balanceRepo:    balanceRepo,
		client:         client,
		limit:          1000,
		workerPool:     NewWorkerPool(10),
		updateInterval: time.Second * 5,
	}
}

func (s *Service) Start(ctx context.Context) {
	zap.L().Info("Accrual service started")
	go s.run(ctx)
}

func (s *Service) run(ctx context.Context) {
	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			zap.L().Info("Context canceled, stopping service")
			return
		case <-ticker.C:
			s.processOrders(ctx)
		}
	}
}

func (s *Service) processOrders(ctx context.Context) {
	orders, err := s.orderRepo.FindForProcessing(ctx, atomic.LoadUint32(&s.limit))
	if err != nil {
		zap.L().Error("Failed to fetch orders for processing", zap.Error(err))
		return
	}

	var g errgroup.Group
	for _, order := range orders {
		order := order

		if _, loaded := processingOrders.LoadOrStore(order.OrderNumber, struct{}{}); loaded {
			continue
		}

		g.Go(func() error {
			err := s.workerPool.AddTask(ctx, func() error {
				defer processingOrders.Delete(order.OrderNumber)
				return s.handleOrder(ctx, order)
			})
			if err != nil {
				processingOrders.Delete(order.OrderNumber)
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		zap.L().Error("Error processing orders", zap.Error(err))
	}
}
func (s *Service) handleOrder(ctx context.Context, order domain.Order) error {
	url := s.url + "/api/orders/" + order.OrderNumber
	var err error
	var statusCode int
	var respBody []byte
	var respHeaders http.Header

	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			statusCode, respBody, respHeaders, err = s.client.Get(url, nil)
			if err != nil {
				if attempt < maxRetries {
					retryAfter := retryInterval * time.Duration(attempt)
					time.Sleep(retryAfter)
					continue
				}
				return fmt.Errorf("failed to process order %s after %d retries: %w", order.OrderNumber, maxRetries, err)
			}

			switch statusCode {
			case http.StatusTooManyRequests:
				return s.handleRateLimit(order, respHeaders, attempt)
			case http.StatusNoContent:
				zap.L().Warn("Order not found in accrual system, retrying", zap.String("orderNumber", order.OrderNumber), zap.Int("attempt", attempt))
				if attempt < maxRetries {
					retryAfter := retryInterval * time.Duration(attempt)
					time.Sleep(retryAfter)
					continue
				}
				return fmt.Errorf("failed to process not found order %s after %d retries", order.OrderNumber, maxRetries)

			case http.StatusOK:
				return s.processAccrual(ctx, order, respBody)

			default:
				zap.L().Error("Unexpected status code", zap.Int("status", statusCode), zap.String("orderNumber", order.OrderNumber))
				return errors.New("unexpected status code")
			}
		}
	}
	return nil
}

func (s *Service) processAccrual(ctx context.Context, order domain.Order, respBody []byte) error {
	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("failed to parse response body: %w", err)
	}

	if response.Order != order.OrderNumber {
		return fmt.Errorf("order number mismatch: expected %s, got %s", order.OrderNumber, response.Order)
	}

	order.Status = response.Status
	switch response.Status {
	case "PROCESSED":
		if response.Accrual > 0 {
			order.Accrual = response.Accrual
			if err := s.UpdateBalance(ctx, order.UserID, response.Accrual); err != nil {
				return fmt.Errorf("failed to update balance for user %d: %w", order.UserID, err)
			}
		}
	case "REGISTERED":
		zap.L().Info("Order registered, no accrual yet", zap.String("orderNumber", order.OrderNumber))
	case "INVALID":
		zap.L().Info("Order is invalid and will not be processed", zap.String("orderNumber", order.OrderNumber))
	default:
		zap.L().Warn("Unrecognized status received", zap.String("orderNumber", order.OrderNumber), zap.String("status", response.Status))
	}

	if err := s.orderRepo.Update(ctx, &order); err != nil {
		return fmt.Errorf("failed to update order in repo: %w", err)
	}
	return nil
}

func (s *Service) UpdateBalance(ctx context.Context, userID int, accrual float64) error {
	balance, err := s.balanceRepo.GetUserBalance(ctx, userID)
	if err != nil {
		return err
	}

	balance.CurrentBalance += accrual
	_, err = s.balanceRepo.UpdateUserBalance(ctx, userID, balance)
	if err != nil {
		return err
	}

	zap.L().Info("Balance updated successfully", zap.Int("userID", userID), zap.Float64("accrual", accrual))
	return nil
}

func (s *Service) handleRateLimit(order domain.Order, respHeaders http.Header, attempt int) error {
	retryAfterHeader := respHeaders.Get("Retry-After")
	retryAfter := retryInterval * time.Duration(attempt)

	if retryAfterHeader != "" {
		if seconds, err := strconv.Atoi(retryAfterHeader); err == nil {
			retryAfter = time.Duration(seconds) * time.Second
		}
	}
	zap.L().Warn(
		"Rate limit detected, retrying",
		zap.String("orderNumber", order.OrderNumber),
		zap.Int("attempt", attempt),
		zap.Duration("retryAfter", retryAfter),
	)
	time.Sleep(retryAfter)
	return nil
}
