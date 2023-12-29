package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"phrase_bot/data"
	"phrase_bot/handler"
	"phrase_bot/view"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MainTestSuite struct {
	suite.Suite
	DB   data.DB
	Pool *pgxpool.Pool
}

func (suite *MainTestSuite) TestHandleDeletePhrase() {
	h := handler.PhraseHandler{DB: suite.DB}
	e := echo.New()
	e.Renderer = view.EchoTemplate
	phrase, err := data.CreatePhrase(suite.DB, "my new phrase")
	if err != nil {
		suite.T().Fatal("failed to insert fixture phrase", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/phrase/:id/delete/")
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(phrase.Id))

	assert := assert.New(suite.T())

	// Assertions
	if assert.NoError(h.HandleDeletePhrase(c)) {
		assert.Equal(http.StatusFound, rec.Code)
		assert.Equal("/phrase/", rec.Result().Header.Get("Location"))
		phrases, err := data.GetAllPhrases(suite.DB)
		phraseStrings := make([]string, len(*phrases))
		for _, phrase := range *phrases {
			phraseStrings = append(phraseStrings, phrase.Phrase)
		}
		if assert.NoError(err) {
			assert.NotContains(phraseStrings, "my new phrase")
		}
	}
}

func (suite *MainTestSuite) TestHandleCreatePhrase() {
	h := handler.PhraseHandler{DB: suite.DB}
	e := echo.New()
	e.Renderer = view.EchoTemplate
	f := make(url.Values)
	f.Set("phrase", "my new phrase")
	req := httptest.NewRequest(http.MethodPost, "/phrase/", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert := assert.New(suite.T())

	// Assertions
	if assert.NoError(h.HandleCreatePhrase(c)) {
		assert.Equal(http.StatusFound, rec.Code)
		assert.Equal("/phrase/", rec.Result().Header.Get("Location"))
		phrases, err := data.GetAllPhrases(suite.DB)
		phraseStrings := make([]string, len(*phrases))
		for _, phrase := range *phrases {
			phraseStrings = append(phraseStrings, phrase.Phrase)
		}
		if assert.NoError(err) {
			assert.Contains(phraseStrings, "my new phrase")
		}
	}
}

func (suite *MainTestSuite) TestHandlePhraseSearchNoMatch() {
	h := handler.PhraseHandler{DB: suite.DB}
	e := echo.New()
	e.Renderer = view.EchoTemplate
	q := make(url.Values)
	q.Set("search", "nonexistent")
	req := httptest.NewRequest(http.MethodGet, "/phrase/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert := assert.New(suite.T())

	// Assertions
	if assert.NoError(h.HandlePhraseShow(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.NotContains(rec.Body.String(), "JIRA")
	}
}

func (suite *MainTestSuite) TestHandlePhraseSearchMatch() {
	h := handler.PhraseHandler{DB: suite.DB}
	e := echo.New()
	e.Renderer = view.EchoTemplate
	q := make(url.Values)
	q.Set("search", "JIRA")
	req := httptest.NewRequest(http.MethodGet, "/phrase/?"+q.Encode(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert := assert.New(suite.T())

	// Assertions
	if assert.NoError(h.HandlePhraseShow(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Contains(rec.Body.String(), "JIRA")
	}
}

func (suite *MainTestSuite) TestHandlePhraseShowAll() {
	h := handler.PhraseHandler{DB: suite.DB}
	e := echo.New()
	e.Renderer = view.EchoTemplate
	req := httptest.NewRequest(http.MethodGet, "/phrase/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert := assert.New(suite.T())

	// Assertions
	if assert.NoError(h.HandlePhraseShow(c)) {
		assert.Equal(http.StatusOK, rec.Code)
		assert.Contains(rec.Body.String(), "JIRA")
	}
}

func (suite *MainTestSuite) SetupSuite() {
	config := Config{
		DatabaseUrl: os.Getenv("DATABASE_URL"),
	}
	db, err := setupDB(config)
	if err != nil {
		suite.T().Fatal("failed to connect to database", err)
	}
	_, err = db.Exec(context.Background(), "truncate phrase")
	if err != nil {
		suite.T().Fatal("failed to clean out database before test", err)
	}
	_, err = db.Exec(context.Background(), "insert into phrase (phrase) values ($1)", "JIRA is bad")
	if err != nil {
		suite.T().Fatal("failed to insert fixture data", err)
	}
	suite.Pool = db
}

func (suite *MainTestSuite) SetupTest() {
	tx, err := suite.Pool.Begin(context.Background())
	if err != nil {
		suite.T().Fatal("failed to insert fixture data", err)
	}
	suite.DB = tx
}

func (suite *MainTestSuite) TearDownTest() {
	tx := suite.DB.(pgx.Tx)
	tx.Rollback(context.Background())
}

func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}
