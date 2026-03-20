# go-chatbot-api

`go-chatbot-api` is the backend for a company-scoped AI chatbot platform. It gives administrators a control plane for companies, users, chatbots, knowledge files, access grants, and token budgets, while end users get a chat runtime limited to the chatbots assigned to their company.

The service is written in Go, uses Huma on top of `http.ServeMux`, persists state in PostgreSQL, relies on Supabase for authentication and file storage, and talks to an OpenAI-compatible chat completion API.

## What the API does

- Creates, updates, lists, and deletes companies.
- Tracks per-company monthly token budgets and current usage.
- Invites, disables, and deletes users through Supabase Auth plus a local user profile store.
- Creates and manages chatbots with system prompts.
- Uploads chatbot knowledge files in `PDF`, `TXT`, `DOCX`, and `MD` formats.
- Extracts text from uploaded files and injects that text directly into prompt context.
- Grants chatbot access per company.
- Persists conversations and messages for end users.
- Blocks new chat turns when a company's token budget is exhausted.

## Architecture

The project follows a hexagonal layout with inward-only dependencies:

```text
adapters/ -> app/ -> domain/
```

Key directories:

```text
cmd/
  server/              HTTP server bootstrap
  bootstrap-admin/     Idempotent admin bootstrap command
domain/
  model/               Core entities and pure domain logic
  errors/              Domain errors
  ports/               Interfaces consumed by app services
app/                   Use-case orchestration
adapters/
  http/                Huma handlers and middleware
  postgres/            sqlc-generated queries and repository adapters
  openai/              OpenAI-compatible chat client
  supabaseauth/        Supabase Auth adapter
  supabasestorage/     Supabase Storage adapter
  files/               PDF/TXT/MD/DOCX text extraction
  logging/             Logger adapter
  system/              Clock, ID generation, token counting
config/                Environment-based configuration
migrations/            Embedded SQL migrations
```

## Runtime behavior

- `cmd/server` loads config, validates it, opens PostgreSQL, runs embedded migrations, wires adapters and services, and starts the HTTP API.
- `cmd/bootstrap-admin` runs the same database migrations and ensures an initial admin exists in both Supabase Auth and `user_profiles`.
- Chatbot knowledge is not indexed with vector search in this MVP. Extracted file text is appended directly to the system prompt for chat completions.
- Conversation history is persisted and trimmed before inference using an approximate token counter.
- OpenTelemetry support is built in. When enabled, traces are exported to stdout.

## Requirements

- Go `1.25` or newer
- PostgreSQL
- A Supabase project with Auth and Storage enabled
- A Supabase service role key
- An OpenAI-compatible chat completion endpoint and API key

## Quick start

1. Provision PostgreSQL and a Supabase project.
2. Export the required environment variables.
3. Bootstrap the first admin user.
4. Start the API server.

Example local setup:

```bash
export DATABASE_DSN="postgres://postgres:postgres@localhost:5432/chatbot?sslmode=disable"
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"
export SUPABASE_INVITE_REDIRECT_URL="http://localhost:3000/auth/invite"
export AI_API_KEY="your-openai-compatible-api-key"

# Optional overrides
export AI_BASE_URL="https://api.openai.com/v1"
export AI_MODEL="gpt-4.1-mini"
export SERVER_HOST="0.0.0.0"
export SERVER_PORT="8080"

# Required only for the bootstrap command
export BOOTSTRAP_ADMIN_EMAIL="admin@example.com"
export BOOTSTRAP_ADMIN_PASSWORD="change-me-now"
```

Bootstrap the initial admin:

```bash
go run ./cmd/bootstrap-admin
```

Start the server:

```bash
go run ./cmd/server
```

Health check:

```bash
curl http://localhost:8080/healthz
```

Both binaries apply embedded migrations automatically on startup.

## Configuration

Required for `cmd/server`:

| Variable | Notes |
| --- | --- |
| `DATABASE_DSN` | PostgreSQL connection string. |
| `SUPABASE_URL` | Base URL of the Supabase project. |
| `SUPABASE_SERVICE_ROLE_KEY` | Used for Supabase Auth admin APIs and Storage. Keep it server-side only. |
| `AI_API_KEY` | API key for the OpenAI-compatible provider. |

Required only for `cmd/bootstrap-admin`:

| Variable | Notes |
| --- | --- |
| `BOOTSTRAP_ADMIN_EMAIL` | Email for the initial admin user. |
| `BOOTSTRAP_ADMIN_PASSWORD` | Password for the initial admin user. |

Common optional variables:

