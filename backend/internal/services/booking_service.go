// Package services — BookingService: the concurrency-safe seat reservation core.
//
// Three-layer defence against overbooking:
//  1. Per-event sync.Mutex  — serialises same-node goroutines for the same event.
//  2. DB transaction        — atomic read+write, auto-rollback on error.
//  3. Conditional UPDATE    — WHERE registered < capacity is the ultimate net
//                             (works even in multi-node / distributed deployments).
package services

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"

	"github.com/Amrutavarshini24/Eventregistration/internal/models"
	"github.com/Amrutavarshini24/Eventregistration/internal/repositories"
)

var (
	ErrEventFull      = errors.New("event is fully booked")
	ErrDuplicateBooking = errors.New("user has already registered for this event")
)

type BookingService interface {
	Book(userID, eventID string) (*models.Registration, error)
	GetEventRegistrations(eventID string) ([]models.Registration, error)
	GetUserRegistrations(userID string) ([]models.Registration, error)
}

type bookingService struct {
	db         *gorm.DB
	regRepo    repositories.RegistrationRepository
	evtRepo    repositories.EventRepository
	eventLocks sync.Map // eventID → *sync.Mutex
}

func NewBookingService(db *gorm.DB, r repositories.RegistrationRepository, e repositories.EventRepository) BookingService {
	return &bookingService{db: db, regRepo: r, evtRepo: e}
}

func (s *bookingService) mu(eventID string) *sync.Mutex {
	mu, _ := s.eventLocks.LoadOrStore(eventID, &sync.Mutex{})
	return mu.(*sync.Mutex)
}

// Book reserves a seat for userID in eventID.
func (s *bookingService) Book(userID, eventID string) (*models.Registration, error) {
	// ── Layer 1: per-event mutex ──────────────────────────────────────────────
	mu := s.mu(eventID)
	mu.Lock()
	defer mu.Unlock()

	log.Printf("USER %s attempted booking for event %s", userID, eventID)

	// ── Duplicate guard ───────────────────────────────────────────────────────
	if existing, err := s.regRepo.FindByUserAndEvent(userID, eventID); err == nil && existing != nil {
		log.Printf("BOOKING FAILED — DUPLICATE | user=%s event=%s", userID, eventID)
		return nil, ErrDuplicateBooking
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("bookingSvc.Book lookup: %w", err)
	}

	// ── Layers 2 & 3: transaction + conditional UPDATE ────────────────────────
	var reg *models.Registration
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		_, ok, err := s.evtRepo.IncrementRegistered(tx, eventID)
		if err != nil {
			return err
		}
		if !ok {
			log.Printf("BOOKING FAILED — FULL | user=%s event=%s", userID, eventID)
			return ErrEventFull
		}
		reg = &models.Registration{UserID: userID, EventID: eventID, Status: models.StatusConfirmed}
		if err := s.regRepo.Create(tx, reg); err != nil {
			return err
		}
		log.Printf("SEAT RESERVED SUCCESSFULLY | user=%s event=%s reg=%s", userID, eventID, reg.ID)
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return reg, nil
}

func (s *bookingService) GetEventRegistrations(id string) ([]models.Registration, error) {
	return s.regRepo.FindByEvent(id)
}
func (s *bookingService) GetUserRegistrations(id string) ([]models.Registration, error) {
	return s.regRepo.FindByUser(id)
}
