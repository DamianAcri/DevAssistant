package handlers

import (
	"log"
	"net/http"
)

func GitHubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-GitHub-Event") //like pull_request, workflow_run...
	log.Println("Event recieved: ", event)

	w.WriteHeader(http.StatusOK)
}