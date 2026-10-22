# Tài liệu hàm API dành cho Admin (toàn hệ thống)

> **Nguồn**: routes/*.go — 53 endpoint yêu cầu RBAC quản trị (RequireRoles SuperAdmin/Admin/Partner).
> Mỗi mục gồm: **Name** (tên func tiếng Anh), **Description** (luồng tương tác tiếng Việt), **Entity**, **Model**, **Databases**.
> README/AGENTS.md mô tả tổng quan hệ thống; tài liệu này chi tiết từng hàm.

## Mục lục

| # | Endpoint | Name |
|---|---|---|
| 1 | GET /api/admin/analytics | Tổng quan dashboard (analytics) |
| 2 | GET /api/admin/feed-metrics | Số liệu feed (rollout rank v2) |
| 3 | GET /api/admin/users | Danh sách người dùng |
| 4 | PUT /api/admin/users/:userID/status | Cập nhật trạng thái người dùng |
| 5 | POST /api/admin/users/:userID/ban | Cấm (ban) người dùng |
| 6 | GET /api/admin/posts | Danh sách bài viết |
| 7 | PUT /api/admin/posts/:postID/hide | Ẩn bài viết |
| 8 | PUT /api/admin/posts/:postID/status | Đổi trạng thái bài viết |
| 9 | GET /api/admin/comments | Danh sách bình luận |
| 10 | PUT /api/admin/comments/:commentID/hide | Ẩn bình luận |
| 11 | PUT /api/admin/comments/:commentID/reveal | Hiện lại bình luận |
| 12 | GET /api/admin/reports | Danh sách báo cáo |
| 13 | GET /api/admin/reports/:reportID | Chi tiết báo cáo |
| 14 | POST /api/admin/reports/:reportID/decision | Xử lý (ra quyết định) báo cáo |
| 15 | GET /api/admin/media/grouped | Media nhóm theo người dùng |
| 16 | GET /api/admin/media/flagged | Media bị gắn cờ |
| 17 | POST /api/admin/media/:id/review | Duyệt media (approve/reject) |
| 18 | POST /api/admin/media/cleanup-rejected | Dọn dẹp media đã reject |
| 19 | GET /api/admin/groups | Danh sách nhóm chat |
| 20 | GET /api/admin/groups/:chatID | Chi tiết nhóm chat |
| 21 | GET /api/admin/groups/:chatID/members | Danh sách thành viên nhóm chat |
| 22 | GET /api/admin/groups/:chatID/logs | Nhật ký kiểm duyệt của nhóm chat |
| 23 | POST /api/admin/groups/:chatID/hide | Ẩn nhóm chat |
| 24 | POST /api/admin/groups/:chatID/unhide | Bỏ ẩn nhóm chat |
| 25 | POST /api/admin/groups/:chatID/archive | Đình chỉ nhóm chat |
| 26 | POST /api/admin/groups/:chatID/warn | Cảnh báo nhóm chat |
| 27 | DELETE /api/admin/groups/:chatID | Xóa nhóm chat |
| 28 | GET /api/admin/communities | Danh sách cộng đồng |
| 29 | GET /api/admin/communities/:id | Chi tiết cộng đồng |
| 30 | GET /api/admin/communities/:id/members | Danh sách thành viên cộng đồng |
| 31 | GET /api/admin/communities/:id/logs | Nhật ký kiểm duyệt của cộng đồng |
| 32 | POST /api/admin/communities/:id/hide | Ẩn cộng đồng |
| 33 | POST /api/admin/communities/:id/unhide | Bỏ ẩn cộng đồng |
| 34 | POST /api/admin/communities/:id/archive | Đình chỉ cộng đồng |
| 35 | POST /api/admin/communities/:id/unarchive | Bỏ đình chỉ cộng đồng |
| 36 | POST /api/admin/communities/:id/warn | Cảnh báo cộng đồng |
| 37 | DELETE /api/admin/communities/:id | Xóa cộng đồng |
| 38 | GET /api/admin/ads | Danh sách quảng cáo (admin) |
| 39 | PATCH /api/admin/ads/:id/status | Duyệt / đổi trạng thái quảng cáo (admin) |
| 40 | DELETE /api/admin/ads/:id | Xoá quảng cáo (admin) |
| 41 | GET /api/admin/settings | Đọc cấu hình site settings |
| 42 | PUT /api/admin/settings | Cập nhật cấu hình site settings |
| 43 | POST /api/ads-management | Tạo quảng cáo (ads-management) |
| 44 | GET /api/ads-management | Danh sách quảng cáo (ads-management, admin hoặc partner) |
| 45 | GET /api/ads-management/:id/analytics | Phân tích hiệu quả 1 quảng cáo (analytics) |
| 46 | PATCH /api/ads-management/:id/status | Đổi trạng thái quảng cáo (partner/admin) |
| 47 | PATCH /api/ads-management/:id | Cập nhật nội dung quảng cáo |
| 48 | DELETE /api/ads-management/:id | Xoá quảng cáo (ads-management) |
| 49 | GET /api/ads-management/overview | Tổng quan quảng cáo (overview) |
| 50 | GET /api/admin/violation-rules | Liệt kê toàn bộ quy tắc vi phạm (admin) |
| 51 | POST /api/admin/violation-rules | Tạo quy tắc vi phạm mới |
| 52 | PUT /api/admin/violation-rules/:id | Cập nhật quy tắc vi phạm |
| 53 | PATCH /api/admin/violation-rules/:id/active | Bật / tắt quy tắc vi phạm |

---
## Admin Endpoints — Module kiểm duyệt nội dung

Nguồn: `routes/admin.routes.go` (group `/api/admin`, `middlewares.AuthMiddleware` + `middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin)`), `controllers/admin.controller.go`, `services/admin.service.go`.

### 1. Tổng quan dashboard (analytics)

- **Name**: GetDashboardAnalytics
- **Description**:
  1. Admin gọi API dashboard kèm bộ lọc khoảng thời gian (tuỳ chọn).
  2. Kiểm tra người gọi có quyền admin (SuperAdmin hoặc Admin), không có quyền thì báo lỗi truy cập.
  3. Chuẩn hoá khoảng thời gian: nếu thiếu ngày kết thúc thì lấy hôm nay, nếu thiếu ngày bắt đầu thì lấy 7 ngày trước.
  4. Lấy đồng thời các số liệu bắt buộc: tổng số người dùng, tổng số bài viết (loại bài đã xoá), tổng số báo cáo — lỗi ở bước này thì huỷ toàn bộ và báo lấy thống kê thất bại.
  5. Lấy đồng thời các số liệu phụ (bỏ qua nếu có lỗi): tổng bình luận, tổng media, tổng nhóm chat, tổng cộng đồng, số lệnh cấm còn hiệu quả, báo cáo chờ xử lý, media bị gắn cờ, người dùng hoạt động trong ngày, tổng lượt thích, tổng lượt chia sẻ, 5 người dùng hoạt động nhất, 5 bài viết tương tác cao nhất, phân bố trạng thái người dùng, phân bố trạng thái báo cáo, 4 chuỗi biểu đồ theo ngày cho người dùng, bài viết, báo cáo, bình luận trong khoảng đã chọn.
  6. Đếm số bản ghi của 7 bảng chủ đạo tại mốc một tháng trước để làm số liệu so sánh.
  7. Tính phần trăm thay đổi so với tháng trước (nếu cả hai tháng đều bằng 0 thì 0 phần trăm, tháng trước bằng 0 mà tháng này có dữ liệu thì 100 phần trăm).
  8. Trả về tổng số liệu, phần trăm thay đổi, biểu đồ, top người dùng, top bài viết, phân bố trạng thái và thời điểm tạo thống kê; chỉ đọc, không ghi dữ liệu.
- **Entity**: Admin (người xem dashboard), Hệ thống
- **Model**: models.User, models.Post, models.Report, models.Comment, models.Media, models.Chat, models.Community, models.Ban, models.PostReaction, models.PostShare
- **Databases**: users, posts, reports, comments, media, chats, communities, bans, post_reactions, post_shares, profiles

### 2. Số liệu feed (rollout rank v2)

- **Name**: GetFeedMetrics
- **Description**:
  1. Admin gọi API số liệu feed kèm số ngày cần xem.
  2. Kiểm tra người gọi có quyền admin (SuperAdmin hoặc Admin), không có quyền thì báo lỗi truy cập.
  3. Chuẩn hoá số ngày: nhỏ hơn hoặc bằng 0 thì lấy 7, lớn hơn 90 thì chặn ở 90; xác định mốc thời gian bắt đầu và ngưỡng người theo dõi cho tác giả nhỏ.
  4. Chạy 6 truy vấn tổng hợp số liệu feed: tổng lượt xem đã khử trùng lặp, lượt xem đến từ nguồn bảng tin, lượt xem đến từ trang chi tiết, lượt xem của bài do tác giả ít người theo dõi sáng tạo, số tác giả nhỏ được hiển thị khác nhau, số báo cáo bài viết trong khoảng thời gian.
  5. Tính các chỉ số dẫn xuất: tỷ lệ mở trang chi tiết trên lượt xem bảng tin, và phần trăm lượt xem dành cho bài của tác giả nhỏ.
  6. Trả về số liệu feed; chỉ đọc, không ghi dữ liệu.
- **Entity**: Admin, Hệ thống (số liệu feed)
- **Model**: models.Post, models.PostView, models.Report
- **Databases**: post_views, posts, reports, follows

### 3. Danh sách người dùng

- **Name**: ListUsers
- **Description**:
  1. Admin gọi API danh sách người dùng kèm bộ lọc từ khóa, trạng thái và phân trang.
  2. Kiểm tra người gọi có quyền admin (SuperAdmin hoặc Admin), không có quyền thì báo lỗi truy cập.
  3. Chuẩn hoá phân trang: trang mặc định 1, kích thước trang mặc định 20, tối đa 100.
  4. Kiểm tra trạng thái phải để trống hoặc thuộc đang hoạt động, bị cấm, bị tạm ngưng — sai thì báo trạng thái không hợp lệ.
  5. Truy vấn danh sách người dùng kèm thông tin hồ sơ, tìm theo tên người dùng, email hoặc tên hiển thị, lọc theo trạng thái, loại trừ toàn bộ người là admin, sắp xếp mới nhất trước và lấy theo trang.
  6. Đếm tổng số người dùng với cùng bộ điều kiện để phục vụ phân trang.
  7. Trả về danh sách người dùng đã phân trang kèm tổng số; nếu không có kết quả thì báo không tìm thấy người dùng.
- **Entity**: Admin (người xem), Người dùng (đối tượng liệt kê)
- **Model**: models.User, models.Profile, models.UserStatus
- **Databases**: users, profiles, user_roles, roles

### 4. Cập nhật trạng thái người dùng

- **Name**: UpdateUserStatus
- **Description**:
  1. SuperAdmin gọi API với mã người dùng cần đổi và trạng thái mới trong thân yêu cầu.
  2. Kiểm tra người gọi là SuperAdmin, admin thường không được dùng chức năng này.
  3. Kiểm tra người dùng đích có tồn tại, không tồn tại thì báo lỗi.
  4. Chặn thao tác nếu người dùng đích cũng là SuperAdmin.
  5. Kiểm tra trạng thái mới chỉ nhận đang hoạt động, bị cấm hoặc bị tạm ngưng — sai thì báo trạng thái không hợp lệ.
  6. Nếu trạng thái đích đã trùng trạng thái hiện tại thì bỏ qua, không ghi gì.
  7. Ghi trạng thái mới cho người dùng vào database.
  8. Huỷ phiên đăng nhập của người dùng bị ảnh hưởng và gửi thông báo cho người đó (bỏ qua lỗi gửi).
  9. Trả về thông báo cập nhật trạng thái người dùng thành công.
- **Entity**: SuperAdmin (người thao tác), Người dùng đích (bị ảnh hưởng)
- **Model**: models.User, models.UserStatus, models.RoleSuperAdmin, models.NotificationTypeMessage
- **Databases**: users, user_roles, roles, notifications

### 5. Cấm (ban) người dùng

- **Name**: BanUser
- **Description**:
  1. SuperAdmin gọi API với mã người dùng cần cấm, lý do và thời hạn cấm.
  2. Kiểm tra người gọi là SuperAdmin, chỉ SuperAdmin mới cấm được người dùng.
  3. Kiểm tra người dùng đích tồn tại; chặn nếu đích là admin, và chặn nếu người này đã ở trạng thái bị cấm.
  4. Tính hạn cấm: nếu chọn vĩnh viễn thì không có ngày hết hạn, ngược lại tra thời hạn trong các mức được hỗ trợ — thời hạn sai thì báo không hợp lệ.
  5. Tạo bản ghi cấm mới kèm mã định danh và thời điểm tạo, lưu vào database.
  6. Chuyển quyền sở hữu cộng đồng và nhóm chat mà người dùng đích đang là chủ sang người thay thế hợp lệ (ưu tiên admin còn lại, rồi thành viên lâu nhất, rồi người đóng góp cao nhất); lỗi ở bước chuyển thì bỏ qua, không chặn.
  7. Cập nhật trạng thái người dùng thành bị cấm và vô hiệu hoá phiên đăng nhập của họ.
  8. Gửi thông báo cấm cho người dùng kèm lý do và thời hạn (bỏ qua lỗi gửi).
  9. Trả về thông báo cấm người dùng thành công kèm hạn cấm.
- **Entity**: SuperAdmin (người ban), Người dùng bị ban, Chủ cộng đồng/group thay thế, Hệ thống
- **Model**: models.User, models.Ban, models.Community, models.Chat, models.GroupMember, models.MemberContribution, models.ChatParticipant, models.Notification
- **Databases**: bans, users, user_roles, roles, communities, group_members, member_contributions, chats, chat_participants, notifications

### 6. Danh sách bài viết

- **Name**: ListPosts
- **Description**:
  1. Admin gọi API danh sách bài viết kèm bộ lọc từ khóa, trạng thái và phân trang.
  2. Kiểm tra người gọi có quyền admin (SuperAdmin hoặc Admin), không có quyền thì báo lỗi truy cập.
  3. Chuẩn hoá phân trang: trang mặc định 1, kích thước trang mặc định 20, tối đa 100.
  4. Kiểm tra trạng thái phải thuộc công khai, riêng tư, bị ẩn, bạn bè hoặc đã xoá — sai thì báo trạng thái không hợp lệ.
  5. Truy vấn danh sách bài viết theo từ khóa, trạng thái và trang, đồng thời đếm tổng số với cùng điều kiện.
  6. Gom danh sách tác giả rồi nạp thông tin của họ theo lô (tên người dùng, tên hiển thị, ảnh đại diện); lỗi từng nguồn thì bỏ qua.
  7. Nạp danh sách media cho từng bài viết, loại media bị từ chối sẽ bị loại khỏi danh sách; lỗi nạp chỉ ghi log, không chặn phản hồi.
  8. Trả về danh sách bài viết đã phân trang kèm tổng số; nếu không có kết quả thì báo không tìm thấy bài viết.
- **Entity**: Admin (người xem), Người dùng (chủ bài viết)
- **Model**: models.Post, models.User, models.Profile, models.Media, models.PostStatus
- **Databases**: posts, users, profiles, media

### 7. Ẩn bài viết

- **Name**: HidePost
- **Description**:
  1. SuperAdmin gọi API với mã bài viết cần ẩn và lý do ẩn (lý do là bắt buộc).
  2. Kiểm tra người gọi là SuperAdmin, chỉ SuperAdmin mới ẩn được bài viết.
  3. Kiểm tra bài viết có tồn tại, không tồn tại thì báo không tìm thấy.
  4. Chặn nếu bài viết đã ở trạng thái bị ẩn.
  5. Ghi nhật ký kiểm duyệt với hành động ẩn, đối tượng là bài viết, người thao tác và lý do.
  6. Cập nhật trạng thái bài viết thành bị ẩn trong database.
  7. Gửi thông báo cho chủ bài viết kèm liên kết tới bài viết đó; lỗi gửi sẽ làm thao tác thất bại.
  8. Trả về thông báo ẩn bài viết thành công.
- **Entity**: SuperAdmin (người thao tác), Chủ bài viết (bị ảnh hưởng), Hệ thống (moderation log)
- **Model**: models.Post, models.ModerationLog, models.PostStatus, models.ModerationAction, models.NotificationTypeMessage
- **Databases**: posts, moderation_logs, notifications

### 8. Đổi trạng thái bài viết

- **Name**: ChangePostStatus
- **Description**:
  1. SuperAdmin gọi API với mã bài viết và trạng thái mới cần đổi.
  2. Kiểm tra người gọi là SuperAdmin.
  3. Kiểm tra bài viết có tồn tại, không tồn tại thì báo không tìm thấy.
  4. Kiểm tra trạng thái mới phải thuộc công khai, riêng tư, bị ẩn, bạn bè hoặc đã xoá — sai thì báo trạng thái không hợp lệ.
  5. Nếu trạng thái mới đã trùng trạng thái hiện tại thì bỏ qua, không ghi gì.
  6. Ghi nhật ký kiểm duyệt với lý do là cập nhật trạng thái sang giá trị mới.
  7. Cập nhật trạng thái bài viết trong database.
  8. Gửi thông báo cho chủ bài viết kèm mã bài viết; lỗi gửi sẽ làm thao tác thất bại.
  9. Trả về thông báo cập nhật trạng thái bài viết thành công.
- **Entity**: SuperAdmin, Chủ bài viết, Hệ thống (moderation log)
- **Model**: models.Post, models.PostStatus, models.ModerationLog, models.ModerationAction, models.NotificationTypeMessage
- **Databases**: posts, moderation_logs, notifications

### 9. Danh sách bình luận

- **Name**: ListComments
- **Description**:
  1. Admin gọi API danh sách bình luận kèm bộ lọc từ khóa, trạng thái và phân trang.
  2. Kiểm tra người gọi có quyền admin (SuperAdmin hoặc Admin), không có quyền thì báo lỗi truy cập.
  3. Chuẩn hoá phân trang: trang mặc định 1, kích thước trang mặc định 20, tối đa 100.
  4. Kiểm tra trạng thái phải thuộc đang hoạt động, bị ẩn hoặc đã xoá — sai thì báo trạng thái không hợp lệ.
  5. Truy vấn danh sách bình luận theo từ khóa, trạng thái và trang, đồng thời đếm tổng số với cùng điều kiện.
  6. Nạp thông tin chủ bình luận theo lô (tên người dùng, tên hiển thị, ảnh đại diện); lỗi từng nguồn thì bỏ qua.
  7. Trả về danh sách bình luận kèm lý do xét duyệt (nếu có); nếu không có kết quả thì báo không tìm thấy bình luận.
  8. Trả về danh sách bình luận đã phân trang kèm tổng số; chỉ đọc, không ghi dữ liệu.
- **Entity**: Admin (người xem), Người dùng (chủ bình luận)
- **Model**: models.Comment, models.User, models.Profile, models.CommentStatus
- **Databases**: comments, users, profiles

### 10. Ẩn bình luận

- **Name**: HideComment
- **Description**:
  1. SuperAdmin gọi API với mã bình luận cần ẩn và lý do ẩn (dùng chung trường lý do).
  2. Kiểm tra người gọi là SuperAdmin.
  3. Kiểm tra bình luận có tồn tại, không tồn tại thì báo không tìm thấy.
  4. Chặn nếu bình luận đã ở trạng thái bị ẩn.
  5. Ghi nhật ký kiểm duyệt với hành động ẩn, đối tượng là bình luận và lý do từ người thao tác.
  6. Cập nhật trạng thái bình luận thành bị ẩn và lưu lại lý do xét duyệt.
  7. Gửi thông báo cho chủ bình luận kèm mã bài viết và mã bình luận; lỗi gửi sẽ làm thao tác thất bại.
  8. Trả về thông báo ẩn bình luận thành công.
- **Entity**: SuperAdmin, Người dùng chủ bình luận, Hệ thống (moderation log)
- **Model**: models.Comment, models.CommentStatus, models.ModerationLog, models.ModerationAction, models.NotificationTypeMessage
- **Databases**: comments, moderation_logs, notifications

### 11. Hiện lại bình luận

- **Name**: RevealComment
- **Description**:
  1. SuperAdmin gọi API với mã bình luận cần hiện lại, không cần thân yêu cầu.
  2. Kiểm tra người gọi là SuperAdmin.
  3. Kiểm tra bình luận có tồn tại, không tồn tại thì báo không tìm thấy.
  4. Chỉ hợp lệ khi bình luận đang ở trạng thái bị ẩn, nếu không thì báo bình luận đã ở trạng thái này.
  5. Ghi nhật ký kiểm duyệt với hành động cập nhật, đối tượng là bình luận và lý do là hiện lại bình luận.
  6. Đổi trạng thái bình luận về đang hoạt động và xoá lý do xét duyệt đã lưu.
  7. Gửi thông báo cho chủ bình luận báo bình luận đã được hiện lại, kèm mã bài viết và mã bình luận.
  8. Trả về thông báo hiện bình luận thành công.
- **Entity**: SuperAdmin, Người dùng chủ bình luận, Hệ thống (moderation log)
- **Model**: models.Comment, models.CommentStatus, models.ModerationLog, models.ModerationAction, models.NotificationTypeMessage
- **Databases**: comments, moderation_logs, notifications

### 12. Danh sách báo cáo

- **Name**: ListReports
- **Description**:
  1. Admin gọi API danh sách báo cáo kèm bộ lọc từ khóa, trạng thái, loại đối tượng, cách sắp xếp và phân trang.
  2. Kiểm tra người gọi có quyền admin (SuperAdmin hoặc Admin); thiếu thông tin người gọi thì báo lỗi truy cập.
  3. Chuẩn hoá phân trang: trang mặc định 1, kích thước trang mặc định 20, tối đa 100.
  4. Chuẩn hoá bộ lọc: trạng thái phải hợp lệ, trường sắp xếp chỉ nhận thời gian tạo hoặc loại đối tượng (mặc định thời gian tạo), thứ tự chỉ nhận tăng hoặc giảm (mặc định giảm), loại đối tượng viết thường (bài viết, người dùng hoặc bình luận).
  5. Truy vấn danh sách báo cáo kèm thông tin người báo cáo và tên quy tắc vi phạm, lọc theo từ khóa, trạng thái và loại đối tượng, sắp xếp rồi lấy theo trang.
  6. Đếm tổng số báo cáo với cùng bộ điều kiện để phục vụ phân trang.
  7. Trả về danh sách báo cáo đã phân trang kèm tổng số; chỉ đọc, không ghi dữ liệu.
- **Entity**: Admin (người xem), Người báo cáo (reporter)
- **Model**: models.Report, models.ReportStatus, models.ViolationRule
- **Databases**: reports, users, violation_rules

### 13. Chi tiết báo cáo

- **Name**: GetReportDetail
- **Description**:
  1. Admin gọi API chi tiết báo cáo với mã báo cáo.
  2. Kiểm tra người gọi có quyền admin (SuperAdmin hoặc Admin); thiếu thông tin người gọi thì báo lỗi truy cập.
  3. Lấy thông tin báo cáo; nếu không lấy được thì báo lỗi, kể cả khi báo cáo chưa tồn tại.
  4. Nếu báo cáo đang chờ xử lý thì chuyển sang trạng thái đã xem để đánh dấu đã có người mở xem.
  5. Lấy thông tin người báo cáo (tên người dùng, email); lỗi ở bước này sẽ làm thao tác thất bại.
  6. Nếu báo cáo gắn với một quy tắc vi phạm thì lấy tên quy tắc đó, lỗi thì bỏ qua.
  7. Nếu đối tượng bị báo cáo là bài viết thì lấy bài viết, ghi rõ loại đối tượng là bài viết, kèm chủ bài, đoạn trích 200 ký tự, danh sách media, số lượt thích và số bình luận; nếu bài viết đã biến mất vẫn ghi loại đối tượng là bài viết.
  8. Nếu đối tượng bị báo cáo là người dùng thì lấy thông tin người dùng kèm hồ sơ (tên người dùng, tên hiển thị, ảnh đại diện, trạng thái).
  9. Nếu đối tượng bị báo cáo là bình luận thì lấy bình luận và tiêu đề bài viết gốc; lỗi thì chỉ ghi loại đối tượng là bình luận.
  10. Trả về chi tiết báo cáo, loại đối tượng mặc định là chưa xác định; đây là thao tác có ghi dữ liệu (đổi trạng thái báo cáo từ chờ xử lý sang đã xem).
- **Entity**: Admin (người xem), Người báo cáo, Chủ nội dung bị báo cáo, Hệ thống
- **Model**: models.Report, models.ReportStatus, models.ViolationRule, models.Post, models.Comment, models.User, models.Profile, models.Media
- **Databases**: reports, violation_rules, posts, comments, users, profiles, media

### 14. Xử lý (ra quyết định) báo cáo

- **Name**: ReviewReport
- **Description**:
  1. SuperAdmin gọi API với mã báo cáo, hành động xử lý, lý do và thời hạn (nếu có).
  2. Kiểm tra người gọi là SuperAdmin, lấy báo cáo và chặn nếu báo cáo đã bị từ chối hoặc đã được xử lý.
  3. Kiểm tra hành động phải thuộc huỷ, ẩn hoặc cấm — sai thì báo hành động không hợp lệ; riêng ẩn và cấm bắt buộc phải có lý do.
  4. Nhánh huỷ: chỉ đổi báo cáo sang trạng thái bị từ chối.
  5. Nhánh ẩn áp dụng cho bài viết: đổi bài viết sang bị ẩn, đổi báo cáo sang đã xử lý và ghi nhật ký kiểm duyệt cho bài viết.
  6. Nhánh ẩn áp dụng cho người dùng: báo lỗi hành động không hợp lệ, vì phải dùng hành động cấm.
  7. Nhánh ẩn áp dụng cho bình luận: kiểm tra bình luận còn tồn tại, đổi bình luận sang bị ẩn kèm lý do xét duyệt, tìm các bình luận con rồi ẩn hàng loạt (lỗi bỏ qua), đổi báo cáo sang đã xử lý và ghi nhật ký kiểm duyệt.
  8. Nhánh ẩn mà báo cáo không trỏ tới đối tượng nào: báo lỗi hành động không hợp lệ.
  9. Nhánh cấm: bắt buộc báo cáo phải trỏ tới người dùng, tính hạn cấm theo thời hạn đã cho (sai thì báo thời hạn không hợp lệ), đổi người dùng sang bị cấm và vô hiệu phiên đăng nhập, tạo bản ghi cấm, đổi báo cáo sang đã xử lý và ghi nhật ký kiểm duyệt.
  10. Chốt trạng thái báo cáo vào database rồi trả về thông báo xử lý báo cáo thành công.
  11. Gửi thông báo cho người báo cáo mô tả đối tượng bị báo cáo bằng tên hoặc tựa đề, lỗi gửi thì bỏ qua.
  12. Gửi thông báo cho chủ nội dung bị ẩn hoặc bị xem xét (bài viết, bình luận hoặc người dùng, riêng nhánh cấm tránh gửi trùng); nếu là cấm thì gửi thêm thông báo cấm riêng cho người dùng, dùng lý do mặc định khi không có lý do.
- **Entity**: SuperAdmin (người ra quyết định), Người báo cáo, Chủ bài/bình luận/người dùng bị báo cáo, Hệ thống
- **Model**: models.Report, models.ReportStatus, models.Post, models.Comment, models.Ban, models.User, models.ModerationLog, models.ModerationAction, models.NotificationTypeMessage
- **Databases**: reports, posts, comments, bans, users, user_roles, moderation_logs, notifications

### 15. Media nhóm theo người dùng

- **Name**: ListMediaGroupedByUser
- **Description**:
  1. Admin gọi API danh sách media theo người dùng kèm từ khóa và phân trang.
  2. Kiểm tra người gọi có quyền admin (SuperAdmin hoặc Admin); thiếu thông tin người gọi thì báo lỗi truy cập.
  3. Chuẩn hoá phân trang: trang mặc định 1, kích thước trang mặc định 20, tối đa 100.
  4. Truy vấn media đã được duyệt, đếm số người dùng khác nhau có media làm tổng, phân trang theo từng người dùng, tìm theo tên người dùng hoặc tên hiển thị, rồi nạp media và thông tin hồ sơ của từng nhóm người dùng.
  5. Ghép kết quả thành từng nhóm người dùng kèm danh sách media của họ, thời gian tạo định dạng chuẩn.
  6. Trả về danh sách nhóm đã phân trang kèm tổng số; chỉ đọc, không ghi dữ liệu.
- **Entity**: Admin (người xem), Người dùng chủ media
- **Model**: models.Media, models.User, models.Profile
- **Databases**: media, users, profiles

### 16. Media bị gắn cờ

- **Name**: ListFlaggedMedia
- **Description**:
  1. Admin gọi API danh sách media bị gắn cờ kèm bộ lọc trạng thái, từ khóa và phân trang.
  2. Kiểm tra người gọi có quyền admin (SuperAdmin hoặc Admin), rồi chuẩn hoá phân trang.
  3. Chuẩn hoá bộ lọc trạng thái: để trống hoặc chọn tất cả thì không lọc, ngược lại theo trạng thái hợp lệ; mặc định danh sách này lấy media bị gắn cờ.
  4. Đếm tổng rồi tìm danh sách theo trang, lọc theo trạng thái, tìm theo tên người dùng hoặc tên hiển thị, sắp xếp mới nhất trước.
  5. Nạp thông tin chủ media của từng người (tên người dùng, tên hiển thị); lỗi thì bỏ qua và tiếp tục.
  6. Ghép kết quả thành danh sách media với thời gian tạo định dạng chuẩn.
  7. Trả về danh sách media kèm tổng số và trang hiện tại; chỉ đọc, không ghi dữ liệu.
- **Entity**: Admin (người xem), Người dùng chủ media
- **Model**: models.Media, models.MediaStatus, models.User, models.Profile
- **Databases**: media, users, profiles

### 17. Duyệt media (approve/reject)

- **Name**: ReviewMedia
- **Description**:
  1. Admin gọi API với mã media, hành động (duyệt hoặc từ chối) và lý do.
  2. Kiểm tra người gọi có quyền admin, rồi kiểm tra media có tồn tại — không có thì báo không tìm thấy.
  3. Chuyển hành động thành trạng thái tương ứng: duyệt thì thành đã duyệt, từ chối thì thành bị từ chối; giá trị khác thì báo hành động không hợp lệ.
  4. Kiểm tra chuyển trạng thái hợp lệ: chỉ cho phép media đang bị gắn cờ được duyệt hoặc từ chối; trạng thái hiện tại không hỗ trợ hoặc trạng thái đích không được phép thì báo trạng thái không hợp lệ.
  5. Ghi trạng thái mới và lý do xét duyệt vào database.
  6. Ghi nhật ký kiểm duyệt với hành động cập nhật, đối tượng là media và lý do; lỗi ghi sẽ làm thao tác thất bại ngay.
  7. Gửi thông báo cho chủ media theo loại thông báo tương ứng với kết quả xét duyệt, kèm mã bài viết; lỗi gửi sẽ làm thao tác thất bại.
  8. Nếu là từ chối thì xoá file trên nền tảng lưu trữ đám mây (nếu chưa cấu hình thì chỉ ghi cảnh báo), sau đó xoá liên kết file trong database; lỗi xoá file sẽ báo xoá file khỏi nền tảng lưu trữ thất bại.
  9. Trả về thông báo xử lý media thành công.
- **Entity**: Admin (người duyệt), Người dùng chủ media, Hệ thống (Cloudinary + moderation log)
- **Model**: models.Media, models.MediaStatus, models.ModerationLog, models.ModerationAction, models.NotificationType
- **Databases**: media, moderation_logs, notifications

### 18. Dọn dẹp media đã reject

- **Name**: CleanupRejectedMedia
- **Description**:
  1. Admin gọi API dọn dẹp media, không cần dữ liệu đầu vào.
  2. Kiểm tra người gọi có quyền admin (SuperAdmin hoặc Admin).
  3. Lấy toàn bộ media bị từ chối từ quá 7 ngày, tính mốc 7 ngày trước so với thời điểm hiện tại.
  4. Nếu không có gì để dọn thì trả về số lượng đã dọn bằng 0.
  5. Với từng media: nếu đã cấu hình nền tảng lưu trữ đám mây thì xoá file tương ứng; lỗi chỉ ghi log, và bỏ qua khi chưa cấu hình hoặc không đọc được liên kết file.
  6. Xoá bản ghi media trong database và cộng lại dung lượng lưu trữ cho chủ media trong cùng một giao dịch, dung lượng không được âm; lỗi thì ghi log và không tính vào số đã dọn.
  7. Trả về thông báo dọn dẹp media thành công kèm số bản ghi đã dọn.
- **Entity**: Admin (người chạy), Người dùng chủ media (storage được cộng lại), Hệ thống (Cloudinary storage)
- **Model**: models.Media, models.MediaStatus, models.User
- **Databases**: media, users

## Admin Endpoints — Quản lý nhóm chat & cộng đồng

Nguồn: `routes/admin.routes.go` (group `/api/admin`, `middlewares.AuthMiddleware` + `middlewares.RequireRoles(db, models.RoleSuperAdmin, models.RoleAdmin)`), `controllers/admin.controller.go`, `services/admin.service.go`.

> Các endpoint GET dùng helper `ensureAdmin` (SuperAdmin **hoặc** Admin). Các endpoint thao tác (hide/unhide/archive/warn/delete) dùng helper `ensureSuperAdmin` (chỉ SuperAdmin) — RBAC middleware chặn trước, service chặn thêm lần nữa.
> Endpoint 34/35 (`archive`/`unarchive`) được đăng ký theo thứ tự `unarchive` trước `archive` trong routes; đường dẫn vẫn đầy đủ `POST /api/admin/communities/:id/archive` và `POST /api/admin/communities/:id/unarchive`.

### 19. Danh sách nhóm chat

- **Name**: ListGroups
- **Description**:
  1. Admin gọi API lấy danh sách nhóm chat, truyền bộ lọc từ khóa, trạng thái, số trang và kích thước trang.
  2. Hệ thống kiểm tra tham số truy vấn hợp lệ, không hợp lệ thì báo lỗi.
  3. Kiểm tra người gọi có quyền quản trị nền tảng (quản trị viên cao cấp hoặc quản trị viên), không đủ quyền thì báo lỗi truy cập bị từ chối.
  4. Chuẩn hóa phân trang (mặc định trang 1, 20 mục mỗi trang, tối đa 100) và kiểm tra trạng thái phải thuộc đang hoạt động, bị ẩn hoặc đã lưu trữ.
  5. Truy vấn danh sách nhóm chat khớp bộ lọc, kèm số thành viên và tên người tạo của từng nhóm.
  6. Đếm tổng số nhóm chat khớp bộ lọc để phục vụ phân trang.
  7. Trả về danh sách nhóm chat đã phân trang kèm tổng số, số trang và kích thước trang.
- **Entity**: Admin (người xem), Hệ thống, Nhóm chat (đối tượng được liệt kê)
- **Model**: models.Chat, models.ChatTypeGroup, models.ChatStatus (ChatStatusActive, ChatStatusHidden, ChatStatusArchived)
- **Databases**: roles, user_roles, chats, chat_participants, profiles

### 20. Chi tiết nhóm chat

- **Name**: GetGroupDetail
- **Description**:
  1. Admin gọi API xem chi tiết nhóm chat, truyền ID nhóm chat trong đường dẫn.
  2. Kiểm tra ID không được để trống và người gọi có quyền quản trị nền tảng.
  3. Tìm nhóm chat theo ID; không tìm thấy thì báo lỗi không tồn tại, nếu không phải nhóm chat (là cuộc trò chuyện 1-1) thì báo lỗi đối tượng không phải nhóm chat.
  4. Lấy danh sách thành viên của nhóm theo thời gian tham gia, kèm vai trò, tên hiển thị và ảnh đại diện.
  5. Nếu nhóm có chủ nhóm thì lấy tên hiển thị của chủ nhóm, không lấy được thì để trống tên.
  6. Trả về thông tin nhóm chat, số thành viên và danh sách thành viên.
- **Entity**: Admin (người xem), Nhóm chat, Chủ nhóm, Thành viên nhóm
- **Model**: models.Chat, models.ChatTypeGroup, models.Profile
- **Databases**: roles, user_roles, chats, chat_participants, profiles

### 21. Danh sách thành viên nhóm chat

- **Name**: ListGroupMembers
- **Description**:
  1. Admin gọi API lấy danh sách thành viên nhóm chat, truyền ID nhóm chat trong đường dẫn.
  2. Kiểm tra ID không được để trống và người gọi có quyền quản trị nền tảng.
  3. Kiểm tra nhóm chat tồn tại và đúng là nhóm chat, không có hoặc sai loại thì báo lỗi tương ứng.
  4. Lấy toàn bộ thành viên của nhóm theo thời gian tham gia, kèm vai trò, tên hiển thị và ảnh đại diện.
  5. Trả về danh sách thành viên của nhóm (không phân trang).
- **Entity**: Admin (người xem), Nhóm chat, Thành viên nhóm
- **Model**: models.Chat, models.ChatTypeGroup, models.ChatParticipant, models.Profile
- **Databases**: roles, user_roles, chats, chat_participants, profiles

### 22. Nhật ký kiểm duyệt của nhóm chat

- **Name**: GetGroupModerationLogs
- **Description**:
  1. Admin gọi API xem nhật ký kiểm duyệt của nhóm chat, truyền ID nhóm trong đường dẫn và số trang, kích thước trang.
  2. Kiểm tra tham số truy vấn hợp lệ và người gọi có quyền quản trị nền tảng.
  3. Chuẩn hóa phân trang (mặc định 20 mục mỗi trang, tối đa 100).
  4. Lấy các bản ghi kiểm duyệt của nhóm theo thời gian giảm dần, kèm tên người kiểm duyệt đã ghi bản ghi đó.
  5. Đếm tổng số bản ghi kiểm duyệt của nhóm để phục vụ phân trang.
  6. Trả về danh sách nhật ký kiểm duyệt đã phân trang kèm tổng số, số trang và kích thước trang.
- **Entity**: Admin (người xem), Nhóm chat, Người kiểm duyệt (moderator) từng ghi log
- **Model**: models.ModerationLog, models.ModerationTargetGroupChat, models.Profile
- **Databases**: roles, user_roles, moderation_logs, profiles

### 23. Ẩn nhóm chat

- **Name**: HideGroup
- **Description**:
  1. Admin gọi API ẩn nhóm chat, truyền ID nhóm trong đường dẫn và lý do ẩn trong thân yêu cầu; thiếu lý do thì báo lỗi.
  2. Kiểm tra người gọi là quản trị viên cao cấp, không đủ quyền thì báo lỗi không đủ quyền.
  3. Tìm nhóm chat theo ID và kiểm tra đúng là nhóm chat; không tìm thấy hoặc sai loại thì báo lỗi tương ứng.
  4. Kiểm tra trạng thái chuyển đổi: nhóm đang lưu trữ thì không được ẩn, nhóm đã ở trạng thái bị ẩn rồi thì báo lỗi đã ở trạng thái này.
  5. Ghi nhật ký kiểm duyệt với hành động ẩn, đối tượng là nhóm chat và lý do từ yêu cầu.
  6. Cập nhật trạng thái của nhóm thành bị ẩn.
  7. Nếu nhóm có chủ nhóm thì gửi thông báo cho chủ nhóm biết nhóm đã bị ẩn kèm lý do; quá trình gửi kiểm tra tùy chọn thông báo, phát realtime và gửi push nếu có thiết bị đăng ký, lỗi gửi bị bỏ qua.
  8. Trả về thông báo ẩn nhóm chat thành công.
- **Entity**: Admin/SuperAdmin (người ra lệnh), Nhóm chat, Chủ nhóm (nhận notification), Người nhận notification/push
- **Model**: models.Chat, models.ChatStatusHidden, models.ModerationLog, models.ModerationActionHide, models.ModerationTargetGroupChat, models.Notification, models.NotificationTypeMessage, models.RoleSuperAdmin
- **Databases**: roles, user_roles, chats, moderation_logs, notification_preferences, notifications, push_tokens

### 24. Bỏ ẩn nhóm chat

- **Name**: UnhideGroup
- **Description**:
  1. Admin gọi API bỏ ẩn nhóm chat, chỉ truyền ID nhóm trong đường dẫn (không có thân yêu cầu, lý do được dùng là bỏ ẩn nhóm).
  2. Kiểm tra người gọi là quản trị viên cao cấp.
  3. Tìm nhóm chat theo ID và kiểm tra đúng là nhóm chat; không tìm thấy hoặc sai loại thì báo lỗi tương ứng.
  4. Kiểm tra trạng thái chuyển đổi: nhóm đang lưu trữ thì không được bỏ ẩn trực tiếp, nhóm đang hoạt động rồi thì báo lỗi đã ở trạng thái này.
  5. Ghi nhật ký kiểm duyệt với hành động cập nhật và lý do là bỏ ẩn nhóm.
  6. Cập nhật trạng thái của nhóm thành đang hoạt động.
  7. Gửi thông báo cho chủ nhóm (nếu có) về việc nhóm được bỏ ẩn; phát realtime và gửi push, lỗi gửi bị bỏ qua.
  8. Trả về thông báo bỏ ẩn nhóm chat thành công.
- **Entity**: Admin/SuperAdmin (người ra lệnh), Nhóm chat, Chủ nhóm (nhận notification)
- **Model**: models.Chat, models.ChatStatusActive, models.ModerationLog, models.ModerationActionUpdate, models.ModerationTargetGroupChat, models.Notification, models.NotificationTypeMessage, models.RoleSuperAdmin
- **Databases**: roles, user_roles, chats, moderation_logs, notification_preferences, notifications, push_tokens

### 25. Đình chỉ nhóm chat

- **Name**: ArchiveGroup
- **Description**:
  1. Admin gọi API đình chỉ nhóm chat, truyền ID nhóm trong đường dẫn và lý do đình chỉ; thiếu lý do thì báo lỗi.
  2. Kiểm tra người gọi là quản trị viên cao cấp.
  3. Tìm nhóm chat theo ID và kiểm tra đúng là nhóm chat; không tìm thấy hoặc sai loại thì báo lỗi tương ứng.
  4. Kiểm tra trạng thái chuyển đổi: nhóm đang lưu trữ thì không được chuyển sang trạng thái khác, nhóm đã ở trạng thái đình chỉ rồi thì báo lỗi đã ở trạng thái này.
  5. Ghi nhật ký kiểm duyệt với hành động đình chỉ và lý do từ yêu cầu.
  6. Cập nhật trạng thái của nhóm thành đã lưu trữ; từ đây các thao tác chuyển trạng thái khác của nhóm sẽ bị chặn.
  7. Gửi thông báo cho chủ nhóm (nếu có) biết nhóm đã bị đình chỉ kèm lý do; phát realtime và gửi push, lỗi gửi bị bỏ qua.
  8. Trả về thông báo đình chỉ nhóm chat thành công.
- **Entity**: Admin/SuperAdmin (người ra lệnh), Nhóm chat, Chủ nhóm (nhận notification)
- **Model**: models.Chat, models.ChatStatusArchived, models.ModerationLog, models.ModerationActionSuspend, models.ModerationTargetGroupChat, models.Notification, models.NotificationTypeMessage, models.RoleSuperAdmin
- **Databases**: roles, user_roles, chats, moderation_logs, notification_preferences, notifications, push_tokens

### 26. Cảnh báo nhóm chat

- **Name**: WarnGroup
- **Description**:
  1. Admin gọi API cảnh báo nhóm chat, truyền ID nhóm trong đường dẫn, lý do và nội dung cảnh báo; thiếu một trong hai thì báo lỗi.
  2. Kiểm tra người gọi là quản trị viên cao cấp.
  3. Tìm nhóm chat theo ID và kiểm tra đúng là nhóm chat; không tìm thấy hoặc sai loại thì báo lỗi tương ứng.
  4. Ghi nhật ký kiểm duyệt với hành động cảnh báo và lý do từ yêu cầu.
  5. Không thay đổi trạng thái của nhóm — cảnh báo chỉ ghi nhật ký, không có kiểm tra hay cập nhật trạng thái.
  6. Nếu nhóm có chủ nhóm thì gửi thông báo cho chủ nhóm với nội dung cảnh báo kèm tên nhóm và lời nhắn; phát realtime và gửi push, lỗi gửi bị bỏ qua.
  7. Trả về thông báo cảnh báo nhóm chat thành công.
- **Entity**: Admin/SuperAdmin (người ra lệnh), Nhóm chat, Chủ nhóm (người bị cảnh báo/nhận notification)
- **Model**: models.Chat, models.ChatTypeGroup, models.ModerationLog, models.ModerationActionWarn, models.ModerationTargetGroupChat, models.Notification, models.NotificationTypeMessage, models.RoleSuperAdmin
- **Databases**: roles, user_roles, chats, moderation_logs, notification_preferences, notifications, push_tokens

### 27. Xóa nhóm chat

- **Name**: DeleteGroup
- **Description**:
  1. Admin gọi API xóa nhóm chat, truyền ID nhóm trong đường dẫn và lý do xóa; thiếu lý do thì báo lỗi.
  2. Kiểm tra người gọi là quản trị viên cao cấp, chỉ quản trị viên cao cấp mới được xóa.
  3. Tìm nhóm chat theo ID, không tìm thấy thì báo lỗi không tồn tại; nếu không phải nhóm chat thì báo lỗi đối tượng không phải nhóm chat.
  4. Lấy danh sách ID thành viên hiện có của nhóm.
  5. Kiểm tra an toàn: nhóm còn nhiều hơn một thành viên thì từ chối, chỉ cho phép xóa nhóm đã bị giải tán xuống còn một người.
  6. Ghi nhật ký kiểm duyệt với hành động xóa và lý do từ yêu cầu.
  7. Tiến hành xóa trong một giao dịch theo thứ tự: trước hết xóa tin nhắn của nhóm, rồi xóa thành viên, cuối cùng xóa chính nhóm; từng bước thất bại đều báo lỗi có ngữ cảnh.
  8. Trả về thông báo xóa nhóm chat thành công; không gửi thông báo cho thành viên.
- **Entity**: Admin/SuperAdmin (người ra lệnh), Nhóm chat bị xóa, Thành viên/nhận tin nhắn của nhóm
- **Model**: models.Chat, models.ChatTypeGroup, models.Message, models.ChatParticipant, models.ModerationLog, models.ModerationActionDelete, models.ModerationTargetGroupChat, models.RoleSuperAdmin
- **Databases**: roles, user_roles, chats, chat_participants, messages, moderation_logs

### 28. Danh sách cộng đồng

- **Name**: ListCommunities
- **Description**:
  1. Admin gọi API lấy danh sách cộng đồng, truyền bộ lọc từ khóa, trạng thái, quyền riêng tư, số trang và kích thước trang.
  2. Hệ thống kiểm tra tham số truy vấn hợp lệ, không hợp lệ thì báo lỗi.
  3. Kiểm tra người gọi có quyền quản trị nền tảng (quản trị viên cao cấp hoặc quản trị viên), không đủ quyền thì báo lỗi truy cập bị từ chối.
  4. Chuẩn hóa phân trang (mặc định trang 1, 20 mục mỗi trang, tối đa 100) và kiểm tra trạng thái phải thuộc đang hoạt động, bị ẩn hoặc đã lưu trữ, quyền riêng tư phải thuộc công khai hoặc chỉ qua lời mời.
  5. Truy vấn danh sách cộng đồng khớp bộ lọc, kèm số thành viên và tên người tạo của từng cộng đồng.
  6. Đếm tổng số cộng đồng khớp bộ lọc để phục vụ phân trang.
  7. Trả về danh sách cộng đồng đã phân trang kèm tổng số, số trang và kích thước trang.
- **Entity**: Admin (người xem), Hệ thống, Cộng đồng (đối tượng được liệt kê)
- **Model**: models.Community, models.CommunityStatus (Active/Hidden/Archived), models.PrivacyPublic, models.PrivacyInvitationOnly
- **Databases**: roles, user_roles, communities, group_members, profiles

### 29. Chi tiết cộng đồng

- **Name**: GetCommunityDetail
- **Description**:
  1. Admin gọi API xem chi tiết cộng đồng, truyền ID cộng đồng trong đường dẫn.
  2. Kiểm tra ID không được để trống và người gọi có quyền quản trị nền tảng.
  3. Tìm cộng đồng theo ID, không tìm thấy thì báo lỗi không tồn tại.
  4. Lấy danh sách thành viên cộng đồng kèm hồ sơ của họ, suy ra vai trò hiển thị trong cộng đồng theo thứ tự ưu tiên từ cao xuống, sắp xếp theo thời gian tham gia.
  5. Lấy tên hiển thị của chủ cộng đồng, không lấy được thì để trống tên.
  6. Trả về thông tin cộng đồng (tên, mô tả, ảnh đại diện, ảnh bìa, quyền riêng tư, trạng thái, tự động duyệt thành viên, thời gian tạo và cập nhật), số thành viên và danh sách thành viên.
- **Entity**: Admin (người xem), Cộng đồng, Chủ cộng đồng, Thành viên
- **Model**: models.Community, models.GroupMember, models.UserRole, models.Role, models.Profile
- **Databases**: roles, user_roles, communities, group_members, profiles

### 30. Danh sách thành viên cộng đồng

- **Name**: ListCommunityMembers
- **Description**:
  1. Admin gọi API lấy danh sách thành viên cộng đồng, truyền ID cộng đồng trong đường dẫn.
  2. Kiểm tra ID không được để trống và người gọi có quyền quản trị nền tảng.
  3. Kiểm tra cộng đồng tồn tại, không có thì báo lỗi không tồn tại.
  4. Lấy toàn bộ thành viên cộng đồng kèm hồ sơ và vai trò trong cộng đồng, sắp xếp theo thời gian tham gia.
  5. Trả về danh sách thành viên (không phân trang).
- **Entity**: Admin (người xem), Cộng đồng, Thành viên cộng đồng
- **Model**: models.Community, models.GroupMember, models.UserRole, models.Role, models.Profile
- **Databases**: roles, user_roles, communities, group_members, profiles

### 31. Nhật ký kiểm duyệt của cộng đồng

- **Name**: GetCommunityModerationLogs
- **Description**:
  1. Admin gọi API xem nhật ký kiểm duyệt của cộng đồng, truyền ID cộng đồng trong đường dẫn và số trang, kích thước trang.
  2. Kiểm tra tham số truy vấn hợp lệ và người gọi có quyền quản trị nền tảng.
  3. Chuẩn hóa phân trang (mặc định 20 mục mỗi trang, tối đa 100).
  4. Lấy các bản ghi kiểm duyệt của cộng đồng theo thời gian giảm dần, kèm tên người kiểm duyệt đã ghi bản ghi đó.
  5. Đếm tổng số bản ghi kiểm duyệt của cộng đồng để phục vụ phân trang.
  6. Trả về danh sách nhật ký kiểm duyệt đã phân trang kèm tổng số, số trang và kích thước trang.
- **Entity**: Admin (người xem), Cộng đồng, Người kiểm duyệt từng ghi log
- **Model**: models.ModerationLog, models.ModerationTargetCommunity, models.Profile
- **Databases**: roles, user_roles, moderation_logs, profiles

### 32. Ẩn cộng đồng

- **Name**: HideCommunity
- **Description**:
  1. Admin gọi API ẩn cộng đồng, truyền ID cộng đồng trong đường dẫn và lý do ẩn trong thân yêu cầu; thiếu lý do thì báo lỗi.
  2. Kiểm tra người gọi là quản trị viên cao cấp, không đủ quyền thì báo lỗi không đủ quyền.
  3. Tìm cộng đồng theo ID, không tìm thấy thì báo lỗi không tồn tại.
  4. Kiểm tra trạng thái chuyển đổi: cộng đồng đang lưu trữ thì không được chuyển sang bị ẩn, cộng đồng đã ở trạng thái bị ẩn rồi thì báo lỗi đã ở trạng thái này.
  5. Ghi nhật ký kiểm duyệt với hành động ẩn, đối tượng là cộng đồng và lý do từ yêu cầu.
  6. Cập nhật trạng thái của cộng đồng thành bị ẩn.
  7. Gửi thông báo cho chủ cộng đồng (cộng đồng luôn có chủ) biết cộng đồng đã bị ẩn kèm lý do; quá trình gửi kiểm tra tùy chọn thông báo, phát realtime và gửi push, lỗi gửi bị bỏ qua.
  8. Trả về thông báo ẩn cộng đồng thành công.
- **Entity**: Admin/SuperAdmin (người ra lệnh), Cộng đồng, Chủ cộng đồng (nhận notification)
- **Model**: models.Community, models.CommunityStatusHidden, models.ModerationLog, models.ModerationActionHide, models.ModerationTargetCommunity, models.Notification, models.NotificationTypeMessage, models.RoleSuperAdmin
- **Databases**: roles, user_roles, communities, moderation_logs, notification_preferences, notifications, push_tokens

### 33. Bỏ ẩn cộng đồng

- **Name**: UnhideCommunity
- **Description**:
  1. Admin gọi API bỏ ẩn cộng đồng, chỉ truyền ID cộng đồng trong đường dẫn (không có thân yêu cầu, lý do được dùng là bỏ ẩn cộng đồng).
  2. Kiểm tra người gọi là quản trị viên cao cấp.
  3. Tìm cộng đồng theo ID, không tìm thấy thì báo lỗi không tồn tại.
  4. Kiểm tra trạng thái chuyển đổi: cộng đồng đang lưu trữ thì bỏ ẩn về đang hoạt động là hợp lệ, cộng đồng đang hoạt động rồi thì báo lỗi đã ở trạng thái này.
  5. Ghi nhật ký kiểm duyệt với hành động cập nhật và lý do là bỏ ẩn cộng đồng.
  6. Cập nhật trạng thái của cộng đồng thành đang hoạt động.
  7. Gửi thông báo cho chủ cộng đồng về việc cộng đồng được bỏ ẩn; phát realtime và gửi push, lỗi gửi bị bỏ qua.
  8. Trả về thông báo bỏ ẩn cộng đồng thành công.
- **Entity**: Admin/SuperAdmin (người ra lệnh), Cộng đồng, Chủ cộng đồng (nhận notification)
- **Model**: models.Community, models.CommunityStatusActive, models.ModerationLog, models.ModerationActionUpdate, models.ModerationTargetCommunity, models.Notification, models.NotificationTypeMessage, models.RoleSuperAdmin
- **Databases**: roles, user_roles, communities, moderation_logs, notification_preferences, notifications, push_tokens

### 34. Đình chỉ cộng đồng

- **Name**: ArchiveCommunity
- **Description**:
  1. Admin gọi API đình chỉ cộng đồng, truyền ID cộng đồng trong đường dẫn và lý do đình chỉ; thiếu lý do thì báo lỗi.
  2. Kiểm tra người gọi là quản trị viên cao cấp.
  3. Tìm cộng đồng theo ID, không tìm thấy thì báo lỗi không tồn tại.
  4. Kiểm tra trạng thái chuyển đổi: cộng đồng đang lưu trữ mà đình chỉ lần nữa thì báo lỗi đã ở trạng thái này, trạng thái khác thì tiến hành bình thường.
  5. Ghi nhật ký kiểm duyệt với hành động đình chỉ và lý do từ yêu cầu.
  6. Cập nhật trạng thái của cộng đồng thành đã lưu trữ.
  7. Gửi thông báo cho chủ cộng đồng biết cộng đồng đã bị đình chỉ kèm lý do; quá trình gửi kiểm tra tùy chọn thông báo, phát realtime và gửi push, lỗi gửi bị bỏ qua.
  8. Trả về thông báo đình chỉ cộng đồng thành công.
- **Entity**: Admin/SuperAdmin (người ra lệnh), Cộng đồng, Chủ cộng đồng (nhận notification)
- **Model**: models.Community, models.CommunityStatusArchived, models.ModerationLog, models.ModerationActionSuspend, models.ModerationTargetCommunity, models.Notification, models.NotificationTypeMessage, models.RoleSuperAdmin
- **Databases**: roles, user_roles, communities, moderation_logs, notification_preferences, notifications, push_tokens

### 35. Bỏ đình chỉ cộng đồng

- **Name**: UnarchiveCommunity
- **Description**:
  1. Admin gọi API bỏ đình chỉ cộng đồng, chỉ truyền ID cộng đồng trong đường dẫn (không có thân yêu cầu, lý do được dùng là bỏ đình chỉ cộng đồng).
  2. Kiểm tra người gọi là quản trị viên cao cấp.
  3. Tìm cộng đồng theo ID, không tìm thấy thì báo lỗi không tồn tại.
  4. Kiểm tra trạng thái chuyển đổi: cộng đồng đang lưu trữ chuyển về đang hoạt động là được phép, cộng đồng đang hoạt động rồi thì báo lỗi đã ở trạng thái này.
  5. Ghi nhật ký kiểm duyệt với hành động cập nhật và lý do là bỏ đình chỉ cộng đồng.
  6. Cập nhật trạng thái của cộng đồng thành đang hoạt động.
  7. Gửi thông báo cho chủ cộng đồng về việc cộng đồng được bỏ đình chỉ; phát realtime và gửi push, lỗi gửi bị bỏ qua.
  8. Trả về thông báo bỏ đình chỉ cộng đồng thành công.
- **Entity**: Admin/SuperAdmin (người ra lệnh), Cộng đồng, Chủ cộng đồng (nhận notification)
- **Model**: models.Community, models.CommunityStatusActive, models.ModerationLog, models.ModerationActionUpdate, models.ModerationTargetCommunity, models.Notification, models.NotificationTypeMessage, models.RoleSuperAdmin
- **Databases**: roles, user_roles, communities, moderation_logs, notification_preferences, notifications, push_tokens

### 36. Cảnh báo cộng đồng

- **Name**: WarnCommunity
- **Description**:
  1. Admin gọi API cảnh báo cộng đồng, truyền ID cộng đồng trong đường dẫn, lý do và nội dung cảnh báo; thiếu một trong hai thì báo lỗi.
  2. Kiểm tra người gọi là quản trị viên cao cấp.
  3. Tìm cộng đồng theo ID, không tìm thấy thì báo lỗi không tồn tại.
  4. Ghi nhật ký kiểm duyệt với hành động cảnh báo và lý do từ yêu cầu.
  5. Không thay đổi trạng thái của cộng đồng — cảnh báo chỉ ghi nhật ký, không có kiểm tra hay cập nhật trạng thái.
  6. Gửi thông báo cho chủ cộng đồng với nội dung cảnh báo kèm tên cộng đồng và lời nhắn; quá trình gửi kiểm tra tùy chọn thông báo, phát realtime và gửi push, lỗi gửi bị bỏ qua.
  7. Trả về thông báo cảnh báo cộng đồng thành công.
- **Entity**: Admin/SuperAdmin (người ra lệnh), Cộng đồng, Chủ cộng đồng (người bị cảnh báo/nhận notification)
- **Model**: models.Community, models.ModerationLog, models.ModerationActionWarn, models.ModerationTargetCommunity, models.Notification, models.NotificationTypeMessage, models.RoleSuperAdmin
- **Databases**: roles, user_roles, communities, moderation_logs, notification_preferences, notifications, push_tokens

### 37. Xóa cộng đồng

- **Name**: DeleteCommunity
- **Description**:
  1. Admin gọi API xóa cộng đồng, truyền ID cộng đồng trong đường dẫn và lý do xóa; thiếu lý do thì báo lỗi.
  2. Kiểm tra người gọi là quản trị viên cao cấp, chỉ quản trị viên cao cấp mới được xóa.
  3. Tìm cộng đồng theo ID, không tìm thấy thì báo lỗi không tồn tại; cộng đồng không có chủ cộng đồng thì báo lỗi thiếu chủ cộng đồng.
  4. Lấy danh sách ID thành viên hiện có của cộng đồng.
  5. Kiểm tra an toàn: cộng đồng còn nhiều hơn một thành viên thì từ chối, chỉ cho phép xóa cộng đồng đã bị giải tán về đúng một người là chủ cộng đồng.
  6. Ghi nhật ký kiểm duyệt với hành động xóa và lý do từ yêu cầu.
  7. Tiến hành xóa trong một giao dịch: gỡ vai trò trong cộng đồng của chủ cộng đồng, rồi xóa tư cách thành viên của người đó; không xóa bản ghi cộng đồng gốc, việc xóa cộng đồng thực chất là giải tán thành viên cuối cùng và ghi nhật ký.
  8. Trả về thông báo xóa cộng đồng thành công; không gửi thông báo cho thành viên.
- **Entity**: Admin/SuperAdmin (người ra lệnh), Cộng đồng, Chủ cộng đồng (thành viên cuối bị gỡ), Thành viên cộng đồng
- **Model**: models.Community, models.GroupMember, models.UserRole, models.ModerationLog, models.ModerationActionDelete, models.ModerationTargetCommunity, models.RoleSuperAdmin
- **Databases**: roles, user_roles, communities, group_members, moderation_logs

## A3 — Admin Ads / Site Settings / Violation Rules (endpoint 38–53)

Nguồn: `routes/admin.routes.go`, `routes/ad.routes.go`, `routes/violation_rule.routes.go` → `controllers/` → `services/`.

---

### 38. Danh sách quảng cáo (admin)

- **Name**: ListAds
- **Description**:
  1. Admin gọi API với các bộ lọc tùy chọn trên query: từ khóa tìm kiếm, trạng thái quảng cáo, trang và số phần tử mỗi trang.
  2. Kiểm tra người dùng đã đăng nhập và có quyền quản trị (SUPER_ADMIN hoặc ADMIN), không có quyền thì báo lỗi truy cập.
  3. Chuẩn hóa tham số phân trang (trang mặc định 1, số phần tử mặc định 20, tối đa 100), làm sạch từ khóa và chuẩn hóa trạng thái.
  4. Đếm tổng số quảng cáo khớp bộ lọc rồi lấy trang hiện tại, gồm thông tin chủ quảng cáo, ảnh đại diện và số liệu lượt hiển thị, click, tỷ lệ click.
  5. Sắp xếp kết quả theo thời gian tạo giảm dần.
  6. Trả về danh sách quảng cáo đã phân trang kèm tổng số, trang và số phần tử mỗi trang; nếu không có kết quả thì trả thông báo không tìm thấy quảng cáo.
- **Entity**: Admin/SuperAdmin (người duyệt), Partner (chủ quảng cáo), User hiển thị trong danh sách
- **Model**: models.Ad, models.AdMedia, models.AdAnalytics
- **Databases**: ads, ad_media, ad_analytics, users, profiles

### 39. Duyệt / đổi trạng thái quảng cáo (admin)

- **Name**: UpdateAdStatus
- **Description**:
  1. Admin gọi API với id quảng cáo cần duyệt và trạng thái mới (đang chờ, đang hoạt động, tạm dừng, hoàn thành hoặc từ chối), kèm lý do từ chối nếu có.
  2. Kiểm tra quyền quản trị (SUPER_ADMIN hoặc ADMIN).
  3. Tìm quảng cáo theo id, không tồn tại thì báo lỗi không tìm thấy.
  4. Nếu trạng thái mới trùng trạng thái hiện tại thì trả về thành công mà không thay đổi gì.
  5. Chỉ được chuyển quảng cáo đang chờ sang trạng thái đang hoạt động hoặc từ chối; sai trạng thái thì báo lỗi trạng thái không hợp lệ.
  6. Kiểm tra lý do từ chối là bắt buộc khi từ chối quảng cáo, thiếu thì báo lỗi yêu cầu lý do.
  7. Lưu trạng thái mới và lý do từ chối (nếu có) vào database.
  8. Gửi thông báo cho Partner về việc quảng cáo được phê duyệt hoặc từ chối, kèm lý do nếu từ chối; lỗi gửi thông báo được bỏ qua.
  9. Trả về thông báo cập nhật trạng thái quảng cáo thành công.
- **Entity**: Admin/SuperAdmin (người duyệt), Partner (nhận thông báo)
- **Model**: models.Ad, models.Notification
- **Databases**: ads, notifications

### 40. Xoá quảng cáo (admin)

- **Name**: DeleteAd (admin)
- **Description**:
  1. Admin gọi API với id quảng cáo cần xoá.
  2. Kiểm tra người dùng có quyền SUPER_ADMIN, nếu không thì báo lỗi không đủ quyền.
  3. Xoá vĩnh viễn quảng cáo khỏi database; nếu không xoá được thì báo lỗi.
  4. Các file media liên kết với quảng cáo cũng bị xoá theo.
  5. Ở nhánh admin hệ thống không giải phóng slot gói đăng ký (khác với nhánh Partner).
  6. Trả về thông báo xoá quảng cáo thành công.
- **Entity**: SuperAdmin
- **Model**: models.Ad, models.AdMedia
- **Databases**: ads, ad_media

### 41. Đọc cấu hình site settings

- **Name**: GetSettings
- **Description**:
  1. Người dùng gọi API để đọc cấu hình hệ thống.
  2. Kiểm tra đã đăng nhập và có quyền SUPER_ADMIN, không có thì báo lỗi truy cập.
  3. Đọc toàn bộ bản ghi cấu hình từ database, mỗi dòng là một cặp khóa và giá trị.
  4. Gom các khóa và giá trị thành danh sách khóa - chuỗi rồi trả về cho người gọi.
  5. Nếu đọc database gặp lỗi thì báo lỗi đọc cấu hình thất bại.
  6. Trả về toàn bộ cấu hình gồm các khóa hợp lệ như tên trang, chế độ bảo trì, độ dài mật khẩu tối thiểu, thời hạn token.
- **Entity**: SuperAdmin
- **Model**: models.SystemConfig
- **Databases**: system_configs, roles, user_roles

### 42. Cập nhật cấu hình site settings

- **Name**: UpdateSettings
- **Description**:
  1. Người dùng gọi API với danh sách khóa và giá trị cấu hình cần cập nhật; thiếu dữ liệu thì báo lỗi.
  2. Kiểm tra đã đăng nhập và có quyền SUPER_ADMIN.
  3. Kiểm tra từng khóa có nằm trong danh sách khóa được phép cập nhật hay không, khóa lạ thì báo lỗi khóa không hợp lệ.
  4. Kiểm tra giá trị theo kiểu và khoảng cho phép: số nguyên phải trong khoảng hợp lệ, boolean chỉ nhận đúng true hoặc false, email phải đúng định dạng; để trống thì bỏ qua.
  5. Ghi đè hàng loạt các cấu hình vào database.
  6. Nếu bật chế độ bảo trì thì vô hiệu mọi phiên đang phát hành của người dùng thường (trừ SUPER_ADMIN và ADMIN) và xóa cache xác thực; lỗi ở bước này được báo lại.
  7. Trả về thông báo lưu cài đặt thành công, các giá trị mới có hiệu lực ngay với việc đăng nhập và kiểm tra phiên.
- **Entity**: SuperAdmin, toàn bộ người dùng (phiên bị vô hiệu khi bật maintenance)
- **Model**: models.SystemConfig
- **Databases**: system_configs, users, roles, user_roles

### 43. Tạo quảng cáo (ads-management)

- **Name**: CreateAd
- **Description**:
  1. Partner gọi API với form đa phần bao gồm thông tin quảng cáo (tiêu đề, nội dung, link đích, định dạng), các trường số (kinh phí, ngân sách ngày, giá mỗi nghìn hiển thị, giá mỗi click, số lần hiển thị tối đa), ngày bắt đầu và kết thúc, dữ liệu nhắm mục tiêu và file media.
  2. Validate các trường bắt buộc, định dạng ngày, số liệu và dữ liệu nhắm mục tiêu; lỗi thì báo dữ liệu không hợp lệ.
  3. Kiểm tra Partner có gói đăng ký đang hoạt động hay chưa; chưa đăng ký thì báo lỗi chưa có gói, dùng hết slot thì báo lỗi hết chỗ; định dạng video hoặc carousel phải được gói hỗ trợ.
  4. Tạo quảng cáo mới với định danh duy nhất, trỏ tới gói đăng ký, lưu dữ liệu nhắm mục tiêu và trạng thái ban đầu rồi ghi vào database.
  5. Upload từng file media lên dịch vụ lưu trữ đám mây và ghi thông tin file kèm thứ tự cho quảng cáo.
  6. Tăng số slot đã dùng của gói đăng ký.
  7. Nếu người tạo không phải SUPER_ADMIN thì đặt quảng cáo vào trạng thái chờ duyệt; SUPER_ADMIN thì quảng cáo chạy ngay.
  8. Trả về thông báo tạo quảng cáo thành công kèm dữ liệu quảng cáo và danh sách media.
- **Entity**: Partner/Admin/SuperAdmin (người tạo), hệ thống upload Cloudinary, gói subscription
- **Model**: models.Ad, models.AdMedia, models.AdPackage, models.PartnerSubscription, models.Media
- **Databases**: ads, ad_media, partner_subscriptions, ad_packages, media

### 44. Danh sách quảng cáo (ads-management, admin hoặc partner)

- **Name**: GetAdminList
- **Description**:
  1. Người dùng gọi API với tham số trang và số phần tử mỗi trang.
  2. Chuẩn hóa phân trang: trang tối thiểu 1, số phần tử mặc định 20, tối đa 100.
  3. Đọc vai trò và danh tính người dùng từ phiên đăng nhập.
  4. Nếu là SUPER_ADMIN hoặc ADMIN thì trả về toàn bộ danh sách quảng cáo kèm media, không phân trang.
  5. Nếu là PARTNER thì đếm tổng số quảng cáo của Partner rồi lấy trang hiện tại gồm thông tin quảng cáo, ảnh đầu tiên và số liệu lượt hiển thị, click, tỷ lệ click, sắp xếp theo thời gian tạo giảm dần.
  6. Trả về danh sách kèm tổng số, trang và số phần tử mỗi trang cho Partner; lỗi database thì báo lỗi máy chủ.
- **Entity**: Admin/SuperAdmin (xem tất cả), Partner (xem quảng cáo của mình)
- **Model**: models.Ad, models.AdMedia
- **Databases**: ads, ad_media, ad_analytics, users, profiles (nhánh admin qua `GetAll`; nhánh partner subquery trên ad_media/ad_analytics)

### 45. Phân tích hiệu quả 1 quảng cáo (analytics)

- **Name**: GetAnalytics
- **Description**:
  1. Người dùng gọi API với id quảng cáo cần xem hiệu quả.
  2. Kiểm tra quyền sở hữu: SUPER_ADMIN và ADMIN được bỏ qua kiểm tra, PARTNER chỉ xem được quảng cáo của mình; không thuộc về thì báo lỗi, id sai thì báo không tìm thấy.
  3. Dịch vụ kiểm tra lại quyền sở hữu một lần nữa, sai thì báo lỗi không phải chủ sở hữu.
  4. Tải quảng cáo kèm media và lấy bộ đếm sự kiện theo từng loại (lượt hiển thị, click, xem, vuốt, bắt đầu và kết thúc video) cùng số người dùng duy nhất.
  5. Tính tổng tương tác, chặn số chi tiêu không vượt quá kinh phí và số dư kinh phí không được âm.
  6. Tính các chỉ số: tỷ lệ click trên lượt hiển thị, chi phí trên nghìn hiển thị và chi phí mỗi click (chỉ khi có dữ liệu mẫu).
  7. Trả về báo cáo hiệu quả gồm ngân sách, chi tiêu, độ tiếp cận, số liệu video và lịch chạy.
- **Entity**: Partner (chủ ad), Admin/SuperAdmin
- **Model**: models.Ad, models.AdAnalytics
- **Databases**: ads, ad_analytics, ad_media

### 46. Đổi trạng thái quảng cáo (partner/admin)

- **Name**: UpdateStatus
- **Description**:
  1. Kiểm tra quyền sở hữu quảng cáo ngay từ middleware, Partner không thể đổi trạng thái quảng cáo của người khác.
  2. Người dùng gọi API với trạng thái mới (đang hoạt động, tạm dừng, hoàn thành, chờ duyệt hoặc từ chối); thiếu hoặc sai thì báo dữ liệu không hợp lệ.
  3. Dịch vụ xác minh lại quyền sở hữu, sai thì báo lỗi không phải chủ sở hữu.
  4. Tải quảng cáo, gán trạng thái mới (giá trị không hợp lệ thì rơi về tạm dừng) và lưu vào database.
  5. Trả về thông báo đổi trạng thái thành công kèm dữ liệu quảng cáo; không sinh thông báo cho Partner (khác với nhánh admin duyệt quảng cáo).
- **Entity**: Partner (chủ ad), Admin/SuperAdmin
- **Model**: models.Ad
- **Databases**: ads

### 47. Cập nhật nội dung quảng cáo

- **Name**: UpdateAd
- **Description**:
  1. Kiểm tra quyền sở hữu ngay từ middleware, Partner chỉ được sửa quảng cáo của mình.
  2. Người dùng gọi API với form đa phần gồm các trường văn bản cần sửa (ngày, số, dữ liệu nhắm mục tiêu) và file media nếu có.
  3. Validate lại toàn bộ dữ liệu theo định dạng như khi tạo; lỗi thì báo dữ liệu không hợp lệ.
  4. Dịch vụ xác minh lại quyền sở hữu, sai thì báo lỗi không phải chủ sở hữu.
  5. Chỉ cho phép sửa khi quảng cáo đang hoạt động hoặc tạm dừng; nếu đang chờ duyệt, bị từ chối hoặc đã hoàn thành thì báo lỗi không được phép sửa.
  6. Chỉ cập nhật các trường có giá trị được gửi lên, kiểm tra tuổi tối đa phải lớn hơn 0 và nhỏ hơn 100, dữ liệu địa điểm được lưu lại rồi ghi vào database.
  7. Tải lại quảng cáo kèm media và trả về thông báo cập nhật thành công kèm dữ liệu quảng cáo.
- **Entity**: Partner (chủ ad), Admin/SuperAdmin
- **Model**: models.Ad, models.AdMedia
- **Databases**: ads, ad_media

### 48. Xoá quảng cáo (ads-management)

- **Name**: DeleteAd (ads-management)
- **Description**:
  1. Kiểm tra quyền sở hữu ngay từ middleware, Partner không thể xoá quảng cáo của người khác.
  2. Dịch vụ xác minh lại quyền sở hữu, sai thì báo lỗi không phải chủ sở hữu.
  3. Chỉ cho phép xoá khi quảng cáo đang chờ duyệt, bị từ chối hoặc tạm dừng; đang hoạt động hoặc đã hoàn thành thì báo lỗi không được phép xoá.
  4. Xoá thông tin media của quảng cáo trước, rồi xoá vĩnh viễn quảng cáo; không xoá được thì báo lỗi.
  5. Giải phóng slot trong gói đăng ký của Partner theo gói đang dùng; lỗi giải phóng bị bỏ qua.
  6. Trả về thông báo xoá quảng cáo thành công.
- **Entity**: Partner (chủ ad), Admin/SuperAdmin
- **Model**: models.Ad, models.AdMedia, models.PartnerSubscription
- **Databases**: ads, ad_media, partner_subscriptions

### 49. Tổng quan quảng cáo (overview)

- **Name**: GetOverview
- **Description**:
  1. Partner gọi API để xem tổng quan quảng cáo của mình.
  2. Hệ thống tổng hợp kinh phí và chi tiêu của các quảng cáo, tính số dư ngân sách (không âm) và số quảng cáo đang hoạt động.
  3. Tổng hợp số lượt hiển thị và click của Partner rồi tính tỷ lệ click khi có dữ liệu hiển thị.
  4. Lấy thông tin gói đăng ký đang hoạt động mới nhất gồm slot đã dùng, slot tối đa, tên gói và thời hạn.
  5. Nếu có lỗi thì báo lỗi máy chủ.
  6. Trả về dữ liệu tổng quan gồm ngân sách, chỉ số hiệu quả và thông tin gói đăng ký.
  7. Lưu ý: đường dẫn tổng quan được khai báo sau đường dẫn chi tiết nhưng hệ thống vẫn khớp đúng đường dẫn tĩnh.
- **Entity**: Partner/Admin/SuperAdmin (người xem)
- **Model**: models.Ad, models.AdAnalytics, models.PartnerSubscription, models.AdPackage
- **Databases**: ads, ad_analytics, partner_subscriptions, ad_packages

### 50. Liệt kê toàn bộ quy tắc vi phạm (admin)

- **Name**: ListAllRules
- **Description**:
  1. Admin gọi API với các bộ lọc tùy chọn: loại đối tượng áp dụng (bài viết, bình luận, người dùng), từ khóa tìm kiếm và lựa chọn hiển thị quy tắc đã tắt.
  2. Kiểm tra xác thực người dùng và quyền SUPER_ADMIN hoặc ADMIN.
  3. Mặc định admin xem cả quy tắc đã tắt; nếu truyền giá trị khác thì chỉ hiển thị quy tắc đang bật.
  4. Lọc danh sách theo trạng thái, loại đối tượng và tìm theo tiêu đề hoặc mô tả, sắp xếp theo thứ tự ưu tiên rồi thời gian tạo.
  5. Trả về danh sách quy tắc kèm tổng số.
- **Entity**: Admin/SuperAdmin
- **Model**: models.ViolationRule
- **Databases**: violation_rules

### 51. Tạo quy tắc vi phạm mới

- **Name**: CreateRule
- **Description**:
  1. Admin gọi API với thông tin quy tắc mới: tiêu đề (bắt buộc), mô tả, phạm vi áp dụng, mức độ nghiêm trọng và thứ tự ưu tiên (tuỳ chọn).
  2. Kiểm tra người dùng đã đăng nhập và có quyền SUPER_ADMIN.
  3. Validate tiêu đề và mô tả; lỗi thì báo dữ liệu không hợp lệ.
  4. Kiểm tra tiêu đề trùng với quy tắc đã có (không phân biệt hoa thường) thì báo lỗi trùng lặp.
  5. Tạo quy tắc mới với định danh duy nhất, chuẩn hóa chuỗi, phạm vi áp dụng mặc định là tất cả, mức độ nghiêm trọng mặc định là trung bình, bật quy tắc và ghi thời gian tạo.
  6. Lưu quy tắc vào database; quy tắc xuất hiện ngay trong danh sách chọn khi người dùng báo cáo vi phạm.
  7. Trả về thông báo tạo quy tắc thành công kèm dữ liệu quy tắc.
- **Entity**: SuperAdmin (người tạo), User thường (dùng rule khi report)
- **Model**: models.ViolationRule
- **Databases**: violation_rules

### 52. Cập nhật quy tắc vi phạm

- **Name**: UpdateRule
- **Description**:
  1. Admin gọi API với id quy tắc và các trường cần cập nhật, tất cả đều tùy chọn.
  2. Kiểm tra người dùng đã đăng nhập và có quyền SUPER_ADMIN.
  3. Tải quy tắc theo id, không tồn tại thì báo lỗi không tìm thấy.
  4. Chỉ cập nhật các trường có giá trị được gửi lên: tiêu đề (validate và kiểm tra trùng với quy tắc khác, loại trừ chính nó), mô tả (validate), phạm vi áp dụng, mức độ nghiêm trọng (giá trị lạ dùng mặc định) và thứ tự ưu tiên (âm thì đưa về 0).
  5. Lưu thay đổi vào database kèm thời gian cập nhật.
  6. Trả về thông báo cập nhật quy tắc thành công kèm dữ liệu quy tắc.
- **Entity**: SuperAdmin, User thường (xem rule ở dropdown report)
- **Model**: models.ViolationRule
- **Databases**: violation_rules

### 53. Bật / tắt quy tắc vi phạm

- **Name**: SetRuleActive
- **Description**:
  1. Admin gọi API với id quy tắc và trạng thái bật hoặc tắt mới; thiếu dữ liệu thì báo lỗi.
  2. Kiểm tra người dùng đã đăng nhập và có quyền SUPER_ADMIN.
  3. Kiểm tra quy tắc tồn tại, không có thì báo lỗi không tìm thấy.
  4. Cập nhật trạng thái hoạt động và thời gian cập nhật của quy tắc theo id; không cập nhật được thì báo lỗi không tìm thấy.
  5. Đây là ẩn mềm: không xoá dữ liệu nào, quy tắc tắt biến mất khỏi danh sách chọn báo cáo của người dùng nhưng các báo cáo đã trỏ tới quy tắc đó vẫn giữ nguyên.
  6. Trả về thông báo cập nhật trạng thái quy tắc thành công.
- **Entity**: SuperAdmin, User thường (dropdown report)
- **Model**: models.ViolationRule
- **Databases**: violation_rules, reports (không thay đổi — chỉ tham chiếu giữ nguyên)
