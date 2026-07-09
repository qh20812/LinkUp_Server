# Hướng dẫn Test Call History với Postman

## 1. Chuẩn bị Environment

### 1.1. Biến môi trường

Tạo một **Environment** mới trong Postman với các biến sau:

| Variable | Initial Value | Description |
|----------|--------------|-------------|
| `base_url` | `http://localhost:8080` | Server URL |
| `token_A` | *(empty)* | JWT token user A |
| `token_B` | *(empty)* | JWT token user B |
| `token_C` | *(empty)* | JWT token user C (3rd party) |
| `userA_id` | *(empty)* | UUID của user A |
| `userB_id` | *(empty)* | UUID của user B |
| `userC_id` | *(empty)* | UUID của user C |
| `call_id` | *(empty)* | ID cuộc gọi để test hide |
| `hidden_call_id` | *(empty)* | ID cuộc gọi đã ẩn |

### 1.2. Seed database

```bash
cd server
go build ./cmd/seed && ./seed.exe
```

Tài khoản mặc định (xem `cmd/seed/main.go` — tất cả đều dùng `Password123!`).

---

## 2. Lấy Token

### 2.1. Login user A

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

### 2.2. Login user B

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

## 3. Setup — Tạo dữ liệu cuộc gọi

Cần có ít nhất vài cuộc gọi trong DB để test history. Thực hiện các bước sau:

### 3.1. Kết bạn A → B

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/friend-requests/send` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"user_id": "{{userB_id}}"}` |

**Script**:
```javascript
pm.collectionVariables.set('request_id', pm.response.json().data.id);
```

### 3.2. B chấp nhận

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/friend-requests/{{request_id}}/accept` |
| **Headers** | `Authorization: Bearer {{token_B}}` |

### 3.3. A gọi B (voice)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/initiate` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"callee_id": "{{userB_id}}", "call_type": "voice"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('call_id', json.data.id);
```

### 3.4. B chấp nhận → connected

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/{{call_id}}/accept` |
| **Headers** | `Authorization: Bearer {{token_B}}` |

### 3.5. Kết thúc cuộc gọi (qua WS)

Dùng WebSocket:
1. New → WebSocket Request
2. URL: `ws://localhost:8080/api/calls/ws?token={{token_A}}`
3. Connect → gửi:
```json
{"type": "call:end", "payload": {"call_id": "{{call_id}}", "action": "end"}}
```

Call chuyển sang status `ended`.

### 3.6. Tạo cuộc gọi nhỡ

Lặp lại 3.3 (A gọi B), **không** accept → kết thúc qua WS:
```json
{"type": "call:end", "payload": {"call_id": "{{new_call_id}}", "action": "end"}}
```

Call chuyển sang status `missed`.

### 3.7. Kết bạn C với A (cho test direction)

Lặp lại 3.1–3.2: C gửi lời mời A, A chấp nhận.

### 3.8. B gọi A (để có incoming call)

Lặp lại 3.3–3.5 với token B gọi A + accept + end.

---

## 4. GetCallHistory

