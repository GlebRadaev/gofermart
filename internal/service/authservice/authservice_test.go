package authservice

import (
	"context"
	"errors"
	"testing"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/handlers/balance"
	"github.com/GlebRadaev/gofermart/pkg/auth"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func NewMock(t *testing.T) (*Service, *MockRepo, *balance.MockService, *auth.MockHashServiceInterface, *auth.MockJWTServiceInterface) {
	ctrl := gomock.NewController(t)
	repo := NewMockRepo(ctrl)
	balanceService := balance.NewMockService(ctrl)
	hashService := auth.NewMockHashServiceInterface(ctrl)
	jwtService := auth.NewMockJWTServiceInterface(ctrl)

	service := New(repo, balanceService, hashService, jwtService)
	defer ctrl.Finish()
	return service, repo, balanceService, hashService, jwtService
}

func TestRegister(t *testing.T) {
	service, userRepo, balanceService, passwordHasher, _ := NewMock(t)

	tests := []struct {
		name          string
		login         string
		password      string
		prepareMock   func()
		expectedUser  *domain.User
		expectedError error
	}{
		{
			name:     "Successful registration",
			login:    "testuser",
			password: "testpassword",
			prepareMock: func() {
				userRepo.EXPECT().FindByLogin(context.Background(), "testuser").Return(nil, nil)
				passwordHasher.EXPECT().HashPassword("testpassword").Return("hashedpassword", nil)
				userRepo.EXPECT().Create(context.Background(), gomock.Any()).DoAndReturn(func(ctx context.Context, user *domain.User) (*domain.User, error) {
					user.ID = 1
					return user, nil
				})
				balanceService.EXPECT().CreateBalance(context.Background(), gomock.Any()).Return(nil, nil)
			},
			expectedUser: &domain.User{
				ID:           1,
				Login:        "testuser",
				PasswordHash: "hashedpassword",
			},
			expectedError: nil,
		},
		{
			name:     "User already exists",
			login:    "testuser",
			password: "testpassword",
			prepareMock: func() {
				userRepo.EXPECT().FindByLogin(context.Background(), "testuser").Return(&domain.User{Login: "testuser"}, nil)
			},
			expectedUser:  nil,
			expectedError: errors.New("username already taken"),
		},
		{
			name:     "Error finding user",
			login:    "testuser",
			password: "testpassword",
			prepareMock: func() {
				userRepo.EXPECT().FindByLogin(context.Background(), "testuser").Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
		{
			name:     "Error hashing password",
			login:    "testuser",
			password: "testpassword",
			prepareMock: func() {
				userRepo.EXPECT().FindByLogin(context.Background(), "testuser").Return(nil, nil)
				passwordHasher.EXPECT().HashPassword("testpassword").Return("", errors.New("hashing error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("hashing error"),
		},
		{
			name:     "Error creating user",
			login:    "testuser",
			password: "testpassword",
			prepareMock: func() {
				userRepo.EXPECT().FindByLogin(context.Background(), "testuser").Return(nil, nil)
				passwordHasher.EXPECT().HashPassword("testpassword").Return("hashedpassword", nil)
				userRepo.EXPECT().Create(context.Background(), gomock.Any()).Return(nil, errors.New("creation failed"))
			},
			expectedUser:  nil,
			expectedError: errors.New("creation failed"),
		},
		{
			name:     "Error creating balance",
			login:    "testuser",
			password: "testpassword",
			prepareMock: func() {
				userRepo.EXPECT().FindByLogin(context.Background(), "testuser").Return(nil, nil)
				passwordHasher.EXPECT().HashPassword("testpassword").Return("hashedpassword", nil)
				userRepo.EXPECT().Create(context.Background(), gomock.Any()).DoAndReturn(func(ctx context.Context, user *domain.User) (*domain.User, error) {
					user.ID = 1
					return user, nil
				})
				balanceService.EXPECT().CreateBalance(context.Background(), gomock.Any()).Return(nil, errors.New("balance creation failed"))
			},
			expectedUser:  nil,
			expectedError: errors.New("balance creation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()

			user, err := service.Register(context.Background(), tt.login, tt.password)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	service, userRepo, _, passwordHasher, _ := NewMock(t)

	tests := []struct {
		name          string
		login         string
		password      string
		prepareMock   func()
		expectedUser  *domain.User
		expectedError error
	}{
		{
			name:     "Successful authentication",
			login:    "testuser",
			password: "testpassword",
			prepareMock: func() {
				userRepo.EXPECT().FindByLogin(context.Background(), "testuser").Return(&domain.User{
					ID:           1,
					Login:        "testuser",
					PasswordHash: "hashedpassword",
				}, nil)
				passwordHasher.EXPECT().ComparePassword("hashedpassword", "testpassword").Return(true)
			},
			expectedUser: &domain.User{
				ID:           1,
				Login:        "testuser",
				PasswordHash: "hashedpassword",
			},
			expectedError: nil,
		},
		{
			name:     "Invalid credentials - user not found",
			login:    "testuser",
			password: "testpassword",
			prepareMock: func() {
				userRepo.EXPECT().FindByLogin(context.Background(), "testuser").Return(nil, nil)
			},
			expectedUser:  nil,
			expectedError: errors.New("invalid credentials"),
		},
		{
			name:     "Invalid credentials - incorrect password",
			login:    "testuser",
			password: "wrongpassword",
			prepareMock: func() {
				userRepo.EXPECT().FindByLogin(context.Background(), "testuser").Return(&domain.User{
					ID:           1,
					Login:        "testuser",
					PasswordHash: "hashedpassword",
				}, nil)
				passwordHasher.EXPECT().ComparePassword("hashedpassword", "wrongpassword").Return(false)
			},
			expectedUser:  nil,
			expectedError: errors.New("invalid credentials"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()

			user, err := service.Authenticate(context.Background(), tt.login, tt.password)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, user)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	service, _, _, _, jwtService := NewMock(t)

	tests := []struct {
		name          string
		userID        int
		prepareMock   func()
		expectedToken string
		expectedError error
	}{
		{
			name:   "Successful token generation",
			userID: 1,
			prepareMock: func() {
				jwtService.EXPECT().GenerateJWT(1, gomock.Any()).Return("generated-token", nil)
			},
			expectedToken: "generated-token",
			expectedError: nil,
		},
		{
			name:   "Error generating token",
			userID: 1,
			prepareMock: func() {
				jwtService.EXPECT().GenerateJWT(1, gomock.Any()).Return("", errors.New("can't generate token"))
			},
			expectedToken: "",
			expectedError: errors.New("can't generate token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()

			token, err := service.GenerateToken(tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}
