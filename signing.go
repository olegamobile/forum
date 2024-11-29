package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/mail"
	"time"
	"unicode"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type SignData struct {
	Message1 string
	Message2 string
}

var signData SignData

func checkString(s string) bool {
	if len(s) < 5 {
		return false
	}
	for _, ch := range s {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && !unicode.IsSymbol(ch) {
			fmt.Println("Only letters, numbers and symbols allowed:", string(ch))
			return false
		}
	}
	return true
}

func nameExists(db *sql.DB, username string) bool {
	selectQueryName := `SELECT username FROM users WHERE username = ?`
	err := db.QueryRow(selectQueryName, username).Scan(&username)
	return err == nil // no error if name found
}

func emailExists(db *sql.DB, mail string) bool {
	selectQueryMail := `SELECT email FROM users WHERE email = ?`
	err := db.QueryRow(selectQueryMail, mail).Scan(&mail)
	return err == nil // no error if email found
}

func addUserHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("username")
		email := r.FormValue("email")
		pass := r.FormValue("password")

		n, p := checkString(name), checkString(pass)
		_, e := mail.ParseAddress(email)

		if !n || !p {
			fmt.Println("Minimum 5 chars, limited chars")
			signData.Message2 = "Minimum 5 characters in usename and password. Only letters, numbers and symbols allowed."
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if e != nil {
			fmt.Println("Invalid email address")
			signData.Message2 = "Invalid email address"
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if nameExists(db, name) {
			fmt.Println("Name already taken")
			signData.Message2 = "Name already taken"
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if emailExists(db, email) {
			fmt.Println("Email already taken")
			signData.Message2 = "Email already taken"
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		hashPass, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

		_, err := db.Exec(`INSERT INTO users (email, username, password) VALUES (?, ?, ?);`, email, name, string(hashPass))
		if err != nil {
			fmt.Println("Adding:", err.Error())
			http.Error(w, "Error adding user", http.StatusInternalServerError)
			return
		}

		signData.Message2 = ""

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func createSession() (string, error) {
	sessionUUID, err := uuid.NewV4() // Generate a new UUID
	if err != nil {
		return "", err
	}
	return sessionUUID.String(), nil
}

func saveSession(db *sql.DB, userID int, sessionToken string, expiresAt time.Time) error {
	query := `INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)`
	_, err := db.Exec(query, userID, sessionToken, expiresAt)
	return err
}

func logUserHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("username")
		email := r.FormValue("email")
		pass := r.FormValue("password")

		// Checking usesr
		if !nameExists(db, name) && !emailExists(db, email) {
			fmt.Println("User not found")
			signData.Message1 = "User not found"
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		// Checking password
		storedHashedPass, userID := "", 0
		if nameExists(db, name) {
			query := `SELECT password, id FROM users WHERE username = ?`
			db.QueryRow(query, name).Scan(&storedHashedPass, &userID)
		} else {
			query := `SELECT password, id FROM users WHERE email = ?`
			db.QueryRow(query, email).Scan(&storedHashedPass, &userID)
		}

		err := bcrypt.CompareHashAndPassword([]byte(storedHashedPass), []byte(pass))

		if err != nil {
			fmt.Println("Password incorrect")
			signData.Message1 = "Password incorrect"
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		} else {
			fmt.Println("Correct password")
		}

		// Cookie and session
		sessionToken, err := createSession()
		if err != nil {
			http.Error(w, "Unable to create session", http.StatusInternalServerError)
			return
		}

		expiresAt := time.Now().Add(2 * time.Hour) // 2-hour expiration

		err = saveSession(db, userID, sessionToken, expiresAt)
		if err != nil {
			http.Error(w, "Unable to save session", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{ // Set the session token as a cookie
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  expiresAt,
			HttpOnly: true,
		})

		// x sign in
		// x set a cookie
		//   Check session token on every request
		//   Log out: delete the cookie and remove it from the database

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
