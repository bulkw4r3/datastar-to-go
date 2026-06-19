# AGENTS.md

## Project overview

Single-file Go web app (`main.go`) using Datastar (SSE-driven hypermedia) + SQLite.
Module name: `numberstore`.

## Commands

```bash
# Run (auto-creates numbers.db if missing)
go run main.go

# Build binary
go build -o numberstore main.go

# Run on custom port
PORT=3000 go run main.go

# Docker build & run
docker build -t numberstore .
docker run -p 8080:8080 -v $(pwd)/numbers.db:/numbers.db numberstore

# Docker Compose
docker compose up -d --build
```

## Architecture

- **Datastar**: Creates reactive frontend via SSE without writing JS. The HTML response
  sets `datastar-selector` and `datastar-mode` headers to patch the DOM.
- **SQLite**: `numbers.db` is created on startup (`CREATE TABLE IF NOT EXISTS`).
  Pure-Go driver (`modernc.org/sqlite`) — no CGO required.
- **Routes**: `GET /` serves the page, `POST /api/numbers` handles form submissions.
- **Datastar CDN script**: loaded from jsDelivr (v1.0.2), distinct from the Go library version.

## Gotchas

- The `numberstore` binary at repo root is a pre-built executable — ignore it; work with `main.go`.
- Datastar signals use `data-bind:seven-digit` (hyphenated) which maps to Go struct field
  `SevenDigit` (camelCase JSON tag). The binding is automatic via `datastar.ReadSignals`.
- Form submissions use `@post('/api/numbers')` — Datastar intercepts the form submit,
  serializes signals, and sends them as a POST with JSON body. The browser never does a
  traditional form POST.
- No tests exist. The project has no Makefile, CI, or lint config.
- Docker files present: `Dockerfile`, `.dockerignore`, `docker-compose.yml`.
