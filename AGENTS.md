# LinkUp Server — AGENTS.md

## Quick start

```bash
go build -o ./tmp/main.exe ./cmd   # or `air` for hot reload
go build ./cmd/seed && ./seed.exe  # full seed run (drops & recreates 32 tables, resets DB)
go build ./...                      # verify all packages compile
```

Tests in `tests/`:
```bash
go test ./tests/... -v              # unit tests (no DB needed)
go test ./tests/... -run TestRegisterHandler_Success   # requires TEST_DSN env var
```

## Architecture

```
cmd/main.go → controller → service → repository (GORM)
cmd/seed/   → raw database/sql (10 ordered steps, resets DB)
cmd/cloudinary-check/ → standalone Cloudinary connectivity test
ws/          → gorilla/websocket Hub (per-user broadcast, token auth via `?token=`)
```

- **Framework**: Gin (`gin.New()` then `.Use(gin.Logger(), gin.Recovery())` in `cmd/main.go:44-45`).
- **DB**: `db.ConnectDb(env)` returns `*sql.DB` (DSN has **no TLS params**); `main.go` wraps with `gorm.Open(mysql.New(mysql.Config{Conn: database}), ...)`.
- **All model IDs are `string` (UUID)**. Foreign keys (`UserID`, `PostID`, etc.) are `string`/`*string`. Models have `json` tags; `db` tags are unused. `gorm` tags appear only in 4 models: `post` (`->` read-only computed counts), `password_history` (`primaryKey`, `index`), `post_share` (`primaryKey`), `notification_preference` (`primaryKey`).
- **32 model files** matching 32 schema tables (seed schema in `cmd/seed/schema/main.go`).

## Config

- `.env` loaded by `config.LoadEnv()` (custom line parser — **not** godotenv). Singleton guard prevents reloads.
- Required env vars: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, `JWT_EXPIRES_IN`, `CLOUDINARY_URL`. `DB_SSL` treated as required (validated as `false` → missing).
- `PORT` defaults to `"8080"`. Optional Gmail: `GMAIL_USER`, `GMAIL_PASSWORD`, `FRONTEND_RESET_URL` (default `http://localhost:3000`).
- `config.LoadCloudinaryEnv()` provides fallback from `CLOUDINARY_CLOUD_NAME`/`API_KEY`/`API_SECRET` but `CLOUDINARY_URL` is primary.
- `config.GetEnv()` returns a **value copy**.

## Routes

All routes wired in `cmd/main.go` (inside `if database != nil { ... }` guard). WebSocket is outside the guard.

| Path | Method | Auth | Handler file |
|---|---|---|---|
| `/health` | GET | No | inline in `main.go` |
| `/ws` | GET | token query | `ws/handler.go` |
| `/api/auth/register` | POST | No | `routes/auth.routes.go` |
| `/api/auth/login` | POST | No | |
| `/api/auth/change-password` | POST | `AuthMiddleware` | |
| `/api/auth/forgot-password` | POST | No | `routes/password_reset.routes.go` |
| `/api/auth/verify-reset-token` | POST | No | |
| `/api/auth/reset-password` | POST | No | |
| `/posts` | GET | No | `routes/post.routes.go` |
| `/posts` | POST | `AuthMiddleware` | |
| `/posts/:id` | GET | No | |
| `/posts/:id/react` | POST | `AuthMiddleware` | |
| `/posts/:id/comments` | POST | `AuthMiddleware` | |
| `/api/profile` | GET/PUT | `AuthMiddleware` | `routes/profile.routes.go` |
| `/api/profile/:userID` | GET | No | |
| `/api/follow/:userID` | POST | `AuthMiddleware` | `routes/follow.routes.go` |
| `/api/follow/stats/:userID` | GET | No | |
| `/api/media/upload` | POST | `AuthMiddleware` | `routes/media.routes.go` |
| `/api/media/storage` | GET | `AuthMiddleware` | |
| `/api/reports` | POST | `AuthMiddleware` | `routes/report.routes.go` |
| `/api/blocks` | POST/GET | `AuthMiddleware` | `routes/block.routes.go` |
| `/api/search` | GET | No | `routes/search.routes.go` |
| `/api/notifications` | GET | `AuthMiddleware` | `routes/notification.routes.go` |
| `/api/notifications/:id/read` | PUT | `AuthMiddleware` | |
| `/api/notifications/read-all` | PUT | `AuthMiddleware` | |
| `/api/notifications/unread-count` | GET | `AuthMiddleware` | |
| `/api/notifications/preferences` | GET/PUT | `AuthMiddleware` | |
| `/api/friend-requests` | GET | `AuthMiddleware` | `routes/friend.routes.go` |
| `/api/friend-requests/:userID` | POST | `AuthMiddleware` | |
| `/api/friend-requests/:id/accept` | PUT | `AuthMiddleware` | |
| `/api/friend-requests/:id` | DELETE | `AuthMiddleware` | |

