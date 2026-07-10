# Hướng dẫn Test Video Call với Postman

## 1. Chuẩn bị Environment

### 1.1. Biến môi trường

Tạo một **Environment** mới trong Postman với các biến sau:

| Variable | Initial Value | Description |
|----------|--------------|-------------|
| `base_url` | `http://localhost:8080` | Server URL |
| `token_A` | *(empty)* | JWT token user A (caller) |
| `token_B` | *(empty)* | JWT token user B (callee) |
| `token_C` | *(empty)* | JWT token user C (3rd party) |
| `userA_id` | *(empty)* | UUID của user A |
| `userB_id` | *(empty)* | UUID của user B |
| `userC_id` | *(empty)* | UUID của user C |
| `video_call_id` | *(empty)* | ID của video call để test |
| `voice_call_id` | *(empty)* | ID của voice call (để test error — toggle video trên voice call) |
| `ice_servers` | *(empty)* | Danh sách ICE servers (GET /ice-servers) |

### 1.2. Seed database

```bash
cd server
go build ./cmd/seed && ./seed.exe
```

Tài khoản mặc định (xem `cmd/seed/main.go` — tất cả đều dùng `Password123!`).

---

## 2. Lấy Token

### 2.1. Login user A (caller)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/auth/login` |
| **Headers** | `Content-Type: application/json` |
| **Body** (raw JSON) | `{"email": "seed_user_A@example.com", "password": "Password123!"}` |

