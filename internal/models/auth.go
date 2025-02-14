package models

// Модель для запроса /api/auth
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Модель для ответа /api/auth
type AuthResponse struct {
	Token string `json:"token"`
}
