package config

import (
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env string
	GitHubWebhookSecret string
}

func LoadConfig() Config {
	godotenv.Load("../../.env") // Obtain the env file info 

	port := os.Getenv("PORT") // whatever we have in the .env file
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")

	return Config {
		Port: port,
		Env: env,
		GitHubWebhookSecret: webhookSecret,
	}

}