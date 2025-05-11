package main

import (
    "context"
    "fmt"
    "log"

    "github.com/joho/godotenv"
    "chatrelay-go/internal/slack"
)

func init() {
    if err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found. Using system environment variables.")
    }
}

func main() {
    fmt.Println("ChatRelay bot is starting...")
    ctx := context.Background()
    slack.StartSlackListener(ctx)
}
