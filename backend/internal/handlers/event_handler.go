package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Amrutavarshini24/Eventregistration/internal/middleware"
	"github.com/Amrutavarshini24/Eventregistration/internal/models"
	"github.com/Amrutavarshini24/Eventregistration/internal/services"
)

type EventHandler struct{ svc services.EventService }

func NewEventHandler(s services.EventService) *EventHandler { return &EventHandler{svc: s} }

// POST /api/events  (organizer)
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req models.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, _ := c.Get(middleware.ContextKeyUserID)
	ev, err := h.svc.CreateEvent(&req, id.(string))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ev)
}

// GET /api/events
func (h *EventHandler) ListEvents(c *gin.Context) {
	evs, err := h.svc.ListEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": evs, "count": len(evs)})
}

// GET /api/events/:id
func (h *EventHandler) GetEvent(c *gin.Context) {
	ev, err := h.svc.GetEvent(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}
	c.JSON(http.StatusOK, ev)
}
