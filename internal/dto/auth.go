package dto

type RegisterRequestDTO struct {
	Login    string `json:"login" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterResponseDTO struct {
	Message string `json:"message"`
}

type LoginRequestDTO struct {
	Login    string `json:"login" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponseDTO struct {
	Message string `json:"message"`
}
