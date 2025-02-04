package handlers

import (
	"fmt"
	"forum/internal/db"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
)

// validateSession returns user id, name and if session is (still) valid
func ValidateSession(r *http.Request) (string, string, bool) {
	validSes := true
	var userID string
	var userName string

	cookie, _ := r.Cookie("session_token")
	if cookie != nil {
		query := `SELECT user_id, username FROM sessions WHERE session_token = ? AND expires_at > ?`
		err := db.DB.QueryRow(query, cookie.Value, time.Now()).Scan(&userID, &userName)
		if err != nil { // invalid session
			validSes = false
		}
	} else {
		validSes = false
	}

	return userID, userName, validSes
}

func CheckUsername(s string) bool {
	if len(s) < 5 || len(s) > 25 {
		return false
	}
	for _, char := range s {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' && char != '_' {
			return false
		}
	}
	return true
}

func CheckPassword(s string) bool {
	if len(s) < 5 || len(s) > 25 {
		return false
	}
	for _, char := range s {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') {
			if char != '-' && char != '_' && char != '!' && char != '@' && char != '#' && char != '$' && char != '%' && char != '^' && char != '&' && char != '*' && char != '(' && char != ')' {
				return false
			}
		}
	}
	return true
}

func NameOremailExists(input string) bool {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ? OR email = ?)`
	err := db.DB.QueryRow(query, input, input).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func CreateSession() (string, error) {
	sessionUUID, err := uuid.NewV4() // Generate a new UUID
	if err != nil {
		return "", err
	}
	return sessionUUID.String(), nil
}

func SaveSession(userID, usname, sessionToken string, expiresAt time.Time) error {
	query := `INSERT INTO sessions (user_id, username, session_token, expires_at) VALUES (?, ?, ?, ?)`
	_, err := db.DB.Exec(query, userID, usname, sessionToken, expiresAt)
	return err
}

// deleteSession removes all sessions from the db by user Id
func deleteSession(w http.ResponseWriter, r *http.Request, usrId string) {
	query := `DELETE FROM sessions WHERE user_id = ?`
	_, err := db.DB.Exec(query, usrId)
	if err != nil {
		goToErrorPage("Failed to delete old session", http.StatusInternalServerError, w, r)
		return
	}
}

// sessionAndToken creates and puts a new session token into the database and into a user cookie
func sessionAndToken(w *http.ResponseWriter, r *http.Request, userID, username string) {
	// New session token
	sessionToken, err := CreateSession()
	if err != nil {
		goToErrorPage("Unable to create session: "+err.Error(), http.StatusInternalServerError, *w, r)
		return
	}
	expiresAt := time.Now().Add(30 * time.Minute)

	// Token into database
	err = SaveSession(userID, username, sessionToken, expiresAt)
	if err != nil {
		fmt.Println("Error saving session", err.Error())
		goToErrorPage("Unable to save session"+err.Error(), http.StatusInternalServerError, *w, r)
		return
	}

	// Token into cookie
	http.SetCookie(*w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		HttpOnly: true,
	})
}
