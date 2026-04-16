# AI GitHub Dev Assistant — Especificación Técnica

> Backend Go + Frontend TypeScript | Portfolio project para Dev Productivity roles

---

## 1. Qué es esto

Una plataforma que escucha eventos de GitHub (PRs, CI failures), los analiza con un LLM, y devuelve resultados estructurados: comentarios automáticos en el PR y un dashboard de observabilidad.

**No es un chatbot.** Es un sistema de automatización orientado a eventos, con LLM como capa de análisis.

---

## 2. Stack

| Capa | Tecnología | Por qué |
|---|---|---|
| Backend | Go 1.22+ | Stack del equipo, concurrencia, rendimiento |
| Frontend | TypeScript + React 18 + Vite | Stack del equipo |
| Base de datos | PostgreSQL 16 | Persistencia de runs y resultados |
| ORM/queries | `sqlc` + `pgx` | Queries tipadas, no magia |
| HTTP router | `chi` | Minimalista, idiomático en Go |
| LLM | Anthropic API (claude-sonnet-4-20250514) | Structured outputs fiables |
| Auth | GitHub OAuth 2.0 | Contextual al producto |
| Infra | Docker Compose | Local reproducible |
| Testing backend | `testing` stdlib + `testify` | |
| Testing frontend | Vitest + React Testing Library | |

---

## 3. Estructura de carpetas

```
ai-github-assistant/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # entrypoint
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handlers/
│   │   │   │   ├── webhook.go       # POST /webhook/github
│   │   │   │   ├── analyze.go       # POST /analyze/pr, POST /analyze/ci
│   │   │   │   └── results.go       # GET /prs, GET /results/:pr_id
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go
│   │   │   │   └── webhook_verify.go
│   │   │   └── router.go
│   │   ├── github/
│   │   │   ├── client.go            # wrapper GitHub API
│   │   │   ├── diff.go              # fetch PR diff
│   │   │   ├── comment.go           # post comment en PR
│   │   │   └── webhook.go           # parsing de eventos
│   │   ├── llm/
│   │   │   ├── client.go            # wrapper Anthropic API
│   │   │   ├── prompts.go           # prompt templates
│   │   │   └── types.go             # structs de input/output
│   │   ├── db/
│   │   │   ├── migrations/
│   │   │   │   ├── 001_init.sql
│   │   │   │   └── 002_ci_failures.sql
│   │   │   ├── queries/             # .sql para sqlc
│   │   │   └── store.go             # interface de DB
│   │   └── config/
│   │       └── config.go            # env vars
│   ├── Dockerfile
│   └── go.mod
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── PRList.tsx
│   │   │   ├── PRDetail.tsx
│   │   │   ├── CIFailureList.tsx
│   │   │   └── AnalysisCard.tsx
│   │   ├── api/
│   │   │   └── client.ts            # fetch wrapper tipado
│   │   ├── types/
│   │   │   └── index.ts             # tipos compartidos
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   └── vite.config.ts
│
├── docker-compose.yml
└── .env.example
```

---

## 4. Modelo de datos

```sql
-- Repositorios conectados
CREATE TABLE repositories (
    id          BIGSERIAL PRIMARY KEY,
    github_id   BIGINT UNIQUE NOT NULL,
    full_name   TEXT NOT NULL,          -- "owner/repo"
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Análisis de PRs
CREATE TABLE pr_analyses (
    id              BIGSERIAL PRIMARY KEY,
    repo_id         BIGINT REFERENCES repositories(id),
    pr_number       INT NOT NULL,
    pr_title        TEXT NOT NULL,
    pr_url          TEXT NOT NULL,
    diff_url        TEXT NOT NULL,
    summary         TEXT,
    risk_level      TEXT CHECK (risk_level IN ('low', 'medium', 'high')),
    possible_bugs   JSONB,             -- array de strings
    missing_tests   JSONB,             -- array de strings
    improvements    JSONB,             -- array de strings
    raw_response    JSONB,             -- respuesta LLM completa
    status          TEXT DEFAULT 'pending',  -- pending | completed | failed
    error           TEXT,
    github_comment_id BIGINT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Análisis de CI failures
CREATE TABLE ci_analyses (
    id              BIGSERIAL PRIMARY KEY,
    repo_id         BIGINT REFERENCES repositories(id),
    run_id          BIGINT NOT NULL,
    workflow_name   TEXT NOT NULL,
    pr_number       INT,               -- nullable, puede ser push directo
    root_cause      TEXT,
    fix_suggestion  TEXT,
    confidence      FLOAT,
    raw_response    JSONB,
    status          TEXT DEFAULT 'pending',
    error           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
```

---

## 5. API endpoints

### Backend (Go, puerto 8080)

