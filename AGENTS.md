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
- **DB**: `db.ConnectDb(env)` returns `*sql.DB`; `main.go` wraps with `gorm.Open(mysql.New(mysql.Config{Conn: database}), ...)`. DSN has **no TLS params**. `registerTLSConfig` in `db/mysql.go` is a dead stub.
- **All model IDs are `string` (UUID)**. Foreign keys (`UserID`, `PostID`, etc.) are `string`/`*string`. Models have `json` and `db` tags; `db` tags are unused relics.
- **29 model files** (28 seed tables + `password_reset_token.model.go` — not in seed schema).

## Config

- `.env` loaded by `config.LoadEnv()` (custom line parser — **not** godotenv). Singleton guard prevents reloads.
- Required env vars: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`, `JWT_EXPIRES_IN`.
  - `DB_SSL` parsed via `strconv.ParseBool` — will fail at parse time if unset, but **not** in `validateRequired`.
  - `PORT` defaults to `"8080"`.
- Optional Gmail: `GMAIL_USER`, `GMAIL_PASSWORD`, `FRONTEND_RESET_URL` (default `http://localhost:3000`).
- Cloudinary: `CLOUDINARY_URL` primary; fallback `CLOUDINARY_CLOUD_NAME`, `CLOUDINARY_API_KEY`, `CLOUDINARY_API_SECRET`. Loaded via `config.LoadCloudinaryEnv()` (never called from `main.go`).
- `config.GetEnv()` returns a **value copy**.
- `config/cloundinary.go` is misnamed (should be `cloudinary`).

## Routes

All routes wired in `cmd/main.go` (inside `if database != nil { ... }` guard):

| Path | Method | Auth | Handler file |
|---|---|---|---|
| `/health` | GET | No | inline in `main.go` |
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
| `/api/follow/:userID` | POST | `middlewares.AuthMiddleware` | `routes/follow.routes.go` |
| `/api/follow/stats/:userID` | GET | No | `routes/follow.routes.go` |
| `/api/media/upload` | POST | `middlewares.AuthMiddleware` | `routes/media.routes.go` |
| `/api/media/storage` | GET | `middlewares.AuthMiddleware` | `routes/media.routes.go` |
| `/api/reports` | POST | `middlewares.AuthMiddleware` | `routes/report.routes.go` |
| `/api/blocks` | POST | `middlewares.AuthMiddleware` | `routes/block.routes.go` |
| `/api/blocks` | GET | `middlewares.AuthMiddleware` | `routes/block.routes.go` |
| `/api/search` | GET | No | `routes/search.routes.go` |

Two separate auth middleware implementations exist:
- `middlewares/auth.middleware.go` — uses `utils.ParseToken`, sets `userID` and `email`
- `routes/post.routes.go` inline (package-local `AuthMiddleware`) — uses `jwt.Parse` directly, sets `userId`

## DTOs & validation

- `validations` package uses explicit validation (sentinel errors + struct methods), **not** Gin binding tags.
- DTOs with `binding` tags are the exception: `profile.dto.go:EditProfileInput`, `post.controller.go:CreatePostInput`/`ReactPostInput`.
- **Query params**: `search.dto.go:SearchInput` uses `form:"keyword" form:"type"` for `c.ShouldBindQuery(&input)`.
- Password rules: ≥8 chars, ≤128, must have upper, lower, digit, special.

## Business logic constraints

**Role protection** — `repository/auth.repository.go:HasRole(ctx, userID, roleName)` checks user_roles via JOIN. Used to protect admin/superadmin users:
- `ReportService` rejects reporting users with `SUPER_ADMIN` or `ADMIN` role (`"cannot report admin or super admin"`).
- `BlockService` rejects blocking users with `SUPER_ADMIN` or `ADMIN` role (`"cannot block admin or super admin"`).
- `SearchRepository.SearchUsers` excludes users with `SUPER_ADMIN` or `ADMIN` role via `NOT EXISTS` subquery.

