package main

import (
	"log"
	"net/http"

	"github.com/DamianAcri/DevAssistant/internal/api"
	"github.com/DamianAcri/DevAssistant/internal/config"
)

func main() {
	cfg := config.LoadConfig()
	router := api.NewRouter(cfg)

	log.Println("Server running on port:", cfg.Port)

	// Start the server
	err := http.ListenAndServe(":"+cfg.Port, router)
	if err != nil {
		log.Fatal(err)
	}
}