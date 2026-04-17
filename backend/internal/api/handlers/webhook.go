package handlers

import (
	"io"
	"log"
	"net/http"
)

func GitHubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-GitHub-Event") //like pull_request, workflow_run...
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Received GitHub event: %s\nPayload: %s\n", event, string(body))
	w.WriteHeader(http.StatusOK)
}