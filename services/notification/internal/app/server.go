package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"kursovaya_aksp/services/notification/internal/store"
)

// Server обслуживает REST ручки уведомлений.
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

	r.Route("/notifications", func(rt chi.Router) {
		rt.Get("/", s.handleList)
		rt.Post("/send", s.handleSend)
		rt.Post("/test", s.handleTest)
		rt.Get("/{id}", s.handleGet)
	})

	return s
}

func (s *Server) Router() http.Handler {
	return s.router
}
