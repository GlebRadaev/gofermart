package dto

type GetOrdersResponseDTO struct {
	Number     string  `json:"number" example:"1234567890"`
	Status     string  `json:"status" example:"PROCESSED"`
	Accrual    float64 `json:"accrual,omitempty" example:"500"`
	UploadedAt string  `json:"uploaded_at" example:"2020-12-09T16:09:57+03:00"`
}
