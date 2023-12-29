package handler

import (
	"io"
	"net/http"
	"phrase_bot/data"

	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

type SlackHandler struct {
	data.DB
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
	if s.Text != "insult jira" {
		response := &Response{Text: "make sure you type \"insult jira\" after the /pf2 command"}
		return c.JSON(http.StatusOK, response)
	}
	phrase, err := data.GetRandomPhrase(h.DB)
	if err != nil {
		return err
	}
	response := &Response{Text: phrase.Phrase, ResponseType: "in_channel"}
	return c.JSON(http.StatusOK, response)
}
