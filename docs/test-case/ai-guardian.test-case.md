# AI Guardian (Moderation) — Test Cases

## REST API Endpoints

| Method | Path | Handler | Auth | Description |
|--------|------|---------|------|-------------|
| POST | `/api/media/upload` | UploadMedia | Token | Upload file + AI moderation (Cloudinary + AWS Rekognition) |
| POST | `/api/communities` | CreateCommunity | Token | Tạo community (avatar được kiểm tra status) |
| POST | `/api/communities/:communityID/background` | SetCommunityBackground | Token | Set background community (kiểm tra status) |
| GET | `/api/admin/media/flagged` | ListFlaggedMedia | Admin | Danh sách media flagged/rejected |
| POST | `/api/admin/media/:id/review` | ReviewMedia | Admin | Approve/reject media |
| POST | `/api/admin/media/cleanup-rejected` | CleanupRejectedMedia | Admin | Xoá media reject cũ |

---

## 1. UploadMedia — AI Moderation

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| AI-UPLOAD-01 | Upload ảnh hợp lệ (nội dung sạch) | POST `/api/media/upload` với file ảnh | `200` response. `status = "approved"`. `file_uri` trả về URL Cloudinary. | ✅ |
| AI-UPLOAD-02 | Upload ảnh vi phạm (AWS Rekognition reject) | POST `/api/media/upload` với ảnh nhạy cảm | `200` response. `status = "rejected"`. `file_uri` trả về URL Cloudinary (file giữ nguyên để admin override). | ✅ |
| AI-UPLOAD-03 | Upload ảnh moderation pending | POST `/api/media/upload` khi moderation trả `pending` | `200` response. `status = "flagged"` (pending → flagged). `file_uri` trả về URL. | ✅ |
| AI-UPLOAD-04 | Upload ảnh không có moderation (thiếu AWS Rek) | POST `/api/media/upload` khi Cloudinary không có moderation | `200` response. `status = "flagged"` (fallback). `file_uri` trả về URL. | ✅ |
| AI-UPLOAD-05 | Upload file quá dung lượng | POST `/api/media/upload` với file > storage quota | `400` error. Không lưu DB, không upload Cloudinary. | ✅ |
| AI-UPLOAD-06 | Upload file sai định dạng | POST `/api/media/upload` với file `.exe` | `400` error từ validation. | ✅ |
| AI-UPLOAD-07 | Upload không có token | POST `/api/media/upload` không gửi Authorization | `401` Unauthorized. | ✅ |
| AI-UPLOAD-08 | Upload không có file | POST `/api/media/upload` với body rỗng | `400` error. | ✅ |
| AI-UPLOAD-09 | Upload ảnh → notification gửi đúng message | Upload ảnh approved | User nhận notification: `"Ảnh/video của bạn đã được hệ thống tự động duyệt."` | ✅ |
| AI-UPLOAD-10 | Upload ảnh bị reject → notification gửi đúng | Upload ảnh bị AI reject | User nhận notification: `"Ảnh/video của bạn bị từ chối do vi phạm tiêu chuẩn cộng đồng."` | ✅ |
| AI-UPLOAD-11 | Upload ảnh flagged → notification gửi đúng | Upload ảnh bị AI flagged | User nhận notification: `"Ảnh/video của bạn đang chờ admin kiểm duyệt."` | ✅ |

---

## 2. Community — Avatar Status Check

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| COM-AV-01 | Tạo community với avatar approved | POST `/api/communities` kèm file avatar (approved) | `200`. `avatar_uri` được set, community tạo thành công. | ✅ |
| COM-AV-02 | Tạo community với avatar rejected | POST `/api/communities` kèm file avatar (rejected) | `400` error `"Ảnh đại diện vi phạm tiêu chuẩn cộng đồng"`. Community không được tạo. | ✅ |
| COM-AV-03 | Tạo community với avatar flagged | POST `/api/communities` kèm file avatar (flagged) | `200`. Community tạo thành công với `avatar_uri` tạm thời. | ✅ |
| COM-AV-04 | Tạo community với avatar pending | POST `/api/communities` kèm file avatar (pending) | `200`. Community tạo thành công với `avatar_uri` tạm thời. | ✅ |
| COM-AV-05 | Tạo community không có avatar | POST `/api/communities` không gửi file avatar | `200`. Community tạo thành công, `avatar_uri = ""`. | ✅ |

