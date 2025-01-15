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

type SignData struct {
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

func addUserHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed: ", http.StatusMethodNotAllowed)
		return
	}

	name := html.EscapeString(r.FormValue("username"))
	email := html.EscapeString(r.FormValue("email"))
	pass := html.EscapeString(r.FormValue("password"))
	var signData SignData
	signData.UsrId, signData.UsrNm, signData.ValidSes = validateSession(r)

	if signData.ValidSes {
		fmt.Println(signData.UsrNm + "trying to create a new user while signed-in")
		signData.Message1 = "Signed in as " + signData.UsrNm + ". Log out first."
		signTmpl.Execute(w, signData)
		return
	}

	nameOk, passOk := checkString(name), checkString(pass)
	_, emailErr := mail.ParseAddress(email)

	if !nameOk || !passOk {
		fmt.Println("Minimum 5 chars, limited chars")
		signData.Message2 = "Minimum 5 characters in usename and password. Only letters, numbers and symbols allowed."
		signTmpl.Execute(w, signData)
		return
	}

	if emailErr != nil {
		fmt.Println("Invalid email address")
		signData.Message2 = "Invalid email address"
		signTmpl.Execute(w, signData)
		return
	}

	if nameExists(db, name) {
		fmt.Println("Name already taken")
		signData.Message2 = "Name already taken"
		signTmpl.Execute(w, signData)
		return
	}

	if emailExists(db, email) {
		fmt.Println("Email already taken")
		signData.Message2 = "Email already taken"
		signTmpl.Execute(w, signData)
		return
	}

	hashPass, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	userId, err := uuid.NewV4() // Generate a new UUID user id
	if err != nil {
		http.Error(w, "Error generating Id for user", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?);`, userId, email, name, string(hashPass))
	if err != nil {
		fmt.Println("Adding:", err.Error())
		http.Error(w, "Error adding user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

}

// deleteSession removes all sessions from the db by user Id
func deleteSession(w http.ResponseWriter, usrId string) {
	query := `DELETE FROM sessions WHERE user_id = ?`
	_, err := db.Exec(query, usrId)
	if err != nil {
		http.Error(w, "Failed to delete old session", http.StatusInternalServerError)
		return
	}
}

func logUserInHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed: ", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("username")
	email := r.FormValue("email")
	pass := r.FormValue("password")
	var signData SignData
	signData.UsrId, signData.UsrNm, signData.ValidSes = validateSession(r)

	if signData.ValidSes {
		fmt.Println(signData.UsrNm + "trying to sign in while already signed-in")
		signData.Message1 = "Already signed in as " + signData.UsrNm + ". Log out first."
		signTmpl.Execute(w, signData)
		return
	}

	// Checking user
	if !nameExists(db, name) && !emailExists(db, email) {
		fmt.Println("User not found")
		signData.Message1 = "User not found"
		signTmpl.Execute(w, signData)
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
		signData.Message1 = "Password incorrect"
		signTmpl.Execute(w, signData)
		return
	} else {
		fmt.Println("Correct password")
	}

	// Remove any old sessions so user can't be active on different browsers (audit question)
	deleteSession(w, userID)

	// Cookie and session
	sessionToken, err := createSession()
	if err != nil {
		http.Error(w, "Unable to create session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	expiresAt := time.Now().Add(30 * time.Minute) // Set session validity length

	err = saveSession(db, userID, name, sessionToken, expiresAt)
	if err != nil {
		fmt.Println("Error saving session", err.Error())
		http.Error(w, "Unable to save session", http.StatusInternalServerError)
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

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed: ", http.StatusMethodNotAllowed)
		return
	}

	_, _, validSes := validateSession(r)
	if !validSes {
		fmt.Println("Trying to log out while not signed-in")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "No session found", http.StatusBadRequest)
		return
	}

	// Delete the session from the database
	query := `DELETE FROM sessions WHERE session_token = ?`
	_, err = db.Exec(query, cookie.Value)
	if err != nil {
		http.Error(w, "Failed to log out", http.StatusInternalServerError)
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

func signInHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed: ", http.StatusMethodNotAllowed)
		return
	}

	var signData SignData
	signData.UsrId, signData.UsrNm, signData.ValidSes = validateSession(r)
	signTmpl.Execute(w, signData)
}
