package main

import (
	"fmt"
	"forum/cmd/router"
	"forum/internal/db"
	"forum/internal/templates"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	err := db.OpenDB() // Open database connection
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	defer db.DB.Close()
	db.MakeTables()
	db.DataCleanup(time.Hour, db.RemoveExpiredSessions, "session")     // Clean up sessions every hour
	db.DataCleanup(6*time.Hour, db.RemoveUnusedCategories, "category") // Clean up categories every 6 hours
	templates.InitTemplates()
	router.SetHandlers()

	// Start the server
	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil)) // Logs the error and exits.
}
