# LinkUp Server

Backend API cho mạng xã hội LinkUp — hỗ trợ bài viết, tin nhắn real-time (WebSocket), voice/video call, nhóm cộng đồng, quảng cáo, và hệ thống điểm đóng góp.

## Tech Stack

- **Ngôn ngữ:** Go 1.26.3
- **Framework:** Gin
- **Database:** MySQL + GORM
- **WebSocket:** gorilla/websocket
- **Authentication:** JWT (HS256)
- **Media:** Cloudinary
- **Encryption:** AES-256-GCM (tin nhắn chat)

## Tính năng chính

- **Xác thực & Người dùng:** Đăng ký, đăng nhập, refresh token, đặt lại mật khẩu qua email
- **Bài viết & Bình luận:** CRUD, react (emoji), chia sẻ, lưu bài
- **Profile:** Quản lý thông tin cá nhân, avatar, storage quota
- **Follow & Bạn bè:** Theo dõi, kết bạn, chặn người dùng
- **Real-time Chat:** Tin nhắn trực tiếp & nhóm, mã hóa đầu cuối, typing indicator
- **Voice/Video Call:** Signaling qua WebSocket, quản lý cuộc gọi
- **Cộng đồng:** Nhóm bài viết, quy tắc, lời mời, mã mời
- **Điểm đóng góp:** Policy, challenge, leaderboard
- **Quảng cáo:** Quản lý quảng cáo cho đối tác (RBAC)
- **Admin:** Kiểm duyệt nội dung, quản lý báo cáo, ban user
- **Tìm kiếm:** Bài viết, người dùng, hashtag
- **Thông báo:** Real-time notification qua WebSocket
- **Media:** Upload ảnh/video lên Cloudinary

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

# Email (SMTP Gmail - tùy chọn)
GMAIL_USER=your-email@gmail.com
GMAIL_PASSWORD=your-app-password

# Server
PORT=8080
FRONTEND_RESET_URL=http://localhost:3000
```

> **Lưu ý:** `DB_SSL` phải được set thành `true` do lỗi parser của config (giá trị `false` bị coi là thiếu). Thực tế DSN không dùng SSL.

### Seed database

```bash
go build ./cmd/seed && ./seed.exe
```

Lệnh này xóa toàn bộ dữ liệu cũ và tạo lại 34+ bảng với dữ liệu mẫu. Tất cả user seed có mật khẩu `Password123!`.

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
│   ├── main.go                 # Entrypoint
│   ├── seed/                   # Seed database
│   └── cloudinary-check/       # Standalone (không dùng trong app)
├── config/                     # Env & cấu hình
├── controllers/                # HTTP handlers (Gin)
├── dto/                        # Data Transfer Objects
├── db/                         # Kết nối MySQL
├── docs/                       # Tài liệu
├── middlewares/                # Auth & RBAC middleware
├── models/                     # GORM models (45 models)
├── repository/                 # Database access layer
├── routes/                     # Route registration (20 files)
├── services/                   # Business logic
├── tests/                      # Integration tests
├── utils/                      # JWT, encryption, hash, UUID...
├── validations/                # Input validation (13 validators)
└── ws/                         # WebSocket hub & client
```

## API Endpoints

### Public

| Method | Path | Mô tả |
|--------|------|-------|
| GET | `/health` | Health check |
| GET | `/ws` | WebSocket (notifications) `?token=` |
| GET | `/posts` | Danh sách bài viết |
| GET | `/posts/:id` | Chi tiết bài viết |
| GET | `/api/tags/:name/posts` | Bài viết theo tag |
| GET | `/api/profile/:userID` | Profile công khai |
| GET | `/api/search` | Tìm kiếm |
| GET | `/api/communities/:id/contributions/leaderboard` | Bảng xếp hạng |
| GET | `/api/communities/:id/contributions/:userID` | Điểm của user |

### Yêu cầu xác thực

Tất cả các endpoint dưới đây yêu cầu `Authorization: Bearer <token>`.

