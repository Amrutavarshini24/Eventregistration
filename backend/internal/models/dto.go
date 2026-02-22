package models

// ── Auth DTOs ─────────────────────────────────────────

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"omitempty,oneof=organizer attendee"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// ── Event DTOs ────────────────────────────────────────

type CreateEventRequest struct {
	Title       string `json:"title" binding:"required,min=3,max=200"`
	Description string `json:"description"`
	Capacity    int    `json:"capacity" binding:"required,min=1"`
	EventDate   string `json:"event_date" binding:"required"` // RFC3339
}

type EventResponse struct {
	*Event
	AvailableSeats int `json:"available_seats"`
}

// ── Booking DTOs ──────────────────────────────────────

type BookingResponse struct {
	Message      string        `json:"message"`
	Registration *Registration `json:"registration,omitempty"`
}
