# LinkUp Server — AGENTS.md

## Quick start

```bash
go build -o ./tmp/main.exe ./cmd   # or `air` for hot reload
go build ./cmd/seed && ./seed.exe  # full seed run (WIP — see Seed section)
go build ./...                      # verify all packages compile
```

No `_test.go` files exist anywhere.

## Architecture

```
cmd/main.go          → controller → service → repository (GORM)
cmd/seed/main.go     → raw database/sql (9 ordered steps)
cmd/cloudinary-check → standalone Cloudinary connectivity test
```

- **Framework**: Gin (`gin.New()` — no default middleware, `.Use(gin.Logger(), gin.Recovery())` at `cmd/main.go:36-37`)
- **DB**: `db.ConnectDb(env)` returns `*sql.DB`; `main.go` wraps with `gorm.Open(mysql.New(mysql.Config{Conn: database}), ...)` for GORM repos. **`DB_SSL=true` but `registerTLSConfig` is a stub that returns an error** — TLS config not wired into the DSN (`db/mysql.go`).
- **All model IDs are `string` (UUID)**. Foreign keys (`UserID`, `PostID`, etc.) are `string`/`*string`. Models have both `json` and `db` tags; `db` tags are unused relics.
- **DTOs have no `binding` tags** — validation is via `validations.AuthValidation` methods in the controller layer.
- **JWT**: `utils.GenerateTokenPair(secret, userID, email, accessTTL, refreshTTL)` — accepts `string` userID. Refresh TTL hardcoded to 7 days (`auth.service.go:106`). Uses `golang-jwt/jwt/v5` HS256.
- **`user.controller.go`** is an empty stub; **`user.repository.go`** has basic CRUD (Create, FindByEmail) but **not wired** in `main.go`. Only auth routes (`/api/auth/register`, `/api/auth/login`) are connected.
- `config/cloundinary.go` is misnamed (should be `cloudinary`).

## Config

- `.env` loaded by `config.LoadEnv()` (custom line parser — **not** godotenv). Singleton guard prevents reloads.
- Required env vars (fail at load time if empty): `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, `JWT_EXPIRES_IN`
  - `DB_SSL` is parsed with `strconv.ParseBool` — `"true"`/`"false"` string required; not in `validateRequired` list but will fail at parse time.
  - `PORT` defaults to `"8080"` if unset.
  - `DB_PORT` is parsed via `strconv.Atoi` — integer required.
  - `JWT_EXPIRES_IN` parsed via `strconv.Atoi` — integer minutes required.
- Cloudinary: `CLOUDINARY_URL` (format `cloudinary://key:secret@cloudname`) is primary; fallback to `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`, `CLOUDINARY_API_SECRET`. Loaded via separate `config.LoadCloudinaryEnv()`.
- `config.GetEnv()` returns a value copy of the singleton `Env` struct.

## Seed system

```bash
go build ./cmd/seed && ./seed.exe
```

10 ordered steps, each opens its own `*sql.DB` and uses raw SQL (not GORM):

1. **reset** — DROP TABLE all 28 tables (`FOREIGN_KEY_CHECKS=0`)
2. **schema** — CREATE TABLE all 28 tables with FKs, indexes, utf8mb4
3. **users** — 20 users (bcrypt `Password123!`), 2 banned, 1 suspended
3. **core** — 3 roles, 10 emojis, 8 violation rules + `user_roles` assignments
4. **profiles** — 20 profiles with avatars and bios, 5 private
5. **social** — 30 posts, 60 comments (nested), 50 tags, 80 reactions, 40 follows, 15 friends, 5 blocks, 20 bookmarks
6. **relationships** — 5 communities, 25 group members with roles and points
7. **messaging** — 8 chats (direct + group), ~24 participants, 50 messages
8. **moderation** — 8 reports, 5 bans, 8 moderation logs
9. **extended** — 5 ads + analytics, 15 media, 10 stories, 20 notifications, 5 calls

Steps share data through `internal.SeedState` passed via the orchestrator (`cmd/seed/main.go`). Each module opens a fresh DB connection.

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
| `cmd/seed/` | Seed scripts (raw SQL, 9 modules + `internal/` helpers) |
| `routes/` | Route registration |

## Notes

- `validations.AuthValidation` exported error vars (`ErrUserNotFound`, `ErrPasswordTooShort`, etc.) are meant for `errors.Is()` checks.
- `repository.ErrUserNotFound` — checked by `auth.service.go` with `errors.Is()`.
- No `middlewares/` directory exists yet (no auth middleware).
- `air` config at `.air.toml` — builds `./tmp/main.exe`, watches `.go` files.
- `go 1.26.3` in `go.mod`.
