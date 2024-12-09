package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	LikedNow    bool
	DislikedNow bool
}

type reaction struct {
	userID   int
	opinion  string
	postType string
}

type threadPageData struct {
	Thread   Thread
	ValidSes bool
	UsrId    int
	UsrNm    string
}

func addThreadHandler(w http.ResponseWriter, r *http.Request) {
	authID, author, valid := validateSession(r)

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

func addReplyHandler(w http.ResponseWriter, r *http.Request, parentType string) {
	authID, author, valid := validateSession(r)

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

func timeStrings(created string) (string, string, error) {
	createdGoTime, err := time.Parse(time.RFC3339, created) // "created" looks something like this: 2024-12-02T15:44:52Z
	if err != nil {
		return "", "", err
	}

	// Convert to Finnish timezone (UTC+2)
	location, err := time.LoadLocation("Europe/Helsinki")
	if err != nil {
		return "", "", err
	}
	createdGoTime = createdGoTime.In(location)

	day := createdGoTime.Format("2.1.2006")
	time := createdGoTime.Format("15.04") //"15.04.05"

	return day, time, nil
}

func getReplies(rows *sql.Rows, thisID int) []Reply {
	var err error
	var replies []Reply

	for rows.Next() {
		var re Reply
		err2 := rows.Scan(&re.ID, &re.BaseID, &re.Author, &re.Content, &re.Created)
		if err2 != nil {
			fmt.Println("Error reading reply rows for reply:", err2.Error())
			return replies
		}
		re.ParentID = thisID
		re.ParentType = "reply"

		re.CreatedDay, re.CreatedTime, err = timeStrings(re.Created)
		if err != nil {
			return replies
		}

		re.Likes, re.Dislikes = countReactions(re.ID, "reply")

		replies = append(replies, re)
	}

	return replies
}

func recurseReplies(db *sql.DB, this *Reply) {
	selectQueryReplies := `SELECT id, base_id, author, content, created_at FROM replies WHERE parent_id = ? AND parent_type = ?;`
	rows, err := db.Query(selectQueryReplies, this.ID, "reply")
	if err != nil {
		fmt.Println("Error getting replies for reply:", err.Error())
		return
	}
	defer rows.Close()

	replies := getReplies(rows, this.ID)

	if len(replies) != 0 {
		this.Replies = replies
		for i := 0; i < len(this.Replies); i++ {
			recurseReplies(db, &this.Replies[i])
		}
	}
}

func likeOrDislike(w http.ResponseWriter, r *http.Request, opinion string) {
	threadId := r.FormValue("base_id")
	postId := r.FormValue("post_id")
	postType := r.FormValue("post_type")
	userID, _, valid := validateSession(r)

	if !valid {
		http.Redirect(w, r, "/thread/"+threadId, http.StatusSeeOther)
		return
	}

	// Try to delete the exact same row from the table (when already liked/disliked)
	res, _ := db.Exec(`DELETE FROM post_reactions WHERE user_id = ? AND post_id = ? AND post_type = ? AND reaction_type = ?;`, userID, postId, postType, opinion)

	// Check if any row was deleted
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		fmt.Println("Affected rows checking failed:", err.Error())
	}

	// Add like/dislike: Update with current value on conflict
	if rowsAffected == 0 {
		_, err2 := db.Exec(`INSERT INTO post_reactions (user_id, post_id, post_type, reaction_type) VALUES (?, ?, ?, ?) ON CONFLICT (user_id, post_id, post_type) DO UPDATE SET reaction_type = excluded.reaction_type;`, userID, postId, postType, opinion)
		if err2 != nil {
			fmt.Println("Adding like or dislike:", err2.Error())
			http.Error(w, "Error adding like or dislike", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/thread/"+threadId, http.StatusSeeOther)
}

func likeHandler(w http.ResponseWriter, r *http.Request) {
	likeOrDislike(w, r, "like")
}

func dislikeHandler(w http.ResponseWriter, r *http.Request) {
	likeOrDislike(w, r, "dislike")
}

func dataToThread(thread Thread) (Thread, error) {
	var err error
	thread.CreatedDay, thread.CreatedTime, err = timeStrings(thread.Created)
	if err != nil {
		return thread, err
	}
	err = json.Unmarshal([]byte(thread.Categories), &thread.CatsSlice)
	if err != nil {
		fmt.Println(err.Error())
		return thread, err
	}
	thread.Likes, thread.Dislikes = countReactions(thread.ID, "thread")
	thread.BaseID = thread.ID
	return thread, nil
}

func findThread(db *sql.DB, id int) (Thread, error) {
	var thread Thread
	selectQueryThread := `SELECT id, author, title, content, created_at, categories FROM threads WHERE id = ?;`
	err := db.QueryRow(selectQueryThread, id).Scan(&thread.ID, &thread.Author, &thread.Title, &thread.Content, &thread.Created, &thread.Categories)
	if err != nil {
		return thread, err
	}

	thread, err = dataToThread(thread)

	selectQueryReplies := `SELECT id, base_id, author, content, created_at FROM replies WHERE parent_id = ? AND parent_type = ?;`
	rows, err2 := db.Query(selectQueryReplies, thread.ID, "thread")
	if err2 != nil {
		return thread, err2
	}
	defer rows.Close()

	replies := getReplies(rows, thread.ID)

	// Add replies to replies recursively
	for i := 0; i < len(replies); i++ {
		recurseReplies(db, &(replies[i]))
	}

	thread.Replies = replies
	return thread, err
}

// markValidity writes to each reply if the session is valid, to show reply button or not
func markValidity(rep *Reply, valid bool, reactMap map[string]reaction) {
	rep.ValidSes = valid

	id := fmt.Sprintf("reply%d", rep.ID)
	if reactMap[id].postType == "reply" {
		if reactMap[id].opinion == "like" {
			rep.LikedNow = true
		}
		if reactMap[id].opinion == "dislike" {
			rep.DislikedNow = true
		}
	}

	for i := range rep.Replies {
		markValidity(&rep.Replies[i], valid, reactMap)
	}
}

func threadPageHandler(w http.ResponseWriter, r *http.Request) {
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

	usId, usName, validSes := validateSession(r)

	// List liked and disliked posts. Only to colour the buttons.
	selectQueryReplies := `SELECT post_id, post_type, reaction_type FROM post_reactions WHERE user_id = ?;`
	rows, err2 := db.Query(selectQueryReplies, usId)
	if err2 != nil {
		fmt.Println("Error querying reactions:", err2.Error())
		return
	}

	reactionMap := make(map[string]reaction)
	for rows.Next() {
		postID := 0
		postType := ""
		opinion := ""

		err2 := rows.Scan(&postID, &postType, &opinion)
		if err2 != nil {
			fmt.Println("Error reading reaction rows:", err2.Error())
			return
		}

		id := fmt.Sprintf("%s%d", postType, postID)
		reactionMap[id] = reaction{usId, opinion, postType}
	}

	for i := range thread.Replies {
		markValidity(&thread.Replies[i], validSes, reactionMap)
	}

	// Markers for coloring the thread buttons too
	id = fmt.Sprintf("thread%d", thread.ID)
	if reactionMap[id].postType == "thread" {
		if reactionMap[id].opinion == "like" {
			thread.LikedNow = true
		}
		if reactionMap[id].opinion == "dislike" {
			thread.DislikedNow = true
		}
	}

	tpd := threadPageData{thread, validSes, usId, usName}
	threadTmpl.Execute(w, tpd)
}