**Script — Tests** tab:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_A', json.data.access_token);
pm.collectionVariables.set('userA_id', json.data.user.id);
```

### 2.2. Login user B (callee)

| Body | `{"email": "seed_user_B@example.com", "password": "Password123!"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_B', json.data.access_token);
pm.collectionVariables.set('userB_id', json.data.user.id);
```

### 2.3. Login user C (3rd party)

| Body | `{"email": "seed_user_C@example.com", "password": "Password123!"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_C', json.data.access_token);
pm.collectionVariables.set('userC_id', json.data.user.id);
```

---

## 3. Get ICE servers
chức năng của ice servers là gì?
trả lời: ICE servers (STUN/TURN) giúp thiết lập kết nối peer-to-peer giữa caller và callee trong WebRTC. STUN giúp tìm địa chỉ IP công cộng, TURN giúp relay dữ liệu khi peer-to-peer không khả dụng (ví dụ NAT/firewall).
Trước khi test video call, lấy danh sách ICE server (cần auth):

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/calls/ice-servers` |
| **Headers** | `Authorization: Bearer {{token_A}}` |

**Script — Tests** tab:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('ice_servers', JSON.stringify(json.ice_servers));
```

**Response mẫu**:
```json
{
    "ice_servers": [
        {"urls": "stun:stun.l.google.com:19302"}
    ]
}
```

---

## 4. Kết bạn (A → B)

Video call yêu cầu A và B phải là bạn bè.

### 4.1. Gửi lời mời (A → B)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/friend-requests/send` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"user_id": "{{userB_id}}"}` |

**Script** để lưu request_id:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('request_id', json.data.id);
```

### 4.2. Chấp nhận (B)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/friend-requests/{{request_id}}/accept` |
| **Headers** | `Authorization: Bearer {{token_B}}` |

### 4.3. Kết bạn C với B (cho busy test)

Lặp lại 4.1 và 4.2 với A→C→B: C gửi lời mời B, B chấp nhận.

---

## 5. Initiate Video Call

### 5.1. Thành công — A gọi B (video)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/initiate` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"callee_id": "{{userB_id}}", "call_type": "video"}` |

**Expected Response** (`200 OK`):
```json
{
    "data": {
        "id": "uuid-call-id-here",
        "caller_id": "{{userA_id}}",
        "callee_id": "{{userB_id}}",
        "call_type": "video",
        "is_group": false,
        "status": "calling",
        "video_enabled_caller": false,
        "video_enabled_callee": false,
        "started_at": null,
        "ended_at": null,
        "duration": 0,
        "muted_caller": false,
        "muted_callee": false,
        "created_at": "2026-07-..."
    }
}
```

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('video_call_id', json.data.id);
```

### 5.2. Lỗi — gọi chính mình

| Body | `{"callee_id": "{{userA_id}}", "call_type": "video"}` |
    
**Expected** (`400`):
```json
{"error": "không thể gọi cho chính mình"}
```

### 5.3. Lỗi — không phải bạn bè

Dùng token C (chưa kết bạn với A) gọi A:

| Headers | `Authorization: Bearer {{token_C}}` |
| Body | `{"callee_id": "{{userA_id}}", "call_type": "video"}` |

**Expected** (`400`):
```json
{"error": "chỉ có thể gọi cho bạn bè"}
```

### 5.4. Busy — callee đang bận

1. A→B video call đang `calling` (từ 5.1)
2. C gọi B:

| Headers | `Authorization: Bearer {{token_C}}` |
| Body | `{"callee_id": "{{userB_id}}", "call_type": "video"}` |

**Expected** (`200 OK`):
```json
{"message": "người dùng đang bận"}
```

> Response là `200` vì request hợp lệ, chỉ callee busy.

### 5.5. Thiếu callee_id

| Body | `{"call_type": "video"}` |

**Expected** (`400`):
```json
{"error": "callee_id và call_type là bắt buộc"}
```

---

## 6. Accept Video Call

> **Prerequisite**: A→B video call đang `calling` (chạy 5.1).

### 6.1. B chấp nhận

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/{{video_call_id}}/accept` |
| **Headers** | `Authorization: Bearer {{token_B}}` |

**Expected** (`200 OK`):
```json
{"message": "cuộc gọi đã được chấp nhận"}
```

Kiểm tra DB: `calls` table → `status = "connected"`, `started_at` không null.

### 6.2. Lỗi — A không thể accept

| Headers | `Authorization: Bearer {{token_A}}` |

**Expected** (`400`):
```json
{"error": "chỉ người nhận mới có thể chấp nhận cuộc gọi"}
```

### 6.3. Lỗi — call không tồn tại

| URL | `{{base_url}}/api/calls/nonexistent/accept` |

**Expected** (`400`):
```json
{"error": "cuộc gọi không tồn tại"}
```

---

## 7. Toggle Video (REST)

> **Prerequisite**: video call đang `connected` (5.1 → 6.1).

### 7.1. Caller bật video

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/{{video_call_id}}/video` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"video_enabled": true}` |

**Expected** (`200 OK`):
```json
{"message": "đã cập nhật trạng thái video"}
```

DB: `calls` → `video_enabled_caller = true`.

### 7.2. Caller tắt video

| Body | `{"video_enabled": false}` |

**Expected** (`200 OK`):
```json
{"message": "đã cập nhật trạng thái video"}
```

DB: `video_enabled_caller = false`.

### 7.3. Callee bật video

| Headers | `Authorization: Bearer {{token_B}}` |
| Body | `{"video_enabled": true}` |

**Expected** (`200`). DB: `video_enabled_callee = true`.

### 7.4. Callee tắt video

| Body | `{"video_enabled": false}` |

**Expected** (`200`). DB: `video_enabled_callee = false`.

### 7.5. Toggle nhiều lần

Bật → tắt → bật liên tiếp. Mỗi lần đều `200`, DB cập nhật đúng.

### 7.6. Lỗi — không phải participant (user C)

| Headers | `Authorization: Bearer {{token_C}}` |

**Expected** (`400`):
```json
{"error": "không phải người tham gia cuộc gọi"}
```

### 7.7. Lỗi — toggle trên voice call

> **Prerequisite**: tạo voice call riêng (call_type=voice), accept → connected.

| URL | `{{base_url}}/api/calls/{{voice_call_id}}/video` |
| Body | `{"video_enabled": true}` |

**Expected** (`400`):
```json
{"error": "cuộc gọi không phải video call"}
```

### 7.8. Lỗi — call chưa connected (calling)

> **Prerequisite**: tạo video call mới, chưa accept.

**Expected** (`400`):
```json
{"error": "cuộc gọi không ở trạng thái kết nối"}
```

### 7.9. Lỗi — call đã kết thúc

Sau khi call đã ended:

**Expected** (`400`):
```json
{"error": "cuộc gọi không ở trạng thái kết nối"}
```

### 7.10. Lỗi — call không tồn tại

| URL | `{{base_url}}/api/calls/nonexistent/video` |

**Expected** (`400`):
```json
{"error": "cuộc gọi không tồn tại"}
```

### 7.11. Lỗi — thiếu video_enabled

| Body | `{}` |

**Expected** (`400`):
```json
{"error": "video_enabled là bắt buộc"}
```

---

## 8. Toggle Mute trong Video Call

Tương tự voice call. Mute độc lập với video.

### 8.1. Caller mute

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/{{video_call_id}}/mute` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"muted": true}` |

**Expected** (`200`). DB: `muted_caller = true`.

### 8.2. Caller unmute

| Body | `{"muted": false}` |

**Expected** (`200`). DB: `muted_caller = false`.

---

## 9. End Video Call

### 9.1. Caller kết thúc (connected → ended)

> **Prerequisite**: video call đang `connected`.

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/{{video_call_id}}/end` |
| **Note** | Endpoint này là **WebSocket event** (`call:end`), không phải REST. |

Để test trên Postman:
1. Mở **WebSocket Request** (New → WebSocket)
2. URL: `ws://localhost:8080/api/calls/ws?token={{token_A}}`
3. Gửi message:
```json
{"type": "call:end", "payload": {"call_id": "{{video_call_id}}", "action": "end"}}
```

**Expected response** qua WebSocket:
```json
{
    "type": "call:status",
    "payload": {
        "call_id": "...",
        "status": "ended",
        "caller_id": "...",
        "callee_id": "...",
        "call_type": "video",
        "video_enabled_caller": false,
        "video_enabled_callee": false,
        "ended_at": 177...,
        "duration": 120
    }
}
```

### 9.2. Lỗi — user không phải participant

Dùng token_C gửi WS `call:end`.

**Expected**: WS error event:
```json
{"type": "error", "payload": {"message": "không phải người tham gia cuộc gọi"}}
```

---

## 10. WebSocket — call:video Event

### 10.1. Mở kết nối WS cho A và B

Trong Postman:
1. **New → WebSocket Request**
2. URL: `ws://localhost:8080/api/calls/ws?token={{token_A}}`
3. Click **Connect** (tab 1 cho A)
4. Mở tab 2: `ws://localhost:8080/api/calls/ws?token={{token_B}}`

### 10.2. A bật video → B nhận event

Trong khi video call đang connected, gọi REST ToggleVideo của A:

POST `{{base_url}}/api/calls/{{video_call_id}}/video` với token A, body `{"video_enabled": true}`

**B nhận** qua WS:
```json
{
    "type": "call:video",
    "payload": {
        "call_id": "...",
        "user_id": "{{userA_id}}",
        "video_enabled": true
    }
}
```

### 10.3. Gửi video_toggle qua WS

Thay vì REST, gửi trực tiếp qua WS:

```json
{
    "type": "call:video_toggle",
    "payload": {
        "call_id": "{{video_call_id}}",
        "video_enabled": true
    }
}
```

**Expected**: Server xử lý `callService.ToggleVideo`, broadcast `call:video` tới A và B.

### 10.4. Payload thiếu trường

```json
{"type": "call:video_toggle", "payload": {"call_id": "{{video_call_id}}"}}
```

**Expected**: WS error:
```json
{"type": "error", "payload": {"message": "dữ liệu video toggle không hợp lệ"}}
```

---

## 11. Test Flow — Kịch bản đầy đủ

### 11.1. Happy path: video call thành công

| Step | Action | Token | API | Expected |
|------|--------|-------|-----|----------|
| 1 | Login A | — | POST /api/auth/login | `token_A` |
| 2 | Login B | — | POST /api/auth/login | `token_B` |
| 3 | Kết bạn A→B | A | POST /api/friend-requests/send | sent |
| 4 | Chấp nhận | B | POST /friend-requests/:id/accept | accepted |
| 5 | A gọi B (video) | A | POST /api/calls/initiate `{callee_id:B, call_type:video}` | status=calling, lưu `video_call_id` |
| 6 | B chấp nhận | B | POST /api/calls/:video_call_id/accept | connected |
| 7 | A bật video | A | POST /api/calls/:video_call_id/video `{video_enabled:true}` | DB: video_enabled_caller=true |
| 8 | B bật video | B | POST /api/calls/:video_call_id/video `{video_enabled:true}` | DB: video_enabled_callee=true |
| 9 | A tắt video | A | POST /api/calls/:video_call_id/video `{video_enabled:false}` | DB: video_enabled_caller=false |
| 10 | A mute | A | POST /api/calls/:video_call_id/mute `{muted:true}` | OK |
| 11 | A kết thúc | A | WS `call:end` | status=ended |
| 12 | Xem history | A | GET /api/calls/history | total≥1 |

### 11.2. Error flows

| Step | Action | Token | API | Expected |
|------|--------|-------|-----|----------|
| 1 | Gọi chính mình | A | POST /api/calls/initiate `{callee_id:A, call_type:video}` | `400` "không thể gọi cho chính mình" |
| 2 | Không phải bạn | C | POST /api/calls/initiate `{callee_id:A, call_type:video}` | `400` "chỉ có thể gọi cho bạn bè" |
| 3 | Busy | C | POST /api/calls/initiate `{callee_id:B, call_type:video}` (khi A→B đang calling) | `200` "người dùng đang bận" |
| 4 | Caller không thể accept | A | POST /api/calls/:video_call_id/accept | `400` "chỉ người nhận mới có thể chấp nhận" |
| 5 | Toggle trên voice call | A | POST /api/calls/:voice_call_id/video | `400` "cuộc gọi không phải video call" |
| 6 | Toggle khi chưa connected | A | POST /api/calls/:new_call/video (calling) | `400` "cuộc gọi không ở trạng thái kết nối" |
| 7 | Toggle khi không phải participant | C | POST /api/calls/:video_call_id/video | `400` "không phải người tham gia cuộc gọi" |
| 8 | Thiếu video_enabled | A | POST /api/calls/:video_call_id/video `{}` | `400` "video_enabled là bắt buộc" |

### 11.3. Missed call flow (video)

| Step | Action | Token | Expected |
|------|--------|-------|----------|
| 1 | A gọi B (video) | A | status=calling |
| 2 | A end call (WS `call:end`) | A | status=missed (vì chưa connected). duration=0 |
| 3 | Xem history | A | duration=0, status=missed |

### 11.4. Reject flow (video)

| Step | Action | Token | Expected |
|------|--------|-------|----------|
| 1 | A gọi B (video) | A | status=calling |
| 2 | B reject | B | POST /api/calls/:video_call_id/reject → status=rejected |
| 3 | B không thể accept lại | B | `400` "cuộc gọi không ở trạng thái chờ" |

---

## 12. Kiểm tra Database

```sql
-- Xem tất cả cuộc gọi
SELECT id, caller_id, callee_id, call_type, status, 
       started_at, ended_at, duration,
       muted_caller, muted_callee, 
       video_enabled_caller, video_enabled_callee,
       created_at
FROM calls ORDER BY created_at DESC;

-- Đếm số cuộc gọi active của 1 user
SELECT COUNT(*) FROM calls 
WHERE (caller_id = '<user_id>' OR callee_id = '<user_id>')
  AND status IN ('calling', 'ringing', 'connected');

-- Kiểm tra video fields
SELECT id, call_type, video_enabled_caller, video_enabled_callee
FROM calls WHERE call_type = 'video';

-- Kiểm tra cột mới không bị NULL
SELECT COUNT(*) FROM calls WHERE video_enabled_caller IS NULL;
```

---

## 13. Postman Collection Export

```
Video Call Tests
├── Auth
│   ├── Login A
│   ├── Login B
│   └── Login C
├── ICE Servers
│   ├── Get ICE servers
│   ├── Không token → 401
│   └── Nhiều STUN + TURN
├── Friend Setup
│   ├── A gửi lời mời B
│   ├── B chấp nhận
│   ├── C gửi lời mời B
│   └── B chấp nhận (C)
├── Initiate Video Call
│   ├── Thành công (A→B video)
│   ├── Lỗi — gọi chính mình
│   ├── Lỗi — không phải bạn
│   ├── Busy — C gọi B
│   └── Lỗi — thiếu callee_id
├── Accept Video Call
│   ├── B chấp nhận
│   ├── Lỗi — A accept
│   └── Lỗi — call không tồn tại
├── ToggleVideo
│   ├── Caller bật video
│   ├── Caller tắt video
│   ├── Callee bật video
│   ├── Callee tắt video
│   ├── Lỗi — không phải participant
│   ├── Lỗi — voice call
│   ├── Lỗi — call chưa connected
│   ├── Lỗi — call không tồn tại
│   ├── Lỗi — thiếu video_enabled
│   └── Lỗi — call đã kết thúc
├── ToggleMute (video call)
│   ├── A mute
│   └── A unmute
├── Reject Video Call
│   ├── B reject
│   └── Lỗi — A reject
├── WebSocket Events
│   ├── WS connect A
│   ├── WS connect B
│   ├── A bật video → B nhận call:video
│   ├── A gửi video_toggle qua WS
│   └── Payload thiếu trường → error
├── Full Flows
│   ├── Happy path video call
│   ├── Missed call flow
│   └── Reject flow
└── GetCallHistory
    ├── A xem history
    ├── Phân trang
    └── Filter theo type=video
```

---

## 14. Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `401 Unauthorized` | Token thiếu/hết hạn | Login lại, cập nhật biến |
| `400 "cuộc gọi không phải video call"` | ToggleVideo trên call_type=voice | Dùng `{{video_call_id}}` (call_type=video) |
| `400 "video_enabled là bắt buộc"` | Thiếu field trong body | Gửi `{"video_enabled": true/false}` |
| `400 "cuộc gọi không ở trạng thái kết nối"` | Call chưa được accept | Accept call trước |
| WS không nhận `call:video` | WS chưa kết nối hoặc sai token | Kết nối qua `/api/calls/ws`, dùng đúng token |
| `"dịch vụ gọi không khả dụng"` | callService nil | Kết nối WS qua `/api/calls/ws` (không phải `/ws`) |
| `{"ice_servers":[]}` | Chưa cấu hình ICE_SERVER_URLS | Set biến môi trường IceServerUrls |
| `"chỉ có thể gọi cho bạn bè"` | A và B chưa là bạn | Hoàn thành bước kết bạn |
| Video không hiện trên UI | Chưa thêm video track vào PeerConnection | Frontend cần getUserMedia + addTrack |
