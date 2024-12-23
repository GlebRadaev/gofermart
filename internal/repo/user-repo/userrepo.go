package userrepo

import (
	"context"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/pg"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Repository struct {
	db pg.Database
}

func New(db pg.Database) *Repository {
	return &Repository{
		db: db,
	}
}

func (repo *Repository) FindByLogin(ctx context.Context, login string) (*domain.User, error) {
	var user domain.User
	err := repo.db.QueryRow(ctx, "SELECT id, login, password_hash FROM users WHERE login = $1", login).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		zap.L().Error("can't find user", zap.Error(err))
		return nil, err
	}
	return &user, nil
}

func (repo *Repository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`
	err := repo.db.QueryRow(ctx, query, user.Login, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		zap.L().Error("can't save user", zap.Error(err))
		return nil, err
	}
	return user, nil
}