**Post status** — `ReportService.CreateReport` validates post exists and is active via `postRepo.FindByID` (only returns active posts). Hidden/deleted posts cannot be reported.

**Toggle pattern** — `BlockService.ToggleBlock` and `FollowService.FollowToggle` use the same pattern: check existing record, if found → delete (unblock/unfollow), if not → create (block/follow).

**Password reset token exposed** — `password_reset.service.go:67` includes the raw token in the API response under `Token` field (flagged `// ⚠️ CHỈ FOR TESTING`).

## JWT

- `utils.GenerateTokenPair` — HS256, access TTL from `JWTExpiresIn` (minutes, fallback 15 min), refresh TTL hardcoded to 7 days.
- `utils.ParseToken` parses into `*utils.TokenClaims` (`UserID`, `Email`, `TokenType`).

## Password reset flow

1. `POST /api/auth/forgot-password` — 32-byte hex token, stored in `password_reset_tokens`, email via Gmail SMTP (`smtp.gmail.com:587`).
2. `POST /api/auth/verify-reset-token` — checks validity & expiry.
3. `POST /api/auth/reset-password` — updates password, marks token used.

Token expiry: 10 min. Email template in Vietnamese.

## Seed system

```bash
go build ./cmd/seed && ./seed.exe
```

10 ordered steps, raw SQL (not GORM), **drops all tables on run**:
1. **reset** — DROP 28 tables (`FOREIGN_KEY_CHECKS=0`)
2. **schema** — CREATE 28 tables with FKs, indexes, utf8mb4
3. **users** — 20 users (bcrypt `Password123!`), 2 banned, 1 suspended
4. **core** — 3 roles, 10 emojis, 8 violation rules + `user_roles`
5. **profiles** — 20 profiles, 5 private
6. **social** — 30 posts, 60 comments, 50 tags, 80 reactions, 40 follows, 15 friends, 5 blocks, 20 bookmarks
7. **relationships** — 5 communities, 25 group members
8. **messaging** — 8 chats, ~24 participants, 50 messages
9. **moderation** — 8 reports, 5 bans, 8 moderation logs
10. **extended** — 5 ads + analytics, 15 media, 10 stories, 20 notifications, 5 calls

Steps share data via `internal.SeedState`. UUIDs via `internal.UUID()` (crypto/rand, RFC 9562).

## Stubs / not wired

- `controllers/user.controller.go` — empty file
- `repository/user.repository.go` — has `Create`, `FindByEmail` but **not wired** in `cmd/main.go`
- `config.LoadCloudinaryEnv()` — never called from `main.go`
- `cmd/cloudinary-check/` — standalone binary

## Inconsistencies to know

- **Repository patterns differ**: most repos use concrete structs; `post.repository.go` uses an interface (`PostRepository`).
- **UUID generation diverges**: `utils.GenerateUUID()` (crypto/rand) vs `github.com/google/uuid.New().String()` in `post.service.go`.
- **ReactPost hardcodes emoji UUIDs**: `post.controller.go:113-124` maps 10 UUIDs to emoji names — fragile if seed data changes.
- **Two auth middlewares**: `middlewares/auth.middleware.go` (English errors, `userID` claim) vs `routes/post.routes.go:25` (Vietnamese errors, `userId` claim).

## Key packages

| Package | Files | Role |
|---|---|---|
| `config/` | 2 | Env loading, Cloudinary creds |
| `controllers/` | 10 | HTTP handlers |
| `services/` | 9 | Business logic |
| `repository/` | 10 | GORM data access |
| `dto/` | 8 | Request/response structs |
| `validations/` | 5 | Explicit validation |
| `utils/` | 5 | JWT, bcrypt, UUID, username gen, email |
| `models/` | 29 | GORM model structs |
| `routes/` | 9 | Route registration |
| `middlewares/` | 1 | Gin JWT auth middleware |
| `cmd/seed/` | ~12 | Seed scripts + internal helpers |
| `db/` | 1 | MySQL connection |
