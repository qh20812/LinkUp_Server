# LinkUp Server — AGENTS.md

> `.gitignore`d (line 61). Local-only. No README, no CI, no Makefile.

## Build & run

```bash
go build -o ./tmp/main.exe ./cmd          # production build
air                                         # hot reload
go build ./cmd/seed && ./seed.exe          # full seed (drops & recreates all tables)
go build ./... && go vet ./...              # verify & vet all packages
```

**Tests**:
```bash
go test ./tests/community/... -v           # validation-only, no DB
go test ./tests/contribution/... -v        # validation-only, no DB
go test ./services/...                     # some need TEST_DSN env var
```
`tests/chat/`, `tests/friend/`, `tests/post/` are empty dirs. No linter configured.

## Architecture

```
cmd/main.go → controller → service → repository (GORM)
middlewares/  → auth.middleware.go + rbac.middleware.go
cmd/seed/     → raw database/sql (10 ordered steps)
ws/           → gorilla/websocket Hub (per-user broadcast + chat rooms)
```

- **Module `linkup`** (Go 1.26.3, Gin). Run from repo root.
- **DB**: `db/mysql.go` returns `*sql.DB`; `cmd/main.go` wraps with `gorm.Open(mysql.New(mysql.Config{Conn: database}), ...)`. All code inside `if database != nil { ... }` guard — WS + health endpoint run without DB.
- **All model IDs are `string` (UUID)**. Foreign keys are `string`/`*string`.
- **Validation split**: DTOs use `binding` tags (community, group_chat, post:`ReactPostInput`, chat). Others use `validations` package (13 validators, sentinel errors, struct methods). Query params: `form:` tags + `c.ShouldBindQuery`.
- **RBAC**: `RequireRoles` checks platform roles (`user_roles` JOIN, scope_id IS NULL). `CheckAdOwnership` guards ads for PARTNERs. `RequireContributionLevel` checks community contribution score threshold.
- **Contribution system**: `PostService.SetContributionService` wired after `ContributionService` init in `cmd/main.go`.
- **`PostService`, `MediaService`, `AdService` are interfaces** in `services/`. All other services use concrete structs.
- **Toggle pattern**: BlockService.ToggleBlock, FollowService.FollowToggle, FriendService.ToggleFriendRequest, postService.ReactPost: check existing → delete or create.
- **Error languages**: All services return Vietnamese. RBAC middleware returns English.

## Config quirks

- `.env` loaded by **custom line parser** (not godotenv). Singleton guard prevents reloads.
- Required: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, `JWT_EXPIRES_IN`, `CLOUDINARY_URL`.
- **`DB_SSL` bug**: `validateRequired` treats `"false"` as missing (`env.go:172`: `if !e.DBSSL`). Always set `DB_SSL=true`. Also not wired into the actual DSN (`db/mysql.go` ignores it).
- `PORT` defaults to `"8080"`. Optional: `GMAIL_USER`, `GMAIL_PASSWORD`, `FRONTEND_RESET_URL` (default `http://localhost:3000`).
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
| `/api/communities/:id/policy\|challenges\|contributions` | Auth* | `contribution.routes.go` |
| `/ads-management*` | Auth+RBAC | `ad.routes.go` |
| `/customer/*` | Auth | `ad.routes.go` |
| `/api/admin/*` | Auth | `admin.routes.go` |
| `/api/calls/*` | Auth | `call.routes.go` |

\* Contribution GET /leaderboard and /:userID are public (no Auth).

## WebSocket

Two endpoints, two Hub instances, one unified `ws.Hub` type:

| Endpoint | Handler | Hub | Chat client | Auth |
|---|---|---|---|---|
| `GET /ws` | `ws/handler.go:ServeWS` | `hub` (notifications) | `service=nil` → reads discarded | `?token=` access JWT |
| `GET /api/chats/ws` | `controllers/chat.controller.go:HandleWebsocket` | `chatHub` | `ChatService` set → processes events | `AuthMiddleware` |

- **Import cycle avoided**: `services` imports `ws`; `ws/chat.service.go` defines a `ChatService interface` that `services/chat.service.go` implements implicitly.
- WS events: `chat:join`, `message:send`, `typing:start/stop`, `message:delete`, `message:search`.
- **Chat messages** use AES-256-GCM encryption (`utils/encryption.go`). Key stored per-chat (`chat.model.go:EncryptionKey`).

## Voice/Video calls

Wired in `cmd/main.go` under `/api/calls/*` (all Auth). Uses the notification `hub` for signaling over WebSocket (`/api/calls/ws`). Separate `VoiceCallService` in `services/voice_call.service.go`.

## Seed system

10 ordered steps (`cmd/seed/main.go`): reset → schema → users → core → profiles → social → relationships → messaging → moderation → extended. Raw SQL (not GORM), drops all 34+ tables. Steps share data via `internal.SeedState`. UUIDs via `internal.UUID()` (crypto/rand, RFC 9562). All seed users have bcrypt `Password123!`. Relationships step also seeds `community_rules` for each community.

## Password reset flow

3-step: `forgot-password` (token in DB, email via Gmail SMTP Vietnamese template, 10 min expiry) → `verify-reset-token` → `reset-password`.

## JWT

`utils.GenerateTokenPair` — HS256, access TTL from `JWTExpiresIn` (minutes, fallback 15), refresh TTL 7 days. `utils.ParseToken` → `*utils.TokenClaims` (`UserID`, `Email`, `TokenType`). Separate `utils.GenerateToken` for single tokens (reset). Auth has `/api/auth/refresh` endpoint.

## Stubs / not wired

- `controllers/user.controller.go` — empty. `repository/user.repository.go` has `Create`, `FindByEmail` but not wired in `cmd/main.go`.
- `cmd/cloudinary-check/` — standalone binary, not part of the app.
- `dto/auth.dto.go` — no `binding` tags (auth validation delegated to `validations.AuthValidation`). Only `RefreshTokenInput` has `binding:"required"`.
- `docs/` — 5 prose files (admin/user function specs, voice-call docs, community plan). Not wired code.

## Quirks

- **UUID divergence**: `utils.GenerateUUID()` (crypto/rand, RFC 9562) used by most services. `ad.service.go` has `uuidGenerate()` using crypto/rand with `ad_` prefix. `github.com/google/uuid` is indirect dep only.
- **Air config** (`.air.toml`) builds `cmd/main.go` specifically (not `./cmd` — the Dockerfile builds `./cmd`).
- **Build artifacts committed**: `cmd.exe`, `seed.exe`, `cloudinary-check.exe` in repo root.
- **gorm tags**: 10+ model files use them for indexes, computed columns (`->`), PKs. Models primarily use `json` tags; `db` tags unused.
