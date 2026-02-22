# DESIGN.md — Concurrency & Race Condition Prevention

## Overview

The core challenge of a ticketing system is preventing **overbooking** when multiple users race to claim the last available seats. This document explains exactly how this system solves that problem.

---

## The Problem: TOCTOU Race Condition

A naïve implementation would:

1. **Read** the current `registered` count.
2. **Check** if `registered < capacity`.
3. **Write** `registered = registered + 1`.

Between steps 1 and 3, another goroutine can execute the same sequence. Both goroutines read `registered = 4` (capacity 5), both decide there's a seat, and both write `registered = 5`. Two users get the last seat — that's a classic **Time-Of-Check-To-Time-Of-Use (TOCTOU)** race.

---

## Three-Layer Defence

### Layer 1 — Per-Event Application Mutex

**File:** `internal/services/booking_service.go`

```go
// sync.Map stores one *sync.Mutex per eventID lazily.
func (s *bookingService) getEventMutex(eventID string) *sync.Mutex {
    mu, _ := s.eventLocks.LoadOrStore(eventID, &sync.Mutex{})
    return mu.(*sync.Mutex)
}

func (s *bookingService) Book(userID, eventID string) (*models.Registration, error) {
    mu := s.getEventMutex(eventID)
    mu.Lock()
    defer mu.Unlock()
    // ... only 1 goroutine per event runs beyond this point
}
```

**Why `sync.Map` instead of a single `sync.RWMutex`?**

A single mutex would serialise all bookings across ALL events. A per-event mutex means events A, B, and C can all be booked concurrently — only concurrent bookings for the **same** event are serialised.

**Trade-off:** `sync.Map` has slightly more overhead than a plain map, but it is safe for concurrent reads/writes without external locking and never blocks across events.

---

### Layer 2 — ACID Database Transaction

**File:** `internal/services/booking_service.go`

```go
txErr := s.db.Transaction(func(tx *gorm.DB) error {
    _, ok, err := s.evtRepo.IncrementRegistered(tx, eventID)
    if !ok { return ErrEventFull }
    return s.regRepo.Create(tx, &Registration{...})
})
```

The transaction ensures:

- If creating the `Registration` record fails after the seat was claimed, the seat counter rolls back automatically.
- Both writes — the `registered` increment and the registration insert — either succeed together or fail together.

**Isolation level:** PostgreSQL's default `READ COMMITTED` isolation combined with `SELECT FOR UPDATE` (see Layer 3) is sufficient to prevent phantom reads on the `registered` column.

---

### Layer 3 — Optimistic Conditional UPDATE

**File:** `internal/repositories/event_repository.go`

```go
// Step 1: Lock the row (PostgreSQL SELECT … FOR UPDATE)
tx.Set("gorm:query_option", "FOR UPDATE").First(&event, "id = ?", eventID)

// Step 2: Conditional UPDATE — the database itself enforces capacity
result := tx.Model(&models.Event{}).
    Where("id = ? AND registered < capacity", eventID).
    UpdateColumn("registered", gorm.Expr("registered + ?", 1))

// Step 3: If RowsAffected == 0, another transaction won the race
if result.RowsAffected == 0 { return nil, false, nil }
```

This is the **ultimate safety net**. Even in a multi-node deployment where the application-level mutex (Layer 1) offers no cross-process protection, the database enforces the constraint atomically.

The `WHERE registered < capacity` predicate makes the UPDATE a **no-op** if the capacity was reached by another transaction between the SELECT and the UPDATE.

---

## Why This Works in Distributed Systems

| Scenario | Layer 1 (Mutex) | Layer 2 (Transaction) | Layer 3 (Conditional UPDATE) |
|---|---|---|---|
| Single node, many goroutines | ✅ Protects | ✅ Protects | ✅ Protects |
| Multi-node, load balanced | ❌ No cross-process mutex | ✅ Isolates per-node writes | ✅ **Guarantees correctness** |
| Network partition / crash | ❌ Mutex lost | ✅ Rollback on failure | ✅ UPDATE is atomic |

**In a distributed deployment**, Layer 3 is the definitive guarantee. The probabilistic rate of false conflicts (two processes hitting the UPDATE at the exact same millisecond) remains low because the DB serialises row-level writes natively.

For **even stronger distributed guarantees** consider:
- PostgreSQL Advisory Locks (`pg_try_advisory_xact_lock`) instead of the application mutex.
- Redis distributed lock (Redlock algorithm) as a cross-process mutex.

---

## Duplicate Booking Prevention

Before entering the transaction, the service checks:

```go
existing, err := s.regRepo.FindByUserAndEvent(userID, eventID)
if existing != nil { return nil, ErrDuplicateBooking }
```

A **unique index** on `(user_id, event_id, status)` in the database is an additional guard against duplicate rows being inserted by concurrent requests from the same user (e.g., double-click).

---

## Logging Protocol

As required by the project spec, the system emits three distinct log lines:

```
USER X attempted booking for event Y
SEAT RESERVED SUCCESSFULLY | user=X event=Y reg=Z
BOOKING FAILED — FULL | user=X event=Y
```

These appear in both the application logs (`log.Printf`) and the test output (`t.Log`).

---

## Data Model Summary

```
User ────────┐
             │ organizer
             ▼
           Event (capacity, registered)
             │
             │ 1:N
             ▼
       Registration (user_id, event_id, status)
             │
             └──── User (attendee)
```

The `registered` counter is denormalised on the `Event` row for O(1) capacity checks. It is always updated inside a transaction to remain consistent with the `Registration` count.
