package api

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/DamianAcri/DevAssistant/internal/api/handlers"
	"github.com/DamianAcri/DevAssistant/internal/api/middleware"
	"github.com/DamianAcri/DevAssistant/internal/config"
	"github.com/DamianAcri/DevAssistant/internal/db"

)

func NewRouter(cfg config.Config, store *db.Store) http.Handler { //creates and returns the router
	r := chi.NewRouter() 

	r.Get("/health", handlers.HealthHandler) // GET /health and we run health handler when someone access it
	r.With(
		middleware.VerifyGitHubWebhookSignature(cfg.GitHubWebhookSecret),
	).Post("/webhook/github", handlers.NewGitHubWebhookHandler(store)) // POST /webhook and we run GitHubWebhookHandler when someone access it, but before that we run the middleware to verify the signature of the webhook request

	return r //return the router
}

