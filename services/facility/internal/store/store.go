package store

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"kursovaya_aksp/services/facility/internal/models"
)

const queryTimeout = 5 * time.Second

var (
	errFacilityNotFound = errors.New("facility not found")
	psql                = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
)

// Store держит данные по площадкам в PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) SeedDemoData() {
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Select("COUNT(*)").From("facilities").ToSql()
	if err != nil {
		return
	}
	var count int
	if err := s.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil || count > 0 {
		return
	}
	_, _ = s.Create(models.Facility{
		Name:        "Sportzal Olimp",
		City:        "Moscow",
		Address:     "Arbat 1",
		Type:        "indoor",
		Amenities:   []string{"shower", "locker_room", "parking"},
		Description: "Multipurpose indoor arena",
	})
	_, _ = s.Create(models.Facility{
		Name:        "Manezh Veter",
		City:        "Saint-Petersburg",
		Address:     "Nevsky 12",
		Type:        "outdoor",
		Amenities:   []string{"lights", "stands"},
		Description: "Comfortable outdoor football field",
	})
}

func (s *Store) List(city, facilityType string) []models.Facility {
	ctx, cancel := s.ctx()
	defer cancel()
	builder := psql.Select(
		"id", "name", "city", "address", "type", "amenities", "description", "maintenance", "created_at", "updated_at",
	).
		From("facilities").
		OrderBy("created_at")
	if city != "" {
		builder = builder.Where(squirrel.Expr("LOWER(city) = ?", strings.ToLower(city)))
	}
	if facilityType != "" {
		builder = builder.Where(squirrel.Expr("LOWER(type) = ?", strings.ToLower(facilityType)))
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return []models.Facility{}
	}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return []models.Facility{}
	}
	defer rows.Close()
	result := make([]models.Facility, 0)
	for rows.Next() {
		facility, err := scanFacility(rows)
		if err != nil {
			return []models.Facility{}
		}
		result = append(result, *facility)
	}
	return result
}

func (s *Store) Create(data models.Facility) (*models.Facility, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	if data.Amenities == nil {
		data.Amenities = []string{}
	}
	if data.Maintenance == nil {
		data.Maintenance = []models.Maintenance{}
	}
	availability := defaultAvailability()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	amJSON, err := json.Marshal(data.Amenities)
	if err != nil {
		return nil, err
	}
	maintJSON, err := json.Marshal(data.Maintenance)
	if err != nil {
		return nil, err
	}
	data.ID = uuid.NewString()
	insert, args, err := psql.Insert("facilities").
		Columns("id", "name", "city", "address", "type", "amenities", "description", "maintenance").
		Values(data.ID, data.Name, data.City, data.Address, data.Type, amJSON, data.Description, maintJSON).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := tx.QueryRow(ctx, insert, args...).
		Scan(&data.CreatedAt, &data.UpdatedAt); err != nil {
		return nil, err
	}
	if err := replaceAvailability(ctx, tx, data.ID, availability); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &data, nil
}

func (s *Store) Update(id string, fn func(f *models.Facility)) (*models.Facility, error) {
	facility, ok := s.Get(id)
	if !ok {
		return nil, errFacilityNotFound
	}
	fn(facility)
	ctx, cancel := s.ctx()
	defer cancel()
	amJSON, err := json.Marshal(facility.Amenities)
	if err != nil {
		return nil, err
	}
	maintJSON, err := json.Marshal(facility.Maintenance)
	if err != nil {
		return nil, err
	}
	query, args, err := psql.Update("facilities").
		Set("name", facility.Name).
		Set("city", facility.City).
		Set("address", facility.Address).
		Set("type", facility.Type).
		Set("amenities", amJSON).
		Set("description", facility.Description).
		Set("maintenance", maintJSON).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, query, args...).
		Scan(&facility.CreatedAt, &facility.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errFacilityNotFound
		}
		return nil, err
	}
	return facility, nil
}

