package handlers

import (
	"database/sql"
	"fmt"
	"forum/internal/db"
	"forum/internal/templates"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// cleanString removes trailing and leading spaces and punctuation
func cleanString(s string) string {
	result := ""
	for _, char := range strings.TrimSpace(s) {
		if !unicode.IsPunct(char) {
			result += string(char)
		} else {
			result += " "
		}
	}
	return result
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
	time := createdGoTime.Format("15:04") //"15.04.05"

	return day, time, nil
}

// createReplies creates a slice of Replies from database rows
func createReplies(rows *sql.Rows, thisID int) []Reply {
	var err error
	var replies []Reply

	for rows.Next() {
		var re Reply
		err2 := rows.Scan(&re.ID, &re.BaseID, &re.Author, &re.Content, &re.Created)
		if err2 != nil {
			fmt.Println("Error reading reply rows for reply:", err2.Error())
			return replies
		}
		re.ParentID, re.ContentMaxLen = thisID, contentMaxLen

		re.CreatedDay, re.CreatedTime, err = timeStrings(re.Created)
		if err != nil {
			return replies
		}

		re.Likes, re.Dislikes = countReactions(re.ID)

		replies = append(replies, re)
	}

	return replies
}

func recurseReplies(this *Reply) {
	selectQueryReplies := `SELECT id, base_id, author, content, created_at FROM posts WHERE parent_id = ?;`
	rows, err := db.DB.Query(selectQueryReplies, this.ID)
	if err != nil {
		fmt.Println("Error getting replies for reply:", err.Error())
		return
	}
	defer rows.Close()

	replies := createReplies(rows, this.ID)

	if len(replies) != 0 {
		this.Replies = replies
		for i := 0; i < len(this.Replies); i++ {
			recurseReplies(&this.Replies[i])
		}
	}
}

func dataToThread(thread Thread) (Thread, error) {
	var err error
	thread.CreatedDay, thread.CreatedTime, err = timeStrings(thread.Created)
	if err != nil {
		return thread, err
	}
	// err = json.Unmarshal([]byte(thread.Categories),
	thread.CatsSlice = strings.Fields(thread.Categories)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return thread, err
	// }

	thread.Likes, thread.Dislikes = countReactions(thread.ID)
	thread.BaseID, thread.ContentMaxLen = thread.ID, contentMaxLen
	return thread, nil
}

func findThread(id int) (Thread, error) {
	var thread Thread
	selectQueryThread := `SELECT id, author, title, content, created_at FROM posts WHERE id = ?;`
	err := db.DB.QueryRow(selectQueryThread, id).Scan(&thread.ID, &thread.Author, &thread.Title, &thread.Content, &thread.Created)
	if err != nil {
		return thread, err
	}
	thread.Categories = fetchCategories(id)

	thread, err = dataToThread(thread)
	selectQueryReplies := `SELECT id, base_id, author, content, created_at FROM posts WHERE parent_id = ?;`
	rows, err2 := db.DB.Query(selectQueryReplies, thread.ID)
	if err2 != nil {
		return thread, err2
	}
	defer rows.Close()

	replies := createReplies(rows, thread.ID)

	// Add replies to replies recursively
	for i := 0; i < len(replies); i++ {
		recurseReplies(&(replies[i]))
	}

	thread.Replies = replies
	return thread, err
}

// markValidity writes to each reply if the session is valid, to show reply button or not
func markValidity(rep *Reply, valid bool, reactMap map[int]reaction) {
	rep.ValidSes = valid

	if reactMap[rep.ID].opinion == "like" {
		rep.LikedNow = true
	}
	if reactMap[rep.ID].opinion == "dislike" {
		rep.DislikedNow = true
	}

	for i := range rep.Replies {
		markValidity(&rep.Replies[i], valid, reactMap)
	}
}

// threadPageHandler handles request from /thread/
func ThreadPageHandler(w http.ResponseWriter, r *http.Request) {

	if !strings.HasPrefix(r.URL.Path, "/thread/") {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	id := r.URL.Path[len("/thread/"):]

	threadID, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Error parsing id:", id)
		goToErrorPage("Invalid thread ID", http.StatusBadRequest, w, r)
		return
	}

	thread, err := findThread(threadID)
	if err != nil {
		fmt.Println("Find thread error:", err.Error())
		goToErrorPage("Thread not found", http.StatusNotFound, w, r)
		return
	}
	// Get linked images for the thread
	images, err := getThreadImageURL(threadID)
	if err != nil {
		fmt.Println("Error finding images:", err.Error())
		goToErrorPage("Error loading images", http.StatusInternalServerError, w, r)
		return
	}

	usId, usName, validSes := ValidateSession(r)

	// List liked and disliked posts. Only to colour the buttons.
	selectQueryReplies := `SELECT post_id, reaction_type FROM post_reactions WHERE user_id = ?;`
	rows, err2 := db.DB.Query(selectQueryReplies, usId)
	if err2 != nil {
		fmt.Println("Error querying reactions:", err2.Error())
		return
	}

	reactionMap := make(map[int]reaction)
	for rows.Next() {
		postID := 0
		opinion := ""

		err2 := rows.Scan(&postID, &opinion)
		if err2 != nil {
			fmt.Println("Error reading reaction rows:", err2.Error())
			return
		}

		reactionMap[postID] = reaction{usId, opinion}
	}

	for i := range thread.Replies {
		markValidity(&thread.Replies[i], validSes, reactionMap)
	}

	// Markers for coloring the thread buttons too
	if reactionMap[thread.ID].opinion == "like" {
		thread.LikedNow = true
	}
	if reactionMap[thread.ID].opinion == "dislike" {
		thread.DislikedNow = true
	}

	loginUrl := "/login?return_url=" + r.URL.Path
	tpd := threadPageData{thread, validSes, usId, usName, loginUrl, images}
	templates.ThreadTmpl.Execute(w, tpd)
}
