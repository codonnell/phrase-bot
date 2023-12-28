package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"phrase_bot/handler"
	"phrase_bot/view"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/slack-go/slack"
)

func main() {
	log.Println("Starting phrase bot")
	godotenv.Load()
	port := os.Getenv("PORT")
	slackToken := os.Getenv("SLACK_TOKEN")
	slackSigningSecret := os.Getenv("SLACK_SIGNING_SECRET")
	client := slack.New(slackToken, slack.OptionDebug(true))
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create dbpoolection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	createTable := `create table if not exists phrase (id serial primary key, phrase text not null, inserted_at timestamptz not null default now())`
	_, err = dbpool.Exec(context.Background(), createTable)
	if err != nil {
		log.Printf("%q: %s\n", err, createTable)
		return
	}
	_, err = dbpool.Exec(context.Background(), "delete from phrase")
	if err != nil {
		log.Fatal("Deletion error", err)
	}
	_, err = dbpool.Exec(context.Background(), "insert into phrase (phrase) values ($1)", "JIRA is bad")
	if err != nil {
		log.Fatal("Insertion error", err)
	}
	row := dbpool.QueryRow(context.Background(), "select id, phrase from phrase")
	var id int
	var phrase string
	err = row.Scan(&id, &phrase)
	if err != nil {
		log.Fatal("Query error", err)
	}
	log.Println("id:", id, "phrase:", phrase)
	log.Println("All done!")

	app := echo.New()
	app.Renderer = view.EchoTemplate
	app.Use(middleware.Logger())
	app.Pre(middleware.AddTrailingSlashWithConfig(middleware.TrailingSlashConfig{RedirectCode: http.StatusTemporaryRedirect}))
	phraseHandler := handler.PhraseHandler{Pool: dbpool}
	slackHandler := handler.SlackHandler{Pool: dbpool, Client: client, SigningSecret: slackSigningSecret}
	phraseGroup := app.Group("/phrase")
	phraseGroup.Use(middleware.BasicAuth(func(username, password string, _ echo.Context) (bool, error) {
		if subtle.ConstantTimeCompare([]byte(username), []byte("pf2")) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte("rulez")) == 1 {
			return true, nil
		}
		return false, nil
	}))
	phraseGroup.GET("/", phraseHandler.HandlePhraseShow)
	phraseGroup.POST("/", phraseHandler.HandleCreatePhrase)
	phraseGroup.POST("/:id/delete/", phraseHandler.HandleDeletePhrase)
	app.POST("/insult/", slackHandler.HandleInsultJira)
	app.Logger.Fatal(app.Start(":" + port))
}