## 3. Community — Background Status Check

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| COM-BG-01 | Set background approved | POST `/api/communities/:id/background` với file background (approved) | `200`. Background cập nhật thành công. | ✅ |
| COM-BG-02 | Set background rejected | POST `/api/communities/:id/background` với file background (rejected) | `400` error `"ảnh background vi phạm tiêu chuẩn cộng đồng"`. Background không đổi. | ✅ |
| COM-BG-03 | Set background flagged/pending | POST `/api/communities/:id/background` với file background (flagged/pending) | `200`. Background cập nhật (tạm thời). | ✅ |
| COM-BG-04 | Set background không có file | POST `/api/communities/:id/background` không có file | `400` error `"Vui lòng chọn ảnh background"`. | ✅ |
| COM-BG-05 | Set background bởi non-admin | POST bởi user không phải admin của community | `400` error (không có quyền). | ✅ |
| COM-BG-06 | Set background với community không tồn tại | POST `/api/communities/fake-id/background` | `400` error `"cộng đồng không tồn tại"`. | ✅ |

---

## 4. Admin — ListFlaggedMedia

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| ADM-LFM-01 | Danh sách mặc định (status=flagged) | GET `/api/admin/media/flagged` | `200`. Trả về items có `status = "flagged"`. `total` = số lượng. | ✅ |
| ADM-LFM-02 | Lọc theo status=rejected | GET `/api/admin/media/flagged?status=rejected` | `200`. Chỉ trả về items `status = "rejected"`. | ✅ |
| ADM-LFM-03 | Lọc status không hợp lệ | GET `/api/admin/media/flagged?status=deleted` | `200`. Fallback về `status = "flagged"` (ParseMediaStatus trả default). | ✅ |
| ADM-LFM-04 | Phân trang | GET `/api/admin/media/flagged?page=1&page_size=5` | `200`. Chỉ 5 items. `total` = tổng số. | ✅ |
| ADM-LFM-05 | page_size > 100 | GET `/api/admin/media/flagged?page_size=200` | Clamp về 100 items. | ✅ |
| ADM-LFM-06 | Không có data | Không có media nào | `items: []`, `total: 0`. | ✅ |
| ADM-LFM-07 | Non-admin gọi API | GET bởi user thường (không admin) | `403` error `"chỉ có admin/superadmin mới được phép"`. | ✅ |
| ADM-LFM-08 | Không có token | GET không gửi Authorization | `401` Unauthorized. | ✅ |

---

## 5. Admin — ReviewMedia

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| ADM-RM-01 | Approve media flagged | POST `/api/admin/media/:id/review` `{"action":"approve","reason":"OK"}` | `200`. Status thành `approved`. Moderation log được tạo. User nhận notification. | ✅ |
| ADM-RM-02 | Reject media flagged | POST `/api/admin/media/:id/review` `{"action":"reject","reason":"Spam"}` | `200`. Status thành `rejected`. File Cloudinary bị Destroy. `FileURI` trong DB clear thành `""`. User nhận notification. Moderation log được tạo. | ✅ |
| ADM-RM-03 | Reject nhưng Cloudinary chưa cấu hình | Set cloudinary = nil, reject media | `200`. Log warning `[Admin] cảnh báo: Cloudinary chưa được cấu hình`. Skip Destroy. | ✅ |
| ADM-RM-04 | Reject mà parse URL thất bại | Media có FileURI không hợp lệ, reject | `200`. Log warning `[Admin] không thể xoá media ... parse URL thất bại`. Skip Destroy. | ✅ |
| ADM-RM-05 | Action không hợp lệ | POST `{"action":"ban","reason":"x"}` | `400` error `"hành động không hợp lệ: ban"`. | ✅ |
| ADM-RM-06 | Media không ở trạng thái flagged | Review media đã approved/rejected | `400` error `"media ở trạng thái approved không thể review"`. | ✅ |
| ADM-RM-07 | Media không tồn tại | POST với id không hợp lệ | `400` error `"media không tồn tại"`. | ✅ |
| ADM-RM-08 | Non-admin gọi API | POST bởi user thường | `403` error. | ✅ |
| ADM-RM-09 | Không có token | POST không gửi Authorization | `401` Unauthorized. | ✅ |
| ADM-RM-10 | Approve/reject không thay đổi storage_used_bytes | So sánh storage trước và sau review | Không thay đổi (review không ảnh hưởng storage). | ✅ |