### 4.1. Lấy lịch sử mặc định

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/calls/history` |
| **Headers** | `Authorization: Bearer {{token_A}}` |

**Expected** (`200 OK`):
```json
{
    "data": [
        {
            "id": "...",
            "other_user": {"id": "...", "display_name": "User B", "avatar_url": ""},
            "call_type": "voice",
            "direction": "outgoing",
            "status": "ended",
            "is_missed": false,
            "duration": 45,
            "started_at": 1770000000000,
            "ended_at": 1770000045000,
            "created_at": 1769999955000
        },
        {
            "id": "...",
            "other_user": {"id": "...", "display_name": "User B", "avatar_url": ""},
            "call_type": "voice",
            "direction": "incoming",
            "status": "ended",
            ...
        }
    ],
    "total": 3,
    "limit": 20,
    "offset": 0
}
```

### 4.2. Phân trang

| URL | `{{base_url}}/api/calls/history?limit=1&offset=0` |

**Expected**: `data` có 1 item, `total` là tổng số records.

### 4.3. Limit > 100

| URL | `{{base_url}}/api/calls/history?limit=200` |

**Expected**: Clamp về `limit=100`.

### 4.4. Limit = 0

| URL | `{{base_url}}/api/calls/history?limit=0` |

**Expected**: Clamp về `limit=20` (default).

### 4.5. Offset âm

| URL | `{{base_url}}/api/calls/history?offset=-5` |

**Expected**: Clamp về `offset=0`.

### 4.6. Filter theo type = voice

| URL | `{{base_url}}/api/calls/history?type=voice` |

**Expected**: Chỉ trả về cuộc gọi voice.

### 4.7. Filter theo type = video

| URL | `{{base_url}}/api/calls/history?type=video` |

**Expected**: Chỉ trả về cuộc gọi video (nếu có).

### 4.8. Filter type không hợp lệ

| URL | `{{base_url}}/api/calls/history?type=fax` |

**Expected**: Type bị bỏ qua, trả tất cả.

### 4.9. Filter theo status = missed

| URL | `{{base_url}}/api/calls/history?status=missed` |

**Expected**: Chỉ trả về cuộc gọi nhỡ.

### 4.10. Filter theo status = ended

| URL | `{{base_url}}/api/calls/history?status=ended` |

**Expected**: Chỉ trả về cuộc gọi đã kết thúc.

### 4.11. Filter status không hợp lệ

| URL | `{{base_url}}/api/calls/history?status=deleted` |

**Expected**: Status bị bỏ qua, trả tất cả.

### 4.12. Sort theo duration

| URL | `{{base_url}}/api/calls/history?sort=duration` |

**Expected**: Kết quả sort theo duration tăng dần.

### 4.13. Sort theo call_type

| URL | `{{base_url}}/api/calls/history?sort=call_type` |

**Expected**: Kết quả sort theo call_type.

### 4.14. Sort column không hợp lệ

| URL | `{{base_url}}/api/calls/history?sort=invalid` |

**Expected**: Fallback về `sort=created_at`.

### 4.15. Order asc

| URL | `{{base_url}}/api/calls/history?sort=duration&order=asc` |

**Expected**: Kết quả sort tăng dần.

### 4.16. Order desc

| URL | `{{base_url}}/api/calls/history?sort=duration&order=desc` |

**Expected**: Kết quả sort giảm dần.

### 4.17. Order không hợp lệ

| URL | `{{base_url}}/api/calls/history?order=ascending` |

**Expected**: Fallback về `order=desc`.

### 4.18. Kết hợp tất cả

| URL | `{{base_url}}/api/calls/history?type=voice&status=ended&sort=duration&order=asc&limit=3&offset=0` |

**Expected**: Kết quả đúng filter, đúng sort, đúng phân trang.

### 4.19. User không có cuộc gọi

| Headers | `Authorization: Bearer {{token_C}}` |

**Expected**: `data: []`, `total: 0`.

### 4.20. Kiểm tra direction

Xem response của A:
- Cuộc gọi A gọi B → `direction = "outgoing"`
- Cuộc gọi B gọi A → `direction = "incoming"`

### 4.21. Kiểm tra other_user

Mỗi item có `other_user` với `id`, `display_name`, `avatar_url` đúng.

### 4.22. Kiểm tra is_missed

- Cuộc gọi ended → `is_missed = false`
- Cuộc gọi missed → `is_missed = true`

### 4.23. Không có token

| Headers | *(none)* |

**Expected**: `401 Unauthorized`.

---

## 5. HideCall

### 5.1. Ẩn cuộc gọi thành công

| Field | Value |
|-------|-------|
| **Method** | `DELETE` |
| **URL** | `{{base_url}}/api/calls/{{call_id}}/hide` |
| **Headers** | `Authorization: Bearer {{token_A}}` |

**Expected** (`200 OK`):
```json
{"message": "cuộc gọi đã được ẩn khỏi lịch sử"}
```

**Script** để lưu hidden_call_id:
```javascript
pm.collectionVariables.set('hidden_call_id', '{{call_id}}');
```

### 5.2. Kiểm tra cuộc gọi đã ẩn không xuất hiện

Sau khi ẩn, GET `/api/calls/history` với token A:
Cuộc gọi đã ẩn KHÔNG xuất hiện trong data.

### 5.3. Kiểm tra user B vẫn thấy cuộc gọi

| Headers | `Authorization: Bearer {{token_B}}` |
| URL | `{{base_url}}/api/calls/history` |

Cuộc gọi vẫn xuất hiện trong history của B (user-level soft delete).

### 5.4. Ẩn cuộc gọi đã ẩn (idempotent)

| Method | `DELETE` |
| URL | `{{base_url}}/api/calls/{{hidden_call_id}}/hide` |
| Headers | `Authorization: Bearer {{token_A}}` |

**Expected** (`200 OK`): Không lỗi, không duplicate record.

### 5.5. Lỗi — user không phải participant

| Method | `DELETE` |
| URL | `{{base_url}}/api/calls/{{call_id}}/hide` |
| Headers | `Authorization: Bearer {{token_C}}` |

**Expected** (`400`):
```json
{"error": "không phải người tham gia cuộc gọi"}
```

### 5.6. Lỗi — cuộc gọi không tồn tại

| URL | `{{base_url}}/api/calls/nonexistent/hide` |

**Expected** (`400`):
```json
{"error": "cuộc gọi không tồn tại"}
```

### 5.7. Không có token

| Headers | *(none)* |

**Expected**: `401 Unauthorized`.

---

## 6. GetMissedCallCount

### 6.1. Đếm cuộc gọi nhỡ (có missed call)

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/calls/missed/count` |
| **Headers** | `Authorization: Bearer {{token_A}}` |

