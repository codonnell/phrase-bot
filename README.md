# Phrase Bot

Small website and slack bot that stores phrases and posts them to slack

## Development

Create a new slack bot with the ability to write chats and put its token in the `SLACK_TOKEN` environment variable and signing secret in the `SLACK_SIGNING_SECRET` environment variable. Then... `make run` and off you go!

For hot reloading, first you need to install [air](https://github.com/cosmtrek/air):

```
go install github.com/cosmtrek/air@latest
```

Then to compile with hot reloading, `make watch`.
