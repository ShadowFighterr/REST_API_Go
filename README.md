# REST API Go

Simple REST API service written in Go.

## Requirements

- Go `1.21+`
- Git
- (Optional) Docker

## Quick Start

```bash
# 1) Clone
git clone <your-repo-url>
cd REST_API_Go

# 2) Install dependencies
go mod tidy

# 3) Create env file
cp .env.example .env

# 4) Run
go run ./cmd/server
```

Server runs on:

- `http://localhost:8080` (default)

---

## `.env` File Pattern

Create a `.env` file in the project root:

```env
# App
APP_NAME=rest-api-go
APP_ENV=development
APP_PORT=8080

# HTTP
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=10s

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=rest_api_go
DB_SSLMODE=disable

# Auth
JWT_SECRET=change_me
JWT_EXPIRES_IN=24h

# Logging
LOG_LEVEL=info
```

Also add `.env.example` (without real secrets) for teammates.

---

## Project Scripts (optional)

```bash
# Run app
go run ./cmd/server

# Build binary
go build -o bin/api ./cmd/server

# Run tests
go test ./...

# Format code
go fmt ./...
```

---

## API Health Check

Example endpoint:

```http
GET /health
```

Expected response:

```json
{
    "status": "ok"
}
```

---

## Recommended `.gitignore`

```gitignore
# Binaries
bin/
dist/

# Env files
.env

# IDE / OS
.vscode/
.idea/
.DS_Store
```

---

## Notes

- Keep secrets only in `.env`, never commit them.
- Use `.env.example` as the public template.
- For production, inject environment variables from your deployment platform.
+-+-+-+-+-+