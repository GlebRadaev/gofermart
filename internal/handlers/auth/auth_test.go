package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/pkg/utils"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func NewMock(t *testing.T) (*AuthHandler, *MockService) {
	ctrl := gomock.NewController(t)
	service := NewMockService(ctrl)
	handler := New(service)
	defer ctrl.Finish()
	return handler, service
}

func TestRegisterHandler(t *testing.T) {
	handler, service := NewMock(t)

	tests := []struct {
		name          string
		body          string
		prepareMock   func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Successful registration",
			body: `{"login":"newuser","password":"password123"}`,
			prepareMock: func() {
				service.EXPECT().Register(context.Background(), "newuser", "password123").Return(&domain.User{
					ID:           1,
					Login:        "newuser",
					PasswordHash: "hashedpassword",
				}, nil)
				service.EXPECT().GenerateToken(1).Return("some-jwt-token", nil)
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
		},
		{
			name: "User already exists",
			body: `{"login":"existinguser","password":"password123"}`,
			prepareMock: func() {
				service.EXPECT().Register(context.Background(), "existinguser", "password123").Return(nil, errors.New("user already exists"))
			},
			expectedCode:  http.StatusConflict,
			expectedError: "user already exists",
		},
		{
			name: "Invalid request body",
			body: `{invalid json`,
			prepareMock: func() {
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name: "Error generating token",
			body: `{"login":"newuser","password":"password123"}`,
			prepareMock: func() {
				service.EXPECT().Register(context.Background(), "newuser", "password123").Return(&domain.User{
					ID:           1,
					Login:        "newuser",
					PasswordHash: "hashedpassword",
				}, nil)
				service.EXPECT().
					GenerateToken(1).
					Return("", errors.New("token generation error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Error generating token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()

			req := httptest.NewRequest("POST", "/api/user/register", bytes.NewReader([]byte(tt.body)))
			rr := httptest.NewRecorder()

			handler.Register(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedError != "" {
				var resp utils.Response
				err := json.NewDecoder(rr.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, resp.Message)
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	handler, service := NewMock(t)

	tests := []struct {
		name          string
		body          string
		prepareMock   func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Successful login",
			body: `{"login":"testuser","password":"password123"}`,
			prepareMock: func() {
				service.EXPECT().
					Authenticate(context.Background(), "testuser", "password123").
					Return(&domain.User{
						ID:           1,
						Login:        "testuser",
						PasswordHash: "hashedpassword",
					}, nil)

				service.EXPECT().
					GenerateToken(1).
					Return("some-jwt-token", nil)
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
		},
		{
			name: "Invalid credentials",
			body: `{"login":"testuser","password":"wrongpassword"}`,
			prepareMock: func() {
				service.EXPECT().
					Authenticate(context.Background(), "testuser", "wrongpassword").
					Return(nil, errors.New("Invalid credentials"))
			},
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Invalid credentials",
		},
		{
			name: "Invalid request body",
			body: `{invalid json`,
			prepareMock: func() {
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid request body",
		},
		{
			name: "Error generating token",
			body: `{"login":"testuser","password":"password123"}`,
			prepareMock: func() {
				service.EXPECT().
					Authenticate(context.Background(), "testuser", "password123").
					Return(&domain.User{
						ID:           1,
						Login:        "testuser",
						PasswordHash: "hashedpassword",
					}, nil)

				service.EXPECT().
					GenerateToken(1).
					Return("", errors.New("token generation error"))
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Error generating token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepareMock()

			req := httptest.NewRequest("POST", "/api/user/login", bytes.NewReader([]byte(tt.body)))
			rr := httptest.NewRecorder()

			handler.Login(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedError != "" {
				var resp utils.Response
				err := json.NewDecoder(rr.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, resp.Message)
			}
		})
	}
}
