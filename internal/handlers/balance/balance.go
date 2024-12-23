package balance

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/dto"
	balanceservice "github.com/GlebRadaev/gofermart/internal/service/balanceservice"
	"github.com/GlebRadaev/gofermart/pkg/auth"
	"github.com/GlebRadaev/gofermart/pkg/utils"
	"github.com/GlebRadaev/gofermart/pkg/validate"
)

type Service interface {
	CreateBalance(ctx context.Context, userID int) (*domain.Balance, error)
	GetBalance(ctx context.Context, userID int) (*domain.Balance, error)
	Withdraw(ctx context.Context, userID int, orderNumber string, amount float64) error
	GetWithdrawals(ctx context.Context, userID int) ([]domain.Withdrawal, error)
}

type BalanceHandler struct {
	balanceService Service
}

func New(balanceService Service) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
	}
}

// GetBalance godoc
//
//	@Summary		Get current user balance
//	@Description	Retrieve the current loyalty points balance and the total amount withdrawn for the authenticated user.
//	@Tags			Баланс
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{object}	dto.BalanceResponseDTO	"Current balance and withdrawn points"
//	@Failure		401	{object}	utils.Response			"User not authorized"
//	@Failure		500	{object}	utils.Response			"Internal server error"
//	@Router			/api/user/balance [get]
func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, dto.BalanceResponseDTO{
		Current:   balance.CurrentBalance,
		Withdrawn: balance.WithdrawnTotal,
	})
}

// Withdraw godoc
//
//	@Summary		Request funds withdrawal
//	@Description	Withdraw points from the user balance for the provided order number.
//	@Tags			Баланс
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.BalanceWithdrawRequestDTO	true	"Withdrawal request payload"
//	@Success		200		{string}	string							"Withdrawal successful"
//	@Failure		401		{object}	utils.Response					"User not authorized"
//	@Failure		402		{object}	utils.Response					"Insufficient balance"
//	@Failure		422		{object}	utils.Response					"Invalid order number"
//	@Failure		500		{object}	utils.Response					"Internal server error"
//	@Router			/api/user/balance/withdraw [post]
func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)

	var req dto.BalanceWithdrawRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ok := validate.IsLuna(req.Order)
	if !ok {
		utils.RespondWithError(w, http.StatusUnprocessableEntity, "Invalid order number")
		return
	}

	err := h.balanceService.Withdraw(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		switch {
		case errors.Is(err, balanceservice.ErrInsufficientBalance):
			utils.RespondWithError(w, http.StatusPaymentRequired, err.Error())
		default:
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, "withdrawal successful")
}

// GetWithdrawals godoc
//
//	@Summary		Get withdrawals history
//	@Description	Get withdrawals history for the authenticated user with sorted by processed at date
//	@Tags			Баланс
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{array}		dto.GetWithdrawalsResponseDTO	"Withdrawals history"
//	@Success		204	{object}	utils.Response					"Withdrawals not found"
//	@Failure		401	{object}	utils.Response					"User not authorized"
//	@Failure		500	{object}	utils.Response					"Internal server error"
//	@Router			/api/user/withdrawals [get]
func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)

	withdrawals, err := h.balanceService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch withdrawals")
		return
	}

	if len(withdrawals) == 0 {
		utils.RespondWithError(w, http.StatusNoContent, "Withdrawals not found")
		return
	}

	response := make([]dto.GetWithdrawalsResponseDTO, len(withdrawals))
	for i, wd := range withdrawals {
		response[i] = dto.GetWithdrawalsResponseDTO{
			Order:       wd.OrderNumber,
			Sum:         wd.Sum,
			ProcessedAt: wd.ProcessedAt,
		}
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}
