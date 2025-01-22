package main

import (
	"fmt"
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
			categories JSON,
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

}
