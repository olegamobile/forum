package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterHandler(t *testing.T) {
	// Setup: create db in mem, not on disk; db is temp and lost upon prog termination
	db, _ = sql.Open("sqlite3", ":memory:")
	defer db.Close()
	makeTables() // create same tables as our prog

	// Create a new HTTP request with mock user login data
	req, err := http.NewRequest("POST", "/register", strings.NewReader("username=testuser&email=test@example.com&password=testpass"))
	if err != nil {
		t.Fatal(err)
	}
	// Set "Content-Type" header of HTTP req to "application/x-www-form-urlencoded"
	// req body contains form data encoded as key-value pairs (standard encoding for HTML form submissions
	// This header is necessary for the server to correctly parse the form data sent in the request body.
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(registerHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Check if the user was added to the database
	var username string
	err = db.QueryRow("SELECT username FROM users WHERE username = ?", "testuser").Scan(&username)
	if err != nil {
		t.Errorf("user not found in database: %v", err)
	}
	if username != "testuser" {
		t.Errorf("handler returned unexpected body: got %v want %v", username, "testuser")
	}
}

func TestLogUserInHandler(t *testing.T) {
	// Setup
	db, _ = sql.Open("sqlite3", ":memory:")
	defer db.Close()
	makeTables()

	// Add a test user to the database
	hashPass, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	db.Exec("INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)", "testid", "test@example.com", "testuser", string(hashPass))

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "/loguserin", strings.NewReader("username=testuser&password=testpass&return_url=/"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(logUserInHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Check if the session was created
	var sessionToken string
	err = db.QueryRow("SELECT session_token FROM sessions WHERE user_id = ?", "testid").Scan(&sessionToken)
	if err != nil {
		t.Errorf("session not found in database: %v", err)
	}
	if sessionToken == "" {
		t.Errorf("handler did not create session")
	}
}

func TestValidateSession(t *testing.T) {
	// Setup
	db, _ = sql.Open("sqlite3", ":memory:")
	defer db.Close()
	makeTables()

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
