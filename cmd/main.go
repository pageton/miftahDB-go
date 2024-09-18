package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pageton/miftahDB-go/db"
)

func main() {
	database, err := db.NewBaseMiftahDB(":memory:")
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer database.Close()

	err = database.Set("user:1", "John Doe", nil)
	if err != nil {
		log.Fatalf("Error setting value: %v", err)
	}
	fmt.Println("Value inserted.")

	entry, err := database.Get("user:1")
	if err != nil {
		log.Fatalf("Error getting value: %v", err)
	} else if entry != nil {
		fmt.Printf("Retrieved Entry: Key=%s, Value=%v\n", entry.Key, entry.Value)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	err = database.Set("user:2", "Jane Doe", &expiresAt)
	if err != nil {
		log.Fatalf("Error setting value with expiration: %v", err)
	}
	fmt.Println("Value with expiration inserted.")

	exists := database.Exists("user:2")
	if exists {
		fmt.Println("Key 'user:2' exists.")
	} else {
		fmt.Println("Key 'user:2' does not exist.")
	}

	err = database.Delete("user:1")
	if err != nil {
		log.Fatalf("Error deleting value: %v", err)
	}
	fmt.Println("Value deleted.")

	err = database.Cleanup()
	if err != nil {
		log.Fatalf("Error during cleanup: %v", err)
	}
	fmt.Println("Cleanup completed.")

	err = database.Backup("backup.db")
	if err != nil {
		log.Fatalf("Error creating backup: %v", err)
	}
	fmt.Println("Backup completed.")
}
