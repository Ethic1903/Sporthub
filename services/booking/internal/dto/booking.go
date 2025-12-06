package dto

import "time"

// CreateBookingRequest описывает POST /bookings.
type CreateBookingRequest struct {
	FacilityID   string    `json:"facility_id"`
	UserID       string    `json:"user_id"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
	Participants int       `json:"participants"`
	Price        float64   `json:"price"`
	Notes        string    `json:"notes"`
}

// PatchBookingRequest описывает PATCH /bookings/{id}.
type PatchBookingRequest struct {
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
	Notes    *string    `json:"notes"`
}
