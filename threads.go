package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"
)

type Reply struct {
	ID          int
	Author      string
	Content     string
	Created     string
	CreatedDay  string
	CreatedTime string
	Likes       int
	Dislikes    int
	ParentID    int
	ParentType  string
	Replies     []Reply
	BaseID      int
	ValidSes    bool
}

type threadPageData struct {
	Thread   Thread
	ValidSes bool
	UsrId    int
	UsrNm    string
}

func addThreadHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	authID, author, valid := validateSession(db, r)

	if valid && r.Method == http.MethodPost {
		title := strings.TrimSpace(r.FormValue("title"))
		content := strings.TrimSpace(r.FormValue("content"))
		rawCats := strings.ToLower(r.FormValue("categories"))

		cleanCats := ""
		for _, char := range strings.TrimSpace(rawCats) {
			if !unicode.IsPunct(char) {
				cleanCats += string(char)
			}
		}
		catsJson, _ := json.Marshal(strings.Fields(cleanCats))

		if content != "" {
			_, err := db.Exec(`INSERT INTO threads (author, authorID, title, content, categories) VALUES (?, ?, ?, ?, ?);`, author, authID, title, content, string(catsJson))
			if err != nil {
				fmt.Println("Adding:", err.Error())
				http.Error(w, "Error adding thread", http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	if !valid {
		// Session perhaps expired during writing
		// Print message?
	}
}

func addReplyHandler(db *sql.DB, w http.ResponseWriter, r *http.Request, parentType string) {
	authID, author, valid := validateSession(db, r)

	if valid && r.Method == http.MethodPost {
		content := strings.TrimSpace(r.FormValue("content"))
		parId := r.FormValue("parentId") // No int conversion necessary
		baseId := r.FormValue("baseId")  // No int conversion necessary

		if content != "" {
			_, err := db.Exec(`INSERT INTO replies (base_id, author, authorID, content, parent_id, parent_type) VALUES (?, ?, ?, ?, ?, ?);`, baseId, author, authID, content, parId, parentType)
			if err != nil {
				fmt.Println("Replying:", err.Error())
				http.Error(w, "Error adding reply", http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "/thread/"+baseId, http.StatusSeeOther)
	}

	if !valid {
		// Session maybe expired during writing
		// Print message?
	}
}

func recurseReplies(db *sql.DB, this *Reply) {
	selectQueryReplies := `SELECT id, base_id, author, content, created_at, likes, dislikes FROM replies WHERE parent_id = ? AND parent_type = ?;`
	rows, err := db.Query(selectQueryReplies, this.ID, "reply")
	if err != nil {
		fmt.Println("Error getting replies for reply:", err.Error())
		return
	}
	defer rows.Close()

	var replies []Reply
	for rows.Next() {
		var re Reply
		err2 := rows.Scan(&re.ID, &re.BaseID, &re.Author, &re.Content, &re.Created, &re.Likes, &re.Dislikes)
		if err2 != nil {
			fmt.Println("Error rading reply rows for reply:", err2.Error())
			return
		}
		re.ParentID = this.ID
		re.ParentType = "reply"

		createdGoTime, err := time.Parse(time.RFC3339, re.Created) // 2024-12-02T15:44:52Z
		if err != nil {
			return
		}

		// Convert to Finnish timezone (UTC+2)
		location, err := time.LoadLocation("Europe/Helsinki")
		if err != nil {
			return
		}
		createdGoTime = createdGoTime.In(location)

		re.CreatedDay = createdGoTime.Format("2.1.2006")
		re.CreatedTime = createdGoTime.Format("15.04.05")

		replies = append(replies, re)
	}

	if len(replies) != 0 {
		this.Replies = replies

		fmt.Println(this.Replies)

		for i := 0; i < len(this.Replies); i++ {
			recurseReplies(db, &this.Replies[i])
		}
	}
}

func findThread(db *sql.DB, id int) (Thread, error) {
	var thread Thread
	selectQueryThread := `SELECT id, author, title, content, created_at, categories, likes, dislikes FROM threads WHERE id = ?;`
	err1 := db.QueryRow(selectQueryThread, id).Scan(&thread.ID, &thread.Author, &thread.Title, &thread.Content, &thread.Created, &thread.Categories, &thread.Likes, &thread.Dislikes)
	if err1 != nil {
		return thread, err1
	}

	selectQueryReplies := `SELECT id, base_id, author, content, created_at, likes, dislikes FROM replies WHERE parent_id = ? AND parent_type = ?;`
	rows, err2 := db.Query(selectQueryReplies, thread.ID, "thread")
	if err2 != nil {
		return thread, err2
	}
	defer rows.Close()

	var replies []Reply
	for rows.Next() {
		var re Reply
		err3 := rows.Scan(&re.ID, &re.BaseID, &re.Author, &re.Content, &re.Created, &re.Likes, &re.Dislikes)
		if err3 != nil {
			return thread, err3
		}
		re.ParentID = thread.ID
		re.ParentType = "thread"
		replies = append(replies, re)
	}

	// Add replies to replies recursively
	for i := 0; i < len(replies); i++ {
		recurseReplies(db, &(replies[i]))
	}

	fmt.Println("Replies 0 times:", replies[0].CreatedDay, replies[0].CreatedTime)

	thread.Replies = replies
	thread.BaseID = thread.ID
	return thread, err1
}

// markValidity writes to each reply if the session is valid, to show reply button or not
func markValidity(rep *Reply, valid bool) {
	rep.ValidSes = valid
	for i := range rep.Replies {
		markValidity(&rep.Replies[i], valid)
	}
}

func threadPageHandler(db *sql.DB, tmpl *template.Template, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/thread/"):]
	threadID, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Error parsing id:", id)
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	thread, err := findThread(db, threadID)
	if err != nil {
		fmt.Println("Find thread error:", err.Error())
		http.Error(w, "Thread not found", http.StatusNotFound)
		return
	}

	usId, usName, validSes := validateSession(db, r)
	for i := range thread.Replies {
		markValidity(&thread.Replies[i], validSes)
	}

	tpd := threadPageData{thread, validSes, usId, usName}
	tmpl.Execute(w, tpd)
}
