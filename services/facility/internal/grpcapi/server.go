package grpcapi

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"kursovaya_aksp/services/facility/internal/models"
	"kursovaya_aksp/services/facility/internal/store"
	"kursovaya_aksp/services/facility/pkg/grpc/facilitypb"
)

// Server предоставляет gRPC-доступ к данным площадок.
type Server struct {
	facilitypb.UnimplementedFacilityServiceServer
	store *store.Store
}

// New создаёт gRPC сервер.
func New(store *store.Store) *Server {
	return &Server{store: store}
}

// GetFacility возвращает площадку по идентификатору.
func (s *Server) GetFacility(ctx context.Context, req *facilitypb.GetFacilityRequest) (*facilitypb.GetFacilityResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}
	facility, ok := s.store.Get(req.GetId())
	if !ok {
		return nil, status.Error(codes.NotFound, "facility not found")
	}
	return &facilitypb.GetFacilityResponse{Facility: toProtoFacility(facility)}, nil
}

// ListAvailability возвращает расписание доступности площадки.
func (s *Server) ListAvailability(ctx context.Context, req *facilitypb.ListAvailabilityRequest) (*facilitypb.ListAvailabilityResponse, error) {
	if strings.TrimSpace(req.GetFacilityId()) == "" {
		return nil, status.Error(codes.InvalidArgument, "facility_id required")
	}
	slots, ok := s.store.GetAvailability(req.GetFacilityId())
	if !ok {
		return nil, status.Error(codes.NotFound, "facility not found")
	}
	return &facilitypb.ListAvailabilityResponse{Slots: toProtoSlots(slots)}, nil
}

func toProtoFacility(src *models.Facility) *facilitypb.Facility {
	if src == nil {
		return nil
	}
	return &facilitypb.Facility{
		Id:          src.ID,
		Name:        src.Name,
		City:        src.City,
		Address:     src.Address,
		Type:        src.Type,
		Amenities:   append([]string{}, src.Amenities...),
		Description: src.Description,
	}
}

func toProtoSlots(slots []models.AvailabilitySlot) []*facilitypb.AvailabilitySlot {
	if len(slots) == 0 {
		return nil
	}
	out := make([]*facilitypb.AvailabilitySlot, 0, len(slots))
	for _, slot := range slots {
		out = append(out, &facilitypb.AvailabilitySlot{
			Start: timestamppb.New(slot.Start),
			End:   timestamppb.New(slot.End),
		})
	}
	return out
}
