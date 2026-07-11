# LinkUp Server

Backend API cho mạng xã hội LinkUp — hỗ trợ bài viết, story 24h, tin nhắn real-time (WebSocket), voice/video call (WebRTC + ICE), nhóm cộng đồng, quảng cáo, hệ thống điểm đóng góp, và lịch sử cuộc gọi.

## Tech Stack

- **Ngôn ngữ:** Go 1.26.3
- **Framework:** Gin
- **Database:** MySQL + GORM
- **WebSocket:** gorilla/websocket (2 Hub types: `ws.Hub` + `groupws.Hub`)
- **Authentication:** JWT (HS256, access + refresh token)
- **Media:** Cloudinary
- **Encryption:** AES-256-GCM (tin nhắn chat)
- **Seed:** raw `database/sql` (10 bước có thứ tự, dùng chung `SeedState`)

## Tính năng chính

- **Xác thực & Người dùng:** Đăng ký, đăng nhập, refresh token, đặt lại mật khẩu qua email (Gmail SMTP)
- **Bài viết & Bình luận:** CRUD, react (emoji), chia sẻ, lưu bài, tags
- **Story 24h:** Đăng tải story, xem (track view), tương tác (react/reply/share), analytics
- **Profile:** Quản lý thông tin cá nhân, avatar, storage quota, last_read_missed_at
- **Follow & Bạn bè:** Theo dõi, kết bạn, chặn người dùng (toggle pattern)
- **Real-time Chat:** Tin nhắn trực tiếp & nhóm, mã hóa AES-256-GCM, typing indicator, tìm kiếm
- **Voice/Video Call:** WebRTC signaling qua WebSocket, ICE server config (STUN/TURN), quản lý cuộc gọi (initiate, accept, reject, end, mute, video toggle, History), lịch sử cuộc gọi có filter + phân trang
- **Cộng đồng:** Nhóm bài viết, quy tắc, lời mời, mã mời (invite code 6 ký tự), join request
- **Điểm đóng góp:** Policy (post/comment/reaction weight), challenge, leaderboard, badge, auto-promote moderator
- **Quảng cáo:** Quản lý quảng cáo cho đối tác (RBAC: PARTNER), analytics
- **Admin:** Kiểm duyệt nội dung, quản lý báo cáo, ban user, moderation log, quản lý media flagged/rejected
- **AI Guardian (Moderation):** Tự động kiểm duyệt ảnh/video khi upload qua Cloudinary + AWS Rekognition, phân luồng (approved/flagged/rejected/pending), ghi moderation log + notification
- **Tìm kiếm:** Bài viết, người dùng, hashtag
- **Thông báo:** Real-time notification qua WebSocket (like, comment, follow, friend request, call, moderation result)
- **Media:** Upload ảnh/video lên Cloudinary kèm AI moderation (single upload, không double upload)

## Bắt đầu

### Yêu cầu

- Go 1.26+
- MySQL 8.0+
- (Tùy chọn) Air — hot reload: `go install github.com/air-verse/air@latest`

### Cài đặt

```bash
# Clone repo
git clone <repo-url>
cd server

# Cài đặt dependencies
go mod download

# Tạo file .env (xem cấu hình bên dưới)
```

### Cấu hình môi trường

Tạo file `.env` tại thư mục gốc:

```env
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=linkup
DB_SSL=true

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRES_IN=15

# Cloudinary
CLOUDINARY_URL=cloudinary://api_key:api_secret@cloud_name

# Email (SMTP Gmail — tùy chọn, cho quên mật khẩu)
GMAIL_USER=your-email@gmail.com
GMAIL_PASSWORD=your-app-password

# ICE Servers (WebRTC — tùy chọn)
# STUN: comma-separated URLs
ICE_SERVER_URLS=stun:stun.l.google.com:19302,stun:stun1.l.google.com:19302
# TURN: để trống nếu chưa dùng
TURN_SERVER_URL=
TURN_USERNAME=
TURN_CREDENTIAL=

# WebSocket origin (tùy chọn, mặc định cho phép tất cả)
WS_ALLOWED_ORIGINS=http://localhost:3000

# Server
PORT=8080
FRONTEND_RESET_URL=http://localhost:3000
```

