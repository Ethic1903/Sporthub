package dto

import "kursovaya_aksp/services/identity/internal/models"

// RegisterRequest описывает тело регистрации.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	Phone    string `json:"phone"`
}

// LoginRequest описывает запрос логина.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ProfileUpdateRequest описывает PATCH /users/me.
type ProfileUpdateRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
}

// TokenRequest используется для внутренних операций с токеном.
type TokenRequest struct {
	Token string `json:"token"`
}

// TokenResponse возвращает пару токен + пользователь.
type TokenResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// StatusResponse сообщает о простом результате операции.
type StatusResponse struct {
	Status string `json:"status"`
}
