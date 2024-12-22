package balancerepo

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/pg"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func NewMock(t *testing.T) (*Repository, pgxmock.PgxPoolIface, *pg.MockTXManager) {
	ctrl := gomock.NewController(t)
	mockTxManager := pg.NewMockTXManager(ctrl)

	mockDB, err := pgxmock.NewPool()
	assert.NoError(t, err)
	repo := New(mockDB, mockTxManager)
	defer mockDB.Close()
	defer ctrl.Finish()

	return repo, mockDB, mockTxManager
}

func TestRepository_GetUserBalance(t *testing.T) {
	repo, mock, _ := NewMock(t)

	tests := []struct {
		name      string
		userID    int
		mockSetup func()
		expectErr bool
		result    *domain.Balance
	}{
		{
			name:   "Valid userID returns balance",
			userID: 1,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "user_id", "current_balance", "withdrawn_total"}).
					AddRow(1, 1, 100.0, 50.0)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, current_balance, withdrawn_total FROM balances WHERE user_id = $1`)).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectErr: false,
			result: &domain.Balance{
				ID:             1,
				UserID:         1,
				CurrentBalance: 100.0,
				WithdrawnTotal: 50.0,
			},
		},
		{
			name:   "Non-existing userID returns nil",
			userID: 99,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, current_balance, withdrawn_total FROM balances WHERE user_id = $1`)).
					WithArgs(99).
					WillReturnError(pgx.ErrNoRows)
			},
			expectErr: false,
			result:    nil,
		},
		{
			name:   "Database error",
			userID: 1,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, current_balance, withdrawn_total FROM balances WHERE user_id = $1`)).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expectErr: true,
			result:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.GetUserBalance(context.Background(), tt.userID)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.result == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.result, result)
			}
		})
	}
}

func TestRepository_CreateUserBalance(t *testing.T) {
	repo, mock, _ := NewMock(t)

	tests := []struct {
		name      string
		userID    int
		mockSetup func()
		expectErr bool
		result    *domain.Balance
	}{
		{
			name:   "Successfully creates balance",
			userID: 1,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`
					INSERT INTO balances (user_id, current_balance, withdrawn_total)
					VALUES ($1, 0, 0)
					RETURNING id, user_id, current_balance, withdrawn_total`)).
					WithArgs(1).
					WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "current_balance", "withdrawn_total"}).
						AddRow(1, 1, 0.0, 0.0),
					)
			},
			expectErr: false,
			result: &domain.Balance{
				ID:             1,
				UserID:         1,
				CurrentBalance: 0.0,
				WithdrawnTotal: 0.0,
			},
		},
		{
			name:   "Database error",
			userID: 1,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`
					INSERT INTO balances (user_id, current_balance, withdrawn_total)
					VALUES ($1, 0, 0)
					RETURNING id, user_id, current_balance, withdrawn_total`)).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expectErr: true,
			result:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.CreateUserBalance(context.Background(), tt.userID)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.result, result)
			}
		})
	}
}

func TestRepository_UpdateUserBalance(t *testing.T) {
	repo, mock, tx := NewMock(t)

	tests := []struct {
		name         string
		userID       int
		inputBalance *domain.Balance
		mockSetup    func()
		expectErr    bool
		expected     *domain.Balance
	}{
		{
			name:   "Successfully updates balance",
			userID: 1,
			inputBalance: &domain.Balance{
				CurrentBalance: 200.0,
				WithdrawnTotal: 100.0,
			},
			mockSetup: func() {
				tx.EXPECT().Begin(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn pg.TransactionalFn) error {
					mock.ExpectQuery(regexp.QuoteMeta(`
					UPDATE balances
					SET current_balance = $1,
						withdrawn_total = $2
					WHERE user_id = $3
					RETURNING id, user_id, current_balance, withdrawn_total`)).
						WithArgs(200.0, 100.0, 1).
						WillReturnRows(
							pgxmock.NewRows([]string{"id", "user_id", "current_balance", "withdrawn_total"}).
								AddRow(1, 1, 200.0, 100.0),
						)
					return fn(ctx)
				})
			},
			expectErr: false,
			expected: &domain.Balance{
				ID:             1,
				UserID:         1,
				CurrentBalance: 200.0,
				WithdrawnTotal: 100.0,
			},
		},
		{
			name:   "Database error",
			userID: 1,
			inputBalance: &domain.Balance{
				CurrentBalance: 200.0,
				WithdrawnTotal: 100.0,
			},
			mockSetup: func() {
				tx.EXPECT().Begin(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn pg.TransactionalFn) error {
					mock.ExpectQuery(regexp.QuoteMeta(`
					UPDATE balances
					SET current_balance = $1,
						withdrawn_total = $2
					WHERE user_id = $3
					RETURNING id, user_id, current_balance, withdrawn_total`)).
						WithArgs(200.0, 100.0, 1).
						WillReturnError(errors.New("database error"))

					return fn(ctx)
				})
			},
			expectErr: true,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.UpdateUserBalance(context.Background(), tt.userID, tt.inputBalance)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