> **Lưu ý:** `DB_SSL` phải set thành `"true"` do lỗi parser config (giá trị `false` bị coi là thiếu). Thực tế DSN không dùng SSL.

### Seed database

Seed gồm 2 bước: (1) schema (tạo 46 bảng + index/FK idempotent) và (2) dữ liệu mẫu.

```bash
# Seed toàn bộ (xóa + tạo lại)
go build ./cmd/seed && ./seed.exe
```

Lệnh này xóa toàn bộ dữ liệu cũ và tạo lại 46 bảng với dữ liệu mẫu (users, profiles, posts, communities, calls, ...). Tất cả user seed có mật khẩu `Password123!`. Các ALTER TABLE dùng helper idempotent (`addColumnIfMissing`, `addIndexIfMissing`, `addForeignKeyIfMissing`) — an toàn khi chạy lại.

## Chạy ứng dụng

### Development (hot reload)

```bash
air
```

### Production

```bash
go build -o ./tmp/main.exe ./cmd
./tmp/main.exe
```

## Cấu trúc thư mục

```
├── cmd/
│   ├── main.go                 # Entrypoint (Gin, GORM, routes, WS)
│   ├── seed/                   # Seed database (main.go + schema/ sub-package)
│   └── cloudinary-check/       # Standalone (không dùng trong app)
├── config/                     # Env parser (custom, singleton guard)
├── controllers/                # HTTP handlers (Gin) — 22 files
├── dto/                        # Data Transfer Objects (binding tags + validators)
├── db/                         # Kết nối MySQL (*sql.DB)
├── docs/                       # Tài liệu + test cases (.test-case.md)
├── groupws/                    # Group chat WebSocket Hub (riêng, không phải ws.Hub)
├── middlewares/                # Auth (JWT) & RBAC middleware
├── models/                     # GORM models — 49 files
├── repository/                 # Database access layer (GORM + raw SQL)
├── routes/                     # Route registration — 22 files
├── services/                   # Business logic (PostService, etc. là interfaces)
├── tests/                      # Validation tests (community, contribution, call)
├── utils/                      # JWT, encryption (AES-256-GCM), hash, UUID, email
├── validations/                # Input validation — 13 validators (struct methods)
└── ws/                         # WebSocket hub, client, chat/call service interfaces
```

## API Endpoints

### Public (không cần auth)

| Method | Path | Mô tả |
|--------|------|-------|
| GET | `/health` | Health check |
| GET | `/ws` | WebSocket (notifications) `?token=` |
| GET | `/posts` | Danh sách bài viết |
| GET | `/posts/:id` | Chi tiết bài viết |
| GET | `/stories/feed` | Danh sách story đầu trang chủ |
| GET | `/api/tags/:name/posts` | Bài viết theo tag |
| GET | `/api/profile/:userID` | Profile công khai |
| GET | `/api/search` | Tìm kiếm (posts, users, hashtags) |
| GET | `/api/communities/:id/contributions/leaderboard` | Bảng xếp hạng điểm đóng góp |
| GET | `/api/communities/:id/contributions/:userID` | Điểm của user trong community |
| GET | `/api/calls/ice-servers` | ICE server config (STUN/TURN) |

### Yêu cầu xác thực

Tất cả endpoint dưới đây yêu cầu `Authorization: Bearer <access_token>`.

#### Auth

| Method | Path | Mô tả |
|--------|------|-------|
| POST | `/api/auth/register` | Đăng ký |
| POST | `/api/auth/login` | Đăng nhập |
| POST | `/api/auth/refresh` | Refresh token |
| POST | `/api/auth/change-password` | Đổi mật khẩu |
| POST | `/api/auth/forgot-password` | Quên mật khẩu (gửi email) |
| POST | `/api/auth/verify-reset-token` | Xác thực token reset |
| POST | `/api/auth/reset-password` | Đặt lại mật khẩu |

