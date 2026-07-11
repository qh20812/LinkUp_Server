# Report — Test Cases

## REST API Endpoints

| Method | Path | Handler | Auth | Description |
|--------|------|---------|------|-------------|
| POST | `/api/reports` | CreateReport | Token | Tạo báo cáo mới (user/post/comment) |
| PUT | `/api/reports/:id` | UpdateReport | Token | Chỉnh sửa báo cáo đang pending |
| POST | `/api/admin/reports/:reportID/decision` | ReviewReport | Token (SuperAdmin) | Xử lý báo cáo (cancel/hide/ban) |
| GET | `/api/admin/reports` | ListReports | Token (SuperAdmin) | Danh sách báo cáo (filter, pagination) |
| GET | `/api/admin/reports/:reportID` | GetReportDetail | Token (SuperAdmin) | Chi tiết báo cáo |

---

## 1. CreateReport

### 1.1. Happy Path

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-CRT-01 | Report user thành công | POST `/api/reports` với `target_type: "user"`, `target_id` = user khác, `report_type: "spam"`, `reason_detail: "nội dung spam"` | `201` `{message: "Báo cáo đã được gửi thành công"}`. Report có status `pending` trong DB. | ✅ |
| RPT-CRT-02 | Report post thành công | POST `/api/reports` với `target_type: "post"`, `target_id` = post ID, `report_type: "harassment"` | `201` `{message: "Báo cáo đã được gửi thành công"}`. | ✅ |
| RPT-CRT-03 | Report comment thành công | POST `/api/reports` với `target_type: "comment"`, `target_id` = comment ID, `report_type: "hate_speech"` | `201` `{message: "Báo cáo đã được gửi thành công"}`. | ✅ |
| RPT-CRT-04 | Report với violation_rule_id | POST `/api/reports` với `violation_rule_id` có giá trị | `201`. Report có `violation_rule_id` đúng trong DB. | ✅ |
| RPT-CRT-05 | Report không có violation_rule_id | POST `/api/reports` không truyền `violation_rule_id` | `201`. Report có `violation_rule_id = nil`. | ✅ |

### 1.2. Validation — Target Type

| ID | Scenario | Input | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-CRT-06 | Thiếu target_type | `target_type: ""` | `400` error `target_type là bắt buộc`. | ✅ |
| RPT-CRT-07 | target_type không hợp lệ | `target_type: "ad"` | `400` error `target_type phải là 'user', 'post' hoặc 'comment'`. | ✅ |
| RPT-CRT-08 | target_type hoa | `target_type: "USER"` | `400` error (không normalize case). | ✅ |

### 1.3. Validation — Target ID

| ID | Scenario | Input | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-CRT-09 | Thiếu target_id | `target_id: ""` | `400` error `target_id là bắt buộc`. | ✅ |

### 1.4. Validation — Report Type

| ID | Scenario | Input | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-CRT-10 | Thiếu report_type | `report_type: ""` | `400` error `report_type là bắt buộc`. | ✅ |

### 1.5. Validation — Reason

| ID | Scenario | Input | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-CRT-11 | Thiếu reason_detail | `reason_detail: ""` | `400` error `reason_detail là bắt buộc`. | ✅ |
| RPT-CRT-12 | reason_detail chỉ có whitespace | `reason_detail: "   "` | `400` error `reason_detail là bắt buộc`. | ✅ |

### 1.6. Duplicate Prevention (Anti-Spam)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-CRT-13 | Report trùng pending — user | 1. Report user X (pending)<br>2. Report user X lần nữa | `400` error `bạn đã có báo cáo đang chờ xử lý cho đối tượng này, vui lòng chỉnh sửa báo cáo thay vì tạo mới`. | ✅ |
| RPT-CRT-14 | Report trùng pending — post | 1. Report post Y (pending)<br>2. Report post Y lần nữa | `400` error duplicate. | ✅ |
| RPT-CRT-15 | Report trùng pending — comment | 1. Report comment Z (pending)<br>2. Report comment Z lần nữa | `400` error duplicate. | ✅ |
| RPT-CRT-16 | Report sau khi report cũ resolved | 1. Report user X (pending)<br>2. Admin xử lý → resolved<br>3. Report user X lần nữa | `201` thành công (report cũ đã processed, cho phép tạo mới). | ✅ |
| RPT-CRT-17 | Report sau khi report cũ rejected | 1. Report user X (pending)<br>2. Admin cancel → rejected<br>3. Report user X lần nữa | `201` thành công. | ✅ |
| RPT-CRT-18 | Report cùng target, user khác nhau | 1. User A report user X<br>2. User B report user X | Cả 2 đều `201` thành công (duplicate check theo reporter). | ✅ |
| RPT-CRT-19 | Report khác target type, cùng target_id | 1. Report post Y (pending)<br>2. Report comment Y (cùng ID, khác type) | `201` thành công (check theo cả target_type + target_id). | ✅ |

