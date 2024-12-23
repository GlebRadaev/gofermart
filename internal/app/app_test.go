package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ApplicationSuite struct {
	suite.Suite
	app       *Application
	testError error
}

func TestApplication(t *testing.T) {
	suite.Run(t, &ApplicationSuite{})
}

func (s *ApplicationSuite) SetupTest() {
	s.app = New()
}

// func (s *ApplicationSuite) TestStart_Success() {
// 	cfg := &config.Config{
// 		Address:  "localhost:8080",
// 		Database: "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable",
// 	}
// 	s.app.cfg = cfg

// 	conn := repo.New(nil, nil)
// 	s.app.repo = conn
// 	s.app.srv = service.New(s.app.repo)
// 	s.app.api = handlers.New(s.app.srv)

// 	ctx := context.Background()
// 	err := s.app.Start(ctx)
// 	if strings.Contains(err.Error(), "can't build pgx pool") {
// 		return
// 	}
// 	s.NoError(err)
// }

func (s *ApplicationSuite) TestWait() {
	ctx, cancel := context.WithCancel(context.Background())

	s.app.errCh = make(chan error)
	go func() {
		s.app.errCh <- fmt.Errorf("mock error")
	}()

	err := s.app.Wait(ctx, cancel)

	s.Require().Error(err)
	s.Contains(err.Error(), "mock error")
}
