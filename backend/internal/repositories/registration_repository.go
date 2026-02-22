package repositories

import (
	"fmt"
	"gorm.io/gorm"
	"github.com/Amrutavarshini24/Eventregistration/internal/models"
)

type RegistrationRepository interface {
	Create(tx *gorm.DB, reg *models.Registration) error
	FindByUserAndEvent(userID, eventID string) (*models.Registration, error)
	FindByEvent(eventID string) ([]models.Registration, error)
	FindByUser(userID string) ([]models.Registration, error)
}

type registrationRepository struct{ db *gorm.DB }

func NewRegistrationRepository(db *gorm.DB) RegistrationRepository {
	return &registrationRepository{db: db}
}

func (r *registrationRepository) Create(tx *gorm.DB, reg *models.Registration) error {
	return tx.Create(reg).Error
}

func (r *registrationRepository) FindByUserAndEvent(userID, eventID string) (*models.Registration, error) {
	var reg models.Registration
	err := r.db.Where("user_id = ? AND event_id = ? AND status = ?",
		userID, eventID, models.StatusConfirmed).First(&reg).Error
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

func (r *registrationRepository) FindByEvent(eventID string) ([]models.Registration, error) {
	var regs []models.Registration
	err := r.db.Preload("User").
		Where("event_id = ? AND status = ?", eventID, models.StatusConfirmed).
		Find(&regs).Error
	if err != nil {
		return nil, fmt.Errorf("regRepo.FindByEvent: %w", err)
	}
	return regs, nil
}

func (r *registrationRepository) FindByUser(userID string) ([]models.Registration, error) {
	var regs []models.Registration
	err := r.db.Preload("Event").Preload("Event.Organizer").
		Where("user_id = ? AND status = ?", userID, models.StatusConfirmed).
		Find(&regs).Error
	if err != nil {
		return nil, fmt.Errorf("regRepo.FindByUser: %w", err)
	}
	return regs, nil
}