Auth middleware (`middlewares/auth.middleware.go`) uses `utils.ParseToken`, sets `userID` and `email` on context.

## DTOs & validation

- `validations` package uses explicit validation (sentinel errors + struct methods), **not** Gin binding tags.
- `binding` tags used only in controller-level input structs: `post.controller.go`.
- Query params via `c.ShouldBindQuery(&input)` with `form:` tags (e.g. `dto/search.dto.go:SearchInput`).
- Password: ≥8, ≤128, must have upper, lower, digit, special.

## Business logic constraints

- **Role protection** — `repository/auth.repository.go:HasRole` checks `user_roles` via JOIN. `ReportService` and `BlockService` reject targeting `SUPER_ADMIN` or `ADMIN` role users. `SearchRepository.SearchUsers` excludes them via `NOT EXISTS`.
- **Post status** — `ReportService.CreateReport` validates post exists & is active via `postRepo.FindByID` (only returns active posts).
- **Toggle pattern** — `BlockService.ToggleBlock` and `FollowService.FollowToggle`: check existing record → delete if found, create if not.
- **Reaction toggle** — `postService.ReactPost` same pattern (check → delete existing or create new).

## JWT

- `utils.GenerateTokenPair` — HS256, access TTL from `JWTExpiresIn` (minutes, fallback 15min), refresh TTL hardcoded to 7 days.
- `utils.ParseToken` → `*utils.TokenClaims` (`UserID`, `Email`, `TokenType`).

## Password reset flow

1. `POST /api/auth/forgot-password` — 32-byte hex token in `password_reset_tokens`, email via Gmail SMTP (`smtp.gmail.com:587`, Vietnamese template).
2. `POST /api/auth/verify-reset-token` — checks validity & expiry.
3. `POST /api/auth/reset-password` — updates password, marks token used.
Token expiry: 10 min.

## Seed system

```bash
go build ./cmd/seed && ./seed.exe
```

10 ordered steps, raw SQL (not GORM), **drops all 32 tables on run**: reset → schema → users → core → profiles → social → relationships → messaging → moderation → extended. Steps share data via `internal.SeedState`. UUIDs via `internal.UUID()` (crypto/rand, RFC 9562). All users have bcrypt `Password123!`.

## WebSocket

- Hub at `ws/hub.go` — per-user broadcast via `userID → []Client` map. `SendToUser` serializes `OutgoingMessage{Type, Data}` to JSON.
- Auth via `?token=` query param (access JWT). No auth middleware — custom JWT parse in handler.
- Client is write-only from server perspective (reads and discards incoming messages).
- Used by `notificationService.Create` in follow/post service to push real-time notifications.

## Stubs / not wired

- `controllers/user.controller.go` — empty file
- `repository/user.repository.go` — has `Create`, `FindByEmail` but **not wired** in `cmd/main.go`
- `cmd/cloudinary-check/` — standalone binary

## Notable quirks

- **UUID generation diverges**: most services use `utils.GenerateUUID()` (crypto/rand), but `media.service.go` uses `github.com/google/uuid`.
- **PostService is an interface** (`services/post.service.go`), `MediaService` is also an interface; all other services use concrete structs.
- **`CLOUDINARY_URL` loaded in `LoadEnv`** (required), with `LoadCloudinaryEnv()` providing fallback for individual `CLOUDINARY_CLOUD_NAME`/`API_KEY`/`API_SECRET` vars.

## Key packages

| Package | Files | Role |
|---|---|---|
| `config/` | 2 | Env loading, Cloudinary creds |
| `controllers/` | 12 | HTTP handlers (1 stub) |
| `services/` | 11 | Business logic |
| `repository/` | 13 | GORM data access (1 unwired) |
| `dto/` | 10 | Request/response structs |
| `validations/` | 6 | Explicit validation |
| `utils/` | 5 | JWT, bcrypt, UUID, username gen, email |
| `models/` | 32 | GORM model structs |
| `routes/` | 11 | Route registration |
| `middlewares/` | 1 | Gin JWT auth middleware |
| `cmd/seed/` | 13 | Seed scripts + internal helpers |
| `db/` | 1 | MySQL connection |
| `ws/` | 3 | WebSocket hub, client, handler |
