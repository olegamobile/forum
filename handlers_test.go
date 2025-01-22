package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func Testinit() {
	// Setup: create in-mem db; db is temp and lost upon prog termination
	db, _ = sql.Open("sqlite3", ":memory:")

	// Run our program
	makeTables()
	initTemplates()

	// Clear existing handlers to avoid duplicate route registration
	http.DefaultServeMux = new(http.ServeMux)
	setHandlers()
}

func TestIndexHandler(t *testing.T) {
	Testinit()
	defer db.Close()

	tests := []struct {
		name     string
		method   string
		url      string
		wantCode int
	}{
		// Test diff http methods
		{
			name:     "POST home / valid",
			method:   "POST",
			url:      "/",
			wantCode: http.StatusOK,
		},
		{
			name:     "GET home / valid",
			method:   "GET",
			url:      "/",
			wantCode: http.StatusOK,
		},
		{
			name:     "PUT home / not allowed",
			method:   "PUT",
			url:      "/",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "DELETE home / not allowed",
			method:   "DELETE",
			url:      "/",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "POST home /forum valid",
			method:   "POST",
			url:      "/forum",
			wantCode: http.StatusOK,
		},
		{
			name:     "GET home /forum valid",
			method:   "GET",
			url:      "/forum",
			wantCode: http.StatusOK,
		},
		{
			name:     "PUT home /forum not allowed",
			method:   "PUT",
			url:      "/forum",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "DELETE home /forum not allowed",
			method:   "DELETE",
			url:      "/forum",
			wantCode: http.StatusMethodNotAllowed,
		},
		// Test bad URL
		{
			name:     "POST /bad URL",
			method:   "POST",
			url:      "/bad",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "GET /bad URL",
			method:   "GET",
			url:      "/bad",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the HTTP request
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Record the response
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				indexHandler(w, r, "")
			})
			handler.ServeHTTP(rr, req)

			// Check the status code
			if rr.Code != tt.wantCode {
				t.Errorf("handler returned wrong status code: got %v but want %v", rr.Code, tt.wantCode)
			}

		})
	}
}

func TestRegisterHandler(t *testing.T) {
	Testinit()
	defer db.Close()

	tests := []struct {
		name       string
		method     string
		url        string
		body       string
		wantCode   int
		wantResult string // Expected db value
	}{
		// Test diff http methods
		{
			name:       "POST register valid",
			method:     "POST",
			url:        "/register",
			body:       "username=testuser&email=test@example.com&password=testpass",
			wantCode:   http.StatusSeeOther,
			wantResult: "testuser",
		},
		{
			name:     "GET register valid",
			method:   "GET",
			url:      "/register",
			body:     "",
			wantCode: http.StatusOK,
		},
		{
			name:     "PUT register not allowed",
			method:   "PUT",
			url:      "/register",
			body:     "",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "DELETE register not allowed",
			method:   "DELETE",
			url:      "/register",
			body:     "",
			wantCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the HTTP request
			req, err := http.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			// Set content type for POST requests
			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req.Header.Set("Content-Type", "")
			}

			// Record the response
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(registerHandler)
			handler.ServeHTTP(rr, req)

			// Check the status code
			if rr.Code != tt.wantCode {
				t.Errorf("handler returned wrong status code: got %v but want %v", rr.Code, tt.wantCode)
			}

			// Verify if user was added to the db for successful POST
			if tt.method == "POST" && tt.wantCode == http.StatusSeeOther && tt.wantResult != "" {
				var username string
				err = db.QueryRow("SELECT username FROM users WHERE username = ?", tt.wantResult).Scan(&username)
				if err != nil {
					t.Errorf("user not found in database: %v", err)
				}
				if username != tt.wantResult {
					t.Errorf("handler returned unexpected body: got %v but want %v", username, tt.wantResult)
				}
			}
		})
	}
}

func TestLogUserInHandler(t *testing.T) {
	Testinit()
	defer db.Close()

	tests := []struct {
		name     string
		method   string
		url      string
		body     string
		wantCode int
	}{
		// Test diff http methods
		{
			name:     "POST /loguserin valid",
			method:   "POST",
			url:      "/loguserin",
			body:     "username-or-email=testuser&password=testpass&return_url=/",
			wantCode: http.StatusSeeOther,
		},
		{
			name:     "GET /loguserin valid",
			method:   "GET",
			url:      "/loguserin",
			body:     "",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "PUT /loguserin not allowed",
			method:   "PUT",
			url:      "/loguserin",
			body:     "",
			wantCode: http.StatusMethodNotAllowed,
		},
		{
			name:     "DELETE /loguserin not allowed",
			method:   "DELETE",
			url:      "/loguserin",
			body:     "",
			wantCode: http.StatusMethodNotAllowed,
		},
	}

	// Add a test user to the database
	hashPass, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	db.Exec("INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)", "testid", "test@example.com", "testuser", string(hashPass))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new HTTP request
			req, err := http.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			// Set content type for POST requests
			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req.Header.Set("Content-Type", "")
			}

			// Record the response
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(logUserInHandler)
			handler.ServeHTTP(rr, req)

			// Check the status code
			if rr.Code != tt.wantCode {
				t.Errorf("handler returned wrong status code: got %v but want %v", rr.Code, tt.wantCode)
			}

			// Check if the session was created for POST requests
			if tt.method == "POST" && tt.wantCode == http.StatusSeeOther {
				var sessionToken string
				err = db.QueryRow("SELECT session_token FROM sessions WHERE user_id = ?", "testid").Scan(&sessionToken)
				if err != nil {
					t.Errorf("session not found in database: %v", err)
				}
				if sessionToken == "" {
					t.Errorf("handler did not create session")
				}
			}
		})
	}
}
