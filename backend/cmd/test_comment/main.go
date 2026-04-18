package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/DamianAcri/DevAssistant/internal/config"
	gh "github.com/DamianAcri/DevAssistant/internal/github"
)

func main() {
	// Expected usage:
	// go run ./cmd/test_comment <owner> <repo> <pr_number> <comment_body>
	if len(os.Args) < 5 {
		log.Fatalf("usage: go run ./cmd/test_comment <owner> <repo> <pr_number> <comment_body>")
	}

	owner := os.Args[1]
	repo := os.Args[2]

	prNumber, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("invalid pr_number: %v", err)
	}

	commentBody := os.Args[4]

	cfg := config.LoadConfig()

	if cfg.GitHubToken == "" {
		log.Fatal("GITHUB_TOKEN is not configured")
	}

	client := gh.NewClient(cfg.GitHubToken)

	comment, err := client.PostPullRequestComment(owner, repo, prNumber, commentBody)
	if err != nil {
		log.Fatalf("failed to post PR comment: %v", err)
	}

	fmt.Printf("Comment created successfully!\n")
	fmt.Printf("Comment ID: %d\n", comment.ID)
	fmt.Printf("Comment URL: %s\n", comment.URL)
	fmt.Printf("Comment Body: %s\n", comment.Body)
}