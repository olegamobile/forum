package handlers

import (
	"fmt"
	"forum/internal/db"
	"forum/internal/templates"
	"html"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
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
	CategoriesList   []string
	TopTenCategories []string
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

type Reply struct {
	ID            int
	Author        string
	Content       string
	Created       string
	CreatedDay    string
	CreatedTime   string
	Likes         int
	Dislikes      int
	ParentID      int
	Replies       []Reply
	BaseID        int
	ValidSes      bool
	LikedNow      bool
	DislikedNow   bool
	ContentMaxLen int
}

type reaction struct {
	userID  string
	opinion string
}

type threadPageData struct {
	Thread   Thread
	ValidSes bool
	UsrId    string
	UsrNm    string
	LoginURL string
	Images   map[string]string
}

type loginData struct {
	Message1  string
	Message2  string
	ValidSes  bool
	UsrId     string
	UsrNm     string
	ReturnURL string
	LoginURL  string
}

func IndexHandler(w http.ResponseWriter, r *http.Request, msg string) {

	if r.URL.Path != "/" && r.URL.Path != "/forum" && r.URL.Path != "/expired" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	usId, usName, validSes := ValidateSession(r)

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

	categories := strings.Fields(fetchCategories(-1))
	var topTen []string
	if len(categories) < 10 {
		topTen = categories
	} else {
		topTen = categories[:10]
	}

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
		CategoriesList:   categories,
		TopTenCategories: topTen,
	}
	templates.IndexTmpl.Execute(w, data)
}

func AddThreadHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/add" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	if r.Method != http.MethodPost {
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	authID, author, valid := ValidateSession(r)

	if valid && r.Method == http.MethodPost {
		title := html.EscapeString(strings.TrimSpace(r.FormValue("title")))
		content := html.EscapeString(strings.TrimSpace(r.FormValue("content")))
		rawCats := html.EscapeString(strings.ToLower(r.FormValue("categories")))
		if !checkRequestSize(r) {
			goToErrorPage("Request size too large", http.StatusRequestEntityTooLarge, w, r)
			return
		}
		if len(title) > titleMaxLen ||
			len(content) > contentMaxLen ||
			len(rawCats) > categoriesMaxLen ||
			title == "" ||
			content == "" ||
			rawCats == "" { // User may try to force a long or short input
			goToErrorPage("Bad request, input length not supported", http.StatusBadRequest, w, r)
			return
		}

		catsList := removeDuplicates(strings.Fields(cleanString(rawCats)))
		threadUrl := "/"

		threadResult, err := db.DB.Exec(`INSERT INTO posts (author, authorID, title, content) 
										 VALUES (?, ?, ?, ?);`, author, authID, title, content)
		if err != nil {
			fmt.Println("Adding:", err.Error())
			goToErrorPage("Error adding thread", http.StatusInternalServerError, w, r)
			return
		}
		threadID, err := threadResult.LastInsertId()
		if err != nil {
			fmt.Println("Failed to get last insert ID:", err.Error())
		} else {
			threadUrl = fmt.Sprintf("/thread/%d", threadID)
		}

		for _, category := range catsList {

			_, err := db.DB.Exec(`INSERT OR IGNORE INTO categories (name) VALUES (?);`, category)
			if err != nil {
				fmt.Println("Adding:", err.Error())
				goToErrorPage("Error adding categories", http.StatusInternalServerError, w, r)
				return
			}

			_, err = db.DB.Exec(`INSERT OR IGNORE INTO posts_categories (post_id, category_id) 
								 VALUES (?, (SELECT id FROM categories WHERE name=?));`, threadID, category)
			if err != nil {
				fmt.Println("Adding:", err.Error())
				goToErrorPage("Error adding posts-categories relations", http.StatusInternalServerError, w, r)
				return
			}
		}

		errMsg, err := ImageUploadHandler(r, threadID, authID)
		if err != nil {
			fmt.Println(errMsg, err.Error())
			goToErrorPage(errMsg, http.StatusInternalServerError, w, r)
			return
		}

		//easteregg error 418 teapot
		if title == "tea" && content == "tea" && rawCats == "tea" {
			errMsg := `<a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/418">
			I'm a teapot. I refuse to brew coffee!</a>`
			goToErrorPage(errMsg, http.StatusTeapot, w, r)
			return
		}

		http.Redirect(w, r, threadUrl, http.StatusSeeOther)
	}

	if !valid {
		// Session perhaps expired during writing
		http.Redirect(w, r, "/expired", http.StatusSeeOther)
	}
}

func AddReplyHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/reply" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	if r.Method != http.MethodPost {
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	authID, author, valid := ValidateSession(r)

	if valid && r.Method == http.MethodPost {
		content := html.EscapeString(strings.TrimSpace(r.FormValue("content")))
		parId := r.FormValue("parentId") // No int conversion necessary
		baseId := r.FormValue("baseId")  // No int conversion necessary

		if len(content) > contentMaxLen || content == "" { // User may try to force a bad input
			goToErrorPage("Bad request, input length not supported", http.StatusBadRequest, w, r)
			return
		}

		if content != "" {
			_, err := db.DB.Exec(`INSERT INTO posts (base_id, author, authorID, content, parent_id) 
								  VALUES (?, ?, ?, ?, ?);`, baseId, author, authID, content, parId)
			if err != nil {
				fmt.Println("Replying:", err.Error())
				goToErrorPage("Error adding reply", http.StatusInternalServerError, w, r)
				return
			}
		}
		http.Redirect(w, r, "/thread/"+baseId, http.StatusSeeOther)
	}

	if !valid {
		// Session maybe expired during writing
		http.Redirect(w, r, "/expired", http.StatusSeeOther)
	}
}

func likeOrDislike(w http.ResponseWriter, r *http.Request, opinion string) {
	if r.URL.Path != "/like" && r.URL.Path != "/dislike" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	if r.Method != http.MethodPost {
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	threadId := r.FormValue("base_id")
	postId := r.FormValue("post_id")
	userID, _, valid := ValidateSession(r)

	if !valid {
		http.Redirect(w, r, "/thread/"+threadId, http.StatusSeeOther)
		return
	}

	// Try to delete the exact same row from the table (when already liked/disliked)
	res, _ := db.DB.Exec(`DELETE FROM post_reactions 
						  WHERE user_id = ? AND post_id = ? AND reaction_type = ?;`, userID, postId, opinion)

	// Check if any row was deleted
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		fmt.Println("Affected rows checking failed:", err.Error())
	}

	// Add like/dislike: Update with current value on conflict
	if rowsAffected == 0 {
		_, err2 := db.DB.Exec(`INSERT INTO post_reactions (user_id, post_id, reaction_type) 
							   VALUES (?, ?, ?) 
							   ON CONFLICT (user_id, post_id) 
							   DO UPDATE SET reaction_type = excluded.reaction_type;`, userID, postId, opinion)
		if err2 != nil {
			fmt.Println("Adding like or dislike:", err2.Error())
			goToErrorPage("Error adding like or dislike", http.StatusInternalServerError, w, r)
			return
		}
	}

	http.Redirect(w, r, "/thread/"+threadId, http.StatusSeeOther)
}

func LikeHandler(w http.ResponseWriter, r *http.Request) {
	likeOrDislike(w, r, "like")
}

func DislikeHandler(w http.ResponseWriter, r *http.Request) {
	likeOrDislike(w, r, "dislike")
}

// Image upldoad

func checkRequestSize(r *http.Request) bool {
	maxTotalSize := int(20 * 1024 * 1024) // 20 MB
	return r.ContentLength <= int64(maxTotalSize)
}

func ImageUploadHandler(r *http.Request, postID int64, userID string) (string, error) {
	errMsg := ""
	files := r.MultipartForm.File["files"]

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			errMsg = "File cannot be opened."
			return errMsg, err
		}
		if !imageTypeCorrect(fileHeader.Filename) {
			errMsg = "Invalid file type."
			return errMsg, err
		}
		saveImageData(postID, userID, fileHeader, file)
		defer file.Close()
	}
	return "", nil
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/register" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	var loginData loginData
	loginData.UsrId, loginData.UsrNm, loginData.ValidSes = ValidateSession(r)
	loginData.LoginURL = "/login"

	if loginData.ValidSes {
		fmt.Println(loginData.UsrNm + " trying to create a new user while logged-in")
		loginData.Message1 = "Logged in as " + loginData.UsrNm + ". Log out first."
		templates.LogTmpl.Execute(w, loginData)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Handle GET request - show registration form
		templates.RegisterTmpl.Execute(w, loginData)
		return

	case http.MethodPost:
		name := html.EscapeString(r.FormValue("username"))
		email := html.EscapeString(r.FormValue("email"))
		pass := html.EscapeString(r.FormValue("password"))

		//nameOk, passOk := checkString(name), checkString(pass)
		nameOk, passOk := CheckUsername(name), CheckPassword(pass)
		_, emailErr := mail.ParseAddress(email)

		if !nameOk || !passOk {
			fmt.Println("Minimum 5 chars, limited chars")
			loginData.Message2 = "5-25 characters in username and password. Only letters, numbers and symbols allowed."
			templates.RegisterTmpl.Execute(w, loginData)
			return
		}

		if emailErr != nil {
			fmt.Println("Invalid email address")
			loginData.Message2 = "Invalid email address"
			templates.RegisterTmpl.Execute(w, loginData)
			return
		}

		if NameOremailExists(name) {
			fmt.Println("Name already taken")
			loginData.Message2 = "Name already taken"
			templates.RegisterTmpl.Execute(w, loginData)
			return
		}

		if NameOremailExists(email) {
			fmt.Println("Email already taken")
			loginData.Message2 = "Email already taken"
			templates.RegisterTmpl.Execute(w, loginData)
			return
		}

		hashPass, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		userId, err := uuid.NewV4() // Generate a new UUID user id
		if err != nil {
			goToErrorPage("Error generating Id for user", http.StatusInternalServerError, w, r)
			return
		}

		_, err = db.DB.Exec(`INSERT INTO users (id, email, username, password) 
						    VALUES (?, ?, ?, ?);`, userId, email, name, string(hashPass))
		if err != nil {
			fmt.Println("Adding:", err.Error())
			goToErrorPage("Error adding user", http.StatusInternalServerError, w, r)
			return
		}

		// Log user in
		sessionAndToken(&w, r, userId.String(), name)

		http.Redirect(w, r, "/", http.StatusSeeOther)

	default:
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}
}

