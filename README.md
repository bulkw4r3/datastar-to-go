# Number Storage

A single-file Go web application that stores and displays pairs of numbers: a 7-digit number and a 10–20 digit number. Built with **Datastar** for reactive UI updates via Server-Sent Events (SSE) and **SQLite** for persistence — no JavaScript required.

## Features

- **Reactive UI**: Datastar handles form submissions and DOM updates via SSE.
- **Validation**: 7-digit and 10–20 digit number inputs are validated server-side.
- **Persistence**: Data is stored in an embedded SQLite database (`numbers.db`).
- **Zero CGO**: Uses `modernc.org/sqlite`, a pure-Go SQLite driver.

## Tech Stack

- **Go** — backend & templating
- **Datastar** — SSE-driven hypermedia (browser-side)
- **SQLite** — embedded database

## Getting Started

### Prerequisites

- [Go](https://go.dev/) 1.26+
- (Optional) [Docker](https://docs.docker.com/get-docker/)

### Run Locally

```bash
# Run the app (auto-creates numbers.db if missing)
go run main.go

# Or build a binary
go build -o numberstore main.go
./numberstore
```

The app will be available at `http://localhost:8080`.

To run on a custom port:

```bash
PORT=3000 go run main.go
```

### Run with Docker

```bash
# Build the image
docker build -t numberstore .

# Run the container (mount local database file for persistence)
docker run -p 8080:8080 -v $(pwd)/numbers.db:/numbers.db numberstore
```

Or use Docker Compose:

```bash
docker compose up -d --build
```

## Project Structure

```
├── main.go              # Single-file Go web app
├── go.mod               # Go module definition
├── Dockerfile           # Multi-stage Docker build (scratch runtime)
├── docker-compose.yml   # Docker Compose service definition
├── .dockerignore        # Docker ignore rules
└── numbers.db           # SQLite database (created on first run)
```

## How It Works

1. **Page Load**: `GET /` returns the full HTML page with the form and stored numbers.
2. **Form Submit**: Datastar intercepts the form submission (`@post('/api/numbers')`) and sends the signals as a JSON POST request.
3. **Server Response**: The server validates the input, inserts into SQLite, and returns an HTML fragment with `datastar-selector` and `datastar-mode` headers.
4. **DOM Update**: Datastar patches the `#app` element with the returned fragment — no full page reload.

## Data Binding

Datastar signals use hyphenated attributes (e.g., `data-bind:seven-digit`) which map to Go struct fields with camelCase JSON tags (`SevenDigit` via `json:"sevenDigit"`). The binding is automatic via `datastar.ReadSignals`.

## License

MIT
