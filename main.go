package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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

type SignData struct {
	Message1 string
	Message2 string
}

var signData SignData

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

	// Create users table if it doesn't exist
	createUsersTableQuery := `
		CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,  -- Store hashed passwords (bonus task)
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createUsersTableQuery); err != nil {
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

func checkString(s string) bool {
	if len(s) < 5 {
		return false
	}
	for _, ch := range s {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && !unicode.IsSymbol(ch) {
			fmt.Println("Only letters, numbers and symbols allowed:", string(ch))
			return false
		}
	}
	return true
}

func nameExists(db *sql.DB, username string) bool {
	selectQueryName := `SELECT username FROM users WHERE username = ?`
	err1 := db.QueryRow(selectQueryName, username).Scan(&username)
	if err1 == nil {
		return true
	}
	return false
}

func emailExists(db *sql.DB, mail string) bool {
	selectQueryMail := `SELECT email FROM users WHERE email = ?`
	err2 := db.QueryRow(selectQueryMail, mail).Scan(&mail)
	if err2 == nil {
		return true
	}
	return false
}

func addUserHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("username")
		email := r.FormValue("email")
		pass := r.FormValue("password")

		n, p := checkString(name), checkString(pass)
		_, e := mail.ParseAddress(email)

		if !n || !p {
			fmt.Println("Minimum 5 chars, limited chars")
			signData.Message2 = "Minimum 5 characters in usename and password. Only letters, numbers and symbols allowed."
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if e != nil {
			fmt.Println("Invalid email address")
			signData.Message2 = "Invalid email address"
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if nameExists(db, name) {
			fmt.Println("Name already taken")
			signData.Message2 = "Name already taken"
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if emailExists(db, email) {
			fmt.Println("Email already taken")
			signData.Message2 = "Email already taken"
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		hashPass, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

		_, err := db.Exec(`INSERT INTO users (email, username, password) VALUES (?, ?, ?);`, email, name, string(hashPass))
		if err != nil {
			fmt.Println("Adding:", err.Error())
			http.Error(w, "Error adding user", http.StatusInternalServerError)
			return
		}

		signData.Message2 = ""

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func logUserHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("username")
		email := r.FormValue("email")
		pass := r.FormValue("password")

		if !nameExists(db, name) && !emailExists(db, email) {
			fmt.Println("User not found")
			signData.Message1 = "User not found"
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		storedHashedPassword := ""
		if nameExists(db, name) {
			// retrieve hashed pass with name
		} else {
			// retrieve hashed pass with email
		}

		err := bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(pass))

		if err != nil {
			fmt.Println("Password incorrect")
			signData.Message1 = "Password incorrect"
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		// sign in
		// set a cookie
		// Check session token on every request
		// Log out: delete the cookie and remove it from the database

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
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
	indexTmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	threadTmpl, err := template.ParseFiles("templates/thread.html", "templates/reply.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	signTmpl, err := template.ParseFiles("templates/signin.html")
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
	http.HandleFunc("/static/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/styles.css")
	})
	http.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		signTmpl.Execute(w, signData)
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		logUserHandler(db, w, r)
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		addUserHandler(db, w, r)
	})

	// Start the server
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}
