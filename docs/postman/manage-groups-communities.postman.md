# Hướng dẫn Test Admin Group & Community Management với Postman

## 1. Chuẩn bị Environment

### 1.1. Biến môi trường

Tạo một **Environment** mới trong Postman với các biến sau:

| Variable | Initial Value | Description |
|----------|--------------|-------------|
| `base_url` | `http://localhost:8080` | Server URL |
| `token_superadmin` | *(empty)* | JWT token superadmin |
| `token_admin` | *(empty)* | JWT token admin |
| `token_creator` | *(empty)* | JWT token user làm chủ group/community test |
| `token_member` | *(empty)* | JWT token thành viên thường |
| `superadmin_id` | *(empty)* | UUID của superadmin |
| `admin_id` | *(empty)* | UUID của admin |
| `creator_id` | *(empty)* | UUID của chủ sở hữu |
| `member_id` | *(empty)* | UUID của thành viên thường |
| `group_chat_id` | *(empty)* | ID của group chat để test |
| `community_id` | *(empty)* | ID của community để test |
| `log_id` | *(empty)* | ID của moderation log |
| `ban_user_id` | *(empty)* | User ID bị ban (để test auto-transfer) |

### 1.2. Seed database

```bash
cd server
go build ./cmd/seed && ./seed.exe
```

Tài khoản mặc định (xem `cmd/seed/main.go` — tất cả đều dùng `Password123!`).

---

## 2. Lấy Token

### 2.1. Login superadmin

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/auth/login` |
| **Headers** | `Content-Type: application/json` |
| **Body** (raw JSON) | `{"email": "superadmin@linkup.com", "password": "Password123!"}` |

**Script — Tests** tab:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_superadmin', json.data.access_token);
pm.collectionVariables.set('superadmin_id', json.data.user.id);
```

### 2.2. Login admin

| URL | Body |
|-----|------|
| `POST {{base_url}}/api/auth/login` | `{"email": "admin@linkup.com", "password": "Password123!"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_admin', json.data.access_token);
pm.collectionVariables.set('admin_id', json.data.user.id);
```

### 2.3. Login creator user

| URL | Body |
|-----|------|
| `POST {{base_url}}/api/auth/login` | `{"email": "seed_user_A@example.com", "password": "Password123!"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_creator', json.data.access_token);
pm.collectionVariables.set('creator_id', json.data.user.id);
```

### 2.4. Login member user

| URL | Body |
|-----|------|
| `POST {{base_url}}/api/auth/login` | `{"email": "seed_user_B@example.com", "password": "Password123!"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token_member', json.data.access_token);
pm.collectionVariables.set('member_id', json.data.user.id);
```

---

## 3. Setup — Tạo Group Chat + Community

### 3.1. Tạo group chat (creator)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/group-chats/create` |
| **Headers** | `Authorization: Bearer {{token_creator}}`, `Content-Type: application/json` |
| **Body** | `{"name": "Test Group", "member_ids": ["{{member_id}}"]}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('group_chat_id', json.data.id);
```

### 3.2. Tạo community (creator)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/communities/create` |
| **Headers** | `Authorization: Bearer {{token_creator}}`, `Content-Type: application/json` |
| **Body** | `{"name": "Test Community", "description": "A test community", "privacy": "public"}` |