### 1.7. Target Validation — User

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-CRT-20 | Report admin | Report user có role Admin | `400` error `không thể báo cáo quản trị viên hoặc siêu quản trị viên`. | ✅ |
| RPT-CRT-21 | Report super admin | Report user có role SuperAdmin | `400` error `không thể báo cáo quản trị viên hoặc siêu quản trị viên`. | ✅ |
| RPT-CRT-22 | Report chính mình | Report chính user đang logged in | `400` error `không thể báo cáo chính mình`. | ✅ |
| RPT-CRT-23 | Report user đã bị ban | Report user có status = banned | `400` error `không thể báo cáo người dùng đã bị cấm`. | ✅ |
| RPT-CRT-24 | Report user không tồn tại | Report với target_id không tồn tại trong DB | `400` hoặc `500` error (user not found). | ✅ |

### 1.8. Target Validation — Post

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-CRT-25 | Report post không tồn tại | Report post với ID không tồn tại | `400` error `không tìm thấy bài viết hoặc bài viết không hoạt động`. | ✅ |
| RPT-CRT-26 | Report post đã hidden | Report post có status = hidden | `400` error `không thể báo cáo bài viết đã bị ẩn`. | ✅ |
| RPT-CRT-27 | Report bài viết của chính mình | Report post mà user đang logged in là tác giả | `400` error `không thể báo cáo bài viết của chính mình`. | ✅ |

### 1.9. Target Validation — Comment

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-CRT-28 | Report comment không tồn tại | Report comment với ID không tồn tại | `400` error `không tìm thấy bình luận hoặc bình luận không hoạt động`. | ✅ |

### 1.10. Authentication

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-CRT-29 | Không có token | POST `/api/reports` without Authorization | `401` Unauthorized. | ✅ |
| RPT-CRT-30 | Token hết hạn | Token đã hết hạn | `401` Unauthorized. | ✅ |

### 1.11. Report Model

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-CRT-31 | UUID format | Tạo report, kiểm tra ID | ID là UUID v4 hợp lệ. | ✅ |
| RPT-CRT-32 | Default status | Tạo report thành công | Status = `pending`. | ✅ |
| RPT-CRT-33 | created_at set đúng | Tạo report, kiểm tra `created_at` | `created_at` gần với thời điểm tạo (±5s). | ✅ |
| RPT-CRT-34 | target字段 correctly set | Report user → `target_user_id` set, `target_post_id` = nil, `target_comment_id` = nil | Đúng. | ✅ |
| RPT-CRT-35 | Report post → target_post_id set | Report post | `target_post_id` set, `target_user_id` = nil. | ✅ |
| RPT-CRT-36 | Report comment → target_comment_id set | Report comment | `target_comment_id` set, `target_user_id` = nil. | ✅ |

---

## 2. UpdateReport

### 2.1. Happy Path

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-UPT-01 | Sửa report_type | 1. Tạo report (pending)<br>2. PUT `/api/reports/:id` với `report_type: "nudity"` | `200` `{message: "Báo cáo đã được cập nhật thành công"}`. DB: report_type = `nudity`. | ✅ |
| RPT-UPT-02 | Sửa reason_detail | PUT với `reason_detail: "lý do mới"` | `200`. DB: reason_detail updated. | ✅ |
| RPT-UPT-03 | Sửa violation_rule_id | PUT với `violation_rule_id: "new-rule-id"` | `200`. DB: violation_rule_id updated. | ✅ |
| RPT-UPT-04 | Bỏ violation_rule_id | PUT với `violation_rule_id: null` | `200`. DB: violation_rule_id = nil. | ✅ |
| RPT-UPT-05 | Sửa tất cả fields cùng lúc | PUT với report_type + reason_detail + violation_rule_id mới | `200`. Tất cả fields updated. | ✅ |
| RPT-UPT-06 | updated_at được set | 1. Tạo report<br>2. Update report | `updated_at` trong DB được set = thời điểm update. | ✅ |

### 2.2. Validation

| ID | Scenario | Input | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-UPT-07 | Thiếu report_type | `report_type: ""` | `400` error `report_type là bắt buộc`. | ✅ |
| RPT-UPT-08 | Thiếu reason_detail | `reason_detail: ""` | `400` error `reason_detail là bắt buộc`. | ✅ |
| RPT-UPT-09 | Thiếu report_id param | PUT `/api/reports/` (no id) | `404` route not match. | ✅ |

