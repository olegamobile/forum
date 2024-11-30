package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
	//_ "modernc.org/sqlite"
)

func setHandlers(db *sql.DB) {
	// Initialize templates
	indexTmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	threadTmpl, err := template.ParseFiles("templates/thread.html", "templates/reply.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	signTmpl, err := template.ParseFiles("templates/signin.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// Set up routes
	http.HandleFunc("/static/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/styles.css")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexHandler(db, indexTmpl, w, r)
	})
	http.HandleFunc("/thread/", func(w http.ResponseWriter, r *http.Request) {
		threadPageHandler(db, threadTmpl, w, r)
	})
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		addThreadHandler(db, w, r)
	})
	http.HandleFunc("/reply", func(w http.ResponseWriter, r *http.Request) {
		addReplyHandler(db, w, r, "thread")
	})
	http.HandleFunc("/replytoreply", func(w http.ResponseWriter, r *http.Request) {
		addReplyHandler(db, w, r, "reply")
	})
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		signTmpl.Execute(w, signData)
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		logUserHandler(db, w, r)
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		addUserHandler(db, w, r)
	})
}

func main() {
	// Open database connection
	db, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	makeTables(db)
	setHandlers(db)

	// Start the server
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
