# Hướng dẫn Test Voice Call với Postman

## 1. Chuẩn bị Environment

### 1.1. Biến môi trường

Tạo một **Environment** mới trong Postman với các biến sau:

| Variable | Initial Value | Description |
|----------|--------------|-------------|
| `base_url` | `http://localhost:8080` | Server URL |
| `token_A` | *(empty)* | JWT token user A (caller) |
| `token_B` | *(empty)* | JWT token user B (callee) |
| `token_C` | *(empty)* | JWT token user C (3rd party) |
| `call_id` | *(empty)* | ID của cuộc gọi vừa tạo |
| `userA_id` | *(empty)* | UUID của user A |
| `userB_id` | *(empty)* | UUID của user B |
| `userC_id` | *(empty)* | UUID của user C |
| `video_call_id` | *(empty)* | ID của video call (type=video) |
| `ice_servers` | *(empty)* | Danh sách ICE servers (GET /ice-servers) |

### 1.2. Seed database

Chạy seed để có dữ liệu user test:

```bash
cd server
go build ./cmd/seed && ./seed.exe
```

Username/password mặc định (xem trong `cmd/seed/main.go` — tất cả đều dùng `Password123!`).

---

## 2. Lấy Token — Đăng nhập

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

> Thay `seed_user_A@example.com` bằng email user có sẵn trong seed data.

### 2.3. Get ICE servers

Trước khi test video call, lấy danh sách ICE server (cần auth — giống như WS call):

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

Response mẫu:
```json
{
  "ice_servers": [
    {"urls": "stun:stun.l.google.com:19302"}
  ]
}
```

### 2.2. Login user B (callee)

Same as above, dùng email khác. **Tests** tab:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_B', json.data.access_token);
pm.collectionVariables.set('userB_id', json.data.user.id);
```

### 2.3. Login user C (3rd party)

**Tests** tab:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_C', json.data.access_token);
pm.collectionVariables.set('userC_id', json.data.user.id);
```

---

## 3. Gửi lời mời kết bạn (A → B)

Voice call yêu cầu A và B phải là bạn bè.

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/friend-requests/send` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"user_id": "{{userB_id}}"}` |

### 3.1. Chấp nhận lời mời (B)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/friend-requests/{{request_id}}/accept` |
| **Headers** | `Authorization: Bearer {{token_B}}` |

---

## 4. Test InitiateCall

### 4.1. Thành công — A gọi B

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/initiate` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"callee_id": "{{userB_id}}", "call_type": "voice"}` |

**Expected Response** (`200 OK`):
```json
{
    "data": {
        "id": "uuid-call-id-here",
        "caller_id": "{{userA_id}}",
        "callee_id": "{{userB_id}}",
        "call_type": "voice",
        "is_group": false,
        "status": "calling",
        "started_at": null,
        "ended_at": null,
        "duration": 0,
        "muted_caller": false,
        "muted_callee": false,
        "created_at": "2026-07-07T..."
    }
}
```

**Script** để lưu `call_id`:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('call_id', json.data.id);
```

### 4.2. Lỗi — gọi chính mình

| Field | Value |
|-------|-------|
| **Body** | `{"callee_id": "{{userA_id}}", "call_type": "voice"}` |

**Expected** (`400 Bad Request`):
```json
{"error": "không thể gọi cho chính mình"}
```

### 4.3. Lỗi — caller đang có cuộc gọi khác

1. Chạy **4.1** (A→B call created) — giữ nguyên trạng thái
2. Gọi tiếp **4.1** với callee_id = userC_id

**Expected** (`400 Bad Request`):
```json
{"error": "bạn đang có cuộc gọi khác"}
```

### 4.4. Lỗi — không phải bạn bè

Dùng user C (chưa kết bạn với ai) gọi A:

| Field | Value |
|-------|-------|
| **Headers** | `Authorization: Bearer {{token_C}}` |
| **Body** | `{"callee_id": "{{userA_id}}", "call_type": "voice"}` |

**Expected** (`400 Bad Request`):
```json
{"error": "chỉ có thể gọi cho bạn bè"}
```

### 4.5. Busy — callee đang bận

1. Kết bạn C với B (lặp lại bước 3)
2. A→B call đang ở trạng thái `calling` (từ 4.1)
3. C gọi B:

| Field | Value |
|-------|-------|
| **Headers** | `Authorization: Bearer {{token_C}}` |
| **Body** | `{"callee_id": "{{userB_id}}", "call_type": "voice"}` |

**Expected** (`200 OK`):
```json
{"message": "người dùng đang bận"}
```

> **Lưu ý**: Response là `200` (không phải `400`) vì request hợp lệ, chỉ là callee busy — controller trả về `200` với message.

### 4.6. Call type video

| Field | Value |
|-------|-------|
| **Body** | `{"callee_id": "{{userB_id}}", "call_type": "video"}` |

**Expected**: `call_type` trong response là `"video"`.

### 4.7. Thiếu callee_id

| Field | Value |
|-------|-------|
| **Body** | `{"call_type": "voice"}` |

**Expected** (`400 Bad Request`):
```json
{"error": "callee_id và call_type là bắt buộc"}
```

---

## 5. Test AcceptCall

> **Prerequisite**: A→B call đang ở status `calling` (chạy 4.1).

### 5.1. B chấp nhận

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/{{call_id}}/accept` |
| **Headers** | `Authorization: Bearer {{token_B}}` |

