package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"kursovaya_aksp/services/identity/internal/dto"
	"kursovaya_aksp/services/identity/internal/models"
)

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	role := req.Role
	if role == "" {
		role = "athlete"
	}
	u, err := s.store.CreateUser(req.Email, req.Password, req.FullName, role, req.Phone)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	user, err := s.store.Authenticate(req.Email, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	token, err := s.store.IssueToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dto.TokenResponse{Token: token, User: user})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		writeError(w, http.StatusBadRequest, "missing token")
		return
	}
	s.store.RevokeToken(token)
	writeJSON(w, http.StatusOK, dto.StatusResponse{Status: "logged_out"})
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		writeError(w, http.StatusBadRequest, "missing token")
		return
	}
	u, err := s.store.UserByToken(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	s.store.RevokeToken(token)
	newToken, err := s.store.IssueToken(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dto.TokenResponse{Token: newToken, User: u})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	u, err := s.userFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	u, err := s.userFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req dto.ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	updated, err := s.store.Update(u.ID, func(target *models.User) {
		if req.FullName != "" {
			target.FullName = req.FullName
		}
		if req.Phone != "" {
			target.Phone = req.Phone
		}
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleGetUserInternal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u, err := s.store.Get(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handlePatchUserInternal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Role == "" {
		writeError(w, http.StatusBadRequest, "role required")
		return
	}
	updated, err := s.store.UpdateRole(id, req.Role)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleValidateToken(w http.ResponseWriter, r *http.Request) {
	var req dto.TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Token) == "" {
		writeError(w, http.StatusBadRequest, "token required")
		return
	}
	u, err := s.store.UserByToken(req.Token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": u})
}

func (s *Server) userFromRequest(r *http.Request) (*models.User, error) {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		return nil, errors.New("missing bearer token")
	}
	return s.store.UserByToken(token)
}