```
POST   /webhook/github              Recibe eventos de GitHub
POST   /analyze/pr                  Analiza un PR manualmente (para testing)
POST   /analyze/ci                  Analiza un CI failure manualmente
GET    /api/prs                     Lista PR analyses
GET    /api/prs/:id                 Detalle de un PR analysis
GET    /api/ci                      Lista CI analyses
GET    /api/ci/:id                  Detalle de un CI analysis
GET    /health                      Healthcheck
```

### Webhook event routing

```
pull_request (opened, synchronize) → handler PR
workflow_run (completed, failure)   → handler CI
```

---

## 6. Interfaces Go clave

### NodeExecutor equivalente — Analyzer interface

```go
// internal/llm/types.go

type PRAnalysisInput struct {
    PRNumber  int
    PRTitle   string
    PRBody    string
    Diff      string
    RepoName  string
}

type PRAnalysisOutput struct {
    Summary      string   `json:"summary"`
    RiskLevel    string   `json:"risk_level"`       // "low" | "medium" | "high"
    PossibleBugs []string `json:"possible_bugs"`
    MissingTests []string `json:"missing_tests"`
    Improvements []string `json:"suggested_improvements"`
}

type CIAnalysisInput struct {
    WorkflowName string
    Logs         string
    CommitSHA    string
    RepoName     string
}

type CIAnalysisOutput struct {
    RootCause     string  `json:"root_cause"`
    FixSuggestion string  `json:"fix_suggestion"`
    Confidence    float64 `json:"confidence"`
}
```

### Analyzer interface

```go
// internal/llm/client.go

type Analyzer interface {
    AnalyzePR(ctx context.Context, input PRAnalysisInput) (*PRAnalysisOutput, error)
    AnalyzeCI(ctx context.Context, input CIAnalysisInput) (*CIAnalysisOutput, error)
}
```

---

## 7. Flujo completo — PR abierto

```
1. Developer abre PR en GitHub
2. GitHub envía POST /webhook/github con event: pull_request, action: opened
3. middleware/webhook_verify.go valida HMAC-SHA256 signature (X-Hub-Signature-256)
4. handlers/webhook.go parsea el evento y despacha goroutine
5. github/diff.go llama GET /repos/{owner}/{repo}/pulls/{pr}/files → raw diff
6. llm/client.go envía diff + prompt a Anthropic API → JSON estructurado
7. db/store.go guarda PRAnalysis en PostgreSQL
8. github/comment.go llama POST /repos/{owner}/{repo}/issues/{pr}/comments
9. GET /api/prs devuelve el resultado al frontend
```

---

## 8. Prompt de PR (base)

```
You are a code review assistant. Analyze the following pull request diff and return ONLY a JSON object, no markdown, no explanation.

PR Title: {{title}}
Repository: {{repo}}

Diff:
{{diff}}

Return this exact JSON structure:
{
  "summary": "2-3 sentence description of what this PR does",
  "risk_level": "low|medium|high",
  "possible_bugs": ["..."],
  "missing_tests": ["..."],
  "suggested_improvements": ["..."]
}

Rules:
- risk_level low = style/docs changes
- risk_level medium = logic changes with existing tests
- risk_level high = untested logic, security-sensitive code, DB schema changes
- Each array max 3 items
- Be specific, not generic
```

---

## 9. Variables de entorno

```bash
# .env.example

# GitHub
GITHUB_APP_ID=
GITHUB_PRIVATE_KEY_PATH=./github-app.pem
GITHUB_WEBHOOK_SECRET=
GITHUB_TOKEN=                        # Personal Access Token para dev

# Anthropic
ANTHROPIC_API_KEY=

# DB
DATABASE_URL=postgres://user:pass@localhost:5432/ghassistant?sslmode=disable

# Server
PORT=8080
ENV=development                      # development | production
```

---

## 10. Fases y criterios de "va bien"

### Fase 1 — Webhook receiver (3-4 días)

**Qué construyes:**
- Go server con `chi`
- `POST /webhook/github` que parsea eventos `pull_request` y `workflow_run`
- Valida HMAC signature
- Loguea el evento en stdout
- Docker Compose con PostgreSQL

**Sabes que va bien cuando:**
- `curl -X POST localhost:8080/webhook/github -H "X-GitHub-Event: pull_request" -d @test_payload.json` → 200 OK y log del evento
- Payload con signature incorrecta → 401
- Payload con event desconocido → 200 OK silencioso (no crashea)
- Tests: `TestWebhookSignatureValid`, `TestWebhookSignatureInvalid`, `TestWebhookUnknownEvent`

---

### Fase 2 — GitHub API integration (2-3 días)

**Qué construyes:**
- `github/diff.go`: dado repo + PR number, devuelve el diff como string
- `github/comment.go`: dado repo + PR number + body, publica comentario
- Guarda `repositories` y `pr_analyses` (status: pending) en DB al recibir webhook

**Sabes que va bien cuando:**
- Script de test local `go run cmd/test_diff/main.go owner/repo 42` imprime el diff
- Comentario de prueba aparece en un PR real de un repo tuyo
- `SELECT * FROM pr_analyses;` muestra una fila con status pending

