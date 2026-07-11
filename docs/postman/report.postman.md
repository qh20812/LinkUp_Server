# Hướng dẫn Test Report Module với Postman

## 1. Chuẩn bị Environment

### 1.1. Biến môi trường

Tạo một **Environment** mới trong Postman với các biến sau:

| Variable | Initial Value | Description |
|----------|--------------|-------------|
| `base_url` | `http://localhost:8080` | Server URL |
| `token_A` | *(empty)* | JWT token user A (reporter) |
| `token_B` | *(empty)* | JWT token user B (post author) |
| `token_C` | *(empty)* | JWT token user C |
| `token_admin` | *(empty)* | JWT token SuperAdmin |
| `userA_id` | *(empty)* | UUID của user A |
| `userB_id` | *(empty)* | UUID của user B |
| `userC_id` | *(empty)* | UUID của user C |
| `admin_id` | *(empty)* | UUID của admin |
| `post_id` | *(empty)* | ID post để test report |
| `comment_id` | *(empty)* | ID comment để test report |
| `report_id` | *(empty)* | ID report vừa tạo |
| `report_id_2` | *(empty)* | ID report thứ 2 |

### 1.2. Seed database

```bash
cd server
go build ./cmd/seed && ./seed.exe
```

Tài khoản mặc định (xem `cmd/seed/main.go` — tất cả đều dùng `Password123!`).

> **Lưu ý**: Seed tạo 10 user (A–J), mỗi user có 3 post. User đầu tiên trong danh sách admin có thể có role SuperAdmin. Kiểm tra DB để xác nhận.

---

## 2. Lấy Token

### 2.1. Login user A (reporter)

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

### 2.2. Login user B (post author)

| Body | `{"email": "seed_user_B@example.com", "password": "Password123!"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_B', json.data.access_token);
pm.collectionVariables.set('userB_id', json.data.user.id);
```

### 2.3. Login user C

| Body | `{"email": "seed_user_C@example.com", "password": "Password123!"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_C', json.data.access_token);
pm.collectionVariables.set('userC_id', json.data.user.id);
```

### 2.4. Login admin

| Body | `{"email": "seed_user_1@example.com", "password": "Password123!"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_admin', json.data.access_token);
pm.collectionVariables.set('admin_id', json.data.user.id);
```

> **Lưu ý**: Nếu user không có role SuperAdmin, cần cập nhật DB: `UPDATE user_roles SET ...` hoặc seed lại với admin role. Hoặc dùng endpoint `POST /api/admin/reports/:reportID/decision` sẽ trả `403` nếu không phải superadmin.

---

## 3. Setup — Tạo dữ liệu test

