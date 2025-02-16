package models

import "fmt"

const (
	ErrBadRequest   = "bad request"
	ErrUnauthorized = "unauthorized"
	ErrInternal     = "internal server error"
)

// Модель для ответа с ошибкой
type ErrorResponse struct {
	Errors string `json:"errors"`
}

func NewErrorResponse(errMsg string) ErrorResponse {
	return ErrorResponse{
		Errors: errMsg,
	}
}

func NewDetailedErrorResponse(errMsg, details string) ErrorResponse {
	return ErrorResponse{
		Errors: fmt.Sprintf("%s: %s", errMsg, details),
	}
}
