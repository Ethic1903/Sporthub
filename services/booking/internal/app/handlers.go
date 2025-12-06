package app

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"kursovaya_aksp/services/booking/internal/dto"
	"kursovaya_aksp/services/booking/internal/models"
)

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	filter := models.Filter{
		FacilityID: r.URL.Query().Get("facility_id"),
		UserID:     r.URL.Query().Get("user_id"),
	}
	items := s.store.List(filter)
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.FacilityID == "" || req.UserID == "" {
		writeError(w, http.StatusBadRequest, "facility_id and user_id required")
		return
	}
	if req.EndsAt.Before(req.StartsAt) {
		writeError(w, http.StatusBadRequest, "ends_at must be after starts_at")
		return
	}
	exists, err := s.ensureFacilityExists(r.Context(), req.FacilityID)
	if err != nil {
		writeError(w, http.StatusBadGateway, "facility service unavailable")
		return
	}
	if !exists {
		writeError(w, http.StatusBadRequest, "unknown facility")
		return
	}
	userExists, err := s.ensureUserExists(r.Context(), req.UserID)
	if err != nil {
		writeError(w, http.StatusBadGateway, "identity service unavailable")
		return
	}
	if !userExists {
		writeError(w, http.StatusBadRequest, "unknown user")
		return
	}
	if s.store.HasOverlap(req.FacilityID, req.StartsAt, req.EndsAt, "") {
		writeError(w, http.StatusConflict, "slot already booked")
		return
	}
	booking, err := s.store.Create(models.CreateInput{
		FacilityID:   req.FacilityID,
		UserID:       req.UserID,
		StartsAt:     req.StartsAt,
		EndsAt:       req.EndsAt,
		Participants: req.Participants,
		Price:        req.Price,
		Notes:        req.Notes,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, booking)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if booking, ok := s.store.Get(id); ok {
		writeJSON(w, http.StatusOK, booking)
		return
	}
	writeError(w, http.StatusNotFound, "booking not found")
}

func (s *Server) handlePatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.PatchBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	existing, ok := s.store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "booking not found")
		return
	}
	newStart := existing.StartsAt
	if req.StartsAt != nil {
		newStart = *req.StartsAt
	}
	newEnd := existing.EndsAt
	if req.EndsAt != nil {
		newEnd = *req.EndsAt
	}
	if newEnd.Before(newStart) {
		writeError(w, http.StatusBadRequest, "ends_at must be after starts_at")
		return
	}
	if s.store.HasOverlap(existing.FacilityID, newStart, newEnd, existing.ID) {
		writeError(w, http.StatusConflict, "slot already booked")
		return
	}
	updated, err := s.store.Update(id, req.StartsAt, req.EndsAt, req.Notes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if s.store.Delete(id) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeError(w, http.StatusNotFound, "booking not found")
}

func (s *Server) handleConfirm(w http.ResponseWriter, r *http.Request) {
	s.updateStatus(w, r, "confirmed")
}

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	s.updateStatus(w, r, "cancelled")
}

func (s *Server) updateStatus(w http.ResponseWriter, r *http.Request, status string) {
	id := chi.URLParam(r, "id")
	updated, err := s.store.UpdateStatus(id, status)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleAvailabilitySearch(w http.ResponseWriter, r *http.Request) {
	facilityID := r.URL.Query().Get("facility_id")
	if facilityID == "" {
		writeError(w, http.StatusBadRequest, "facility_id required")
		return
	}
	dateStr := r.URL.Query().Get("date")
	day := time.Now()
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid date format")
			return
		}
		day = parsed
	}
	exists, err := s.ensureFacilityExists(r.Context(), facilityID)
	if err != nil {
		writeError(w, http.StatusBadGateway, "facility service unavailable")
		return
	}
	if !exists {
		writeError(w, http.StatusBadRequest, "unknown facility")
		return
	}
	slots := s.computeAvailability(facilityID, day)
	writeJSON(w, http.StatusOK, slots)
}

func (s *Server) computeAvailability(facilityID string, day time.Time) []map[string]any {
	result := make([]map[string]any, 0)
	start := day.Truncate(24 * time.Hour).Add(8 * time.Hour)
	for i := 0; i < 8; i++ {
		slotStart := start.Add(time.Duration(i) * time.Hour)
		slotEnd := slotStart.Add(time.Hour)
		booked := s.store.HasOverlap(facilityID, slotStart, slotEnd, "")
		result = append(result, map[string]any{
			"start": slotStart,
			"end":   slotEnd,
			"free":  !booked,
		})
	}
	return result
}
