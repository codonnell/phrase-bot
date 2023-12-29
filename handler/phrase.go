package handler

import (
	"net/http"
	"phrase_bot/data"
	"phrase_bot/types"
	"phrase_bot/view"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type PhraseHandler struct {
	data.DB
}

func (h PhraseHandler) HandlePhraseShow(c echo.Context) error {
	search := c.QueryParam("search")
	var phrases *[]types.Phrase
	var err error
	if search != "" {
		phrases, err = data.SearchPhrases(h.DB, search)
	} else {
		phrases, err = data.GetAllPhrases(h.DB)
	}
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "show_phrases", view.ShowPhrases{PhraseError: "", Phrases: *phrases, Search: search})
}

func (h PhraseHandler) HandleCreatePhrase(c echo.Context) error {
	formPhrase := c.FormValue("phrase")
	if len(strings.TrimSpace(formPhrase)) == 0 {
		phrases, err := data.GetAllPhrases(h.DB)
		if err != nil {
			return err
		}
		return c.Render(http.StatusOK, "show_phrases", view.ShowPhrases{PhraseError: "Phrase cannot be blank", Phrases: *phrases})
	}
	_, err := data.CreatePhrase(h.DB, formPhrase)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/phrase/")
}

func (h PhraseHandler) HandleDeletePhrase(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	err = data.DeletePhrase(h.DB, id)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/phrase/")
}
