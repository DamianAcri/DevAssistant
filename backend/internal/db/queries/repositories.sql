-- name: CreateRepository :one
INSERT INTO repositories (
    github_id,
    full_name
) VALUES (
    $1,
    $2
)
RETURNING *;

-- name: GetRepositoryByGitHubID :one
SELECT *
FROM repositories
WHERE github_id = $1;

-- name: UpdateRepositoryFullName :one
UPDATE repositories
SET full_name = $2
WHERE github_id = $1
RETURNING *;