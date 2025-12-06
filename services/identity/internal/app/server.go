package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"kursovaya_aksp/services/identity/internal/store"
)

// Server инкапсулирует HTTP-роутер и стор пользователей.
type Server struct {
	store  *store.Store
	router *chi.Mux
}

// New создаёт сервер, настраивает middleware и маршруты.
func New(store *store.Store) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	s := &Server{store: store, router: r}
	r.Route("/auth", func(rt chi.Router) {
		rt.Post("/register", s.handleRegister)
		rt.Post("/login", s.handleLogin)
		rt.Post("/logout", s.handleLogout)
		rt.Post("/refresh", s.handleRefresh)
	})

	r.Route("/users", func(rt chi.Router) {
		rt.Get("/me", s.handleMe)
		rt.Patch("/me", s.handleUpdateMe)
	})

	r.Route("/internal", func(rt chi.Router) {
		rt.Get("/users/{id}", s.handleGetUserInternal)
		rt.Patch("/users/{id}", s.handlePatchUserInternal)
		rt.Post("/auth/validate", s.handleValidateToken)
	})

	return s
}

// Router возвращает настроенный http.Handler для запуска сервиса.
func (s *Server) Router() http.Handler {
	return s.router
}
