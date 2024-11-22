package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
	//_ "modernc.org/sqlite"
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

type Reply struct {
	ID         int
	Author     string
	Content    string
	Created    string
	Likes      int
	Dislikes   int
	ParentID   int
	ParentType string
	Replies    []Reply
	BaseID     int
}

type PageData struct {
	Threads []Thread
}

func makeTables(db *sql.DB) {
	// Create threads table if it doesn't exist
	createThreadsTableQuery := `
	CREATE TABLE IF NOT EXISTS threads (
		id INTEGER PRIMARY KEY AUTOINCREMENT,		
		author TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		categories JSON,
		likes INTEGER DEFAULT 0,
		dislikes INTEGER DEFAULT 0
	);`
	if _, err := db.Exec(createThreadsTableQuery); err != nil {
		fmt.Println("Error creating threads table:", err)
		return
	}

	// Create replies table if it doesn't exist
	createRepliesTableQuery := `
		CREATE TABLE IF NOT EXISTS replies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			base_id INTEGER DEFAULT 0,
			author TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			likes INTEGER DEFAULT 0,
			dislikes INTEGER DEFAULT 0,
			parent_id INTEGER NOT NULL,
			parent_type TEXT
		);`
	if _, err := db.Exec(createRepliesTableQuery); err != nil {
		fmt.Println("Error creating replies table:", err)
		return
	}
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

func addThreadHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		author := "Auteur"
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
			_, err := db.Exec(`INSERT INTO threads (author, title, content, categories) VALUES (?, ?, ?, ?);`, author, title, content, string(catsJson)) //categories <=> catsJson, Ongelma JSOnin kanssa
			if err != nil {
				fmt.Println("Adding:", err.Error())
				http.Error(w, "Error adding thread", http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func addReplyHandler(db *sql.DB, w http.ResponseWriter, r *http.Request, parentType string) {
	if r.Method == http.MethodPost {
		author := "Auteur"
		content := strings.TrimSpace(r.FormValue("content"))
		parId := r.FormValue("parentId") // No int conversion necessary
		baseId := r.FormValue("baseId")  // No int conversion necessary

		if content != "" {
			_, err := db.Exec(`INSERT INTO replies (base_id, author, content, parent_id, parent_type) VALUES (?, ?, ?, ?, ?);`, baseId, author, content, parId, parentType)
			if err != nil {
				fmt.Println("Replying:", err.Error())
				http.Error(w, "Error adding reply", http.StatusInternalServerError)
				return
			}
		}
		http.Redirect(w, r, "/thread/"+baseId, http.StatusSeeOther)
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
		replies = append(replies, re)
	}

	if len(replies) != 0 {

		this.Replies = replies
		for i := 0; i < len(this.Replies); i++ {
			recurseReplies(db, &this.Replies[i])
		}
	}
}

func findThread(db *sql.DB, id int) (Thread, error) {
	var thread Thread
	selectQueryThread := `SELECT id, author, title, content, created_at, categories, likes, dislikes FROM threads WHERE id = ?;`
	err1 := db.QueryRow(selectQueryThread, id).Scan(&thread.ID, &thread.Author, &thread.Title, &thread.Content, &thread.Created, &thread.Categories, &thread.Likes, &thread.Dislikes)

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

	thread.Replies = replies
	thread.BaseID = thread.ID
	return thread, err1
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

	tmpl.Execute(w, thread)
}

func deleteThreadHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id := r.FormValue("id")
		_, err := db.Exec(`DELETE FROM threads WHERE id = ?;`, id)
		if err != nil {
			http.Error(w, "Error deleting thread", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
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

	data := PageData{Threads: threads}
	tmpl.Execute(w, data)
}

func main() {
	// Open database connection
	db, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	makeTables(db)

	// Initialize templates
	indexTmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	threadTmpl, err := template.ParseFiles("static/thread.html", "static/reply.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// Set up routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexHandler(db, indexTmpl, w, r)
	})
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		addThreadHandler(db, w, r)
	})
	http.HandleFunc("/reply", func(w http.ResponseWriter, r *http.Request) {
		addReplyHandler(db, w, r, "thread")
	})
	http.HandleFunc("/replytoreply", func(w http.ResponseWriter, r *http.Request) {
		addReplyHandler(db, w, r, "reply")
	})
	http.HandleFunc("/thread/", func(w http.ResponseWriter, r *http.Request) {
		threadPageHandler(db, threadTmpl, w, r)
	})
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		deleteThreadHandler(db, w, r)
	})
	http.HandleFunc("/static/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/styles.css")
	})

	// Start the server
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}