| Method | Path | Mô tả |
|--------|------|-------|
| POST | `/api/auth/register` | Đăng ký |
| POST | `/api/auth/login` | Đăng nhập |
| POST | `/api/auth/refresh` | Refresh token |
| POST | `/api/auth/change-password` | Đổi mật khẩu |
| POST | `/api/auth/forgot-password` | Quên mật khẩu |
| POST | `/api/auth/reset-password` | Đặt lại mật khẩu |
| POST | `/api/auth/verify-reset-token` | Xác thực token reset |
| POST | `/posts` | Tạo bài viết |
| POST | `/posts/:id/react` | React bài viết |
| POST | `/posts/:id/comments` | Bình luận |
| POST | `/posts/:id/share` | Chia sẻ |
| POST | `/posts/:id/save` | Lưu bài |
| PATCH | `/api/profile` | Cập nhật profile |
| GET/POST | `/api/follow/*` | Follow/Unfollow |
| GET/POST | `/api/friend-requests/*` | Kết bạn |
| GET/POST | `/api/blocks` | Chặn/Bỏ chặn |
| GET/POST | `/api/media/*` | Upload media |
| POST | `/api/reports` | Báo cáo vi phạm |
| GET/POST | `/api/notifications*` | Thông báo |
| GET/POST | `/api/chats/*` | Chat (message, typing, delete, search) |
| GET/POST | `/api/group-chats/*` | Group chat |
| GET/POST | `/api/communities*` | Quản lý cộng đồng |
| GET/POST/PUT/DELETE | `/api/communities/:id/policy\|challenges\|contributions` | Điểm đóng góp |
| GET/POST/PUT/DELETE | `/ads-management*` | Quản lý quảng cáo (RBAC: PARTNER) |
| GET/POST | `/customer/*` | Quảng cáo (RBAC: PARTNER) |
| GET/POST | `/api/admin/*` | Admin |
| GET/POST | `/api/calls/*` | Voice/Video call |

## WebSocket

### Notification Hub (`GET /ws`)

- Auth: `?token=<access_token>`
- Dùng cho thông báo real-time (like, comment, follow, friend request, call signaling)
- Không xử lý message từ client (ghi nhận nhưng bỏ qua)

### Chat Hub (`GET /api/chats/ws`)

- Auth: `Authorization: Bearer <token>`
- Xử lý các event:
  - `chat:join` — tham gia phòng chat
  - `message:send` — gửi tin nhắn (mã hóa AES-256-GCM)
  - `message:delete` — xóa tin nhắn
  - `message:search` — tìm kiếm tin nhắn
  - `typing:start` / `typing:stop` — đang gõ

## Testing

```bash
# Validation tests (không cần DB)
go test ./tests/community/... -v
go test ./tests/contribution/... -v

# Service tests (một số cần TEST_DSN)
go test ./services/... -v

# Verify code
go build ./... && go vet ./...
```

## Docker

```bash
docker build -t linkup-server .
docker run -p 8080:8080 --env-file .env linkup-server
```

## Kiến trúc

```
Client → Router (Gin) → Middleware (Auth/RBAC) → Controller → Service → Repository (GORM) → MySQL
                                                    ↕
                                               WebSocket Hub (gorilla/websocket)
```

- **Controller:** Xử lý HTTP request/response, parse input
- **Service:** Business logic, gọi repository
- **Repository:** GORM queries
- **WebSocket Hub:** Quản lý kết nối real-time, broadcast message

### Ghi chú kiến trúc

- `PostService`, `MediaService`, `AdService` là interfaces; các service còn lại dùng concrete structs
- Tất cả ID là UUID dạng string
- Service trả về lỗi tiếng Việt; middleware RBAC trả về tiếng Anh
- Pattern toggle: BlockService.ToggleBlock, FollowService.FollowToggle, FriendService.ToggleFriendRequest, postService.ReactPost (check tồn tại → xóa hoặc tạo)

## Giấy phép

© 2026 LinkUp
