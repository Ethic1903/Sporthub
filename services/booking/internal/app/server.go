package app

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"kursovaya_aksp/services/booking/internal/store"
	"kursovaya_aksp/services/facility/pkg/grpc/facilitypb"
	"kursovaya_aksp/services/identity/pkg/grpc/identitypb"
)

const grpcTimeout = 5 * time.Second

// Server обслуживает HTTP ручки бронирования.
type Server struct {
	store          *store.Store
	facilityClient facilitypb.FacilityServiceClient
	identityClient identitypb.IdentityServiceClient
	router         *chi.Mux
}

func New(store *store.Store, facilityClient facilitypb.FacilityServiceClient, identityClient identitypb.IdentityServiceClient) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	s := &Server{
		store:          store,
		facilityClient: facilityClient,
		identityClient: identityClient,
		router:         r,
	}

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	r.Route("/bookings", func(rt chi.Router) {
		rt.Get("/", s.handleList)
		rt.Post("/", s.handleCreate)
		rt.Route("/{id}", func(sub chi.Router) {
			sub.Get("/", s.handleGet)
			sub.Patch("/", s.handlePatch)
			sub.Delete("/", s.handleDelete)
			sub.Post("/confirm", s.handleConfirm)
			sub.Post("/cancel", s.handleCancel)
		})
	})

	r.Get("/availability/search", s.handleAvailabilitySearch)

	return s
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) ensureFacilityExists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := s.facilityClient.GetFacility(ctx, &facilitypb.GetFacilityRequest{Id: id})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Server) ensureUserExists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, nil
	}
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()
	_, err := s.identityClient.GetUser(ctx, &identitypb.GetUserRequest{Id: id})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
