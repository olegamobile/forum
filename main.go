package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
	//_ "modernc.org/sqlite"
)

var (
	db         *sql.DB
	indexTmpl  *template.Template
	threadTmpl *template.Template
	signTmpl   *template.Template
)

func setHandlers() {
	// Initialize templates
	var err error
	indexTmpl, err = template.ParseFiles("templates/index.html", "templates/header.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	threadTmpl, err = template.ParseFiles("templates/thread.html", "templates/header.html", "templates/reply.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	signTmpl, err = template.ParseFiles("templates/signin.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// Set up routes
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/static/styles.css", http.StripPrefix("/static/", fileServer))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/thread/", threadPageHandler)
	http.HandleFunc("/add", addThreadHandler)
	http.HandleFunc("/reply", func(w http.ResponseWriter, r *http.Request) {
		addReplyHandler(w, r, "thread")
	})
	http.HandleFunc("/replytoreply", func(w http.ResponseWriter, r *http.Request) {
		addReplyHandler(w, r, "reply")
	})
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) { // Load page to sign in
		signTmpl.Execute(w, signData)
	})
	http.HandleFunc("/login", logUserHandler)
	http.HandleFunc("/register", addUserHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/like", likeHandler)
	http.HandleFunc("/dislike", dislikeHandler)
}

func main() {
	// Open database connection
	var err error
	db, err = sql.Open("sqlite3", "forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	makeTables()
	setHandlers()

	// Start the server
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
