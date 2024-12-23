package authservice

import (
	"context"
	"errors"
	"time"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/handlers/balance"
	"github.com/GlebRadaev/gofermart/pkg/auth"
	"go.uber.org/zap"
)

type Repo interface {
	FindByLogin(ctx context.Context, login string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
}
type Service struct {
	userRepo       Repo
	balanceService balance.Service
	hashService    auth.HashServiceInterface
	jwtService     auth.JWTServiceInterface
}

func New(repo Repo, balanceService balance.Service, hashService auth.HashServiceInterface, jwtService auth.JWTServiceInterface) *Service {
	return &Service{
		userRepo:       repo,
		balanceService: balanceService,
		hashService:    hashService,
		jwtService:     jwtService,
	}
}

func (s *Service) Register(ctx context.Context, login, password string) (*domain.User, error) {
	existingUser, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil {
		zap.L().Error("can't find user: ", zap.Error(err))
		return nil, err
	}
	if existingUser != nil {
		zap.L().Info("user already exists, login: ", zap.String("login", login))
		return nil, errors.New("username already taken")
	}
	hashedPassword, err := s.hashService.HashPassword(password)
	if err != nil {
		zap.L().Error("can't hash password: ", zap.Error(err))
		return nil, err
	}
	user := &domain.User{
		Login:        login,
		PasswordHash: hashedPassword,
	}
	newUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		zap.L().Error("can't create user: ", zap.Error(err))
		return nil, err
	}

	_, err = s.balanceService.CreateBalance(ctx, newUser.ID)
	if err != nil {
		zap.L().Error("can't create balance: ", zap.Error(err))
		return nil, err
	}

	zap.L().Info("user successfully registered", zap.String("login", login))
	return user, nil
}

func (s *Service) Authenticate(ctx context.Context, login, password string) (*domain.User, error) {
	user, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil || user == nil {
		zap.L().Error("invalid credentials", zap.Error(err))
		return nil, errors.New("invalid credentials")
	}
	if ok := s.hashService.ComparePassword(user.PasswordHash, password); !ok {
		zap.L().Error("invalid credentials", zap.Error(err))
		return nil, errors.New("invalid credentials")
	}
	zap.L().Info("user successfully authenticated", zap.String("login", login))
	return user, nil
}

func (s *Service) GenerateToken(userID int) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)

	token, err := s.jwtService.GenerateJWT(userID, expirationTime)
	if err != nil {
		zap.L().Error("can't generate token: ", zap.Error(err))
		return "", err
	}
	return token, nil
}
