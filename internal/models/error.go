package models

const (
	// Ошибки
	ErrBadRequest   = "bad request"
	ErrUnauthorized = "unauthorized"
	ErrIternal      = "internal server error"
)

// Модель для ответа с ошибкой
type ErrorResponse struct {
	Errors string `json:"errors"`
}

func NewErrorResponse(title, details string) ErrorResponse {
	return ErrorResponse{
		Errors: title + ": " + details,
	}
}
