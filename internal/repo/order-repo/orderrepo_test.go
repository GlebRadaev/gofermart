package orderrepo

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

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

func TestRepository_FindByOrderNumber(t *testing.T) {
	repo, mock, _ := NewMock(t)
	time := time.Now()
	tests := []struct {
		name        string
		orderNumber string
		mockSetup   func()
		expectErr   bool
		result      *domain.Order
	}{
		{
			name:        "Order exists",
			orderNumber: "12345",
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "user_id", "order_number", "status", "accrual", "uploaded_at"}).
					AddRow(1, 1, "12345", "NEW", 100.0, time)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM orders WHERE order_number = $1")).
					WithArgs("12345").
					WillReturnRows(rows)
			},
			expectErr: false,
			result: &domain.Order{
				ID:          1,
				UserID:      1,
				OrderNumber: "12345",
				Status:      "NEW",
				Accrual:     100.0,
				UploadedAt:  time,
			},
		},
		{
			name:        "Order does not exist",
			orderNumber: "99999",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM orders WHERE order_number = $1")).
					WithArgs("99999").
					WillReturnError(pgx.ErrNoRows)
			},
			expectErr: false,
			result:    nil,
		},
		{
			name:        "Database error",
			orderNumber: "12345",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM orders WHERE order_number = $1")).
					WithArgs("12345").
					WillReturnError(errors.New("database error"))
			},
			expectErr: true,
			result:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.FindByOrderNumber(context.Background(), tt.orderNumber)
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

func TestRepository_FindOrdersByUserID(t *testing.T) {
	repo, mock, _ := NewMock(t)
	timeNow := time.Now()

	tests := []struct {
		name      string
		userID    int
		mockSetup func()
		expectErr bool
		result    []domain.Order
	}{
		{
			name:   "Orders found",
			userID: 1,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "user_id", "order_number", "status", "accrual", "uploaded_at"}).
					AddRow(1, 1, "12345", "NEW", 100.0, timeNow).
					AddRow(2, 1, "67890", "PROCESSING", 200.0, timeNow)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC")).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectErr: false,
			result: []domain.Order{
				{ID: 1, UserID: 1, OrderNumber: "12345", Status: "NEW", Accrual: 100.0, UploadedAt: timeNow},
				{ID: 2, UserID: 1, OrderNumber: "67890", Status: "PROCESSING", Accrual: 200.0, UploadedAt: timeNow},
			},
		},
		{
			name:   "Database error",
			userID: 1,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC")).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			expectErr: true,
			result:    nil,
		},
		{
			name:   "Scan row error",
			userID: 1,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "user_id", "order_number", "status", "accrual", "uploaded_at"}).
					AddRow(1, 1, "12345", "NEW", "invalid_value", timeNow)
				mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC")).
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
			result, err := repo.FindOrdersByUserID(context.Background(), tt.userID)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.result, result)
			}
		})
	}
}

