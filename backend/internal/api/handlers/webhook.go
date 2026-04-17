package handlers

import (
	"io"
	"log"
	"net/http"
	"encoding/json"
)

type PullRequestEvent struct {
	Action      string `json:"action"` // opened, synchronize, closed, etc.
	Number      int    `json:"number"` // PR number
	PullRequest struct {
		Title string `json:"title"` // PR title
		Body  string `json:"body"`  // PR description
		URL   string `json:"html_url"`
	} `json:"pull_request"` // we only care about the pull_request field of the event, so we define an anonymous struct for it	
	Repository struct {
		FullName string `json:"full_name"` // owner/repo
	} `json:"repository"` // we only care about the repository field of the event, so we define an anonymous struct for it
}

type WorkflowRunEvent struct {
	Action      string `json:"action"` // like completed, requested, etc.
	WorkflowRun struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		Conclusion string `json:"conclusion"` // success, failure, cancelled...
		HeadSHA    string `json:"head_sha"`
	} `json:"workflow_run"` // we only care about the workflow_run field of the event, so we define an anonymous struct for it
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"` // we only care about the repository field of the event, so we define an anonymous struct for it
}

func GitHubWebhookHandler(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-GitHub-Event") //like pull_request, workflow_run...
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	// we decide what to based on the event type
	switch event {
		case "pull_request":
			var prEvent PullRequestEvent

			//JSON to struct
			if err := json.Unmarshal(body, &prEvent); err != nil { // we convert the JSON payload to our PullRequestEvent struct
				http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
				return
			}

			// for now we only care about opened and synchronize actions, but we can add more later
			if prEvent.Action != "opened" && prEvent.Action != "synchronize" {
				log.Printf("ignoring pull_request action=%s", prEvent.Action)
				w.WriteHeader(http.StatusOK)
				return
			}
			
			// log relevant info about the PR event
			log.Printf("Received pull_request event: action=%s number=%d title=%s repo=%s\n", prEvent.Action, prEvent.Number, prEvent.PullRequest.Title, prEvent.Repository.FullName)

		case "workflow_run":
			var wrEvent WorkflowRunEvent
			
			//JSON to struct
			if err := json.Unmarshal(body, &wrEvent); err != nil { // we convert the JSON payload to our WorkflowRunEvent struct
				http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
				return
			}

			// for now we only care about completed actions, but we can add more later
			if wrEvent.Action != "completed" {
				log.Printf("ignoring workflow_run action=%s", wrEvent.Action)
				w.WriteHeader(http.StatusOK)
				return
			}
			
			// log relevant info about the workflow run event
			log.Printf("Received workflow_run event: action=%s name=%s conclusion=%s repo=%s\n", wrEvent.Action, wrEvent.WorkflowRun.Name, wrEvent.WorkflowRun.Conclusion, wrEvent.Repository.FullName)
		
		default: // if it's an event type we don't care about, we just log it and return 200 OK so that GitHub doesn't keep retrying
			log.Printf("ignoring event type=%s", event)
			w.WriteHeader(http.StatusOK)
			return
	}
	
	w.WriteHeader(http.StatusOK)
}