---

## 6. Admin — CleanupRejectedMedia

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| ADM-CRM-01 | Cleanup thành công | Có 3 media reject > 7 ngày | POST `/api/admin/media/cleanup-rejected` | `200`. `cleaned = 3`. File Cloudinary bị Destroy. Record DB bị xoá (DeleteWithStorageAdjustment). | ✅ |
| ADM-CRM-02 | Cleanup không có media cũ | Chỉ có media reject < 7 ngày | `200`. `cleaned = 0`. | ✅ |
| ADM-CRM-03 | Cleanup không ảnh hưởng media flagged | Có media flagged cũ + reject cũ | `cleaned = 1` (chỉ reject cũ). Flagged không bị động. | ✅ |
| ADM-CRM-04 | Cloudinary Destroy fail | Cloudinary không available | Log warning, vẫn delete DB record. `cleaned` vẫn tăng. | ✅ |
| ADM-CRM-05 | Non-admin gọi API | Người dùng thường | `403` error. | ✅ |
| ADM-CRM-06 | Không có token | Không gửi Authorization | `401` Unauthorized. | ✅ |

---

## 7. Notification Messages

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| NOTIF-01 | AI auto-approve | Upload ảnh → AI trả approved | Notification content: `"Ảnh/video của bạn đã được hệ thống tự động duyệt."` | ✅ |
| NOTIF-02 | AI reject | Upload ảnh → AI trả rejected | Notification content: `"Ảnh/video của bạn bị từ chối do vi phạm tiêu chuẩn cộng đồng."` | ✅ |
| NOTIF-03 | AI flagged | Upload ảnh → AI trả flagged/pending | Notification content: `"Ảnh/video của bạn đang chờ admin kiểm duyệt."` | ✅ |
| NOTIF-04 | Admin approve | Admin approve media | Notification content: `"Ảnh/video của bạn đã được admin phê duyệt."` | ✅ |
| NOTIF-05 | Admin reject | Admin reject media kèm reason | Notification content: `"Ảnh/video của bạn đã bị admin từ chối: {reason}"` | ✅ |

---

## 8. Integration — Transition Validation

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| TRANS-01 | Flagged → Approved hợp lệ | ReviewMedia flagged media với action=approve | `200`. Status thành `approved`. | ✅ |
| TRANS-02 | Flagged → Rejected hợp lệ | ReviewMedia flagged media với action=reject | `200`. Status thành `rejected`. | ✅ |
| TRANS-03 | Approved → Rejected không hợp lệ | ReviewMedia media đã approved với action=reject | `400` error `"media ở trạng thái approved không thể review"`. | ✅ |
| TRANS-04 | Rejected → Approved không hợp lệ | ReviewMedia media đã rejected với action=approve | `400` error. | ✅ |
| TRANS-05 | Pending → Approved không hợp lệ | ReviewMedia media pending với action=approve | `400` error. | ✅ |

---

## 9. Security

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| SEC-01 | Admin endpoints require admin role | User thường gọi bất kỳ admin endpoint | `403` error. | ✅ |
| SEC-02 | Upload yêu cầu auth | Upload không token | `401` Unauthorized. | ✅ |
| SEC-03 | Review media không flag (replay) | Review cùng media 2 lần với action=approve | Lần 1: `200`. Lần 2: `400` (đã approved). | ✅ |
| SEC-04 | Cleanup media của user khác | Admin cleanup chỉ xoá media của user khác | Chỉ xoá media đúng điều kiện (reject > 7 ngày), không phân biệt user. | ✅ |
