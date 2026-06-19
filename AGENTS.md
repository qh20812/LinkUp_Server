# LinkUp Server — AGENTS.md

## Quick start

```bash
go build -o ./tmp/main.exe ./cmd   # or `air` for hot reload
go build ./cmd/seed && ./seed.exe  # full seed run
go build ./...                      # verify all packages compile
```

No `_test.go` files exist anywhere.

## Architecture

```
cmd/main.go          → controller → service → repository (GORM)
cmd/seed/main.go     → raw database/sql (10 ordered steps)
cmd/cloudinary-check → standalone Cloudinary connectivity test
```

- **Framework**: Gin (`gin.New()`, no default middleware; `.Use(gin.Logger(), gin.Recovery())` at `cmd/main.go:36-37`)
- **DB**: `db.ConnectDb(env)` returns `*sql.DB`; `main.go` wraps with `gorm.Open(mysql.New(mysql.Config{Conn: database}), ...)` for GORM repos. DSN has **no TLS params** — `registerTLSConfig` in `db/mysql.go` is a dead stub.
- **All model IDs are `string` (UUID)**. Foreign keys (`UserID`, `PostID`, etc.) are `string`/`*string`. Models have `json` and `db` tags; `db` tags are unused relics.
- **DTOs have no `binding` tags** — validation is via `validations.AuthValidation` methods in the controller.
- **JWT**: `utils.GenerateTokenPair(secret, userID, email, accessTTL, refreshTTL)` — uses `golang-jwt/jwt/v5` HS256. Refresh TTL hardcoded to 7 days (`auth.service.go:106`).
- **`user.controller.go`** is an empty stub; **`user.repository.go`** has basic CRUD (Create, FindByEmail) but **not wired** in `cmd/main.go`. Only auth routes (`/api/auth/register`, `/api/auth/login`) are connected.
- **`middlewares/`** directory exists but is empty — no auth middleware yet.
- `config/cloundinary.go` is misnamed (should be `cloudinary`).

## Config

- `.env` loaded by `config.LoadEnv()` (custom line parser — **not** godotenv). Singleton guard prevents reloads.
- Required env vars (fail at load if empty): `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, `JWT_EXPIRES_IN`
  - `DB_SSL` parsed via `strconv.ParseBool` — `"true"`/`"false"` string required; not in `validateRequired` list but will fail at parse time.
  - `PORT` defaults to `"8080"` if unset.
- Cloudinary: `CLOUDINARY_URL` (`cloudinary://key:secret@cloudname`) is primary; fallback to individual vars (`CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`, `CLOUDINARY_API_SECRET`). Loaded via `config.LoadCloudinaryEnv()`.
- `config.GetEnv()` returns a value copy of the singleton `Env` struct.

## Seed system

```bash
go build ./cmd/seed && ./seed.exe
```

10 ordered steps, each opens its own `*sql.DB` using raw SQL (not GORM):

1. **reset** — DROP TABLE all 28 tables (`FOREIGN_KEY_CHECKS=0`)
2. **schema** — CREATE TABLE all 28 tables with FKs, indexes, utf8mb4
3. **users** — 20 users (bcrypt `Password123!`), 2 banned, 1 suspended
4. **core** — 3 roles, 10 emojis, 8 violation rules + `user_roles` assignments
5. **profiles** — 20 profiles with avatars and bios, 5 private
6. **social** — 30 posts, 60 comments (nested), 50 tags, 80 reactions, 40 follows, 15 friends, 5 blocks, 20 bookmarks
7. **relationships** — 5 communities, 25 group members with roles and points
8. **messaging** — 8 chats (direct + group), ~24 participants, 50 messages
9. **moderation** — 8 reports, 5 bans, 8 moderation logs
10. **extended** — 5 ads + analytics, 15 media, 10 stories, 20 notifications, 5 calls

Steps share data through `internal.SeedState` (`cmd/seed/internal/state.go`). Each module opens a fresh DB connection.

Table list for manual DROP (if skipping reset):
```
ad_analytics, moderation_logs, bans, reports, notifications, calls, messages,
chat_participants, chats, group_members, communities, tags, post_reactions,
bookmarks, blocks, friends, follows, comments, posts, media, stories, ads,
user_roles, profiles, violation_rules, emojis, roles, users
```

## Key packages

| Package | Role |
|---|---|
| `config/` | Env loading (custom parser), Cloudinary creds |
| `controllers/` | HTTP handlers (Gin context) |
| `services/` | Business logic |
| `repository/` | GORM data access |
| `dto/` | Request/response structs |
| `validations/` | Explicit validation (not struct tags) |
| `utils/` | JWT (golang-jwt/jwt/v5), bcrypt hashing, username generation |
| `db/` | MySQL connection via `database/sql` + `go-sql-driver/mysql` |
| `models/` | GORM model structs (28 files) |
| `cmd/seed/` | Seed scripts (raw SQL, 10 modules + `internal/` helpers) |
| `routes/` | Route registration |

## Validation & errors

- `validations.AuthValidation` exports sentinel errors (`ErrUserNotFound`, `ErrPasswordTooShort`, etc.) for `errors.Is()` checks.
- `repository.ErrUserNotFound` — checked by `auth.service.go` with `errors.Is()`.

## Key files

| File | Purpose |
|---|---|
| `cmd/main.go` | Server entrypoint: env load → DB connect → GORM wrap → wire auth → start Gin |
| `cmd/seed/main.go` | Seed orchestrator: runs 10 modules sequentially |
| `routes/auth.routes.go` | `/api/auth/register`, `/api/auth/login` |
| `config/env.go` | Custom `.env` loader with singleton guard |
| `db/mysql.go` | `ConnectDb()` returns `*sql.DB` (no TLS in DSN) |