---

### Fase 3 — LLM integration (2-3 días)

**Qué construyes:**
- `llm/client.go` que llama Anthropic API
- Parsea la respuesta JSON forzada
- Actualiza `pr_analyses` con resultado y status: completed

**Sabes que va bien cuando:**
- `curl -X POST localhost:8080/analyze/pr -d '{"pr_number":42,"repo":"owner/repo"}'` devuelve el JSON de análisis
- El JSON tiene los 5 campos: summary, risk_level, possible_bugs, missing_tests, improvements
- Si el LLM no devuelve JSON válido → status: failed + error guardado en DB, no crash
- Tests: `TestParsePRAnalysisOutput` con fixtures de respuestas LLM válidas e inválidas

---

### Fase 4 — End-to-end (1-2 días)

**Qué construyes:**
- Pipeline completo: webhook → diff → LLM → comentario en GitHub
- Usa `ngrok` para exponer localhost a GitHub webhooks reales

**Sabes que va bien cuando:**
- Abres un PR real en un repo tuyo de prueba
- En menos de 30 segundos aparece un comentario automático del bot
- El comentario tiene el summary, risk_level y bullets de mejoras
- `SELECT * FROM pr_analyses WHERE status = 'completed';` muestra la fila

---

### Fase 5 — Frontend TypeScript (3-4 días)

**Qué construyes:**
- React app con Vite + TypeScript
- `GET /api/prs` → tabla de PRs analizados con risk_level coloreado
- Click en PR → detalle con todos los campos del análisis
- `GET /api/ci` → tabla de CI failures

**Sabes que va bien cuando:**
- `npm run build` sin errores TypeScript
- La tabla carga datos reales de la DB
- Risk level high → badge rojo, medium → amarillo, low → verde
- Si la API está caída → mensaje de error, no pantalla en blanco
- Tests: `PRList renders rows from API response`, `PRDetail shows all analysis fields`

---

## 11. Testing strategy

### Backend

```bash
# Unit tests (sin DB, sin network)
go test ./internal/llm/...
go test ./internal/github/...

# Integration tests (requieren DB real o testcontainers)
go test ./internal/db/...

# E2E manual
go run cmd/server/main.go
# + ngrok + PR real en GitHub
```

Tests mínimos obligatorios:
- `TestWebhookSignatureVerification`
- `TestParsePRAnalysisOutput_ValidJSON`
- `TestParsePRAnalysisOutput_InvalidJSON`
- `TestAnalyzePR_Integration` (con mock del LLM client)
- `TestGitHubDiffFetch_Integration` (con mock del HTTP client)

### Frontend

```bash
npm run test          # Vitest
npm run type-check    # tsc --noEmit
npm run lint          # eslint
```

Tests mínimos:
- `PRList` renderiza filas correctamente
- `PRDetail` muestra todos los campos
- Manejo de estado de carga y error

---

## 12. Comandos de setup inicial

```bash
# Clonar y arrancar
git init ai-github-assistant
cd ai-github-assistant

# Backend
mkdir -p backend/cmd/server
cd backend
go mod init github.com/tuuser/ai-github-assistant
go get github.com/go-chi/chi/v5
go get github.com/jackc/pgx/v5
go get github.com/joho/godotenv

# Frontend
cd ../
npm create vite@latest frontend -- --template react-ts
cd frontend && npm install

# Infra
# docker-compose.yml en la raíz (ver abajo)
docker compose up -d postgres
```

### docker-compose.yml mínimo

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ghassistant
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

---

## 13. Lo que NO construyes en el MVP

- GitHub App (usas Personal Access Token en dev, suficiente para portfolio)
- Auth de usuarios (una sola instancia, un solo token)
- Rate limiting
- Queue system (las goroutines directas son suficientes para el MVP)
- CI analysis automático (el PR analysis ya es suficientemente fuerte; CI es bonus)

---

## 14. Qué demuestra este proyecto

| Habilidad | Evidencia concreta |
|---|---|
| Backend Go | HTTP server, webhooks, concurrencia, interfaces, error handling |
| System design | Event-driven, async processing, structured LLM outputs |
| GitHub API | Webhooks, REST API, auth |
| AI engineering | Prompt design, JSON forcing, manejo de outputs no deterministas |
| TypeScript | Frontend tipado, async data fetching, estado de UI |
| PostgreSQL | Schema design, queries, migraciones |
| DevOps básico | Docker Compose, env management |

---

## 15. Orden de commits recomendado

```
feat: initial Go server with health endpoint
feat: webhook receiver with HMAC verification
feat: GitHub diff fetching
feat: database schema and migrations
feat: LLM integration with structured output
feat: GitHub comment posting
feat: end-to-end PR analysis pipeline
feat: REST API endpoints for frontend
feat: React dashboard with PR list
feat: PR detail view
feat: CI failure analysis (bonus)
```

Cada commit debe dejar el proyecto en estado funcionando. No commitees broken states.