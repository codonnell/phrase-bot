package handler

import (
	"net/http"
	"phrase_bot/data"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/slack-go/slack"
)

type SlackHandler struct {
	*pgxpool.Pool
	*slack.Client
}

type Response struct {
	Text         string `json:"text"`
	ResponseType string `json:"response_type"`
}

func (h SlackHandler) HandleInsultJira(c echo.Context) error {
	if strings.TrimSpace(c.FormValue("text")) != "insult jira" {
		response := &Response{Text: "make sure you type \"insult jira\" after the /pf2 command"}
		return c.JSON(http.StatusOK, response)
	}
	phrase, err := data.GetRandomPhrase(h.Pool)
	if err != nil {
		log.Error(err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}
	response := &Response{Text: phrase.Phrase, ResponseType: "in_channel"}
	return c.JSON(http.StatusOK, response)
}