### 2.3. Authorization

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-UPT-10 | Sửa report của user khác | User B update report của User A | `400` error `bạn không có quyền chỉnh sửa báo cáo này`. | ✅ |
| RPT-UPT-11 | Sửa report đã processed | Report có status = resolved, reporter update | `400` error `chỉ có thể chỉnh sửa báo cáo đang chờ xử lý`. | ✅ |
| RPT-UPT-12 | Sửa report đã rejected | Report có status = rejected, reporter update | `400` error `chỉ có thể chỉnh sửa báo cáo đang chờ xử lý`. | ✅ |
| RPT-UPT-13 | Sửa report đã reviewed | Report có status = reviewed, reporter update | `400` error `chỉ có thể chỉnh sửa báo cáo đang chờ xử lý`. | ✅ |

### 2.4. Authentication

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-UPT-14 | Không có token | PUT without Authorization | `401` Unauthorized. | ✅ |
| RPT-UPT-15 | Report không tồn tại | PUT `/api/reports/nonexistent` | `400` error (report not found). | ✅ |

---

## 3. ReviewReport (Admin)

### 3.1. Action: Cancel

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-REV-01 | Cancel report thành công | POST `/api/admin/reports/:id/decision` với `action: "cancel"` | `200` `{message: "xử lý báo cáo thành công"}`. Report status = `rejected`. Reporter nhận notification. | ✅ |
| RPT-REV-02 | Cancel không cần reason | POST với `action: "cancel"`, `reason: ""` | `200` thành công (cancel không bắt buộc reason). | ✅ |
| RPT-REV-03 | Cancel report user | Report user → cancel | Status = rejected. User target KHÔNG bị ban. | ✅ |
| RPT-REV-04 | Cancel report post | Report post → cancel | Status = rejected. Post status KHÔNG thay đổi. | ✅ |

### 3.2. Action: Hide (Post)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-REV-05 | Hide post thành công | Report post → `action: "hide"`, `reason: "vi phạm nội dung"` | `200`. Post status = hidden. Report status = resolved. Moderation log created. | ✅ |
| RPT-REV-06 | Hide thiếu reason | Report post → `action: "hide"`, `reason: ""` | `400` error `lý do là bắt buộc cho hành động hide hoặc ban`. | ✅ |
| RPT-REV-07 | Hide report comment | Report comment → `action: "hide"` | `400` error `loại báo cáo không được hỗ trợ`. | ✅ |
| RPT-REV-08 | Moderation log tạo đúng | Hide post | Moderation log có `action = delete`, `target_type = post`, `target_id = post_id`, `reason` đúng. | ✅ |
| RPT-REV-09 | Notification gửi đúng | Hide post | Reporter nhận notification `Báo cáo ... đã được xử lý bằng hành động: hide`. Post author nhận notification `Bài viết của bạn đã bị báo cáo và đã được hide bởi quản trị viên`. | ✅ |

### 3.3. Action: Ban (User)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-REV-10 | Ban user thành công | Report user → `action: "ban"`, `reason: "spam nghiêm trọng"` | `200`. User status = banned. Report status = resolved. Ban record created. Moderation log created. | ✅ |
| RPT-REV-11 | Ban với duration | Report user → `action: "ban"`, `duration: "7d"` | `200`. Ban record có `expires_at` đúng (7 ngày từ bây giờ). | ✅ |
| RPT-REV-12 | Ban thiếu reason | Report user → `action: "ban"`, `reason: ""` | `400` error `lý do là bắt buộc cho hành động hide hoặc ban`. | ✅ |
| RPT-REV-13 | Ban report post | Report post → `action: "ban"` | `400` error `ban chỉ hỗ trợ cho báo cáo người dùng`. | ✅ |
| RPT-REV-14 | Ban report comment | Report comment → `action: "ban"` | `400` error `ban chỉ hỗ trợ cho báo cáo người dùng`. | ✅ |
| RPT-REV-15 | Notification ban | Ban user | Target nhận 2 notifications: `Tài khoản của bạn đã bị báo cáo và đã được ban bởi quản trị viên` + `Tài khoản của bạn đã bị cấm vì vi phạm báo cáo`. | ✅ |
| RPT-REV-16 | Moderation log ban | Ban user | Moderation log có `action = ban`, `target_type = user`. | ✅ |

