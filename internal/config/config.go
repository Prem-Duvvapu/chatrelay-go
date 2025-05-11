package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SlackBotToken string
	SlackAppToken string
	BackendURL    string
	OTelEndpoint  string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found. Using system environment variables.")
	}

	return &Config{
		SlackBotToken: os.Getenv("SLACK_BOT_TOKEN"),
		SlackAppToken: os.Getenv("SLACK_APP_TOKEN"),
		BackendURL:    os.Getenv("BACKEND_URL"),
		OTelEndpoint:  os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}
}
