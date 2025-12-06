package dto

import "kursovaya_aksp/services/facility/internal/models"

// FacilityRequest описывает POST /facilities.
type FacilityRequest struct {
	Name        string               `json:"name"`
	City        string               `json:"city"`
	Address     string               `json:"address"`
	Type        string               `json:"type"`
	Amenities   []string             `json:"amenities"`
	Description string               `json:"description"`
	Maintenance []models.Maintenance `json:"maintenance"`
}

// FacilityUpdateRequest описывает PATCH /facilities/{id}.
type FacilityUpdateRequest struct {
	Name        *string   `json:"name"`
	City        *string   `json:"city"`
	Address     *string   `json:"address"`
	Type        *string   `json:"type"`
	Amenities   *[]string `json:"amenities"`
	Description *string   `json:"description"`
}

// AvailabilityUpdateRequest описывает PUT /facilities/{id}/availability.
type AvailabilityUpdateRequest struct {
	Slots []models.AvailabilitySlot `json:"slots"`
}
