package main

import (
	"fmt"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID   string
	Name string
}

type Event struct {
	ID    string
	Title string
}

func main() {
	dbIdx := "event_ticketing_dev.db"
	if _, err := os.Stat(dbIdx); os.IsNotExist(err) {
		log.Fatalf("Database file %s not found in current directory", dbIdx)
	}

	db, err := gorm.Open(sqlite.Open(dbIdx), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("--- USERS ---")
	var users []User
	db.Raw("SELECT id, name FROM users").Scan(&users)
	for _, u := range users {
		fmt.Printf("ID: %s | Name: %s\n", u.ID, u.Name)
	}

	fmt.Println("\n--- EVENTS ---")
	var events []Event
	db.Raw("SELECT id, title FROM events").Scan(&events)
	for _, e := range events {
		fmt.Printf("ID: %s | Title: %s\n", e.ID, e.Title)
	}
}
