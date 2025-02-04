package handlers

import (
	"database/sql"
	"fmt"
	"forum/internal/db"
	"forum/internal/templates"
	"html"
	"net/http"
	"strings"
	"time"
)

func countReactions(id int) (int, int) {

	reactionsQuery := `SELECT reaction_type, COUNT(*) AS count FROM post_reactions WHERE post_id = ? GROUP BY reaction_type;`
	rows, err := db.DB.Query(reactionsQuery, id)
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

func fetchCategories(postId int) string {
	var selectQuery string
	if postId == -1 {
		selectQuery = `SELECT categories.name FROM posts_categories JOIN categories ON posts_categories.category_id = categories.id GROUP BY posts_categories.category_id ORDER BY COUNT(posts_categories.post_id) DESC;`
	} else {
		selectQuery = `SELECT categories.name AS category FROM categories JOIN posts_categories ON posts_categories.category_id = categories.id WHERE post_id = ?;`
	}

	rowsCategories, err := db.DB.Query(selectQuery, postId)
	if err != nil {
		fmt.Println("fetchCategories selectQuery failed", err.Error())
		return ""
	}
	defer rowsCategories.Close()

	var category, categories string
	for rowsCategories.Next() {
		err = rowsCategories.Scan(&category)
		if err != nil {
			fmt.Println("Error reading category:", err.Error())
			return categories
		}
		categories += category + " "
	}
	if err := rowsCategories.Err(); err != nil {
		fmt.Println("Error iterating through rows:", err.Error())
	}
	if len(categories) > 0 {
		categories = categories[:len(categories)-1]
	}
	return categories
}

// fetchThreads
func fetchThreads(rowsThreads *sql.Rows) ([]Thread, error) {
	var threads []Thread
	for rowsThreads.Next() {
		var th Thread
		err := rowsThreads.Scan(&th.ID, &th.Author, &th.Title, &th.Content, &th.Created)
		if err != nil {
			fmt.Println("fetchThreads rows scanning:", err.Error())
			return nil, err
		}
		th.Categories = fetchCategories(th.ID)
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
		// Should we add escaping here ^^^ or somewhere else?
	}

	if multisearch == "any" {
		query = fmt.Sprintf(`SELECT DISTINCT p.id, p.author, p.title, p.content, p.created_at FROM posts p JOIN posts_categories pc ON pc.post_id = p.id JOIN categories cats ON cats.id = pc.category_id WHERE cats.name IN (%s);`, strings.Join(searches, ", "))
	} else {
		query = fmt.Sprintf(`SELECT DISTINCT p.id, p.author, p.title, p.content, p.created_at FROM posts p JOIN posts_categories pc ON pc.post_id = p.id JOIN categories cats ON cats.id = pc.category_id WHERE cats.name IN (%s) GROUP BY p.id, p.author, p.title, p.content, p.created_at HAVING COUNT(DISTINCT cats.name) = %v;`, strings.Join(searches, ", "), searchesCount)
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

	usId, _, validSes := ValidateSession(r)
	selection := r.FormValue("todisplay")
	search := r.FormValue("usersearch")
	multisearch := r.FormValue("multisearch")

	// Find all threads by default
	selectQuery := `SELECT id, author, title, content, created_at FROM posts WHERE title != "";`
	rowsThreads, err := db.DB.Query(selectQuery)
	if err != nil {
		fmt.Println("findThreads selectQuery failed", err.Error())
		return nil, selection, search, multisearch, err
	}
	defer rowsThreads.Close()

	if r.Method == http.MethodPost { // If user did a POST, session should be valid

		if validSes && (r.FormValue("updatesel") == "update" && (selection == "created" || selection == "liked" || selection == "disliked")) {
			switch selection {
			case "created":
				selectQuery = `SELECT p.id, p.author, p.title, p.content, p.created_at FROM posts p WHERE title != "" AND authorID = ?;`
			case "liked":
				selectQuery = `SELECT p.id, p.author, p.title, p.content, p.created_at FROM posts p JOIN post_reactions pr ON p.id = pr.post_id WHERE p.title != "" AND pr.reaction_type = 'like' AND pr.user_id = ?`
			case "disliked":
				selectQuery = `SELECT p.id, p.author, p.title, p.content, p.created_at FROM posts p JOIN post_reactions pr ON p.id = pr.post_id WHERE p.title != "" AND pr.reaction_type = 'dislike' AND pr.user_id = ?`
			}

			rowsThreads, err = db.DB.Query(selectQuery, usId)
			if err != nil {
				fmt.Println("findThreads selectQuery to filter selected failed", err.Error())
				return nil, selection, search, multisearch, err
			}
			defer rowsThreads.Close()

			search = ""
		}

		if r.FormValue("searchcat") == "search" {
			searches := strings.Fields(cleanString(html.EscapeString(strings.ToLower(search))))

			selectQuery := getMultipleSearch(multisearch, searches)

			if len(searches) > 0 {
				rowsThreads, err = db.DB.Query(selectQuery)

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
	rows, err := db.DB.Query(selectQueryReplies, thisID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	replies := createReplies(rows, thisID)
	return replies, nil
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
	_, usName, validSes := ValidateSession(r)
	errData := errorData{msg, code, validSes, usName, "/login"}
	w.WriteHeader(code)
	templates.ErrorTmpl.Execute(w, errData)
}
