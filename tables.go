package main

import (
	"fmt"
)

func makeTables() {
	// Create threads table if it doesn't exist
	createThreadsTableQuery := `
	CREATE TABLE IF NOT EXISTS threads (
		id INTEGER PRIMARY KEY AUTOINCREMENT,		
		author TEXT NOT NULL,
		authorID INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		categories JSON
	);`
	if _, err := db.Exec(createThreadsTableQuery); err != nil {
		fmt.Println("Error creating threads table:", err)
		return
	}

	// Create replies table if it doesn't exist
	createRepliesTableQuery := `
		CREATE TABLE IF NOT EXISTS replies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			base_id INTEGER DEFAULT 0,
			author TEXT NOT NULL,
			authorID INTEGER NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			parent_id INTEGER NOT NULL,
			parent_type TEXT
		);`
	if _, err := db.Exec(createRepliesTableQuery); err != nil {
		fmt.Println("Error creating replies table:", err)
		return
	}

	// Create users table if it doesn't exist
	createUsersTableQuery := `
		CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
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
	creatSessionsTableQuery := `
		CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		username TEXT,
		session_token TEXT UNIQUE NOT NULL,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
	);`
	if _, err := db.Exec(creatSessionsTableQuery); err != nil {
		fmt.Println("Error creating sessions table:", err)
		return
	}

	// Create reactions table if it doesn't exist
	createReactionsTableQuery := `
	CREATE TABLE IF NOT EXISTS post_reactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,          -- User who reacted
		post_id INTEGER NOT NULL,          -- ID of the thread or reply
		post_type TEXT NOT NULL,           -- 'thread' or 'reply'
		reaction_type TEXT NOT NULL,       -- 'like' or 'dislike'
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE (user_id, post_id, post_type)  -- Prevents duplicate reactions for the same post type (no simultaneous like and dislike)
	);`
	if _, err := db.Exec(createReactionsTableQuery); err != nil {
		fmt.Println("Error creating reactions table:", err)
		return
	}

}
