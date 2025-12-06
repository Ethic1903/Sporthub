package store

import (
	"context"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"kursovaya_aksp/services/booking/internal/models"
)

const queryTimeout = 5 * time.Second

var (
	errBookingNotFound = errors.New("booking not found")
	psql               = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
)

// Store хранит брони в PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) List(filter models.Filter) []models.Booking {
	ctx, cancel := s.ctx()
	defer cancel()
	builder := psql.Select(
		"id", "facility_id", "user_id", "status", "starts_at", "ends_at", "participants", "price", "notes", "created_at", "updated_at",
	).
		From("bookings").
		OrderBy("starts_at")
	if filter.FacilityID != "" {
		builder = builder.Where(squirrel.Eq{"facility_id": filter.FacilityID})
	}
	if filter.UserID != "" {
		builder = builder.Where(squirrel.Eq{"user_id": filter.UserID})
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return []models.Booking{}
	}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return []models.Booking{}
	}
	defer rows.Close()
	result := make([]models.Booking, 0)
	for rows.Next() {
		booking, err := scanBooking(rows)
		if err != nil {
			return []models.Booking{}
		}
		result = append(result, *booking)
	}
	return result
}

func (s *Store) Create(input models.CreateInput) (*models.Booking, error) {
	booking := &models.Booking{
		ID:           uuid.NewString(),
		FacilityID:   input.FacilityID,
		UserID:       input.UserID,
		Status:       "pending",
		StartsAt:     input.StartsAt,
		EndsAt:       input.EndsAt,
		Participants: input.Participants,
		Price:        input.Price,
		Notes:        input.Notes,
	}
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Insert("bookings").
		Columns("id", "facility_id", "user_id", "status", "starts_at", "ends_at", "participants", "price", "notes").
		Values(booking.ID, booking.FacilityID, booking.UserID, booking.Status, booking.StartsAt, booking.EndsAt, booking.Participants, booking.Price, booking.Notes).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, query, args...).
		Scan(&booking.CreatedAt, &booking.UpdatedAt); err != nil {
		return nil, err
	}
	return booking, nil
}

func (s *Store) Get(id string) (*models.Booking, bool) {
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Select(
		"id", "facility_id", "user_id", "status", "starts_at", "ends_at", "participants", "price", "notes", "created_at", "updated_at",
	).
		From("bookings").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, false
	}
	row := s.pool.QueryRow(ctx, query, args...)
	booking, err := scanBooking(row)
	if err != nil {
		return nil, false
	}
	return booking, true
}

func (s *Store) Update(id string, startsAt, endsAt *time.Time, notes *string) (*models.Booking, error) {
	booking, ok := s.Get(id)
	if !ok {
		return nil, errBookingNotFound
	}
	if startsAt != nil {
		booking.StartsAt = *startsAt
	}
	if endsAt != nil {
		booking.EndsAt = *endsAt
	}
	if notes != nil {
		booking.Notes = *notes
	}
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Update("bookings").
		Set("starts_at", booking.StartsAt).
		Set("ends_at", booking.EndsAt).
		Set("notes", booking.Notes).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, query, args...).
		Scan(&booking.CreatedAt, &booking.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errBookingNotFound
		}
		return nil, err
	}
	return booking, nil
}

func (s *Store) Delete(id string) bool {
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Delete("bookings").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return false
	}
	cmd, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return false
	}
	return cmd.RowsAffected() > 0
}

func (s *Store) UpdateStatus(id, status string) (*models.Booking, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Update("bookings").
		Set("status", status).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING id, facility_id, user_id, status, starts_at, ends_at, participants, price, notes, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, err
	}
	row := s.pool.QueryRow(ctx, query, args...)
	booking, err := scanBooking(row)
	if err != nil {
		return nil, err
	}
	return booking, nil
}

// HasOverlap проверяет пересечения по слотам.
func (s *Store) HasOverlap(facilityID string, start, end time.Time, excludeID string) bool {
	ctx, cancel := s.ctx()
	defer cancel()
	builder := psql.Select("1").From("bookings").
		Where(squirrel.Eq{"facility_id": facilityID}).
		Where(squirrel.Expr("status <> 'cancelled'")).
		Where(squirrel.Expr("? < ends_at", start)).
		Where(squirrel.Expr("? > starts_at", end)).
		Limit(1)
	if excludeID != "" {
		builder = builder.Where(squirrel.NotEq{"id": excludeID})
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return false
	}
	var dummy int
	if err := s.pool.QueryRow(ctx, query, args...).Scan(&dummy); err != nil {
		return false
	}
	return true
}

func (s *Store) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), queryTimeout)
}

func scanBooking(row pgx.Row) (*models.Booking, error) {
	var booking models.Booking
	if err := row.Scan(
		&booking.ID,
		&booking.FacilityID,
		&booking.UserID,
		&booking.Status,
		&booking.StartsAt,
		&booking.EndsAt,
		&booking.Participants,
		&booking.Price,
		&booking.Notes,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errBookingNotFound
		}
		return nil, err
	}
	return &booking, nil
}
