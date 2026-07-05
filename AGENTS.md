# LinkUp Server — AGENTS.md

> **Note**: This file is `.gitignore`d (line 61). It is local-only and won't be committed.

## Build & run

```bash
go build -o ./tmp/main.exe ./cmd          # production build
air                                         # hot reload (`.air.toml` ready)
go build ./cmd/seed && ./seed.exe          # full seed (drops & recreates all tables)
go build ./... && go vet ./...              # verify & vet all packages
```

**Tests** — validation-only (no DB) in `auth`, `community`, and `contribution`:
```bash
go test ./tests/auth/... -v -run TestValidate
go test ./tests/community/... -v
go test ./tests/contribution/... -v
go test ./tests/... -run TestRegisterHandler_Success  # needs TEST_DSN env var
```
`tests/chat/`, `tests/friend/`, `tests/post/` are empty. No linter — `go build && go vet` is the verification gate. `services/contribution_internal_test.go` is an internal package test, run with `go test ./services/...`.

## Architecture

```
cmd/main.go → controller → service → repository (GORM)
middlewares/  → auth.middleware.go + rbac.middleware.go (RequireRoles, CheckAdOwnership)
cmd/seed/     → raw database/sql (10 ordered steps)
ws/           → gorilla/websocket Hub (per-user broadcast + chat rooms)
docs/         → admin/user function specification tables
```

- **Gin** with `gin.New()` + `gin.Logger(), gin.Recovery()`.
- **DB**: `db/mysql.go` returns `*sql.DB` (DSN: `parseTime=true&charset=utf8mb4`, no TLS); `cmd/main.go` wraps with `gorm.Open(mysql.New(mysql.Config{Conn: database}), ...)`.
- **Module**: `linkup` (Go 1.26.3), run from repo root.
- **All model IDs are `string` (UUID)**. Foreign keys are `string`/`*string`. 43 model files.
- **Validation split**: DTOs use `binding` tags (`community`, `community_rule`, `group_chat`, `post:ReactPostInput`, `chat:3 structs`). Others use `validations` package (13 validators, sentinel errors, struct methods). Query params: `form:` tags + `c.ShouldBindQuery`.
- **RBAC**: `middlewares/rbac.middleware.go` has `RequireRoles` (platform-level: SUPER_ADMIN/ADMIN/PARTNER via `user_roles` JOIN) and `CheckAdOwnership` (ad ownership guard for PARTNERs).
- **Contribution system**: `PostService.SetContributionService` wired after `ContributionService` init. Community challenges, policies, and leaderboard.

## Config quirks

- `.env` loaded by `config.LoadEnv()` — **custom line parser** (not godotenv). Singleton guard prevents reloads.
- Required: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, `JWT_EXPIRES_IN`, `CLOUDINARY_URL`.
- **`DB_SSL` bug**: `validateRequired` treats `"false"` as missing. Always set `DB_SSL=true`. Also, `DB_SSL` is not wired into the actual DSN (`db/mysql.go` ignores it). It was meant for TLS but never connected.
- `PORT` defaults to `"8080"`. Optional: `GMAIL_USER`, `GMAIL_PASSWORD`, `FRONTEND_RESET_URL` (default `http://localhost:3000`).
- `config.GetEnv()` returns a **value copy** — mutations to the returned struct don't affect the singleton.
- `CLOUDINARY_URL` is primary; `LoadCloudinaryEnv()` falls back to `CLOUDINARY_CLOUD_NAME`/`API_KEY`/`API_SECRET`.

## Routes

All wired in `cmd/main.go` inside `if database != nil { ... }` guard (WS + health are outside). Auth middleware sets `userID`/`email` on Gin context. Uses `Bearer` token in `Authorization` header. `cmd/main.go` has Vietnamese comments throughout.

| Path | Method | Auth | File |
|---|---|---|---|
| `/health` | GET | No | `cmd/main.go` inline |
| `/ws` | GET | `?token=` query | `ws/handler.go` |
| `/api/auth/*` | various | varies | `routes/auth.routes.go`, `routes/password_reset.routes.go` |
| `/api/tags/:name/posts` | GET | No | `routes/tag.routes.go` |
| `/posts` | GET/POST | POST=Auth | `routes/post.routes.go` |
| `/posts/:id` | GET | No | |
| `/posts/:id/react\|comments\|share\|save` | POST | Auth | |
| `/api/profile` | GET/PATCH | Auth | `routes/profile.routes.go` |
| `/api/profile/:userID` | GET | No | |
| `/api/follow/*` | POST/GET | Auth | `routes/follow.routes.go` |
| `/api/media/*` | all | Auth | `routes/media.routes.go` |
| `/api/reports` | POST | Auth | `routes/report.routes.go` |
| `/api/blocks` | POST/GET | Auth | `routes/block.routes.go` |
| `/api/search` | GET | No | `routes/search.routes.go` |
| `/api/notifications*` | all | Auth | `routes/notification.routes.go` |
| `/api/friend-requests*` | GET/POST/PUT/DELETE | Auth | `routes/friend.routes.go` |
| `/api/chats/*` | POST/GET/DELETE | Auth | `routes/chat.routes.go` |
| `/api/group-chats/*` | POST/GET/PUT/DELETE | Auth | `routes/group_chat.routes.go` |
| `/api/communities*` | POST/GET/PUT/DELETE | Auth | `routes/community.routes.go`, `routes/community_rule.routes.go`, `routes/contribution.routes.go` |
| `/ads-management` | all | Auth+RBAC | `routes/ad.routes.go` |
| `/customer/feed\|/customer/ads/:id/track` | GET/POST | Auth | |