**Script**:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('community_id', json.data.id);
```

---

## 4. Admin — List Groups

### 4.1. Thành công (superadmin)

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/groups` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}` |

**Expected** (`200 OK`):
```json
{
    "groups": [
        {
            "id": "...",
            "name": "Test Group",
            "creator_id": "...",
            "creator_name": "...",
            "member_count": 2,
            "status": "active",
            "created_at": "..."
        }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
}
```

### 4.2. Filter theo keyword

| URL | `{{base_url}}/api/admin/groups?keyword=Test` |

### 4.3. Filter theo status

| URL | `{{base_url}}/api/admin/groups?status=archived` |

### 4.4. Phân trang

| URL | `{{base_url}}/api/admin/groups?page=1&page_size=5` |

### 4.5. Admin (không phải superadmin) vẫn có quyền

| Headers | `Authorization: Bearer {{token_admin}}` |

### 4.6. Không có token → 401

| Headers | *(none)* |

**Expected** (`401 Unauthorized`).

---

## 5. Admin — Get Group Detail

### 5.1. Thành công

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/groups/{{group_chat_id}}` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}` |

**Expected** (`200 OK`):
```json
{
    "id": "...",
    "name": "Test Group",
    "avatar_uri": "",
    "creator_id": "...",
    "creator_name": "...",
    "type": "group",
    "status": "active",
    "member_count": 2,
    "members": [
        {"user_id": "...", "display_name": "...", "avatar_uri": "", "role": "CHAT_ADMIN"},
        {"user_id": "...", "display_name": "...", "avatar_uri": "", "role": "CHAT_MEMBER"}
    ],
    "created_at": "..."
}
```

### 5.2. Group không tồn tại

| URL | `{{base_url}}/api/admin/groups/nonexistent-id` |

**Expected** (`400 Bad Request`):
```json
{"error": "group chat không tồn tại"}
```

---

## 6. Admin — List Group Members

### 6.1. Thành công

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/groups/{{group_chat_id}}/members` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}` |

**Expected** (`200 OK`):
```json
{"members": [
    {"user_id": "...", "display_name": "...", "avatar_uri": "", "role": "CHAT_ADMIN"},
    {"user_id": "...", "display_name": "...", "avatar_uri": "", "role": "CHAT_MEMBER"}
]}
```

### 6.2. Group không tồn tại

| URL | `{{base_url}}/api/admin/groups/nonexistent/members` |

**Expected** (`400`): `"group chat không tồn tại"`.

---

## 7. Admin — Get Group Moderation Logs

### 7.1. Thành công (chưa có log)

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/groups/{{group_chat_id}}/logs` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}` |

**Expected** (`200 OK`):
```json
{"logs": [], "total": 0, "page": 1, "page_size": 20}
```

### 7.2. Phân trang

| URL | `{{base_url}}/api/admin/groups/{{group_chat_id}}/logs?page=1&page_size=5` |

---

## 8. Admin — Moderate Group (Hide / Unhide / Archive / Warn)

### 8.1. Hide group

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/groups/{{group_chat_id}}/hide` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}`, `Content-Type: application/json` |
| **Body** | `{"reason": "Vi phạm nội quy"}` |

**Expected** (`200 OK`):
```json
{"message": "Ẩn group thành công"}
```

### 8.2. Unhide group

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/groups/{{group_chat_id}}/unhide` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}` |

**Expected** (`200 OK`):
```json
{"message": "Bỏ ẩn group thành công"}
```

### 8.3. Archive (đình chỉ) group

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/groups/{{group_chat_id}}/archive` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}`, `Content-Type: application/json` |
| **Body** | `{"reason": "Spam liên tục"}` |

**Expected** (`200 OK`):
```json
{"message": "Đình chỉ group thành công"}
```

### 8.4. Warn group

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/groups/{{group_chat_id}}/warn` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}`, `Content-Type: application/json` |
| **Body** | `{"reason": "Cảnh báo lần 1", "message": "Vui lòng tuân thủ nội quy"}` |

**Expected** (`200 OK`):
```json
{"message": "Cảnh báo group thành công"}
```

### 8.5. State transition — archived → không thể hide/unhide

Sau khi archive (8.3), thử hide:
**Expected** (`400`): `"không thể thao tác trên group đã bị đình chỉ"`

### 8.6. Thiếu reason (hide/archive)

**Body**: `{}`

**Expected** (`400`): validation error.

### 8.7. Admin (không superadmin) có quyền

| Headers | `Authorization: Bearer {{token_admin}}` |

---

## 9. Admin — List Communities

### 9.1. Thành công (superadmin)

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/communities` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}` |

**Expected** (`200 OK`):
```json
{
    "communities": [
        {
            "id": "...",
            "name": "Test Community",
            "creator_id": "...",
            "creator_name": "...",
            "member_count": 1,
            "privacy": "public",
            "status": "active",
            "created_at": "..."
        }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
}
```

### 9.2. Filter theo keyword

| URL | `{{base_url}}/api/admin/communities?keyword=Test` |

### 9.3. Filter theo status

| URL | `{{base_url}}/api/admin/communities?status=archived` |

### 9.4. Filter theo privacy

| URL | `{{base_url}}/api/admin/communities?privacy=public` |

### 9.5. Phân trang

| URL | `{{base_url}}/api/admin/communities?page=1&page_size=5` |

### 9.6. Admin (không superadmin) có quyền

| Headers | `Authorization: Bearer {{token_admin}}` |

---

## 10. Admin — Get Community Detail

### 10.1. Thành công

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/communities/{{community_id}}` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}` |

**Expected** (`200 OK`):
```json
{
    "id": "...",
    "name": "Test Community",
    "description": "A test community",
    "creator_id": "...",
    "creator_name": "...",
    "privacy": "public",
    "status": "active",
    "auto_approve": false,
    "member_count": 1,
    "members": [
        {"user_id": "...", "display_name": "...", "avatar_uri": "", "role": "COMMUNITY_ADMIN"}
    ],
    "created_at": "..."
}
```

### 10.2. Community không tồn tại

| URL | `{{base_url}}/api/admin/communities/nonexistent` |

**Expected** (`400`): `"cộng đồng không tồn tại"`.

---

## 11. Admin — List Community Members

### 11.1. Thành công

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/communities/{{community_id}}/members` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}` |

**Expected** (`200 OK`):
```json
{"members": [{"user_id": "...", "display_name": "...", "avatar_uri": "", "role": "COMMUNITY_ADMIN"}]}
```

### 11.2. Community không tồn tại

| URL | `{{base_url}}/api/admin/communities/nonexistent/members` |

---

## 12. Admin — Get Community Moderation Logs

### 12.1. Thành công (chưa có log)

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/communities/{{community_id}}/logs` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}` |

**Expected** (`200 OK`): `{"logs": [], "total": 0, ...}`

### 12.2. Phân trang

| URL | `{{base_url}}/api/admin/communities/{{community_id}}/logs?page=1&page_size=5` |

---

## 13. Admin — Moderate Community (Hide / Unhide / Archive / Warn)

### 13.1. Hide community

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/communities/{{community_id}}/hide` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}`, `Content-Type: application/json` |
| **Body** | `{"reason": "Nội dung không phù hợp"}` |

**Expected** (`200 OK`):
```json
{"message": "Ẩn cộng đồng thành công"}
```

### 13.2. Unhide community

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/communities/{{community_id}}/unhide` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}` |

**Expected** (`200 OK`):
```json
{"message": "Bỏ ẩn cộng đồng thành công"}
```

### 13.3. Archive (đình chỉ) community

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/communities/{{community_id}}/archive` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}`, `Content-Type: application/json` |
| **Body** | `{"reason": "Vi phạm nhiều lần"}` |

**Expected** (`200 OK`):
```json
{"message": "Đình chỉ cộng đồng thành công"}
```

### 13.4. Warn community

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/communities/{{community_id}}/warn` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}`, `Content-Type: application/json` |
| **Body** | `{"reason": "Cảnh báo", "message": "Vui lòng tuân thủ quy tắc cộng đồng"}` |

**Expected** (`200 OK`):
```json
{"message": "Cảnh báo cộng đồng thành công"}
```

### 13.5. State transition — archived → không thể hide/unhide

Sau khi archive (13.3), thử hide:
**Expected** (`400`): `"không thể thao tác trên cộng đồng đã bị đình chỉ"`

---

## 14. Delete Group (Phase 4)

### 14.1. Thành công — group chỉ còn 1 member (creator)

> **Prerequisite**: group chat chỉ có 1 thành viên (creator là người duy nhất).

| Field | Value |
|-------|-------|
| **Method** | `DELETE` |
| **URL** | `{{base_url}}/api/admin/groups/{{group_chat_id}}` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}`, `Content-Type: application/json` |
| **Body** | `{"reason": "Xoá group vi phạm"}` |