func (s *Store) Delete(id string) bool {
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Delete("facilities").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return false
	}
	cmd, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return false
	}
	return cmd.RowsAffected() > 0
}

func (s *Store) Get(id string) (*models.Facility, bool) {
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Select("id", "name", "city", "address", "type", "amenities", "description", "maintenance", "created_at", "updated_at").
		From("facilities").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, false
	}
	row := s.pool.QueryRow(ctx, query, args...)
	facility, err := scanFacility(row)
	if err != nil {
		return nil, false
	}
	return facility, true
}

func (s *Store) GetAvailability(id string) ([]models.AvailabilitySlot, bool) {
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Select("start_at", "end_at").
		From("facility_availability").
		Where(squirrel.Eq{"facility_id": id}).
		OrderBy("start_at").
		ToSql()
	if err != nil {
		return nil, false
	}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false
	}
	defer rows.Close()
	slots := make([]models.AvailabilitySlot, 0)
	for rows.Next() {
		var slot models.AvailabilitySlot
		if err := rows.Scan(&slot.Start, &slot.End); err != nil {
			return nil, false
		}
		slots = append(slots, slot)
	}
	if len(slots) == 0 && !s.facilityExists(ctx, id) {
		return nil, false
	}
	return slots, true
}

func (s *Store) UpdateAvailability(id string, slots []models.AvailabilitySlot) error {
	ctx, cancel := s.ctx()
	defer cancel()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if !s.facilityExistsTx(ctx, tx, id) {
		return errFacilityNotFound
	}
	if err := replaceAvailability(ctx, tx, id, slots); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func defaultAvailability() []models.AvailabilitySlot {
	slots := make([]models.AvailabilitySlot, 0, 7)
	now := time.Now().UTC().Truncate(24 * time.Hour)
	for i := 0; i < 7; i++ {
		start := now.AddDate(0, 0, i).Add(8 * time.Hour)
		end := start.Add(10 * time.Hour)
		slots = append(slots, models.AvailabilitySlot{Start: start, End: end})
	}
	return slots
}

func (s *Store) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), queryTimeout)
}

func scanFacility(row pgx.Row) (*models.Facility, error) {
	var (
		facility      models.Facility
		amenitiesJSON []byte
		maintJSON     []byte
	)
	if err := row.Scan(
		&facility.ID,
		&facility.Name,
		&facility.City,
		&facility.Address,
		&facility.Type,
		&amenitiesJSON,
		&facility.Description,
		&maintJSON,
		&facility.CreatedAt,
		&facility.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errFacilityNotFound
		}
		return nil, err
	}
	if len(amenitiesJSON) > 0 {
		_ = json.Unmarshal(amenitiesJSON, &facility.Amenities)
	}
	if len(maintJSON) > 0 {
		_ = json.Unmarshal(maintJSON, &facility.Maintenance)
	}
	return &facility, nil
}

func replaceAvailability(ctx context.Context, tx pgx.Tx, facilityID string, slots []models.AvailabilitySlot) error {
	deleteSQL, deleteArgs, err := psql.Delete("facility_availability").
		Where(squirrel.Eq{"facility_id": facilityID}).
		ToSql()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, deleteSQL, deleteArgs...); err != nil {
		return err
	}
	for _, slot := range slots {
		insertSQL, insertArgs, err := psql.Insert("facility_availability").
			Columns("id", "facility_id", "start_at", "end_at").
			Values(uuid.NewString(), facilityID, slot.Start, slot.End).
			ToSql()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, insertSQL, insertArgs...); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) facilityExists(ctx context.Context, id string) bool {
	query, args, err := psql.Select("COUNT(1)").From("facilities").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return false
	}
	var count int
	if err := s.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func (s *Store) facilityExistsTx(ctx context.Context, tx pgx.Tx, id string) bool {
	query, args, err := psql.Select("COUNT(1)").From("facilities").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return false
	}
	var count int
	if err := tx.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return false
	}
	return count > 0
}
