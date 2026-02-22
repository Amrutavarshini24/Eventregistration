package services

import (
	"fmt"
	"time"

	"github.com/Amrutavarshini24/Eventregistration/internal/models"
	"github.com/Amrutavarshini24/Eventregistration/internal/repositories"
)

type EventService interface {
	CreateEvent(req *models.CreateEventRequest, organizerID string) (*models.EventResponse, error)
	GetEvent(id string) (*models.EventResponse, error)
	ListEvents() ([]models.EventResponse, error)
}

type eventService struct{ eventRepo repositories.EventRepository }

func NewEventService(r repositories.EventRepository) EventService { return &eventService{eventRepo: r} }

func (s *eventService) CreateEvent(req *models.CreateEventRequest, organizerID string) (*models.EventResponse, error) {
	date, err := time.Parse(time.RFC3339, req.EventDate)
	if err != nil {
		return nil, fmt.Errorf("invalid event_date (use RFC3339 e.g. 2025-12-31T18:00:00Z): %w", err)
	}
	ev := &models.Event{
		Title: req.Title, Description: req.Description,
		Capacity: req.Capacity, EventDate: date, OrganizerID: organizerID,
	}
	if err := s.eventRepo.Create(ev); err != nil {
		return nil, err
	}
	return toEventResponse(ev), nil
}

func (s *eventService) GetEvent(id string) (*models.EventResponse, error) {
	ev, err := s.eventRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return toEventResponse(ev), nil
}

func (s *eventService) ListEvents() ([]models.EventResponse, error) {
	evs, err := s.eventRepo.List()
	if err != nil {
		return nil, err
	}
	resp := make([]models.EventResponse, len(evs))
	for i := range evs {
		resp[i] = *toEventResponse(&evs[i])
	}
	return resp, nil
}

func toEventResponse(e *models.Event) *models.EventResponse {
	return &models.EventResponse{Event: e, AvailableSeats: e.AvailableSeats()}
}
