# LinkUp Server — AGENTS.md

## Build & run

```bash
go build -o ./tmp/main.exe ./cmd          # production build
air                                         # hot reload (`.air.toml` ready)
go build ./cmd/seed && ./seed.exe          # full seed (drops & recreates all tables)
go build ./...                              # verify all packages compile
```

**Tests** — two categories:
```bash
go test ./tests/auth/... -v -run TestValidate  # validation-only, no DB
go test ./tests/... -run TestRegisterHandler_Success  # needs TEST_DSN env var
```

`tests/post/` is empty (no post tests yet). No linter/typecheck config — `go build ./...` is the main verification.

## Architecture

```
cmd/main.go → controller → service → repository (GORM)
cmd/seed/   → raw database/sql (10 ordered steps)
ws/          → gorilla/websocket Hub (per-user broadcast)
```

- **Framework**: Gin (`gin.New()`, `.Use(gin.Logger(), gin.Recovery())`).
- **DB**: `db/mysql.go` returns `*sql.DB` (DSN has **no TLS params**); `cmd/main.go` wraps with `gorm.Open(mysql.New(mysql.Config{Conn: database}), ...)`.
- **Module**: `linkup` (Go 1.26.3), run all build/test from repo root.
- **All model IDs are `string` (UUID)**. Foreign keys (`UserID`, `PostID`, etc.) are `string`/`*string`.
- **Validation**: `binding` tags in `dto/post.dto.go` (3 structs) and `dto/chat.dto.go` (3 structs). Elsewhere: explicit validation via `validations` package (sentinel errors, struct methods). Query params use `form:` tags with `c.ShouldBindQuery`.

## Config quirks

- `.env` loaded by `config.LoadEnv()` — **custom line parser** (not godotenv). Singleton guard prevents reloads.
- Required: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, `JWT_EXPIRES_IN`, `CLOUDINARY_URL`.
- **`DB_SSL` bug**: `validateRequired` treats `false` as missing. Always set `DB_SSL=true`.
- `PORT` defaults to `"8080"`. Optional: `GMAIL_USER`, `GMAIL_PASSWORD`, `FRONTEND_RESET_URL` (default `http://localhost:3000`).
- `config.GetEnv()` returns a **value copy**.
- `CLOUDINARY_URL` is primary; `LoadCloudinaryEnv()` falls back to individual `CLOUDINARY_CLOUD_NAME`/`API_KEY`/`API_SECRET`.

## Routes

All wired in `cmd/main.go` inside `if database != nil { ... }` guard (WS is outside). Auth middleware sets `userID` and `email` on Gin context. Uses `Bearer` token in `Authorization` header.

| Path | Method | Auth | Registration file |
|---|---|---|---|
| `/health` | GET | No | inline in `main.go` |
| `/ws` | GET | `?token=` query | `ws/handler.go` |
| `/api/auth/register` | POST | No | `routes/auth.routes.go` |
| `/api/auth/login` | POST | No | |
| `/api/auth/change-password` | POST | Auth | |
| `/api/auth/forgot-password` | POST | No | `routes/password_reset.routes.go` |
| `/api/auth/verify-reset-token` | POST | No | |
| `/api/auth/reset-password` | POST | No | |
| `/posts` | GET/POST | POST=Auth | `routes/post.routes.go` |
| `/posts/:id` | GET | No | |
| `/posts/:id/react` | POST | Auth | |
| `/posts/:id/comments` | GET/POST | POST=Auth | |
| `/posts/:id/share` | POST | Auth | |
| `/posts/:id/save` | POST | Auth | |
| `/api/profile` | GET/PATCH | Auth | `routes/profile.routes.go` |
| `/api/profile/:userID` | GET | No | |
| `/api/follow/:userID` | POST | Auth | `routes/follow.routes.go` |
| `/api/follow/stats/:userID` | GET | Auth | |
| `/api/media/*` | all | Auth | `routes/media.routes.go` |
| `/api/reports` | POST | Auth | `routes/report.routes.go` |
| `/api/blocks` | POST/GET | Auth | `routes/block.routes.go` |
| `/api/search` | GET | No | `routes/search.routes.go` |
| `/api/notifications*` | all | Auth | `routes/notification.routes.go` |
| `/api/friend-requests` | GET | Auth | `routes/friend.routes.go` |
| `/api/friend-requests/:userID` | POST | Auth | |
| `/api/friend-requests/:id/accept` | PUT | Auth | |
| `/api/friend-requests/:id` | DELETE | Auth | |
| `/api/chats/direct` | POST | Auth | `routes/chat.routes.go` |
| `/api/chats/invite` | POST | Auth | |
| `/api/chats/invite/respond` | POST | Auth | |
| `/api/chats/ws` | GET | Auth (middleware) | |

