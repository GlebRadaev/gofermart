package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type key int

const (
	keyWithTX key = iota
)

func TXFrom(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(keyWithTX).(pgx.Tx)
	return tx, ok
}

type ContextOpt func(ctx context.Context) context.Context

func With(ctx context.Context, opts ...ContextOpt) context.Context {
	for _, o := range opts {
		ctx = o(ctx)
	}
	return ctx
}

func WithTransaction(tx pgx.Tx) ContextOpt {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, keyWithTX, tx)
	}
}