**Expected** (`200 OK`):
```json
{"message": "Xóa group chat thành công"}
```

### 14.2. Lỗi — group còn nhiều hơn 1 member

> **Prerequisite**: group chat có ≥ 2 thành viên.

**Expected** (`400`):
```json
{"error": "không thể xóa group chat còn thành viên khác; hãy chuyển quyền sở hữu trước"}
```

### 14.3. Group không tồn tại

| URL | `{{base_url}}/api/admin/groups/nonexistent` |

**Expected** (`400`): `"không tìm thấy chat"`.

---

## 15. Delete Community (Phase 4)

### 15.1. Thành công — community chỉ còn 1 member (creator)

> **Prerequisite**: community chỉ có 1 thành viên là creator.

| Field | Value |
|-------|-------|
| **Method** | `DELETE` |
| **URL** | `{{base_url}}/api/admin/communities/{{community_id}}` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}`, `Content-Type: application/json` |
| **Body** | `{"reason": "Xoá cộng đồng vi phạm"}` |

**Expected** (`200 OK`):
```json
{"message": "Xóa cộng đồng thành công"}
```

### 15.2. Lỗi — community còn nhiều hơn 1 member

> **Prerequisite**: community có ≥ 2 thành viên.

**Expected** (`400`):
```json
{"error": "không thể xóa cộng đồng còn thành viên khác; hãy chuyển quyền sở hữu trước"}
```

### 15.3. Community không tồn tại

| URL | `{{base_url}}/api/admin/communities/nonexistent` |

**Expected** (`400`): `"cộng đồng không tồn tại"`.

---

## 16. Ban User — Auto-transfer Ownership (Phase 4)

### 16.1. Tạo data test

1. Login làm user A (creator): tạo 1 community, 1 group chat
2. Mời user B (member) vào cả 2
3. Login superadmin → ban user A

### 16.2. Ban user A (creator) — cộng đồng + group chat tự động chuyển

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/users/{{creator_id}}/ban` |
| **Headers** | `Authorization: Bearer {{token_superadmin}}`, `Content-Type: application/json` |
| **Body** | `{"reason": "Spam", "duration": "permanent"}` |