**Expected** (`200 OK`):
```json
{"message": "cuộc gọi đã được chấp nhận"}
```

Kiểm tra DB: `calls` table → `status = "connected"`, `started_at` không null.

### 5.2. Lỗi — A không thể accept

| Field | Value |
|-------|-------|
| **Headers** | `Authorization: Bearer {{token_A}}` |

**Expected** (`400 Bad Request`):
```json
{"error": "chỉ người nhận mới có thể chấp nhận cuộc gọi"}
```

### 5.3. Lỗi — cuộc gọi đã kết thúc

Sau khi call đã ended (chạy 6.1 trước):

**Expected** (`400 Bad Request`):
```json
{"error": "cuộc gọi không ở trạng thái chờ"}
```

### 5.4. Lỗi — call_id không tồn tại

| Field | Value |
|-------|-------|
| **URL** | `{{base_url}}/api/calls/nonexistent-id/accept` |

**Expected** (`400 Bad Request`):
```json
{"error": "cuộc gọi không tồn tại"}
```

---

## 6. Test EndCall

### 6.1. Caller kết thúc cuộc gọi (connected)

> **Prerequisite**: call đang ở status `connected` (chạy 4.1 → 5.1).

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/{{call_id}}/end` |
| **Note** | Endpoint này là WebSocket event (`call:end`), không phải REST. |

⚠️ **`/api/calls/:callID/end` KHÔNG có REST route** — chỉ có WS event `call:end`.

Để test trên Postman:
1. Mở **WebSocket Request** (New → WebSocket)
2. URL: `ws://localhost:8080/api/calls/ws?token={{token_A}}`
3. Gửi message:
```json
{"type": "call:end", "payload": {"call_id": "{{call_id}}", "action": "end"}}
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
        ...
        "ended_at": 177...,
        "duration": 120
    }
}
```

Hoặc nếu call chưa connected (vẫn calling), status sẽ là `"missed"`.

### 6.2. Lỗi — user không phải participant

Dùng token_C gửi WS `call:end`.

**Expected**: WS error event:
```json
{"type": "error", "payload": {"message": "không phải người tham gia cuộc gọi"}}
```

### 6.3. Lỗi — cuộc gọi đã kết thúc

Gọi `call:end` lần thứ 2 trên cùng call.

**Expected**: WS error "cuộc gọi đã kết thúc".

---

## 7. Test RejectCall

### 7.1. B từ chối cuộc gọi

