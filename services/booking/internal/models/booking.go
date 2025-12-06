package models

import "time"

// Booking описывает бронь площадки.
type Booking struct {
	ID           string    `json:"id"`
	FacilityID   string    `json:"facility_id"`
	UserID       string    `json:"user_id"`
	Status       string    `json:"status"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
	Participants int       `json:"participants"`
	Price        float64   `json:"price"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateInput содержит обязательные поля для новой брони.
type CreateInput struct {
	FacilityID   string
	UserID       string
	StartsAt     time.Time
	EndsAt       time.Time
	Participants int
	Price        float64
	Notes        string
}

// Filter задаёт условия выборки.
type Filter struct {
	FacilityID string
	UserID     string
}
