# LinkUp Server — AGENTS.md

## Quick start

```bash
go build -o ./tmp/main.exe ./cmd   # or `air` for hot reload
go build ./cmd/seed && ./seed.exe  # full seed run (resets DB!)
go build ./...                      # verify all packages compile
```

No `_test.go` files exist anywhere.

## Architecture

```
cmd/main.go → controller → service → repository (GORM)
cmd/seed/   → raw database/sql (10 ordered steps, resets DB)
cmd/cloudinary-check/ → standalone Cloudinary connectivity test
```

- **Framework**: Gin (`gin.New()`; `.Use(gin.Logger(), gin.Recovery())` in `cmd/main.go:36-37`)
- **DB**: `db.ConnectDb(env)` returns `*sql.DB`; `main.go` wraps with `gorm.Open(mysql.New(mysql.Config{Conn: database}), ...)` for GORM repos. DSN has **no TLS params** — `DB_SSL` is parsed in config but never used in DSN. `registerTLSConfig` in `db/mysql.go` is a dead stub.
- **All model IDs are `string` (UUID)**. Foreign keys (`UserID`, `PostID`, etc.) are `string`/`*string`. Models have `json` and `db` tags; `db` tags are unused relics.
- **29 model files** (28 seed tables + `password_reset_token.model.go` — not in seed schema).

## Config

- `.env` loaded by `config.LoadEnv()` (custom line parser — **not** godotenv). Singleton guard prevents reloads.
- Required env vars (fail if empty): `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, `JWT_EXPIRES_IN`
  - `DB_SSL` parsed via `strconv.ParseBool` — **not in `validateRequired`** but will fail at parse time if unset.
  - `PORT` defaults to `"8080"`.
- Optional Gmail: `GMAIL_USER`, `GMAIL_PASSWORD`, `FRONTEND_RESET_URL` (default `http://localhost:3000`)
- Cloudinary: `CLOUDINARY_URL` (`cloudinary://key:secret@cloudname`) primary; fallback `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`, `CLOUDINARY_API_SECRET`. Loaded via `config.LoadCloudinaryEnv()` (never called from `main.go`).
- `config.GetEnv()` returns a **value copy** of the singleton `Env` struct.
- `config/cloundinary.go` is misnamed (should be `cloudinary`).

## Routes

All routes wired in `cmd/main.go` (inside `if database != nil { ... }` guard):

| Path | Method | Auth | Handler file |
|---|---|---|---|
| `/health` | GET | No | inline in `main.go` |
| `/` | GET | No | inline in `main.go` |
| `/api/auth/register` | POST | No | `routes/auth.routes.go` |
| `/api/auth/login` | POST | No | `routes/auth.routes.go` |
| `/api/auth/change-password` | POST | `middlewares.AuthMiddleware` | `routes/auth.routes.go` |
| `/api/auth/forgot-password` | POST | No | `routes/password_reset.routes.go` |
| `/api/auth/verify-reset-token` | POST | No | `routes/password_reset.routes.go` |
| `/api/auth/reset-password` | POST | No | `routes/password_reset.routes.go` |
| `/posts` | GET | No | `routes/post.routes.go` |
| `/posts` | POST | inline in `post.routes.go` | `routes/post.routes.go` |
| `/posts/:id` | GET | No | `routes/post.routes.go` |
| `/posts/:id/react` | POST | inline in `post.routes.go` | `routes/post.routes.go` |
| `/api/profile` | GET | `middlewares.AuthMiddleware` | `routes/profile.routes.go` |
| `/api/profile` | PUT | `middlewares.AuthMiddleware` | `routes/profile.routes.go` |
| `/api/profile/:userID` | GET | No | `routes/profile.routes.go` |

Two separate auth middleware implementations exist:
- `middlewares/auth.middleware.go` — uses `utils.ParseToken`, sets `userID` and `email`
- `routes/post.routes.go` inline (package-local `AuthMiddleware`) — uses `jwt.Parse` directly, sets `userId`

## JWT

- `utils.GenerateTokenPair(secret, userID, email, accessTTL, refreshTTL)` — `golang-jwt/jwt/v5` HS256.
- Access TTL from `JWTExpiresIn` env var (minutes); falls back to 15 min.
- Refresh TTL hardcoded to `7 * 24 * time.Hour` (`auth.service.go:118`).
- `utils.ParseToken` parses into `*utils.TokenClaims` (fields: `UserID`, `Email`, `TokenType`).
- `utils.GenerateToken` for single-token cases (used by password reset).

