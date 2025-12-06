package graphql

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kursovaya_aksp/services/api-gateway/internal/graphql/model"
)

const graphTimeout = 7 * time.Second

var errNotFound = errors.New("resource not found")

// ResolverConfig contains dependencies required by GraphQL resolvers.
type ResolverConfig struct {
	FacilityURL     string
	BookingURL      string
	IdentityURL     string
	NotificationURL string
	Client          *http.Client
}

// Resolver wires business dependencies into generated resolvers.
type Resolver struct {
	facilityURL     string
	bookingURL      string
	identityURL     string
	notificationURL string
	client          *http.Client
}

// NewResolver clones the provided config and ensures defaults are available for resolvers.
func NewResolver(cfg ResolverConfig) *Resolver {
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Resolver{
		facilityURL:     strings.TrimSuffix(cfg.FacilityURL, "/"),
		bookingURL:      strings.TrimSuffix(cfg.BookingURL, "/"),
		identityURL:     strings.TrimSuffix(cfg.IdentityURL, "/"),
		notificationURL: strings.TrimSuffix(cfg.NotificationURL, "/"),
		client:          client,
	}
}

type facilityDTO struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	City        string           `json:"city"`
	Address     string           `json:"address"`
	Type        string           `json:"type"`
	Amenities   []string         `json:"amenities"`
	Description string           `json:"description"`
	Maintenance []maintenanceDTO `json:"maintenance"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type maintenanceDTO struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Note  string    `json:"note"`
}

type availabilityDTO struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type bookingDTO struct {
	ID           string    `json:"id"`
	FacilityID   string    `json:"facility_id"`
	UserID       string    `json:"user_id"`
	Status       string    `json:"status"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
	Participants int       `json:"participants"`
	Price        float64   `json:"price"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type userDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type authResponse struct {
	Token string  `json:"token"`
	User  userDTO `json:"user"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type profileUpdateRequest struct {
	FullName string `json:"full_name,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

type facilityCreateRequest struct {
	Name        string           `json:"name"`
	City        string           `json:"city"`
	Address     string           `json:"address"`
	Type        string           `json:"type"`
	Amenities   []string         `json:"amenities,omitempty"`
	Description string           `json:"description,omitempty"`
	Maintenance []maintenanceDTO `json:"maintenance,omitempty"`
}

type facilityUpdateRequest struct {
	Name        *string   `json:"name,omitempty"`
	City        *string   `json:"city,omitempty"`
	Address     *string   `json:"address,omitempty"`
	Type        *string   `json:"type,omitempty"`
	Amenities   *[]string `json:"amenities,omitempty"`
	Description *string   `json:"description,omitempty"`
}

type availabilityUpdateRequest struct {
	Slots []availabilityDTO `json:"slots"`
}

type bookingCreateRequest struct {
	FacilityID   string    `json:"facility_id"`
	UserID       string    `json:"user_id"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
	Participants *int      `json:"participants,omitempty"`
	Price        *float64  `json:"price,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
}

type bookingPatchRequest struct {
	StartsAt *time.Time `json:"starts_at,omitempty"`
	EndsAt   *time.Time `json:"ends_at,omitempty"`
	Notes    *string    `json:"notes,omitempty"`
}

func (r *Resolver) listFacilities(ctx context.Context, city, facilityType string) ([]facilityDTO, error) {
	endpoint := r.facilityURL + "/facilities"
	params := url.Values{}
	if city != "" {
		params.Set("city", city)
	}
	if facilityType != "" {
		params.Set("type", facilityType)
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	var payload []facilityDTO
	if err := r.doGet(ctx, endpoint, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (r *Resolver) getFacility(ctx context.Context, id string) (*facilityDTO, error) {
	endpoint := fmt.Sprintf("%s/facilities/%s", r.facilityURL, id)
	var payload facilityDTO
	if err := r.doGet(ctx, endpoint, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (r *Resolver) listBookings(ctx context.Context, facilityID, userID string) ([]bookingDTO, error) {
	endpoint := r.bookingURL + "/bookings"
	params := url.Values{}
	if facilityID != "" {
		params.Set("facility_id", facilityID)
	}
	if userID != "" {
		params.Set("user_id", userID)
	}
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	var payload []bookingDTO
	if err := r.doGet(ctx, endpoint, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (r *Resolver) getAvailability(ctx context.Context, facilityID string) ([]availabilityDTO, error) {
	endpoint := fmt.Sprintf("%s/facilities/%s/availability", r.facilityURL, facilityID)
	var payload []availabilityDTO
	if err := r.doGet(ctx, endpoint, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (r *Resolver) doJSON(ctx context.Context, method, target string, payload any, out any) error {
	ctx, cancel := context.WithTimeout(ctx, graphTimeout)
	defer cancel()
	var body io.Reader
	if payload != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(payload); err != nil {
			return err
		}
		body = buf
	}
	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth := authorizationFromContext(ctx); auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		io.Copy(io.Discard, resp.Body)
		return errNotFound
	}
	if resp.StatusCode >= http.StatusBadRequest {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		if len(snippet) == 0 {
			return fmt.Errorf("upstream %s returned %s", target, resp.Status)
		}
		return fmt.Errorf("upstream %s returned %s: %s", target, resp.Status, strings.TrimSpace(string(snippet)))
	}
	if out == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (r *Resolver) doGet(ctx context.Context, target string, out any) error {
	return r.doJSON(ctx, http.MethodGet, target, nil, out)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func parseRFC3339(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

type contextKey string

const headersContextKey contextKey = "graphqlHeaders"

// WithRequestHeaders attaches HTTP headers to context for downstream authorization.
func WithRequestHeaders(ctx context.Context, headers http.Header) context.Context {
	if headers == nil {
		return ctx
	}
	return context.WithValue(ctx, headersContextKey, headers.Clone())
}

func authorizationFromContext(ctx context.Context) string {
	headers, _ := ctx.Value(headersContextKey).(http.Header)
	if headers == nil {
		return ""
	}
	return headers.Get("Authorization")
}

func facilityResult(items []facilityDTO) []*model.Facility {
	res := make([]*model.Facility, 0, len(items))
	for _, item := range items {
		copy := item
		res = append(res, facilityMap(copy))
	}
	return res
}

func facilityMap(item facilityDTO) *model.Facility {
	return &model.Facility{
		ID:          item.ID,
		Name:        stringPtr(item.Name),
		City:        stringPtr(item.City),
		Address:     stringPtr(item.Address),
		Type:        stringPtr(item.Type),
		Amenities:   append([]string{}, item.Amenities...),
		Description: stringPtr(item.Description),
		Maintenance: maintenanceResult(item.Maintenance),
		CreatedAt:   stringPtr(formatTime(item.CreatedAt)),
		UpdatedAt:   stringPtr(formatTime(item.UpdatedAt)),
	}
}

func maintenanceResult(items []maintenanceDTO) []*model.Maintenance {
	res := make([]*model.Maintenance, 0, len(items))
	for _, m := range items {
		copy := m
		res = append(res, &model.Maintenance{
			Start: formatTime(copy.Start),
			End:   formatTime(copy.End),
			Note:  stringPtr(copy.Note),
		})
	}
	return res
}

func availabilityResult(items []availabilityDTO) []*model.AvailabilitySlot {
	res := make([]*model.AvailabilitySlot, 0, len(items))
	for _, slot := range items {
		copy := slot
		res = append(res, &model.AvailabilitySlot{
			Start: formatTime(copy.Start),
			End:   formatTime(copy.End),
		})
	}
	return res
}

func bookingResult(items []bookingDTO) []*model.Booking {
	res := make([]*model.Booking, 0, len(items))
	for _, b := range items {
		copy := b
		res = append(res, bookingMap(copy))
	}
	return res
}

func bookingMap(item bookingDTO) *model.Booking {
	return &model.Booking{
		ID:           item.ID,
		FacilityID:   stringPtr(item.FacilityID),
		UserID:       stringPtr(item.UserID),
		Status:       stringPtr(item.Status),
		StartsAt:     stringPtr(formatTime(item.StartsAt)),
		EndsAt:       stringPtr(formatTime(item.EndsAt)),
		Participants: intPtr(item.Participants),
		Price:        floatPtr(item.Price),
		Notes:        stringPtr(item.Notes),
		CreatedAt:    stringPtr(formatTime(item.CreatedAt)),
		UpdatedAt:    stringPtr(formatTime(item.UpdatedAt)),
	}
}

func userMap(item userDTO) *model.User {
	return &model.User{
		ID:        item.ID,
		Email:     stringPtr(item.Email),
		FullName:  stringPtr(item.FullName),
		Role:      stringPtr(item.Role),
		Phone:     stringPtr(item.Phone),
		CreatedAt: stringPtr(formatTime(item.CreatedAt)),
		UpdatedAt: stringPtr(formatTime(item.UpdatedAt)),
	}
}

func maintenanceFromInput(inputs []*model.MaintenanceInput) ([]maintenanceDTO, error) {
	if len(inputs) == 0 {
		return nil, nil
	}
	res := make([]maintenanceDTO, 0, len(inputs))
	for _, in := range inputs {
		start, err := parseRFC3339(in.Start)
		if err != nil {
			return nil, fmt.Errorf("invalid maintenance start: %w", err)
		}
		end, err := parseRFC3339(in.End)
		if err != nil {
			return nil, fmt.Errorf("invalid maintenance end: %w", err)
		}
		note := ""
		if in.Note != nil {
			note = *in.Note
		}
		res = append(res, maintenanceDTO{Start: start, End: end, Note: note})
	}
	return res, nil
}

func availabilityFromInput(inputs []*model.AvailabilitySlotInput) ([]availabilityDTO, error) {
	if len(inputs) == 0 {
		return []availabilityDTO{}, nil
	}
	res := make([]availabilityDTO, 0, len(inputs))
	for _, slot := range inputs {
		start, err := parseRFC3339(slot.Start)
		if err != nil {
			return nil, fmt.Errorf("invalid slot start: %w", err)
		}
		end, err := parseRFC3339(slot.End)
		if err != nil {
			return nil, fmt.Errorf("invalid slot end: %w", err)
		}
		res = append(res, availabilityDTO{Start: start, End: end})
	}
	return res, nil
}

func timePtrFromString(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	parsed, err := parseRFC3339(*value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func stringPtr(v string) *string {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}

func stringVal(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
