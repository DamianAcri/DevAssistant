package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env string
	GitHubWebhookSecret string
	GitHubToken string
	DatabaseURL string
}

func LoadConfig() Config {
	err := godotenv.Load(".env")
	if err != nil {
		err = godotenv.Load("../../.env")
		if err != nil {
			log.Println("warning: .env file not found")
		}
	} // Obtain the env file info 

	port := os.Getenv("PORT") // whatever we have in the .env file
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	githubToken := os.Getenv("GITHUB_TOKEN")
	databaseURL := os.Getenv("DATABASE_URL")

	return Config {
		Port: port,
		Env: env,
		GitHubWebhookSecret: webhookSecret,
		GitHubToken: githubToken,
		DatabaseURL: databaseURL,
	}

}