> **Prerequisite**: A→B call đang `calling` (chạy 4.1 mới).

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/{{call_id}}/reject` |
| **Headers** | `Authorization: Bearer {{token_B}}` |

**Expected**:
```json
{"message": "cuộc gọi đã bị từ chối"}
```

### 7.2. Lỗi — A không thể reject

| Field | Value |
|-------|-------|
| **Headers** | `Authorization: Bearer {{token_A}}` |

**Expected** (`400 Bad Request`):
```json
{"error": "chỉ người nhận mới có thể từ chối cuộc gọi"}
```

---

## 8. Test ToggleMute

### 8.1. Caller mute

> **Prerequisite**: call đang `connected` (4.1 → 5.1).

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/{{call_id}}/mute` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"muted": true}` |

**Expected**:
```json
{"message": "đã cập nhật trạng thái tắt tiếng"}
```

Kiểm tra DB: `calls` → `muted_caller = true`.

### 8.2. Caller unmute

| Field | Value |
|-------|-------|
| **Body** | `{"muted": false}` |

**Expected**: `muted_caller = false` trong DB.

### 8.3. Callee mute

| Field | Value |
|-------|-------|
| **Headers** | `Authorization: Bearer {{token_B}}` |
| **Body** | `{"muted": true}` |

**Expected**: DB → `muted_callee = true`.

### 8.4. Lỗi — user không phải participant

| Field | Value |
|-------|-------|
| **Headers** | `Authorization: Bearer {{token_C}}` |

**Expected**:
```json
{"error": "không phải người tham gia cuộc gọi"}
```

### 8.5. Lỗi — call chưa connected

> **Prerequisite**: call đang `calling` (chưa accept).

**Expected**:
```json
{"error": "cuộc gọi không ở trạng thái kết nối"}
```

### 8.6. Lỗi — thiếu muted

| Field | Value |
|-------|-------|
| **Body** | `{}` |

**Expected**:
```json
{"error": "muted là bắt buộc"}
```

---

## 9. Test GetCallDetail

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/calls/{{call_id}}` |
| **Headers** | `Authorization: Bearer {{token_A}}` |

**Expected** (`200 OK`):
```json
{
    "data": {
        "id": "...",
        "caller_id": "...",
        "callee_id": "...",
        "call_type": "voice",
        "is_group": false,
        "status": "connected",
        "started_at": "...",
        "ended_at": null,
        ...
    }
}
```

### 9.1. Lỗi — không phải participant

| Field | Value |
|-------|-------|
| **Headers** | `Authorization: Bearer {{token_C}}` |

**Expected**:
```json
{"error": "không phải người tham gia cuộc gọi"}
```

---

## 10. Test GetCallHistory

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/calls/history` |
| **Headers** | `Authorization: Bearer {{token_A}}` |

**Expected** (`200 OK`):
```json
{
    "data": [...],
    "total": 3,
    "limit": 20,
    "offset": 0
}
```

### 10.1. Phân trang

| Field | Value |
|-------|-------|
| **URL** | `{{base_url}}/api/calls/history?limit=2&offset=1` |

**Expected**: `data` có 2 items, bắt đầu từ offset 1.

### 10.2. Limit > 100

| Field | Value |
|-------|-------|
| **URL** | `{{base_url}}/api/calls/history?limit=200` |

**Expected**: Fallback về 20 — `limit` trong response là `20`.

### 10.3. User chưa có lịch sử

| Field | Value |
|-------|-------|
| **Headers** | `Authorization: Bearer {{token_C}}` |

**Expected**: `data: []`, `total: 0`.

---

## 11. Test WebSocket Signal

### 11.1. Mở kết nối WS

Trong Postman:
1. **New → WebSocket Request**
2. URL: `ws://localhost:8080/api/calls/ws?token={{token_A}}`
3. Click **Connect**

### 11.2. Gửi signal (caller → callee)

Sau khi call đã connected, gửi message:

```json
{
    "type": "call:signal",
    "payload": {
        "call_id": "{{call_id}}",
        "signal": {"type": "offer", "sdp": "test_sdp_offer"}
    }
}
```

**Expected**: callee (B) nhận được qua WS của B:
```json
{
    "type": "call:signal",
    "payload": {
        "call_id": "...",
        "sender_id": "{{userA_id}}",
        "signal": {"type": "offer", "sdp": "test_sdp_offer"}
    }
}
```

> Để test callee nhận WS, mở tab WebSocket thứ 2 với `token={{token_B}}` kết nối tới cùng URL.

### 11.3. Lỗi — signal với call_id không tồn tại

**Expected**: WS error event.

---

## 12. Test Flow — Kịch bản đầy đủ

### 12.1. Happy path: voice call thành công

