-- name: CreatePendingPRAnalysis :one
INSERT INTO pr_analyses (
    repo_id,
    pr_number,
    pr_title,
    pr_url,
    diff_url,
    status
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    'pending'
)
RETURNING *;