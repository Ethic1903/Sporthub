package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"kursovaya_aksp/services/facility/internal/dto"
	"kursovaya_aksp/services/facility/internal/models"
)

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	city := strings.ToLower(r.URL.Query().Get("city"))
	facilityType := strings.ToLower(r.URL.Query().Get("type"))
	items := s.store.List(city, facilityType)
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.FacilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	facility := models.Facility{
		Name:        req.Name,
		City:        req.City,
		Address:     req.Address,
		Type:        req.Type,
		Amenities:   req.Amenities,
		Description: req.Description,
		Maintenance: req.Maintenance,
	}
	created, err := s.store.Create(facility)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if facility, ok := s.store.Get(id); ok {
		writeJSON(w, http.StatusOK, facility)
		return
	}
	writeError(w, http.StatusNotFound, "facility not found")
}

func (s *Server) handlePatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.FacilityUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	updated, err := s.store.Update(id, func(target *models.Facility) {
		if req.Name != nil {
			target.Name = *req.Name
		}
		if req.City != nil {
			target.City = *req.City
		}
		if req.Address != nil {
			target.Address = *req.Address
		}
		if req.Type != nil {
			target.Type = *req.Type
		}
		if req.Description != nil {
			target.Description = *req.Description
		}
		if req.Amenities != nil {
			target.Amenities = *req.Amenities
		}
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
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
	writeError(w, http.StatusNotFound, "facility not found")
}

func (s *Server) handleAvailability(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	slots, ok := s.store.GetAvailability(id)
	if !ok {
		writeError(w, http.StatusNotFound, "facility not found")
		return
	}
	writeJSON(w, http.StatusOK, slots)
}

func (s *Server) handleUpdateAvailability(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.AvailabilityUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := s.store.UpdateAvailability(id, req.Slots); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
