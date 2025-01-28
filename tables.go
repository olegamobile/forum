package main

import (
	"fmt"
	"log"
	"time"
)

func makeTables() {

	// Create posts table if it doesn't exist
	createPostsTableQuery := `
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			base_id INTEGER DEFAULT 0,
			author TEXT NOT NULL,
			authorID TEXT NOT NULL,
			title TEXT DEFAULT '',
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			parent_id INTEGER DEFAULT 0
		);`
	if _, err := db.Exec(createPostsTableQuery); err != nil {
		fmt.Println("Error creating posts table:", err)
		return
	}

	// Create users table if it doesn't exist
	createUsersTableQuery := `
		CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,  -- Hashed passwords
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createUsersTableQuery); err != nil {
		fmt.Println("Error creating users table:", err)
		return
	}

	// Create sessions table if it doesn't exist
	createSessionsTableQuery := `
		CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		username TEXT,
		session_token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	);`
	if _, err := db.Exec(createSessionsTableQuery); err != nil {
		fmt.Println("Error creating sessions table:", err)
		return
	}

	// Create reactions table if it doesn't exist
	createReactionsTableQuery := `
	CREATE TABLE IF NOT EXISTS post_reactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,          -- User who reacted
		post_id INTEGER NOT NULL,          -- ID of the thread or reply
		reaction_type TEXT NOT NULL,       -- 'like' or 'dislike'
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE (user_id, post_id)  -- Prevents duplicate reactions for the same post type (no simultaneous like and dislike)
	);`
	if _, err := db.Exec(createReactionsTableQuery); err != nil {
		fmt.Println("Error creating reactions table:", err)
		return
	}

	// Create categories table if it doesn't exist
	createCategoriesTableQuery := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE (name)  -- Prevents duplicate categories
	);`
	if _, err := db.Exec(createCategoriesTableQuery); err != nil {
		fmt.Println("Error creating reactions table:", err)
		return
	}

	// Create table to connect posts and categories if it doesn't exist
	createPostsCategoriesTableQuery := `
	CREATE TABLE IF NOT EXISTS posts_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (post_id) REFERENCES posts(id),
		FOREIGN KEY (category_id) REFERENCES categories(id),
		UNIQUE (post_id, category_id) 
	);`
	if _, err := db.Exec(createPostsCategoriesTableQuery); err != nil {
		fmt.Println("Error creating reactions table:", err)
		return
	}

	// Create images table if it doesn't exist
	createImagesTableQuery := `
CREATE TABLE IF NOT EXISTS images (
	id TEXT PRIMARY KEY,  -- includes file extension (like [UUID].jpg)
	post_id INTEGER DEFAULT NULL,  -- if NOT NULL it is a post image, 
	user_id INTEGER NOT NULL,
	original_name TEXT,
	file_size INT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (post_id) REFERENCES posts(id),
	FOREIGN KEY (user_id) REFERENCES users(id)
);`
	if _, err := db.Exec(createImagesTableQuery); err != nil {
		fmt.Println("Error creating reactions table:", err)
		return
	}

}

// removeExpiredSessions deletes all expired sessions, runs with dataCleanup()
func removeExpiredSessions() {
	_, err := db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	if err != nil {
		log.Printf("Error deleting expired sessions: %v\n", err.Error())
	}
}

// removeUnusedCategories deletes unused categories, runs with dataCleanup()
func removeUnusedCategories() {
	delUnusedCatsQuery := `DELETE FROM categories WHERE id NOT IN (SELECT DISTINCT category_id	FROM posts_categories);`
	_, err := db.Exec(delUnusedCatsQuery)
	if err != nil {
		log.Printf("Error deleting unused categories: %v\n", err.Error())
	}
}

// dataCleanup removes expired sessions or unused categories every given time interval
func dataCleanup(interval time.Duration, f func(), name string) {
	ticker := time.NewTicker(interval)
	f() // run cleanup at the start
	go func() {
		for range ticker.C {
			log.Println("Running", name, "cleanup...")
			f()
		}
	}()
}
