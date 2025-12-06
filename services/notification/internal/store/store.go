package store

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Notification описывает отправленное сообщение.
type Notification struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	Recipient string    `json:"recipient"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Store хранит уведомления в памяти.
type Store struct {
	mu   sync.RWMutex
	data map[string]*Notification
}

func New() *Store {
	return &Store{data: make(map[string]*Notification)}
}

func (s *Store) List(channel string) []Notification {
	result := make([]Notification, 0)
	s.mu.RLock()
	for _, item := range s.data {
		if channel != "" && item.Channel != channel {
			continue
		}
		result = append(result, *item)
	}
	s.mu.RUnlock()
	return result
}

func (s *Store) Create(channel, recipient, subject, body string) *Notification {
	notification := &Notification{
		ID:        uuid.NewString(),
		Channel:   channel,
		Recipient: recipient,
		Subject:   subject,
		Body:      body,
		Status:    "delivered",
		CreatedAt: time.Now(),
	}
	s.mu.Lock()
	s.data[notification.ID] = notification
	s.mu.Unlock()
	return notification
}

func (s *Store) Get(id string) (*Notification, bool) {
	s.mu.RLock()
	notification, ok := s.data[id]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	copy := *notification
	return &copy, true
}
