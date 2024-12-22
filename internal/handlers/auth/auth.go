package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/dto"
	"github.com/GlebRadaev/gofermart/pkg/utils"
)

type Service interface {
	Register(ctx context.Context, login, password string) (*domain.User, error)
	Authenticate(ctx context.Context, login, password string) (*domain.User, error)
	GenerateToken(userID int) (string, error)
}

type AuthHandler struct {
	authService Service
}

func New(authService Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Create a new user account with login and password
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequestDTO	true	"Register request body"
//	@Success		200		{object}	dto.RegisterResponseDTO
//	@Failure		400		{object}	utils.Response	"Invalid request body"
//	@Failure		409		{object}	utils.Response	"User already exists"
//	@Failure		500		{object}	utils.Response	"Internal server error"
//	@Router			/api/user/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequestDTO
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	user, err := h.authService.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusConflict, err.Error())
		return
	}
	token, err := h.authService.GenerateToken(user.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error generating token")
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
	utils.RespondWithJSON(w, http.StatusOK, dto.RegisterResponseDTO{
		Message: "User successfully registered",
	})
}

// Login godoc
//
//	@Summary		Authenticate user
//	@Description	Log in with a user account and get a JWT token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequestDTO	true	"Login request body"
//	@Success		200		{object}	dto.LoginResponseDTO
//	@Failure		400		{object}	utils.Response	"Invalid request body"
//	@Failure		401		{object}	utils.Response	"Invalid credentials"
//	@Failure		500		{object}	utils.Response	"Internal server error"
//	@Router			/api/user/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequestDTO
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	user, err := h.authService.Authenticate(r.Context(), req.Login, req.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	token, err := h.authService.GenerateToken(user.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error generating token")
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
	utils.RespondWithJSON(w, http.StatusOK, dto.LoginResponseDTO{
		Message: "User successfully authenticated",
	})
}
