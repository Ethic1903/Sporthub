package app

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	graphapi "kursovaya_aksp/services/api-gateway/internal/graphql"
	"kursovaya_aksp/services/api-gateway/internal/graphql/generated"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Config описывает URL внутренних сервисов и HTTP-клиента.
type Config struct {
	IdentityURL     string
	FacilityURL     string
	BookingURL      string
	NotificationURL string
	Client          *http.Client
}

// Server реализует API Gateway поверх chi.
type Server struct {
	cfg    Config
	router *chi.Mux
}

// New создаёт сервер и настраивает маршруты/proxy.
func New(cfg Config) *Server {
	if cfg.Client == nil {
		cfg.Client = &http.Client{Timeout: 10 * time.Second}
	}
	resolver := graphapi.NewResolver(graphapi.ResolverConfig{
		IdentityURL:     cfg.IdentityURL,
		FacilityURL:     cfg.FacilityURL,
		BookingURL:      cfg.BookingURL,
		NotificationURL: cfg.NotificationURL,
		Client:          cfg.Client,
	})
	execSchema := generated.NewExecutableSchema(generated.Config{Resolvers: resolver})
	graphqlHandler := handler.NewDefaultServer(execSchema)
	playgroundHandler := playground.Handler("GraphQL playground", "/graphql")
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	s := &Server{cfg: cfg, router: r}
	r.Handle("/graphql", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			playgroundHandler.ServeHTTP(w, req)
			return
		}
		ctx := graphapi.WithRequestHeaders(req.Context(), req.Header)
		graphqlHandler.ServeHTTP(w, req.WithContext(ctx))
	}))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	r.Route("/auth", func(rt chi.Router) {
		rt.Post("/register", s.proxyFixed(cfg.IdentityURL+"/auth/register"))
		rt.Post("/login", s.proxyFixed(cfg.IdentityURL+"/auth/login"))
		rt.Post("/logout", s.proxyFixed(cfg.IdentityURL+"/auth/logout"))
		rt.Post("/refresh", s.proxyFixed(cfg.IdentityURL+"/auth/refresh"))
	})

	r.Route("/users", func(rt chi.Router) {
		rt.Get("/me", s.proxyFixed(cfg.IdentityURL+"/users/me"))
		rt.Patch("/me", s.proxyFixed(cfg.IdentityURL+"/users/me"))
	})

	r.Mount("/facilities", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s.proxyWithPrefix(w, req, cfg.FacilityURL, "/facilities")
	}))

	r.Mount("/bookings", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s.proxyWithPrefix(w, req, cfg.BookingURL, "/bookings")
	}))

	r.Mount("/availability", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s.proxyWithPrefix(w, req, cfg.BookingURL, "/availability")
	}))

	r.Mount("/notifications", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s.proxyWithPrefix(w, req, cfg.NotificationURL, "/notifications")
	}))

	return s
}

// Router возвращает настроенный http.Handler.
func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) proxyFixed(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.forward(w, r, target)
	}
}

func (s *Server) proxyWithPrefix(w http.ResponseWriter, r *http.Request, base, prefix string) {
	suffix := strings.TrimPrefix(r.URL.Path, prefix)
	if suffix == "" {
		suffix = "/"
	}
	target := strings.TrimSuffix(base, "/") + suffix
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	s.forward(w, r, target)
}

func (s *Server) forward(w http.ResponseWriter, r *http.Request, target string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, bytes.NewReader(body))
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}
	copyHeaders(req.Header, r.Header)
	resp, err := s.cfg.Client.Do(req)
	if err != nil {
		http.Error(w, "target unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func copyHeaders(dst, src http.Header) {
	for k, values := range src {
		for _, v := range values {
			dst.Add(k, v)
		}
	}
}
