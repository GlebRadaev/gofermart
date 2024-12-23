package orders

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/GlebRadaev/gofermart/internal/domain"
	"github.com/GlebRadaev/gofermart/internal/dto"

	orderservice "github.com/GlebRadaev/gofermart/internal/service/orderservice"
	"github.com/GlebRadaev/gofermart/pkg/auth"
	"github.com/GlebRadaev/gofermart/pkg/utils"
	"github.com/GlebRadaev/gofermart/pkg/validate"
)

type Service interface {
	ProcessOrder(ctx context.Context, userID int, orderNumber string) (*domain.Order, error)
	GetOrders(ctx context.Context, userID int) ([]domain.Order, error)
}

type OrderHandler struct {
	orderService Service
}

func New(orderService Service) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// AddOrder godoc
//
//	@Summary		Add a new order
//	@Description	Allow authenticated users to add a new order using a valid order number.
//	@Tags			Orders
//	@Accept			text/plain
//	@Produce		json
//	@Param			orderNumber	body	string	true	"Order number to be added"
//	@Security		BearerAuth
//	@Success		202	{object}	utils.Response	"New order has been accepted for processing"
//	@Success		200	{object}	utils.Response	"Order already uploaded by this user"
//	@Failure		400	{object}	utils.Response	"Bad request due to incorrect order number format"
//	@Failure		401	{object}	utils.Response	"User not authorized"
//	@Failure		409	{object}	utils.Response	"Order already uploaded by another user"
//	@Failure		422	{object}	utils.Response	"Invalid order number format"
//	@Failure		500	{object}	utils.Response	"Internal server error"
//	@Router			/api/user/orders [post]
func (h *OrderHandler) AddOrder(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	orderNumber := string(body)

	if orderNumber == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Order number is required")
		return
	}

	ok := validate.IsLuna(orderNumber)
	if !ok {
		utils.RespondWithError(w, http.StatusUnprocessableEntity, "Invalid order number")
		return
	}
	resp, err := h.orderService.ProcessOrder(r.Context(), userID, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, orderservice.ErrOrderAlreadyExistsByUser):
			utils.RespondWithError(w, http.StatusOK, err.Error())
		case errors.Is(err, orderservice.ErrOrderAlreadyExists):
			utils.RespondWithError(w, http.StatusConflict, err.Error())
		default:
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	utils.RespondWithJSON(w, http.StatusAccepted, resp)
}

// GetOrders godoc
//
//	@Summary		Get orders list for user
//	@Description	Retrieve a list of uploaded orders for the authorized user
//	@Tags			Orders
//	@Produce		json
//	@Success		200	{array}		dto.GetOrdersResponseDTO
//	@Failure		204	{object}	utils.Response	"No data available"
//	@Failure		401	{object}	utils.Response	"User not authorized"
//	@Failure		500	{object}	utils.Response	"Internal server error"
//	@Router			/api/user/orders [get]
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(int)

	orders, err := h.orderService.GetOrders(r.Context(), userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if len(orders) == 0 {
		utils.RespondWithError(w, http.StatusNoContent, "No data available")
		return
	}

	var response []dto.GetOrdersResponseDTO
	for _, order := range orders {
		response = append(response, dto.GetOrdersResponseDTO{
			Number:     order.OrderNumber,
			Status:     order.Status,
			Accrual:    order.Accrual,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		})
	}
	utils.RespondWithJSON(w, http.StatusOK, response)
}