#### Posts

| Method | Path | Mô tả |
|--------|------|-------|
| POST | `/posts` | Tạo bài viết |
| POST | `/posts/:id/react` | React bài viết (toggle) |
| POST | `/posts/:id/comments` | Bình luận |
| POST | `/posts/:id/share` | Chia sẻ |
| POST | `/posts/:id/save` | Lưu bài (bookmark) |

#### Stories

| Method | Path | Mô tả |
|--------|------|-------|
| POST | `/stories` | Đăng tải story mới |
| GET | `/stories/:id` | Xem story (ghi nhận view) |
| POST | `/stories/:id/interact` | Tương tác (react, reply, share) |
| GET | `/stories/:id/analytics` | Analytics cho chủ story |

#### Profile & Social

| Method | Path | Mô tả |
|--------|------|-------|
| PATCH | `/api/profile` | Cập nhật profile |
| GET/POST | `/api/follow/*` | Follow/Unfollow (toggle) |
| GET/POST | `/api/friend-requests/*` | Kết bạn (toggle + accept/reject) |
| GET/POST | `/api/blocks` | Chặn/Bỏ chặn (toggle) |

#### Media & Reports

| Method | Path | Mô tả |
|--------|------|-------|
| GET/POST | `/api/media/*` | Upload media lên Cloudinary |
| POST | `/api/reports` | Báo cáo vi phạm |

#### Notifications

| Method | Path | Mô tả |
|--------|------|-------|
| GET/POST | `/api/notifications*` | Thông báo (mark read, list) |

#### Chat

| Method | Path | Mô tả |
|--------|------|-------|
| GET/POST | `/api/chats/*` | Chat CRUD + tin nhắn |
| GET | `/api/chats/ws` | WebSocket Chat Hub (Bearer auth) |
| GET/POST | `/api/group-chats/*` | Group chat settings, members, mutes |
| GET | `/api/group-chats/ws` | WebSocket Group Chat Hub (`?token=` access JWT) |

#### Communities

| Method | Path | Mô tả |
|--------|------|-------|
| GET/POST | `/api/communities*` | Quản lý cộng đồng, rules, invitations, invite codes |
| GET/POST/PUT/DELETE | `/api/communities/:id/policy\|challenges\|contributions` | Điểm đóng góp |

#### Ads (RBAC: PARTNER)

| Method | Path | Mô tả |
|--------|------|-------|
| GET/POST/PUT/DELETE | `/ads-management*` | Quản lý quảng cáo |
| GET/POST | `/customer/*` | Quảng cáo (partner) |

#### Admin

| Method | Path | Mô tả |
|--------|------|-------|
| GET/POST | `/api/admin/*` | Admin dashboard, reports, bans, moderation |
| GET | `/api/admin/media/flagged` | Danh sách media flagged/rejected (phân trang, filter status) |
| POST | `/api/admin/media/:id/review` | Duyệt media (approved / rejected) — kèm cleanup Cloudinary + notification |
| POST | `/api/admin/media/cleanup-rejected` | Dọn dẹp media rejected quá 7 ngày |

#### Calls

| Method | Path | Mô tả |
|--------|------|-------|
| GET | `/api/calls/history` | Lịch sử cuộc gọi (phân trang, filter theo `status`/`call_type`/`direction`) |
| GET | `/api/calls/history/missed-count` | Số cuộc gọi nhỡ chưa đọc |
| POST | `/api/calls/history/mark-read` | Đánh dấu đã đọc cuộc gọi nhỡ |
| POST | `/api/calls/history/hide` | Ẩn cuộc gọi khỏi lịch sử (soft-delete theo user) |
| GET | `/api/calls/:callID` | Chi tiết cuộc gọi |
| POST | `/api/calls/:callID/mute` | Bật/tắt mic |
| POST | `/api/calls/:callID/video` | Bật/tắt camera |
| GET | `/api/calls/ws` | WebSocket Call Hub (signaling + control) |

## WebSocket