## Business logic conventions

- **Role protection**: `repository/auth.repository.go:HasRole` checks `user_roles` via JOIN. `ReportService` and `BlockService` reject targeting `SUPER_ADMIN` or `ADMIN`. `SearchRepository.SearchUsers` excludes them via `NOT EXISTS`.
- **Toggle pattern**: `BlockService.ToggleBlock`, `FollowService.FollowToggle`, `FriendService.ToggleFriendRequest`, `postService.ReactPost`: check existing → delete if found, else create.
- **Post status**: `FindByID` only returns active posts. Used by `ReportService` to validate post exists.
- **Vietnamese error messages**: post/friend/chat services return Vietnamese error strings. Auth service returns English.

## Service patterns

- **`PostService` and `MediaService` are interfaces** (in `services/`). All other services use concrete structs.
- **`notificationService.Create`** is called from follow/friend/post services to push real-time via WebSocket Hub.

## WebSocket

Two endpoints, two separate Hub instances, one unified `ws.Hub` type:

| Endpoint | Handler | Hub | Client type | Auth |
|---|---|---|---|---|
| `GET /ws` | `ws/handler.go:ServeWS` | `hub` (notification) | `Client` with `service=nil` → reads discarded | `?token=` access JWT |
| `GET /api/chats/ws` | `controllers/chat.controller.go:HandleWebsocket` | `chatHub` | `Client` with `ChatService` → processes chat events | `AuthMiddleware` |

- **`ws/hub.go`** unifies both patterns in one struct: `rooms map[string]map[*Client]bool` (chat broadcast) + `clients map[string]map[*Client]bool` (per-user notification). Has `SendToUser`, `JoinChat`, `RegisterClient`, and a `broadcast` channel.
- **Import cycle avoided**: `services` imports `ws` (for `*ws.Hub`, `ws.OutgoingMessage`), but `ws` does NOT import `services`. Instead `ws/chat.service.go` defines a `ChatService interface` that `services/chat.service.go` implements implicitly.
- **`ws/client.go`** handles chat events (`chat:join`, `message:send`, `typing:start/stop`) when `service != nil`; otherwise discards incoming (notification-only client).

## Seed system

10 ordered steps: reset → schema → users → core → profiles → social → relationships → messaging → moderation → extended. Raw SQL (not GORM), drops all 32 tables. Steps share data via `internal.SeedState`. UUIDs via `internal.UUID()` (crypto/rand, RFC 9562). All seed users have bcrypt `Password123!`.

## Password reset flow

1. `POST /api/auth/forgot-password` — 32-byte hex token in `password_reset_tokens`, email via Gmail SMTP (`smtp.gmail.com:587`, Vietnamese template).
2. `POST /api/auth/verify-reset-token` — checks validity & expiry.
3. `POST /api/auth/reset-password` — updates password, marks token used. Token expiry: 10 min.

## JWT

`utils.GenerateTokenPair` — HS256, access TTL from `JWTExpiresIn` (minutes, fallback 15), refresh TTL hardcoded to 7 days. `utils.ParseToken` → `*utils.TokenClaims` (`UserID`, `Email`, `TokenType`). A separate `utils.GenerateToken` exists for single-token generation (used by reset tokens).

## Stubs / not wired

- `controllers/user.controller.go` — empty (2 lines)
- `repository/user.repository.go` — `Create`, `FindByEmail` exist but **not wired** in `cmd/main.go`
- `cmd/cloudinary-check/` — standalone binary

## Quirks

- **UUID divergence**: most services use `utils.GenerateUUID()` (crypto/rand), but `media.service.go` uses `github.com/google/uuid`.
- **`gorm` tags** appear in only 4 models: `post` (computed `->`), `password_history`/`post_share`/`notification_preference` (`primaryKey`). Models use `json` tags; `db` tags are unused.
- **`validations` package**: 6 validators exist (`auth`, `block`, `friend`, `media`, `report`, `search`) but not all services use them — some just use `binding` tags or inline checks.
- **Air config** builds `cmd/main.go` specifically (not `./cmd`), excludes `_test.go` via regex.
