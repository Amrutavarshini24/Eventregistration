package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Amrutavarshini24/Eventregistration/internal/models"
	"github.com/Amrutavarshini24/Eventregistration/internal/services"
)

type AuthHandler struct{ svc services.AuthService }

func NewAuthHandler(s services.AuthService) *AuthHandler { return &AuthHandler{svc: s} }

// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.svc.Register(&req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	// Auto-login: generate token
	token, _, _ := h.svc.Login(&models.LoginRequest{Email: req.Email, Password: req.Password})
	c.JSON(http.StatusCreated, models.AuthResponse{Token: token, User: user})
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, user, err := h.svc.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.AuthResponse{Token: token, User: user})
}
