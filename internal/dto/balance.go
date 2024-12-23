package dto

import "time"

type BalanceResponseDTO struct {
	Current   float64 `json:"current" example:"500.5"`
	Withdrawn float64 `json:"withdrawn" example:"42"`
}
type BalanceWithdrawRequestDTO struct {
	Order string  `json:"order" example:"2377225624"`
	Sum   float64 `json:"sum" example:"500"`
}

type GetWithdrawalsResponseDTO struct {
	Order       string    `json:"order" example:"2377225624"`
	Sum         float64   `json:"sum" example:"500"`
	ProcessedAt time.Time `json:"processed_at" example:"2020-12-09T16:09:57+03:00"`
}
