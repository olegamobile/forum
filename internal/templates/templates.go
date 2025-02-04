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

	head := "internal/static/templates/header.html"
	foot := "internal/static/templates/footer.html"

	IndexTmpl, err = template.ParseFiles("internal/static/templates/index.html", head, foot)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	ThreadTmpl, err = template.ParseFiles("internal/static/templates/thread.html", head, "internal/static/templates/reply.html", foot)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	LogTmpl, err = template.ParseFiles("internal/static/templates/login.html", head, foot)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	RegisterTmpl, err = template.ParseFiles("internal/static/templates/registerUser.html", head, foot)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
	ErrorTmpl, err = template.ParseFiles("internal/static/templates/error.html", head, foot)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}
}
