package handler

import (
	"io"
	"net/http"
	"phrase_bot/data"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

type SlackHandler struct {
	*pgxpool.Pool
	*slack.Client
	SigningSecret string
}

type Response struct {
	Text         string `json:"text"`
	ResponseType string `json:"response_type"`
}

func (h SlackHandler) HandleInsultJira(c echo.Context) error {
	sv, err := slack.NewSecretsVerifier(c.Request().Header, h.SigningSecret)
	if err != nil {
		return err
	}
	c.Request().Body = io.NopCloser(io.TeeReader(c.Request().Body, &sv))
	s, err := slack.SlashCommandParse(c.Request())
	if err != nil {
		return err
	}
	err = sv.Ensure()
	if err != nil {
		return err
	}
	if s.Text != "insult jira" && s.Command != "/pf2" {
		response := &Response{Text: "make sure you type \"insult jira\" after the /pf2 command"}
		return c.JSON(http.StatusOK, response)
	}
	phrase, err := data.GetRandomPhrase(h.Pool)
	if err != nil {
		return err
	}
	response := &Response{Text: phrase.Phrase, ResponseType: "in_channel"}
	return c.JSON(http.StatusOK, response)
}