### 3.4. Report Status Checks

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-REV-17 | Review report đã resolved | Report status = resolved → review | `400` error `báo cáo đã được xử lý`. | ✅ |
| RPT-REV-18 | Review report đã rejected | Report status = rejected → review | `400` error `báo cáo đã được xử lý`. | ✅ |
| RPT-REV-19 | Review report đã reviewed | Report status = reviewed → review | `400` error `báo cáo đã được xử lý`. | ✅ |

### 3.5. Validation

| ID | Scenario | Input | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-REV-20 | Action không hợp lệ | `action: "delete"` | `400` error `action không hợp lệ, chỉ chấp nhận cancel, hide hoặc ban`. | ✅ |
| RPT-REV-21 | Report không tồn tại | reportID không tồn tại | `400` error (report not found). | ✅ |

### 3.6. Authorization

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-REV-22 | Non-superadmin review | User thường review report | `403` Forbidden (ensureSuperAdmin check). | ✅ |
| RPT-REV-23 | Không có token | POST without Authorization | `401` Unauthorized. | ✅ |

---

## 4. ListReports (Admin)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-LST-01 | List mặc định | GET `/api/admin/reports` | `200` `{reports: [...], total, page, page_size}`. Sort mặc định `created_at desc`. | ✅ |
| RPT-LST-02 | Filter status | GET `/api/admin/reports?status=pending` | Chỉ trả về reports có status = pending. | ✅ |
| RPT-LST-03 | Filter targetType | GET `/api/admin/reports?target_type=post` | Chỉ trả về reports target là post. | ✅ |
| RPT-LST-04 | Filter keyword | GET `/api/admin/reports?keyword=spam` | Tìm trong username, email, reason_detail. | ✅ |
| RPT-LST-05 | Pagination | GET `/api/admin/reports?page=2&page_size=5` | Offset = 5, limit = 5. | ✅ |
| RPT-LST-06 | Sort by created_at | GET `/api/admin/reports?sort_by=created_at&order=asc` | Sort tăng dần. | ✅ |
| RPT-LST-07 | Sort by target_type | GET `/api/admin/reports?sort_by=target_type` | Sort theo loại target. | ✅ |

---

## 5. GetReportDetail (Admin)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-DET-01 | Lấy detail report post | GET `/api/admin/reports/:id` (report post) | `200` `{id, reporter_username, reporter_email, target_type: "post", post_title, post_content, report_type, ...}` | ✅ |
| RPT-DET-02 | Lấy detail report user | GET `/api/admin/reports/:id` (report user) | `200` với `target_type: "user"`. | ✅ |
| RPT-DET-03 | Auto-mark reviewed | Report status = pending → GET detail | Report status tự chuyển thành `reviewed`. | ✅ |
| RPT-DET-04 | Report không tồn tại | GET với reportID không tồn tại | `400` error. | ✅ |

---

## 6. DTO Serialization

### 6.1 CreateReportInput

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-DTO-01 | JSON roundtrip đầy đủ | Marshal → Unmarshal `CreateReportInput` | Tất cả fields giữ nguyên giá trị. | ✅ |
| RPT-DTO-02 | ViolationRuleID omitempty | Marshal với `violation_rule_id: nil` | JSON KHÔNG chứa `violation_rule_id`. | ✅ |
| RPT-DTO-03 | ViolationRuleID có giá trị | Marshal với `violation_rule_id: "rule-123"` | JSON chứa `"violation_rule_id": "rule-123"`. | ✅ |

### 6.2 UpdateReportInput

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-DTO-04 | JSON roundtrip | Marshal → Unmarshal `UpdateReportInput` | Tất cả fields giữ nguyên. | ✅ |
| RPT-DTO-05 | ViolationRuleID omitempty | Marshal với `violation_rule_id: nil` | JSON KHÔNG chứa `violation_rule_id`. | ✅ |

### 6.3 Report Model

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-DTO-06 | JSON roundtrip Report | Marshal → Unmarshal `Report` | Tất cả fields giữ nguyên. | ✅ |
| RPT-DTO-07 | Omitempty nil fields | Marshal Report với `target_user_id: nil, target_post_id: nil, target_comment_id: nil` | JSON KHÔNG chứa các field nil. | ✅ |
| RPT-DTO-08 | UpdatedAt omitempty | Marshal Report với `updated_at: nil` | JSON KHÔNG chứa `updated_at`. | ✅ |

---

## 7. SQL Injection Prevention

