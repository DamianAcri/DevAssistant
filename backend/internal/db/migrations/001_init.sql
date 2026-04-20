CREATE TABLE IF NOT EXISTS repositories (
    id          BIGSERIAL PRIMARY KEY,
    github_id   BIGINT UNIQUE NOT NULL,
    full_name   TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pr_analyses (
    id                  BIGSERIAL PRIMARY KEY,
    repo_id             BIGINT NOT NULL REFERENCES repositories(id),
    pr_number           INT NOT NULL,
    pr_title            TEXT NOT NULL,
    pr_url              TEXT NOT NULL,
    diff_url            TEXT NOT NULL,
    summary             TEXT,
    risk_level          TEXT CHECK (risk_level IN ('low', 'medium', 'high')),
    possible_bugs       JSONB,
    missing_tests       JSONB,
    improvements        JSONB,
    raw_response        JSONB,
    status              TEXT DEFAULT 'pending',
    error               TEXT,
    github_comment_id   BIGINT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);