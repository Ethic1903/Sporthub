package app

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	channel := r.URL.Query().Get("channel")
	items := s.store.List(channel)
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Channel   string `json:"channel"`
		Recipient string `json:"recipient"`
		Subject   string `json:"subject"`
		Body      string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Channel == "" || req.Recipient == "" {
		writeError(w, http.StatusBadRequest, "channel and recipient required")
		return
	}
	msg := s.store.Create(req.Channel, req.Recipient, req.Subject, req.Body)
	writeJSON(w, http.StatusCreated, msg)
}

func (s *Server) handleTest(w http.ResponseWriter, _ *http.Request) {
	msg := s.store.Create("email", "demo@example.com", "Test message", "Notification service is up")
	writeJSON(w, http.StatusOK, msg)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if msg, ok := s.store.Get(id); ok {
		writeJSON(w, http.StatusOK, msg)
		return
	}
	writeError(w, http.StatusNotFound, "notification not found")
}
