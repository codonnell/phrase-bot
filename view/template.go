package view

import (
	"html/template"
	"io"
	"phrase_bot/types"

	"github.com/labstack/echo/v4"
)

type ShowPhrases struct {
	PhraseError string
	Search      string
	Phrases     []types.Phrase
}

type Template struct {
	templates *template.Template
}

var EchoTemplate *Template = &Template{
	templates: template.Must(template.ParseGlob("templates/*.html")),
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
