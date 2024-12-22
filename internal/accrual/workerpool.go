package accrual

import (
	"context"

	"go.uber.org/zap"
)

type WorkerPoolI interface {
	AddTask(ctx context.Context, task Task) error
	Close()
}

type Task func() error

type WorkerPool struct {
	pool chan Task
}

func NewWorkerPool(size int) *WorkerPool {
	pool := make(chan Task, size)
	wp := &WorkerPool{pool: pool}

	for i := 0; i < size; i++ {
		go wp.worker()
	}
	return wp
}

func (wp *WorkerPool) worker() {
	for task := range wp.pool {
		if err := task(); err != nil {
			zap.L().Error("Task execution failed", zap.Error(err))
		}
	}
}

func (wp *WorkerPool) AddTask(ctx context.Context, task Task) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case wp.pool <- task:
		return nil
	}
}

func (wp *WorkerPool) Close() {
	select {
	case <-wp.pool:
	default:
		close(wp.pool)
	}
}