## Business logic conventions

- **Role protection**: `repository/auth.repository.go:HasRole` checks `user_roles` via JOIN. `RequireRoles` middleware checks platform roles. `ReportService` and `BlockService` reject targeting `SUPER_ADMIN` or `ADMIN`. `SearchRepository.SearchUsers` excludes them via `NOT EXISTS`.
- **Toggle pattern**: `BlockService.ToggleBlock`, `FollowService.FollowToggle`, `FriendService.ToggleFriendRequest`, `postService.ReactPost`: check existing → delete if found, else create.
- **Post status**: `FindByID` only returns active posts.
- **Error languages**: post/friend/chat/community services return Vietnamese error strings. Auth service returns English.

## Service patterns

- **`PostService` and `MediaService` are interfaces** (in `services/`). All other services use concrete structs.
- **`friendService`** (`NewFriendService`) takes 5: `friendRepository`, `authRepository`, `profileRepository`, `friendValidation`, `notificationService`.
- **`chatService`** (`NewChatService`) takes 6: `chatRepository`, `friendRepository`, `inviteRepository`, `mediaRepository`, `notificationService`, `chatValidation`.
- **`notificationService.Create`** is called from follow/friend/post services to push real-time via WebSocket Hub.
- **Contribution wiring**: `PostService` embeds `ContributionService` via `SetContributionService` (called in `cmd/main.go` after both are constructed).

## WebSocket

Two endpoints, two separate Hub instances, one unified `ws.Hub` type:

| Endpoint | Handler | Hub | Client type | Auth |
|---|---|---|---|---|
| `GET /ws` | `ws/handler.go:ServeWS` | `hub` (notification) | `Client` with `service=nil` → reads discarded | `?token=` access JWT |
| `GET /api/chats/ws` | `controllers/chat.controller.go:HandleWebsocket` | `chatHub` | `Client` with `ChatService` → processes chat events | `AuthMiddleware` |

- **`ws/hub.go`**: `rooms` (chat broadcast) + `clients` (per-user notification). Methods: `SendToUser`, `JoinChat`, `RegisterClient`.
- **Import cycle avoided**: `services` imports `ws`; `ws` does NOT import `services` — instead `ws/chat.service.go` defines a `ChatService interface` that `services/chat.service.go` implements implicitly.
- **`ws/client.go`**: handles `chat:join`, `message:send`, `typing:start/stop`, `message:delete`, `message:search` when `service != nil`; otherwise discards incoming (notification-only client).
- **Chat messages** use AES-256-GCM encryption (`utils/encryption.go`).

## Seed system

10 ordered steps (`cmd/seed/main.go`): reset → schema → users → core → profiles → social → relationships → messaging → moderation → extended. Raw SQL (not GORM), drops all 34+ tables. Steps share data via `internal.SeedState`. UUIDs via `internal.UUID()` (crypto/rand, RFC 9562). All seed users have bcrypt `Password123!`. Relationships step also seeds `community_rules` for each community.

## Password reset flow

3-step: `forgot-password` (token in DB, email via Gmail SMTP Vietnamese template, 10 min expiry) → `verify-reset-token` → `reset-password`.

## JWT

`utils.GenerateTokenPair` — HS256, access TTL from `JWTExpiresIn` (minutes, fallback 15), refresh TTL 7 days. `utils.ParseToken` → `*utils.TokenClaims` (`UserID`, `Email`, `TokenType`). Separate `utils.GenerateToken` for single tokens (reset). Auth has `/api/auth/refresh` endpoint for token refresh.

## Stubs / not wired

- `controllers/user.controller.go` — empty. `repository/user.repository.go` — `Create`, `FindByEmail` exist but **not wired** in `cmd/main.go`.
- `cmd/cloudinary-check/` — standalone binary, not part of the app.
- `dto/auth.dto.go` — no `binding` tags. Auth validation is fully delegated to `validations.AuthValidation`.
- `docs/` — admin/user function spec tables (prose, not wired code).

## Quirks

- **UUID divergence**: `utils.GenerateUUID()` (crypto/rand, RFC 9562) is used by most services. `ad.service.go` has its own local `uuidGenerate()` using crypto/rand. `github.com/google/uuid` is an indirect dep only (cloudinary-go transitive).
- **`utils/email.go`** reads `GMAIL_USER`/`GMAIL_PASSWORD` via `os.Getenv` directly — not from `config.Env` struct (which also stores them unused). New email features should follow the same `os.Getenv` pattern or reconcile both paths.
- **Chat encryption**: AES-256-GCM (`utils/encryption.go`). Key stored per-chat in `chat.model.go:EncryptionKey` (generated via `GenerateEncryptionKey` at chat creation). No env var for it.
- **`gorm` tags** appear in 10+ model files for indexes, computed columns (`->`), and primary keys. Models primarily use `json` tags; `db` tags are unused.
- **`validations` package**: 13 validators (`auth`, `block`, `chat`, `comment`, `community`, `community_rule`, `contribution`, `friend`, `group_chat`, `media`, `post`, `report`, `search`) but not all services use them — some rely on `binding` tags or inline checks. `dto/auth.dto.go` has no `binding` tags except `RefreshTokenInput.binding:"required"`.
- **Air config** (`air.toml`) builds `cmd/main.go` specifically (not `./cmd` — the Dockerfile builds `./cmd`), excludes `_test.go` via regex.
- **No Makefile, Taskfile, or script runner**. Everything is manual `go` commands.
- **Build artifacts committed**: `cmd.exe`, `seed.exe`, `cloudinary-check.exe` in repo root.
