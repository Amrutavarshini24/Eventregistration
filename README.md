# Event Registration & Ticketing System API

## Project Overview
This project is a RESTful Event Registration and Ticketing System built using Go (Golang), Gin, GORM, and SQLite.

The system allows:
- Organizers to create and manage events
- Attendees to browse events and register for seats
- Real-time tracking of available capacity
- High-concurrency handling to prevent overbooking

This project was implemented as part of the Go Capstone project, specifically focusing on the challenge of managing multiple users trying to book the last remaining spots for an event simultaneously.

## Tech Stack
- Go 1.25: The fast and efficient language powering the app.
- Gin: Handles the web connections and API.
- GORM: Manages how data is saved and retrieved.
- SQLite: A simple, built-in database that needs no setup.
- UUID: Creates unique identities for every user and event.
- Concurrency Protection: Prevents overbooking even when many people click at once.

## How to Run the Project

You can run these commands directly in your **VS Code Terminal** (press `Ctrl + ` to open it):

1. Clone the repository
```bash
git clone https://github.com/Amrutavarshini24/Eventregistration.git
cd Eventregistration
```

2. Install dependencies
```bash
go mod tidy
```

3. Setup environment
Copy the example environment file:
```bash
cp .env.example .env
```

4. Run the server
```bash
go run main.go
```

Server runs at: http://localhost:8080

## Database
- SQLite database file: backend/event_ticketing_dev.db
- Tables are automatically created using GORM AutoMigrate.
- Database constraints (Unique Indices) ensure data integrity at the row level.

## API Endpoints

### Auth
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST   | /auth/register | Register a new user |
| POST   | /auth/login    | Login and receive JWT token |

### Events
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST   | /events | Create a new event (Organizer only) |
| GET    | /events | Get all listed events |
| GET    | /events/:id | Get detailed information for a single event |

### Booking
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST   | /events/:id/register | Book a seat for an event |
| GET    | /me/registrations   | View all bookings for the logged-in user |

## Concurrency Handling
The booking operation is handled using a three-layered defense system to prevent race conditions and overbooking:

1. Per-Event Mutex: Serializes booking requests for the same event at the application level.
2. Database Transactions: Ensures all operations (checking capacity, incrementing seats, and creating registration) succeed or fail together.
3. Conditional UPDATE: Uses a WHERE clause (registered < capacity) at the database level as the ultimate net to guarantee no overbooking occurs even in multi-node deployments.

## Testing
The system includes specialized stress tests to simulate high-concurrency environments:

- Concurrent Booking Test: Simulates 50 users racing for 5 seats to verify that no overbooking occurs.
- Race Detector: All tests are verified using the Go race detector to ensure thread safety.

## Key Features
- Transaction-safe booking
- Real-time capacity management
- Role-based security (Organizer / Attendee)
- Three-layer concurrency protection
- Persistent SQLite database
- RESTful design

## How AI was used
I used AI prompts as a guide to help build different parts of this project:
- Setting up the main project structure and folders.
- Writing the code that prevents two people from booking the same seat.
- Transforming the website design from one page to several separate pages.
- Making the website look modern with nice colors and smooth animations.
- Helping organize and write the descriptions in this README.

## Author
Amrutavarshini Beernalli
Go Capstone Project

