package main

import (
	"database/sql"
	"fmt"
	"net/http"
)

type Thread struct {
	ID          int
	Author      string
	Title       string
	Content     string
	Created     string
	CreatedDay  string
	CreatedTime string
	Categories  string
	CatsSlice   []string
	Likes       int
	Dislikes    int
	RepliesN    int
	Replies     []Reply
	BaseID      int
	LikedNow    bool
	DislikedNow bool
}

type PageData struct {
	Threads   []Thread
	ValidSes  bool
	UsrId     int
	UsrNm     string
	Message   string
	Selection string
	Search    string
}

func countReactions(id int) (int, int) {
	reactionsQuery := `SELECT reaction_type, COUNT(*) AS count FROM post_reactions WHERE post_id = ? GROUP BY reaction_type;`
	rows, err := db.Query(reactionsQuery, id)
	if err != nil {
		fmt.Println("Fetching reactions query failed", err.Error())
		return 0, 0
	}
	defer rows.Close()

	var likes, dislikes int
	for rows.Next() {
		var reactionType string
		var count int
		if err := rows.Scan(&reactionType, &count); err != nil {
			fmt.Printf("Failed to scan row: %v\n", err)
		}

		// Assign counts based on reaction type
		switch reactionType {
		case "like":
			likes = count
		case "dislike":
			dislikes = count
		}
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		fmt.Printf("Error iterating rows: %v\n", err)
	}

	return likes, dislikes
}

func fetchThreads(rowsThreads *sql.Rows) ([]Thread, error) {

	// Find all posts without an empty title
	/* 	rowsThreads, err := db.Query(selectQuery)
	   	if err != nil {
	   		fmt.Println("fetchThreads selectQuery failed", err.Error())
	   		return nil, err
	   	}
	   	defer rowsThreads.Close() */

	var threads []Thread
	for rowsThreads.Next() {
		var th Thread
		err := rowsThreads.Scan(&th.ID, &th.Author, &th.Title, &th.Content, &th.Created, &th.Categories)
		if err != nil {
			fmt.Println("fetchThreads rows scanning:", err.Error())
			return nil, err
		}

		th, err = dataToThread(th)
		if err != nil {
			return nil, err
		}

		threads = append(threads, th)
	}

	return threads, nil
}

func findThreads(filter string, authId int) ([]Thread, error) {

	var selectQuery string
	var rowsThreads *sql.Rows
	var err error

	if filter == "" || filter == "all" {
		selectQuery = `SELECT id, author, title, content, created_at, categories FROM posts WHERE title != "";`
		rowsThreads, err = db.Query(selectQuery)
		if err != nil {
			fmt.Println("findThreads selectQuery failed", err.Error())
			return nil, err
		}
		defer rowsThreads.Close()
	}

	if filter == "created" {
		selectQuery = `SELECT id, author, title, content, created_at, categories FROM posts WHERE title != "" AND authorID = ?;`
		rowsThreads, err = db.Query(selectQuery, authId)
		if err != nil {
			fmt.Println("findThreads selectQuery failed", err.Error())
			return nil, err
		}
		defer rowsThreads.Close()
	}
	if filter == "liked" {
		// Find all threads and filter them otherwise?
	}
	if filter == "disliked" {
		// Find all threads and filter them otherwise?
	}

	return fetchThreads(rowsThreads)
}

func fetchReplies(thisID int) ([]Reply, error) {
	selectQueryReplies := `SELECT id, base_id, author, content, created_at FROM posts WHERE base_id = ? AND title = '';` // All replies
	rows, err := db.Query(selectQueryReplies, thisID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	replies := getReplies(rows, thisID)
	return replies, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request, msg string) {

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed: ", http.StatusMethodNotAllowed)
		return
	}

	usId, usName, validSes := validateSession(r)
	selection := r.FormValue("todisplay")
	search := r.FormValue("usersearch")

	var threads []Thread
	var err error
	if r.Method == http.MethodGet || !validSes {
		threads, err = findThreads("", usId)
	} else {

		// If user did a POST, session should be valid

		if r.FormValue("updatesel") == "update" {
			switch r.FormValue("todisplay") {
			case "created":
				threads, err = findThreads("created", usId)
			case "liked":
				threads, err = findThreads("liked", usId)
			case "disliked":
				threads, err = findThreads("disliked", usId)
			default:
				threads, err = findThreads("", usId)
			}
			search = ""
		}
		if r.FormValue("serchcat") == "search" {
			// filter threads by category
			threads, err = findThreads("", usId)
			selection = ""
		}
	}

	if err != nil {
		http.Error(w, "Error fetching threads", http.StatusInternalServerError)
		return
	}

	for i, th := range threads {
		replies, err := fetchReplies(th.ID)
		if err != nil {
			fmt.Println("Error fetching replies:", err.Error())
			http.Error(w, "Error fetching replies", http.StatusInternalServerError)
			return
		}
		threads[i].RepliesN = len(replies)
	}

	data := PageData{Threads: threads, ValidSes: validSes, UsrId: usId, UsrNm: usName, Message: msg, Selection: selection, Search: search}
	indexTmpl.Execute(w, data)
}
