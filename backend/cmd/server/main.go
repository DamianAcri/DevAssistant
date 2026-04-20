package main

import (
	"context"
	"log"
	"net/http"

	"github.com/DamianAcri/DevAssistant/internal/api"
	"github.com/DamianAcri/DevAssistant/internal/config"
	"github.com/DamianAcri/DevAssistant/internal/db"
)

func main() {
	cfg := config.LoadConfig()

	// Create a context used for startup operations like DB connection
	ctx := context.Background()

	// Initialize the PostgreSQL store
	store, err := db.NewStore(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer store.Close()

	router := api.NewRouter(cfg, store)

	log.Println("Server running on port:", cfg.Port)

	err = http.ListenAndServe(":"+cfg.Port, router)
	if err != nil {
		log.Fatal(err)
	}
}