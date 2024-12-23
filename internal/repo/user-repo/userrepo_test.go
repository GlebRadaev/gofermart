package userrepo

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func NewMock(t *testing.T) (*Repository, pgxmock.PgxPoolIface) {
	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	repo := New(mockDB)
	defer mockDB.Close()

	return repo, mockDB
}

func TestRepository_FindByLogin(t *testing.T) {
	repo, mock := NewMock(t)

	tests := []struct {
		name      string
		login     string
		mockSetup func()
		expectErr bool
		result    *domain.User
	}{
		{
			name:  "User found",
			login: "test_user",
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "login", "password_hash"}).
					AddRow(1, "test_user", "hashed_password")
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, login, password_hash FROM users WHERE login = $1")).
					WithArgs("test_user").
					WillReturnRows(rows)
			},
			expectErr: false,
			result: &domain.User{
				ID:           1,
				Login:        "test_user",
				PasswordHash: "hashed_password",
			},
		},
		{
			name:  "User not found",
			login: "non_existing_user",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, login, password_hash FROM users WHERE login = $1")).
					WithArgs("non_existing_user").
					WillReturnError(pgx.ErrNoRows)
			},
			expectErr: false,
			result:    nil,
		},
		{
			name:  "Database error",
			login: "test_user",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, login, password_hash FROM users WHERE login = $1")).
					WithArgs("test_user").
					WillReturnError(errors.New("database error"))
			},
			expectErr: true,
			result:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.FindByLogin(context.Background(), tt.login)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.result, result)
			}
		})
	}
}

func TestRepository_Create(t *testing.T) {
	repo, mock := NewMock(t)

	tests := []struct {
		name      string
		user      *domain.User
		mockSetup func()
		expectErr bool
		result    *domain.User
	}{
		{
			name: "Create user successfully",
			user: &domain.User{
				Login:        "new_user",
				PasswordHash: "hashed_password",
			},
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`
					INSERT INTO users (login, password_hash)
					VALUES ($1, $2)
					RETURNING id
				`)).
					WithArgs("new_user", "hashed_password").
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectErr: false,
			result: &domain.User{
				ID:           1,
				Login:        "new_user",
				PasswordHash: "hashed_password",
			},
		},
		{
			name: "Database error",
			user: &domain.User{
				Login:        "new_user",
				PasswordHash: "hashed_password",
			},
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`
					INSERT INTO users (login, password_hash)
					VALUES ($1, $2)
					RETURNING id
				`)).
					WithArgs("new_user", "hashed_password").
					WillReturnError(errors.New("database error"))
			},
			expectErr: true,
			result:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.Create(context.Background(), tt.user)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.result, result)
			}
		})
	}
}
