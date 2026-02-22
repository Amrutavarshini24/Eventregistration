// Package models defines GORM database models.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a system user (organizer or attendee).
type User struct {
	ID           string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(100);not null" json:"name"`
	Email        string    `gorm:"type:varchar(150);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	Role         string    `gorm:"type:varchar(20);default:'attendee'" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Events        []Event        `gorm:"foreignKey:OrganizerID" json:"-"`
	Registrations []Registration `gorm:"foreignKey:UserID" json:"-"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// Event represents a ticketed event created by an organizer.
type Event struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Title       string    `gorm:"type:varchar(200);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Capacity    int       `gorm:"not null;check:capacity > 0" json:"capacity"`
	Registered  int       `gorm:"default:0" json:"registered"`
	EventDate   time.Time `gorm:"not null" json:"event_date"`
	OrganizerID string    `gorm:"type:varchar(36);not null" json:"organizer_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Organizer     User           `gorm:"foreignKey:OrganizerID" json:"organizer,omitempty"`
	Registrations []Registration `gorm:"foreignKey:EventID" json:"-"`
}

func (e *Event) BeforeCreate(_ *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

func (e *Event) AvailableSeats() int { return e.Capacity - e.Registered }

// RegistrationStatus enumerates booking states.
type RegistrationStatus string

const (
	StatusConfirmed RegistrationStatus = "confirmed"
	StatusCancelled RegistrationStatus = "cancelled"
)

// Registration links a User to an Event.
type Registration struct {
	ID        string             `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID    string             `gorm:"type:varchar(36);not null;uniqueIndex:idx_user_event_status" json:"user_id"`
	EventID   string             `gorm:"type:varchar(36);not null;uniqueIndex:idx_user_event_status" json:"event_id"`
	Status    RegistrationStatus `gorm:"type:varchar(20);default:'confirmed';uniqueIndex:idx_user_event_status" json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`

	User  User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Event Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
}

func (r *Registration) BeforeCreate(_ *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}
