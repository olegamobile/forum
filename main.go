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
	db           *sql.DB
	indexTmpl    *template.Template
	threadTmpl   *template.Template
	logTmpl      *template.Template
	registerTmpl *template.Template
	errorTmpl    *template.Template
)

func initTemplates() {
	var err error
	indexTmpl, err = template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	threadTmpl, err = template.ParseFiles("templates/thread.html", "templates/header.html", "templates/reply.html", "templates/footer.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	logTmpl, err = template.ParseFiles("templates/login.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	registerTmpl, err = template.ParseFiles("templates/registerUser.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	errorTmpl, err = template.ParseFiles("templates/error.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
}

func setHandlers() {
	fileServer := http.FileServer(http.Dir("./"))
	http.Handle("/static/css/styles.css", fileServer)
	http.Handle("/static/css/darkmode.css", fileServer)
	http.Handle("/static/css/headerfooter.css", fileServer)
	http.Handle("/static/css/image_upload_style.css", fileServer)
	http.Handle("/static/home-functions.js", fileServer)
	http.Handle("/static/ui-functions.js", fileServer)
	http.Handle("/static/image_upload.js", fileServer)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexHandler(w, r, "")
	})
	http.HandleFunc("/thread/", threadPageHandler)
	http.HandleFunc("/add", addThreadHandler)
	http.HandleFunc("/reply", addReplyHandler)
	http.HandleFunc("/login", logInHandler)
	http.HandleFunc("/loguserin", logUserInHandler)
	http.HandleFunc("/register", registerHandler)
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
	log.Fatal(http.ListenAndServe(":8080", nil)) // Logs the error and exits.
}
