package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"
)

type Thread struct {
	ID         int
	Author     string
	Title      string
	Content    string
	Created    string
	Categories string
	CatsSlice  []string
	Likes      int
	Dislikes   int
	RepliesN   int
	Replies    []Reply
	BaseID     int
}

type PageData struct {
	Threads []Thread
}

func fetchThreads(db *sql.DB) ([]Thread, error) {
	selectQuery := `SELECT id, author, title, content, created_at, categories, likes, dislikes FROM threads;`
	rows, err := db.Query(selectQuery)
	if err != nil {
		fmt.Println("fetchThreads selectQuery failed", err.Error())
		return nil, err
	}
	defer rows.Close()

	var threads []Thread
	for rows.Next() {
		var th Thread
		err := rows.Scan(&th.ID, &th.Author, &th.Title, &th.Content, &th.Created, &th.Categories, &th.Likes, &th.Dislikes)
		if err != nil {
			fmt.Println("fetchThreads rows scanning:", err.Error())
			return nil, err
		}
		th.BaseID = th.ID
		threads = append(threads, th)
	}

	return threads, nil
}

func fetchReplies(db *sql.DB) ([]Reply, error) {
	selectQuery := `SELECT id, base_id, author, content, created_at, likes, dislikes, parent_id, parent_type FROM replies;`
	rows, err := db.Query(selectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []Reply
	for rows.Next() {
		var re Reply
		err := rows.Scan(&re.ID, &re.BaseID, &re.Author, &re.Content, &re.Created, &re.Likes, &re.Dislikes, &re.ParentID, &re.ParentType)
		if err != nil {
			return nil, err
		}
		replies = append(replies, re)
	}
	return replies, nil
}

func indexHandler(db *sql.DB, tmpl *template.Template, w http.ResponseWriter, r *http.Request) {
	threads, err := fetchThreads(db)
	if err != nil {
		http.Error(w, "Error fetching threads", http.StatusInternalServerError)
		return
	}

	replies, err := fetchReplies(db)
	if err != nil {
		fmt.Println("Error fetching replies:", err.Error())
		http.Error(w, "Error fetching replies", http.StatusInternalServerError)
		return
	}

	for i, po := range threads {
		for _, re := range replies {
			if po.ID == re.ParentID {
				threads[i].RepliesN++
			}
		}
	}

	signData.Message1 = ""
	signData.Message2 = ""
	data := PageData{Threads: threads}
	tmpl.Execute(w, data)
}
