package config

import (
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env string
}

func LoadConfig() Config {
	godotenv.Load() // Obtain the env file info 

	port := os.Getenv("PORT") // whatever we have in the .env file
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	return Config {
		Port: port,
		Env: env,
	}

}