**Expected** (`200 OK`):
```json
{"message": "ban user thành công"}
```

**Verification**:
- DB `users` → status = `"banned"`
- DB `communities` → `creator_id` đã đổi thành member (hoặc admin) khác
- DB `chats` → `creator_id` đã đổi thành participant khác

### 16.3. Ban user không có community/group

Ban user B (không phải creator của group/community nào):

| URL | `{{base_url}}/api/admin/users/{{member_id}}/ban` |
| Body | `{"reason": "Vi phạm", "duration": "7d"}` |

**Expected** (`200`): Ban thành công, không ảnh hưởng tới các group/community khác.

### 16.4. Ban user đã bị ban

Gọi lại 16.2 một lần nữa:
**Expected** (`400`): `"người dùng đã bị ban"`.

### 16.5. Lỗi — không phải superadmin

| Headers | `Authorization: Bearer {{token_creator}}` |

**Expected** (`400`): `"chỉ có superadmin mới có được phép"`.

---

## 17. Transfer Ownership (Phase 2)

### 17.1. Chuyển quyền sở hữu community

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/communities/{{community_id}}/transfer-ownership` |
| **Headers** | `Authorization: Bearer {{token_creator}}`, `Content-Type: application/json` |
| **Body** | `{"target_user_id": "{{member_id}}", "keep_admin": false}` |

**Expected** (`200 OK`):
```json
{"message": "chuyển quyền sở hữu cộng đồng thành công"}
```

### 17.2. Chuyển quyền sở hữu group chat

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/group-chats/{{group_chat_id}}/transfer-ownership` |
| **Headers** | `Authorization: Bearer {{token_creator}}`, `Content-Type: application/json` |
| **Body** | `{"target_user_id": "{{member_id}}", "keep_admin": false}` |

