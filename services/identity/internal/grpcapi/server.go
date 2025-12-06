package grpcapi

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"kursovaya_aksp/services/identity/internal/models"
	"kursovaya_aksp/services/identity/internal/store"
	"kursovaya_aksp/services/identity/pkg/grpc/identitypb"
)

// Server реализует gRPC API поверх стора пользователей.
type Server struct {
	identitypb.UnimplementedIdentityServiceServer
	store *store.Store
}

// New создаёт gRPC сервер с доступом к стору.
func New(store *store.Store) *Server {
	return &Server{store: store}
}

// GetUser возвращает профиль пользователя по идентификатору.
func (s *Server) GetUser(ctx context.Context, req *identitypb.GetUserRequest) (*identitypb.GetUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	user, err := s.store.Get(req.GetId())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &identitypb.GetUserResponse{User: toProtoUser(user)}, nil
}

// ValidateToken валидирует токен и возвращает профиль владельца.
func (s *Server) ValidateToken(ctx context.Context, req *identitypb.ValidateTokenRequest) (*identitypb.ValidateTokenResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token required")
	}
	user, err := s.store.UserByToken(req.GetToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &identitypb.ValidateTokenResponse{User: toProtoUser(user)}, nil
}

func toProtoUser(user *models.User) *identitypb.User {
	if user == nil {
		return nil
	}
	return &identitypb.User{
		Id:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		Phone:     user.Phone,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}
