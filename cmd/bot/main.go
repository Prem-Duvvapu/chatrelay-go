package main

import (
    "context"
    "fmt"
    "log"

    "github.com/joho/godotenv"
    "chatrelay-go/internal/slack"
    "chatrelay-go/internal/telemetry"
)

func init() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found. Using system environment variables.")
    }
}

func main() {
    shutdown := telemetry.InitTracer("chatrelay-bot")
    defer func() {
        if err := shutdown(context.Background()); err != nil {
            log.Fatalf("Failed to shutdown tracer: %v", err)
        }
    }()

    fmt.Printf("ChatRelay bot is starting...");
    ctx := context.Background()
    slack.StartSlackListener(ctx)
}
