package router

import (
	"forum/internal/handlers"
	"net/http"
)

func SetHandlers() {

	fileServer := http.FileServer(http.Dir("./"))
	http.Handle("/static/", fileServer)
	http.Handle("/images/", fileServer)

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.IndexHandler(w, r, "")
	})
	http.HandleFunc("/thread/", handlers.ThreadPageHandler)
	http.HandleFunc("/add", handlers.AddThreadHandler)
	http.HandleFunc("/reply", handlers.AddReplyHandler)
	http.HandleFunc("/login", handlers.LogInHandler)
	http.HandleFunc("/loguserin", handlers.LogUserInHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/like", handlers.LikeHandler)
	http.HandleFunc("/dislike", handlers.DislikeHandler)
	http.HandleFunc("/expired", func(w http.ResponseWriter, r *http.Request) {
		handlers.IndexHandler(w, r, "Session expired")
	})
}