**Expected** (`200 OK`):
```json
{"count": 1}
```

### 6.2. User không có missed call

| Headers | `Authorization: Bearer {{token_C}}` |

**Expected**: `{"count": 0}`.

### 6.3. Sau khi MarkMissedRead, count = 0

1. POST `/api/calls/missed/read` (token A)
2. GET `/api/calls/missed/count` (token A)

**Expected**: `{"count": 0}`.

### 6.4. Ẩn cuộc gọi nhỡ vẫn tính

Ẩn một missed call → count vẫn bằng 1 trước khi read (hide không ảnh hưởng missed count).

### 6.5. Không có token

| Headers | *(none)* |

**Expected**: `401 Unauthorized`.

---

## 7. MarkMissedRead

### 7.1. Đánh dấu đã đọc thành công

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/calls/missed/read` |
| **Headers** | `Authorization: Bearer {{token_A}}` |

**Expected** (`200 OK`):
```json
{"message": "đã đánh dấu đã đọc"}
```

DB: `profiles.last_read_missed_at` được cập nhật.

### 7.2. Gọi nhiều lần (idempotent)

POST lại `/api/calls/missed/read` → không lỗi, `last_read_missed_at` cập nhật lại.

### 7.3. Missed call mới sau khi read

1. POST `/api/calls/missed/read` (token A)
2. Tạo cuộc gọi nhỡ mới cho A (B gọi A, end trước khi connected)
3. GET `/api/calls/missed/count` (token A)

**Expected**: `{"count": 1}` (cuộc gọi mới sau thời điểm read).

### 7.4. Không có token

| Headers | *(none)* |

**Expected**: `401 Unauthorized`.

---

## 8. WebSocket — call:missed Event

### 8.1. End call khi chưa connected → missed WS

**Prerequisite**: A gọi B (calling), B đang online qua WS.

1. Kết nối WS của B: `ws://localhost:8080/api/calls/ws?token={{token_B}}`
2. Kết nối WS của A: `ws://localhost:8080/api/calls/ws?token={{token_A}}`
3. A gọi B: POST `/api/calls/initiate`
4. **Chưa accept**
5. A gửi WS `call:end`:
```json
{"type": "call:end", "payload": {"call_id": "{{call_id}}", "action": "end"}}
```

**B nhận**:
```json
{
    "type": "call:missed",
    "payload": {
        "call_id": "...",
        "caller_id": "{{userA_id}}",
        "timestamp": 1770000000000
    }
}
```

**A nhận**:
```json
{
    "type": "call:status",
    "payload": {
        "call_id": "...",
        "status": "missed",
        ...
    }
}
```

### 8.2. End call khi connected → ended, không có missed WS

1. A gọi B, B accept → connected
2. A gửi WS `call:end`

**Expected**: Status = `ended`. KHÔNG có `call:missed` event.

---

## 9. Test Flow — Kịch bản đầy đủ

### 9.1. Happy path: history sau nhiều cuộc gọi

| Step | Action | Token | API | Expected |
|------|--------|-------|-----|----------|
| 1 | Login A + B | — | POST /api/auth/login | tokens |
| 2 | Kết bạn A→B | A | POST /api/friend-requests/send | accepted |
| 3 | A gọi B → accept → end (connected) | A/B | POST /api/calls/initiate → accept → WS end | status=ended |
| 4 | A gọi B → end trước khi connected | A | POST /api/calls/initiate → WS end | status=missed |
| 5 | B gọi A → accept → end | B/A | POST /api/calls/initiate → accept → WS end | status=ended, direction=incoming (A view) |
| 6 | A xem history | A | GET /api/calls/history | total=3, data có ended + missed + incoming |
| 7 | A ẩn cuộc gọi ended | A | DELETE /api/calls/:call_id/hide | OK |
| 8 | A xem history lại | A | GET /api/calls/history | total=2 (đã ẩn 1) |
| 9 | B xem history | B | GET /api/calls/history | total=3 (vẫn thấy đủ) |
| 10 | A đếm missed | A | GET /api/calls/missed/count | count=1 |
| 11 | A đánh dấu đã đọc | A | POST /api/calls/missed/read | OK |
| 12 | A đếm lại | A | GET /api/calls/missed/count | count=0 |

### 9.2. Filter + pagination flow

