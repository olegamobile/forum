package main

import (
	"database/sql"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"
)

type Thread struct {
	ID            int
	Author        string
	Title         string
	Content       string
	Created       string
	CreatedDay    string
	CreatedTime   string
	Categories    string
	CatsSlice     []string
	Likes         int
	Dislikes      int
	RepliesN      int
	Replies       []Reply
	BaseID        int
	LikedNow      bool
	DislikedNow   bool
	ContentMaxLen int
}

type PageData struct {
	Threads          []Thread
	ValidSes         bool
	UsrId            string
	UsrNm            string
	Message          string
	Selection        string
	Search           string
	Multisearch      string
	TitleMaxLen      int
	ContentMaxLen    int
	CategoriesMaxLen int
	LoginURL         string
}

type errorData struct {
	Message   string
	ErrorCode int
	ValidSes  bool
	UsrNm     string
	LoginURL  string
}

const (
	titleMaxLen      int = 200
	contentMaxLen    int = 3000
	categoriesMaxLen int = 200
)

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

// fetchThreads
func fetchThreads(rowsThreads *sql.Rows) ([]Thread, error) {
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

// getMultipleSearch returns a search query that looks for either any or all matches to the search terms
func getMultipleSearch(multisearch string, searches []string) string {
	query := ""
	searchesCount := len(searches)
	for i := range searches {
		searches[i] = "'" + searches[i] + "'"
	}

	if multisearch == "any" {
		query = fmt.Sprintf(`SELECT DISTINCT p.id, p.author, p.title, p.content, p.created_at, p.categories FROM posts p, json_each(p.categories) WHERE LOWER(json_each.value) IN (%s);`, strings.Join(searches, ", "))
	} else {
		query = fmt.Sprintf(`SELECT p.id, p.author, p.title, p.content, p.created_at, p.categories FROM posts p, json_each(p.categories) WHERE LOWER(json_each.value) IN (%s) GROUP BY p.id HAVING COUNT(DISTINCT LOWER(json_each.value)) = %v;`, strings.Join(searches, ", "), searchesCount)
		// HAVING COUNT to have equal number of matching categories to search terms
	}

	// json_each expands the JSON array in p.categories into rows, allowing filtering by category strings
	// DISTINCT to avoid duplicates in case of repeated category in post

	return query
}

// removeDuplicates returns a slice of strings without duplicates
func removeDuplicates(searches []string) []string {
	result := []string{}
	for i := 0; i < len(searches); i++ {
		found := false
		for j := 0; j < len(result); j++ {
			if searches[i] == result[j] {
				found = true
			}
		}
		if !found {
			result = append(result, searches[i])
		}
	}
	return result
}

func findThreads(r *http.Request) ([]Thread, string, string, string, error) {

	usId, _, validSes := validateSession(r)
	selection := r.FormValue("todisplay")
	search := r.FormValue("usersearch")
	multisearch := r.FormValue("multisearch")

	// Find all threads by default
	selectQuery := `SELECT id, author, title, content, created_at, categories FROM posts WHERE title != "";`
	rowsThreads, err := db.Query(selectQuery)
	if err != nil {
		fmt.Println("findThreads selectQuery failed", err.Error())
		return nil, selection, search, multisearch, err
	}
	defer rowsThreads.Close()

	if r.Method == http.MethodPost { // If user did a POST, session should be valid

		if validSes && (r.FormValue("updatesel") == "update" && (selection == "created" || selection == "liked" || selection == "disliked")) {
			switch selection {
			case "created":
				selectQuery = `SELECT p.id, p.author, p.title, p.content, p.created_at, p.categories FROM posts p WHERE title != "" AND authorID = ?;`
			case "liked":
				selectQuery = `SELECT p.id, p.author, p.title, p.content, p.created_at, p.categories FROM posts p JOIN post_reactions pr ON p.id = pr.post_id WHERE p.title != "" AND pr.reaction_type = 'like' AND pr.user_id = ?`
			case "disliked":
				selectQuery = `SELECT p.id, p.author, p.title, p.content, p.created_at, p.categories FROM posts p JOIN post_reactions pr ON p.id = pr.post_id WHERE p.title != "" AND pr.reaction_type = 'dislike' AND pr.user_id = ?`
			}

			rowsThreads, err = db.Query(selectQuery, usId)
			if err != nil {
				fmt.Println("findThreads selectQuery to filter selected failed", err.Error())
				return nil, selection, search, multisearch, err
			}
			defer rowsThreads.Close()

			search = ""
		}

		if r.FormValue("serchcat") == "search" {
			searches := strings.Fields(cleanString(html.EscapeString(strings.ToLower(search))))

			selectQuery := getMultipleSearch(multisearch, searches)

			if len(searches) > 0 {
				rowsThreads, err = db.Query(selectQuery)

				if err != nil {
					fmt.Println("findThreads selectQuery to search categories failed", err.Error())
					return nil, selection, search, multisearch, err
				}
				defer rowsThreads.Close()
			}

			selection = ""
		}

		if r.FormValue("reset") == "reset" {
			search = ""
			selection = ""
		}
	}

	var threads []Thread
	threads, err = fetchThreads(rowsThreads)

	return threads, selection, search, multisearch, err
}

// fetchReplies returns replies based on post ID
func fetchReplies(thisID int) ([]Reply, error) {
	selectQueryReplies := `SELECT id, base_id, author, content, created_at FROM posts WHERE base_id = ? AND title = '';` // All replies
	rows, err := db.Query(selectQueryReplies, thisID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	replies := createReplies(rows, thisID)
	return replies, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request, msg string) {

	if r.URL.Path != "/" && r.URL.Path != "/forum" && r.URL.Path != "/expired" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	usId, usName, validSes := validateSession(r)

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	threads, selection, search, multisearch, err := findThreads(r)

	if err != nil {
		goToErrorPage("Error fetching threads", http.StatusInternalServerError, w, r)
		return
	}

	for i, th := range threads {
		replies, err := fetchReplies(th.ID)
		if err != nil {
			fmt.Println("Error fetching replies:", err.Error())
			goToErrorPage("Error fetching replies", http.StatusInternalServerError, w, r)
			return
		}
		threads[i].Replies = replies
		threads[i].RepliesN = len(replies)
	}

	sortByRecentInteraction(&threads, w, r)

	data := PageData{
		Threads:          threads,
		ValidSes:         validSes,
		UsrId:            usId,
		UsrNm:            usName,
		Message:          msg,
		Selection:        selection,
		Search:           search,
		Multisearch:      multisearch,
		TitleMaxLen:      titleMaxLen,
		ContentMaxLen:    contentMaxLen,
		CategoriesMaxLen: categoriesMaxLen,
		LoginURL:         "/login",
	}
	indexTmpl.Execute(w, data)
}

// newestReply finds time of the  newest reply in the tree
func newestReply(this *Reply, w http.ResponseWriter, r *http.Request) time.Time {
	thisTime, err := time.Parse(time.RFC3339, this.Created)
	if err != nil {
		fmt.Println("Error parsing date:", err.Error())
		goToErrorPage("Error parsing date", http.StatusInternalServerError, w, r)
		return thisTime
	}

	if len(this.Replies) == 0 {
		return thisTime
	} else {
		newest := thisTime
		for _, rep := range this.Replies {
			repTime := newestReply(&rep, w, r)
			if repTime.After(newest) {
				newest = repTime
			}
		}
		return newest
	}
}

// getThreadTime finds the time the thread got its most recent post
func getThreadTime(th *Thread, w http.ResponseWriter, r *http.Request) time.Time {
	threadTime, err := time.Parse(time.RFC3339, th.Created) // "created" looks something like this: 2024-12-02T15:44:52Z
	if err != nil {
		fmt.Println("Error parsing date:", err.Error())
		goToErrorPage("Error parsing date", http.StatusInternalServerError, w, r)
		return threadTime
	}

	for _, rep := range th.Replies {
		replyTime := newestReply(&rep, w, r)
		if replyTime.After(threadTime) {
			threadTime = replyTime
		}
	}

	return threadTime
}

// sortByRecentInteraction sorts the threads by most recent post within the thread
func sortByRecentInteraction(threads *[]Thread, w http.ResponseWriter, r *http.Request) {
	for i := 0; i < len(*threads)-1; i++ {
		for j := i + 1; j < len(*threads); j++ {
			time1 := getThreadTime(&(*threads)[i], w, r)
			time2 := getThreadTime(&(*threads)[j], w, r)
			if time1.Before(time2) {
				(*threads)[i], (*threads)[j] = (*threads)[j], (*threads)[i]
			}
		}
	}
}

func goToErrorPage(msg string, code int, w http.ResponseWriter, r *http.Request) {
	_, usName, validSes := validateSession(r)
	errData := errorData{msg, code, validSes, usName, "/login"}
	w.WriteHeader(code)
	errorTmpl.Execute(w, errData)
}
