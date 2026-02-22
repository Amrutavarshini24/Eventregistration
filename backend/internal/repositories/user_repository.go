package repositories

import (
	"fmt"
	"gorm.io/gorm"
	"github.com/Amrutavarshini24/Eventregistration/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
}

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepository{db: db} }

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}
func (r *userRepository) FindByID(id string) (*models.User, error) {
	var u models.User
	if err := r.db.First(&u, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("userRepo.FindByID: %w", err)
	}
	return &u, nil
}
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var u models.User
	if err := r.db.First(&u, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("userRepo.FindByEmail: %w", err)
	}
	return &u, nil
}
