# LinkUp Server — AGENTS.md

> `.gitignore`d. Local-only. No CI.

## Build & run

```bash
go build -o ./tmp/main.exe ./cmd          # production build
air                                         # hot reload (builds ./cmd/main.go)
go build ./cmd/seed && ./seed.exe          # full seed (drops & recreates all tables)
go build ./... && go vet ./...              # verify & vet all packages
```

**Tests**:
```bash
go test ./tests/community/... -v           # validation-only, no DB
go test ./tests/contribution/... -v        # validation-only, no DB
go test ./tests/call/... -v                # validation-only, no DB
go test ./services/...                     # some need TEST_DSN env var; includes internal integration tests
```
`tests/chat/`, `tests/friend/`, `tests/post/` are empty dirs. No linter configured. `.air.toml` excludes `_test.go` from watch.

## Architecture

```
cmd/main.go → controller → service → repository (GORM)
middlewares/  → auth.middleware.go + rbac.middleware.go
cmd/seed/     → raw database/sql (10 ordered steps)
ws/           → gorilla/websocket Hub (per-user broadcast + chat rooms)
```

- **Module `linkup`** (Go 1.26.3, Gin). Run from repo root.
- **DB**: `db/mysql.go` returns `*sql.DB`; `cmd/main.go` wraps with `gorm.Open(mysql.New(mysql.Config{Conn: database}), ...)`. All code inside `if database != nil { ... }` guard — WS + health endpoint run without DB.
- **46 model files** (`models/*.model.go`), 39 tables in DB seed schema.
- **All model IDs are `string` (UUID)**. Foreign keys are `string`/`*string`.
- **Validation split**: DTOs use `binding` tags (community, group_chat, post:`ReactPostInput`, chat). Others use `validations` package (13 validators, sentinel errors, struct methods). Query params: `form:` tags + `c.ShouldBindQuery`.
- **RBAC**: `RequireRoles` checks platform roles (`user_roles` JOIN, scope_id IS NULL). `CheckAdOwnership` guards ads for PARTNERs. `RequireContributionLevel` checks community contribution score threshold.
- **Contribution system**: `PostService.SetContributionService` wired after `ContributionService` init in `cmd/main.go`.
- **`PostService`, `MediaService`, `AdService` are interfaces** in `services/`. `ws.CallService` and `ws.ChatService` are also interfaces (defined in `ws/` package to avoid import cycles). All other services use concrete structs.
- **Toggle pattern**: BlockService.ToggleBlock, FollowService.FollowToggle, FriendService.ToggleFriendRequest, postService.ReactPost, VoiceCallService.ToggleMute/ToggleVideo: check existing → delete or create.
- **Atomic call ops**: `AcceptCallAtomic`/`RejectCallAtomic` use conditional `UPDATE ... WHERE status IN (?, ?)` — no TOCTOU. `CreateIfNotBusy` uses `SELECT COUNT(*) ... FOR UPDATE` (gap lock). InitiateCall has no redundant pre-check.
- **Batch profile load**: `GetCallHistoryFiltered` loads profiles in 2 queries (SELECT WHERE user_id IN ?), not JOIN. Capped at 100 IDs.
- **Call history soft-delete**: `call_hidden` table (composite PK call_id, user_id). Row stays for other party.
- **Idempotent schema**: `addColumnIfMissing`/`addIndexIfMissing`/`addForeignKeyIfMissing` query `information_schema` before ALTER. DDL identifiers validated via regex whitelist (`^[a-zA-Z_][a-zA-Z0-9_]*$`).
- **Error languages**: All services return Vietnamese. RBAC middleware returns English.

## Config quirks

- `.env` loaded by **custom line parser** (not godotenv). Singleton guard prevents reloads.
- Required: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, `JWT_EXPIRES_IN`, `CLOUDINARY_URL`.
- **`DB_SSL` bug**: `validateRequired` treats `"false"` as missing (`env.go:182`: `if !e.DBSSL`). Always set `DB_SSL=true`. Also not wired into the actual DSN (`db/mysql.go` ignores it).
- `PORT` defaults to `"8080"`. Optional: `GMAIL_USER`, `GMAIL_PASSWORD`, `FRONTEND_RESET_URL` (default `http://localhost:3000`), `WS_ALLOWED_ORIGINS` (comma-separated, default `*`).
- `config.GetEnv()` returns a **value copy** — mutations don't affect the singleton.
- `utils/email.go` reads `GMAIL_USER`/`GMAIL_PASSWORD` via `os.Getenv` directly (not from `config.Env`). New email features should follow the same pattern. `password_reset.service.go` also reads `FRONTEND_RESET_URL` via `os.Getenv`.
- `CLOUDINARY_URL` is primary; `LoadCloudinaryEnv()` falls back to `CLOUDINARY_CLOUD_NAME`/`API_KEY`/`API_SECRET`.

## Routes

Auth middleware sets `userID`/`email` on Gin context. Uses `Bearer` token in `Authorization` header. `cmd/main.go` has Vietnamese comments.

