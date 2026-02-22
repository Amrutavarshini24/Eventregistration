package repositories

import (
	"fmt"
	"gorm.io/gorm"
	"github.com/Amrutavarshini24/Eventregistration/internal/models"
)

type EventRepository interface {
	Create(event *models.Event) error
	FindByID(id string) (*models.Event, error)
	List() ([]models.Event, error)
	// IncrementRegistered claims one seat atomically inside tx.
	// Returns (event, true) on success, (event, false) when full.
	IncrementRegistered(tx *gorm.DB, eventID string) (*models.Event, bool, error)
}

type eventRepository struct{ db *gorm.DB }

func NewEventRepository(db *gorm.DB) EventRepository { return &eventRepository{db: db} }

func (r *eventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) FindByID(id string) (*models.Event, error) {
	var e models.Event
	if err := r.db.Preload("Organizer").First(&e, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("eventRepo.FindByID: %w", err)
	}
	return &e, nil
}

func (r *eventRepository) List() ([]models.Event, error) {
	var evs []models.Event
	if err := r.db.Preload("Organizer").Order("event_date asc").Find(&evs).Error; err != nil {
		return nil, fmt.Errorf("eventRepo.List: %w", err)
	}
	return evs, nil
}

// IncrementRegistered — Layer 3 of the concurrency defence.
// Uses SELECT FOR UPDATE (Postgres) + conditional UPDATE to guarantee
// no overbooking even across multiple server nodes.
func (r *eventRepository) IncrementRegistered(tx *gorm.DB, eventID string) (*models.Event, bool, error) {
	var e models.Event
	// Lock the row for the transaction duration (Postgres).
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&e, "id = ?", eventID).Error; err != nil {
		return nil, false, fmt.Errorf("eventRepo.IncrementRegistered lock: %w", err)
	}
	if e.Registered >= e.Capacity {
		return &e, false, nil
	}
	res := tx.Model(&models.Event{}).
		Where("id = ? AND registered < capacity", eventID).
		UpdateColumn("registered", gorm.Expr("registered + ?", 1))
	if res.Error != nil {
		return nil, false, fmt.Errorf("eventRepo.IncrementRegistered update: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return &e, false, nil // race — another tx won
	}
	e.Registered++
	return &e, true, nil
}