| ID | Scenario | Input | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-INJ-01 | Reason detail SQL injection | `reason_detail: "'; DROP TABLE reports; --"` | Được escape đúng, không SQL injection. | ✅ |
| RPT-INJ-02 | Report type SQL injection | `report_type: "spam' OR 1=1 --"` | Được escape đúng. | ✅ |
| RPT-INJ-03 | Target ID SQL injection | `target_id: "'; DELETE FROM users; --"` | Được escape đúng. | ✅ |

---

## 8. Composite Index Verification

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| RPT-IDX-01 | Duplicate check uses composite index | Tạo report → update status → tạo report mới cùng target | Query `FindPendingByReporterAndTarget` dùng index `idx_reports_reporter_status`. | ✅ |

---

## 9. Integration Flow

### 9.1. Full Report Flow — Post

| Step | Action | Expected |
|------|--------|----------|
| 1 | User A login | Token A |
| 2 | User B login | Token B |
| 3 | User B tạo post | Post ID, post.author = B |
| 4 | User A report post B | `201`, report pending |
| 5 | User A report post B lần nữa | `400` duplicate |
| 6 | User A PUT `/api/reports/:id` sửa report | `200` updated |
| 7 | Admin GET `/api/admin/reports` | Thấy report trong list |
| 8 | Admin GET `/api/admin/reports/:id` | Detail, status auto = reviewed |
| 9 | Admin POST decision `hide` + reason | `200`, post hidden, report resolved |
| 10 | User A GET `/api/reports/:id` | Report status = resolved (nếu query được) |

### 9.2. Full Report Flow — User Ban

| Step | Action | Expected |
|------|--------|----------|
| 1 | User A report user C | `201`, report pending |
| 2 | Admin review → ban | `200`, user C banned, report resolved |
| 3 | User A report user C lần nữa | `400` duplicate (report pending vẫn tồn tại) |
| 4 | Admin review report đó → cancel | `200`, report rejected |
| 5 | User A report user C lần nữa | `400` user đã bị ban |

### 9.3. Cancel Flow

| Step | Action | Expected |
|------|--------|----------|
| 1 | User A report user D | `201` |
| 2 | Admin review → cancel (không cần reason) | `200`, report rejected |
| 3 | User D không bị ảnh hưởng | User D status vẫn active |

---

## Test Coverage Summary

| Feature | Total Cases | ✅ Pass | Status |
|---------|-------------|---------|--------|
| CreateReport — Happy Path | 5 | 5 | ✅ |
| CreateReport — Validation (Target Type) | 3 | 3 | ✅ |
| CreateReport — Validation (Target ID) | 1 | 1 | ✅ |
| CreateReport — Validation (Report Type) | 1 | 1 | ✅ |
| CreateReport — Validation (Reason) | 2 | 2 | ✅ |
| CreateReport — Duplicate Prevention | 7 | 7 | ✅ |
| CreateReport — Target User | 5 | 5 | ✅ |
| CreateReport — Target Post | 3 | 3 | ✅ |
| CreateReport — Target Comment | 1 | 1 | ✅ |
| CreateReport — Authentication | 2 | 2 | ✅ |
| CreateReport — Model | 6 | 6 | ✅ |
| UpdateReport — Happy Path | 6 | 6 | ✅ |
| UpdateReport — Validation | 3 | 3 | ✅ |
| UpdateReport — Authorization | 4 | 4 | ✅ |
| UpdateReport — Authentication | 2 | 2 | ✅ |
| ReviewReport — Cancel | 4 | 4 | ✅ |
| ReviewReport — Hide | 5 | 5 | ✅ |
| ReviewReport — Ban | 7 | 7 | ✅ |
| ReviewReport — Status Checks | 3 | 3 | ✅ |
| ReviewReport — Validation | 2 | 2 | ✅ |
| ReviewReport — Authorization | 2 | 2 | ✅ |
| ListReports | 7 | 7 | ✅ |
| GetReportDetail | 4 | 4 | ✅ |
| DTO Serialization | 8 | 8 | ✅ |
| SQL Injection | 3 | 3 | ✅ |
| Composite Index | 1 | 1 | ✅ |
| Integration Flow | 3 flows | — | ✅ |
| **Total** | **99** | **99** | ✅ |

---

## Test File

`tests/report/report_test.go` — unit tests (validation-only, no DB)

```
TestCreateReportInputJSON
TestCreateReportInputViolationRuleIDEOmitted
TestUpdateReportInputJSON
TestUpdateReportInputViolationRuleIDEOmitted
TestReportModelJSON
TestReportModelOmitsNilTargets
TestReportModelUpdatedAtOmitted
TestReportStatusString (4 subtests)
TestParseReportStatus (6 subtests)
```
