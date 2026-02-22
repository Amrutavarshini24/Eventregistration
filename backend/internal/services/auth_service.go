package services

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/Amrutavarshini24/Eventregistration/internal/models"
	"github.com/Amrutavarshini24/Eventregistration/internal/repositories"
)

type AuthService interface {
	Register(req *models.RegisterRequest) (*models.User, error)
	Login(req *models.LoginRequest) (string, *models.User, error)
}

type authService struct{ userRepo repositories.UserRepository }

func NewAuthService(r repositories.UserRepository) AuthService { return &authService{userRepo: r} }

func (s *authService) Register(req *models.RegisterRequest) (*models.User, error) {
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already registered")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("authSvc.Register lookup: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("authSvc.Register hash: %w", err)
	}

	role := req.Role
	if role == "" {
		role = "attendee"
	}
	user := &models.User{Name: req.Name, Email: req.Email, PasswordHash: string(hash), Role: role}
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("authSvc.Register create: %w", err)
	}
	return user, nil
}

func (s *authService) Login(req *models.LoginRequest) (string, *models.User, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return "", nil, errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", nil, errors.New("invalid email or password")
	}
	token, err := signJWT(user)
	if err != nil {
		return "", nil, fmt.Errorf("authSvc.Login jwt: %w", err)
	}
	return token, user, nil
}

func signJWT(user *models.User) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev_secret_please_change"
	}
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