### 3.1. Tạo post của user B

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/posts` |
| **Headers** | `Authorization: Bearer {{token_B}}`, `Content-Type: application/json` |
| **Body** | `{"title": "Test post for report", "content": "Nội dung cần report", "status": "public"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('post_id', json.data.id);
```

### 3.2. Tạo comment trên post

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/posts/{{post_id}}/comments` |
| **Headers** | `Authorization: Bearer {{token_B}}`, `Content-Type: application/json` |
| **Body** | `{"content": "Comment cần report"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('comment_id', json.data.id);
```

---

## 4. CreateReport

### 4.1. Report post thành công

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/reports` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"target_type": "post", "target_id": "{{post_id}}", "report_type": "spam", "reason_detail": "Bài viết chứa nội dung spam"}` |

**Expected** (`201 Created`):
```json
{"message": "Báo cáo đã được gửi thành công"}
```

**Script**:
```javascript
// Lưu report_id từ DB hoặc dùng ListReports để lấy
// (POST response không trả ID, cần query DB hoặc dùng admin endpoint)
```

### 4.2. Report user thành công

| Body | `{"target_type": "user", "target_id": "{{userC_id}}", "report_type": "harassment", "reason_detail": "Quấy rối người dùng khác"}` |

**Expected** (`201 Created`)

### 4.3. Report comment thành công

| Body | `{"target_type": "comment", "target_id": "{{comment_id}}", "report_type": "hate_speech", "reason_detail": "Bình luận chứa ngôn từ thù hận"}` |

**Expected** (`201 Created`)

### 4.4. Report với violation_rule_id

| Body | `{"target_type": "post", "target_id": "{{post_id}}", "report_type": "spam", "violation_rule_id": "rule-spam-001", "reason_detail": "Spam theo rule mới"}` |

**Expected** (`201 Created`)

### 4.5. Thiếu target_type

| Body | `{"target_id": "{{post_id}}", "report_type": "spam", "reason_detail": "test"}` |

**Expected** (`400`):
```json
{"error": "target_type là bắt buộc"}
```

### 4.6. target_type không hợp lệ

| Body | `{"target_type": "ad", "target_id": "{{post_id}}", "report_type": "spam", "reason_detail": "test"}` |

**Expected** (`400`):
```json
{"error": "target_type phải là 'user', 'post' hoặc 'comment'"}
```

### 4.7. Thiếu report_type

| Body | `{"target_type": "post", "target_id": "{{post_id}}", "reason_detail": "test"}` |

**Expected** (`400`):
```json
{"error": "report_type là bắt buộc"}
```

### 4.8. Thiếu reason_detail

| Body | `{"target_type": "post", "target_id": "{{post_id}}", "report_type": "spam"}` |

**Expected** (`400`):
```json
{"error": "reason_detail là bắt buộc"}
```

### 4.9. Report trùng — spam prevention

| Body | `{"target_type": "post", "target_id": "{{post_id}}", "report_type": "spam", "reason_detail": "report lần 2"}` |

**Expected** (`400`):
```json
{"error": "bạn đã có báo cáo đang chờ xử lý cho đối tượng này, vui lòng chỉnh sửa báo cáo thay vì tạo mới"}
```

### 4.10. Report bài viết của chính mình

| Headers | `Authorization: Bearer {{token_B}}` |
| Body | `{"target_type": "post", "target_id": "{{post_id}}", "report_type": "spam", "reason_detail": "test"}` |

**Expected** (`400`):
```json
{"error": "không thể báo cáo bài viết của chính mình"}
```

### 4.11. Report chính mình (user)

| Body | `{"target_type": "user", "target_id": "{{userA_id}}", "report_type": "spam", "reason_detail": "test"}` |

**Expected** (`400`):
```json
{"error": "không thể báo cáo chính mình"}
```

### 4.12. Report admin

| Body | `{"target_type": "user", "target_id": "{{admin_id}}", "report_type": "spam", "reason_detail": "test"}` |

**Expected** (`400`):
```json
{"error": "không thể báo cáo quản trị viên hoặc siêu quản trị viên"}
```

### 4.13. Report post không tồn tại

| Body | `{"target_type": "post", "target_id": "nonexistent-id", "report_type": "spam", "reason_detail": "test"}` |

**Expected** (`400`):
```json
{"error": "không tìm thấy bài viết hoặc bài viết không hoạt động"}
```

### 4.14. Report comment không tồn tại

| Body | `{"target_type": "comment", "target_id": "nonexistent-id", "report_type": "spam", "reason_detail": "test"}` |

**Expected** (`400`):
```json
{"error": "không tìm thấy bình luận hoặc bình luận không hoạt động"}
```

### 4.15. Không có token

| Headers | *(không có Authorization)* |

**Expected** (`401 Unauthorized`)

---

## 5. UpdateReport

> **Lưu ý**: Cần lấy `report_id` từ DB sau khi tạo report, vì CreateReport không trả ID.

```sql
-- Chạy trên DB để lấy report pending của user A
SELECT id FROM reports WHERE reporter_id = '<userA_id>' AND status = 'pending' LIMIT 1;
```

### 5.1. Sửa report thành công

| Field | Value |
|-------|-------|
| **Method** | `PUT` |
| **URL** | `{{base_url}}/api/reports/{{report_id}}` |
| **Headers** | `Authorization: Bearer {{token_A}}`, `Content-Type: application/json` |
| **Body** | `{"report_type": "nudity", "reason_detail": "Lý do cập nhật: nội dung khiêu dâm"}` |

**Expected** (`200 OK`):
```json
{"message": "Báo cáo đã được cập nhật thành công"}
```

**Script — Tests** tab:
```javascript
pm.test("Report updated", function () {
    pm.response.to.have.status(200);
    const json = pm.response.json();
    pm.expect(json.message).to.eql("Báo cáo đã được cập nhật thành công");
});
```

### 5.2. Sửa violation_rule_id

| Body | `{"report_type": "spam", "reason_detail": "Spam nghiêm trọng", "violation_rule_id": "rule-spam-002"}` |

**Expected** (`200 OK`)

### 5.3. Bỏ violation_rule_id

| Body | `{"report_type": "spam", "reason_detail": "Spam", "violation_rule_id": null}` |

**Expected** (`200 OK`). DB: `violation_rule_id = NULL`.

### 5.4. Thiếu report_type

| Body | `{"reason_detail": "test"}` |

**Expected** (`400`):
```json
{"error": "report_type là bắt buộc"}
```

### 5.5. Thiếu reason_detail

| Body | `{"report_type": "spam"}` |

**Expected** (`400`):
```json
{"error": "reason_detail là bắt buộc"}
```

### 5.6. Sửa report của user khác

| Headers | `Authorization: Bearer {{token_B}}` |
| Body | `{"report_type": "spam", "reason_detail": "test"}` |

**Expected** (`400`):
```json
{"error": "bạn không có quyền chỉnh sửa báo cáo này"}
```

### 5.7. Sửa report đã processed

> Cần admin xử lý report trước, sau đó reporter cố gắng update.

| Headers | `Authorization: Bearer {{token_A}}` |

**Expected** (`400`):
```json
{"error": "chỉ có thể chỉnh sửa báo cáo đang chờ xử lý"}
```

### 5.8. Report không tồn tại

| URL | `{{base_url}}/api/reports/nonexistent-id` |

**Expected** (`400`):
```json
{"error": "report not found: ..."}
```

### 5.9. Không có token

| Headers | *(không có Authorization)* |

**Expected** (`401 Unauthorized`)

---

## 6. ReviewReport (Admin)

> Cần lấy report_id từ DB: `SELECT id, target_user_id, target_post_id FROM reports WHERE status = 'pending' LIMIT 1;`

### 6.1. Cancel report

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/reports/{{report_id}}/decision` |
| **Headers** | `Authorization: Bearer {{token_admin}}`, `Content-Type: application/json` |
| **Body** | `{"action": "cancel"}` |

**Expected** (`200 OK`):
```json
{"message": "xử lý báo cáo thành công"}
```

**Kiểm tra DB**: Report status = `rejected`.

### 6.2. Hide post

| Body | `{"action": "hide", "reason": "Nội dung vi phạm tiêu chuẩn cộng đồng"}` |

**Expected** (`200 OK`)

**Kiểm tra DB**:
- Report status = `resolved`
- Post status = `hidden`
- Moderation log created (action = `delete`)

### 6.3. Ban user

| Body | `{"action": "ban", "reason": "Spam nghiêm trọng, tái phạm nhiều lần", "duration": "7d"}` |

**Expected** (`200 OK`)

**Kiểm tra DB**:
- User status = `banned`
- Report status = `resolved`
- Ban record created với `expires_at` = 7 ngày từ bây giờ
- Moderation log created (action = `ban`)

### 6.4. Hide thiếu reason

| Body | `{"action": "hide"}` |

**Expected** (`400`):
```json
{"error": "lý do là bắt buộc cho hành động hide hoặc ban"}
```

### 6.5. Ban thiếu reason

| Body | `{"action": "ban"}` |

**Expected** (`400`):
```json
{"error": "lý do là bắt buộc cho hành động hide hoặc ban"}
```

### 6.6. Cancel không cần reason

| Body | `{"action": "cancel", "reason": ""}` |

**Expected** (`200 OK`) — cancel không bắt buộc reason.

### 6.7. Action không hợp lệ

| Body | `{"action": "delete", "reason": "test"}` |

**Expected** (`400`):
```json
{"error": "action không hợp lệ, chỉ chấp nhận cancel, hide hoặc ban"}
```

### 6.8. Report đã processed

| Body | `{"action": "cancel"}` |

**Expected** (`400`):
```json
{"error": "báo cáo đã được xử lý"}
```

### 6.9. Report không tồn tại

| URL | `{{base_url}}/api/admin/reports/nonexistent/decision` |

**Expected** (`400` error)

### 6.10. Không phải superadmin

| Headers | `Authorization: Bearer {{token_A}}` |

**Expected** (`403` Forbidden)

### 6.11. Không có token

| Headers | *(không có Authorization)* |

**Expected** (`401 Unauthorized`)

---

## 7. ListReports (Admin)

### 7.1. List mặc định

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/reports` |
| **Headers** | `Authorization: Bearer {{token_admin}}` |

**Expected** (`200 OK`):
```json
{
    "reports": [
        {
            "id": "...",
            "reporter_id": "...",
            "reporter_username": "user_a",
            "reporter_email": "...",
            "target_user_id": null,
            "target_post_id": "...",
            "target_comment_id": null,
            "report_type": "spam",
            "reason_detail": "...",
            "status": "pending",
            "created_at": "...",
            "target_type": "post"
        }
    ],
    "total": 5,
    "page": 1,
    "page_size": 20
}
```

### 7.2. Filter status

| URL | `{{base_url}}/api/admin/reports?status=pending` |

**Expected**: Chỉ reports có status = `pending`.

### 7.3. Filter targetType

| URL | `{{base_url}}/api/admin/reports?target_type=post` |

**Expected**: Chỉ reports target là post.

### 7.4. Filter keyword

| URL | `{{base_url}}/api/admin/reports?keyword=spam` |

**Expected**: Tìm thấy reports có reason_detail chứa "spam".

### 7.5. Pagination

| URL | `{{base_url}}/api/admin/reports?page=2&page_size=2` |

**Expected**: Offset = 2, limit = 2.

---

## 8. GetReportDetail (Admin)

### 8.1. Lấy detail report post

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/reports/{{report_id}}` |
| **Headers** | `Authorization: Bearer {{token_admin}}` |

**Expected** (`200 OK`):
```json
{
    "id": "...",
    "reporter_id": "...",
    "reporter_username": "...",
    "reporter_email": "...",
    "target_type": "post",
    "target_user_id": null,
    "target_post_id": "...",
    "post_title": "...",
    "post_content": "...",
    "report_type": "spam",
    "reason_detail": "...",
    "status": "reviewed",
    "created_at": "..."
}
```

> **Lưu ý**: Status tự chuyển thành `reviewed` khi lấy detail.

---

## 9. Test Flow — Kịch bản đầy đủ

### 9.1. Happy path: Report → Update → Process

| Step | Action | Token | API | Expected |
|------|--------|-------|-----|----------|
| 1 | Login A + B + C + Admin | — | POST /api/auth/login | tokens |
| 2 | Tạo post của B | B | POST /posts | post_id |
| 3 | A report post B | A | POST /api/reports | 201 |
| 4 | A update report | A | PUT /api/reports/:id | 200 |
| 5 | A report post B lần nữa | A | POST /api/reports | 400 (duplicate) |
| 6 | Admin xem list reports | Admin | GET /api/admin/reports | 200, có report |
| 7 | Admin xem detail | Admin | GET /api/admin/reports/:id | 200, status=reviewed |
| 8 | Admin hide post | Admin | POST /api/admin/reports/:id/decision | 200 |
| 9 | Verify post hidden | — | GET /posts/:id | Post status = hidden |
| 10 | User C report user D | C | POST /api/reports | 201 |
| 11 | Admin ban user D | Admin | POST /api/admin/reports/:id/decision | 200 |
| 12 | Verify user banned | — | DB query | User status = banned |

### 9.2. Cancel flow

| Step | Action | Expected |
|------|--------|----------|
| 1 | User A report user C | 201 |
| 2 | Admin cancel (không cần reason) | 200 |
| 3 | User C không bị ảnh hưởng | Status vẫn active |
| 4 | User A report user C lần nữa | 201 (report mới) |

### 9.3. Anti-spam flow

| Step | Action | Expected |
|------|--------|----------|
| 1 | User A report post B | 201 |
| 2 | User A report post B lần 2 | 400 duplicate |
| 3 | User A update report | 200 |
| 4 | User A report post B lần 3 | 400 duplicate |
| 5 | Admin xử lý report | 200 |
| 6 | User A report post B lần 4 | 201 (report mới, report cũ đã processed) |
| 7 | User B report post B | 400 (self-report) |

---

## 10. Kiểm tra Database

```sql
-- Xem tất cả reports
SELECT id, reporter_id, report_type, target_user_id, target_post_id,
       target_comment_id, violation_rule_id, reason_detail, status,
       created_at, updated_at
FROM reports ORDER BY created_at DESC;

-- Xem reports pending
SELECT id, reporter_id, target_type, report_type, status
FROM reports WHERE status = 'pending';

-- Xem duplicate check: cùng reporter + target + status=pending
SELECT reporter_id, target_user_id, target_post_id, target_comment_id, COUNT(*)
FROM reports WHERE status = 'pending'
GROUP BY reporter_id, target_user_id, target_post_id, target_comment_id
HAVING COUNT(*) > 1;

-- Xem moderation logs từ report
SELECT ml.*, r.report_type, r.reason_detail
FROM moderation_logs ml
JOIN reports r ON r.target_post_id = ml.target_id OR r.target_user_id = ml.target_id
ORDER BY ml.created_at DESC;

-- Xem bans từ report
SELECT b.*, r.report_type
FROM bans b
JOIN reports r ON r.target_user_id = b.user_id
ORDER BY b.created_at DESC;

-- Kiểm tra composite index
SHOW INDEX FROM reports WHERE Key_name = 'idx_reports_reporter_status';

-- Kiểm tra updated_at
SELECT id, created_at, updated_at FROM reports WHERE updated_at IS NOT NULL;
```

---

## 11. Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `401 Unauthorized` | Token thiếu/hết hạn | Login lại, cập nhật biến |
| `403 Forbidden` khi review report | User không có role SuperAdmin | Dùng đúng token admin có role SuperAdmin |
| `400 duplicate` khi report mới | Report pending chưa xử lý | Admin xử lý report cũ trước, hoặc dùng PUT để update |
| Report không trả ID | CreateReport chỉ trả message | Query DB để lấy report_id: `SELECT id FROM reports ORDER BY created_at DESC LIMIT 1` |
| `target_id` không hoạt động | ID sai hoặc record không tồn tại | Kiểm tra DB, đảm bảo ID đúng format UUID |
| `báo cáo đã được xử lý` | Report đã ở trạng thái final (resolved/rejected/reviewed) | Tạo report mới thay vì review lại |
| Cancel vẫn lỗi reason | Kiểm tra body có `action: "cancel"` đúng | Cancel không yêu cầu reason, chỉ hide/ban mới cần |
| Updated_at không set | Report chưa bao giờ được update | PUT report để trigger updated_at |
| `400` self-report khi report user khác | Sai target_id (trùng reporter) | Đảm bảo target_id ≠ reporter_id |

---

## 12. Postman Collection Export

```
Report Module Tests
├── Auth
│   ├── Login User A (reporter)
│   ├── Login User B (post author)
│   ├── Login User C
│   └── Login Admin (SuperAdmin)
├── Setup Test Data
│   ├── Tạo post của B
│   └── Tạo comment trên post
├── CreateReport
│   ├── Report post thành công
│   ├── Report user thành công
│   ├── Report comment thành công
│   ├── Report với violation_rule_id
│   ├── Thiếu target_type → 400
│   ├── target_type không hợp lệ → 400
│   ├── Thiếu report_type → 400
│   ├── Thiếu reason_detail → 400
│   ├── Duplicate (spam prevention) → 400
│   ├── Report bài viết của mình → 400
│   ├── Report chính mình → 400
│   ├── Report admin → 400
│   ├── Report post không tồn tại → 400
│   ├── Report comment không tồn tại → 400
│   └── Không token → 401
├── UpdateReport
│   ├── Sửa report thành công
│   ├── Sửa violation_rule_id
│   ├── Bỏ violation_rule_id
│   ├── Thiếu report_type → 400
│   ├── Thiếu reason_detail → 400
│   ├── Sửa report của user khác → 400
│   ├── Sửa report đã processed → 400
│   ├── Report không tồn tại → 400
│   └── Không token → 401
├── ReviewReport (Admin)
│   ├── Cancel report
│   ├── Hide post
│   ├── Ban user
│   ├── Hide thiếu reason → 400
│   ├── Ban thiếu reason → 400
│   ├── Cancel không cần reason
│   ├── Action không hợp lệ → 400
│   ├── Report đã processed → 400
│   ├── Report không tồn tại → 400
│   ├── Non-admin → 403
│   └── Không token → 401
├── ListReports (Admin)
│   ├── List mặc định
│   ├── Filter status
│   ├── Filter targetType
│   ├── Filter keyword
│   └── Pagination
├── GetReportDetail (Admin)
│   └── Lấy detail report post
└── Full Flow
    ├── Report → Update → Process (happy path)
    ├── Cancel flow
    └── Anti-spam flow
```