| Step | Action | Expected |
|------|--------|----------|
| 1 | GET `/api/calls/history?type=voice&status=ended` | Chỉ ended voice calls |
| 2 | GET `/api/calls/history?status=missed` | Chỉ missed calls |
| 3 | GET `/api/calls/history?sort=duration&order=asc&limit=2` | 2 items, sort duration ASC |
| 4 | GET `/api/calls/history?type=video` | Empty nếu chưa có video call |

---

## 10. Kiểm tra Database

```sql
-- Xem tất cả cuộc gọi
SELECT id, caller_id, callee_id, call_type, status,
       started_at, ended_at, duration,
       created_at
FROM calls ORDER BY created_at DESC;

-- Xem cuộc gọi đã ẩn
SELECT call_id, user_id, created_at FROM call_hidden;

-- Xem last_read_missed_at của user
SELECT user_id, display_name, last_read_missed_at
FROM profiles WHERE last_read_missed_at IS NOT NULL;

-- Đếm missed call chưa đọc (manual)
SELECT COUNT(*) FROM calls c
WHERE c.callee_id = '<user_id>'
  AND c.status = 'missed'
  AND NOT EXISTS (
    SELECT 1 FROM call_hidden h
    WHERE h.call_id = c.id AND h.user_id = '<user_id>'
  )
  AND (c.created_at >= (SELECT last_read_missed_at FROM profiles WHERE user_id = '<user_id>')
       OR (SELECT last_read_missed_at FROM profiles WHERE user_id = '<user_id>') IS NULL);
```

---

## 11. Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `401 Unauthorized` | Token thiếu/hết hạn | Login lại, cập nhật biến |
| `data: []` khi mong đợi có data | Chưa tạo call history | Chạy InitiateCall → Accept → End trước |
| `total: 0` | User chưa có cuộc gọi nào | Tạo ít nhất 1 cuộc gọi |
| Limit không clamp | Gửi sai param name | Dùng `limit`, không phải `page_size` |
| Filter không hoạt động | Sai param name | Dùng `type`, `status`, không phải `call_type` |
| `"cuộc gọi không tồn tại"` khi hide | Sai call_id hoặc call đã bị xoá | Kiểm tra DB |
| `"không phải người tham gia cuộc gọi"` | User không phải caller/callee | Dùng đúng token của participant |
| `count` luôn = 0 dù có missed call | `last_read_missed_at` đã được set | Kiểm tra DB hoặc tạo missed call mới |
| WS không nhận `call:missed` | WS chưa kết nối hoặc sai endpoint | Kết nối qua `/api/calls/ws` |
| Missed count không giảm sau hide | Hide không ảnh hưởng missed count | Dùng MarkMissedRead để reset |

---

## 12. Postman Collection Export

```
Call History Tests
├── Auth
│   ├── Login A
│   ├── Login B
│   └── Login C
├── Setup Test Data
│   ├── Kết bạn A→B
│   ├── B chấp nhận
│   ├── A gọi B → accept → end (connected)
│   ├── A gọi B → end trước connected (missed)
│   ├── B gọi A → accept → end (incoming)
│   └── Kết bạn C với A
├── GetCallHistory
│   ├── Mặc định (limit=20)
│   ├── Phân trang (limit=1, offset=0)
│   ├── Limit > 100 → clamp
│   ├── Limit = 0 → fallback
│   ├── Offset âm → clamp
│   ├── Filter type=voice
│   ├── Filter type=video
│   ├── Filter type invalid → bỏ qua
│   ├── Filter status=missed
│   ├── Filter status=ended
│   ├── Filter status invalid → bỏ qua
│   ├── Sort=duration
│   ├── Sort=call_type
│   ├── Sort invalid → fallback
│   ├── Order=asc
│   ├── Order=desc
│   ├── Order invalid → fallback
│   ├── Combined (type+status+sort+order+page)
│   ├── User không có history → empty
│   ├── Direction check
│   ├── OtherUser check
│   ├── is_missed check
│   └── Không token → 401
├── HideCall
│   ├── Ẩn thành công
│   ├── Verify: ẩn không xuất hiện
│   ├── Verify: user kia vẫn thấy
│   ├── Ẩn lại (idempotent)
│   ├── Lỗi — không phải participant
│   ├── Lỗi — call không tồn tại
│   └── Lỗi — không token
├── GetMissedCallCount
│   ├── Có missed call
│   ├── Không có missed call
│   ├── Sau MarkMissedRead → 0
│   ├── Ẩn không ảnh hưởng count
│   └── Không token → 401
├── MarkMissedRead
│   ├── Đánh dấu đã đọc
│   ├── Idempotent (gọi lại)
│   ├── Missed call mới sau read
│   └── Không token → 401
├── WebSocket call:missed
│   ├── End trước connected → missed
│   ├── End sau connected → ended (no missed)
│   └── B nhận call:missed event
└── Full Flow
    ├── Happy path (nhiều call + hide + missed)
    └── Filter + pagination
```
