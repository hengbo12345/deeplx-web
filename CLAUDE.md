# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DeepLX Web is a translation service with a Go backend (Go 1.22) and React frontend (React 18 + TypeScript + Tailwind CSS + Vite). It provides text translation and async document translation (.docx, .txt) using a worker pool architecture. Documents are processed asynchronously with task-based progress tracking.

## Development Commands

### Quick Start (Development)
```bash
./start.sh                     # Builds frontend, copies to nginx, starts backend on :8449 with debug logs
```

### Backend
```bash
cd backend
go run cmd/server/main.go      # Run dev server (binds 127.0.0.1:9448)
go build -o deeplx-web ./cmd/server  # Build binary
go mod tidy                    # Update dependencies
```

### Frontend
```bash
cd frontend
npm run dev                    # Vite dev server on :3000, proxies /api and /health to localhost:9448
npm run build                  # TypeScript check + Vite production build to dist/
npm run lint                   # ESLint
```

### Docker
```bash
docker-compose up -d           # Backend on :8449, Frontend on :9449
docker-compose logs -f
docker-compose down
```

## Architecture

### Backend Structure
```
backend/
├── cmd/server/main.go          # Entry point, wires all dependencies via chi router
├── internal/
│   ├── config/config.go        # All config from env vars with defaults
│   ├── handler/                # HTTP handlers: translate, document, task, health
│   ├── middleware/             # CORS, logging middleware
│   ├── models/                 # Data models (Task with Snowflake ID, request/response types)
│   ├── service/
│   │   ├── deeplx.go           # DeepLX API client with retry
│   │   ├── docx.go             # DOCX/TXT processing via python-docx subprocess
│   │   └── task_manager.go     # In-memory task tracking
│   └── worker/
│       └── document_worker.go  # Worker pool for async document translation
└── pkg/utils/                  # Zap logger with Lumberjack rotation, cleanup service
```

### Key Architecture Patterns

**Dependency wiring** (in `main.go`): Config → Logger → DeepLXService → DocxService → TaskManager → DocumentWorker → Handlers → Chi Router.

**Async Document Processing Flow:**
1. POST `/api/translate/document` → upload file, create task (Snowflake ID), enqueue to channel
2. Worker pool (default 5 workers, buffered channel capacity 100) picks up task
3. Task status: pending → processing → completed/failed (in-memory tracking)
4. Results stored in `uploads/` directory
5. Frontend polls `/api/tasks/{id}/status` for progress

**DOCX Translation Strategy** (`internal/service/docx.go` + `scripts/docx_helper.py`):
1. Go calls Python helper (`scripts/docx_helper.py`) via subprocess
2. Phase 1 (extract): Python uses `python-docx` to traverse all `<w:p>` elements (including table cells) in XML document order
3. Phase 2 (translate): Go batches paragraphs by character size (max 2000 chars/batch), translates each batch using numbered markers (`⟨⟨001⟩⟩ text` format). If batch translation fails, falls back to per-paragraph translation. 2-second delay between batches
4. Phase 3 (replace): Python writes translated text back preserving paragraph styles and run formatting
5. Multi-run paragraphs: translated text written to first run, remaining runs cleared (format from first run preserved)
6. Empty/whitespace paragraphs: preserved unchanged (no translation entry in map)

**TXT Translation**: Split into chunks by character size (max 2000 chars), preserving line boundaries where possible. Translate each chunk, join with newlines.

### Frontend Structure
```
frontend/src/
├── App.tsx                     # Router: / → Home (text), /document → Document (file)
├── components/                 # Header, Footer, DocumentUpload, TranslateForm, ui/
├── lib/
│   ├── api.ts                  # API client (VITE_API_URL env or /api default)
│   ├── constants.ts            # Language mappings
│   └── types.ts                # TypeScript types
└── pages/
    ├── Home.tsx                # Text translation
    └── Document.tsx            # Document upload with status polling
```

Vite config (`vite.config.ts`): Path alias `@` → `./src`, dev proxy `/api` and `/health` → `localhost:9448`.

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/translate` | POST | Text translation (max 10,000 chars) |
| `/api/translate/document` | POST | Upload document for async translation |
| `/api/tasks/{id}/status` | GET | Get task status and progress |
| `/api/tasks/{id}/download` | GET | Download translated document |

## Environment Variables

All config in `internal/config/config.go`, env vars with defaults. Key variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DEEPLX_URL` | `http://localhost:1188` | DeepLX service endpoint |
| `DEEPLX_TOKEN` | (empty) | Authentication token |
| `SERVER_PORT` | `9448` | Backend port (binds 127.0.0.1) |
| `LOG_LEVEL` | `info` | debug/info/warn/error |
| `WORKER_COUNT` | `5` | Document worker pool size |
| `MAX_FILE_SIZE` | `10MB` | Max upload file size |
| `TASK_MAX_AGE` | `1h` | Task retention period |
| `TASK_CLEANUP_INTERVAL` | `5m` | Task cleanup frequency |

Full list in `.env.example`.

## Key Dependencies

**Backend**: `go-chi/chi/v5` (router), `go-chi/cors`, `zap` (logging), `lumberjack` (log rotation), `bwmarrin/snowflake` (task IDs)

**Python**: `python-docx` (DOCX structured read/write via subprocess)

**Frontend**: React 18, TypeScript, Vite 5, Tailwind CSS 3, react-router-dom v6, lucide-react (icons)

## Deployment Notes

- Backend binds to `127.0.0.1` — nginx proxy required for external access
- `start.sh` builds frontend, copies `dist/` to `~/DockerUse/nginx/html/deeplx-html/`, runs backend on port 8449
- Docker: Backend on `:8449`, Frontend (nginx) on `:9449`, SSL certs in `frontend/ssl/`
- Docker volumes: `./uploads` and `./logs` persist on host

## Debugging

- Check `backend/logs/app.log` (Zap structured logs with task_id, file_name context)
- Search by task ID to trace full lifecycle
- Use curl commands in `api.md` to test DeepLX API directly
- Log rotation: 100MB max, 3 backups, 7-day retention

## Common Tasks

### Adding New Translation Features
1. `internal/service/deeplx.go` — API client changes
2. `internal/handler/translate.go` — Request/response handling
3. `frontend/src/lib/api.ts` — Frontend API client

### Modifying Document Processing
1. `internal/service/docx.go` — Core DOCX/TXT processing logic
2. `internal/worker/document_worker.go` — Worker pool handling
3. `internal/service/task_manager.go` — Task lifecycle

### Adding New API Endpoints
1. Create handler in `internal/handler/`
2. Register route in `cmd/server/main.go`
3. Add frontend API client in `frontend/src/lib/api.ts`
