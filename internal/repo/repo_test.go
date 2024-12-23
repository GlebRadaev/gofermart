package repo

import (
	"testing"

	"github.com/GlebRadaev/gofermart/internal/pg"
	balancerepo "github.com/GlebRadaev/gofermart/internal/repo/balance-repo"
	orderrepo "github.com/GlebRadaev/gofermart/internal/repo/order-repo"
	userrepo "github.com/GlebRadaev/gofermart/internal/repo/user-repo"
	withdrawalrepo "github.com/GlebRadaev/gofermart/internal/repo/withdrawal-repo"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func NewMock(t *testing.T) (*Repositories, pgxmock.PgxPoolIface) {
	ctrl := gomock.NewController(t)
	mockDB, err := pgxmock.NewPool()
	mockTxManager := pg.NewMockTXManager(ctrl)
	assert.NoError(t, err)
	repo := New(mockDB, mockTxManager)
	defer mockDB.Close()

	return repo, mockDB
}

func TestNew(t *testing.T) {
	repo, mock := NewMock(t)

	assert.NotNil(t, repo.UserRepo)
	assert.NotNil(t, repo.OrderRepo)
	assert.NotNil(t, repo.BalanceRepo)
	assert.NotNil(t, repo.Withdrawal)

	assert.IsType(t, &userrepo.Repository{}, repo.UserRepo)
	assert.IsType(t, &orderrepo.Repository{}, repo.OrderRepo)
	assert.IsType(t, &balancerepo.Repository{}, repo.BalanceRepo)
	assert.IsType(t, &withdrawalrepo.Repository{}, repo.Withdrawal)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}
