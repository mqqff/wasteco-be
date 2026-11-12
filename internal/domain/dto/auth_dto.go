package dto

type GoogleCallbackRequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state" validate:"required"`
	Error string `json:"error"`
}

type RegisterRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	PhoneNumber string `json:"phone_number"`
	AreaName    string `json:"area_name"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}