## Password reset flow

1. `POST /api/auth/forgot-password` — generates 32-byte hex token, stores in `password_reset_tokens`, sends email via Gmail SMTP (`smtp.gmail.com:587`)
2. `POST /api/auth/verify-reset-token` — checks token validity & expiry
3. `POST /api/auth/reset-password` — updates password, marks token used

Token expiry: 10 minutes. Email template is in Vietnamese, uses `GMAIL_USER`/`GMAIL_PASSWORD` env vars. **Forgot-password response leaks the token** (via `Token` field, for testing).

## DTOs & validation

- `validations.AuthValidation` exports sentinel errors (`ErrUserNotFound`, `ErrPasswordTooShort`, etc.) for `errors.Is()` checks.
- `repository.ErrUserNotFound` (`auth.repository.go:13`) — checked by `auth.service.go` with `errors.Is()`.
- Password rules: ≥8 chars, ≤128, must have upper, lower, digit, special.
- Most DTOs have no `binding` tags; exceptions: `profile.dto.go:EditProfileInput`, `post.controller.go:CreatePostInput`, `post.controller.go:ReactPostInput`.

## Stubs / not wired

- `controllers/user.controller.go` — empty file (just `package controllers`)
- `repository/user.repository.go` — has `Create`, `FindByEmail` but **not wired** in `cmd/main.go`
- `config.LoadCloudinaryEnv()`, `config/cloundinary.go` — loaded but never called from `main.go`
- `cmd/cloudinary-check/` — standalone binary, not part of server

## Inconsistencies to know

- **Repository patterns differ**: `auth`/`profile`/`password_reset` repos use concrete structs; `post` uses an interface (`repository.PostRepository`).
- **UUID generation diverges**: `auth.service.go` uses `utils.GenerateUUID()` (crypto/rand); `post.service.go` uses `github.com/google/uuid.New().String()`.
- **ReactPost hardcodes emoji UUIDs**: `post.controller.go:113-124` maps 10 UUIDs to emoji names — fragile if seed data changes.
- **Two auth middlewares**: `middlewares/auth.middleware.go` vs `routes/post.routes.go:25` — different error messages (English vs Vietnamese), different claim key (`userID` vs `userId`).
- **Password reset token exposed**: `password_reset.service.go:67` includes the token in the API response for testing.

## Seed system

```bash
go build ./cmd/seed && ./seed.exe
```

10 ordered steps, raw SQL (not GORM), **drops all tables on run**:

1. **reset** — DROP TABLE 28 tables (`FOREIGN_KEY_CHECKS=0`)
2. **schema** — CREATE TABLE 28 tables with FKs, indexes, utf8mb4
3. **users** — 20 users (bcrypt `Password123!`), 2 banned, 1 suspended
4. **core** — 3 roles, 10 emojis, 8 violation rules + `user_roles`
5. **profiles** — 20 profiles with avatars and bios, 5 private
6. **social** — 30 posts, 60 comments (nested), 50 tags, 80 reactions, 40 follows, 15 friends, 5 blocks, 20 bookmarks
7. **relationships** — 5 communities, 25 group members
8. **messaging** — 8 chats, ~24 participants, 50 messages
9. **moderation** — 8 reports, 5 bans, 8 moderation logs
10. **extended** — 5 ads + analytics, 15 media, 10 stories, 20 notifications, 5 calls

Steps share data through `internal.SeedState` (`cmd/seed/internal/state.go`). UUIDs generated via `internal.UUID()` (crypto/rand, RFC 9562 variant).

Table list for manual DROP:
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
| `controllers/` | HTTP handlers (Gin context) — 5 files |
| `services/` | Business logic — 4 files (auth, password_reset, post, profile) |
| `repository/` | GORM data access — 5 files |
| `dto/` | Request/response structs — 3 files (auth, password_reset, profile) |
| `validations/` | Explicit validation (not struct tags) — `AuthValidation` |
| `utils/` | JWT, bcrypt hashing, username generation, UUID, Gmail SMTP |
| `db/` | MySQL connection via `database/sql` + `go-sql-driver/mysql` |
| `models/` | GORM model structs (29 files) |
| `cmd/seed/` | Seed scripts (raw SQL, 10 modules + `internal/` helpers) |
| `routes/` | Route registration — 4 files |
| `middlewares/` | `auth.middleware.go` (Gin JWT middleware) |
