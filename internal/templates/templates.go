package templates

import (
	"fmt"
	"text/template"
)

var (
	IndexTmpl    *template.Template
	ThreadTmpl   *template.Template
	LogTmpl      *template.Template
	RegisterTmpl *template.Template
	ErrorTmpl    *template.Template
)

func InitTemplates() {
	var err error
	IndexTmpl, err = template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	ThreadTmpl, err = template.ParseFiles("templates/thread.html", "templates/header.html", "templates/reply.html", "templates/footer.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	LogTmpl, err = template.ParseFiles("templates/login.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	RegisterTmpl, err = template.ParseFiles("templates/registerUser.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	ErrorTmpl, err = template.ParseFiles("templates/error.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
}