**Expected** (`200 OK`):
```json
{"message": "chuyển quyền sở hữu nhóm chat thành công"}
```

### 17.3. Chuyển admin (không phải creator) group chat

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/group-chats/{{group_chat_id}}/transfer-admin` |
| **Headers** | `Authorization: Bearer {{token_creator}}`, `Content-Type: application/json` |
| **Body** | `{"target_user_id": "{{member_id}}"}` |

**Expected** (`200 OK`).

---

## 18. Kiểm tra Database

```sql
-- Xem groups
SELECT id, name, creator_id, status, type FROM chats WHERE type = 'group';

-- Xem communities
SELECT id, name, creator_id, status FROM communities;

-- Xem moderation logs
SELECT id, moderator_id, action, target_type, target_id, reason, created_at
FROM moderation_logs ORDER BY created_at DESC;

-- Kiểm tra user status
SELECT id, username, email, status FROM users WHERE status != 'active';

-- Kiểm tra ownership sau transfer
SELECT id, name, creator_id FROM communities WHERE id = '<community_id>';
SELECT id, name, creator_id FROM chats WHERE id = '<group_chat_id>';
```

---

## 19. Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `401 Unauthorized` | Token thiếu/hết hạn | Login lại, cập nhật biến |
| `400 "không có quyền truy cập"` | Token không hợp lệ hoặc thiếu role | Dùng superadmin token |
| `"chỉ có superadmin mới có được phép"` | User không phải superadmin | Dùng `token_superadmin` |
| `"cộng đồng không tồn tại"` | Sai community_id | Kiểm tra seed data |
| `"không tìm thấy chat"` | Sai chat_id | Kiểm tra seed data |
| `"không thể thao tác trên group/cộng đồng đã bị đình chỉ"` | Trạng thái archived | Unarchive không hỗ trợ |
| `"group chat không tồn tại"` | Sai chatID hoặc chat type != group | Kiểm tra `type` trong DB |
| Không tìm thấy user seed | Seed chưa chạy | `go build ./cmd/seed && ./seed.exe` |
| `403 Forbidden` | Thiếu role requirement | Kiểm tra `RequireRoles` middleware |

---

## 20. Postman Collection Export

```
Admin Group & Community Management
├── Auth
│   ├── Login superadmin
│   ├── Login admin
│   ├── Login creator
│   └── Login member
├── Setup
│   ├── Tạo group chat (creator)
│   └── Tạo community (creator)
├── Group Admin
│   ├── List groups
│   ├── Get group detail
│   ├── List group members
│   ├── Get group moderation logs
│   ├── Hide group
│   ├── Unhide group
│   ├── Archive group
│   ├── Warn group
│   └── Delete group
├── Community Admin
│   ├── List communities
│   ├── Get community detail
│   ├── List community members
│   ├── Get community moderation logs
│   ├── Hide community
│   ├── Unhide community
│   ├── Archive community
│   ├── Warn community
│   └── Delete community
├── Ban User (auto-transfer)
│   ├── Ban creator (có community/group)
│   ├── Ban member (không có)
│   ├── Ban đã bị ban → lỗi
│   └── Không phải superadmin → lỗi
├── Transfer Ownership
│   ├── Chuyển ownership community
│   ├── Chuyển ownership group chat
│   └── Chuyển admin group chat (không phải creator)
└── Error Cases
    ├── Group/community không tồn tại
    ├── Archived → không thể hide/unhide
    ├── Thiếu reason → 400
    └── Không token → 401
```