// logUserInHandler starts a session for the user if login is succesful
func LogUserInHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/loguserin" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	if r.Method != http.MethodPost {
		fmt.Println("Bad method on logUserInHandler:", r.Method)
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	nameOrEmail := r.FormValue("username-or-email")
	pass := r.FormValue("password")
	returnUrl := r.FormValue("return_url")

	var loginData loginData
	loginData.UsrId, loginData.UsrNm, loginData.ValidSes = ValidateSession(r)
	loginData.ReturnURL, loginData.LoginURL = returnUrl, "/login?return_url="+returnUrl

	if loginData.ValidSes {
		fmt.Println(loginData.UsrNm + " trying to log in while already logged-in")
		loginData.Message1 = "Already logged in as " + loginData.UsrNm + ". Log out first."
		templates.LogTmpl.Execute(w, loginData)
		return
	}

	// Checking if user or email exists
	if !NameOremailExists(nameOrEmail) {
		fmt.Println("User not found")
		loginData.Message1 = "Invalid username/email or password"
		templates.LogTmpl.Execute(w, loginData)
		return
	}

	// Get user information and check password
	var storedHashedPass, userID, username string
	query := `SELECT password, id, username FROM users WHERE username = ? OR email = ?`
	err := db.DB.QueryRow(query, nameOrEmail, nameOrEmail).Scan(&storedHashedPass, &userID, &username)
	if err != nil {
		loginData.Message1 = "Invalid username/email or password"
		templates.LogTmpl.Execute(w, loginData)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHashedPass), []byte(pass))
	if err != nil {
		fmt.Println("Password incorrect")
		loginData.Message1 = "Invalid username/email or password"
		templates.LogTmpl.Execute(w, loginData)
		return
	}

	// Remove any old sessions
	deleteSession(w, r, userID)
	// Create new session and token
	sessionAndToken(&w, r, userID, username)

	http.Redirect(w, r, returnUrl, http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/logout" {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	if r.Method != http.MethodPost {
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	_, _, validSes := ValidateSession(r)
	if !validSes {
		fmt.Println("Trying to log out while not logged-in")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		goToErrorPage("No session found", http.StatusBadRequest, w, r)
		return
	}

	// Delete the session from the database
	query := `DELETE FROM sessions WHERE session_token = ?`
	_, err = db.DB.Exec(query, cookie.Value)
	if err != nil {
		goToErrorPage("Failed to log out", http.StatusInternalServerError, w, r)
		return
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour), // Expire immediately
		HttpOnly: true,
	})

	// Determine the previous page
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}

	http.Redirect(w, r, referer, http.StatusSeeOther)
}

// logInHandler handles user clicking log-in link
func LogInHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/login") {
		goToErrorPage("Page does not exist", http.StatusNotFound, w, r)
		return
	}
	if r.Method != http.MethodGet {
		fmt.Println("Disallowed method at logInHandler:", r.Method)
		goToErrorPage("Method not allowed", http.StatusMethodNotAllowed, w, r)
		return
	}

	var loginData loginData
	loginData.UsrId, loginData.UsrNm, loginData.ValidSes = ValidateSession(r)
	loginData.ReturnURL, loginData.LoginURL = r.URL.Query().Get("return_url"), "/login"
	if loginData.ReturnURL == "" {
		loginData.ReturnURL = "/"
	}

	templates.LogTmpl.Execute(w, loginData)
}
