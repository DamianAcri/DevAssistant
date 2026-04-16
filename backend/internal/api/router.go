package api

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/DamianAcri/DevAssistant/internal/api/handlers"
)

func NewRouter() http.Handler { //creates and returns the router
	r := chi.NewRouter() 

	r.Get("/health", handlers.HealthHandler) // GET /health and we run health handler when someone access it
	r.Post("/webhook/github", handlers.GitHubWebhookHandler) // POST /webhook/github, so that we can recieve GH events

	return r //return the router
}

