package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    "chatrelay-go/internal/slack"
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

    fmt.Printf("ChatRelay bot is starting on port %s...\n", port)
    ctx := context.Background()
    slack.StartSlackListener(ctx)
    // If you later add an HTTP server, bind it to ":" + port
}
