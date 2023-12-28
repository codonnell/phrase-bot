package main

import (
	"context"
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"phrase_bot/data"
	"phrase_bot/handler"
	"phrase_bot/view"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/slack-go/slack"
)

func setupDB(config Config) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), config.DatabaseUrl)
	if err != nil {
		log.Fatalf("Unable to create dbpoolection pool: %v\n", err)
		os.Exit(1)
	}

	createTable := `create table if not exists phrase (id serial primary key, phrase text not null, inserted_at timestamptz not null default now())`
	_, err = dbpool.Exec(context.Background(), createTable)
	if err != nil {
		log.Fatal("create table error", err)
		return nil, err
	}
	_, err = dbpool.Exec(context.Background(), "create index if not exists phrase_search on phrase using gin (to_tsvector('english', phrase))")
	if err != nil {
		log.Fatal("create index error", err)
		return nil, err
	}
	_, err = dbpool.Exec(context.Background(), "delete from phrase")
	if err != nil {
		log.Fatal("Deletion error", err)
		return nil, err
	}
	_, err = dbpool.Exec(context.Background(), "insert into phrase (phrase) values ($1)", "JIRA is bad")
	if err != nil {
		log.Fatal("Insertion error", err)
		return nil, err
	}
	_, err = dbpool.Exec(context.Background(), "insert into phrase (phrase) values ($1)", "not matching")
	if err != nil {
		log.Fatal("Insertion error", err)
		return nil, err
	}
	phrases, err := data.SearchPhrases(dbpool, "JIRA")
	if err != nil {
		log.Fatal("Query error", err)
		return nil, err
	}
	for _, phrase := range *phrases {
		log.Println("phrase:", phrase.Id, "phrase:", phrase.Phrase)
	}
	log.Println("All done!")
	return dbpool, nil
}

type Config struct {
	BasicAuthUser      string
	BasicAuthPass      string
	Port               string
	SlackToken         string
	SlackSigningSecret string
	DatabaseUrl        string
}

func setupWebServer(config Config, dbpool *pgxpool.Pool, slackClient *slack.Client) *echo.Echo {
	app := echo.New()
	app.Renderer = view.EchoTemplate
	app.Use(middleware.Logger())
	app.Pre(middleware.AddTrailingSlashWithConfig(middleware.TrailingSlashConfig{RedirectCode: http.StatusTemporaryRedirect}))
	phraseHandler := handler.PhraseHandler{Pool: dbpool}
	slackHandler := handler.SlackHandler{Pool: dbpool, Client: slackClient, SigningSecret: config.SlackSigningSecret}
	phraseGroup := app.Group("/phrase")
	phraseGroup.Use(middleware.BasicAuth(func(username, password string, _ echo.Context) (bool, error) {
		if subtle.ConstantTimeCompare([]byte(username), []byte(config.BasicAuthUser)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(config.BasicAuthPass)) == 1 {
			return true, nil
		}
		return false, nil
	}))
	phraseGroup.GET("/", phraseHandler.HandlePhraseShow)
	phraseGroup.POST("/", phraseHandler.HandleCreatePhrase)
	phraseGroup.POST("/:id/delete/", phraseHandler.HandleDeletePhrase)
	app.POST("/insult/", slackHandler.HandleInsultJira)
	return app
}

func main() {
	log.Println("Starting phrase bot")
	godotenv.Load()

	config := Config{
		BasicAuthUser:      os.Getenv("BASIC_AUTH_USER"),
		BasicAuthPass:      os.Getenv("BASIC_AUTH_PASS"),
		Port:               os.Getenv("PORT"),
		SlackToken:         os.Getenv("SLACK_TOKEN"),
		SlackSigningSecret: os.Getenv("SLACK_SIGNING_SECRET"),
		DatabaseUrl:        os.Getenv("DATABASE_URL"),
	}

	slackClient := slack.New(config.SlackToken, slack.OptionDebug(true))

	dbpool, err := setupDB(config)
	if err != nil {
		log.Fatal("Unable to connect with database", err)
		return
	}

	app := setupWebServer(config, dbpool, slackClient)
	app.Logger.Fatal(app.Start(":" + config.Port))
}