Kiến trúc WS: **2 Hub types (`ws.Hub`, `groupws.Hub`), 3 Hub instances, 4 endpoints**.

| Endpoint | Hub type | Instance | Service | Auth | Mục đích |
|---|---|---|---|---|---|
| `GET /ws` | `ws.Hub` | `hub` | `service=nil, callService=nil` | `?token=` access JWT | Thông báo real-time |
| `GET /api/chats/ws` | `ws.Hub` | `chatHub` | `ChatService` set | Bearer | Chat mã hóa |
| `GET /api/calls/ws` | `ws.Hub` | `hub` (shared) | `callService` set | Bearer | WebRTC signaling |
| `GET /api/group-chats/ws` | `groupws.Hub` | `groupHub` | `GroupMessageService` + `GroupChatService` | `?token=` access JWT | Group chat real-time |

### Notification Hub (`GET /ws`)

- Chỉ nhận, không xử lý event từ client (chat/call events bị discard).

### Chat Hub (`GET /api/chats/ws`)

- **Client → Server:**
  - `chat:join` — tham gia phòng chat, nhận lịch sử
  - `message:send` — gửi tin nhắn (mã hóa AES-256-GCM)
  - `message:delete` — xóa tin nhắn (self hoặc all)
  - `message:search` — tìm kiếm tin nhắn
  - `typing:start` / `typing:stop` — đang gõ

- **Server → Client:**
  - `message:new` — tin nhắn mới (broadcast trong phòng)
  - `message:deleted` — xác nhận xóa
  - `message:search_result` — kết quả tìm kiếm
  - `typing` — trạng thái đang gõ của user khác
  - `message:history` — lịch sử khi join

### Group Chat Hub (`GET /api/group-chats/ws`)

Sử dụng Hub type riêng (`groupws.Hub`), không phải `ws.Hub`.

- **Client → Server:**
  - `group:join` — tham gia nhóm chat, nhận lịch sử tin nhắn
  - `group:message:send` — gửi tin nhắn
  - `group:typing:start` / `group:typing:stop` — đang gõ
  - `group:message:search` — tìm kiếm tin nhắn
  - `group:leave` — rời nhóm (public/quiet, kèm history mode)
  - `group:member:add` — thêm thành viên
  - `group:member:ban` — chặn thành viên
  - `group:member:mute` — mute thành viên (có thời hạn / vĩnh viễn)
  - `group:member:unmute` — bỏ mute
  - `group:admin:transfer` — chuyển quyền admin
  - `group:settings:update` — cập nhật settings nhóm

- **Server → Client:**
  - `group:history` — lịch sử tin nhắn khi join
  - `group:message:new` — tin nhắn mới
  - `group:typing` — trạng thái typing
  - `group:message:search_result` — kết quả tìm kiếm
  - `group:member:left` — thành viên rời nhóm (public leave)
  - `group:member:added` — thành viên mới
  - `group:member:banned` — thành viên bị ban
  - `group:member:muted` — member bị mute
  - `group:member:unmuted` — member được unmute
  - `group:admin:transferred` — chuyển admin
  - `group:settings:updated` — settings đã cập nhật

### Call Hub (`GET /api/calls/ws`)

- **Client → Server:**
  - `call:initiate` — tạo cuộc gọi
  - `call:accept` — chấp nhận
  - `call:reject` — từ chối
  - `call:end` — kết thúc
  - `call:signal` — WebRTC signaling (SDP, ICE candidate)
  - `call:video_toggle` — bật/tắt camera (`{call_id, video_enabled}`)
  - `call:toggle_mute` — bật/tắt mic (`{call_id, muted}`)

- **Server → Client:**
  - `call:incoming` — có cuộc gọi đến (`{call_id, caller_id, call_type, timestamp}`)
  - `call:status` — cập nhật trạng thái (`calling`, `connected`, `ended`, `missed`, `rejected`, `cancelled`)
  - `call:busy` — callee đang bận
  - `call:mute` — trạng thái mute (`{call_id, user_id, muted}`)
  - `call:video` — trạng thái camera (`{call_id, user_id, video_enabled}`)
  - `call:signal` — chuyển tiếp WebRTC signal
  - `call:missed` — thông báo cuộc gọi nhỡ (khi caller kết thúc trước khi callee trả lời)
  - `call:cancelled` — thông báo caller đã hủy cuộc gọi

