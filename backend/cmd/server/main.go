package main

import (
	"log"
	"net/http"
	"github.com/DamianAcri/DevAssistant/internal/api"
	"github.com/DamianAcri/DevAssistant/internal/config"
)

func main() {
	cfg := config.LoadConfig()
	router := api.NewRouter()

	log.Println("Server running on port: ", cfg.Port)

	//and we run the server
	http.ListenAndServe(":"+cfg.Port, router)
}