package orderservice

import (
	"context"
	"errors"
	"testing"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func NewMock(t *testing.T) (*Service, *MockRepo) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepo(ctrl)
	service := New(repo)
	defer ctrl.Finish()
	return service, repo
}

func TestProcessOrder(t *testing.T) {
	service, repo := NewMock(t)
	tests := []struct {
		name          string
		userID        int
		orderNumber   string
		prepareMock   func()
		expectedOrder *domain.Order
		expectedError error
	}{
		{
			name:        "Order already exists by the same user",
			userID:      1,
			orderNumber: "12345",
			prepareMock: func() {
				repo.EXPECT().FindByOrderNumber(gomock.Any(), "12345").Return(&domain.Order{UserID: 1}, nil)
			},
			expectedOrder: nil,
			expectedError: ErrOrderAlreadyExistsByUser,
		},
		{
			name:        "Order already exists by another user",
			userID:      2,
			orderNumber: "12345",
			prepareMock: func() {
				repo.EXPECT().FindByOrderNumber(gomock.Any(), "12345").Return(&domain.Order{UserID: 1}, nil)
			},
			expectedOrder: nil,
			expectedError: ErrOrderAlreadyExists,
		},
		{
			name:        "New order is created successfully",
			userID:      1,
			orderNumber: "12345",
			prepareMock: func() {
				repo.EXPECT().FindByOrderNumber(gomock.Any(), "12345").Return(nil, nil)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedOrder: &domain.Order{
				UserID:      1,
				OrderNumber: "12345",
				Status:      NewOrderStatus,
			},
			expectedError: nil,
		},
		{
			name:        "Cannot find order by order number",
			userID:      1,
			orderNumber: "12345",
			prepareMock: func() {
				repo.EXPECT().FindByOrderNumber(gomock.Any(), "12345").Return(nil, errors.New("some error"))
			},
			expectedOrder: nil,
			expectedError: errors.New("some error"),
		},
		{
			name:        "Cannot save new order",
			userID:      1,
			orderNumber: "12345",
			prepareMock: func() {
				repo.EXPECT().FindByOrderNumber(gomock.Any(), "12345").Return(nil, nil)
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("some error"))
			},
			expectedOrder: nil,
			expectedError: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepareMock != nil {
				tt.prepareMock()
			}

			order, err := service.ProcessOrder(context.Background(), tt.userID, tt.orderNumber)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				assert.Equal(t, tt.expectedOrder.UserID, order.UserID)
				assert.Equal(t, tt.expectedOrder.OrderNumber, order.OrderNumber)
				assert.Equal(t, tt.expectedOrder.Status, order.Status)
			}
		})
	}
}

func TestGetOrders(t *testing.T) {
	service, repo := NewMock(t)

	tests := []struct {
		name           string
		userID         int
		prepareMock    func()
		expectedOrders []domain.Order
		expectedError  error
	}{
		{
			name:   "No orders found",
			userID: 1,
			prepareMock: func() {
				repo.EXPECT().FindOrdersByUserID(gomock.Any(), 1).Return(nil, nil)
			},
			expectedOrders: nil,
			expectedError:  nil,
		},
		{
			name:   "Filter out registered orders",
			userID: 1,
			prepareMock: func() {
				repo.EXPECT().FindOrdersByUserID(gomock.Any(), 1).Return([]domain.Order{
					{OrderNumber: "123", Status: RegisteredOrderStatus},
					{OrderNumber: "124", Status: NewOrderStatus},
				}, nil)
			},
			expectedOrders: []domain.Order{
				{OrderNumber: "124", Status: NewOrderStatus},
			},
			expectedError: nil,
		},
		{
			name:   "Error fetching orders",
			userID: 1,
			prepareMock: func() {
				repo.EXPECT().FindOrdersByUserID(gomock.Any(), 1).Return(nil, errors.New("db error"))
			},
			expectedOrders: nil,
			expectedError:  errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prepareMock != nil {
				tt.prepareMock()
			}

			orders, err := service.GetOrders(context.Background(), tt.userID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOrders, orders)
			}
		})
	}
}
