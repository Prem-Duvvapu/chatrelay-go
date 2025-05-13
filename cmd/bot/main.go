package main

import (
    "context"
    "fmt"
    "log"
    "os"

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
    port := os.Getenv("CHATRELAY_PORT")
    if port == "" {
        port = "8081"
    }

    shutdown := telemetry.InitTracer("chatrelay-bot")
    defer func() {
        if err := shutdown(context.Background()); err != nil {
            log.Fatalf("Failed to shutdown tracer: %v", err)
        }
    }()

    fmt.Printf("ChatRelay bot is starting on port %s...\n", port)
    ctx := context.Background()
    slack.StartSlackListener(ctx)
}
