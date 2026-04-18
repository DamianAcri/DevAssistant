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
	// go run ./cmd/test_diff owner repo 42
	if len(os.Args) != 4 {
		log.Fatalf("usage: go run ./cmd/test_diff <owner> <repo> <pr_number>")
	}

	owner := os.Args[1]
	repo := os.Args[2]

	prNumber, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatalf("invalid pr_number: %v", err)
	}

	cfg := config.LoadConfig()

	if cfg.GitHubToken == "" {
		log.Fatal("GITHUB_TOKEN is not configured")
	}

	client := gh.NewClient(cfg.GitHubToken)

	diff, err := client.GetPullRequestDiff(owner, repo, prNumber)
	if err != nil {
		log.Fatalf("failed to fetch PR diff: %v", err)
	}

	fmt.Println(diff)
}