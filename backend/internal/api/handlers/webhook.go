package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/DamianAcri/DevAssistant/internal/db"
	dbsqlc "github.com/DamianAcri/DevAssistant/internal/db/sqlc"
)

// GitHubWebhookHandler handles GitHub webhook requests and persists
// relevant PR metadata into PostgreSQL.
type GitHubWebhookHandler struct {
	store *db.Store
}

// NewGitHubWebhookHandler creates a webhook handler with DB access.
func NewGitHubWebhookHandler(store *db.Store) http.HandlerFunc {
	h := &GitHubWebhookHandler{
		store: store,
	}

	return h.Handle
}

// PullRequestEvent represents only the pull_request fields we need.
type PullRequestEvent struct {
	Action string `json:"action"`
	Number int32  `json:"number"`

	PullRequest struct {
		Title   string `json:"title"`
		Body    string `json:"body"`
		URL     string `json:"html_url"`
		DiffURL string `json:"diff_url"`
	} `json:"pull_request"`

	Repository struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// WorkflowRunEvent represents only the workflow_run fields we need.
type WorkflowRunEvent struct {
	Action string `json:"action"`

	WorkflowRun struct {
		ID         int64  `json:"id"`
		Name       string `json:"name"`
		Conclusion string `json:"conclusion"`
		HeadSHA    string `json:"head_sha"`
	} `json:"workflow_run"`

	Repository struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// Handle processes validated GitHub webhook requests.
func (h *GitHubWebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("X-GitHub-Event")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	switch eventType {
	case "pull_request":
		h.handlePullRequestEvent(w, body)

	case "workflow_run":
		h.handleWorkflowRunEvent(w, body)

	default:
		// Unknown events should be ignored silently with 200 OK.
		log.Printf("ignoring unsupported event type=%s", eventType)
		w.WriteHeader(http.StatusOK)
	}
}

// handlePullRequestEvent parses the PR payload and creates a pending PR analysis row.
func (h *GitHubWebhookHandler) handlePullRequestEvent(w http.ResponseWriter, body []byte) {
	var event PullRequestEvent

	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid pull_request payload", http.StatusBadRequest)
		return
	}

	// Only process the actions required by the spec.
	if event.Action != "opened" && event.Action != "synchronize" {
		log.Printf("ignoring pull_request action=%s", event.Action)
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx := context.Background()

	// Try to load the repository first.
	repo, err := h.store.Queries.GetRepositoryByGitHubID(ctx, event.Repository.ID)
	if err != nil {
		// If it does not exist yet, create it.
		if errors.Is(err, pgx.ErrNoRows) {
			repo, err = h.store.Queries.CreateRepository(ctx, dbsqlc.CreateRepositoryParams{
				GithubID: event.Repository.ID,
				FullName: event.Repository.FullName,
			})
			if err != nil {
				http.Error(w, "failed to create repository", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "failed to load repository", http.StatusInternalServerError)
			return
		}
	}

	// Insert the pending PR analysis row.
	analysis, err := h.store.Queries.CreatePendingPRAnalysis(ctx, dbsqlc.CreatePendingPRAnalysisParams{
		RepoID:   repo.ID,
		PrNumber: event.Number,
		PrTitle:  event.PullRequest.Title,
		PrUrl:    event.PullRequest.URL,
		DiffUrl:  event.PullRequest.DiffURL,
	})
	if err != nil {
		http.Error(w, "failed to create pending PR analysis", http.StatusInternalServerError)
		return
	}

	log.Printf(
		"saved pending PR analysis: analysis_id=%d action=%s pr_number=%d title=%q repo=%s",
		analysis.ID,
		event.Action,
		event.Number,
		event.PullRequest.Title,
		event.Repository.FullName,
	)

	w.WriteHeader(http.StatusOK)
}

// handleWorkflowRunEvent parses workflow_run payloads and logs relevant CI failures.
// We are not persisting ci_analyses yet in this step.
func (h *GitHubWebhookHandler) handleWorkflowRunEvent(w http.ResponseWriter, body []byte) {
	var event WorkflowRunEvent

	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid workflow_run payload", http.StatusBadRequest)
		return
	}

	if event.Action != "completed" {
		log.Printf("ignoring workflow_run action=%s", event.Action)
		w.WriteHeader(http.StatusOK)
		return
	}

	if event.WorkflowRun.Conclusion != "failure" {
		log.Printf(
			"ignoring workflow_run conclusion=%s repo=%s workflow=%q",
			event.WorkflowRun.Conclusion,
			event.Repository.FullName,
			event.WorkflowRun.Name,
		)
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf(
		"received workflow_run event: action=%s name=%s conclusion=%s repo=%s",
		event.Action,
		event.WorkflowRun.Name,
		event.WorkflowRun.Conclusion,
		event.Repository.FullName,
	)

	w.WriteHeader(http.StatusOK)
}