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

type Post struct {
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
}

type Reply struct {
	ID       int
	Author   string
	Content  string
	Created  string
	Likes    int
	Dislikes int
	ParentID int
}

type PageData struct {
	Posts []Post
}

func makeTables(db *sql.DB) {
	// Create posts table if it doesn't exist
	createPostsTableQuery := `
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		author TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		categories JSON,
		likes INTEGER DEFAULT 0,
		dislikes INTEGER DEFAULT 0
	);`
	if _, err := db.Exec(createPostsTableQuery); err != nil {
		fmt.Println("Error creating posts table:", err)
		return
	}

	// Create replies table if it doesn't exist
	createRepliesTableQuery := `
		CREATE TABLE IF NOT EXISTS replies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			author TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			likes INTEGER DEFAULT 0,
			dislikes INTEGER DEFAULT 0,
			parent_id INTEGER NOT NULL
		);`
	if _, err := db.Exec(createRepliesTableQuery); err != nil {
		fmt.Println("Error creating replies table:", err)
		return
	}
}

func fetchPosts(db *sql.DB) ([]Post, error) {
	selectQuery := `SELECT id, author, title, content, created_at, categories, likes, dislikes FROM posts;`
	rows, err := db.Query(selectQuery)
	if err != nil {
		fmt.Println("fetchPosts selectQuery failed", err.Error())
		return nil, err
	}
	defer rows.Close()
	//fmt.Println(rows)

	var posts []Post
	for rows.Next() {
		var po Post

		err := rows.Scan(&po.ID, &po.Author, &po.Title, &po.Content, &po.Created, &po.Categories, &po.Likes, &po.Dislikes)
		if err != nil {
			fmt.Println("fetchPosts rows scanning:", err.Error())
			return nil, err
		}
		posts = append(posts, po)
	}
	return posts, nil
}

func fetchReplies(db *sql.DB) ([]Reply, error) {
	selectQuery := `SELECT id, author, content, created_at, likes, dislikes, parent_id FROM replies;`
	rows, err := db.Query(selectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []Reply
	for rows.Next() {
		var re Reply
		err := rows.Scan(&re.ID, &re.Author, &re.Content, &re.Created, &re.Likes, &re.Dislikes, &re.ParentID)
		if err != nil {
			return nil, err
		}
		replies = append(replies, re)
	}
	return replies, nil
}

func addPostHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		author := "Auteur"
		title := r.FormValue("title")
		content := r.FormValue("content")
		rawCats := r.FormValue("categories")

		cleanCats := ""
		for _, char := range strings.TrimSpace(rawCats) {
			if !unicode.IsPunct(char) {
				cleanCats += string(char)
			}
		}
		catsJson, _ := json.Marshal(strings.Fields(cleanCats))
		//fmt.Println(string(catsJson))

		_, err := db.Exec(`INSERT INTO posts (author, title, content, categories) VALUES (?, ?, ?, ?);`, author, title, content, string(catsJson)) //categories <=> catsJson, Ongelma JSOnin kanssa
		if err != nil {
			fmt.Println("Adding:", err.Error())
			http.Error(w, "Error adding post", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func findPost(db *sql.DB, id int) (Post, error) {
	var post Post
	selectQueryPost := `SELECT id, author, title, content, created_at, categories, likes, dislikes FROM posts WHERE id = ?;`
	err1 := db.QueryRow(selectQueryPost, id).Scan(&post.ID, &post.Author, &post.Title, &post.Content, &post.Created, &post.Categories, &post.Likes, &post.Dislikes)

	selectQueryReplies := `SELECT id, author, content, created_at, likes, dislikes FROM replies WHERE parent_id = ?;`
	rows, err2 := db.Query(selectQueryReplies, post.ID)
	if err2 != nil {
		return post, err2
	}
	defer rows.Close()

	var replies []Reply
	for rows.Next() {
		var re Reply
		err3 := rows.Scan(&re.ID, &re.Author, &re.Content, &re.Created, &re.Likes, &re.Dislikes)
		if err3 != nil {
			return post, err3
		}
		replies = append(replies, re)
	}
	post.Replies = replies

	return post, err1
}

func postPageHandler(db *sql.DB, tmpl *template.Template, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/post/"):]
	postID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := findPost(db, postID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	tmpl.Execute(w, post)
}

func deletePostHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id := r.FormValue("id")
		_, err := db.Exec(`DELETE FROM todos WHERE id = ?;`, id)
		if err != nil {
			http.Error(w, "Error deleting task", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func indexHandler(db *sql.DB, tmpl *template.Template, w http.ResponseWriter, r *http.Request) {
	posts, err := fetchPosts(db)
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	replies, err := fetchReplies(db)
	if err != nil {
		http.Error(w, "Error fetching replies", http.StatusInternalServerError)
		return
	}

	for i, po := range posts {
		for _, re := range replies {
			if po.ID == re.ParentID {
				posts[i].RepliesN++
			}
		}
	}

	data := PageData{Posts: posts}
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
	postTmpl, err := template.ParseFiles("static/post.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	// Set up routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexHandler(db, indexTmpl, w, r)
	})
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		addPostHandler(db, w, r)
	})
	http.HandleFunc("/post/", func(w http.ResponseWriter, r *http.Request) {
		postPageHandler(db, postTmpl, w, r)
	})
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		deletePostHandler(db, w, r)
	})
	http.HandleFunc("/static/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/styles.css")
	})

	// Start the server
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}