| Variable | Default |
| --- | --- |
| `SERVER_HOST` | `0.0.0.0` |
| `SERVER_PORT` | `8080` |
| `SERVER_READ_TIMEOUT` | `5s` |
| `SERVER_WRITE_TIMEOUT` | `30s` |
| `SERVER_SHUTDOWN_TIMEOUT` | `15s` |
| `SUPABASE_STORAGE_BUCKET` | `chatbot-files` |
| `SUPABASE_AUTH_USER_PATH` | `/auth/v1/user` |
| `SUPABASE_ADMIN_PATH` | `/auth/v1/admin` |
| `AI_BASE_URL` | `https://api.openai.com/v1` |
| `AI_MODEL` | `gpt-4.1-mini` |
| `AI_TIMEOUT` | `45s` |
| `AI_KNOWLEDGE_MAX_TOKENS` | `24000` |
| `AI_HISTORY_MAX_TOKENS` | `12000` |
| `UPLOAD_MAX_FILE_BYTES` | `10485760` |
| `UPLOAD_MAX_FILES_PER_BOT` | `5` |
| `UPLOAD_ALLOWED_TYPES` | `application/pdf,text/plain,text/markdown,application/vnd.openxmlformats-officedocument.wordprocessingml.document` |
| `LOG_LEVEL` | `info` |
| `OTEL_SERVICE_NAME` | `go-chatbot-api` |
| `OTEL_ENABLED` | `false` |

## HTTP API overview

Protected endpoints expect:

```text
Authorization: Bearer <supabase-access-token>
```

Public endpoint:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Basic liveness check. |

Admin endpoints under `/admin`:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/admin/companies` | List companies. |
| `POST` | `/admin/companies` | Create a company. |
| `PATCH` | `/admin/companies/{company_id}` | Update a company. |
| `DELETE` | `/admin/companies/{company_id}` | Delete a company. |
| `GET` | `/admin/companies/{company_id}/usage/current` | Get current monthly usage. |
| `POST` | `/admin/companies/{company_id}/usage/adjust` | Apply a manual usage adjustment. |
| `POST` | `/admin/companies/{company_id}/usage/reset` | Reset current monthly usage. |
| `GET` | `/admin/users` | List users. |
| `POST` | `/admin/users/invitations` | Invite a user. |
| `POST` | `/admin/users/{user_id}/disable` | Disable a user. |
| `DELETE` | `/admin/users/{user_id}` | Delete a user. |
| `GET` | `/admin/chatbots` | List chatbots with files and assignments. |
| `POST` | `/admin/chatbots` | Create a chatbot. |
| `PATCH` | `/admin/chatbots/{chatbot_id}` | Update a chatbot. |
| `DELETE` | `/admin/chatbots/{chatbot_id}` | Delete a chatbot. |
| `POST` | `/admin/chatbots/{chatbot_id}/files` | Upload one knowledge file using multipart form data. |
| `DELETE` | `/admin/chatbots/{chatbot_id}/files/{file_id}` | Delete a knowledge file. |
| `PUT` | `/admin/companies/{company_id}/chatbots/{chatbot_id}` | Grant company access to a chatbot. |
| `DELETE` | `/admin/companies/{company_id}/chatbots/{chatbot_id}` | Revoke company access to a chatbot. |

End-user endpoints under `/me`:

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/me` | Get the authenticated profile. |
| `GET` | `/me/chatbots` | List chatbots assigned to the user's company. |
| `GET` | `/me/chatbots/{chatbot_id}/conversation` | Get or create the active conversation. |
| `POST` | `/me/chatbots/{chatbot_id}/conversations` | Start a fresh conversation. |
| `POST` | `/me/chatbots/{chatbot_id}/messages` | Send a chat message and persist the reply. |

The API is registered with Huma, so OpenAPI metadata is generated from the handler definitions and struct tags in `adapters/http`.

## Knowledge file constraints

- Supported file formats: `PDF`, `TXT`, `DOCX`, `MD`
- Default per-file size limit: `10 MB`
- Default maximum files per chatbot: `5`
- Default total extracted knowledge budget per chatbot: `24000` tokens

## Testing

Run the test suite with:

```bash
go test ./...
```

The repository currently uses:

- table-driven unit tests with `testify`
- handler tests with `humatest`
- mocks generated into `mocks/`

## Notes for operators

- The server uses the Supabase service role key for admin auth operations and storage access. Do not expose that key to browsers or mobile clients.
- A request ID is returned in the `X-Request-ID` header.
- When a company exceeds its monthly token budget, chat requests fail with a domain-mapped `429` response.
