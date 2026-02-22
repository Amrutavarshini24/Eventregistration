package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Amrutavarshini24/Eventregistration/internal/middleware"
	"github.com/Amrutavarshini24/Eventregistration/internal/services"
)

type BookingHandler struct{ svc services.BookingService }

func NewBookingHandler(s services.BookingService) *BookingHandler { return &BookingHandler{svc: s} }

// POST /api/events/:id/register
func (h *BookingHandler) BookEvent(c *gin.Context) {
	uid, _ := c.Get(middleware.ContextKeyUserID)
	reg, err := h.svc.Book(uid.(string), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEventFull):
			c.JSON(http.StatusConflict, gin.H{"error": "event is fully booked"})
		case errors.Is(err, services.ErrDuplicateBooking):
			c.JSON(http.StatusConflict, gin.H{"error": "you have already registered for this event"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Seat reserved successfully", "registration": reg})
}

// GET /api/events/:id/registrations
func (h *BookingHandler) GetEventRegistrations(c *gin.Context) {
	regs, err := h.svc.GetEventRegistrations(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"registrations": regs, "count": len(regs)})
}

// GET /api/me/registrations
func (h *BookingHandler) GetMyRegistrations(c *gin.Context) {
	uid, _ := c.Get(middleware.ContextKeyUserID)
	regs, err := h.svc.GetUserRegistrations(uid.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"registrations": regs, "count": len(regs)})
}
