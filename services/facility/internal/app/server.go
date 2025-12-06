package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"kursovaya_aksp/services/facility/internal/store"
)

// Server обслуживает HTTP-запросы к сервису площадок.
type Server struct {
	store  *store.Store
	router *chi.Mux
}

func New(store *store.Store) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	s := &Server{store: store, router: r}

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	r.Route("/facilities", func(rt chi.Router) {
		rt.Get("/", s.handleList)
		rt.Post("/", s.handleCreate)
		rt.Get("/{id}", s.handleGet)
		rt.Patch("/{id}", s.handlePatch)
		rt.Delete("/{id}", s.handleDelete)
		rt.Get("/{id}/availability", s.handleAvailability)
		rt.Put("/{id}/availability", s.handleUpdateAvailability)
	})

	return s
}

func (s *Server) Router() http.Handler {
	return s.router
}