## Testing

```bash
# Validation tests (không cần DB) — community
go test ./tests/community/... -v

# Validation tests — contribution
go test ./tests/contribution/... -v

# Validation tests — call (DTO, model, history)
go test ./tests/call/... -v

# Service tests (một số cần TEST_DSN env var)
go test ./services/... -v

# Verify toàn bộ codebase
go build ./... && go vet ./...
```

```bash
# Integration tests — AI moderation (cần TEST_DSN)
go test ./services/... -run TestAdminService -v
```

Test cases chi tiết cho từng module nằm trong `docs/test-case/` (xem `docs/test-case/ai-guardian.test-case.md` cho AI moderation).

## Kiến trúc

```
Client → Router (Gin) → Middleware (Auth/RBAC) → Controller → Service → Repository (GORM) → MySQL
                                                     ↕
                                          WebSocket Hub (ws.Hub / groupws.Hub)
```

### Layers

- **Controller:** Xử lý HTTP request/response, parse input (`ShouldBindJSON`, `ShouldBindQuery`)
- **Service:** Business logic (interface-based: `PostService`, `MediaService`, `AdService`, `AIModerationService`, `CallService`; concrete struct cho phần còn lại, `AdminService` có setter tiện ích)
- **Repository:** GORM queries + raw SQL khi cần (innodb lock, batch operations)
- **WebSocket Hub:** `ws.Hub` cho notification/chat/call, `groupws.Hub` riêng cho group chat real-time

### Design patterns

- **Toggle pattern:** `ToggleBlock`, `FollowToggle`, `ToggleFriendRequest`, `ReactPost`, `ToggleMute`, `ToggleVideo` — check tồn tại → xóa hoặc tạo
- **Atomic call status:** `AcceptCallAtomic` / `RejectCallAtomic` dùng conditional `UPDATE ... WHERE status IN (?, ?)` để tránh TOCTOU
- **Concurrency guard:** `CreateIfNotBusy` dùng `SELECT COUNT(*) ... FOR UPDATE` (gap lock) để ngăn duplicate call
- **Soft-delete call history:** Bảng `call_hidden` (composite PK `call_id, user_id`), không xóa row khỏi `calls`
- **Batch profile load:** 2-query pattern (SELECT * FROM profiles WHERE user_id IN ?) thay vì JOIN
- **Single-upload moderation:** Upload lên Cloudinary với `Moderation: "aws_rek"`, Cloudinary tự gọi AWS Rekognition → trả về SecureURL + PublicID + moderation result trong 1 response (không upload 2 lần)
- **Transition whitelist:** `ReviewMedia` dùng map `mediaReviewTransitions` — chỉ cho phép `flagged → approved | rejected`
- **Cleanup on reject:** Admin reject → destroy Cloudinary file + clear `FileURI` trong DB + notification phân biệt AI vs Admin
- **Idempotent schema:** `addColumnIfMissing` / `addIndexIfMissing` / `addForeignKeyIfMissing` kiểm tra `information_schema` trước khi ALTER
- **SQL injection prevention:** DDL identifiers được whitelist qua regex `^[a-zA-Z_][a-zA-Z0-9_]*$`

### Conventions

- Tất cả ID là UUID string (crypto/rand, RFC 9562)
- Service trả về lỗi tiếng Việt; middleware RBAC trả về tiếng Anh
- Chat messages dùng AES-256-GCM (`utils/encryption.go`), key lưu per-chat (`encryption_key`)
- Validation: DTOs dùng `binding` tags; các chỗ khác dùng `validations` package (13 validators, sentinel errors)

## Docker

```bash
docker build -t linkup-server .
docker run -p 8080:8080 --env-file .env linkup-server
```

## Giấy phép

© 2026 LinkUp
