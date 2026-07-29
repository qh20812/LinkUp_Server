# Hướng dẫn Test Email Verification với Postman

## 1. Chuẩn bị Environment

### 1.1. Biến môi trường

| Variable | Initial Value | Mô tả |
|----------|--------------|-------|
| `base_url` | `http://localhost:8080` | Server URL |
| `token` | *(empty)* | JWT token sau khi login |
| `verify_token` | *(empty)* | Token xác thực email (lấy từ DB) |

### 1.2. Yêu cầu

- Server đang chạy với Gmail SMTP được cấu hình (hoặc test bằng cách đọc token từ DB).
- `require_email_verify` được bật trong system_configs.

### 1.3. Lấy verify token từ DB (khi test không cần gửi email thật)

```sql
SELECT token FROM email_verification_tokens
WHERE used_at IS NULL AND expires_at > NOW()
ORDER BY created_at DESC LIMIT 1;
```

---

## 2. Test Cases

### 2.1. Register khi require_email_verify = true

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/auth/register` |
| **Headers** | `Content-Type: application/json` |

**Body:**
```json
{
  "display_name": "Nguyễn Văn A",
  "email": "nguyenvana@example.com",
  "password": "Password123!"
}
```

**Expected Response (201):**
```json
{
  "user": {
    "id": "uuid...",
    "username": "nguyenvana...",
    "email": "nguyenvana@example.com",
    "status": "active",
    "created_at": "2026-07-28T..."
  },
  "tokens": {
    "access_token": "",
    "refresh_token": "",
    "token_type": "Bearer",
    "expires_in": 0,
    "refresh_ttl_in": 604800
  },
  "verify_email": true
}
```

✅ **Kiểm tra:** `verify_email: true` — frontend hiển thị thông báo "Vui lòng kiểm tra email để xác thực tài khoản".

---

### 2.2. Login khi chưa xác thực email

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/auth/login` |
| **Headers** | `Content-Type: application/json` |

**Body:**
```json
{
  "email": "nguyenvana@example.com",
  "password": "Password123!"
}
```

**Expected Response (401):**
```json
{
  "error": "vui lòng xác thực email trước khi đăng nhập"
}
```

✅ **Kiểm tra:** User bị chặn đăng nhập với message tiếng Việt.

---

### 2.3. Xác thực email thành công

Lấy token từ DB (câu SQL ở mục 1.3) → gán vào biến `verify_token`.

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/auth/verify-email` |
| **Headers** | `Content-Type: application/json` |

**Body:**
```json
{
  "token": "{{verify_token}}"
}
```

**Expected Response (200):**
```json
{
  "message": "Xác thực email thành công",
  "verified": true,
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "role": "USER"
}
```

✅ **Kiểm tra:** `verified: true` + trả access/refresh token → user tự động đăng nhập.

Gán `access_token` vào biến `token` để dùng cho các request tiếp theo.

---

### 2.4. Login sau khi đã xác thực

Lặp lại **2.2** → **Expected Response (200):** trả tokens bình thường, không còn bị chặn.

---

### 2.5. Token không hợp lệ (sai token)

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/auth/verify-email` |
| **Headers** | `Content-Type: application/json` |

**Body:**
```json
{
  "token": "invalid_token_123"
}
```

**Expected Response (400):**
```json
{
  "message": "Token xác thực không hợp lệ",
  "verified": false
}
```

---

### 2.6. Token đã hết hạn

Dùng token đã hết hạn (hơn 1 giờ) hoặc giả mạo:

**Expected Response (400):**
```json
{
  "message": "Token xác thực đã hết hạn",
  "verified": false
}
```

---

### 2.7. Token đã được sử dụng

Dùng lại token đã verify ở **2.3**:

**Expected Response (400):**
```json
{
  "message": "Token xác thực đã được sử dụng",
  "verified": false
}
```

---

### 2.8. Gửi lại email xác thực

| Field | Value |
|-------|-------|
| **Method** | `POST` |
| **URL** | `{{base_url}}/api/auth/resend-verification` |
| **Headers** | `Content-Type: application/json` |

**Body:**
```json
{
  "email": "nguyenvana@example.com"
}
```

**Expected Response (200):**
```json
{
  "message": "Nếu email tồn tại, bạn sẽ nhận được hướng dẫn xác thực"
}
```

Lưu ý: Token cũ bị xóa, token mới được tạo và gửi qua email.

---

### 2.9. Resend khi email đã được xác thực

Gọi **2.8** sau khi verify thành công → response vẫn là 200 với message:
```json
{
  "message": "Email đã được xác thực trước đó"
}
```

---

### 2.10. Register khi require_email_verify = false

Set `require_email_verify` về `false` trong DB hoặc qua admin settings:

```sql
UPDATE system_configs SET value = 'false' WHERE `key` = 'require_email_verify';
```

Register như **2.1** → **Expected Response (201):** trả tokens đầy đủ, `verify_email: false` (hoặc omitted).

---

## 3. Kiểm tra database

### 3.1. Xem token xác thực

```sql
SELECT * FROM email_verification_tokens ORDER BY created_at DESC;
```

### 3.2. Kiểm tra trạng thái verified của user

```sql
SELECT id, email, email_verified_at FROM users WHERE email = 'nguyenvana@example.com';
```

- `NULL` → chưa verify
- có giá trị datetime → đã verify

### 3.3. Xóa token cũ (khi cần test lại)

```sql
DELETE FROM email_verification_tokens WHERE user_id = (SELECT id FROM users WHERE email = 'nguyenvana@example.com');
UPDATE users SET email_verified_at = NULL WHERE email = 'nguyenvana@example.com';
```
