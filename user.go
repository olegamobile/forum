package main

import (
	"database/sql"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type loginData struct {
	Message1  string
	Message2  string
	ValidSes  bool
	UsrId     string
	UsrNm     string
	ReturnURL string
	LoginURL  string
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

func checkUsername(s string) bool {
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

func checkPassword(s string) bool {
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

func nameOremailExists(db *sql.DB, input string) bool {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ? OR email = ?)`
	err := db.QueryRow(query, input, input).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func createSession() (string, error) {
	sessionUUID, err := uuid.NewV4() // Generate a new UUID
	if err != nil {
		return "", err
	}
	return sessionUUID.String(), nil
}

func saveSession(userID, usname, sessionToken string, expiresAt time.Time) error {
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
	loginData.LoginURL = "/login"

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

		//nameOk, passOk := checkString(name), checkString(pass)
		nameOk, passOk := checkUsername(name), checkPassword(pass)
		_, emailErr := mail.ParseAddress(email)

		if !nameOk || !passOk {
			fmt.Println("Minimum 5 chars, limited chars")
			loginData.Message2 = "5-25 characters in username and password. Only letters, numbers and symbols allowed."
			registerTmpl.Execute(w, loginData)
			return
		}

		if emailErr != nil {
			fmt.Println("Invalid email address")
			loginData.Message2 = "Invalid email address"
			registerTmpl.Execute(w, loginData)
			return
		}

		if nameOremailExists(db, name) {
			fmt.Println("Name already taken")
			loginData.Message2 = "Name already taken"
			registerTmpl.Execute(w, loginData)
			return
		}

		if nameOremailExists(db, email) {
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

		// Log user in
		sessionAndToken(&w, r, userId.String(), name)

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

// sessionAndToken creates and puts a new session token into the database and into a user cookie
func sessionAndToken(w *http.ResponseWriter, r *http.Request, userID, username string) {
	// New session token
	sessionToken, err := createSession()
	if err != nil {
		goToErrorPage("Unable to create session: "+err.Error(), http.StatusInternalServerError, *w, r)
		return
	}
	expiresAt := time.Now().Add(30 * time.Minute)

	// Token into database
	err = saveSession(userID, username, sessionToken, expiresAt)
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

// logUserInHandler starts a session for the user if login is succesful
func logUserInHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/loguserin" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	if r.Method != http.MethodPost {
		fmt.Println("Bad method on logUserInHandler:", r.Method)
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	nameOrEmail := r.FormValue("username-or-email")
	pass := r.FormValue("password")
	returnUrl := r.FormValue("return_url")

	var loginData loginData
	loginData.UsrId, loginData.UsrNm, loginData.ValidSes = validateSession(r)
	loginData.ReturnURL, loginData.LoginURL = returnUrl, "/login?return_url="+returnUrl

	if loginData.ValidSes {
		fmt.Println(loginData.UsrNm + " trying to log in while already logged-in")
		loginData.Message1 = "Already logged in as " + loginData.UsrNm + ". Log out first."
		logTmpl.Execute(w, loginData)
		return
	}

	// Checking if user or email exists
	if !nameOremailExists(db, nameOrEmail) {
		fmt.Println("User not found")
		loginData.Message1 = "Invalid username/email or password"
		logTmpl.Execute(w, loginData)
		return
	}

	// Get user information and check password
	var storedHashedPass, userID, username string
	query := `SELECT password, id, username FROM users WHERE username = ? OR email = ?`
	err := db.QueryRow(query, nameOrEmail, nameOrEmail).Scan(&storedHashedPass, &userID, &username)
	if err != nil {
		loginData.Message1 = "Invalid username/email or password"
		logTmpl.Execute(w, loginData)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHashedPass), []byte(pass))
	if err != nil {
		fmt.Println("Password incorrect")
		loginData.Message1 = "Invalid username/email or password"
		logTmpl.Execute(w, loginData)
		return
	}

	// Remove any old sessions
	deleteSession(w, r, userID)
	// Create new session and token
	sessionAndToken(&w, r, userID, username)

	http.Redirect(w, r, returnUrl, http.StatusSeeOther)
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

	// Determine the previous page
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}

	http.Redirect(w, r, referer, http.StatusSeeOther)
}

// logInHandler handles user clicking log-in link
func logInHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/login") {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	if r.Method != http.MethodGet {
		fmt.Println("Disallowed method at logInHandler:", r.Method)
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	var loginData loginData
	loginData.UsrId, loginData.UsrNm, loginData.ValidSes = validateSession(r)
	loginData.ReturnURL, loginData.LoginURL = r.URL.Query().Get("return_url"), "/login"
	if loginData.ReturnURL == "" {
		loginData.ReturnURL = "/"
	}

	logTmpl.Execute(w, loginData)
}
