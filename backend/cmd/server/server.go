// Package server wires all backend dependencies and starts the Gin HTTP server.
// This is a PURE REST API server — it serves only JSON, no static HTML.
// The frontend (frontend/) is a completely separate project.
package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/Amrutavarshini24/Eventregistration/internal/handlers"
	"github.com/Amrutavarshini24/Eventregistration/internal/middleware"
	"github.com/Amrutavarshini24/Eventregistration/internal/repositories"
	"github.com/Amrutavarshini24/Eventregistration/internal/services"
)

type Server struct {
	engine *gin.Engine
	port   string
}

func New(db *gorm.DB) *Server {
	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo  := repositories.NewUserRepository(db)
	eventRepo := repositories.NewEventRepository(db)
	regRepo   := repositories.NewRegistrationRepository(db)

	// ── Services ─────────────────────────────────────────────────────────────
	authSvc    := services.NewAuthService(userRepo)
	eventSvc   := services.NewEventService(eventRepo)
	bookingSvc := services.NewBookingService(db, regRepo, eventRepo)

	// ── Handlers ─────────────────────────────────────────────────────────────
	authH    := handlers.NewAuthHandler(authSvc)
	eventH   := handlers.NewEventHandler(eventSvc)
	bookingH := handlers.NewBookingHandler(bookingSvc)

	// ── Gin engine ───────────────────────────────────────────────────────────
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// ── CORS middleware ───────────────────────────────────────────────────────
	// Reads CORS_ORIGINS from .env (comma-separated).
	// Use * for development; set exact frontend URL in production.
	allowedOrigins := os.Getenv("CORS_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}
	engine.Use(corsMiddleware(allowedOrigins))

	// ── Health ────────────────────────────────────────────────────────────────
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "event-ticketing-backend"})
	})

	// ── API routes (all under /api) ───────────────────────────────────────────
	api := engine.Group("/api")

	// Auth (public)
	auth := api.Group("/auth")
	auth.POST("/register", authH.Register)
	auth.POST("/login",    authH.Login)

	// Events
	evts := api.Group("/events")
	evts.GET("",     eventH.ListEvents)
	evts.GET("/:id", eventH.GetEvent)
	evts.POST("",
		middleware.AuthRequired(),
		middleware.OrganizerRequired(),
		eventH.CreateEvent,
	)
	evts.POST("/:id/register",
		middleware.AuthRequired(),
		bookingH.BookEvent,
	)
	evts.GET("/:id/registrations",
		middleware.AuthRequired(),
		bookingH.GetEventRegistrations,
	)

	// Me
	me := api.Group("/me", middleware.AuthRequired())
	me.GET("/registrations", bookingH.GetMyRegistrations)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	return &Server{engine: engine, port: port}
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%s", s.port)
	return s.engine.Run(addr)
}

// corsMiddleware adds CORS headers for cross-origin requests from the frontend.
func corsMiddleware(origins string) gin.HandlerFunc {
	allowed := strings.Split(origins, ",")
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allow  := "*"
		for _, o := range allowed {
			o = strings.TrimSpace(o)
			if o == "*" || o == origin {
				allow = o
				break
			}
		}
		c.Header("Access-Control-Allow-Origin",  allow)
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
