package db

import (
	"database/sql"
	"log"
	"time"
)

var DB *sql.DB

func OpenDB() error {
	var err error
	DB, err = sql.Open("sqlite3", "data/forum.db")
	if err != nil {
		return err
	}

	// Ensure the database is accessible
	if err = DB.Ping(); err != nil {
		return err
	}
	return nil
}

// removeExpiredSessions deletes all expired sessions, runs with dataCleanup()
func RemoveExpiredSessions() {
	_, err := DB.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	if err != nil {
		log.Printf("Error deleting expired sessions: %v\n", err.Error())
	}
}

// RemoveUnusedCategories deletes unused categories, runs with dataCleanup()
func RemoveUnusedCategories() {
	delUnusedCatsQuery := `DELETE FROM categories WHERE id NOT IN (SELECT DISTINCT category_id	FROM posts_categories);`
	_, err := DB.Exec(delUnusedCatsQuery)
	if err != nil {
		log.Printf("Error deleting unused categories: %v\n", err.Error())
	}
}

// dataCleanup removes expired sessions or unused categories every given time interval
func DataCleanup(interval time.Duration, f func(), name string) {
	ticker := time.NewTicker(interval)
	f() // run cleanup at the start
	go func() {
		for range ticker.C {
			log.Println("Running", name, "cleanup...")
			f()
		}
	}()
}
