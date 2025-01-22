package main

import (
	"net/http"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestValidateSession(t *testing.T) {
	Testinit()
	defer db.Close()

	// Add a test user and session to the database
	db.Exec("INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)", "testid", "test@example.com", "testuser", "testpass")
	sessionToken := "testtoken"
	expiresAt := time.Now().Add(30 * time.Minute)
	db.Exec("INSERT INTO sessions (user_id, username, session_token, expires_at) VALUES (?, ?, ?, ?)", "testid", "testuser", sessionToken, expiresAt)

	// Create a new HTTP request with the session cookie
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})

	// Validate the session
	userID, userName, valid := validateSession(req)
	if !valid {
		t.Errorf("session should be valid")
	}
	if userID != "testid" {
		t.Errorf("handler returned unexpected userID: got %v want %v", userID, "testid")
	}
	if userName != "testuser" {
		t.Errorf("handler returned unexpected userName: got %v want %v", userName, "testuser")
	}
}
