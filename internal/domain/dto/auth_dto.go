package dto

type GoogleCallbackRequest struct {
	Code  string `json:"code" validate:"required"`
	Error string `json:"error"`
}
