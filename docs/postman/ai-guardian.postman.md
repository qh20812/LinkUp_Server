# Hướng dẫn Test AI Guardian (Moderation) với Postman

## 1. Chuẩn bị Environment

### 1.1. Biến môi trường

Tạo một **Environment** mới trong Postman với các biến sau:

| Variable | Initial Value | Description |
|----------|--------------|-------------|
| `base_url` | `http://localhost:8080` | Server URL |
| `token` | *(empty)* | JWT access token |
| `media_id` | *(empty)* | ID của media vừa upload |
| `admin_token` | *(empty)* | JWT token của admin |
| `community_id` | *(empty)* | ID của community vừa tạo |

### 1.2. Seed database

Chạy seed để có dữ liệu test:

```bash
cd server
go build ./cmd/seed && ./seed.exe
```

Tất cả seed users đều dùng password `Password123!`.

---

## 2. Lấy Token

### 2.1. Login user thường

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/auth/login` |
| **Headers** | `Content-Type: application/json` |
| **Body** | `{"email": "seed_user_A@example.com", "password": "Password123!"}` |

**Script — Tests** tab:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('token', json.data.access_token);
```

### 2.2. Login admin

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/auth/login` |
| **Headers** | `Content-Type: application/json` |
| **Body** | `{"email": "admin@linkup.com", "password": "Password123!"}` |

**Script — Tests** tab:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('admin_token', json.data.access_token);
```

> Email admin phụ thuộc vào seed data. Thay `admin@linkup.com` bằng email admin có thật.

---

## 3. Upload Media + AI Moderation

Upload file ảnh lên server. Hệ thống sẽ tự động upload lên Cloudinary kèm moderation `aws_rek`.

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/media/upload` |
| **Headers** | `Authorization: Bearer {{token}}` |
| **Body** | `form-data` — key `file` (type File), chọn ảnh bất kỳ |

**Script — Tests** tab:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('media_id', json.data.id);
```

**Response mẫu** (approved):
```json
{
    "id": "uuid-media-id",
    "user_id": "uuid-user",
    "file_uri": "https://res.cloudinary.com/.../image/upload/v1/abc123.jpg",
    "file_type": "image/jpeg",
    "file_size": 102400,
    "status": "approved",
    "created_at": "2026-07-11T10:00:00Z"
}
```

**Các status có thể nhận:**

| Status | Ý nghĩa | Hành động tiếp theo |
|--------|---------|---------------------|
| `approved` | Ảnh hợp lệ, được dùng ngay | Không cần làm gì |
| `flagged` | Ảnh đang chờ admin kiểm duyệt | Có thể dùng tạm, admin sẽ review sau |
| `rejected` | Ảnh vi phạm tiêu chuẩn | Không được dùng (lỗi hiển thị) |
| `pending` | Moderation chưa chạy (fallback) | Giống flagged |

---

## 4. Tạo Community với Avatar + Background

### 4.1. Tạo community kèm avatar

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/communities` |
| **Headers** | `Authorization: Bearer {{token}}` |
| **Body** | `form-data` — `name` (text), `description` (text), `avatar` (file) |

**Script — Tests** tab:
```javascript
const json = pm.response.json();
pm.collectionVariables.set('community_id', json.community_id);
```

**Response mẫu** (avatar approved):
```json
{
    "message": "Tạo cộng đồng thành công!",
    "community_id": "uuid-community",
    "auto_approve": false,
    "default_group_chat": {
        "id": "uuid-chat",
        "name": "Tên Cộng Đồng"
    }
}
```

**Note:** Nếu avatar bị `rejected`, API trả lỗi `400 "Ảnh đại diện vi phạm tiêu chuẩn cộng đồng"`.

### 4.2. Set background community

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/communities/{{community_id}}/background` |
| **Headers** | `Authorization: Bearer {{token}}` |
| **Body** | `form-data` — `background` (file) |

**Response mẫu** (success):
```json
{
    "message": "Cập nhật background cộng đồng thành công!"
}
```

**Note:** Nếu background bị `rejected`, API trả lỗi `400 "ảnh background vi phạm tiêu chuẩn cộng đồng"`.

---

## 5. Admin — Danh sách Media

### 5.1. Danh sách media flagged (mặc định)

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/media/flagged` |
| **Headers** | `Authorization: Bearer {{admin_token}}` |

**Response mẫu**:
```json
{
    "items": [
        {
            "id": "uuid-media",
            "user_id": "uuid-user",
            "file_uri": "https://res.cloudinary.com/...",
            "file_type": "image/jpeg",
            "file_size": 102400,
            "status": "flagged",
            "created_at": "2026-07-11T10:00:00Z"
        }
    ],
    "total": 1,
    "page": 1
}
```

### 5.2. Danh sách media theo status

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/admin/media/flagged?status=rejected` |
| **Headers** | `Authorization: Bearer {{admin_token}}` |

| Query Param | Giá trị hợp lệ | Mô tả |
|-------------|---------------|-------|
| `status` | `flagged`, `rejected` | Lọc theo status (mặc định: `flagged`) |
| `page` | `1..N` | Số trang |
| `page_size` | `1..100` | Số item mỗi trang (mặc định: 20) |

---

## 6. Admin — Review Media

### 6.1. Approve media

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/media/{{media_id}}/review` |
| **Headers** | `Authorization: Bearer {{admin_token}}`, `Content-Type: application/json` |
| **Body** | `{"action": "approve", "reason": "Nội dung phù hợp"}` |

**Response**:
```json
{
    "message": "Xử lý media thành công"
}
```

### 6.2. Reject media

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/media/{{media_id}}/review` |
| **Headers** | `Authorization: Bearer {{admin_token}}`, `Content-Type: application/json` |
| **Body** | `{"action": "reject", "reason": "Vi phạm tiêu chuẩn cộng đồng"}` |

**Kết quả:**
- Status media chuyển thành `rejected`
- File trên Cloudinary bị xoá (Destroy)
- `FileURI` trong DB được clear thành `""`
- User nhận notification "đã bị admin từ chối"
- Moderation log được ghi

### 6.3. Action không hợp lệ

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/media/{{media_id}}/review` |
| **Body** | `{"action": "ban", "reason": "test"}` |

**Response**: `400 "hành động không hợp lệ: ban"`

---

## 7. Admin — Cleanup Rejected Media

Xoá toàn bộ media bị AI reject quá 7 ngày.

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/admin/media/cleanup-rejected` |
| **Headers** | `Authorization: Bearer {{admin_token}}` |

**Response mẫu**:
```json
{
    "message": "Dọn dẹp media thành công",
    "cleaned": 5
}
```

---

## 8. Notification User

Sau khi admin review, user nhận notification. Kiểm tra bằng:

| Field | Value |
|-------|-------|
| **Method** | `GET` |
| **URL** | `{{base_url}}/api/notifications` |
| **Headers** | `Authorization: Bearer {{token}}` |

**Các notification messages:**

| Action | Message |
|--------|---------|
| Upload → AI auto-approved | `Ảnh/video của bạn đã được hệ thống tự động duyệt.` |
| Upload → AI rejected | `Ảnh/video của bạn bị từ chối do vi phạm tiêu chuẩn cộng đồng.` |
| Upload → AI flagged/pending | `Ảnh/video của bạn đang chờ admin kiểm duyệt.` |
| Admin approve | `Ảnh/video của bạn đã được admin phê duyệt.` |
| Admin reject | `Ảnh/video của bạn đã bị admin từ chối: <lý do>` |