| Path | Auth | File(s) |
|---|---|---|
| `/health` | No | `cmd/main.go` inline |
| `/ws` | `?token=` | `ws/handler.go` |
| `/api/auth/*` | varies | `auth.routes.go`, `password_reset.routes.go` |
| `/posts`, `/posts/:id/*` | POST+Auth, GET=No | `post.routes.go` |
| `/api/tags/:name/posts` | No | `tag.routes.go` |
| `/api/profile*` | PATCH=Auth, GET/:userID=No | `profile.routes.go` |
| `/api/follow/*` | Auth | `follow.routes.go` |
| `/api/media/*` | Auth | `media.routes.go` |
| `/api/reports` | Auth | `report.routes.go` |
| `/api/blocks` | Auth | `block.routes.go` |
| `/api/search` | No | `search.routes.go` |
| `/api/notifications*` | Auth | `notification.routes.go` |
| `/api/friend-requests*` | Auth | `friend.routes.go` |
| `/api/chats/*` (incl. `/ws`) | Auth | `chat.routes.go`, `chat.controller.go` |
| `/api/group-chats/*` | Auth | `group_chat.routes.go` |
| `/api/communities*` | Auth | `community.routes.go`, `community_rule.routes.go` |
| `/api/communities/:id/policy\|challenges\|contributions` | Auth\* | `contribution.routes.go` |
| `/ads-management*` | Auth+RBAC | `ad.routes.go` |
| `/customer/*` | Auth | `ad.routes.go` |
| `/api/admin/*` | Auth | `admin.routes.go` |
| `/api/calls/*` | Auth | `call.routes.go` |

\* Contribution GET /leaderboard and /:userID are public (no Auth).

## WebSocket

**Two Hub instances, three WS endpoints, unified `ws.Hub` type:**

| Endpoint | Hub | Service | Auth | Purpose |
|---|---|---|---|---|
| `GET /ws` | `hub` (notification) | `service=nil, callService=nil` → call events skipped, chat events discarded | `?token=` access JWT | Real-time notifications |
| `GET /api/chats/ws` | `chatHub` | `ChatService` set → processes chat events | `AuthMiddleware` (Bearer) | Encrypted direct/group chat |
| `GET /api/calls/ws` | `hub` (notification, shared) | `callService` set, `service=nil` → processes call events | `AuthMiddleware` (Bearer) | WebRTC signaling |

- **Import cycle avoided**: `services` imports `ws`; `ws/chat.service.go` and `ws/call.service.go` define interfaces implemented in `services/`.
- **Chat WS events**: `chat:join`, `message:send`, `typing:start/stop`, `message:delete`, `message:search`.
- **Call WS events** (client→server): `call:initiate`, `call:accept`, `call:reject`, `call:end`, `call:signal`, `call:busy`, `call:video_toggle`, `call:toggle_mute`.
- **Call WS events** (server→client): `call:incoming`, `call:status`, `call:busy`, `call:mute`, `call:video`, `call:signal`, `call:missed`, `call:cancelled`.
- **Chat messages** use AES-256-GCM encryption (`utils/encryption.go`). Key stored per-chat (`chat.model.go:EncryptionKey`).

## Seed system

10 ordered steps (`cmd/seed/main.go`): reset → schema → users → core → profiles → social → relationships → messaging → moderation → extended. Raw SQL (not GORM), drops all tables. Steps share data via `internal.SeedState`. UUIDs via `internal.UUID()` (crypto/rand, RFC 9562). All seed users have bcrypt `Password123!`. Relationships step also seeds `community_rules` for each community.

**Schema sub-package** (`cmd/seed/schema/`): 39 `CREATE TABLE IF NOT EXISTS` statements, then idempotent ALTER TABLE via `addColumnIfMissing`/`addIndexIfMissing`/`addForeignKeyIfMissing` (queries `information_schema` first). DDL identifiers validated via regex whitelist.

## Password reset flow

3-step: `forgot-password` (token in DB, email via Gmail SMTP Vietnamese template, 10 min expiry) → `verify-reset-token` → `reset-password`.

## JWT

`utils.GenerateTokenPair` — HS256, access TTL from `JWTExpiresIn` (minutes, fallback 15), refresh TTL 7 days. `utils.ParseToken` → `*utils.TokenClaims` (`UserID`, `Email`, `TokenType`). Separate `utils.GenerateToken` for single tokens (reset). Auth has `/api/auth/refresh` endpoint.

## Stubs / not wired

- `controllers/user.controller.go` — empty. `repository/user.repository.go` has `Create`, `FindByEmail` but not wired in `cmd/main.go`.
- `cmd/cloudinary-check/` — standalone binary, not part of the app.
- `dto/auth.dto.go` — no `binding` tags (auth validation delegated to `validations.AuthValidation`). Only `RefreshTokenInput` has `binding:"required"`.
- `docs/` — 5 prose files (admin/user function specs, voice-call docs, community plan) + `docs/test-case/call-history.test-case.md` (110 test cases). Not wired code.

## Quirks

- **UUID divergence**: `utils.GenerateUUID()` (crypto/rand, RFC 9562) used by most services. `ad.service.go` has `uuidGenerate()` using crypto/rand with `ad_` prefix. `github.com/google/uuid` is indirect dep only.
- **Air config** (`.air.toml`) builds `cmd/main.go` specifically; the Dockerfile builds `./cmd` (package, not file).
- **Build artifacts committed**: `cmd.exe`, `seed.exe`, `cloudinary-check.exe` in repo root.
- **gorm tags**: 46 model files use them for indexes, computed columns (`->`), PKs. Models primarily use `json` tags; `db` tags unused.
- **DB_SSL not wired**: `validateRequired` treats `"false"` as missing, but even when set, `db/mysql.go` ignores it in the DSN.
- **WsConfig note**: `ws/handler.go` reads `WS_ALLOWED_ORIGINS` at import time (`os.Getenv`). `controllers/call.controller.go` has its own `callUpgrader` (decoupled from chat controller's). `controllers/chat.controller.go`'s upgrader still hardcodes `CheckOrigin: return true`.
