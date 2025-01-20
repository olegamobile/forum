package main

import (
	"database/sql"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/mail"
	"time"
	"unicode"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type loginData struct {
	Message1 string
	Message2 string
	ValidSes bool
	UsrId    string
	UsrNm    string
}

// removeExpiredSessions deletes all expired sessions
func removeExpiredSessions() {
	_, err := db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	if err != nil {
		log.Printf("Error deleting expired sessions: %v\n", err.Error())
	}
}

// validateSession returns user id, name and if session is (still) valid
func validateSession(r *http.Request) (string, string, bool) {
	//removeExpiredSessions()	// Doing this every hour instead from main()

	validSes := true
	var userID string
	var userName string

	cookie, _ := r.Cookie("session_token")
	if cookie != nil {
		query := `SELECT user_id, username FROM sessions WHERE session_token = ? AND expires_at > ?`
		err := db.QueryRow(query, cookie.Value, time.Now()).Scan(&userID, &userName)
		if err != nil { // invalid session
			//fmt.Println("Invalid session:", err.Error())
			validSes = false
		}
	} else {
		validSes = false
	}

	return userID, userName, validSes
}

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

func createSession() (string, error) {
	sessionUUID, err := uuid.NewV4() // Generate a new UUID
	if err != nil {
		return "", err
	}
	return sessionUUID.String(), nil
}

func saveSession(db *sql.DB, userID, usname, sessionToken string, expiresAt time.Time) error {
	query := `INSERT INTO sessions (user_id, username, session_token, expires_at) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, userID, usname, sessionToken, expiresAt)
	return err
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/register" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	var loginData loginData
	loginData.UsrId, loginData.UsrNm, loginData.ValidSes = validateSession(r)

	if loginData.ValidSes {
		fmt.Println(loginData.UsrNm + " trying to create a new user while logged-in")
		loginData.Message1 = "Logged in as " + loginData.UsrNm + ". Log out first."
		logTmpl.Execute(w, loginData)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Handle GET request - show registration form
		registerTmpl.Execute(w, loginData)
		return

	case http.MethodPost:
		name := html.EscapeString(r.FormValue("username"))
		email := html.EscapeString(r.FormValue("email"))
		pass := html.EscapeString(r.FormValue("password"))

		nameOk, passOk := checkString(name), checkString(pass)
		_, emailErr := mail.ParseAddress(email)

		if !nameOk || !passOk {
			fmt.Println("Minimum 5 chars, limited chars")
			loginData.Message2 = "Minimum 5 characters in username and password. Only letters, numbers and symbols allowed."
			registerTmpl.Execute(w, loginData)
			return
		}

		if emailErr != nil {
			fmt.Println("Invalid email address")
			loginData.Message2 = "Invalid email address"
			registerTmpl.Execute(w, loginData)
			return
		}

		if nameExists(db, name) {
			fmt.Println("Name already taken")
			loginData.Message2 = "Name already taken"
			registerTmpl.Execute(w, loginData)
			return
		}

		if emailExists(db, email) {
			fmt.Println("Email already taken")
			loginData.Message2 = "Email already taken"
			registerTmpl.Execute(w, loginData)
			return
		}

		hashPass, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		userId, err := uuid.NewV4() // Generate a new UUID user id
		if err != nil {
			goToErrorPage("Error generating Id for user", http.StatusInternalServerError, w, r)
			return
		}

		_, err = db.Exec(`INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?);`, userId, email, name, string(hashPass))
		if err != nil {
			fmt.Println("Adding:", err.Error())
			goToErrorPage("Error adding user", http.StatusInternalServerError, w, r)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)

	default:
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}
}

// deleteSession removes all sessions from the db by user Id
func deleteSession(w http.ResponseWriter, r *http.Request, usrId string) {
	query := `DELETE FROM sessions WHERE user_id = ?`
	_, err := db.Exec(query, usrId)
	if err != nil {
		goToErrorPage("Failed to delete old session", http.StatusInternalServerError, w, r)
		return
	}
}

func logUserInHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/loguserin" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	if r.Method != http.MethodPost {
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	name := r.FormValue("username")
	email := r.FormValue("email")
	pass := r.FormValue("password")
	var loginData loginData
	loginData.UsrId, loginData.UsrNm, loginData.ValidSes = validateSession(r)

	if loginData.ValidSes {
		fmt.Println(loginData.UsrNm + "trying to log in while already logged-in")
		loginData.Message1 = "Already logged in as " + loginData.UsrNm + ". Log out first."
		logTmpl.Execute(w, loginData)
		return
	}

	// Checking user
	if !nameExists(db, name) && !emailExists(db, email) {
		fmt.Println("User not found")
		loginData.Message1 = "User not found"
		logTmpl.Execute(w, loginData)
		return
	}

	// Checking password, find with either name or email
	storedHashedPass, userID := "", ""
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
		loginData.Message1 = "Password incorrect"
		logTmpl.Execute(w, loginData)
		return
	} else {
		fmt.Println("Correct password")
	}

	// Remove any old sessions so user can't be active on different browsers (audit question)
	deleteSession(w, r, userID)

	// Cookie and session
	sessionToken, err := createSession()
	if err != nil {
		goToErrorPage("Unable to create session: "+err.Error(), http.StatusInternalServerError, w, r)
		return
	}
	expiresAt := time.Now().Add(30 * time.Minute) // Set session validity length

	err = saveSession(db, userID, name, sessionToken, expiresAt)
	if err != nil {
		fmt.Println("Error saving session", err.Error())
		goToErrorPage("Unable to save session"+err.Error(), http.StatusInternalServerError, w, r)
		return
	}

	// Set the session token as a cookie. Cookie is added to the writer's header.
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/logout" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	if r.Method != http.MethodPost {
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	_, _, validSes := validateSession(r)
	if !validSes {
		fmt.Println("Trying to log out while not logged-in")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		goToErrorPage("No session found", http.StatusBadRequest, w, r)
		return
	}

	// Delete the session from the database
	query := `DELETE FROM sessions WHERE session_token = ?`
	_, err = db.Exec(query, cookie.Value)
	if err != nil {
		goToErrorPage("Failed to log out", http.StatusInternalServerError, w, r)
		return
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour), // Expire immediately
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// logInHandler handles user clicking log-in link
func logInHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	if r.Method != http.MethodGet {
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	var loginData loginData
	loginData.UsrId, loginData.UsrNm, loginData.ValidSes = validateSession(r)
	logTmpl.Execute(w, loginData)
}
