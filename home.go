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
	Threads  []Thread
	ValidSes bool
	UsrId    int
	UsrNm    string
}

func countReactions(id int, postType string) (int, int) {

	reactionsQuery := `SELECT reaction_type, COUNT(*) AS count FROM post_reactions WHERE post_id = ? AND post_type = ? GROUP BY reaction_type;`
	rows, err := db.Query(reactionsQuery, id, postType)
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

func fetchThreads(db *sql.DB) ([]Thread, error) {
	selectQuery := `SELECT id, author, title, content, created_at, categories FROM threads;`
	rowsThreads, err := db.Query(selectQuery)
	if err != nil {
		fmt.Println("fetchThreads selectQuery failed", err.Error())
		return nil, err
	}
	defer rowsThreads.Close()

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

func fetchReplies(db *sql.DB, thisID int) ([]Reply, error) {

	//selectQueryReplies := `SELECT id, base_id, author, content, created_at FROM replies WHERE parent_id = ? AND parent_type = ?;`
	//rows, err := db.Query(selectQueryReplies, thisID, "thread")
	selectQueryReplies := `SELECT id, base_id, author, content, created_at FROM replies WHERE base_id = ?;` // All or direct replies?
	rows, err := db.Query(selectQueryReplies, thisID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	replies := getReplies(rows, thisID)
	return replies, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	threads, err := fetchThreads(db)
	if err != nil {
		http.Error(w, "Error fetching threads", http.StatusInternalServerError)
		return
	}

	for i, th := range threads {
		replies, err := fetchReplies(db, th.ID)
		if err != nil {
			fmt.Println("Error fetching replies:", err.Error())
			http.Error(w, "Error fetching replies", http.StatusInternalServerError)
			return
		}
		threads[i].RepliesN = len(replies)
	}

	signData.Message1 = ""
	signData.Message2 = ""

	usId, usName, validSes := validateSession(r)

	data := PageData{Threads: threads, ValidSes: validSes, UsrId: usId, UsrNm: usName}
	indexTmpl.Execute(w, data)
}
