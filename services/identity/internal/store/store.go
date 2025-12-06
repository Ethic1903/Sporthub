package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"kursovaya_aksp/services/identity/internal/models"
)

const (
	queryTimeout = 5 * time.Second
	tokenTTL     = 24 * time.Hour
)

var (
	errUserNotFound = errors.New("user not found")
	errTokenInvalid = errors.New("token not found")
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// Store работает поверх PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// SeedDemoUsers добавляет пару тестовых аккаунтов для быстрого старта.
func (s *Store) SeedDemoUsers() {
	_, _ = s.CreateUser("admin@sporthub.io", "admin", "System Admin", "admin", "")
	_, _ = s.CreateUser("coach@sporthub.io", "coach", "Demo Coach", "coach", "")
}

func (s *Store) CreateUser(email, password, fullName, role, phone string) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, errors.New("email required")
	}
	if password == "" {
		return nil, errors.New("password required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	id := uuid.NewString()
	ctx, cancel := s.ctx()
	defer cancel()
	var user models.User
	query, args, err := psql.Insert("users").
		Columns("id", "email", "password_hash", "full_name", "role", "phone").
		Values(id, email, hash, fullName, role, phone).
		Suffix("RETURNING id, email, full_name, role, phone, created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, query, args...).
		Scan(&user.ID, &user.Email, &user.FullName, &user.Role, &user.Phone, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, errors.New("email already registered")
		}
		return nil, err
	}
	return &user, nil
}

func (s *Store) Authenticate(email, password string) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	ctx, cancel := s.ctx()
	defer cancel()
	var (
		storedHash []byte
		user       models.User
	)
	query, args, err := psql.Select("id", "password_hash", "email", "full_name", "role", "phone", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, query, args...).
		Scan(&user.ID, &storedHash, &user.Email, &user.FullName, &user.Role, &user.Phone, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(storedHash, []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &user, nil
}

func (s *Store) IssueToken(userID string) (string, error) {
	token := uuid.NewString()
	expiresAt := time.Now().Add(tokenTTL)
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Insert("sessions").
		Columns("token", "user_id", "expires_at").
		Values(token, userID, expiresAt).
		ToSql()
	if err != nil {
		return "", err
	}
	if _, err := s.pool.Exec(ctx, query, args...); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Store) RevokeToken(token string) {
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Delete("sessions").Where(squirrel.Eq{"token": token}).ToSql()
	if err != nil {
		return
	}
	_, _ = s.pool.Exec(ctx, query, args...)
}

func (s *Store) UserByToken(token string) (*models.User, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Select(
		"u.id", "u.email", "u.full_name", "u.role", "u.phone", "u.created_at", "u.updated_at",
	).
		From("sessions s").
		Join("users u ON u.id = s.user_id").
		Where(squirrel.Eq{"s.token": token}).
		Where(squirrel.Expr("s.expires_at > NOW()")).
		ToSql()
	if err != nil {
		return nil, err
	}
	row := s.pool.QueryRow(ctx, query, args...)
	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, errUserNotFound) {
			return nil, errTokenInvalid
		}
		return nil, err
	}
	return user, nil
}

func (s *Store) Get(id string) (*models.User, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Select("id", "email", "full_name", "role", "phone", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}
	row := s.pool.QueryRow(ctx, query, args...)
	return scanUser(row)
}

func (s *Store) Update(id string, fn func(u *models.User)) (*models.User, error) {
	user, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	fn(user)
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Update("users").
		Set("full_name", user.FullName).
		Set("role", user.Role).
		Set("phone", user.Phone).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": id}).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := s.pool.QueryRow(ctx, query, args...).
		Scan(&user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (s *Store) UpdateRole(id, role string) (*models.User, error) {
	return s.Update(id, func(u *models.User) {
		u.Role = role
	})
}

func (s *Store) ListRaw() ([]*models.User, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	query, args, err := psql.Select("id", "email", "full_name", "role", "phone", "created_at", "updated_at").
		From("users").
		OrderBy("created_at").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]*models.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Store) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), queryTimeout)
}

func scanUser(row pgx.Row) (*models.User, error) {
	var user models.User
	if err := row.Scan(&user.ID, &user.Email, &user.FullName, &user.Role, &user.Phone, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}