| Step | Action | Token | Endpoint | Expected |
|------|--------|-------|----------|----------|
| 1 | Login A | — | POST /api/auth/login | `token_A` |
| 2 | Login B | — | POST /api/auth/login | `token_B` |
| 3 | Kết bạn A→B | A | POST /api/friend-requests/send | sent |
| 4 | Chấp nhận | B | POST /friend-requests/:id/accept | accepted |
| 5 | A gọi B (voice) | A | POST /api/calls/initiate | status=calling, lưu `call_id` |
| 6 | B chấp nhận | B | POST /api/calls/:call_id/accept | connected |
| 7 | A mute | A | POST /api/calls/:call_id/mute `{muted:true}` | OK |
| 8 | B unmute A | B | POST /api/calls/:call_id/mute `{muted:false}` | OK |
| 9 | Gửi signal A→B | A | WS `call:signal` | B nhận signal |
| 10 | Gửi signal B→A | B | WS `call:signal` | A nhận signal |
| 11 | A kết thúc | A | WS `call:end` | status=ended |
| 12 | Xem history | A | GET /api/calls/history | total≥1 |

### 12.2. Busy flow: 2 caller → 1 callee

| Step | Action | Token | Expected |
|------|--------|-------|----------|
| 1 | Kết bạn A→B + C→B | — | accepted |
| 2 | A gọi B | A | status=calling |
| 3 | C gọi B | C | message="người dùng đang bận" |
| 4 | Xem history của C | C | total=0 (không tạo call) |

### 12.3. Missed call flow

| Step | Action | Token | Expected |
|------|--------|-------|----------|
| 1 | A gọi B | A | status=calling, B offline |
| 2 | A end call (WS `call:end`) | A | status=missed (vì chưa connected) |
| 3 | Xem history | A | duration=0, status=missed |

### 12.4. Reject flow

| Step | Action | Token | Expected |
|------|--------|-------|----------|
| 1 | A gọi B | A | status=calling |
| 2 | B reject | B | status=rejected |
| 3 | B không thể accept lại | B | error "cuộc gọi không ở trạng thái chờ" |

### 12.5. Video call flow

1. Khởi tạo video call riêng (dùng `video_call_id` để không ảnh hưởng voice call test):

| Step | Action | Token | API | Expected |
|------|--------|-------|-----|----------|
| 1 | A gọi B video | A | POST `/api/calls/initiate` `{"callee_id":"{{userB_id}}","call_type":"video"}` | `200`, DB: `call_type=video` |
| 2 | Lưu `video_call_id` | — | — | Script: `pm.collectionVariables.set('video_call_id', json.data.id)` |
| 3 | B chấp nhận | B | POST `/api/calls/{{video_call_id}}/accept` | `200` |
| 4 | A bật video | A | POST `/api/calls/{{video_call_id}}/video` `{"video_enabled":true}` | `200`, DB: `video_enabled_caller=true` |
| 5 | B bật video | B | POST `/api/calls/{{video_call_id}}/video` `{"video_enabled":true}` | `200`, DB: `video_enabled_callee=true` |
| 6 | A tắt video | A | POST `/api/calls/{{video_call_id}}/video` `{"video_enabled":false}` | `200`, DB: `video_enabled_caller=false` |
| 7 | A kết thúc | A | WS `call:end` | status=ended |

2. Lỗi — toggle video trên voice call:

| Step | Action | Token | API | Expected |
|------|--------|-------|-----|----------|
| 1 | Dùng `{{call_id}}` (voice call) | A | POST `/api/calls/{{call_id}}/video` `{"video_enabled":true}` | `400` "cuộc gọi không phải video call" |

---

## 13. GetIceServers

Test endpoint public lấy cấu hình ICE server cho WebRTC.

