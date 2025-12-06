package models

import "time"

// Facility описывает объект бронирования.
type Facility struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	City        string        `json:"city"`
	Address     string        `json:"address"`
	Type        string        `json:"type"`
	Amenities   []string      `json:"amenities"`
	Description string        `json:"description"`
	Maintenance []Maintenance `json:"maintenance"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// Maintenance описывает окно техобслуживания.
type Maintenance struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Note  string    `json:"note"`
}

// AvailabilitySlot описывает интервал доступности.
type AvailabilitySlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
