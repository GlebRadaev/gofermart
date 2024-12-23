package withdrawalrepo

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/GlebRadaev/gofermart/internal/domain"
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

func TestRepository_CreateWithdrawal(t *testing.T) {
	ctx := context.Background()
	repo, mock := NewMock(t)
	time := time.Now()
	tests := []struct {
		name       string
		withdrawal *domain.Withdrawal
		mockSetup  func()
		expectErr  bool
		result     *domain.Withdrawal
	}{
		{
			name: "Create withdrawal successfully",
			withdrawal: &domain.Withdrawal{
				UserID:      1,
				OrderNumber: "12345",
				Sum:         100.0,
				ProcessedAt: time,
			},
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`
					INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
					VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs(1, "12345", 100.0, pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))
			},
			expectErr: false,
			result: &domain.Withdrawal{
				ID:          1,
				UserID:      1,
				OrderNumber: "12345",
				Sum:         100.0,
				ProcessedAt: time,
			},
		},
		{
			name: "Database error",
			withdrawal: &domain.Withdrawal{
				UserID:      1,
				OrderNumber: "12345",
				Sum:         100.0,
				ProcessedAt: time,
			},
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`
					INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
					VALUES ($1, $2, $3, $4) RETURNING id`)).
					WithArgs(1, "12345", 100.0, pgxmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.CreateWithdrawal(ctx, tt.withdrawal)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.result, result)
			}
		})
	}
}

func TestRepository_GetWithdrawalsByUserID(t *testing.T) {
	ctx := context.Background()
	repo, mock := NewMock(t)
	time := time.Now()

	tests := []struct {
		name      string
		userID    int
		mockSetup func()
		expectErr bool
		result    []domain.Withdrawal
	}{
		{
			name:   "Withdrawals found",
			userID: 1,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "user_id", "order_number", "sum", "processed_at"}).
					AddRow(1, 1, "12345", 100.0, time).
					AddRow(2, 1, "67890", 200.0, time)
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id, user_id, order_number, sum, processed_at
					FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`)).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectErr: false,
			result: []domain.Withdrawal{
				{ID: 1, UserID: 1, OrderNumber: "12345", Sum: 100.0, ProcessedAt: time},
				{ID: 2, UserID: 1, OrderNumber: "67890", Sum: 200.0, ProcessedAt: time},
			},
		},
		{
			name:   "No withdrawals found",
			userID: 1,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "user_id", "order_number", "sum", "processed_at"})
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id, user_id, order_number, sum, processed_at
					FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`)).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectErr: false,
			result:    nil,
		},
		{
			name:   "Database error",
			userID: 1,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id, user_id, order_number, sum, processed_at
					FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`)).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expectErr: true,
			result:    nil,
		},
		{
			name:   "Error scanning row",
			userID: 1,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "user_id", "order_number", "sum", "processed_at"}).
					AddRow(1, 1, "12345", "invalid_data", "invalid_data")
				mock.ExpectQuery(regexp.QuoteMeta(`
					SELECT id, user_id, order_number, sum, processed_at
					FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`)).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectErr: true,
			result:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.GetWithdrawalsByUserID(ctx, tt.userID)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.result, result)
			}
		})
	}
}