### 13.1. Lấy danh sách ICE servers

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/calls/ice-servers` |
| **Headers** | `Authorization: Bearer {{token_A}}` |

**Expected** (`200 OK`):
```json
{
    "ice_servers": [
        {"urls": "stun:stun.l.google.com:19302"}
    ]
}
```

### 13.2. Verify yêu cầu token

Gọi không có `Authorization` header → `401 Unauthorized`.

---

## 14. Test ToggleVideo

> **Prerequisite**: video call đang `connected` (section 12.5).

### 14.1. Caller bật video

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/{{video_call_id}}/video` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"video_enabled": true}` |

**Expected**:
```json
{"message": "đã cập nhật trạng thái video"}
```
DB: `video_enabled_caller = true`.

### 14.2. Caller tắt video

| Field | Value |
|-------|-------|
| **Body** | `{"video_enabled": false}` |

**Expected**: DB `video_enabled_caller = false`.

### 14.3. Callee bật video

| Field | Value |
|-------|-------|
| **Headers** | `Authorization: Bearer {{token_B}}` |
| **Body** | `{"video_enabled": true}` |

**Expected**: DB `video_enabled_callee = true`.

### 14.4. Lỗi — user không phải participant

| Field | Value |
|-------|-------|
| **Headers** | `Authorization: Bearer {{token_C}}` |

**Expected**:
```json
{"error": "không phải người tham gia cuộc gọi"}
```

### 14.5. Lỗi — call chưa connected

> **Prerequisite**: tạo video call mới, chưa accept.

**Expected**:
```json
{"error": "cuộc gọi không ở trạng thái kết nối"}
```

### 14.6. Lỗi — voice call (call_type=voice)

Dùng `{{call_id}}` (voice call đã connected):

**Expected**:
```json
{"error": "cuộc gọi không phải video call"}
```

### 14.7. Lỗi — thiếu video_enabled

| Field | Value |
|-------|-------|
| **Body** | `{}` |

**Expected**:
```json
{"error": "video_enabled là bắt buộc"}
```

---

## 15. Kiểm tra Database

Sau mỗi test, kiểm tra database để verify:

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

## 16. Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `401 Unauthorized` | Token thiếu/hết hạn | Login lại, cập nhật biến |
| `404 Not Found` cho route | Route không tồn tại | Kiểm tra method+path |
| WS không kết nối | Server chưa chạy | `air` hoặc `go run ./cmd` |
| WS `"loại sự kiện không xác định"` | Payload sai format | Kiểm tra JSON type/payload |
| `"dịch vụ gọi không khả dụng"` | callService nil trong WS client | Connect qua `/api/calls/ws` (không phải `/ws`) |
| Call không được tạo | Friend check fail | Kiểm tra bảng `friends` |
| `"bạn đang có cuộc gọi khác"` | Caller đã có active call | End call cũ trước |
| `"cuộc gọi không phải video call"` | ToggleVideo trên call_type=voice | Dùng video call để test |
| `{"ice_servers":[]}` | Chưa cấu hình ICE_SERVER_URLS trong .env | Set biến môi trường |
| `401` trên `/ice-servers` | Thiếu Authorization header | Gửi kèm token |

---

## 17. Postman Collection Export

Để import nhanh, tạo Collection với cấu trúc:

```
Voice Call Tests
├── Auth
│   ├── Login A
│   ├── Login B
│   └── Login C
├── Friend Setup
│   ├── A gửi lời mời B
│   └── B chấp nhận
├── InitiateCall
│   ├── Thành công (A→B)
│   ├── Lỗi — gọi chính mình
│   ├── Lỗi — không phải bạn
│   ├── Busy — C gọi B
│   └── Video call
├── AcceptCall
│   ├── B chấp nhận
│   ├── Lỗi — A accept
│   └── Lỗi — call không tồn tại
├── RejectCall
│   ├── B reject
│   └── Lỗi — A reject
├── ToggleMute
│   ├── A mute
│   ├── A unmute
│   ├── B mute
│   ├── Lỗi — không phải participant
│   └── Lỗi — call chưa connected
├── GetIceServers
│   ├── Lấy danh sách
│   └── Yêu cầu token → 401
├── Video Call Flow
│   ├── A gọi B (video)
│   ├── B chấp nhận
│   ├── A bật video
│   ├── B bật video
│   ├── A tắt video
│   └── Lỗi — toggle trên voice call
├── ToggleVideo
│   ├── A bật
│   ├── A tắt
│   ├── B bật
│   ├── Lỗi — không phải participant
│   ├── Lỗi — call chưa connected
│   ├── Lỗi — voice call
│   └── Lỗi — thiếu video_enabled
├── GetCallDetail
│   ├── A xem detail
│   └── Lỗi — C xem detail
└── GetCallHistory
    ├── A xem history
    ├── Phân trang
    └── C không có history
```
