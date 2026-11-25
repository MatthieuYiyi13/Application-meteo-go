package helpers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

var templates *template.Template

func LoadTemplates() {
	templates = template.Must(template.ParseGlob("templates/*.html"))
}
func Render(w http.ResponseWriter, status int, name string, data any) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("Erreur template (%s) : %v\n", name, err)
		RenderError(w, http.StatusInternalServerError, "Erreur chargement de la page")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

type ErrorPage struct {
	Code    int
	Message string
}

func RenderError(w http.ResponseWriter, status int, message string) {
	Render(w, status, "error", ErrorPage{Code: status, Message: message})
}
