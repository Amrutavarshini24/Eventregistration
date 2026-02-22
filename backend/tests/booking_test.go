// Package tests — concurrent booking stress tests for the backend.
//
// Run from the backend/ directory:
//
//	go test ./tests/... -v -race -run TestConcurrentBooking
package tests

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Amrutavarshini24/Eventregistration/internal/database"
	"github.com/Amrutavarshini24/Eventregistration/internal/models"
	"github.com/Amrutavarshini24/Eventregistration/internal/repositories"
	"github.com/Amrutavarshini24/Eventregistration/internal/services"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

func createTestUser(t *testing.T, db *gorm.DB, index int) string {
	t.Helper()
	u := &models.User{
		Name: fmt.Sprintf("User %d", index), Email: fmt.Sprintf("user%d@test.com", index),
		PasswordHash: "hashed", Role: "attendee",
	}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("createTestUser(%d): %v", index, err)
	}
	return u.ID
}

func createTestEvent(t *testing.T, db *gorm.DB, organizerID string, capacity int) string {
	t.Helper()
	ev := &models.Event{
		Title: fmt.Sprintf("Test Event (cap=%d)", capacity), Capacity: capacity,
		EventDate: time.Now().Add(24 * time.Hour), OrganizerID: organizerID,
	}
	if err := db.Create(ev).Error; err != nil {
		t.Fatalf("createTestEvent: %v", err)
	}
	return ev.ID
}

// TestConcurrentBooking spawns 50 goroutines racing for 5 seats.
// Asserts: exactly 5 succeed, 45 fail, no data races.
func TestConcurrentBooking(t *testing.T) {
	const (
		totalUsers = 50
		capacity   = 5
	)
	db := setupTestDB(t)

	eventRepo := repositories.NewEventRepository(db)
	regRepo   := repositories.NewRegistrationRepository(db)
	svc       := services.NewBookingService(db, regRepo, eventRepo)

	org := &models.User{Name: "Organizer", Email: "org@test.com", PasswordHash: "h", Role: "organizer"}
	db.Create(org)

	userIDs := make([]string, totalUsers)
	for i := range userIDs {
		userIDs[i] = createTestUser(t, db, i)
	}
	eventID := createTestEvent(t, db, org.ID, capacity)

	var (
		wg         sync.WaitGroup
		successCnt int64
		failCnt    int64
		startGun   = make(chan struct{})
	)
	results := make([]string, totalUsers)

	for i := 0; i < totalUsers; i++ {
		wg.Add(1)
		go func(idx int, uid string) {
			defer wg.Done()
			<-startGun
			log.Printf("USER %s attempted booking for event %s", uid, eventID)
			_, err := svc.Book(uid, eventID)
			if err == nil {
				atomic.AddInt64(&successCnt, 1)
				results[idx] = fmt.Sprintf("✅ User %d — SEAT RESERVED SUCCESSFULLY", idx)
				log.Printf("SEAT RESERVED SUCCESSFULLY | user=%s", uid)
			} else {
				atomic.AddInt64(&failCnt, 1)
				results[idx] = fmt.Sprintf("❌ User %d — BOOKING FAILED (%v)", idx, err)
				log.Printf("BOOKING FAILED — FULL | user=%s", uid)
			}
		}(i, userIDs[i])
	}
	close(startGun)
	wg.Wait()

	t.Log("\n─── Booking Results ───")
	for _, r := range results {
		t.Log(r)
	}
	t.Logf("Summary: %d succeeded, %d failed", successCnt, failCnt)

	if successCnt != capacity {
		t.Errorf("want %d successes, got %d", capacity, successCnt)
	}
	if failCnt != totalUsers-capacity {
		t.Errorf("want %d failures, got %d", totalUsers-capacity, failCnt)
	}

	var ev models.Event
	db.First(&ev, "id = ?", eventID)
	if ev.Registered != capacity {
		t.Errorf("event.Registered = %d, want %d", ev.Registered, capacity)
	}
	t.Logf("✅ No overbooking: event.Registered == %d", ev.Registered)
}

// TestNoDuplicateBooking verifies the same user cannot book twice.
func TestNoDuplicateBooking(t *testing.T) {
	db := setupTestDB(t)
	eventRepo := repositories.NewEventRepository(db)
	regRepo   := repositories.NewRegistrationRepository(db)
	svc       := services.NewBookingService(db, regRepo, eventRepo)

	org := &models.User{Name: "Org", Email: "org2@t.com", PasswordHash: "h", Role: "organizer"}
	db.Create(org)
	att := &models.User{Name: "A", Email: "a@t.com", PasswordHash: "h", Role: "attendee"}
	db.Create(att)
	ev := &models.Event{Title: "Dup Test", Capacity: 10, EventDate: time.Now().Add(time.Hour), OrganizerID: org.ID}
	db.Create(ev)

	if _, err := svc.Book(att.ID, ev.ID); err != nil {
		t.Fatalf("first booking failed: %v", err)
	}
	if _, err := svc.Book(att.ID, ev.ID); err == nil {
		t.Fatal("duplicate booking was not rejected")
	} else {
		t.Logf("✅ Duplicate rejected: %v", err)
	}
}