func TestRepository_Save(t *testing.T) {
	repo, mock, tx := NewMock(t)
	timeNow := time.Now()

	tests := []struct {
		name      string
		order     *domain.Order
		mockSetup func()
		expectErr bool
	}{
		{
			name: "Save order successfully",
			order: &domain.Order{
				UserID:      1,
				OrderNumber: "12345",
				Status:      "NEW",
				Accrual:     100.0,
				UploadedAt:  timeNow,
			},
			mockSetup: func() {
				tx.EXPECT().Begin(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn pg.TransactionalFn) error {
					mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders (user_id, order_number, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5)`)).
						WithArgs(1, "12345", "NEW", 100.0, timeNow).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
					return fn(ctx)
				})
			},
			expectErr: false,
		},
		{
			name: "Database error",
			order: &domain.Order{
				UserID:      1,
				OrderNumber: "12345",
				Status:      "NEW",
				Accrual:     100.0,
				UploadedAt:  timeNow,
			},
			mockSetup: func() {
				tx.EXPECT().Begin(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn pg.TransactionalFn) error {
					mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders (user_id, order_number, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5)`)).
						WithArgs(1, "12345", "NEW", 100.0, timeNow).
						WillReturnError(errors.New("dtabase error"))

					return fn(ctx)
				})
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := repo.Save(context.Background(), tt.order)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepository_Update(t *testing.T) {
	repo, mock, tx := NewMock(t)

	tests := []struct {
		name      string
		order     *domain.Order
		mockSetup func()
		expectErr bool
		result    *domain.Order
	}{
		{
			name: "Update order successfully",
			order: &domain.Order{
				ID:      1,
				Status:  "PROCESSED",
				Accrual: 150.0,
			},
			mockSetup: func() {

				tx.EXPECT().Begin(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn pg.TransactionalFn) error {
					mock.ExpectExec(regexp.QuoteMeta(`UPDATE orders SET status = $1, accrual = $2 WHERE id = $3`)).
						WithArgs("PROCESSED", 150.0, 1).
						WillReturnResult(pgxmock.NewResult("UPDATE", 1))
					return fn(ctx)
				})
			},
			expectErr: false,
		},
		{
			name: "Database error",
			order: &domain.Order{
				UserID:      1,
				OrderNumber: "12345",
				Status:      "NEW",
				Accrual:     100.0,
			},
			mockSetup: func() {
				tx.EXPECT().Begin(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, fn pg.TransactionalFn) error {
					mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders (user_id, order_number, status, accrual, uploaded_at) VALUES ($1, $2, $3, $4, $5)`)).
						WithArgs(1, "12345", "NEW", 100.0).
						WillReturnError(errors.New("dtabase error"))

					return fn(ctx)
				})
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := repo.Update(context.Background(), tt.order)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepository_FindForProcessing(t *testing.T) {
	repo, mock, _ := NewMock(t)
	time := time.Now()

	tests := []struct {
		name      string
		limit     uint32
		mockSetup func()
		expectErr bool
		result    []domain.Order
	}{
		{
			name:  "Orders found",
			limit: 2,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "user_id", "order_number", "status", "accrual", "uploaded_at"}).
					AddRow(1, 1, "12345", "NEW", 100.0, time).
					AddRow(2, 1, "67890", "PROCESSING", 200.0, time)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM orders WHERE status = 'NEW' OR status = 'PROCESSING' ORDER BY uploaded_at ASC LIMIT $1`)).
					WithArgs(2).
					WillReturnRows(rows)
			},
			expectErr: false,
			result: []domain.Order{
				{ID: 1, UserID: 1, OrderNumber: "12345", Status: "NEW", Accrual: 100.0, UploadedAt: time},
				{ID: 2, UserID: 1, OrderNumber: "67890", Status: "PROCESSING", Accrual: 200.0, UploadedAt: time},
			},
		},
		{
			name:  "No orders found",
			limit: 2,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "user_id", "order_number", "status", "accrual", "uploaded_at"})
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM orders WHERE status = 'NEW' OR status = 'PROCESSING' ORDER BY uploaded_at ASC LIMIT $1`)).
					WithArgs(2).
					WillReturnRows(rows)
			},
			expectErr: false,
			result:    nil,
		},
		{
			name:  "Database error",
			limit: 2,
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM orders WHERE status = 'NEW' OR status = 'PROCESSING' ORDER BY uploaded_at ASC LIMIT $1`)).
					WithArgs(2).
					WillReturnError(errors.New("database error"))
			},
			expectErr: true,
			result:    nil,
		},
		{
			name:  "Error scanning row",
			limit: 2,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "user_id", "order_number", "status", "accrual", "uploaded_at"}).
					AddRow(1, 1, "12345", "NEW", 100.0, time).
					AddRow(2, 1, "67890", "PROCESSING", "invalid_data", time)
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM orders WHERE status = 'NEW' OR status = 'PROCESSING' ORDER BY uploaded_at ASC LIMIT $1`)).
					WithArgs(2).
					WillReturnRows(rows)
			},
			expectErr: true,
			result:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			result, err := repo.FindForProcessing(context.Background(), tt.limit)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.result, result)
			}
		})
	}
}
