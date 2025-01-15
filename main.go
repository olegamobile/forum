package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db         *sql.DB
	indexTmpl  *template.Template
	threadTmpl *template.Template
	signTmpl   *template.Template
)

func initTemplates() {
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
	signTmpl, err = template.ParseFiles("templates/signin.html", "templates/header.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
}

func setHandlers() {
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/static/styles.css", http.StripPrefix("/static/", fileServer))
	http.Handle("/static/ui-functions.js", http.StripPrefix("/static/", fileServer))

	http.Handle("/favicon.ico", http.NotFoundHandler()) //accessing favicon will cause 404

	//http.HandleFunc("/", indexHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexHandler(w, r, "")
	})
	http.HandleFunc("/thread/", threadPageHandler)
	http.HandleFunc("/add", addThreadHandler)
	http.HandleFunc("/reply", addReplyHandler)
	http.HandleFunc("/signin", signInHandler)
	http.HandleFunc("/login", logUserInHandler)
	http.HandleFunc("/register", addUserHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/like", likeHandler)
	http.HandleFunc("/dislike", dislikeHandler)
	http.HandleFunc("/expired", func(w http.ResponseWriter, r *http.Request) {
		indexHandler(w, r, "Session expired")
	})
}

// sessionCleanup removes expired sessions every given time interval
func sessionCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			log.Println("Running session cleanup...")
			removeExpiredSessions()
		}
	}()
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
	sessionCleanup(time.Hour) // Remove expired sessions every hour
	initTemplates()
	setHandlers()

	// Start the server
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
