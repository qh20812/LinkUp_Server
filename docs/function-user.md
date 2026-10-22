# Tài liệu hàm API dành cho User

> **Nguồn**: routes/*.go — 194 endpoint user/public + /health + /metrics = 196 mục.
> Mỗi mục gồm: **Name** (tên func tiếng Anh), **Description** (luồng tương tác tiếng Việt), **Entity**, **Model**, **Databases**.
> Endpoint admin (RBAC quản trị) nằm trong docs/function-admin.md.

## Mục lục

| # | Endpoint | Name |
|---|---|---|
| 1 | POST /api/auth/register | Đăng ký tài khoản |
| 2 | POST /api/auth/login | Đăng nhập |
| 3 | POST /api/auth/google | Đăng nhập Google |
| 4 | POST /api/auth/refresh | Làm mới token (refresh) |
| 5 | POST /api/auth/verify-email | Xác thực email |
| 6 | POST /api/auth/resend-verification | Gửi lại email xác thực |
| 7 | GET /api/auth/password-policy | Xem chính sách mật khẩu |
| 8 | POST /api/auth/change-password | Đổi mật khẩu |
| 9 | POST /api/auth/logout | Đăng xuất |
| 10 | POST /api/auth/forgot-password | Quên mật khẩu (gửi email đặt lại) |
| 11 | POST /api/auth/verify-reset-token | Xác minh token đặt lại mật khẩu |
| 12 | POST /api/auth/reset-password | Đặt lại mật khẩu |
| 13 | GET /api/profile | Xem hồ sơ của chính mình |
| 14 | PATCH /api/profile | Chỉnh sửa hồ sơ |
| 15 | GET /api/profile/:userID | Xem hồ sơ người khác theo userID |
| 16 | GET /api/settings/privacy | Xem cài đặt quyền riêng tư |
| 17 | PUT /api/settings/privacy | Cập nhật cài đặt quyền riêng tư |
| 18 | GET /api/settings/storage | Xem thông tin dung lượng lưu trữ |
| 19 | POST /api/settings/deactivate | Vô hiệu hóa tài khoản tạm thời (Deactivate) |
| 20 | GET /api/settings/appearance | Xem cài đặt giao diện (theme/language) |
| 21 | PUT /api/settings/appearance | Cập nhật cài đặt giao diện |
| 22 | GET /api/settings/sessions | Liệt kê các phiên đăng nhập đang hoạt động |
| 23 | DELETE /api/settings/sessions/:id | Thu hồi một phiên đăng nhập cụ thể |
| 24 | POST /api/settings/sessions/revoke-others | Đăng xuất các thiết bị khác |
| 25 | GET /api/presence/online/:userID | Xem trạng thái online/offline của một người dùng |
| 26 | POST /api/presence/batch | Lấy presence theo batch (nhiều user) |
| 27 | GET /api/presence/online | Liệt kê danh sách user đang online |
| 28 | GET /api/presence/count | Đếm số user đang online |
| 29 | GET /api/presence/settings | Xem cài đặt presence (activity status + last seen visibility) |
| 30 | PUT /api/presence/settings | Cập nhật cài đặt presence |
| 31 | GET /api/locations/provinces | Liệt kê danh sách tỉnh/thành phố |
| 32 | GET /api/locations/wards?province_id= | Liệt kê phường/xã theo tỉnh |
| 33 | POST /api/locations/reverse-geocode | Reverse geocode tọa độ → tỉnh/phường |
| 34 | GET /api/posts | Lấy danh sách bài viết (Feed) |
| 35 | GET /api/posts/saved | Lấy bài viết đã lưu (Bookmark) |
| 36 | GET /api/posts/user/:userID | Lấy bài viết của một người dùng |
| 37 | GET /api/posts/:id | Xem chi tiết bài viết |
| 38 | POST /api/posts/:id/view | Ghi nhận lượt xem bài viết (impression feed) |
| 39 | POST /api/posts | Tạo bài viết mới |
| 40 | DELETE /api/posts/:id | Xóa bài viết |
| 41 | POST /api/posts/:id/react | Thả cảm xúc bài viết (toggle) |
| 42 | POST /api/posts/:id/comments | Tạo bình luận |
| 43 | GET /api/posts/:id/comments | Lấy danh sách bình luận |
| 44 | POST /api/posts/comments/:commentID/react | Thả cảm xúc bình luận (toggle) |
| 45 | POST /api/posts/:id/share | Chia sẻ bài viết (repost) |
| 46 | POST /api/posts/:id/save | Lưu/bỏ lưu bài viết (Bookmark toggle) |
| 47 | GET /api/posts/hashtag/:name | Tìm bài viết theo Hashtag (posts) |
| 48 | GET /api/emojis | Lấy danh sách Emoji |
| 49 | POST /api/posts/:id/pin | Ghim bài viết |
| 50 | DELETE /api/posts/:id/pin | Bỏ ghim bài viết |
| 51 | PATCH /api/posts/:id/comments | Bật/tắt bình luận của bài viết |
| 52 | GET /api/posts/user/:userID/media | Lấy media của một người dùng |
| 53 | GET /api/tags/:name/posts | Tag Controller — Bài viết theo Hashtag |
| 54 | POST /api/reports | Tạo báo cáo vi phạm |
| 55 | PUT /api/reports/:id | Chỉnh sửa báo cáo |
| 56 | POST /api/blocks | Chặn/bỏ chặn người dùng (toggle) |
| 57 | GET /api/blocks | Danh sách người dùng đã chặn |
| 58 | POST /api/stories | Đăng tải Story |
| 59 | GET /api/stories/feed | Feed Story trang chủ |
| 60 | GET /api/stories/user/:userID/active | Kiểm tra user có Story active |
| 61 | GET /api/stories/user/:userID | Lấy danh sách Story của 1 user |
| 62 | GET /api/stories/:id | Mở xem Story (ghi nhận view) |
| 63 | POST /api/stories/:id/interact | Tương tác với Story (react / reply / share) |
| 64 | GET /api/stories/:id/analytics | Thống kê Analytics Story |
| 65 | DELETE /api/stories/:id | Xóa Story |
| 66 | PUT /api/stories/mutes/:userID | Bật/tắt ẩn Story của 1 user (mute) |
| 67 | POST /api/follow/:userID | Follow / Unfollow (toggle) |
| 68 | GET /api/follow/stats/:userID | Thống kê Follow |
| 69 | GET /api/follow/suggestions | Gợi ý Follow |
| 70 | GET /api/follow/:userID/followers | Danh sách Followers |
| 71 | GET /api/follow/:userID/following | Danh sách Following |
| 72 | GET /api/follow/:userID/mutual | Danh sách Follow chung (mutual) |
| 73 | GET /api/friend-requests | Danh sách Lời mời kết bạn |
| 74 | GET /api/friends | Danh sách Bạn bè |
| 75 | POST /api/friend-requests/:userID | Gửi / Thu hồi Lời mời kết bạn (toggle) |
| 76 | PUT /api/friend-requests/:id/accept | Chấp nhận Lời mời kết bạn |
| 77 | DELETE /api/friend-requests/:id | Từ chối Lời mời kết bạn |
| 78 | GET /api/friends/search (route group thật là `/api/friends` — `routes/friend.routes.go:26`, KHÔNG phải `/api/friend-requests/search`) | Tìm kiếm trong danh sách Bạn bè |
| 79 | GET /api/friends/suggestions (route group thật là `/api/friends` — `routes/friend.routes.go:27`, KHÔNG phải `/api/friend-requests/suggestions`) | Gợi ý Kết bạn |
| 80 | DELETE /api/friends/:userID | Hủy kết bạn |
| 81 | GET /api/friend-status/:userID | Kiểm tra Quan hệ bạn bè với 1 user |
| 82 | POST /api/media/upload | Tải lên media (upload Cloudinary) |
| 83 | DELETE /api/media/:id | Xóa media (destroy Cloudinary) |
| 84 | GET /api/media/storage | Kiểm tra trạng thái storage |
| 85 | GET /api/media/user-media | Lấy danh sách media của user |
| 86 | GET /api/notifications | Lấy danh sách thông báo (phân trang) |
| 87 | GET /api/notifications/summary | Lấy summary thông báo (unread-count + 5 mới nhất) |
| 88 | PUT /api/notifications/:id/read | Đánh dấu 1 thông báo đã đọc |
| 89 | PUT /api/notifications/read-all | Đánh dấu tất cả thông báo đã đọc |
| 90 | GET /api/notifications/unread-count | Đếm thông báo chưa đọc |
| 91 | GET /api/notifications/preferences | Lấy preferences thông báo |
| 92 | PUT /api/notifications/preferences | Cập nhật preferences thông báo |
| 93 | POST /api/notifications/push-token | Đăng ký push token (Expo) |
| 94 | GET /api/search | Tìm kiếm (users/posts/hashtags/communities) |
| 95 | GET /api/trending | Lấy hashtag trending |
| 96 | GET /api/ads/packages | Lấy danh sách gói quảng cáo |
| 97 | POST /api/ads-management/subscribe | Đăng ký gói quảng cáo (subscribe) |
| 98 | GET /api/ads-management/subscription | Lấy subscription hiện tại của user |
| 99 | GET /api/customer/feed | Lấy feed quảng cáo cho customer (targeting + frequency cap) |
| 100 | POST /api/customer/ads/:id/track | Track click/impression quảng cáo (ad_analytics + trừ ngân sách) |
| 101 | GET /api/chats | Danh sách chat trực tiếp |
| 102 | GET /api/chats/ws | WebSocket chat (real-time, mã hóa đầu cuối) |
| 103 | GET /api/chats/invites | Danh sách lời mời chat đã nhận |
| 104 | POST /api/chats/direct | Tạo (hoặc lấy lại) chat trực tiếp |
| 105 | POST /api/chats/invite | Gửi lời mời chat (khi chưa kết bạn) |
| 106 | POST /api/chats/invite/respond | Phản hồi lời mời chat (chấp nhận / từ chối) |
| 107 | POST /api/chats/media (multipart/form-data: file, chat_id, duration_seconds) | Upload media vào chat (Cloudinary) |
| 108 | GET /api/chats/messages/:messageID/download | Tải về media của tin nhắn |
| 109 | DELETE /api/chats/:chatID | Xóa chat |
| 110 | POST /api/chats/share | Chia sẻ bài viết vào chat trực tiếp |
| 111 | GET /api/chats/:chatID/background | Lấy background của chat |
| 112 | PUT /api/chats/:chatID/background | Đặt / cập nhật background của chat |
| 113 | DELETE /api/chats/:chatID/background | Xóa background của chat |
| 114 | GET /api/chats/:chatID/shared-content?tab=all|media|files|links|posts&offset=0&limit=30 | Nội dung đã chia sẻ trong chat (Shared content sidebar) |
| 115 | PUT /api/e2e/keys | Đăng ký public key E2E của user |
| 116 | GET /api/e2e/keys/:userID | Lấy public key E2E của user khác |
| 117 | POST /api/e2e/chats/keys | Lưu khóa chat E2E (wrapped key, batch) |
| 118 | GET /api/e2e/chats/:chatID/keys | Lấy khóa chat E2E của chính mình |
| 119 | PATCH /api/e2e/chats/:chatID/keys | Re-key khóa chat (sau khi đối phương đổi identity) |
| 120 | PUT /api/e2e/recovery | Lưu backup khôi phục khóa chat |
| 121 | GET /api/e2e/recovery | Trạng thái backup khôi phục (metadata) |
| 122 | POST /api/e2e/recovery/unlock | Mở khóa backup khôi phục (PIN / recovery key) |
| 123 | DELETE /api/e2e/recovery | Xóa backup khôi phục |
| 124 | GET /api/group-chats | Danh sách nhóm chat của tôi |
| 125 | POST /api/group-chats | Tạo nhóm chat mới |
| 126 | POST /api/group-chats/:chatID/add-member | Thêm thành viên (tạo lời mời) |
| 127 | POST /api/group-chats/:chatID/member-requests/:requestID/approve | Chấp nhận lời mời vào nhóm |
| 128 | POST /api/group-chats/:chatID/member-requests/:requestID/reject | Từ chối lời mời vào nhóm |
| 129 | POST /api/group-chats/:chatID/ban | Trục xuất & chặn thành viên (ban) |
| 130 | POST /api/group-chats/:chatID/messages | Gửi tin nhắn vào nhóm (HTTP) |
| 131 | POST /api/group-chats/:chatID/transfer-admin | Chuyển quyền admin |
| 132 | POST /api/group-chats/:chatID/transfer-ownership | Chuyển quyền sở hữu nhóm (owner) |
| 133 | GET /api/group-chats/:chatID/settings | Xem cấu hình nhóm |
| 134 | PUT /api/group-chats/:chatID/settings | Cập nhật cấu hình nhóm |
| 135 | POST /api/group-chats/:chatID/mute | Tắt tiếng thành viên (mute, có thời hạn) |
| 136 | POST /api/group-chats/:chatID/unmute | Bỏ tắt tiếng thành viên (unmute) |
| 137 | POST /api/group-chats/:chatID/leave | Rời khỏi nhóm |
| 138 | GET /api/group-chats/:chatID/background | Xem nền (background) chat nhóm |
| 139 | PUT /api/group-chats/:chatID/background | Đặt/cập nhật nền chat nhóm |
| 140 | DELETE /api/group-chats/:chatID/background | Xóa nền chat nhóm |
| 141 | GET /api/group-chats/:chatID/shared-content?tab=all|media|files|links|posts&offset=0&limit=30 | Nội dung đã chia sẻ trong nhóm (sidebar) |
| 142 | GET /api/group-chats/ws?token={access JWT} | Group Chat WebSocket (real-time nhóm) |
| 143 | GET /api/group-calls/ws?token={access JWT} | Group Call WebSocket (WebRTC signaling + lịch sử MongoDB) |
| 144 | GET /api/calls/ws | Kết nối WebSocket tín hiệu cuộc gọi |
| 145 | GET /api/calls/ice-servers | Lấy danh sách ICE/TURN servers |
| 146 | GET /api/calls/history | Lấy lịch sử cuộc gọi (filter + phân trang) |
| 147 | GET /api/calls/missed/count | Đếm cuộc gọi nhỡ chưa đọc |
| 148 | POST /api/calls/missed/read | Đánh dấu cuộc gọi nhỡ đã đọc |
| 149 | DELETE /api/calls/:callID/hide | Ẩn cuộc gọi khỏi lịch sử (soft-delete) |
| 150 | GET /api/calls/:callID | Xem chi tiết một cuộc gọi |
| 151 | POST /api/calls/initiate | Khởi tạo cuộc gọi (chống bận) |
| 152 | POST /api/calls/:callID/accept | Chấp nhận cuộc gọi (atomic) |
| 153 | POST /api/calls/:callID/reject | Từ chối cuộc gọi (atomic) |
| 154 | POST /api/calls/:callID/mute | Bật/tắt tiếng (mute) trong cuộc gọi |
| 155 | POST /api/calls/:callID/video | Bật/tắt camera (video) trong cuộc gọi |
| 156 | GET /api/communities | Danh sách cộng đồng (khám phá) |
| 157 | GET /api/communities/:communityID | Chi tiết cộng đồng |
| 158 | GET /api/communities/:communityID/posts | Bài viết trong cộng đồng |
| 159 | GET /api/communities/joined | Cộng đồng đã tham gia |
| 160 | GET /api/communities/created | Cộng đồng đã tạo |
| 161 | POST /api/communities | Tạo cộng đồng |
| 162 | PUT /api/communities/:communityID | Cập nhật thông tin cộng đồng |
| 163 | PUT /api/communities/:communityID/background | Cập nhật ảnh bìa (background) cộng đồng |
| 164 | POST /api/communities/:communityID/join | Yêu cầu tham gia cộng đồng |
| 165 | GET /api/communities/:communityID/join-requests | Danh sách yêu cầu tham gia đang chờ |
| 166 | PUT /api/communities/:communityID/join-requests/:requestID/approve | Duyệt yêu cầu tham gia |
| 167 | PUT /api/communities/:communityID/join-requests/:requestID/reject | Từ chối yêu cầu tham gia |
| 168 | GET /api/communities/:communityID/members | Danh sách thành viên cộng đồng |
| 169 | PUT /api/communities/:communityID/members/:memberID/role | Đổi vai trò thành viên |
| 170 | DELETE /api/communities/:communityID/members/:memberID | Đuổi thành viên khỏi cộng đồng |
| 171 | DELETE /api/communities/:communityID/leave | Rời cộng đồng |
| 172 | POST /api/communities/:communityID/transfer-ownership | Chuyển quyền sở hữu cộng đồng |
| 173 | POST /api/communities/:communityID/invite-codes | Tạo mã mời |
| 174 | GET /api/communities/:communityID/invite-codes | Danh sách mã mời |
| 175 | DELETE /api/communities/:communityID/invite-codes/:codeID | Vô hiệu hóa mã mời |
| 176 | POST /api/communities/:communityID/invitations | Gửi lời mời tham gia cộng đồng |
| 177 | GET /api/communities/invitations | Danh sách lời mời của tôi |
| 178 | PUT /api/communities/invitations/:invitationID/respond | Phản hồi lời mời tham gia |
| 179 | GET /api/communities/:communityID/rules | Xem nội quy cộng đồng |
| 180 | POST /api/communities/:communityID/rules | Tạo nội quy cộng đồng |
| 181 | PUT /api/communities/:communityID/rules/:ruleID | Cập nhật nội quy cộng đồng |
| 182 | DELETE /api/communities/:communityID/rules/:ruleID | Xoá nội quy cộng đồng |
| 183 | GET /api/communities/:communityID/policy | Xem chính sách đóng góp của cộng đồng |
| 184 | PUT /api/communities/:communityID/policy | Cập nhật chính sách đóng góp |
| 185 | GET /api/communities/:communityID/contributions/leaderboard | Bảng xếp hạng đóng góp (leaderboard) |
| 186 | GET /api/communities/:communityID/contributions/me | Xem điểm đóng góp của chính mình |
| 187 | GET /api/communities/:communityID/contributions/:userID | Xem điểm đóng góp của user khác (public) |
| 188 | GET /api/communities/:communityID/contributions/members | Danh sách thành viên cộng đồng kèm điểm đóng góp |
| 189 | GET /api/communities/:communityID/challenges | Danh sách challenge đang diễn ra |
| 190 | POST /api/communities/:communityID/challenges | Tạo challenge mới |
| 191 | POST /api/communities/:communityID/challenges/:challengeID/join | Tham gia challenge |
| 192 | GET /api/communities/:communityID/challenges/:challengeID/participants | Danh sách người tham gia challenge |
| 193 | GET /api/violation-rules | Danh sách quy tắc vi phạm (đổ dropdown khi report) |
| 194 | GET /api/violation-rules/:id | Chi tiết một quy tắc vi phạm |
| 195 | GET /health | Kiểm tra sức khỏe server |
| 196 | GET /metrics | Prometheus metrics endpoint |

---
## B1 — Tài liệu hóa Endpoint Auth + Password Reset

Tài liệu tự động trích xuất từ source. Prefix chung: `/api/auth` (routes/auth.routes.go, routes/password_reset.routes.go). Auth middleware cho nhóm protected: `middlewares.AuthMiddleware(env, db)`.

Validation: các endpoint Auth dùng `validations.AuthValidation` (không dùng `binding` tag) — controller chỉ `ShouldBindJSON` để parse JSON rồi gọi `h.validation.ValidateXxx`.

---

### 1. Đăng ký tài khoản

- **Name**: Register
- **Description**:
  1. User nhập email và password, username
  2. Hiển thị chính sách của nền tảng và yêu cầu người dùng tích vào
  3. Validate dữ liệu (email format, password length + format, tích chính sách)
  4. Kiểm tra email đã tồn tại chưa
  5. Kiểm tra cài đặt cho phép đăng ký và không đang bảo trì, sinh username tự động đảm bảo không trùng
  6. Hash password để bảo mật
  7. Tạo user object
  8. Lưu user vào database kèm lịch sử mật khẩu, tạo hồ sơ, gán vai trò mặc định và sinh ảnh đại diện
  9. Nếu bật yêu cầu xác thực email thì gửi email xác thực kèm token hết hạn
  10. Ngược lại trả về thông báo đăng ký thành công tính là đăng nhập luôn
- **Entity**: User (tạo mới, kèm Profile + Role + PasswordHistory)
- **Model**: models.User, models.Profile, models.PasswordHistory, models.UserRole, models.Role, models.EmailVerificationToken
- **Databases**: users, profiles, password_histories, user_roles, roles, system_configs (đọc), email_verification_tokens (nếu require_email_verify=true)

### 2. Đăng nhập

- **Name**: Login
- **Description**:
  1. Người dùng nhập email và mật khẩu rồi gọi API đăng nhập
  2. Validate dữ liệu (định dạng email, mật khẩu) và thu thập thông tin thiết bị để ghi phiên đăng nhập
  3. Chuẩn hóa email rồi tìm tài khoản theo email; không thấy thì báo sai thông tin đăng nhập, không tiết lộ email có tồn tại
  4. Nếu tài khoản đang bị chặn thì tra lệnh chặn mới nhất: hết hạn thì tự kích hoạt lại, còn hạn thì báo lỗi chặn
  5. Kiểm tra trạng thái tài khoản: từng tự ngừng hoạt động thì cho đăng nhập lại, bị quản trị viên đình chỉ thì chặn, đang bị khóa do nhập sai quá nhiều lần thì chặn (quản trị viên được miễn)
  6. So sánh mật khẩu với giá trị đã băm; sai thì tăng số lần nhập sai, còn quyền báo số lần còn lại, hết thì khóa tài khoản
  7. Đăng nhập thành công: đặt lại số lần nhập sai và kích hoạt lại tài khoản từng tự ngừng hoạt động
  8. Kiểm tra chế độ bảo trì (chỉ quản trị viên được vào) và yêu cầu xác thực email (chưa xác thực thì chặn)
  9. Tạo phiên đăng nhập mới lưu thiết bị và địa chỉ IP, phiên được gắn vào token
  10. Tạo cặp token truy cập và làm mới rồi trả về thông tin người dùng, token và dung lượng lưu trữ
- **Entity**: User (tồn tại)
- **Model**: models.User, models.UserSession
- **Databases**: users, user_roles, roles, user_sessions, system_configs (đọc), bans (đọc qua ensureBanStatus)

### 3. Đăng nhập Google

- **Name**: GoogleLogin
- **Description**:
  1. Người dùng chọn đăng nhập Google và gửi mã định danh hoặc mã ủy quyền từ Google
  2. Nếu gửi mã ủy quyền thì đổi lấy mã định danh từ Google; chưa cấu hình tài khoản Google thì báo lỗi chưa cấu hình
  3. Xác minh mã định danh Google: chữ ký hợp lệ, đúng nguồn phát hành, còn thời hạn, email đã xác minh, thuộc danh sách ứng dụng đã cấu hình và email hoặc định danh không rỗng
  4. Chuẩn hóa email rồi tìm tài khoản theo email hoặc theo định danh Google; chưa có thì tạo tài khoản mới
  5. Khi tạo mới: kiểm tra cài đặt cho phép đăng ký, sinh username, đánh dấu email đã xác thực ngay, mật khẩu để trống, gắn định danh Google, tải ảnh đại diện Google, tạo hồ sơ và gán vai trò mặc định
  6. Nếu tài khoản đã tồn tại nhưng chưa liên kết Google thì liên kết; định danh Google đã gắn với tài khoản khác thì báo lỗi không khớp
  7. Kiểm tra chặn và tài khoản hoạt động; nếu tài khoản đăng ký bằng mật khẩu chưa xác thực email thì đánh dấu đã xác thực vì Google đã xác thực
  8. Kiểm tra chế độ bảo trì, chỉ quản trị viên được đăng nhập
  9. Tạo cặp token truy cập và làm mới; đăng nhập Google không tạo phiên đăng nhập nên token không gắn phiên
  10. Trả về thông tin người dùng, token và dung lượng lưu trữ
- **Entity**: User (tìm hoặc tạo mới từ Google)
- **Model**: models.User, models.Profile
- **Databases**: users, profiles, user_roles, roles, system_configs (đọc)

### 4. Làm mới token (refresh)

- **Name**: RefreshToken
- **Description**:
  1. Người dùng gửi token làm mới để lấy token truy cập mới
  2. Thu thập thông tin thiết bị cho phiên đăng nhập mới
  3. Xác thực chữ ký của token; token sai hoặc không hợp lệ thì báo token làm mới không hợp lệ
  4. Kiểm tra loại token phải là token làm mới, gửi nhầm token truy cập thì báo lỗi sai loại token
  5. Tìm tài khoản theo định danh trong token rồi kiểm tra chặn và tài khoản còn hoạt động
  6. So sánh phiên bản token trong token với giá trị lưu trong cơ sở dữ liệu; token cấp trước lần đăng xuất hoặc đổi mật khẩu gần nhất bị từ chối
  7. Lấy vai trò hiện tại của tài khoản vì vai trò có thể đã đổi
  8. Thu hồi phiên cũ và tạo phiên đăng nhập mới, dọn các phiên hết hạn trước
  9. Cấp cặp token truy cập và làm mới mới gắn với phiên mới rồi trả về
- **Entity**: User (phiên bản hiện tại)
- **Model**: models.User, models.UserSession
- **Databases**: users, user_roles, roles, user_sessions

### 5. Xác thực email

- **Name**: VerifyEmail
- **Description**:
  1. Người dùng nhấn liên kết xác thực trong email và gửi mã token
  2. Tra token trong cơ sở dữ liệu theo giá trị đã băm; không thấy thì báo token không tồn tại
  3. Kiểm tra token đã hết hạn (hạn 1 giờ) hay đã được sử dụng, trả thông báo tương ứng cho từng trường hợp
  4. Nếu hợp lệ: trong một giao dịch, cập nhật thời gian xác thực email cho tài khoản và đánh dấu token đã dùng
  5. Tải lại thông tin tài khoản và vai trò hiện tại
  6. Tạo cặp token truy cập và làm mới rồi trả kèm, xác thực email xong là đăng nhập luôn
  7. Trả thông báo xác thực email thành công
- **Entity**: User (đánh dấu email verified)
- **Model**: models.EmailVerificationToken, models.User
- **Databases**: email_verification_tokens, users, user_roles, roles, system_configs (đọc)

### 6. Gửi lại email xác thực

- **Name**: ResendVerification
- **Description**:
  1. Người dùng nhập email rồi gọi API gửi lại email xác thực
  2. Validate định dạng email
  3. Tìm tài khoản theo email; không tồn tại thì trả thông báo chung chung rằng nếu email tồn tại bạn sẽ nhận được hướng dẫn xác thực, không tiết lộ email có tồn tại
  4. Nếu email đã được xác thực rồi thì trả thông báo đã hoàn tất xác thực
  5. Kiểm tra hạn gửi trong 60 giây kể từ lần gửi trước, chưa vượt thì báo còn số giây và không gửi lại
  6. Dọn token hết hạn và token cũ của người dùng, sinh token mới (token thô dùng trong liên kết, bản băm lưu trong cơ sở dữ liệu, hạn 1 giờ) rồi gửi email chứa liên kết web và liên kết mở ứng dụng; lỗi gửi chỉ ghi nhận, không làm hỏng yêu cầu
  7. Trả thông báo chung chung
- **Entity**: User
- **Model**: models.EmailVerificationToken, models.User
- **Databases**: email_verification_tokens, users

### 7. Xem chính sách mật khẩu

- **Name**: GetPasswordPolicy
- **Description**:
  1. Người dùng gọi API xem chính sách mật khẩu mà không cần nhập dữ liệu
  2. Hệ thống không validate gì thêm, không có thay đổi dữ liệu
  3. Đọc độ dài mật khẩu tối thiểu từ cấu hình hệ thống (mặc định 8, trong khoảng 8 đến 50), độ dài tối đa luôn là 50
  4. Trả về độ dài mật khẩu tối thiểu và tối đa để ứng dụng hiển thị ràng buộc khi đăng ký hoặc đổi mật khẩu
- **Entity**: Hệ thống (admin settings)
- **Model**: models.SystemConfig
- **Databases**: system_configs (đọc)

### 8. Đổi mật khẩu

- **Name**: ChangePassword
- **Description**:
  1. Người dùng đã đăng nhập nhập mật khẩu hiện tại và mật khẩu mới rồi gọi API đổi mật khẩu
  2. Xác thực token truy cập trước khi xử lý, không hợp lệ thì từ chối
  3. Validate mật khẩu mới về định dạng và độ dài theo cấu hình (tối thiểu theo cài đặt hệ thống, tối đa 50)
  4. Tìm tài khoản theo định danh người dùng, không thấy thì báo tài khoản không tồn tại
  5. So sánh mật khẩu hiện tại với giá trị đã băm, sai thì báo thông tin đăng nhập không đúng
  6. Chặn mật khẩu mới trùng mật khẩu hiện tại và trùng bất kỳ mật khẩu cũ nào trong lịch sử
  7. Băm mật khẩu mới, cập nhật cho tài khoản và ghi vào lịch sử mật khẩu
  8. Thu hồi toàn bộ phiên đăng nhập, tăng phiên bản token để vô hiệu mọi token cũ và xóa dữ liệu xác thực tạm trong bộ nhớ
  9. Trả thông báo đổi mật khẩu thành công
- **Entity**: User
- **Model**: models.User
- **Databases**: users, password_histories, user_sessions

### 9. Đăng xuất

- **Name**: Logout
- **Description**:
  1. Người dùng đã đăng nhập gọi API đăng xuất, không cần nhập dữ liệu
  2. Xác thực token truy cập để lấy người dùng, không hợp lệ thì từ chối
  3. Thu hồi toàn bộ phiên đăng nhập của người dùng, đăng xuất khỏi mọi thiết bị
  4. Tăng phiên bản token để mọi token cũ bị từ chối khi xác thực và làm mới, đồng thời xóa dữ liệu xác thực tạm trong bộ nhớ
  5. Trả thông báo đăng xuất thành công
- **Entity**: User
- **Model**: models.User, models.UserSession
- **Databases**: users, user_sessions

### 10. Quên mật khẩu (gửi email đặt lại)

- **Name**: ForgotPassword
- **Description**:
  1. Người dùng nhập email rồi gọi API quên mật khẩu
  2. Validate định dạng email
  3. Tìm tài khoản theo email; không tồn tại thì vẫn trả thông báo chung chung, không tiết lộ email có tồn tại
  4. Dọn token đặt lại: xóa token hết hạn toàn hệ thống và token cũ của người dùng
  5. Sinh token đặt lại mới (token thô dùng trong liên kết, bản băm lưu trong cơ sở dữ liệu, hạn 10 phút) và lưu lại
  6. Gửi email đặt lại mật khẩu chứa liên kết cho web và liên kết mở ứng dụng cho mobile; lỗi gửi chỉ ghi nhận, không làm hỏng yêu cầu
  7. Trả thông báo nếu email tồn tại bạn sẽ nhận được hướng dẫn đặt lại mật khẩu
- **Entity**: User
- **Model**: models.PasswordResetToken, models.User
- **Databases**: password_reset_tokens, users

### 11. Xác minh token đặt lại mật khẩu

- **Name**: VerifyResetToken
- **Description**:
  1. Người dùng mở trang đặt lại mật khẩu và gửi token lấy từ liên kết trong email
  2. Tra token trong cơ sở dữ liệu theo giá trị đã băm; không thấy thì trả kết quả không hợp lệ kèm thông báo token không hợp lệ
  3. Token hết hạn (hạn 10 phút) thì trả không hợp lệ kèm thông báo token đã hết hạn
  4. Token đã được sử dụng thì trả không hợp lệ kèm thông báo token đã được sử dụng
  5. Token hợp lệ thì trả kết quả hợp lệ; mọi trường hợp đều trả về thành công với kết quả tương ứng, ứng dụng dùng để cho phép hiển thị form đặt lại mật khẩu và không có thay đổi dữ liệu
- **Entity**: Hệ thống (token)
- **Model**: models.PasswordResetToken
- **Databases**: password_reset_tokens

### 12. Đặt lại mật khẩu

- **Name**: ResetPassword
- **Description**:
  1. Người dùng nhập mật khẩu mới và gửi token lấy từ liên kết đặt lại mật khẩu
  2. Validate định dạng mật khẩu mới
  3. Tra token trong cơ sở dữ liệu theo giá trị đã băm; không thấy thì báo token không tồn tại, hết hạn (10 phút) thì báo hết hạn, đã dùng thì báo đã sử dụng
  4. Validate lại mật khẩu mới theo quy tắc và độ dài theo cấu hình (tối thiểu theo cài đặt hệ thống, tối đa 50)
  5. Băm mật khẩu mới để bảo mật
  6. Cập nhật mật khẩu cho tài khoản sở hữu token và đánh dấu token đã dùng để không dùng lại được
  7. Trả thông báo mật khẩu đã được đặt lại thành công; token hoặc mật khẩu không hợp lệ thì báo lỗi tương ứng, chức năng này không thu hồi phiên đăng nhập hay phiên bản token
- **Entity**: User
- **Model**: models.PasswordResetToken, models.User
- **Databases**: password_reset_tokens, users

## B2 — Profile + Settings + Presence + Locations (Endpoint 13–33)

> Tài liệu hóa 21 endpoint: Profile (13–15), User Settings (16–24), Presence (25–30), Locations (31–33).
> Route files: `routes/profile.routes.go`, `routes/settings.routes.go`, `routes/presence.routes.go`, `routes/location.routes.go`.

---

### 13. Xem hồ sơ của chính mình

- **Name**: ViewProfile
- **Description**:
  1. Người dùng mở trang hồ sơ của chính mình với access token hợp lệ.
  2. Hệ thống xác thực token và lấy người dùng đang gọi; nếu không có → báo lỗi chưa đăng nhập.
  3. Truy vấn hồ sơ của người dùng đó kèm số liệu tổng hợp: số bài viết, số bạn bè và tên hiển thị từ tài khoản.
  4. Nếu hồ sơ chưa tồn tại → báo lỗi không tìm thấy hồ sơ.
  5. Trả về hồ sơ của chính người dùng kèm các số liệu thống kê.
  6. Chức năng này chỉ xem được hồ sơ của chính người gọi, không nhận tham số người dùng khác.
- **Entity**: User (người gọi request)
- **Model**: models.Profile (qua repository.ProfileView enriched view)
- **Databases**: profiles, users, posts, friends (join đếm post_count/friend_count)

---

### 14. Chỉnh sửa hồ sơ

- **Name**: EditProfile
- **Description**:
  1. Người dùng chỉnh sửa hồ sơ và gửi các trường muốn thay đổi (có thể chỉ gửi một phần; trường không gửi giữ nguyên).
  2. Hệ thống xác thực người gọi rồi kiểm tra dữ liệu gửi lên: nếu không có trường nào được gửi hoặc họ tên hiển thị bị để trống → báo lỗi dữ liệu không hợp lệ.
  3. Tải hồ sơ hiện tại của người dùng; nếu chưa có → báo lỗi không tìm thấy hồ sơ.
  4. Nếu có cập nhật số điện thoại, kiểm tra số đó đã được người dùng khác sử dụng chưa; trùng → báo lỗi số điện thoại đã tồn tại.
  5. Kiểm tra riêng từng trường: họ tên hiển thị đúng định dạng; ảnh đại diện và ảnh bìa phải là tệp media hợp lệ và không bị từ chối khi duyệt; giới tính chỉ nhận các giá trị cho phép.
  6. Kiểm tra nghề nghiệp và học vấn theo danh sách giá trị cho phép.
  7. Quy tắc liên kết giữa các trường: nếu nghề nghiệp không phải dạng khác thì tự động xóa ô nghề nghiệp khác, ngược lại ô nghề nghiệp khác là bắt buộc.
  8. Chỉ cập nhật các trường người dùng thực sự gửi lên, kể cả giá trị rỗng hay false; các trường còn lại giữ nguyên.
  9. Cập nhật thời gian sửa rồi lưu thay đổi vào database.
  10. Trả về hồ sơ đã cập nhật kèm số liệu thống kê và thông báo tiếng Việt "Cập nhật hồ sơ thành công".
- **Entity**: User (người gọi), Media (kiểm tra avatar/cover)
- **Model**: models.Profile, models.Media
- **Databases**: profiles, media

---

### 15. Xem hồ sơ người khác theo userID

- **Name**: ViewProfileByID
- **Description**:
  1. Người dùng mở hồ sơ của người khác (endpoint công khai, không bắt buộc đăng nhập; nếu có gửi token thì dùng người xem đó).
  2. Lấy mã người dùng cần xem từ đường dẫn; nếu thiếu → báo lỗi dữ liệu không hợp lệ.
  3. Truy vấn hồ sơ kèm số liệu của người đó; nếu không tồn tại → báo lỗi không tìm thấy hồ sơ.
  4. Nếu người xem chính là người đang được xem → trả hồ sơ đầy đủ bất kể cài đặt riêng tư.
  5. Nếu là người khác mà hồ sơ đang bật chế độ riêng tư → báo lỗi hồ sơ riêng tư.
  6. Nếu hồ sơ công khai → trả về hồ sơ kèm số liệu thống kê.
  7. Quy tắc riêng tư được kiểm tra ngay trong lúc xử lý yêu cầu, nên route vẫn công khai mà vẫn bảo vệ được hồ sơ riêng tư.
- **Entity**: Viewer (optional, anonymous nếu không đăng nhập), Target user
- **Model**: models.Profile (qua repository.ProfileView)
- **Databases**: profiles, users, posts, friends

---

### 16. Xem cài đặt quyền riêng tư

- **Name**: GetPrivacy
- **Description**:
  1. Người dùng mở trang cài đặt quyền riêng tư với access token hợp lệ.
  2. Hệ thống xác thực người gọi; nếu không có → báo lỗi chưa đăng nhập.
  3. Đọc cài đặt của người dùng; nếu chưa từng chỉnh sửa thì dùng giá trị mặc định (hiện trong kết quả tìm kiếm: bật; cho người lạ nhắn tin: tắt).
  4. Trả về hai cài đặt: có xuất hiện trong kết quả tìm kiếm không và có cho người lạ nhắn tin không.
- **Entity**: User (người gọi)
- **Model**: models.UserSetting
- **Databases**: user_settings

---

### 17. Cập nhật cài đặt quyền riêng tư

- **Name**: UpdatePrivacy
- **Description**:
  1. Người dùng bật/tắt các công tắc quyền riêng tư và gửi lên.
  2. Hệ thống xác thực người gọi; nếu không gửi cài đặt nào → báo lỗi dữ liệu không hợp lệ.
  3. Đọc cài đặt hiện tại; nếu chưa có thì khởi tạo với giá trị mặc định.
  4. Chỉ áp dụng các trường được gửi lên; trường không gửi giữ nguyên giá trị cũ.
  5. Lưu (thêm mới hoặc cập nhật) cài đặt của người dùng.
  6. Trả về cài đặt quyền riêng tư đã cập nhật.
- **Entity**: User (người gọi)
- **Model**: models.UserSetting
- **Databases**: user_settings

---

### 18. Xem thông tin dung lượng lưu trữ

- **Name**: GetStorage
- **Description**:
  1. Người dùng mở trang thông tin dung lượng lưu trữ với access token hợp lệ.
  2. Hệ thống xác thực người gọi; nếu không có → báo lỗi chưa đăng nhập.
  3. Lấy thông tin tài khoản của người gọi rồi đọc hạn mức dung lượng và dung lượng đã dùng.
  4. Tính số byte còn lại từ hạn mức trừ đi số đã dùng.
  5. Trả về hạn mức, dung lượng đã dùng và dung lượng còn lại.
- **Entity**: User (người gọi)
- **Model**: models.User (qua AuthRepository)
- **Databases**: users

---

### 19. Vô hiệu hóa tài khoản tạm thời (Deactivate)

- **Name**: Deactivate
- **Description**:
  1. Người dùng xác nhận vô hiệu hóa tài khoản và nhập mật khẩu hiện tại để xác nhận danh tính.
  2. Hệ thống xác thực người gọi; nếu thiếu mật khẩu → báo lỗi dữ liệu không hợp lệ.
  3. So sánh mật khẩu đã nhập với mật khẩu đã lưu; sai → báo lỗi mật khẩu không đúng.
  4. Nếu đúng → đánh dấu tài khoản ở trạng thái vô hiệu hóa tạm thời, có thể khôi phục khi đăng nhập lại.
  5. Thu hồi toàn bộ phiên đăng nhập đang hoạt động của người dùng.
  6. Trả về thông báo tiếng Việt "tài khoản đã được vô hiệu hóa tạm thời".
- **Entity**: User (người gọi, cần xác thực mật khẩu)
- **Model**: models.User (qua AuthRepository), models.UserSession
- **Databases**: users, user_sessions

---

### 20. Xem cài đặt giao diện (theme/language)

- **Name**: GetAppearance
- **Description**:
  1. Người dùng mở trang cài đặt giao diện với access token hợp lệ.
  2. Hệ thống xác thực người gọi; nếu không có → báo lỗi chưa đăng nhập.
  3. Đọc cài đặt của người dùng; nếu chưa từng chỉnh sửa thì dùng giá trị mặc định (giao diện sáng, ngôn ngữ tiếng Việt).
  4. Trả về giao diện và ngôn ngữ hiện tại.
  5. Các cài đặt này được lưu chung trong một bản ghi cài đặt cùng với quyền riêng tư và trạng thái hoạt động.
- **Entity**: User (người gọi)
- **Model**: models.UserSetting
- **Databases**: user_settings

---

### 21. Cập nhật cài đặt giao diện

- **Name**: UpdateAppearance
- **Description**:
  1. Người dùng chọn giao diện hoặc ngôn ngữ mới và gửi lên.
  2. Hệ thống xác thực người gọi; nếu không gửi cài đặt nào → báo lỗi dữ liệu không hợp lệ.
  3. Đọc cài đặt hiện tại; nếu chưa có thì khởi tạo với giá trị mặc định.
  4. Nếu có gửi giao diện: chuẩn hóa rồi chỉ cho phép giá trị sáng hoặc tối; sai → báo lỗi giao diện không hợp lệ.
  5. Nếu có gửi ngôn ngữ: chuẩn hóa rồi chỉ cho phép tiếng Việt hoặc tiếng Anh; sai → báo lỗi ngôn ngữ không hợp lệ.
  6. Lưu cài đặt của người dùng.
  7. Trả về cài đặt giao diện đã cập nhật.
- **Entity**: User (người gọi)
- **Model**: models.UserSetting
- **Databases**: user_settings

---

### 22. Liệt kê các phiên đăng nhập đang hoạt động

- **Name**: ListSessions
- **Description**:
  1. Người dùng mở trang quản lý phiên đăng nhập với access token hợp lệ.
  2. Hệ thống xác thực người gọi và nhận diện phiên của request hiện tại.
  3. Lấy danh sách phiên đăng nhập còn hiệu lực (chưa bị thu hồi, chưa hết hạn) của người dùng.
  4. Với mỗi phiên, ghi nhận thiết bị, địa chỉ IP, tác nhân trình duyệt, thời điểm tạo, thời điểm hết hạn, thời điểm hoạt động gần nhất và đánh dấu phiên đang dùng request hiện tại.
  5. Trả về danh sách các phiên; nếu chưa có dữ liệu phiên thì trả danh sách rỗng.
- **Entity**: User (người gọi), Session (các phiên đăng nhập của user)
- **Model**: models.UserSession
- **Databases**: user_sessions

---

### 23. Thu hồi một phiên đăng nhập cụ thể

- **Name**: RevokeSession
- **Description**:
  1. Người dùng chọn một phiên đăng nhập trong danh sách và bấm thu hồi.
  2. Hệ thống xác thực người gọi rồi lấy mã phiên cần thu hồi; nếu thiếu → báo lỗi dữ liệu không hợp lệ.
  3. Thu hồi phiên đó kèm kiểm tra quyền sở hữu: chỉ thu hồi được phiên thuộc về chính người gọi, tránh dùng mã phiên của người khác.
  4. Nếu phiên không tồn tại hoặc không thuộc người gọi → báo lỗi không tìm thấy phiên đăng nhập.
  5. Trả về thông báo tiếng Việt "đã thu hồi phiên đăng nhập".
  6. Người dùng cũng có thể thu hồi chính phiên đang dùng; khi đó client tự lo phần đăng nhập lại.
- **Entity**: User (người gọi), Session (phiên bị thu hồi)
- **Model**: models.UserSession
- **Databases**: user_sessions

---

### 24. Đăng xuất các thiết bị khác

- **Name**: RevokeOtherSessions
- **Description**:
  1. Người dùng bấm nút "Đăng xuất tất cả thiết bị khác" trong trang quản lý bảo mật.
  2. Hệ thống xác thực người gọi và nhận diện phiên đăng nhập hiện tại của request.
  3. Thu hồi toàn bộ phiên đăng nhập của người dùng, trừ phiên hiện tại đang thao tác.
  4. Người dùng vẫn đăng nhập trên thiết bị đang dùng, còn các thiết bị khác bị đăng xuất ngay vì refresh token của chúng mất hiệu lực.
  5. Nếu chưa có dữ liệu phiên thì bỏ qua việc thu hồi, vẫn coi là thành công.
  6. Trả về thông báo tiếng Việt "đã đăng xuất các thiết bị khác".
- **Entity**: User (người gọi), Session (các phiên khác bị thu hồi)
- **Model**: models.UserSession
- **Databases**: user_sessions

---

### 25. Xem trạng thái online/offline của một người dùng

- **Name**: GetPresence
- **Description**:
  1. Người dùng xem trạng thái của một người khác với access token hợp lệ.
  2. Hệ thống xác thực người gọi rồi lấy mã người cần xem; nếu thiếu → báo lỗi dữ liệu không hợp lệ.
  3. Kiểm tra quy tắc hai chiều: cả người xem và người được xem đều phải bật hiển thị trạng thái hoạt động thì mới thấy nhau.
  4. Tiếp theo kiểm tra mức hiển thị thời điểm hoạt động gần nhất của người được xem: tắt với tất cả thì không thấy; chỉ tin nhắn riêng thì chỉ thấy nếu hai người có hộp thoại riêng; chỉ bạn bè thì chỉ thấy nếu đã kết bạn.
  5. Nếu không được phép thấy → trả trạng thái ngoại tuyến với thời điểm hoạt động gần nhất bị ẩn, không tiết lộ lý do.
  6. Nếu được thấy → đọc trạng thái lưu trong bộ nhớ của máy chủ, được cập nhật real-time khi người dùng kết nối hoặc ngắt kết nối.
  7. Nếu chưa có trong bộ nhớ thì đọc dữ liệu lưu tạm trong database làm nguồn dự phòng.
  8. Trả về mã người dùng, trạng thái online/offline và thời điểm hoạt động gần nhất.
  9. Trạng thái online/offline phản ánh việc người dùng có đang mở kết nối real-time hay không; khi ngắt kết nối thì hệ thống đánh dấu ngoại tuyến.
- **Entity**: Viewer (người gọi), Target user
- **Model**: models.UserPresence, models.PresenceCacheEntry
- **Databases**: presences (fallback last_seen), user_settings (cờ visibility), friends, chats (kiểm tra quan hệ)

---

### 26. Lấy presence theo batch (nhiều user)

- **Name**: BatchGetPresence
- **Description**:
  1. Người dùng gọi lấy trạng thái của nhiều người cùng lúc và gửi danh sách mã người dùng cần xem.
  2. Hệ thống xác thực người gọi; nếu danh sách rỗng → báo lỗi dữ liệu không hợp lệ.
  3. Lặp qua từng người và áp dụng đầy đủ quy tắc quyền riêng tư như khi xem trạng thái của một người.
  4. Nếu một người gặp lỗi thì bỏ qua người đó, vẫn trả kết quả cho các người còn lại.
  5. Trả về kết quả dưới dạng ánh xạ từ mã người dùng sang trạng thái của người đó.
  6. Hữu ích cho màn hình danh sách bạn bè hoặc hộp thoại, giảm số lần gọi hơn là xem từng người một.
- **Entity**: Viewer (người gọi), Target users (nhiều)
- **Model**: models.UserPresence, models.PresenceCacheEntry
- **Databases**: presences, user_settings, friends, chats

---

### 27. Liệt kê danh sách user đang online

- **Name**: GetOnlineUsers
- **Description**:
  1. Người dùng (hoặc trang quản trị) gọi lấy danh sách người đang online với access token hợp lệ.
  2. Hệ thống xác thực yêu cầu rồi đọc toàn bộ bộ nhớ trạng thái của máy chủ.
  3. Lọc ra những người đang có kết nối real-time đang mở.
  4. Trả về danh sách mã người dùng đang online, không kèm thời điểm hoạt động gần nhất.
  5. Đây là danh sách toàn hệ thống, không lọc theo quyền riêng tư như khi xem trạng thái của một người.
  6. Bộ nhớ được cập nhật real-time; khi máy chủ khởi động lại thì danh sách trống cho tới khi các kết nối được mở lại.
- **Entity**: User (người gọi), tất cả user online
- **Model**: models.PresenceCacheEntry (in-memory, không có row DB trực tiếp)
- **Databases**: (không đọc DB trực tiếp — cache in-memory)

---

### 28. Đếm số user đang online

- **Name**: GetOnlineCount
- **Description**:
  1. Người dùng gọi lấy số lượng người đang online với access token hợp lệ.
  2. Hệ thống đọc toàn bộ bộ nhớ trạng thái của máy chủ.
  3. Đếm số đối tượng đang có kết nối real-time đang mở.
  4. Trả về con số tổng số người đang online.
  5. Không cần mã người dùng cụ thể, chỉ trả về một con số tổng.
  6. Việc đếm quét toàn bộ bộ nhớ nên tổng có thể bao gồm cả người đã ngắt kết nối nhưng chưa bị xóa khỏi bộ nhớ.
- **Entity**: Tất cả user online trên hệ thống
- **Model**: models.PresenceCacheEntry (in-memory)
- **Databases**: (không đọc DB — cache in-memory)

---

### 29. Xem cài đặt presence (activity status + last seen visibility)

- **Name**: GetPresenceSettings
- **Description**:
  1. Người dùng mở trang cài đặt trạng thái hoạt động với access token hợp lệ.
  2. Hệ thống xác thực người gọi; nếu không có → báo lỗi chưa đăng nhập.
  3. Đọc cài đặt của người dùng; nếu chưa từng chỉnh sửa thì dùng giá trị mặc định: bật hiển thị trạng thái hoạt động và thời điểm hoạt động gần nhất hiển thị cho bạn bè.
  4. Trả về hai cài đặt: bật/tắt hiển thị trạng thái hoạt động và mức độ ai được thấy thời điểm hoạt động gần nhất (tất cả bạn bè, chỉ tin nhắn riêng hoặc không ai).
  5. Hai cài đặt này nằm cùng bản ghi cài đặt với quyền riêng tư và giao diện.
  6. Được dùng để lọc quyền xem trạng thái khi xem trạng thái của một người hoặc của nhiều người.
- **Entity**: User (người gọi)
- **Model**: models.UserSetting
- **Databases**: user_settings

---

### 30. Cập nhật cài đặt presence

- **Name**: UpdatePresenceSettings
- **Description**:
  1. Người dùng đổi cài đặt trạng thái hoạt động và gửi lên.
  2. Hệ thống xác thực người gọi rồi đọc cài đặt hiện tại để dùng làm giá trị giữ nguyên cho các trường người dùng không gửi.
  3. Nếu gửi bật/tắt hiển thị trạng thái thì dùng giá trị đó, nếu không thì giữ nguyên.
  4. Nếu gửi mức hiển thị thời điểm hoạt động gần nhất thì dùng giá trị đó, nếu không thì giữ nguyên.
  5. Cập nhật hai trường cờ này trong bản ghi cài đặt của người dùng.
  6. Trả về hai cài đặt đã cập nhật.
  7. Lưu ý: tắt hiển thị trạng thái hoạt động thì người khác không thấy người này online và người này cũng không thấy người khác; mức hiển thị thời điểm hoạt động gần nhất quyết định ai được thấy thời điểm đó.
- **Entity**: User (người gọi)
- **Model**: models.UserSetting
- **Databases**: user_settings

---

### 31. Liệt kê danh sách tỉnh/thành phố

- **Name**: ListProvinces
- **Description**:
  1. Người dùng mở danh sách tỉnh/thành phố (endpoint công khai, không cần đăng nhập).
  2. Đọc toàn bộ danh sách tỉnh/thành phố từ dữ liệu được nhúng sẵn trong chương trình.
  3. Dữ liệu là 34 tỉnh/thành theo đơn vị hành chính Việt Nam sau sắp xếp 2025, mỗi phần tử gồm mã và tên tiếng Việt, tiếng Anh.
  4. Trả về danh sách theo thứ tự mã chính thức; dữ liệu tĩnh, không truy vấn database và không gọi dịch vụ ngoài.
  5. Chỉ có hai cấp là tỉnh/thành và phường/xã, không còn cấp quận/huyện.
- **Entity**: Không có user cụ thể (Public)
- **Model**: internal/locations.Place (embedded, không phải GORM model)
- **Databases**: (không truy vấn DB — data embedded trong Go source/binary qua locations.Store)

---

### 32. Liệt kê phường/xã theo tỉnh

- **Name**: ListWards
- **Description**:
  1. Người dùng chọn một tỉnh/thành để xem danh sách phường/xã (endpoint công khai, không cần đăng nhập).
  2. Hệ thống lấy mã tỉnh từ tham số truy vấn; nếu thiếu → báo lỗi dữ liệu không hợp lệ.
  3. Kiểm tra mã tỉnh có tồn tại trong dữ liệu nhúng hay không; không có → báo lỗi không tìm thấy tỉnh.
  4. Nếu hợp lệ → đọc danh sách phường/xã của tỉnh đó và trả về.
  5. Dữ liệu phường/xã theo sau sắp xếp 2025 (không còn cấp quận/huyện); mỗi phường/xã có tọa độ tâm dùng cho tính năng dò vị trí.
  6. Lỗi không tìm thấy tỉnh được xử lý riêng để trả đúng thông báo.
- **Entity**: Không có user cụ thể (Public)
- **Model**: internal/locations.Place, internal/locations.Ward (embedded)
- **Databases**: (không truy vấn DB — data embedded trong locations.Store)

---

### 33. Reverse geocode tọa độ → tỉnh/phường

- **Name**: ReverseGeocode
- **Description**:
  1. Người dùng bấm "dùng vị trí hiện tại" khi chỉnh sửa hồ sơ; hệ thống gửi tọa độ vĩ độ và kinh độ lên.
  2. Hệ thống xác thực người gọi rồi kiểm tra tọa độ có bị thiếu hoặc sai định dạng không; thiếu hoặc sai → báo lỗi dữ liệu không hợp lệ.
  3. Kiểm tra tọa độ nằm trong khoảng hợp lý (vĩ độ từ -90 đến 90, kinh độ từ -180 đến 180, không phải giá trị không xác định); sai → báo lỗi.
  4. Gọi dịch vụ dò địa chỉ Nominatim của OpenStreetMap để nhận chuỗi địa chỉ dạng phân tách dấu phẩy.
  5. Nếu dịch vụ lỗi hoặc trả về chuỗi rỗng thì không làm hỏng yêu cầu, tiếp tục ở chế độ tốt nhất có thể.
  6. Cắt chuỗi địa chỉ thành từng phần theo dấu phẩy.
  7. Tìm tỉnh: quét từ cuối lên, bỏ qua phần quốc gia, phần đầu tiên còn lại khớp với danh sách 34 tỉnh (so tên không dấu).
  8. Tìm phường/xã: lấy phần chi tiết nhất và khớp trong danh sách phường/xã của tỉnh đã tìm được.
  9. Nếu khớp theo tên thất bại thì tính khoảng cách tới tâm của các phường/xã, lấy phường gần nhất trong bán kính 50 km; nếu phường này thuộc tỉnh nào thì suy ra luôn tỉnh.
  10. Trả về cờ đã khớp cùng tỉnh và phường tìm được; nếu không khớp (tọa độ ngoài lãnh thổ hoặc dịch vụ không trả dữ liệu) thì cờ báo chưa khớp.
  11. Hữu ích cho tính năng dùng vị trí hiện tại khi chỉnh sửa hồ sơ, client gọi qua server để không phải gọi dịch vụ ngoài trực tiếp từ trình duyệt.
- **Entity**: User (người gọi, Auth)
- **Model**: internal/locations.GeoResult, internal/locations.Province, internal/locations.Ward (embedded)
- **Databases**: (không truy vấn DB — data embedded + gọi API Nominatim ngoài)

## B3 — Posts + Tag + Report + Block API Documentation

> Tài liệu hóa endpoint Posts, Tag, Report, Block. Mỗi endpoint gồm Name, Description, Entity, Model, Databases.

---

### 34. Lấy danh sách bài viết (Feed)

- **Name**: GetPosts
- **Description**:
  1. Người dùng mở trang bài viết và gọi API lấy danh sách bài (bắt buộc đăng nhập; không có thông tin xác thực thì báo chưa đăng nhập).
  2. Hệ thống đọc tham số phân trang, số bài mỗi trang (mặc định 10, giới hạn 1–100), bộ lọc theo người đang theo dõi (khi đó số bài mỗi trang bị giới hạn tối đa 5) và chế độ xếp hạng.
  3. Với xếp hạng cơ bản, hệ thống lấy các bài đang hiển thị theo điểm xếp hạng của trang, giữ nguyên mốc thời gian để bài không bị xáo trộn khi chuyển trang.
  4. Với xếp hạng cá nhân hóa (chỉ khi đăng nhập), hệ thống lấy bài theo hai giai đoạn gồm bài chính và các luồng khám phá, bảo đảm nội dung đa dạng; khi không có người dùng thì quay về xếp hạng cơ bản.
  5. Hệ thống tổng hợp cho cả trang: số lượt thích, bình luận, chia sẻ, trạng thái đã thích, đã lưu, đã chia sẻ của người xem, ảnh kèm bài và bài gốc nếu là chia sẻ lại.
  6. Trả về danh sách bài viết kèm thông tin cho trang kế tiếp; chỉ đọc dữ liệu, không ghi gì.
- **Entity**: người dùng đăng nhập (viewer), chủ bài viết, người follow
- **Model**: models.Post, models.Media
- **Databases**: posts, post_reactions, comments, post_shares, bookmarks, media

### 35. Lấy bài viết đã lưu (Bookmark)

- **Name**: GetSavedPosts
- **Description**:
  1. Người dùng mở mục bài viết đã lưu và gọi API lấy danh sách (bắt buộc đăng nhập; không có thông tin xác thực thì báo chưa đăng nhập).
  2. Hệ thống đọc số bài mỗi trang (mặc định 10, giới hạn 1–100) và tham số phân trang theo thời điểm lưu bài.
  3. Hệ thống lấy các bài mà người dùng đã đánh dấu lưu, sắp xếp theo thời điểm lưu mới nhất.
  4. Hệ thống tổng hợp số lượt thích, bình luận, chia sẻ, trạng thái đã thích, đã lưu, đã chia sẻ của người xem, ảnh kèm bài và bài gốc nếu là chia sẻ lại.
  5. Trả về danh sách bài đã lưu kèm thông tin cho trang kế tiếp; chỉ đọc dữ liệu, không ghi gì.
- **Entity**: người dùng đăng nhập, chủ bài viết
- **Model**: models.Post, models.Media
- **Databases**: bookmarks, posts, post_reactions, comments, post_shares, media

### 36. Lấy bài viết của một người dùng

- **Name**: GetUserPosts
- **Description**:
  1. Người dùng vào trang cá nhân của người khác và gọi API lấy bài viết của người đó (bắt buộc đăng nhập; không có thông tin xác thực thì báo chưa đăng nhập).
  2. Hệ thống kiểm tra định danh người dùng trong đường dẫn không được rỗng, đọc số bài mỗi trang (mặc định 10, giới hạn 1–100) và tham số phân trang theo thời điểm tạo bài.
  3. Hệ thống lấy các bài công khai đang hiển thị của người dùng đó, lọc trạng thái bài theo quan hệ với người xem và sắp xếp theo thời điểm tạo mới nhất.
  4. Hệ thống tổng hợp số lượt thích, bình luận, chia sẻ, trạng thái đã thích, đã lưu, đã chia sẻ của người xem, ảnh kèm bài và bài gốc nếu là chia sẻ lại.
  5. Trả về danh sách bài viết kèm thông tin cho trang kế tiếp.
- **Entity**: người xem (viewer), chủ profile
- **Model**: models.Post, models.Media
- **Databases**: posts, post_reactions, comments, post_shares, bookmarks, media

### 37. Xem chi tiết bài viết

- **Name**: ViewPostDetail
- **Description**:
  1. Người dùng mở chi tiết một bài viết; nếu gửi kèm thông tin xác thực hợp lệ thì hệ thống nhận diện người xem, không có thông tin xác thực vẫn xem được.
  2. Hệ thống kiểm tra định danh bài viết rồi lấy bài tương ứng; bài không tồn tại hoặc bị ẩn, riêng tư thì báo lỗi.
  3. Hệ thống nạp ảnh kèm bài và bài gốc nếu đây là chia sẻ lại.
  4. Nếu người xem đã đăng nhập, hệ thống ghi nhận lượt xem: mỗi người chỉ được tính một lượt mỗi ngày, ghi bản ghi xem nhưng lượt xem của chính chủ bài không tăng số đếm; đồng thời ghi nhận sở thích nền theo từ khóa của bài và tăng tổng số lượt xem.
  5. Trả về chi tiết bài viết; lỗi trong khi ghi nhận lượt xem không làm thất bại yêu cầu.
- **Entity**: người xem, chủ bài viết
- **Model**: models.Post, models.PostView, models.Media
- **Databases**: posts, post_views, media, tags, user_interests

### 38. Ghi nhận lượt xem bài viết (impression feed)

- **Name**: TrackPostView
- **Description**:
  1. Người dùng lướt danh sách bài viết hoặc mở chi tiết bài, hệ thống gọi API ghi nhận lượt xem (bắt buộc đăng nhập) kèm định danh bài và nguồn xem, mặc định là danh sách bài viết.
  2. Hệ thống kiểm tra bài; bài bị ẩn, riêng tư hoặc đã xóa thì trả về kết quả lượt xem không được tính, không báo lỗi.
  3. Nếu người dùng đã xem bài này trong ngày hôm nay thì trả về kết quả lượt xem không được tính.
  4. Hệ thống ghi bản ghi lượt xem kèm nguồn; lượt xem của chính chủ bài được ghi lại nhưng không tăng số đếm.
  5. Nền hệ thống ghi nhận sở thích của người dùng theo từ khóa của bài, với mức độ cao hơn khi xem từ chi tiết so với xem ở danh sách bài viết.
  6. Nếu lượt xem được tính thì tăng tổng số lượt xem của bài và trả về kết quả lượt xem có được tính hay không.
- **Entity**: người xem, chủ bài viết
- **Model**: models.PostView, models.Post
- **Databases**: post_views, posts, tags, user_interests

### 39. Tạo bài viết mới

- **Name**: CreatePost
- **Description**:
  1. Người dùng điền tiêu đề, nội dung, trạng thái, cộng đồng nếu có, tệp GIF, lựa chọn bật bình luận và chọn ảnh, video đính kèm rồi gửi đăng bài.
  2. Hệ thống kiểm tra dữ liệu biểu mẫu: tiêu đề, nội dung, trạng thái phải hợp lệ và bắt buộc có ít nhất một ảnh, video hoặc GIF.
  3. Nếu đăng trong cộng đồng, hệ thống kiểm tra người dùng có phải thành viên hợp lệ của cộng đồng đó không; không phải thành viên thì báo lỗi.
  4. Hệ thống kiểm tra khóa chống gửi trùng do giao diện gửi kèm; nếu đã có bài được tạo từ khóa này thì trả về bài cũ thay vì tạo lại.
  5. Hệ thống tạo bài viết mới và lưu vào cơ sở dữ liệu, rồi ghi khóa chống gửi trùng trỏ tới bài vừa tạo; nếu thua trong cuộc trùng lặp thì trả về bài của yêu cầu thắng.
  6. Hệ thống xử lý từng tệp media: kiểm tra và đưa vào chờ tải lên nền, tệp lỗi được gom thành cảnh báo không làm hỏng bài; gắn các tệp đã tải vào bài; GIF tải từ ngoài được ghi nhận như một mục media.
  7. Hệ thống hoàn tất bài: nạp ảnh kèm và tác giả, trích xuất từ khóa gắn thẻ cho bài; nếu bài thuộc cộng đồng thì xử lý thử thách và cộng điểm đóng góp nền.
  8. Trả về bài viết vừa tạo kèm danh sách cảnh báo nếu có.
- **Entity**: người tạo bài, thành viên cộng đồng (nếu community_id)
- **Model**: models.Post, models.Media, models.Tag
- **Databases**: posts, post_idempotency, media, tags, user_interests, member_contributions

### 40. Xóa bài viết

- **Name**: DeletePost
- **Description**:
  1. Người dùng bấm xóa bài viết của mình (bắt buộc đăng nhập).
  2. Hệ thống lấy bài theo định danh; bài không tồn tại thì báo không tìm thấy bài viết.
  3. Hệ thống kiểm tra người xóa có phải chủ bài không; không phải thì báo chỉ chủ bài mới xóa được.
  4. Hệ thống xóa bài cùng toàn bộ dữ liệu liên quan gồm cảm xúc, bình luận, chia sẻ, lưu bài, gắn thẻ, ảnh kèm bài, đồng thời thu về danh sách người đã lưu bài này.
  5. Nền hệ thống gửi thông báo cho từng người đã lưu bài, trừ chính chủ, báo bài viết trong danh sách dấu trang đã bị chủ sở hữu xóa.
  6. Trả về thông báo xóa bài thành công bằng tiếng Việt.
- **Entity**: chủ bài viết, người đã bookmark
- **Model**: models.Post
- **Databases**: posts, bookmarks, post_reactions, comments, post_shares, tags, media, notifications

### 41. Thả cảm xúc bài viết (toggle)

- **Name**: ReactPost
- **Description**:
  1. Người dùng chọn một biểu tượng cảm xúc và bấm thả cho bài viết (bắt buộc đăng nhập, phải chọn biểu tượng hợp lệ có trong hệ thống).
  2. Hệ thống kiểm tra người này đã thả đúng biểu tượng đó cho bài này chưa.
  3. Nếu đã thả rồi thì xóa cảm xúc đi; nếu là bài của người khác nằm trong cộng đồng thì nền hệ thống trừ điểm đóng góp và trả về kết quả đã gỡ cảm xúc.
  4. Nếu chưa thả thì tạo cảm xúc mới; nếu là bài của người khác thì gửi thông báo cho chủ bài rằng bạn đã thích bài viết, nếu nằm trong cộng đồng thì nền hệ thống cộng điểm đóng góp.
  5. Nền hệ thống ghi nhận sở thích của người dùng theo từ khóa của bài.
  6. Trả về kết quả đã thả hay đã gỡ kèm thông báo tiếng Việt theo biểu tượng.
- **Entity**: người thả cảm xúc, chủ bài viết
- **Model**: models.PostReaction, models.Emoji, models.Post
- **Databases**: post_reactions, emojis, posts, notifications, member_contributions, tags, user_interests

### 42. Tạo bình luận

- **Name**: CreateComment
- **Description**:
  1. Người dùng nhập nội dung bình luận, nếu là trả lời thì chọn bình luận cha rồi gửi (bắt buộc đăng nhập).
  2. Hệ thống kiểm tra dữ liệu bình luận rồi lấy bài tương ứng; bài bị ẩn hoặc riêng tư thì báo lỗi, bài đã tắt bình luận thì báo chủ bài đã tắt bình luận.
  3. Nếu là trả lời, hệ thống kiểm tra bình luận cha có tồn tại và thuộc đúng bài này không; không hợp lệ thì báo lỗi.
  4. Hệ thống tạo bình luận mới và lưu vào cơ sở dữ liệu, đồng thời trích xuất từ khóa gắn thẻ trong nội dung bình luận.
  5. Nền hệ thống ghi nhận sở thích người dùng; nếu bài thuộc cộng đồng thì cộng điểm chất lượng bình luận; gửi thông báo cho chủ bài khi không phải mình bình luận; gửi thông báo cho người trả lời khi không trùng với chủ bài hoặc người bình luận.
  6. Trả về toàn bộ danh sách bình luận của bài viết.
- **Entity**: người comment, chủ bài viết, người trả lời (parent author)
- **Model**: models.Comment, models.Post, models.Tag
- **Databases**: comments, posts, tags, notifications, member_contributions, user_interests

### 43. Lấy danh sách bình luận

- **Name**: GetComments
- **Description**:
  1. Người dùng mở phần bình luận của một bài viết (không bắt buộc đăng nhập) kèm tham số trang, số bình luận mỗi trang và cách sắp xếp theo mới nhất, cũ nhất hoặc phù hợp nhất; mặc định là mới nhất nếu giá trị lạ.
  2. Hệ thống kiểm tra bài viết; bài không tồn tại, bị ẩn hoặc riêng tư thì báo không truy cập được.
  3. Hệ thống chuẩn hóa phân trang, đếm tổng số bình luận và lấy trang dữ liệu tương ứng, đồng thời đánh dấu trạng thái đã thích của người xem nếu có đăng nhập.
  4. Trả về danh sách bình luận theo trang kèm tổng số; chỉ đọc dữ liệu, không ghi gì.
- **Entity**: người đọc bình luận (public), chủ bài viết
- **Model**: models.Comment
- **Databases**: comments, posts

### 44. Thả cảm xúc bình luận (toggle)

- **Name**: ToggleCommentReaction
- **Description**:
  1. Người dùng thả biểu tượng cảm xúc cho một bình luận (bắt buộc đăng nhập, phải chọn biểu tượng hợp lệ).
  2. Hệ thống kiểm tra bình luận có tồn tại và đang hiển thị không; không có thì báo không tìm thấy.
  3. Nếu người này đã thả cảm xúc cho bình luận này rồi thì xóa cảm xúc và giảm số lượt thích của bình luận, trả về kết quả đã gỡ.
  4. Nếu chưa thì tạo cảm xúc mới và tăng số lượt thích của bình luận, trả về kết quả đã tạo.
  5. Mỗi người chỉ có một cảm xúc cho mỗi bình luận nên đổi biểu tượng thì hệ thống gỡ cái cũ rồi tạo cái mới; trả về kết quả đã tạo hay đã gỡ.
- **Entity**: người thả cảm xúc, bình luận
- **Model**: models.CommentReaction
- **Databases**: comment_reactions (comment likes), comments

### 45. Chia sẻ bài viết (repost)

- **Name**: SharePost (post)
- **Description**:
  1. Người dùng bấm chia sẻ một bài viết, có thể kèm lời chia sẻ hoặc để trống (bắt buộc đăng nhập).
  2. Hệ thống kiểm tra bài gốc; bài không tồn tại, bị ẩn hoặc riêng tư thì báo không thể chia sẻ; chia sẻ bài của chính mình thì báo lỗi; người này đã chia sẻ bài đó rồi thì báo đã chia sẻ.
  3. Hệ thống tạo một bài chia sẻ mới ở chế độ công khai trỏ tới bài gốc kèm lời chia sẻ, đồng thời ghi nhận bản ghi chia sẻ để đếm số lượt và đánh dấu trạng thái đã chia sẻ.
  4. Nền hệ thống gửi thông báo cho chủ bài gốc, ghi nhận sở thích của người dùng và trích xuất từ khóa từ lời chia sẻ để gắn thẻ.
  5. Hệ thống nạp ảnh kèm, tác giả và số liệu của bài chia sẻ để giao diện chèn ngay đầu danh sách bài viết.
  6. Trả về thông báo tiếng Việt kèm bài chia sẻ vừa tạo.
- **Entity**: người share, chủ bài gốc
- **Model**: models.Post (repost), models.PostShare
- **Databases**: posts, post_shares, tags, notifications, user_interests

### 46. Lưu/bỏ lưu bài viết (Bookmark toggle)

- **Name**: SavePost
- **Description**:
  1. Người dùng bấm lưu hoặc bỏ lưu một bài viết (bắt buộc đăng nhập).
  2. Hệ thống kiểm tra bài; bài không tồn tại, bị ẩn hoặc riêng tư thì báo không tìm thấy bài viết; lưu bài của chính mình thì báo lỗi.
  3. Nếu bài đã được lưu rồi thì hủy lưu và trả về kết quả đã bỏ lưu.
  4. Nếu chưa lưu thì lưu bài mới và trả về kết quả đã lưu; nền hệ thống ghi nhận sở thích của người dùng theo từ khóa của bài.
  5. Trả về kết quả kèm thông báo tiếng Việt.
- **Entity**: người bookmark, chủ bài viết
- **Model**: models.Bookmark, models.Post
- **Databases**: bookmarks, posts, tags, user_interests

### 47. Tìm bài viết theo Hashtag (posts)

- **Name**: GetPostsByHashtag (posts)
- **Description**:
  1. Người dùng mở danh sách bài viết theo một từ khóa gắn thẻ (không bắt buộc đăng nhập) kèm tham số trang và số bài mỗi trang.
  2. Hệ thống chuẩn hóa phân trang và tìm các bài viết được gắn thẻ đó.
  3. Không có bài nào gắn thẻ đó thì trả về danh sách rỗng, không báo lỗi; nếu có thì lấy các bài theo danh sách và phân trang.
  4. Trả về từ khóa kèm danh sách bài viết; chỉ đọc dữ liệu.
- **Entity**: khách truy cập, chủ bài viết
- **Model**: models.Post, models.Tag
- **Databases**: tags, posts

### 48. Lấy danh sách Emoji

- **Name**: GetEmojis
- **Description**:
  1. Người dùng mở bộ chọn biểu tượng cảm xúc (không bắt buộc đăng nhập) kèm bộ lọc phạm vi, danh mục, từ khóa tìm kiếm, số lượng và vị trí bắt đầu.
  2. Nếu phạm vi là cảm xúc bài viết thì hệ thống trả trọn bộ biểu tượng cảm xúc mà không phân trang; ngược lại giới hạn số lượng từ 1 đến 100, mặc định 50, và vị trí bắt đầu không âm.
  3. Hệ thống đếm và lấy danh sách biểu tượng theo bộ lọc.
  4. Trả về danh sách biểu tượng, tổng số và cờ còn trang kế tiếp, luôn là không khi phạm vi là cảm xúc bài viết.
- **Entity**: khách truy cập (picker emoji)
- **Model**: models.Emoji
- **Databases**: emojis

### 49. Ghim bài viết

- **Name**: PinPost
- **Description**:
  1. Người dùng bấm ghim một bài viết của mình (bắt buộc đăng nhập).
  2. Hệ thống lấy bài theo định danh; bài không tồn tại thì báo lỗi.
  3. Hệ thống kiểm tra người ghim có phải chủ bài không; không phải thì báo không có quyền, chỉ chủ bài mới ghim được.
  4. Hệ thống đánh dấu bài đã ghim và trả về thông báo đã ghim bài viết.
- **Entity**: chủ bài viết
- **Model**: models.Post
- **Databases**: posts

### 50. Bỏ ghim bài viết

- **Name**: UnpinPost
- **Description**:
  1. Người dùng bấm bỏ ghim một bài viết của mình (bắt buộc đăng nhập).
  2. Hệ thống lấy bài theo định danh; bài không tồn tại thì báo lỗi.
  3. Hệ thống kiểm tra người bỏ ghim có phải chủ bài không; không phải thì báo không có quyền, chỉ chủ bài mới bỏ ghim được.
  4. Hệ thống gỡ trạng thái ghim của bài và trả về thông báo đã bỏ ghim bài viết.
- **Entity**: chủ bài viết
- **Model**: models.Post
- **Databases**: posts

### 51. Bật/tắt bình luận của bài viết

- **Name**: SetCommentsEnabled
- **Description**:
  1. Người dùng bật hoặc tắt tính năng bình luận trên bài viết của mình (bắt buộc đăng nhập).
  2. Hệ thống lấy bài theo định danh; bài không tồn tại thì báo không tìm thấy bài viết; không phải chủ sở hữu thì báo bị từ chối.
  3. Hệ thống cập nhật trạng thái bật tắt bình luận của bài rồi lấy lại bài vừa cập nhật.
  4. Trả về thông báo đã bật hoặc tắt bình luận kèm bài viết mới nhất; khi tắt, các yêu cầu bình luận cho bài này sẽ bị chặn.
- **Entity**: chủ bài viết
- **Model**: models.Post
- **Databases**: posts

### 52. Lấy media của một người dùng

- **Name**: GetUserMedia (posts)
- **Description**:
  1. Người dùng mở mục ảnh video của một người dùng khác (bắt buộc đăng nhập) kèm tham số trang và số mục mỗi trang, mặc định 20, giới hạn 1–50.
  2. Hệ thống kiểm tra định danh người dùng cần xem rồi đếm tổng số media của người đó qua các bài viết.
  3. Hệ thống lấy trang media tương ứng, sắp xếp và phân trang theo các bài viết thuộc người dùng đó.
  4. Trả về danh sách media kèm tổng số, trang hiện tại và cờ còn trang kế tiếp; chỉ đọc dữ liệu.
- **Entity**: người xem (đăng nhập), chủ media
- **Model**: models.Media
- **Databases**: media, posts

### 53. Tag Controller — Bài viết theo Hashtag

- **Name**: GetPostsByHashtag (tags)
- **Description**:
  1. Người dùng gọi API lấy định danh các bài viết theo một từ khóa gắn thẻ (không bắt buộc đăng nhập).
  2. Hệ thống kiểm tra tên từ khóa trong đường dẫn không được rỗng; rỗng thì báo lỗi.
  3. Hệ thống tìm các bài viết được gắn thẻ đó và thu về danh sách định danh bài viết duy nhất.
  4. Trả về từ khóa kèm danh sách định danh bài viết, không kèm nội dung bài, để giao diện dùng danh sách này gọi API lấy chi tiết từng bài.
- **Entity**: khách truy cập
- **Model**: models.Tag
- **Databases**: tags

### 54. Tạo báo cáo vi phạm

- **Name**: CreateReport
- **Description**:
  1. Người dùng chọn đối tượng báo cáo là người dùng, bài viết hoặc bình luận, rồi chọn loại vi phạm, nhập mô tả chi tiết và tùy chọn nội quy vi phạm rồi gửi báo cáo (bắt buộc đăng nhập).
  2. Hệ thống kiểm tra dữ liệu báo cáo.
  3. Nếu chọn nội quy vi phạm, hệ thống kiểm tra nội quy có tồn tại, đang hiệu lực và phù hợp với loại đối tượng không; không phù hợp thì báo lỗi.
  4. Hệ thống chống spam: nếu người này đã có báo cáo đang chờ xử lý cho cùng đối tượng thì báo nên chỉnh sửa báo cáo cũ thay vì tạo mới.
  5. Với đối tượng là người dùng: không được tự báo mình, không được báo quản trị viên, người đã bị cấm thì báo lỗi.
  6. Với đối tượng là bài viết: bài không tồn tại, bị ẩn hoặc là bài của chính mình thì báo lỗi.
  7. Với đối tượng là bình luận: bình luận không tồn tại hoặc bị ẩn thì báo lỗi.
  8. Hệ thống tạo báo cáo mới ở trạng thái đang chờ xử lý, trỏ tới đúng đối tượng được báo và lưu vào cơ sở dữ liệu.
  9. Trả về thông báo báo cáo đã được gửi thành công; không gửi thông báo nào đi kèm.
- **Entity**: người báo cáo, đối tượng bị báo (user/post/comment), admin xử lý (downstream)
- **Model**: models.Report, models.ViolationRule
- **Databases**: reports, violation_rules, users, posts, comments

### 55. Chỉnh sửa báo cáo

- **Name**: UpdateReport
- **Description**:
  1. Người dùng chỉnh sửa loại vi phạm, mô tả lý do và nội quy của báo cáo mình đã gửi (bắt buộc đăng nhập).
  2. Hệ thống kiểm tra định danh báo cáo rồi lấy báo cáo tương ứng; không tồn tại thì báo lỗi.
  3. Hệ thống kiểm tra quyền: không phải người tạo báo cáo thì báo bạn không có quyền chỉnh sửa báo cáo này.
  4. Hệ thống kiểm tra trạng thái: báo cáo không còn ở trạng thái đang chờ xử lý thì báo chỉ chỉnh sửa được báo cáo đang chờ xử lý.
  5. Hệ thống xác định loại đối tượng từ báo cáo hiện có rồi kiểm tra nội quy vi phạm mới có phù hợp không.
  6. Hệ thống cập nhật loại vi phạm, mô tả lý do và nội quy vi phạm, rồi trả về thông báo báo cáo đã được cập nhật thành công.
- **Entity**: người báo cáo (chủ report), violation rule
- **Model**: models.Report, models.ViolationRule
- **Databases**: reports, violation_rules

### 56. Chặn/bỏ chặn người dùng (toggle)

- **Name**: ToggleBlock
- **Description**:
  1. Người dùng bấm chặn hoặc bỏ chặn một người dùng khác (bắt buộc đăng nhập, không được tự chặn chính mình).
  2. Hệ thống kiểm tra người bị chặn có phải quản trị viên không; nếu có thì báo không được phép chặn quản trị viên.
  3. Nếu trước đây đã chặn người này thì gỡ lệnh chặn và trả về kết quả đã bỏ chặn.
  4. Nếu chưa chặn thì tạo lệnh chặn mới và trả về kết quả đã chặn.
  5. Trả về kết quả kèm thông báo tiếng Việt; việc chặn được áp dụng để lọc ở danh sách bài viết, trò chuyện và các truy vấn khác.
- **Entity**: người chặn, người bị chặn
- **Model**: models.Block
- **Databases**: blocks, user_roles

### 57. Danh sách người dùng đã chặn

- **Name**: GetBlockedUsers
- **Description**:
  1. Người dùng mở danh sách người mình đã chặn (bắt buộc đăng nhập).
  2. Hệ thống lấy toàn bộ lệnh chặn của người dùng này, sắp xếp theo thời điểm chặn.
  3. Hệ thống chuyển mỗi lệnh chặn sang thông tin người bị chặn kèm thời điểm chặn.
  4. Trả về danh sách người dùng đã chặn; chỉ đọc dữ liệu, không ghi gì.
- **Entity**: người xem (đang chặn), người bị chặn
- **Model**: models.Block
- **Databases**: blocks

---

**Tổng số mục**: 24 (34–57) — Posts (34–52), Tag (53), Report (54–55), Block (56–57).

## B4 — Stories + Follow + Friend (Endpoints 58–81)

Tài liệu hóa chi tiết 24 endpoint thuộc 3 nhóm tính năng: Stories (`routes/story.routes.go`), Follow (`routes/follow.routes.go`), Friend (`routes/friend.routes.go`). Code paste nguyên văn từ controller/service.

---

## Nhóm Stories (`/api/stories`)

### 58. Đăng tải Story

- **Name**: CreateStory
- **Description**:
  1. Người dùng chọn tải lên ảnh/video hoặc nhập nội dung text để đăng story.
  2. Nếu không có cả file lẫn nội dung text thì báo lỗi nội dung story là bắt buộc.
  3. Xác định định dạng file: ảnh (jpg, jpeg, png) hoặc video (mp4, mov, webm); nếu là định dạng khác thì báo lỗi định dạng không hợp lệ.
  4. Tải file lên hệ thống với chế độ tự động duyệt, không qua kiểm duyệt nội dung, và lấy đường dẫn file đã lưu.
  5. Nếu story chỉ có text, không có file thì nội dung text bắt buộc phải khác rỗng và story không có loại phương tiện.
  6. Tạo bản ghi story mới với thời gian tạo là thời điểm hiện tại và thời gian hết hạn sau 24 giờ, rồi lưu vào database.
  7. Với mỗi file được tải lên thì tạo một story riêng, nên một lần đăng có thể tạo nhiều story.
  8. Trả về thông báo đăng tải story thành công kèm danh sách các story đã tạo.
- **Entity**: User (uploader), Media
- **Model**: models.Story (models.StoryMediaType), dto.CreateStoryResponse
- **Databases**: stories, media

### 59. Feed Story trang chủ

- **Name**: GetHomeFeed
- **Description**:
  1. Hệ thống xác định người xem, có thể ẩn danh, và bộ lọc phạm vi: tất cả hoặc chỉ những người mình đang theo dõi.
  2. Lọc các story còn hiệu lực, chưa hết hạn 24 giờ, theo quyền riêng tư: bỏ story của người đã bị người xem ẩn, của người đã chặn hoặc ẩn người xem; nếu chọn lọc theo dõi thì chỉ giữ story của những người mình đang theo dõi và của chính mình.
  3. Gom danh sách các story mà người xem đã xem trước đó để gắn cờ đã xem cho từng story.
  4. Tải thông tin hồ sơ của các chủ story theo lô để tránh truy vấn lặp từng người.
  5. Gom story theo từng người dùng thành từng nhóm gồm thông tin chủ và danh sách story, sắp xếp mới nhất trước.
  6. Trả về danh sách các nhóm story; người xem ẩn danh vẫn xem được toàn bộ story sau khi đã lọc.
- **Entity**: Viewer (có thể ẩn danh), chủ Story
- **Model**: models.Story, models.Profile, dto.StoryFeedGroup, dto.StoryResponse
- **Databases**: stories, story_views, story_mutes, blocks, follows, profiles

### 60. Kiểm tra user có Story active

- **Name**: CheckUserStory
- **Description**:
  1. Nhận mã người dùng cần kiểm tra từ đường dẫn yêu cầu.
  2. Nếu mã người dùng rỗng thì báo lỗi đầu vào không hợp lệ.
  3. Kiểm tra người dùng đó có story nào còn hiệu lực, chưa hết hạn 24 giờ, hay không.
  4. Trả về kết quả người dùng có story đang hiển thị hay không; kết quả này dùng để hiển thị viền hoặc chấm story quanh ảnh đại diện.
- **Entity**: chủ Story
- **Model**: models.Story
- **Databases**: stories

### 61. Lấy danh sách Story của 1 user

- **Name**: GetUserStories
- **Description**:
  1. Xác định người xem, ẩn danh hoặc đã đăng nhập, để tính trạng thái đã xem.
  2. Lấy toàn bộ story còn hiệu lực của người dùng được chỉ định, sắp xếp theo thời gian tạo tăng dần.
  3. Tải thông tin hồ sơ của chủ story, gồm tên hiển thị và ảnh đại diện, kèm theo từng mục.
  4. Nếu người xem không phải chủ story thì kiểm tra từng story đã được xem hay chưa để gắn cờ đã xem cho từng cái riêng lẻ.
  5. Trả về danh sách story của người dùng đó.
- **Entity**: chủ Story, Viewer (ẩn danh hoặc đăng nhập)
- **Model**: models.Story, models.Profile, dto.StoryResponse
- **Databases**: stories, story_views, profiles

### 62. Mở xem Story (ghi nhận view)

- **Name**: ViewStory
- **Description**:
  1. Tìm story theo mã; nếu không tồn tại thì báo lỗi không tìm thấy story.
  2. Nếu người xem chính là chủ story thì không ghi nhận lượt xem và trả trực tiếp thông tin story.
  3. Nếu người xem khác chủ story: kiểm tra xem người này đã xem story này chưa; nếu chưa xem thì ghi nhận lượt xem, có cập nhật thời gian xem nếu bản ghi đã tồn tại; nếu đã xem rồi thì bỏ qua, lượt xem không tăng thêm.
  4. Trả về chi tiết của story.
- **Entity**: Viewer, chủ Story
- **Model**: models.Story, repository.StoryView
- **Databases**: stories, story_views

### 63. Tương tác với Story (react / reply / share)

- **Name**: Interact
- **Description**:
  1. Người dùng gửi yêu cầu tương tác với story, kiểu tương tác là react, reply hoặc share; nếu dữ liệu gửi lên không hợp lệ thì báo lỗi.
  2. Tìm story cần tương tác; nếu không tồn tại thì báo lỗi không tìm thấy story để tương tác.
  3. Với kiểu react: bắt buộc phải chọn biểu tượng cảm xúc; nếu người dùng đã react story này rồi thì tăng số lần bấm, tối đa 5 lần, vượt quá thì báo lỗi đã đạt giới hạn, và cập nhật lại biểu tượng cảm xúc; nếu chưa react thì kiểm tra biểu tượng cảm xúc có tồn tại hay không, rồi tạo bản ghi tương tác mới và gửi thông báo react cho chủ story.
  4. Với kiểu reply: bắt buộc có nội dung khác rỗng, tạo bản ghi tương tác và gửi thông báo bình luận cho chủ story, kèm nội dung phản hồi.
  5. Với kiểu share: tạo bản ghi tương tác và gửi thông báo chia sẻ cho chủ story.
  6. Không gửi thông báo khi người dùng tương tác với story của chính mình.
- **Entity**: Người tương tác, chủ Story
- **Model**: models.StoryInteract, models.NotificationType, dto.InteractStoryRequest
- **Databases**: story_interacts, stories, emojis, notifications (notification qua NotificationService — có thể đẩy qua notification_preferences/push_tokens)

### 64. Thống kê Analytics Story

- **Name**: GetAnalytics
- **Description**:
  1. Tìm story theo mã; nếu không tồn tại thì báo lỗi không tìm thấy story.
  2. Kiểm tra quyền riêng tư: chỉ chủ story mới được xem thống kê; người khác bị báo lỗi không có quyền xem.
  3. Đếm tổng số lượt xem, và đếm số lượt react, reply, chia sẻ theo từng loại.
  4. Lấy chi tiết người xem và các tương tác, gom theo từng người dùng để mỗi người chỉ có một dòng: gán kèm biểu tượng cảm xúc và số lần bấm đã react, đồng thời chỉ giữ phản hồi mới nhất.
  5. Tải thông tin hồ sơ theo lô để hiển thị tên và ảnh đại diện của từng người xem, người tương tác.
  6. Trả về kết quả thống kê chi tiết của story.
- **Entity**: Chủ Story (viewer duy nhất được phép), người xem/tương tác
- **Model**: models.Story, models.StoryInteract, repository.StoryView, dto.StoryAnalyticsResponse, dto.ViewerDetailResponse
- **Databases**: stories, story_views, story_interacts, profiles

### 65. Xóa Story

- **Name**: DeleteStory
- **Description**:
  1. Tìm story theo mã; nếu không tồn tại thì báo lỗi không tìm thấy story.
  2. Kiểm tra quyền sở hữu: chỉ chủ story mới được xóa; người khác bị báo lỗi không có quyền xóa.
  3. Xóa dữ liệu liên quan của story, gồm lượt tương tác và lượt xem, cùng với chính story, trong một giao dịch.
  4. Trả về thông báo xóa story thành công.
- **Entity**: Chủ Story
- **Model**: models.Story, models.StoryInteract, repository.StoryView
- **Databases**: stories, story_views, story_interacts

### 66. Bật/tắt ẩn Story của 1 user (mute)

- **Name**: ToggleMute (stories)
- **Description**:
  1. Người dùng chọn ẩn hoặc hiện story của một người dùng khác.
  2. Kiểm tra nếu cố ẩn story của chính mình thì báo lỗi không được ẩn story của bản thân.
  3. Kiểm tra xem người dùng này đã bị ẩn story hay chưa.
  4. Nếu ĐÃ ẩn rồi thì bỏ ghi nhận ẩn và trả về trạng thái đã hiện lại; story của người đó sẽ xuất hiện trở lại trong feed.
  5. Nếu CHƯA ẩn thì ghi nhận ẩn và trả về trạng thái đã ẩn; toàn bộ story của người đó sẽ bị loại khỏi feed của mình.
- **Entity**: Viewer (người mute), User bị mute
- **Model**: models.StoryMute
- **Databases**: story_mutes

---

## Nhóm Follow (`/api/follow`)

### 67. Follow / Unfollow (toggle)

- **Name**: FollowToggle
- **Description**:
  1. Người dùng bấm follow hoặc unfollow một người dùng khác.
  2. Validate mã người được follow: bắt buộc khác rỗng và không được là chính mình.
  3. Kiểm tra người được follow có tồn tại và đang hoạt động hay không.
  4. Không cho follow quản trị viên cấp cao; với quản trị viên thông thường thì hạn chế follow, báo lỗi tương ứng.
  5. Kiểm tra đã follow người này chưa: nếu ĐÃ follow thì hủy theo dõi và trả về trạng thái đã unfollow; nếu CHƯA follow thì tạo quan hệ theo dõi mới và trả về trạng thái đã follow.
  6. Khi follow thành công thì gửi thông báo cho người được follow.
  7. Trả về kết quả thao tác kèm số lượng người theo dõi và số người đang theo dõi được cập nhật mới nhất.
- **Entity**: Follower (người follow), Following (bị follow), Admin/SuperAdmin (bị chặn)
- **Model**: models.Follow, dto.FollowToggleResponse
- **Databases**: follows, users, roles, user_roles, notifications

### 68. Thống kê Follow

- **Name**: GetFollowStats
- **Description**:
  1. Nhận mã người dùng cần xem thống kê; người xem phải đã đăng nhập.
  2. Đếm số người đang theo dõi người dùng này và số người mà người dùng này đang theo dõi.
  3. Nếu người xem đã đăng nhập thì kiểm tra xem mình có đang theo dõi người này không để gắn cờ tương ứng.
  4. Trả về thống kê theo dõi kèm cờ đang theo dõi hay chưa.
- **Entity**: Viewer, User được xem thống kê
- **Model**: dto.FollowStatsResponse
- **Databases**: follows

### 69. Gợi ý Follow

- **Name**: GetSuggestions
- **Description**:
  1. Chuẩn hóa phân trang: số trang tối thiểu 1, kích thước trang trong khoảng 1 đến 20, mặc định 5.
  2. Lọc những người dùng đủ điều kiện để gợi ý: không phải chính mình, đang hoạt động, mình chưa follow, không có quan hệ chặn hai chiều và không phải quản trị viên.
  3. Sắp xếp ưu tiên những người có nhiều người chung đang theo dõi nhất, tính qua số người mình đang theo dõi mà họ cũng theo dõi; nếu bằng nhau thì sắp theo mã người dùng.
  4. Trả về danh sách gợi ý follow kèm phân trang và cờ còn trang sau hay không.
- **Entity**: Viewer, user được gợi ý, Admin/SuperAdmin (loại trừ), user bị block (loại trừ)
- **Model**: dto.FollowSuggestionsResponse, dto.FollowSuggestionItem
- **Databases**: users, profiles, follows, blocks, roles, user_roles

### 70. Danh sách Followers

- **Name**: GetFollowers
- **Description**:
  1. Đây là route công khai, không yêu cầu đăng nhập; nếu người xem đã đăng nhập thì dùng để tính cờ đang theo dõi cho từng mục, còn ẩn danh thì cờ luôn là chưa theo dõi.
  2. Chuẩn hóa phân trang: số trang tối thiểu 1, kích thước trang trong khoảng 1 đến 50, mặc định 20.
  3. Đếm tổng số người theo dõi của người dùng được chỉ định và lấy trang danh sách follower kèm thông tin tên người dùng, tên hiển thị và ảnh đại diện.
  4. Với mỗi follower, kiểm tra xem người xem có đang theo dõi người đó không để gắn cờ tương ứng.
  5. Trả về danh sách follower kèm phân trang gồm trang, kích thước trang, tổng số và cờ còn trang sau hay không.
- **Entity**: Viewer (có thể ẩn danh), target User, followers
- **Model**: dto.FollowListResponse, dto.FollowListItem
- **Databases**: follows, users, profiles

### 71. Danh sách Following

- **Name**: GetFollowing
- **Description**:
  1. Đây là route công khai, không yêu cầu đăng nhập; người xem ẩn danh thì không tính cờ đang theo dõi.
  2. Chuẩn hóa phân trang: số trang tối thiểu 1, kích thước trang trong khoảng 1 đến 50, mặc định 20.
  3. Đếm tổng số người mà người dùng này đang theo dõi và lấy trang danh sách kèm thông tin tên người dùng, tên hiển thị và ảnh đại diện của từng người.
  4. Với mỗi người trong danh sách, kiểm tra xem người xem có đang theo dõi người đó không để gắn cờ tương ứng.
  5. Trả về danh sách người đang theo dõi kèm phân trang.
- **Entity**: Viewer (có thể ẩn danh), target User, danh sách user được target follow
- **Model**: dto.FollowListResponse, dto.FollowListItem
- **Databases**: follows, users, profiles

### 72. Danh sách Follow chung (mutual)

- **Name**: GetMutualFollows
- **Description**:
  1. Người dùng bắt buộc đã đăng nhập để xem follow chung; nếu chọn chính mình làm đối tượng so sánh thì trả về danh sách rỗng.
  2. Tìm những người mà cả người xem lẫn người dùng được chọn đều đang theo dõi, tức là tính giao của hai danh sách theo dõi.
  3. Tải thông tin hồ sơ, gồm tên hiển thị và ảnh đại diện, của những người chung này.
  4. Trả về danh sách follow chung kèm tổng số, giới hạn tối đa 10 người.
- **Entity**: Viewer, Target User, user chung
- **Model**: dto.FollowListItem
- **Databases**: follows, users, profiles

---

## Nhóm Friend (`/api/friend-requests`, `/api/friends`, `/api/friend-status`)

### 73. Danh sách Lời mời kết bạn

- **Name**: GetFriendRequests
- **Description**:
  1. Hệ thống lấy hai danh sách lời mời kết bạn đang chờ xử lý của người dùng: những lời mời mình đã GỬI và những lời mời mình NHẬN.
  2. Với mỗi lời mời, tải thông tin hồ sơ của người còn lại: lời mời đã gửi thì hiển thị hồ sơ người nhận, lời mời đã nhận thì hiển thị hồ sơ người gửi; bỏ qua nếu không có hồ sơ.
  3. Gắn hướng của từng lời mời là đã gửi hoặc đã nhận.
  4. Trả về danh sách lời mời kết bạn gồm hai nhóm: đã gửi và đã nhận.
- **Entity**: User (chủ tài khoản), người gửi/nhận lời mời
- **Model**: models.Friend (models.FriendStatusPending), models.Profile, dto.FriendRequestListResponse
- **Databases**: friends, profiles

### 74. Danh sách Bạn bè

- **Name**: GetFriends
- **Description**:
  1. Chuẩn hóa phân trang: số trang tối thiểu 1, kích thước trang trong khoảng 1 đến 50, mặc định 20.
  2. Lấy danh sách bạn bè đã chấp nhận, tức là những quan hệ mà mình là người gửi hoặc người nhận.
  3. Xác định người bạn còn lại trong từng cặp kết bạn và tải thông tin tên hiển thị, ảnh đại diện của họ; sắp xếp theo thời gian kết bạn mới nhất trước.
  4. Trả về danh sách bạn bè kèm phân trang gồm trang, kích thước trang, tổng số và cờ còn trang sau hay không.
- **Entity**: User, bạn bè đã chấp nhận
- **Model**: models.Friend (models.FriendStatusAccepted), dto.FriendListResponse, dto.FriendItem
- **Databases**: friends, profiles

### 75. Gửi / Thu hồi Lời mời kết bạn (toggle)

- **Name**: ToggleFriendRequest
- **Description**:
  1. Người dùng gửi hoặc thu hồi lời mời kết bạn tới một người dùng khác.
  2. Validate: không được tự gửi lời mời kết bạn cho chính mình.
  3. Kiểm tra người được gửi lời mời có tồn tại và đang hoạt động hay không.
  4. Không cho gửi lời mời kết bạn tới quản trị viên, báo lỗi đã bị hạn chế.
  5. Kiểm tra lời mời hiện có giữa hai người: nếu lời mời đã được xử lý rồi, không còn ở trạng thái chờ, thì báo lỗi lời mời đã được xử lý trước đó; nếu lời mời đang chờ thì xóa lời mời và trả về trạng thái đã thu hồi; nếu chưa có lời mời nào thì tạo lời mời mới đang chờ xử lý, lưu thời gian theo giờ UTC, gửi thông báo lời mời kết bạn cho người nhận và trả về trạng thái đã gửi.
  6. Trả về kết quả thao tác lời mời kết bạn.
- **Entity**: Người gửi lời mời, target User, Admin/SuperAdmin (bị chặn)
- **Model**: models.Friend (models.FriendStatusPending), dto.FriendRequestResponse
- **Databases**: friends, users, roles, user_roles, notifications

### 76. Chấp nhận Lời mời kết bạn

- **Name**: AcceptFriendRequest
- **Description**:
  1. Người dùng chọn chấp nhận một lời mời kết bạn theo mã lời mời.
  2. Tìm lời mời; nếu không tồn tại thì báo lỗi không tìm thấy lời mời.
  3. Kiểm tra quyền: chỉ người NHẬN lời mời mới được chấp nhận; người khác bị báo lỗi không có quyền xử lý.
  4. Kiểm tra lời mời phải đang ở trạng thái chờ; nếu đã được xử lý rồi thì báo lỗi đã xử lý trước đó.
  5. Cập nhật trạng thái lời mời thành đã chấp nhận.
  6. Gửi thông báo cho người GỬI lời mời để biết lời mời đã được chấp nhận.
  7. Trả về kết quả đã chấp nhận.
- **Entity**: Người nhận lời mời (receiver), người gửi lời mời (sender)
- **Model**: models.Friend (models.FriendStatusAccepted), dto.FriendActionResponse
- **Databases**: friends, notifications

### 77. Từ chối Lời mời kết bạn

- **Name**: RejectFriendRequest
- **Description**:
  1. Người dùng chọn từ chối một lời mời kết bạn theo mã lời mời.
  2. Tìm lời mời; nếu không tồn tại thì báo lỗi không tìm thấy lời mời.
  3. Kiểm tra quyền: chỉ người NHẬN lời mời mới được từ chối; người khác bị báo lỗi không có quyền xử lý.
  4. Kiểm tra lời mời phải đang ở trạng thái chờ; nếu đã được xử lý rồi thì báo lỗi đã xử lý trước đó.
  5. Khác với chấp nhận: từ chối sẽ xóa hẳn lời mời khỏi hệ thống, không giữ lại ở trạng thái đã từ chối.
  6. Không gửi thông báo khi từ chối lời mời.
  7. Trả về kết quả đã từ chối.
- **Entity**: Người nhận lời mời (receiver), người gửi lời mời (sender)
- **Model**: models.Friend, dto.FriendActionResponse
- **Databases**: friends

### 78. Tìm kiếm trong danh sách Bạn bè

- **Name**: SearchFriends
- **Description**:
  1. Người dùng nhập từ khóa để tìm kiếm trong danh sách bạn bè của mình.
  2. Bắt buộc đã đăng nhập; từ khóa sau khi bỏ khoảng trắng thừa phải khác rỗng, nếu không thì báo lỗi đầu vào không hợp lệ.
  3. Chỉ tìm trong nhóm bạn bè ĐÃ chấp nhận; khớp theo tên người dùng hoặc tên hiển thị theo kiểu chứa từ khóa.
  4. Sắp xếp kết quả theo tên hiển thị tăng dần và giới hạn tối đa 10 kết quả.
  5. Trả về danh sách bạn bè khớp tìm kiếm kèm mã người dùng, tên người dùng, tên hiển thị và ảnh đại diện.
- **Entity**: User, bạn bè đã accept
- **Model**: dto.UserSearchResult
- **Databases**: friends, users, profiles

### 79. Gợi ý Kết bạn

- **Name**: GetFriendSuggestions
- **Description**:
  1. Chuẩn hóa phân trang: số trang tối thiểu 1, kích thước trang trong khoảng 1 đến 50, mặc định 20.
  2. Lọc những người dùng đủ điều kiện để gợi ý: không phải chính mình, đang hoạt động, chưa có quan hệ bạn bè, kể cả đã chấp nhận lẫn đang chờ, không có quan hệ chặn hai chiều và không phải quản trị viên.
  3. Tính số bạn chung cho từng người được gợi ý, dựa trên quan hệ bạn bè của cả hai phía; sắp xếp ưu tiên người có nhiều bạn chung nhất.
  4. Trả về danh sách gợi ý kết bạn kèm phân trang gồm trang, kích thước trang, tổng số và cờ còn trang sau hay không.
- **Entity**: Viewer, user được gợi ý, bạn chung, Admin/SuperAdmin (loại trừ), user bị block (loại trừ)
- **Model**: dto.FriendSuggestionsResponse, dto.FriendSuggestionItem
- **Databases**: users, profiles, friends, blocks, roles, user_roles

### 80. Hủy kết bạn

- **Name**: Unfriend
- **Description**:
  1. Người dùng chọn hủy kết bạn với một người bạn.
  2. Validate: mã người bạn bắt buộc khác rỗng và không được tự hủy kết bạn với chính mình.
  3. Kiểm tra quan hệ bạn bè giữa hai người theo cặp hai chiều, bất kể hướng gửi lời mời; nếu không có quan hệ hoặc chưa ở trạng thái đã chấp nhận thì báo lỗi hai người không phải bạn bè.
  4. Xóa quan hệ bạn bè đã chấp nhận của cặp hai người.
  5. Trả về kết quả đã hủy kết bạn.
- **Entity**: User, bạn bè bị hủy
- **Model**: models.Friend, dto.FriendActionResponse
- **Databases**: friends

### 81. Kiểm tra Quan hệ bạn bè với 1 user

- **Name**: GetFriendStatus
- **Description**:
  1. Người dùng kiểm tra quan hệ bạn bè với một người dùng khác; bắt buộc đã đăng nhập.
  2. Nếu kiểm tra chính mình thì trả về trạng thái là bản thân.
  3. Tìm quan hệ bạn bè theo cặp hai chiều; nếu không có thì trả về trạng thái không có quan hệ.
  4. Nếu quan hệ đã được chấp nhận thì trả về trạng thái đã kết bạn.
  5. Nếu lời mời đang chờ thì phân biệt hướng: mình là người gửi thì trả về trạng thái đã gửi kèm mã lời mời; mình là người nhận thì trả về trạng thái đã nhận kèm mã lời mời, client dùng mã này để gọi chấp nhận hoặc từ chối.
  6. Nếu ở trạng thái khác không xác định thì trả về không có quan hệ; trả về kết quả trạng thái quan hệ bạn bè.
- **Entity**: User, target User, lời mời kết bạn (nếu có)
- **Model**: models.Friend (FriendStatusAccepted/Pending), dto.FriendStatusResponse
- **Databases**: friends

---

## Ghi chú chung

- Nhóm Friend dùng **2 route group**: `/api/friend-requests` (lời mời: list, toggle, accept, reject — #73/#75/#76/#77) và `/api/friends` (bạn bè: list, search, suggestions, unfriend — #74/#78/#79/#80); plus route đứng riêng `/api/friend-status/:userID` (#81). Tất cả đều qua `AuthMiddleware` (group-level hoặc từng route).
- Follow/Friend đều gửi notification qua `NotificationService` (bảng `notifications`, có thể đẩy tiếp qua `notification_preferences`/`push_tokens`).
- Toggle pattern: Follow (#67), Friend request (#75), Story mute (#66) — tồn tại → xóa, chưa có → tạo.
- Block check: feed story (#59) và suggestions (#69, #79) lọc 2 chiều qua bảng `blocks`.

## B5 — Media + Notification + Search + Ads Packages + Customer Ads API Documentation

> Tài liệu hóa endpoint Media, Notification, Search, Ads Packages, Customer Ads. Mỗi endpoint gồm Name, Description, Entity, Model, Databases.

---

### 82. Tải lên media (upload Cloudinary)

- **Name**: UploadMedia
- **Description**:
  1. Người dùng chọn và tải file lên (yêu cầu đăng nhập bằng token).
  2. Hệ thống kiểm tra file đã được đính kèm hay chưa; thiếu file thì báo lỗi nhập liệu.
  3. Kiểm tra loại file và dung lượng file có hợp lệ theo giới hạn của hệ thống không.
  4. Kiểm tra dung lượng lưu trữ còn trống của người dùng; nếu đã đầy thì báo lỗi dung lượng đã đầy.
  5. Tải file lên dịch vụ lưu trữ đám mây và tạo bản ghi media với trạng thái được phê duyệt ngay lập tức.
  6. Cộng dồn dung lượng đã sử dụng vào hồ sơ người dùng và lấy lại thông tin dung lượng hiện tại.
  7. Trả về thông tin media vừa tải lên kèm thông tin dung lượng (tổng, đã dùng, còn trống).
- **Entity**: User (đăng nhập), Media
- **Model**: models.Media
- **Databases**: media (INSERT + SELECT quota/used), profiles (UPDATE storage_used_bytes)

---

### 83. Xóa media (destroy Cloudinary)

- **Name**: DeleteMedia
- **Description**:
  1. Người dùng chọn xóa một media của mình (yêu cầu đăng nhập).
  2. Hệ thống kiểm tra media cần xóa có tồn tại không; nếu không thì báo lỗi không tìm thấy.
  3. Kiểm tra người dùng có phải chủ sở hữu media không; nếu không thì báo lỗi không có quyền xóa.
  4. Xóa file vật lý trên dịch vụ lưu trữ đám mây (bỏ qua nếu không phân tích được định danh file).
  5. Xóa bản ghi media và trừ ngược dung lượng đã sử dụng trong hồ sơ người dùng.
  6. Trả về thông báo xóa thành công kèm mã media đã xóa.
- **Entity**: User (chủ media), Media
- **Model**: models.Media
- **Databases**: media (SELECT + DELETE), profiles (UPDATE storage_used_bytes)

---

### 84. Kiểm tra trạng thái storage

- **Name**: GetStorageStatus
- **Description**:
  1. Người dùng mở màn hình kiểm tra dung lượng lưu trữ (yêu cầu đăng nhập).
  2. Hệ thống lấy tổng dung lượng được cấp và dung lượng đã sử dụng từ hồ sơ người dùng.
  3. Tính dung lượng còn trống (không âm nếu dư) và phần trăm dung lượng đã sử dụng.
  4. Trả về tổng dung lượng, dung lượng đã dùng, dung lượng còn trống và phần trăm sử dụng.
- **Entity**: User, Profile
- **Model**: models.Profile (không có model Media trực tiếp)
- **Databases**: profiles (SELECT storage_quota_bytes, storage_used_bytes)

---

### 85. Lấy danh sách media của user

- **Name**: GetUserMedia (media)
- **Description**:
  1. Người dùng mở danh sách media của mình (yêu cầu đăng nhập).
  2. Hệ thống xác thực phiên đăng nhập để xác định người dùng đang xem.
  3. Hệ thống truy vấn toàn bộ media thuộc người dùng đang đăng nhập.
  4. Trả về danh sách media kèm các thông tin như loại tệp, dung lượng, trạng thái, thời lượng và thời điểm tạo.
- **Entity**: User, Media
- **Model**: models.Media
- **Databases**: media (SELECT WHERE user_id)

---

### 86. Lấy danh sách thông báo (phân trang)

- **Name**: GetNotifications
- **Description**:
  1. Người dùng mở màn hình thông báo, có thể chọn trang, số lượng mỗi trang và chỉ xem thông báo chưa đọc (yêu cầu đăng nhập).
  2. Hệ thống chuẩn hóa thông số phân trang và truy vấn thông báo của người dùng theo thứ tự mới nhất trước.
  3. Đếm tổng số thông báo khớp bộ lọc để trả về kèm trang hiện tại.
  4. Tải thông tin hồ sơ (tên hiển thị, ảnh đại diện) của những người gửi thông báo trong trang hiện tại.
  5. Trả về danh sách thông báo đã phân trang kèm tổng số và trang hiện tại.
- **Entity**: User (receiver + sender), Notification, Profile
- **Model**: models.Notification, models.Profile
- **Databases**: notifications (SELECT + COUNT), profiles (SELECT WHERE user_id IN)

---

### 87. Lấy summary thông báo (unread-count + 5 mới nhất)

- **Name**: GetSummary
- **Description**:
  1. Người dùng mở khu vực tóm tắt thông báo như dropdown thông báo (yêu cầu đăng nhập).
  2. Hệ thống đếm số thông báo chưa đọc của người dùng.
  3. Lấy 5 thông báo mới nhất kèm thông tin người gửi (tên hiển thị, ảnh đại diện).
  4. Đảm bảo danh sách xem trước luôn trả về danh sách rỗng nếu không có thông báo nào.
  5. Trả về số lượng chưa đọc và 5 thông báo mới nhất trong một lần gọi để hiển thị trên thanh thông báo.
- **Entity**: User (receiver + sender), Notification, Profile
- **Model**: models.Notification, models.Profile
- **Databases**: notifications (COUNT unread + SELECT LIMIT 5), profiles (SELECT WHERE user_id IN)

---

### 88. Đánh dấu 1 thông báo đã đọc

- **Name**: MarkAsRead
- **Description**:
  1. Người dùng chọn một thông báo để đánh dấu đã đọc (yêu cầu đăng nhập).
  2. Hệ thống xác thực phiên đăng nhập của người dùng.
  3. Hệ thống cập nhật trạng thái đã đọc, chỉ giới hạn trên thông báo của chính người dùng để không đánh dấu được thông báo của người khác.
  4. Trả về thông báo cập nhật thành công.
- **Entity**: User (receiver), Notification
- **Model**: models.Notification
- **Databases**: notifications (UPDATE is_read = true)

---

### 89. Đánh dấu tất cả thông báo đã đọc

- **Name**: MarkAllAsRead
- **Description**:
  1. Người dùng chọn đánh dấu tất cả thông báo đã đọc (yêu cầu đăng nhập).
  2. Hệ thống xác thực phiên đăng nhập của người dùng.
  3. Hệ thống đánh dấu tất cả thông báo chưa đọc của người dùng là đã đọc.
  4. Trả về thông báo cập nhật thành công.
- **Entity**: User (receiver), Notification
- **Model**: models.Notification
- **Databases**: notifications (UPDATE is_read = true WHERE receiver_id)

---

### 90. Đếm thông báo chưa đọc

- **Name**: GetUnreadCount
- **Description**:
  1. Người dùng tải số thông báo chưa đọc (yêu cầu đăng nhập).
  2. Hệ thống xác thực phiên đăng nhập của người dùng.
  3. Hệ thống đếm số thông báo chưa đọc của người dùng.
  4. Trả về số lượng thông báo chưa đọc; giá trị này cũng dùng làm huy hiệu đếm trên thông báo đẩy iOS.
- **Entity**: User (receiver), Notification
- **Model**: models.Notification
- **Databases**: notifications (COUNT unread)

---

### 91. Lấy preferences thông báo

- **Name**: GetPreferences
- **Description**:
  1. Người dùng mở cài đặt thông báo (yêu cầu đăng nhập).
  2. Hệ thống lấy cấu hình ưu tiên thông báo của người dùng, gồm 10 loại: thích, bình luận, theo dõi, tin nhắn, kết bạn, cộng đồng, gọi thoại, tương tác story, chia sẻ và media.
  3. Nếu người dùng chưa từng cấu hình thì trả về trạng thái mặc định coi như tất cả các loại đều được bật.
  4. Trả về cấu hình thông báo hiện tại của người dùng.
- **Entity**: User, NotificationPreference
- **Model**: models.NotificationPreference
- **Databases**: notification_preferences (SELECT WHERE user_id)

---

### 92. Cập nhật preferences thông báo

- **Name**: UpdatePreferences
- **Description**:
  1. Người dùng bật hoặc tắt các loại thông báo trong cài đặt (yêu cầu đăng nhập); hệ thống kiểm tra dữ liệu gửi lên phải có ít nhất một loại cấu hình.
  2. Nếu người dùng chưa có cấu hình nào thì khởi tạo mặc định tất cả các loại đều bật.
  3. Chỉ cập nhật các loại được người dùng gửi lên; các loại bỏ trống giữ nguyên giá trị hiện tại.
  4. Lưu cấu hình vào cơ sở dữ liệu bằng cách chèn mới hoặc cập nhật.
  5. Trả về thông báo cập nhật thành công; từ sau các loại bị tắt sẽ không được tạo thông báo nữa.
- **Entity**: User, NotificationPreference
- **Model**: models.NotificationPreference
- **Databases**: notification_preferences (SELECT + UPSERT)

---

### 93. Đăng ký push token (Expo)

- **Name**: RegisterPushToken
- **Description**:
  1. Người dùng đăng ký nhận thông báo đẩy (yêu cầu đăng nhập); hệ thống kiểm tra phải có token đẩy.
  2. Hệ thống tạo bản ghi token với nền tảng Expo.
  3. Nếu token đã tồn tại thì cập nhật, ngược lại thêm mới vào danh sách token.
  4. Trả về thông báo đăng ký thành công; token được dùng để gửi thông báo đẩy và tự xóa khi token không còn hiệu lực.
- **Entity**: User, PushToken
- **Model**: models.PushToken
- **Databases**: push_tokens (UPSERT)

---

### 94. Tìm kiếm (users/posts/hashtags/communities)

- **Name**: Search
- **Description**:
  1. Người dùng nhập từ khóa và chọn loại tìm kiếm: người dùng, bài viết, hashtag hoặc cộng đồng (không cần đăng nhập).
  2. Hệ thống kiểm tra từ khóa và loại tìm kiếm hợp lệ; loại mặc định là tìm kiếm tất cả.
  3. Tìm kiếm theo loại đã chọn; nếu chọn tất cả thì tìm kiếm đồng thời trên cả bốn loại, lỗi ở một loại không làm hỏng cả yêu cầu.
  4. Nếu không có kết quả ở tất cả các loại thì gắn thông báo không tìm thấy kết quả phù hợp.
  5. Trả về kết quả tìm kiếm theo từng loại kèm thông báo kết quả.
- **Entity**: User, Post, Community (không có auth actor — public)
- **Model**: models.User, models.Post, models.Community (qua SearchRepository)
- **Databases**: users, posts, communities (JOIN tag/hashtag trong repository — SELECT WHERE keyword LIKE)

---

### 95. Lấy hashtag trending

- **Name**: GetTrending
- **Description**:
  1. Người dùng mở danh sách hashtag thịnh hành (không cần đăng nhập).
  2. Hệ thống kiểm tra bộ nhớ đệm có còn hiệu lực hay không (thời hạn 5 phút); nếu còn hạn thì trả dữ liệu từ bộ nhớ đệm ngay.
  3. Nếu bộ nhớ đệm hết hạn thì truy vấn hashtag xuất hiện nhiều nhất trong các bài viết gần đây.
  4. Nếu truy vấn thất bại mà vẫn còn dữ liệu cũ thì trả dữ liệu cũ; thất bại hoàn toàn thì báo lỗi hệ thống.
  5. Lưu kết quả mới vào bộ nhớ đệm và trả về danh sách hashtag thịnh hành.
- **Entity**: Post/Hashtag (public, không user actor)
- **Model**: (qua SearchRepository — đếm từ posts + tag)
- **Databases**: posts, tags (SELECT GROUP BY hashtag ORDER BY count DESC LIMIT)

---

### 96. Lấy danh sách gói quảng cáo

- **Name**: GetPackages
- **Description**:
  1. Người dùng, kể cả khi chưa đăng nhập, mở danh sách các gói quảng cáo để tham khảo và chọn trước khi đăng ký.
  2. Hệ thống truy vấn toàn bộ gói quảng cáo có trong hệ thống, không yêu cầu phiên đăng nhập.
  3. Hệ thống tổng hợp thông tin từng gói: tên, giá theo tháng, số slot tối đa, thời hạn tối đa, hỗ trợ video hoặc carousel và các tính năng đi kèm.
  4. Trả về danh sách gói để người dùng so sánh và lựa chọn.
- **Entity**: AdPackage (danh mục public)
- **Model**: models.AdPackage
- **Databases**: ad_packages (SELECT ALL)

---

### 97. Đăng ký gói quảng cáo (subscribe)

- **Name**: Subscribe
- **Description**:
  1. Người dùng chọn một gói quảng cáo và đăng ký (yêu cầu đăng nhập; mọi vai trò đều được phép nâng cấp).
  2. Hệ thống kiểm tra gói đăng ký có tồn tại không; nếu không thì báo lỗi không tìm thấy.
  3. Tính thời hạn sử dụng dựa trên số ngày tối đa của gói và tạo bản đăng ký mới với trạng thái đang hoạt động, tự động gia hạn, số slot đã dùng bằng 0.
  4. Trong một giao dịch: hủy gói đang hoạt động nếu có, lưu gói mới và nâng cấp vai trò người dùng lên đối tác để truy cập các chức năng quảng cáo.
  5. Cấp lại token đăng nhập để vai trò mới có hiệu lực ngay; nếu cấp lại thất bại thì client dùng cơ chế làm mới token.
  6. Trả về thông báo đăng ký gói quảng cáo thành công kèm thông tin gói vừa đăng ký và token mới nếu có.
- **Entity**: User (Partner-candidate), AdPackage, PartnerSubscription
- **Model**: models.PartnerSubscription, models.AdPackage
- **Databases**: ad_packages (SELECT), partner_subscriptions (SELECT active + UPDATE old + INSERT new), user_roles (UPDATE role → PARTNER trong transaction)

---

### 98. Lấy subscription hiện tại của user

- **Name**: GetMySubscription
- **Description**:
  1. Người dùng xem gói đăng ký hiện tại của mình (yêu cầu đăng nhập).
  2. Hệ thống truy vấn gói đang hoạt động của người dùng kèm thông tin gói đã đăng ký.
  3. Nếu không có gói đang hoạt động thì báo lỗi chưa đăng ký gói.
  4. Tính số slot còn lại (không âm nếu dư) và dựng thông tin gói: tên, tổng slot, slot đã dùng, slot còn lại, giá theo tháng, ngày bắt đầu, ngày hết hạn và trạng thái.
  5. Trả về thông tin gói đăng ký hiện tại.
- **Entity**: User, PartnerSubscription, AdPackage
- **Model**: models.PartnerSubscription, models.AdPackage
- **Databases**: partner_subscriptions (SELECT JOIN ad_packages WHERE status active)

---

### 99. Lấy feed quảng cáo cho customer (targeting + frequency cap)

- **Name**: GetUserFeed
- **Description**:
  1. Người dùng đăng nhập mở luồng hiển thị quảng cáo (mọi vai trò đều xem được, không giới hạn đối tác hoặc quản trị).
  2. Hệ thống lấy thông tin nhân khẩu học gồm giới tính, tuổi và khu vực từ hồ sơ người dùng; nếu lỗi hoặc chưa có hồ sơ thì coi như tất cả.
  3. Truy vấn các quảng cáo đang hoạt động, còn trong thời gian chạy, còn ngân sách và khớp với đối tượng mục tiêu (giới tính, độ tuổi, khu vực), sắp xếp theo mức độ phù hợp và lấy tối đa 50 quảng cáo.
  4. Lọc theo tần suất: loại bỏ quảng cáo mà người dùng đã thấy trong thời gian gần đây, giữ tối đa 20 quảng cáo; nếu lọc thất bại thì trả nguyên danh sách để không làm hỏng luồng.
  5. Trả về danh sách quảng cáo kèm nội dung media.
- **Entity**: User (customer), Ad, AdAnalytics (đọc frequency), Profile (demographics)
- **Model**: models.Ad, models.Profile
- **Databases**: ads (SELECT ranked active), ad_analytics (SELECT recent impressions cho frequency cap), profiles (SELECT demographics)

---

### 100. Track click/impression quảng cáo (ad_analytics + trừ ngân sách)

- **Name**: TrackAction
- **Description**:
  1. Hệ thống ghi nhận sự kiện hiển thị hoặc nhấp chuột trên một quảng cáo khi người dùng đăng nhập tương tác; loại sự kiện do client gửi.
  2. Trích địa chỉ IP của người dùng để ghi kèm vào log sự kiện.
  3. Kiểm tra quảng cáo có tồn tại không; nếu không thì báo lỗi không tìm thấy.
  4. Chống gian lận: nếu người dùng đã ghi nhận cùng loại sự kiện trên quảng cáo này trong 24 giờ qua thì bỏ qua, không ghi log trùng và không trừ tiền.
  5. Nếu quảng cáo có ngân sách theo ngày và tổng chi tiêu hôm nay đã đạt giới hạn thì bỏ qua việc trừ ngân sách.
  6. Ghi log sự kiện và trừ ngân sách quảng cáo theo chi phí hiển thị hoặc nhấp chuột trong một giao dịch, chỉ trừ khi ngân sách còn đủ.
  7. Trả về xác nhận đã ghi nhận; nếu lỗi liên quan đến quảng cáo không tồn tại thì báo lỗi tương ứng.
- **Entity**: User (customer — gắn dấu vết impression/click), Ad, AdAnalytics
- **Model**: models.Ad, models.AdAnalytics
- **Databases**: ads (SELECT + UPDATE budget/deduct), ad_analytics (SELECT HasRecentAction/GetDailySpend + INSERT log)

---

**Hết tài liệu B5 — 19 endpoint (82-100): Media (4), Notification (8), Search (2), Ads Packages (3), Customer Ads (2).**

## B6 — Chat HTTP + Chat WebSocket + E2E Encryption API Documentation

> Tài liệu hóa endpoint Chat (HTTP), Chat WebSocket real-time và E2E encryption. Mỗi endpoint gồm Name, Description, Entity, Model, Databases.

- Route chat: `routes/chat.routes.go` — nhánh HTTP `router.Group("/api/chats")` gắn `middlewares.AuthMiddleware(env, db)`; endpoint WS `GET /api/chats/ws` được đăng ký TRƯỚC group nên không đi qua AuthMiddleware, handler tự xác thực bằng `?token=`.
- Route E2E: `routes/e2e.routes.go` — `router.Group("/api/e2e")` gắn `middlewares.AuthMiddleware(env, db)` (Auth cho toàn bộ 9 endpoint E2E).
- Wire trong `cmd/main.go`: `chatHub := ws.NewHub()` → `NewChatController(chatHub, chatService, mediaService, postRepository, env)` → `e2eService.SetChatHub(chatHub)` (phát event `chat:e2e_key_updated` cho client đang online trên cùng hub chat).
- Mã hóa: tin nhắn legacy (`e2e_version = 0`) server mã hóa AES-256-GCM bằng `chats.encryption_key`; tin E2E (`e2e_version = 1`) client mã hóa AES-256-GCM bằng khóa chat per-chat (bọc qua ECDH) — server chỉ lưu/relay ciphertext.

---

### 101. Danh sách chat trực tiếp

- **Name**: ListChats
- **Description**:
  1. Người dùng mở màn hình tin nhắn, hệ thống xác thực rồi lấy danh sách các cuộc trò chuyện trực tiếp của người dùng.
  2. Hệ thống thu thập mỗi cuộc chat trực tiếp kèm thông tin đối phương, tin nhắn cuối chưa bị xóa, loại file và thời lượng của tin cuối, cùng nền tảng chat riêng của người dùng trong cuộc đó.
  3. Sắp xếp danh sách theo thời gian hoạt động gần nhất của tin nhắn cuối (không có tin cuối thì theo thời điểm tạo chat) để chat mới hoạt động lên đầu.
  4. Đánh dấu cuộc chat có mã hóa đầu cuối nếu đã tồn tại khóa mã hóa được tạo cho chat đó (không phụ thuộc tin nhắn cuối, vì chat mã hóa đầu cuối vẫn có thể còn tin cũ khi đối phương chưa tạo lại khóa cho lịch sử).
  5. Giải mã trước nội dung tin nhắn cuối bằng khóa mã hóa lưu của chat để hiển thị bản xem trước; tin nhắn mã hóa đầu cuối giữ nguyên dạng đã mã hóa để ứng dụng của người dùng tự giải mã.
  6. Trả về danh sách cuộc trò chuyện kèm đối tác, tin nhắn cuối, nền tảng, thời điểm cập nhật và trạng thái mã hóa.
- **Entity**: người dùng (chủ danh sách), đối tác chat
- **Model**: models.Chat, models.ChatParticipant, models.Message, models.Media, models.Profile, models.ChatUserSettings
- **Databases**: chats, chat_participants, messages, media, profiles, chat_user_settings, chat_e2e_keys

---

### 102. WebSocket chat (real-time, mã hóa đầu cuối)

- **Name**: HandleWebsocket (chat)
- **Description**:
  1. **Kết nối**: Ứng dụng mở kết nối WebSocket tới endpoint chat; hệ thống đọc token truy cập từ chuỗi truy vấn, xác thực token (sai hoặc hết hạn → từ chối kết nối, chỉ nhận token truy cập), rồi nâng cấp kết nối HTTP sang WebSocket; mọi nguồn truy cập đều được chấp nhận. Hệ thống đưa kết nối vào phòng toàn cục, đăng ký theo người dùng, bắt đầu chạy nền gửi nhịp tim định kỳ (mỗi khoảng 54 giây, tự đóng nếu quá 60 giây không phản hồi, gộp nhiều khung dữ liệu vào một frame), với giới hạn kích thước khung 8192 byte.
  2. Các sự kiện gọi trên hub chat này không khả dụng (cuộc gọi dùng endpoint riêng); mỗi khung dữ liệu nhận được gồm loại sự kiện và phần nội dung.
  3. **Sự kiện từ ứng dụng người dùng**:
     - chat:join — tham gia phòng chat: kiểm tra người dùng là thành viên, đưa kết nối vào phòng, trả lịch sử tin nhắn (mặc định 30, tối đa 100, phân trang theo thời điểm và định danh tin nhắn) kèm danh sách tin ghim nếu có;
     - chat:history:more — tải thêm trang lịch sử cũ hơn theo con trỏ phân trang;
     - message:send — gửi tin nhắn với nội dung, phiên bản mã hóa, emoji, tệp đính kèm, gif, tin trả lời, bài viết được chia sẻ, nhóm media hoặc tin chuyển tiếp; lưu tin rồi phát tin mới cho cả phòng chat;
     - typing:start / typing:stop — báo đang gõ hoặc ngừng gõ; hệ thống khử trùng lặp rồi phát trạng thái kèm định danh người dùng;
     - message:delete — xóa tin nhắn; chế độ xóa cho tất cả phát sự kiện xóa cho cả phòng, chế độ xóa cho mình chỉ xác nhận về người gửi;
     - message:search — tìm kiếm tin nhắn trong chat rồi trả kết quả về đúng người dùng;
     - message:pin / message:unpin — ghim hoặc bỏ ghim tin nhắn rồi phát sự kiện cho phòng;
     - message:read — cập nhật mốc đã đọc rồi phát cho các thành viên;
     - message:react — thêm phản ứng trên tin nhắn rồi phát kết quả.
  4. **Sự kiện gửi đến ứng dụng người dùng**: lịch sử tin nhắn, lịch sử trang cũ, danh sách tin ghim, tin nhắn mới, trạng thái đang gõ, tin nhắn bị xóa, kết quả tìm kiếm, tin được ghim hoặc bỏ ghim, mốc đã đọc, phản ứng trên tin, thông báo lỗi kèm nội dung, và thông báo khóa mã hóa đầu cuối của chat đã được cập nhật (gửi tới từng kết nối của thành viên).
  5. **Mã hóa phiên bản cũ**: máy chủ mã hóa nội dung tin nhắn bằng AES-256-GCM với khóa lưu của chat (nonce 12 byte đặt trước dữ liệu mã hóa rồi mã hóa base64); khi phát tin mới, trả lịch sử hoặc bản xem trước, máy chủ giải mã rồi gửi nội dung gốc cho thành viên đã xác thực.
  6. **Mã hóa đầu cuối**: ứng dụng người dùng lấy khóa chat đối xứng riêng cho từng chat (bọc bằng bí mật chung ECDH từ khóa công khai hai bên), mã hóa nội dung AES-256-GCM ngay trên máy rồi gửi bản mã; máy chủ lưu nguyên trạng và phát nguyên bản mã — mỗi ứng dụng tự bóc khóa và tự giải mã. Máy chủ không đọc được nội dung; tìm kiếm bỏ qua tin mã hóa đầu cuối (chỉ tìm được trên máy người dùng); bản xem trước tin cuối và tin trả lời dạng mã hóa cũng giữ nguyên bản mã.
  7. Lịch sử tin được làm giàu kèm thông tin tệp đính kèm, thời lượng ghi âm, bài viết được chia sẻ, bản xem trước tin trả lời, danh sách thành viên đã xem theo mốc đã đọc và các phản ứng; lỗi nghiệp vụ được gửi về người dùng qua sự kiện lỗi.
- **Entity**: người gửi tin, thành viên chat (người nhận), Hub chat (chatHub)
- **Model**: models.Message, models.Chat, models.ChatParticipant, models.PinnedMessage, models.ChatRead, models.Media, models.Post
- **Databases**: chats, chat_participants, messages, media, pinned_messages, chat_reads, posts, chat_e2e_keys

---
### 103. Danh sách lời mời chat đã nhận

- **Name**: ListChatInvites
- **Description**:
  1. Người dùng mở danh sách lời mời chat đã nhận, hệ thống xác thực rồi lấy các lời mời đang chờ phản hồi (chưa được chấp nhận hoặc từ chối) của người dùng.
  2. Gom toàn bộ người gửi lời mời và tải thông tin hồ sơ của họ một lần để tránh tra cứu lặp; tên hiển thị trống thì dùng tên mặc định.
  3. Định dạng từng lời mời gồm định danh lời mời, người gửi, tên hiển thị, ảnh đại diện và thời điểm gửi.
  4. Trả về danh sách lời mời; không có lời mời thì trả danh sách rỗng; thao tác không làm phát sinh thay đổi dữ liệu nào khác.
- **Entity**: người nhận lời mời (caller), người gửi lời mời
- **Model**: models.ChatInvite, models.Profile
- **Databases**: chat_invitations, profiles

---

### 104. Tạo (hoặc lấy lại) chat trực tiếp

- **Name**: CreateDirectChat
- **Description**:
  1. Người dùng chọn người khác để bắt đầu chat trực tiếp, hệ thống xác thực rồi nhận định danh người đối diện (bắt buộc).
  2. Chặn trường hợp người dùng tự chat với chính mình.
  3. Kiểm tra cài đặt của đối phương: nếu cho phép nhận tin nhắn từ người lạ thì ai cũng mở được chat; ngược lại yêu cầu hai bên đã kết bạn.
  4. Nếu chat trực tiếp giữa hai người đã tồn tại thì trả về ngay chat đó cùng thông báo đã có chat giữa hai người.
  5. Chưa có → kiểm tra tình trạng bạn bè nếu bắt buộc, rồi tạo chat trực tiếp và thêm hai thành viên trong một giao dịch để đảm bảo nguyên tử.
  6. Nếu hai yêu cầu cùng tạo chat đồng thời gây trùng lặp thì truy vấn lại chat đã có và trả về (bảo đảm xử lý nhiều lần cho cùng một yêu cầu).
  7. Trả về định danh chat cùng thông báo; lỗi nghiệp vụ được báo lại cho người dùng.
- **Entity**: người mở chat, đối phương
- **Model**: models.Chat, models.ChatParticipant
- **Databases**: chats, chat_participants, user_settings, friends

---

### 105. Gửi lời mời chat (khi chưa kết bạn)

- **Name**: CreateChatInvite
- **Description**:
  1. Người dùng gửi lời mời chat tới người chưa kết bạn, hệ thống xác thực rồi nhận định danh người được mời.
  2. Chặn trường hợp mời chính mình.
  3. Ngăn trùng lặp theo nhiều lớp: hai người đã là bạn bè; đang có lời mời chờ phản hồi giữa hai người; lời mời cũ còn hiệu lực (đang chờ hoặc đã được chấp nhận).
  4. Tạo lời mời chat với trạng thái đang chờ phản hồi.
  5. Gửi thông báo cho người được mời biết họ được mời tham gia chat.
  6. Trả về định danh lời mời cùng thông báo; mọi lỗi nghiệp vụ được báo lại cho người dùng.
- **Entity**: người gửi lời mời, người được mời
- **Model**: models.ChatInvite
- **Databases**: chat_invitations, friends, notifications

---
### 106. Phản hồi lời mời chat (chấp nhận / từ chối)

- **Name**: ResponseChatInvite
- **Description**:
  1. Người dùng mở lời mời và chọn chấp nhận hoặc từ chối, hệ thống xác thực rồi đọc lời mời theo định danh kèm trạng thái đang chờ phản hồi — lời mời đã được xử lý hoặc không tồn tại sẽ báo lỗi.
  2. Kiểm tra người phản hồi chính là người nhận lời mời của lời mời đó — không ai phản hồi thay được.
  3. Từ chối → cập nhật trạng thái lời mời là đã từ chối, trả thông báo người dùng không chấp nhận chat, không kèm định danh chat.
  4. Chấp nhận → tạo hoặc lấy chat trực tiếp giữa hai người (không kiểm tra bạn bè vì việc chấp nhận lời mời đã ngầm đồng ý), cập nhật lời mời là đã chấp nhận kèm định danh chat.
  5. Gửi thông báo cho người gửi lời mời biết lời mời đã được chấp nhận.
  6. Trả về định danh lời mời, định danh chat và thông báo chat đã được kích hoạt; có lỗi thì báo lại.
- **Entity**: người nhận lời mời (caller), người gửi lời mời
- **Model**: models.ChatInvite, models.Chat, models.ChatParticipant
- **Databases**: chat_invitations, chats, chat_participants, notifications

---

### 107. Upload media vào chat (Cloudinary)

- **Name**: UploadChatMedia
- **Description**:
  1. Người dùng chọn tệp đính kèm để gửi lên chat, hệ thống xác thực rồi đọc định danh chat từ biểu mẫu (bắt buộc) và kiểm tra người gửi là thành viên chat — không phải thành viên bị từ chối.
  2. Nhận tệp tải lên kèm thời lượng (tùy chọn, chỉ dùng cho ghi âm).
  3. Kiểm tra hạn mức lưu trữ của người dùng; cho phép ảnh, video và âm thanh (âm thanh tối đa 10MB); chặn khi vượt quá dung lượng còn trống và báo cần nâng cấp dung lượng.
  4. Tải tệp lên Cloudinary với định danh mới — media chat được duyệt ngay, không xếp hàng kiểm duyệt bằng AI như bài viết.
  5. Nếu là tệp âm thanh có thời lượng thì lưu thời lượng và kiểm tra không quá 5 phút.
  6. Tạo bản ghi media với đường dẫn an toàn từ Cloudinary và cộng dồn dung lượng đã dùng của người dùng.
  7. Trả về thông tin tệp đã tải kèm thống kê dung lượng (tổng hạn mức, đã dùng, còn trống); người dùng dùng định danh tệp nhận được khi gửi tin nhắn đính kèm media.
- **Entity**: người upload, thành viên chat
- **Model**: models.Media
- **Databases**: media, users, chats, chat_participants

---

### 108. Tải về media của tin nhắn

- **Name**: DownloadMessageMedia
- **Description**:
  1. Người dùng chọn tệp trong tin nhắn để tải về, hệ thống xác thực rồi nhận định danh tin nhắn (thiếu → báo yêu cầu không hợp lệ).
  2. Tìm tin nhắn; không thấy → báo tin nhắn không tồn tại.
  3. Kiểm tra người tải là thành viên của chat chứa tin nhắn — không phải → bị từ chối.
  4. Tin nhắn phải có tệp đính kèm; lấy thông tin tệp để lấy đường dẫn tải.
  5. Server làm trung gian tải lại tệp từ Cloudinary rồi chuyển về cho client — người dùng không gọi thẳng dịch vụ lưu trữ.
  6. Trả về tệp với đúng loại nội dung và tên tệp đính kèm (phần mở rộng được suy ra từ loại MIME).
  7. Thao tác chỉ trả thông tin tệp và nội dung file tải về; nội dung tin nhắn mã hóa đầu cuối không bị đọc hay giải mã tại đây.
- **Entity**: người tải (participant), người gửi media
- **Model**: models.Message, models.Media
- **Databases**: messages, media, chat_participants

---
### 109. Xóa chat

- **Name**: DeleteChat
- **Description**:
  1. Người dùng chọn xóa một cuộc chat, hệ thống xác thực rồi nhận định danh chat.
  2. Kiểm tra chat tồn tại — không có → báo không tìm thấy chat.
  3. Kiểm tra người xóa là thành viên của chat — không phải → bị từ chối.
  4. Xóa cứng trong một giao dịch theo thứ tự an toàn: toàn bộ tin nhắn của chat, các thành viên, rồi chính chat.
  5. Trả thông báo xóa phòng chat thành công.
- **Entity**: thành viên chat (người xóa)
- **Model**: models.Chat, models.ChatParticipant, models.Message
- **Databases**: chats, chat_participants, messages

---

### 110. Chia sẻ bài viết vào chat trực tiếp

- **Name**: SharePost (chat)
- **Description**:
  1. Người dùng chia sẻ một bài viết tới người khác qua chat, hệ thống xác thực rồi nhận định danh người nhận và bài viết cần chia sẻ (thiếu → báo lỗi).
  2. Tìm hoặc tạo chat trực tiếp với người nhận theo quy tắc bạn bè hoặc cho phép tin nhắn từ người lạ như mục tạo chat trực tiếp.
  3. Gửi một tin nhắn dạng chia sẻ bài viết: nội dung tin để trống, không bật mã hóa đầu cuối, chỉ trỏ tới bài viết cần chia sẻ.
  4. Khi lưu tin: kiểm tra người gửi không bị chặn chat, chat tồn tại, người gửi là thành viên; vì là chia sẻ bài viết nên bỏ qua kiểm tra nội dung văn bản, emoji và media; kiểm tra bài viết phải tồn tại và không bị ẩn hay riêng tư — nếu không báo bài viết không tồn tại hoặc không khả dụng.
  5. Tạo tin nhắn dạng chia sẻ bài viết; vì không bật mã hóa đầu cuối nên nội dung đang trống được mã hóa AES-256-GCM bằng khóa lưu của chat rồi lưu.
  6. Gửi thông báo cho từng thành viên còn lại trong chat biết đã có tin nhắn mới, với gợi ý nội dung là đã gửi một bài viết.
  7. Trả về thông báo chia sẻ bài viết thành công kèm tin nhắn vừa tạo. Lưu ý: thao tác này chỉ lưu tin nhắn, không phát sự kiện tin mới lên hub chat — người cùng chat nhận tin khi tải lịch sử chat hoặc khi gửi chia sẻ qua WebSocket.
- **Entity**: người chia sẻ, đối phương trong chat, tác giả bài viết
- **Model**: models.Message, models.Post, models.Chat, models.Notification
- **Databases**: messages, posts, chats, chat_participants, notifications

---
### 111. Lấy background của chat

- **Name**: GetChatBackground (chat)
- **Description**:
  1. Người dùng mở cài đặt giao diện chat, hệ thống xác thực rồi lấy định danh chat và người dùng.
  2. Đọc cài đặt nền của riêng người dùng trong chat đó (mỗi người tự cài cho mình, không ảnh hưởng đối phương).
  3. Chưa từng cài đặt → trả về loại nền và giá trị trống để ứng dụng dùng nền mặc định (không coi là lỗi).
  4. Có cài đặt → trả về loại nền, giá trị và độ mạnh phủ; lỗi dữ liệu → báo lỗi hệ thống.
- **Entity**: thành viên chat (tự cài background của chính mình)
- **Model**: models.ChatUserSettings
- **Databases**: chat_user_settings

---

### 112. Đặt / cập nhật background của chat

- **Name**: UpdateChatBackground (chat)
- **Description**:
  1. Người dùng chọn nền cho chat (màu đặc, gradient, mẫu có sẵn hoặc ảnh tùy chỉnh), hệ thống xác thực rồi nhận loại nền, giá trị và độ mạnh phủ.
  2. Không gửi độ mạnh phủ thì mặc định 100 (dải hợp lệ từ 30 đến 150).
  3. Kiểm tra loại nền chỉ nhận một trong bốn giá trị: màu đặc, gradient, mẫu có sẵn, tùy chỉnh — sai → báo lỗi.
  4. Kiểm tra độ mạnh phủ nằm trong khoảng 30 đến 150, sai → báo lỗi.
  5. Lưu cài đặt theo cặp chat và người dùng: chưa có thì thêm mới, đã có thì cập nhật loại nền, giá trị nền, độ mạnh phủ và thời điểm cập nhật.
  6. Trả về loại nền, giá trị và độ mạnh phủ vừa lưu; nền này cũng hiển thị trong danh sách chat.
- **Entity**: thành viên chat (tự cài background cho chính mình)
- **Model**: models.ChatUserSettings
- **Databases**: chat_user_settings

---

### 113. Xóa background của chat

- **Name**: DeleteChatBackground (chat)
- **Description**:
  1. Người dùng khôi phục nền mặc định cho chat, hệ thống xác thực rồi nhận định danh chat và người dùng.
  2. Chỉ xóa nền của chính người dùng trong chat đó, không đụng tới cài đặt của đối phương.
  3. Dữ liệu không tồn tại cũng trả thành công (xử lý nhiều lần không sinh lỗi).
  4. Trả thông báo đã xóa nền; lỗi dữ liệu → báo lỗi hệ thống; sau đó khi lấy nền sẽ nhận giá trị trống (mặc định).
- **Entity**: thành viên chat (xóa background của mình)
- **Model**: models.ChatUserSettings
- **Databases**: chat_user_settings

---

### 114. Nội dung đã chia sẻ trong chat (Shared content sidebar)

- **Name**: GetSharedContent (chat)
- **Description**:
  1. Người dùng mở danh sách nội dung đã chia sẻ trong chat, hệ thống xác thực rồi nhận bộ lọc tab (mặc định tất cả) và phân trang (mặc định từ đầu, 30 phần tử mỗi trang).
  2. Kiểm tra người xem là thành viên chat — không phải → bị từ chối.
  3. Lọc theo tab: media (ảnh và video), tệp (âm thanh và các tệp khác), liên kết (trích xuất đường dẫn đầu tiên trong nội dung tin), bài viết (tin chia sẻ bài viết); tab tất cả trả cả bốn danh sách, mỗi danh sách giới hạn số phần tử.
  4. Mọi truy vấn đều loại tin đã xóa, bao gồm tin đã xóa theo phía của người xem, sắp xếp theo thời gian giảm dần và phân trang theo số phần tử đã lấy.
  5. Trả về danh sách tương ứng với tab đã chọn (chỉ danh sách của tab được chọn được điền dữ liệu).
- **Entity**: thành viên chat (người xem sidebar)
- **Model**: models.Message, models.Media, models.Post
- **Databases**: messages, media, posts

---
### 115. Đăng ký public key E2E của user

- **Name**: RegisterUserKey
- **Description**:
  1. Người dùng bật mã hóa đầu cuối, hệ thống xác thực rồi lấy khóa công khai từ yêu cầu — người dùng chỉ đăng ký khóa của chính mình, khóa riêng tư không bao giờ rời khỏi thiết bị.
  2. Chưa từng đăng ký → lưu khóa công khai với phiên bản khóa là 1.
  3. Đã đăng ký mà khóa không đổi → chỉ cập nhật thời điểm cập nhật, giữ nguyên phiên bản.
  4. Khóa công khai thay đổi (danh tính hoặc thiết bị mới) → cập nhật khóa và tăng phiên bản lên một đơn vị, để đối phương phát hiện sự thay đổi rồi tạo lại khóa.
  5. Trả thông báo khóa đã được đăng ký; lỗi dữ liệu → báo lỗi hệ thống.
- **Entity**: user đang đăng ký khóa (caller)
- **Model**: models.UserE2EKey
- **Databases**: user_e2e_keys

---

### 116. Lấy public key E2E của user khác

- **Name**: GetUserKey
- **Description**:
  1. Người dùng cần khóa công khai mã hóa đầu cuối của đối phương, hệ thống xác thực rồi đọc định danh người cần lấy khóa.
  2. Chưa từng đăng ký khóa → báo không tìm thấy người dùng.
  3. Trả về định danh người dùng, khóa công khai và phiên bản khóa.
  4. Ứng dụng dùng khóa công khai của đối phương kèm khóa riêng tư của mình để tính bí mật chung ECDH, từ đó bóc và mở khóa chat đối xứng cho từng thành viên; so sánh phiên bản khóa với bản cũ để phát hiện đối phương đổi danh tính rồi tạo lại khóa.
- **Entity**: caller (người cần khóa), user sở hữu khóa
- **Model**: models.UserE2EKey
- **Databases**: user_e2e_keys

---

### 117. Lưu khóa chat E2E (wrapped key, batch)

- **Name**: StoreChatKeys
- **Description**:
  1. Người dùng hoặc thiết bị gửi lên danh sách khóa chat đã bọc cho nhiều chat, hệ thống xác thực rồi nhận danh sách gồm định danh chat, người nhận khóa, khóa đã bọc và nonce — validate từng phần tử của danh sách.
  2. Với từng khóa, kiểm tra người gửi là thành viên của chat đó (khóa trong danh sách có thể dành cho thành viên khác, người gửi bọc khóa cho đối phương); vi phạm → báo lỗi.
  3. Luồng mã hóa: ứng dụng tạo khóa chat đối xứng AES-256 riêng cho từng chat, mã hóa nội dung tin bằng AES-256-GCM; bọc khóa đó cho từng người bằng bí mật chung ECDH (khóa công khai lấy từ mục lấy khóa đối phương) để thu khóa đã bọc và nonce; máy chủ chỉ lưu dữ liệu opaque, không thể mở khóa.
  4. Lưu khóa trong một giao dịch: khóa các dòng chat liên quan (sắp xếp định danh chat để tránh deadlock) rồi ghi đè khóa chat theo cặp chat và người dùng theo nguyên tắc ai ghi trước thắng — hai thiết bị khởi tạo khóa đồng thời được tuần tự hóa về cùng một khóa cho cả hai thành viên.
  5. Sau khi lưu, gom người nhận thông báo (mọi người trong danh sách trừ người gửi, khử trùng lặp theo chat) và gửi sự kiện khóa mã hóa đầu cuối của chat đã được cập nhật tới từng kết nối của họ; ứng dụng nhận sự kiện thì tự lấy khóa mới và giải mã lại tin đang kẹt, không cần tải lại trang; nếu hub chat chưa được kết nối thì bỏ qua bước báo này.
  6. Trả thông báo khóa chat đã được lưu.
- **Entity**: caller (người bọc khóa), participant khác trong chat, chatHub (người nhận notify)
- **Model**: models.ChatE2EKey
- **Databases**: chat_e2e_keys, chats, chat_participants

---
### 118. Lấy khóa chat E2E của chính mình

- **Name**: GetChatKey
- **Description**:
  1. Người dùng tải khóa chat mã hóa đầu cuối của chính mình, hệ thống xác thực rồi nhận định danh chat.
  2. Kiểm tra người dùng là thành viên chat — không phải → bị từ chối (không thể đọc khóa của chat mình không thuộc).
  3. Đọc khóa trong giao dịch có khóa hàng để luôn nhận snapshot nhất quán, không đọc được nửa cặp khi bên kia đang thiết lập khóa.
  4. Chưa có khóa (chat chưa bật mã hóa đầu cuối) → báo không tìm thấy khóa chat.
  5. Trả về định danh chat, khóa đã bọc và nonce; ứng dụng tự bóc khóa bằng bí mật chung ECDH để lấy khóa chat AES-256 rồi giải mã các tin mã hóa đầu cuối nhận qua lịch sử hoặc tin mới.
- **Entity**: caller (participant), participant cùng chat
- **Model**: models.ChatE2EKey
- **Databases**: chat_e2e_keys, chats, chat_participants

---

### 119. Re-key khóa chat (sau khi đối phương đổi identity)

- **Name**: RekeyChat
- **Description**:
  1. Người dùng hoặc thiết bị cập nhật khóa bọc sau khi đối phương đổi danh tính, hệ thống xác thực rồi nhận định danh chat kèm khóa đã bọc và nonce.
  2. Chỉ thành viên của chat mới được tạo lại khóa cho chat đó.
  3. Luồng mã hóa: ứng dụng phát hiện đối phương đổi khóa công khai (qua phiên bản khóa) → tính lại bí mật chung ECDH mới → bọc lại khóa chat đối xứng vẫn giữ nguyên bằng bí mật mới rồi gửi lên.
  4. Ghi đè khóa theo cặp chat và người dùng của chính người gọi (nguyên tắc ghi đè, khác với ghi trước thắng ở mục lưu khóa theo lô); không tạo khóa chat mới, không đụng khóa của người khác.
  5. Trả thông báo khóa chat đã được cập nhật; không phải thành viên hoặc lỗi dữ liệu → báo lỗi.
  6. Vì khóa chuẩn không đổi nên lịch sử tin cũ vẫn giải mã được — chỉ cần bóc khóa lại trên thiết bị mới hoặc phía đối phương.
- **Entity**: caller (participant re-key), đối phương (đã đổi identity)
- **Model**: models.ChatE2EKey
- **Databases**: chat_e2e_keys, chats, chat_participants

---

### 120. Lưu backup khôi phục khóa chat

- **Name**: PutRecovery
- **Description**:
  1. Người dùng bật sao lưu khôi phục khóa, hệ thống xác thực rồi nhận muối, dữ liệu sao lưu đã mã hóa, giá trị xác minh PIN và giá trị xác minh khóa khôi phục — cả bốn là bắt buộc.
  2. Luồng mã hóa: ứng dụng gom toàn bộ khóa chat đã bọc, mã hóa thành dữ liệu sao lưu AES-256-GCM bằng khóa dẫn từ mật khẩu khôi phục (dùng muối); đồng thời băm giá trị xác minh PIN và giá trị xác minh khóa khôi phục; máy chủ chỉ nhận dữ liệu opaque — không đọc được dữ liệu sao lưu, không biết PIN.
  3. Khởi tạo bộ đếm thử sai bằng 0 và không đặt thời gian khóa — mỗi lần bật hoặc đổi PIN sẽ reset bộ đếm.
  4. Ghi đè bản ghi theo người dùng: muối, dữ liệu sao lưu, hai giá trị xác minh, bộ đếm thử sai, thời gian khóa và thời điểm cập nhật.
  5. Trả thông báo đã lưu dữ liệu khôi phục; lỗi → báo lỗi hệ thống.
- **Entity**: user (chủ khóa), server (chỉ giữ blob)
- **Model**: models.UserE2ERecovery
- **Databases**: user_e2e_recovery

---
### 121. Trạng thái backup khôi phục (metadata)

- **Name**: GetRecoveryMeta
- **Description**:
  1. Người dùng kiểm tra trạng thái sao lưu khôi phục khóa, hệ thống xác thực rồi lấy người dùng hiện hành.
  2. Chưa từng bật khôi phục → trả trạng thái chưa có dữ liệu sao lưu (không coi là lỗi).
  3. Đã có → trả trạng thái đã có dữ liệu, kèm muối và thời điểm cập nhật.
  4. Không trả về dữ liệu sao lưu, hai giá trị xác minh, bộ đếm thử sai hay thời gian khóa (thông tin nhạy cảm).
  5. Thao tác này là cờ để giao diện quyết định có hiển thị màn hình bật hoặc khôi phục khóa hay không.
- **Entity**: user (chủ backup)
- **Model**: models.UserE2ERecovery
- **Databases**: user_e2e_recovery

---

### 122. Mở khóa backup khôi phục (PIN / recovery key)

- **Name**: UnlockRecovery
- **Description**:
  1. Người dùng nhập PIN hoặc khóa khôi phục, hệ thống xác thực rồi nhận giá trị đã băm — ứng dụng gửi giá trị băm, không gửi PIN thô.
  2. Chưa từng bật sao lưu → bị từ chối, báo chưa có dữ liệu khôi phục.
  3. Đang trong thời gian khóa tạm do nhập sai nhiều lần → bị từ chối, báo thử lại sau ít phút.
  4. Đối chiếu giá trị gửi lên với cả giá trị xác minh PIN lẫn giá trị xác minh khóa khôi phục — khớp cái nào cũng mở được.
  5. Sai: tăng bộ đếm thử sai; khi đủ 5 lần sai thì khóa tạm tăng dần theo số lần (10 phút, 20 phút, 30 phút…); trả báo mã xác minh không đúng.
  6. Đúng: reset bộ đếm thử sai, xóa thời gian khóa tạm, cập nhật thời điểm rồi trả dữ liệu sao lưu; ứng dụng tự giải dữ liệu sao lưu (AES-256-GCM, khóa dẫn từ mật khẩu khôi phục và muối lấy ở mục trạng thái) để lấy lại toàn bộ khóa chat, từ đó đọc lại lịch sử tin mã hóa đầu cuối trên thiết bị mới.
- **Entity**: user (người mở khóa), server (đối chiếu hash, trả blob)
- **Model**: models.UserE2ERecovery
- **Databases**: user_e2e_recovery

---

### 123. Xóa backup khôi phục

- **Name**: DeleteRecovery
- **Description**:
  1. Người dùng tắt sao lưu khôi phục khóa, hệ thống xác thực rồi chỉ xóa dữ liệu của chính người dùng (không nhận định danh từ yêu cầu).
  2. Xóa bản ghi sao lưu khôi phục của người dùng.
  3. Bản ghi không tồn tại cũng trả thành công (xử lý nhiều lần không sinh lỗi); lỗi dữ liệu → báo lỗi hệ thống.
  4. Trả thông báo đã xóa dữ liệu khôi phục; sau khi xóa, kiểm tra trạng thái sao lưu trả về chưa có dữ liệu và người dùng không còn lối khôi phục khóa chat trên thiết bị mới bằng mật khẩu khôi phục.
- **Entity**: user (chủ backup, người tắt tính năng)
- **Model**: models.UserE2ERecovery
- **Databases**: user_e2e_recovery

---

**Tổng số mục**: 23 (101—123) — Chat HTTP (101, 103—114), Chat WebSocket (102), E2E encryption (115—123).

## B7 — Group Chat HTTP + Group Chat WS + Group Call WS

Tài liệu hóa các endpoint từ **124–143**: Group Chat HTTP API (`routes/group_chat.routes.go`), Group Chat WS (`GET /api/group-chats/ws`) và Group Call WS (`GET /api/group-calls/ws`).

---

### 124. Danh sách nhóm chat của tôi

- **Name**: ListGroups
- **Description**:
  1. Người dùng mở danh sách nhóm chat của mình; hệ thống lấy danh tính người gọi từ phiên đăng nhập, nếu phiên không hợp lệ thì báo chưa đăng nhập.
  2. Tra cứu các cuộc trò chuyện dạng nhóm mà người dùng là thành viên, kèm tin nhắn gần nhất và thông tin nhóm.
  3. Với mỗi cuộc trò chuyện có tin nhắn gần nhất, giải mã nội dung tin nhắn bằng khóa mã hóa của cuộc trò chuyện để hiển thị dạng văn bản thuần.
  4. Trả về danh sách nhóm chat của người dùng.
- **Entity**: user (người gọi), group chat (chat type=group)
- **Model**: models.Chat (type `group`), models.Message (last message), dto.GroupChatConversationDTO, dto.GroupChatListResponse
- **Databases**: chats, chat_participants, messages, profiles, users

---

### 125. Tạo nhóm chat mới

- **Name**: CreateGroup
- **Description**:
  1. Người dùng nhập tên nhóm, chọn ảnh đại diện và các thành viên muốn mời; hệ thống nhận thông tin từ phiên đăng nhập và yêu cầu gửi lên.
  2. Kiểm tra tên nhóm phải có từ 3 đến 50 ký tự.
  3. Tạo cuộc trò chuyện nhóm mới với định danh duy nhất và sinh khóa mã hóa riêng cho nhóm.
  4. Đặt người tạo làm quản trị viên, các thành viên được mời làm thành viên thường (bỏ qua chính người tạo và các định danh rỗng).
  5. Lưu cuộc trò chuyện và danh sách thành viên vào database theo một giao dịch.
  6. Gửi thông báo "đã thêm bạn vào nhóm ..." cho các thành viên có bật thông báo nhóm.
  7. Trả về thông báo tạo nhóm thành công kèm định danh nhóm.
- **Entity**: user (creator/admin), các member được mời
- **Model**: models.Chat, models.ChatParticipant, models.Notification
- **Databases**: chats, chat_participants, group_chat_member_settings, notifications, profiles, users

---

### 126. Thêm thành viên (tạo lời mời)

- **Name**: AddMember
- **Description**:
  1. Người dùng chọn một thành viên khác và gửi lời mời vào nhóm; hệ thống lấy danh tính người gọi và nhóm được chọn.
  2. Không cho phép tự mời chính mình; kiểm tra người gọi phải là thành viên của nhóm.
  3. Nếu nhóm tắt tính năng cho phép thành viên tự thêm người thì chỉ quản trị viên được mời, ngược lại thành viên thường cũng được phép.
  4. Kiểm tra người được mời không bị cấm trong nhóm, chưa phải thành viên và chưa có lời mời đang chờ xử lý.
  5. Tạo lời mời vào nhóm ở trạng thái chờ; người được mời phải tự chấp nhận hoặc từ chối ở các thao tác riêng.
  6. Gửi thông báo lời mời, tin nhắn hệ thống báo có người được mời trong nhóm và tin nhắn lời mời trong cuộc trò chuyện trực tiếp giữa người mời và người được mời.
  7. Báo cho cả nhóm qua kênh thời gian thực rằng có người được thêm vào; trả về thông báo thành công kèm định danh lời mời.
- **Entity**: requester (member/admin), target user (người được mời)
- **Model**: models.GroupChatMemberRequest, models.Message (system + group_invite), models.Notification
- **Databases**: group_chat_member_requests, group_chat_settings, group_chat_bans, group_chat_member_settings, chats, chat_participants, messages, notifications, profiles, users

---

### 127. Chấp nhận lời mời vào nhóm

- **Name**: ApproveMemberRequest
- **Description**:
  1. Người dùng mở lời mời vào nhóm và chọn chấp nhận; hệ thống lấy danh tính người gọi, nhóm và lời mời được chọn.
  2. Tìm lời mời và xác minh nó thuộc đúng nhóm, đúng người được mời và còn đang chờ xử lý.
  3. Kiểm tra người dùng chưa bị cấm trong nhóm và chưa phải thành viên.
  4. Chuyển lời mời sang trạng thái đã chấp nhận và thêm người dùng vào danh sách thành viên của nhóm.
  5. Cập nhật tin nhắn lời mời trong cuộc trò chuyện trực tiếp của hai bên thành trạng thái đã chấp nhận.
  6. Thông báo cho người mời rằng người được mời đã chấp nhận tham gia nhóm; tạo tin nhắn hệ thống báo có thành viên mới trong nhóm.
  7. Báo cho cả nhóm qua kênh thời gian thực; trả về thông báo thành công.
- **Entity**: user được mời (target), requester, các thành viên nhóm
- **Model**: models.GroupChatMemberRequest, models.ChatParticipant, models.Message, models.Notification
- **Databases**: group_chat_member_requests, group_chat_bans, chat_participants, chats, messages, group_chat_member_settings, notifications, profiles, users

---

### 128. Từ chối lời mời vào nhóm

- **Name**: RejectMemberRequest
- **Description**:
  1. Người dùng mở lời mời vào nhóm và chọn từ chối; hệ thống lấy danh tính người gọi, nhóm và lời mời được chọn.
  2. Tìm lời mời và xác minh nó thuộc đúng nhóm, đúng người được mời và còn đang chờ xử lý.
  3. Đánh dấu lời mời ở trạng thái đã từ chối kèm thời điểm phản hồi.
  4. Cập nhật tin nhắn lời mời trong cuộc trò chuyện trực tiếp của hai bên thành trạng thái đã từ chối.
  5. Thông báo cho người mời rằng lời mời tham gia nhóm đã bị từ chối; trả về thông báo thành công mà không thêm vào nhóm và không phát sóng cho cả nhóm.
- **Entity**: user được mời (target), requester
- **Model**: models.GroupChatMemberRequest, models.Message, models.Notification
- **Databases**: group_chat_member_requests, messages, group_chat_member_settings, notifications, profiles, users

---

### 129. Trục xuất & chặn thành viên (ban)

- **Name**: BanMember
- **Description**:
  1. Quản trị viên chọn một thành viên và thực hiện chặn khỏi nhóm; hệ thống lấy danh tính người gọi và nhóm được chọn.
  2. Không cho phép tự chặn chính mình; kiểm tra người gọi phải là quản trị viên của nhóm.
  3. Kiểm tra người bị chọn chưa từng bị chặn trong nhóm; tạo bản ghi chặn kèm người thực hiện và thời điểm.
  4. Thông báo cho người bị chặn rằng họ đã bị chặn khỏi nhóm chat (nếu người đó bật thông báo nhóm).
  5. Trả về thông báo thành công; từ khi bị chặn, mọi thao tác vào nhóm của người này như kết nối thời gian thực, gửi tin nhắn, thêm thành viên... đều bị từ chối.
- **Entity**: admin/owner (người ban), target user (bị ban)
- **Model**: models.GroupChatBan, models.Notification
- **Databases**: group_chat_bans, chat_participants, group_chat_member_settings, notifications, profiles, users

---

### 130. Gửi tin nhắn vào nhóm (HTTP)

- **Name**: SendGroupMessage
- **Description**:
  1. Người dùng nhập nội dung tin nhắn hoặc chọn ảnh, gif, bài viết để gửi vào nhóm; hệ thống nhận thông tin gửi kèm biểu tượng cảm xúc, tin trả lời và tin chuyển tiếp.
  2. Kiểm tra người dùng có đang bị tắt tiếng trong nhóm hay không; nếu đang bị tắt thì từ chối kèm lý do và hạn còn lại.
  3. Kiểm tra cuộc trò chuyện tồn tại và người gửi là thành viên của nhóm.
  4. Nếu gửi ảnh động thì tạo bản ghi phương tiện dạng ảnh; nếu chia sẻ bài viết thì kiểm tra bài viết còn khả dụng; ngược lại kiểm tra nội dung, biểu tượng cảm xúc và phương tiện.
  5. Kiểm tra biểu tượng cảm xúc hợp lệ, phương tiện thuộc về người gửi, tin trả lời thuộc đúng nhóm và tin chuyển tiếp hợp lệ.
  6. Mã hóa nội dung tin nhắn bằng khóa mã hóa riêng của nhóm rồi lưu tin nhắn, kèm thông tin trả lời, phương tiện, bài viết được chia sẻ và tin chuyển tiếp.
  7. Gửi thông báo cho các thành viên còn lại rằng có tin nhắn mới; trả về thông báo thành công kèm nội dung tin nhắn vừa gửi.
- **Entity**: sender (member, không bị mute), các thành viên nhóm (nhận notify)
- **Model**: models.Message, models.Media, models.Notification
- **Databases**: messages, media, chats, chat_participants, group_chat_mutes, group_chat_member_settings, notifications, posts

---

### 131. Chuyển quyền admin

- **Name**: TransferAdmin
- **Description**:
  1. Quản trị viên chọn một thành viên để trao quyền quản trị; hệ thống lấy danh tính người gọi và nhóm được chọn.
  2. Không cho phép tự chuyển quyền cho chính mình; kiểm tra người gọi phải là quản trị viên và người được chọn phải là thành viên của nhóm.
  3. Kiểm tra thời gian chờ giữa hai lần chuyển quyền quản trị phải cách nhau ít nhất một tháng; nếu chưa đủ thì báo lỗi kèm thời điểm lần chuyển trước.
  4. Nâng quyền người được chọn lên quản trị viên và ghi nhận thời điểm chuyển quyền.
  5. Tạo tin nhắn hệ thống báo đã chuyển quyền quản trị; thông báo cho người được nhận quyền rằng họ đã được chuyển quyền quản trị nhóm.
  6. Báo cho cả nhóm qua kênh thời gian thực kèm tên người vừa được làm quản trị viên; trả về thông báo thành công.
- **Entity**: admin hiện tại (requester), target user (được làm admin)
- **Model**: models.ChatParticipant, models.GroupChatSettings, models.Message, models.Notification
- **Databases**: chat_participants, group_chat_settings, messages, group_chat_member_settings, notifications, profiles, users

---

### 132. Chuyển quyền sở hữu nhóm (owner)

- **Name**: TransferOwnership (group)
- **Description**:
  1. Chủ nhóm chọn một thành viên để chuyển quyền sở hữu; hệ thống lấy danh tính người gọi và nhóm được chọn.
  2. Không cho phép tự chuyển cho chính mình; chỉ chủ sở hữu nhóm mới được chuyển quyền và nhóm phải là nhóm chat.
  3. Người được nhận quyền phải đang là thành viên của nhóm.
  4. Đổi chủ sở hữu nhóm sang người được chọn, cập nhật vai trò của chủ cũ về thành viên thường hoặc giữ lại quyền quản trị tùy lựa chọn, đồng thời ghi nhận thời điểm chuyển quyền.
  5. Thông báo cho người được nhận quyền rằng họ đã được chuyển quyền sở hữu nhóm; trả về thông báo thành công.
- **Entity**: owner (creator), target user (trở thành owner mới)
- **Model**: models.Chat (creator_id), models.ChatParticipant, models.Notification
- **Databases**: chats, chat_participants, group_chat_settings, group_chat_member_settings, notifications, profiles, users

---

### 133. Xem cấu hình nhóm

- **Name**: GetSettings
- **Description**:
  1. Người dùng mở màn hình cấu hình nhóm; hệ thống lấy danh tính người gọi và nhóm được chọn.
  2. Kiểm tra người gọi phải là thành viên của nhóm, nếu không thì báo lỗi.
  3. Lấy thông tin nhóm gồm tên và ảnh đại diện, cấu hình chung của nhóm gồm tính năng cho phép thành viên tự thêm người, và cấu hình cá nhân của người gọi gồm việc bật tắt thông báo.
  4. Lấy danh sách thành viên kèm hồ sơ cá nhân của từng người.
  5. Trả về thông tin cấu hình nhóm.
- **Entity**: member (người gọi), group chat
- **Model**: models.Chat, models.GroupChatSettings, models.GroupChatMemberSettings, models.ChatParticipant (join profiles)
- **Databases**: chats, group_chat_settings, group_chat_member_settings, chat_participants, profiles, users

---

### 134. Cập nhật cấu hình nhóm

- **Name**: UpdateSettings
- **Description**:
  1. Người dùng chỉnh sửa cấu hình nhóm; hệ thống nhận các thông tin: bật tắt thông báo cá nhân, cho phép thành viên tự thêm người, đổi tên và đổi ảnh đại diện nhóm.
  2. Kiểm tra người gọi là thành viên của nhóm; người dùng thường chỉ được tự đổi thông báo cá nhân, còn đổi tên (3 đến 50 ký tự), ảnh đại diện và tính năng thêm thành viên chỉ quản trị viên mới được phép.
  3. Nếu đổi tính năng cho phép thêm thành viên thì lưu vào cấu hình chung của nhóm; nếu đổi tên hoặc ảnh đại diện thì cập nhật thông tin nhóm và tạo tin nhắn hệ thống báo đã đổi tên hoặc đổi ảnh đại diện.
  4. Khi có đổi tên hoặc đổi ảnh đại diện, báo cho cả nhóm qua kênh thời gian thực kèm loại thay đổi và nội dung thông báo thay đổi.
  5. Trả về cấu hình mới của nhóm kèm thông báo cập nhật thành công.
- **Entity**: member (settings cá nhân), admin/owner (settings chung), các thành viên nhóm (nhận broadcast)
- **Model**: models.GroupChatSettings, models.GroupChatMemberSettings, models.Chat, models.Message
- **Databases**: group_chat_settings, group_chat_member_settings, chats, messages, chat_participants, profiles, users

---

### 135. Tắt tiếng thành viên (mute, có thời hạn)

- **Name**: MuteMember
- **Description**:
  1. Quản trị viên chọn một thành viên và tắt tiếng kèm lý do và thời hạn; hệ thống lấy danh tính người gọi và nhóm được chọn.
  2. Không cho phép tự tắt tiếng chính mình; kiểm tra người gọi phải là quản trị viên và người bị chọn phải là thành viên của nhóm.
  3. Kiểm tra lý do và thời lượng tắt tiếng phải hợp lệ.
  4. Tính thời điểm hết hạn bằng thời điểm hiện tại cộng thời lượng; nếu không chỉ định thời hạn thì tắt tiếng vĩnh viễn và không có ngày hết hạn; tạo bản ghi tắt tiếng.
  5. Thông báo cho người bị tắt tiếng rằng họ đã bị tắt tiếng trong nhóm kèm lý do và thời điểm hết hạn.
  6. Trả về thông báo thành công kèm bản ghi tắt tiếng; người bị tắt tiếng sẽ không gửi được tin nhắn qua cả đường truyền thường và thời gian thực.
- **Entity**: admin (người mute), target member (bị mute)
- **Model**: models.GroupChatMute, models.Notification
- **Databases**: group_chat_mutes, chat_participants, group_chat_member_settings, notifications, profiles, users

---

### 136. Bỏ tắt tiếng thành viên (unmute)

- **Name**: UnmuteMember
- **Description**:
  1. Quản trị viên chọn một thành viên đang bị tắt tiếng để bỏ tắt tiếng; hệ thống lấy danh tính người gọi và nhóm được chọn.
  2. Kiểm tra người gọi phải là quản trị viên của nhóm.
  3. Xóa bản ghi tắt tiếng của thành viên đó trong nhóm.
  4. Thông báo cho người vừa được mở lại quyền gửi tin nhắn trong nhóm (nếu người đó bật thông báo); trả về thông báo thành công.
- **Entity**: admin, target member
- **Model**: models.GroupChatMute, models.Notification
- **Databases**: group_chat_mutes, chat_participants, group_chat_member_settings, notifications, profiles, users

---

### 137. Rời khỏi nhóm

- **Name**: LeaveGroup
- **Description**:
  1. Người dùng chọn rời khỏi nhóm; hệ thống nhận cách rời nhóm là công khai hoặc im lặng và cách xử lý lịch sử tin nhắn là giữ lại hoặc ẩn danh hóa, mặc định là công khai và giữ lại.
  2. Kiểm tra người dùng vẫn đang là thành viên của nhóm và hai cách chọn phải hợp lệ.
  3. Nếu chọn ẩn danh hóa lịch sử thì xóa danh tính trên toàn bộ tin nhắn do người dùng đã gửi trong nhóm.
  4. Nếu chọn công khai thì tạo tin nhắn hệ thống báo có thành viên đã rời nhóm; nếu chọn im lặng thì không có tin nhắn công khai.
  5. Xóa người dùng khỏi danh sách thành viên của nhóm; gửi thông báo cho các thành viên còn lại, chọn im lặng thì chỉ thông báo cho quản trị viên, chọn công khai thì thông báo cho tất cả những người có bật thông báo nhóm.
  6. Nếu chọn công khai thì báo cho cả nhóm qua kênh thời gian thực; trả về thông báo thành công.
- **Entity**: member rời nhóm, admin (nhận notify khi silent), các thành viên còn lại
- **Model**: models.ChatParticipant, models.Message (system), models.Message (anonymize), models.Notification
- **Databases**: chat_participants, messages, group_chat_member_settings, notifications, profiles, users

---

### 138. Xem nền (background) chat nhóm

- **Name**: GetChatBackground (group)
- **Description**:
  1. Người dùng mở màn hình chat nhóm và muốn xem nền đang đặt.
  2. Hệ thống xác thực phiên đăng nhập để lấy danh tính người gọi và xác định nhóm được chọn.
  3. Tra cứu thiết lập nền cá nhân của người dùng cho nhóm này; nếu chưa từng đặt thì trả về loại nền và giá trị rỗng.
  4. Trả về loại nền, giá trị nền và độ mạnh phủ nền của người dùng.
- **Entity**: user (thành viên nhóm), group chat
- **Model**: models.ChatUserSettings, dto.ChatBackgroundResponse
- **Databases**: chat_user_settings, chats, chat_participants

---

### 139. Đặt/cập nhật nền chat nhóm

- **Name**: UpdateChatBackground (group)
- **Description**:
  1. Người dùng chọn nền cho chat nhóm; hệ thống nhận loại nền, giá trị nền và độ mạnh phủ nền, nếu không chỉ định độ mạnh thì mặc định là 100.
  2. Kiểm tra loại nền phải hợp lệ như màu hoặc ảnh; kiểm tra độ mạnh phủ nền nằm trong khoảng từ 30 đến 150.
  3. Lưu hoặc cập nhật thiết lập nền cá nhân của người dùng cho nhóm này; nền là thiết lập riêng nên không ảnh hưởng đến thành viên khác.
  4. Trả về loại nền, giá trị nền và độ mạnh phủ nền vừa lưu.
- **Entity**: user, group chat
- **Model**: models.ChatUserSettings, dto.ChatBackgroundResponse
- **Databases**: chat_user_settings, chats

---

### 140. Xóa nền chat nhóm

- **Name**: DeleteChatBackground (group)
- **Description**:
  1. Người dùng chọn xóa nền đang đặt cho chat nhóm.
  2. Hệ thống xác thực phiên đăng nhập để lấy danh tính người gọi và xác định nhóm được chọn.
  3. Xóa thiết lập nền cá nhân của người dùng cho nhóm này, đưa chat về nền mặc định.
  4. Trả về thông báo đã xóa nền; nếu xóa thất bại thì báo lỗi hệ thống.
- **Entity**: user, group chat
- **Model**: models.ChatUserSettings
- **Databases**: chat_user_settings

---

### 141. Nội dung đã chia sẻ trong nhóm (sidebar)

- **Name**: GetSharedContent (group)
- **Description**:
  1. Người dùng mở mục nội dung đã chia sẻ trong sidebar của nhóm; hệ thống nhận bộ lọc theo thẻ, phân trang và số lượng tối đa mỗi loại, mặc định là tất cả và 30 phần tử.
  2. Kiểm tra người gọi phải là thành viên của nhóm, nếu không thì báo lỗi.
  3. Lọc theo thẻ đã chọn: thẻ phương tiện, tệp, liên kết hoặc bài viết thì lấy loại tương ứng; thẻ tất cả thì lấy cả bốn loại, mỗi loại giới hạn số phần tử.
  4. Trả về danh sách nội dung đã chia sẻ gồm phương tiện, tệp, liên kết và bài viết.
- **Entity**: member (người xem), group chat, nội dung chia sẻ (media/file/link/post trong messages)
- **Model**: models.Message (media_id, shared_post_id, content rỗng link), models.Media, models.Post, dto.SharedContentResponse
- **Databases**: messages, media, posts, chat_participants, chats

---

### 142. Group Chat WebSocket (real-time nhóm)

- **Name**: ServeGroupWS
- **Description**:
  1. **Kết nối**: đọc mã truy cập từ tham số truy vấn của địa chỉ kết nối, kiểm tra người dùng còn hoạt động, nâng cấp kết nối sang WebSocket và đăng ký kết nối vào nhóm phát sóng của chat nhóm. Kết nối này chỉ dành cho nhắn tin nhóm, mọi sự kiện gọi nhóm bị chặn và phải dùng kết nối gọi nhóm riêng.
  2. **client→server events**:
     - group:join — kiểm tra quyền vào phòng gồm tư cách thành viên và không bị cấm, tham gia phòng, tải toàn bộ lịch sử tin nhắn gồm nội dung đã giải mã, hồ sơ người gửi, thông tin phương tiện và thời lượng, bài viết được chia sẻ, bản xem trước tin trả lời, trạng thái đã xem và biểu tượng cảm xúc; tải thêm lịch sử cuộc gọi nhóm từ cơ sở dữ liệu lịch sử cuộc gọi; trả về lịch sử và kèm danh sách tin được ghim nếu có.
     - group:message:send — kiểm tra tư cách thành viên, lệnh cấm và tắt tiếng, mã hóa nội dung tin nhắn, kiểm tra tin trả lời, phương tiện, bài viết được chia sẻ và tin chuyển tiếp; phát sóng tin nhắn mới cho cả phòng.
     - group:typing:start và group:typing:stop — chống gửi trùng theo trạng thái gõ của từng người trong từng nhóm, phát sóng trạng thái đang gõ kèm định danh nhóm, người dùng và cờ đang gõ.
     - group:message:search — tìm kiếm tin nhắn bằng cách giải mã rồi lọc trong bộ nhớ; trả kết quả tìm kiếm riêng cho người gửi.
     - group:leave — rời nhóm theo cách công khai hoặc im lặng và giữ lại hoặc ẩn danh hóa lịch sử tin nhắn, rời phòng phát sóng; nếu công khai thì phát sóng sự kiện thành viên đã rời.
     - group:member:add — tạo lời mời vào nhóm ở trạng thái chờ xử lý; phát sóng sự kiện thành viên được thêm.
     - group:member:ban — chặn thành viên khỏi nhóm; phát sóng sự kiện thành viên bị chặn.
     - group:member:mute và group:member:unmute — tắt tiếng hoặc bỏ tắt tiếng thành viên kèm thời hạn; phát sóng sự kiện tắt tiếng kèm bản ghi tắt tiếng, hoặc sự kiện bỏ tắt tiếng.
     - group:admin:transfer — chuyển quyền quản trị kèm thời gian chờ giữa hai lần chuyển là một tháng; phát sóng sự kiện đã chuyển quyền quản trị.
     - group:settings:update — cập nhật thông báo cá nhân, hoặc các thiết lập chỉ quản trị viên được đổi gồm tên nhóm, ảnh đại diện và tính năng cho phép thành viên tự thêm người; phát sóng sự kiện cập nhật cấu hình kèm người thực hiện, loại thay đổi và nội dung thông báo.
     - Các sự kiện bổ sung: group:message:pin và group:message:unpin ghim hoặc bỏ ghim tin nhắn kèm phát sóng, tối đa hai tin ghim; group:message:read đánh dấu đã đọc qua điểm đọc tin nhắn và phát sóng sự kiện đã đọc; group:message:react bật tắt biểu tượng cảm xúc trên tin nhắn và phát sóng sự kiện phản ứng; group:message:delete xóa tin nhắn cho riêng mình hoặc cho cả nhóm và phát sóng sự kiện tin nhắn đã bị xóa.
  3. **server→client events**: group:history, group:message:new, group:typing, group:message:search_result, group:member:left, group:member:added, group:member:banned, group:member:muted, group:member:unmuted, group:admin:transferred, group:settings:updated, group:message:pinned, group:message:unpinned, group:message:pinned_list, group:message:read, group:message:reacted, group:message:deleted, group:call:incoming (khi có cuộc gọi đến từ hệ thống gọi), error.
  4. **Ngắt kết nối**: khi kết nối ngắt thì dọn dẹp phiên gọi nếu kết nối đang ở chế độ gọi, hủy đăng ký khỏi nhóm phát sóng, xóa khỏi mọi phòng và danh sách kết nối.
- **Entity**: user (thành viên nhóm kết nối WS), các thành viên cùng nhóm, admin/owner
- **Model**: models.Chat, models.ChatParticipant, models.Message, models.GroupChatBan, models.GroupChatMute, models.GroupChatSettings, models.GroupChatMemberSettings, models.PinnedMessage, models.MessageReaction, models.ChatRead, models.GroupCallDocument (MongoDB, lịch sử call)
- **Databases**: chats, chat_participants, group_members (chat_participants role), group_chat_settings, group_chat_bans, group_chat_mutes, group_chat_member_requests, group_chat_member_settings, messages, media, pinned_messages, chat_reads, chat_user_settings, notifications, profiles, users. **MongoDB**: collection `group_calls` (lịch sử cuộc gọi nhóm trả kèm `group:history`)

---

### 143. Group Call WebSocket (WebRTC signaling + lịch sử MongoDB)

- **Name**: ServeGroupCallWS
- **Description**:
  1. **Kết nối**: xác thực mã truy cập từ tham số truy vấn và kiểm tra người dùng còn hoạt động, nâng cấp kết nối WebSocket, đăng ký kết nối vào nhóm phát sóng gọi nhóm có liên kết với nhóm phát sóng chat nhóm; endpoint này chỉ chấp nhận sự kiện gọi nhóm và chặn mọi sự kiện chat. Nhóm phát sóng gọi tự chạy định kỳ để dọn dẹp các phiên gọi cũ.
  2. **client→server events** (tín hiệu WebRTC):
     - group:call:create — kiểm tra tư cách thành viên, chỉ hỗ trợ gọi video; tạo phiên gọi mới, mỗi nhóm chỉ có một cuộc gọi đang sống; chọn người tham gia là các thành viên được mời hoặc toàn bộ thành viên; cuộc gọi tự hết hạn sau 60 giây nếu không ai tham gia; lưu bản ghi cuộc gọi ở trạng thái đang gọi vào cơ sở dữ liệu lịch sử cuộc gọi; trả về sự kiện đã tạo cuộc gọi cho người tạo, phát sóng sự kiện cuộc gọi đến cho cả phòng nhóm và gửi riêng cho từng thành viên được chọn; tạo tin nhắn hệ thống báo cuộc gọi đã bắt đầu trong nhóm.
     - group:call:request-join — nếu người dùng đã được mời thì tham gia ngay, trả xác nhận đã tham gia và phát sóng sự kiện đã tham gia cho những người đang trong cuộc gọi; nếu chưa được mời thì tạo yêu cầu chờ, gửi yêu cầu duyệt cho người tạo và xác nhận đã gửi yêu cầu cho người xin; yêu cầu chờ quá 30 giây sẽ bị từ chối vì hết thời gian.
     - group:call:approve-join và group:call:reject-join — chỉ người tạo cuộc gọi được duyệt hoặc từ chối; duyệt thì cập nhật danh sách người tham gia, lưu lại vào cơ sở dữ liệu lịch sử cuộc gọi và phát sóng sự kiện đã tham gia; từ chối thì gửi sự kiện yêu cầu tham gia bị từ chối.
     - group:call:signal — chuyển tiếp thông tin tín hiệu thiết lập kết nối WebRTC giữa những người tham gia đã vào cuộc gọi, đánh dấu người tham gia đang hoạt động rồi gửi tín hiệu tới các người tham gia còn lại kèm định danh cuộc gọi, người gửi và nội dung tín hiệu.
     - group:call:toggle-mic — người dùng tự tắt hoặc mở mic, phát sóng sự kiện trạng thái mic kèm người dùng, trạng thái tắt mic và người thao tác.
     - group:call:toggle-mute — người tạo cuộc gọi tắt mic của người tham gia khác, không được tự tắt mic chính mình, phát sóng sự kiện trạng thái mic.
     - group:call:toggle-video — bật hoặc tắt camera, phát sóng sự kiện trạng thái camera kèm người dùng và trạng thái bật camera.
     - group:call:list-participants và group:call:get-participants — trả bản chụp danh sách người tham gia kèm danh sách đã tham gia và danh sách người tham gia đang hoạt động.
     - group:call:end — nếu là người tạo thì đóng cả phiên gọi; nếu là người tham gia thường thì chỉ rời đi, phiên gọi kết thúc khi không còn ai; khi phiên gọi đóng thì lưu trạng thái đã kết thúc hoặc đã hủy nếu có tối đa một người tham gia kèm thời điểm kết thúc; gửi sự kiện cuộc gọi đã kết thúc cho những người đang trong cuộc gọi, phát sóng cho phòng nhóm và tạo tin nhắn hệ thống báo cuộc gọi đã kết thúc; nếu chỉ rời đi thì gửi sự kiện đã rời cuộc gọi.
  3. **Cleanup khi mất kết nối / timeout**: khi kết nối ngắt thì kết thúc mọi cuộc gọi mà người dùng đang tham gia; tiến trình định kỳ mỗi 30 giây dọn các cuộc gọi hết hạn 60 giây bằng cách gửi sự kiện cuộc gọi đã kết thúc vì hết thời gian, tạo tin nhắn hệ thống báo gọi quá thời gian và lưu trạng thái kết thúc; đồng thời từ chối các yêu cầu tham gia chờ quá 30 giây với lý do hết thời gian.
  4. **server→client events**: group:call:created, group:call:incoming, group:call:joined, group:call:joined_ack, group:call:join-request, group:call:join-request-sent, group:call:join-rejected, group:call:signal, group:call:mic, group:call:video, group:call:participants, group:call:left, group:call:ended, error.
- **Entity**: caller/creator (người tạo call), participants (được invite), người xin join, các thành viên nhóm (nhận `group:call:incoming` qua groupHub + tin hệ thống)
- **Model**: models.Chat, models.ChatParticipant, models.Message (tin hệ thống call_started/call_ended/call_timeout), repository.GroupCallDocument (MongoDB, struct `bson` tags)
- **Databases**: **MongoDB**: collection `group_calls` (group-call records — tạo khi `group:call:create`, update `participants` khi approve-join, update `status`/`ended_at` khi end/timeout). **MySQL** (phụ trợ): chats, chat_participants, messages, group_chat_member_settings, profiles, users

## B8 — Voice/Video Call API Documentation (HTTP + WebSocket)

> Tài liệu hóa 12 endpoint cuộc gọi thoại/video (144–155). Mỗi endpoint gồm Name, Description, Entity, Model, Databases.

---

### 144. Kết nối WebSocket tín hiệu cuộc gọi

- **Name**: HandleWebsocket (call)
- **Description**:
  1. **Kết nối**: Route GET /api/calls/ws được đăng ký riêng trước nhóm route /api/calls nên không đi qua middleware xác thực chung; endpoint tự xác thực bằng tham số truy vấn token dạng truy cập.
  2. **Xác thực**: Nếu ngữ cảnh đã có người dùng thì dùng thẳng; ngược lại đọc token từ truy vấn, phân tích và bắt buộc phải là token truy cập — thiếu hoặc sai thì báo lỗi chưa được xác thực.
  3. **Nâng cấp kết nối**: Nâng cấp kết nối WebSocket; kiểm tra nguồn gốc hiện chấp nhận mọi nguồn, khác với endpoint WebSocket khác vốn tôn trọng cấu hình phạm vi nguồn cho phép — điểm cần lưu ý khi vá bảo mật.
  4. **Đăng ký client**: Tạo client cho người dùng và đăng ký vào Hub; đồng thời chạy hai luồng: luồng ghi (ghi sự kiện ra socket và giữ kết nối bằng ping/pong) và luồng đọc (chặn vòng lặp xử lý sự kiện).
  5. **Sự kiện từ client**:
     - call:initiate — khởi tạo cuộc gọi, phản hồi call:initiated kèm mã cuộc gọi
     - call:accept — chấp nhận cuộc gọi
     - call:reject — từ chối cuộc gọi
     - call:end — kết thúc cuộc gọi
     - call:signal — kiểm tra dữ liệu không quá 8192 byte để chống gửi dữ liệu lớn rồi chuyển tiếp tín hiệu WebRTC
     - call:toggle_mute — bật/tắt tiếng
     - call:video_toggle — bật/tắt camera
     - call:busy — client xác nhận đã nhận thông báo bận, server bỏ qua
  6. **Sự kiện gửi về client**: call:incoming, call:status, call:busy, call:mute, call:video, call:signal, call:missed, call:cancelled — đẩy tới một hoặc nhiều người dùng.
  7. **Dọn dẹp khi mất kết nối**: Khi socket cuối cùng của người dùng đóng thì dọn dẹp các cuộc gọi còn dang dở, chuyển trạng thái đang gọi/đang chuông/đang kết nối sang đã hủy/bạn bỏ lỡ/đã kết thúc để tránh kẹt trạng thái chặn việc tạo cuộc gọi về sau.
  8. **Lưu ý**: Nếu dịch vụ gọi không khả dụng (endpoint dùng chung Hub cho thông báo) thì mọi sự kiện call bị bỏ qua kèm lỗi dịch vụ gọi không khả dụng.
- **Entity**: người gọi (caller), người nghe (callee), server Hub (ws.Hub)
- **Model**: models.Call, models.CallHidden, models.Friend, models.Profile, models.Notification
- **Databases**: calls, call_hidden, friends, profiles, notifications, push_tokens, notification_preferences

---

### 145. Lấy danh sách ICE/TURN servers

- **Name**: GetIceServers
- **Description**:
  1. Đọc cấu hình danh sách địa chỉ STUN từ môi trường (các địa chỉ cách nhau bởi dấu phẩy), tách từng địa chỉ, bỏ khoảng trắng thừa và loại địa chỉ rỗng.
  2. Nếu có cấu hình máy chủ TURN thì bổ sung thêm một máy chủ TURN kèm tên người dùng và mật mã truy cập.
  3. Nếu không có cấu hình nào thì trả về danh sách rỗng, không báo lỗi.
  4. Trả về danh sách máy chủ ICE để client khởi tạo kết nối WebRTC trước khi gửi tín hiệu qua WebSocket cuộc gọi.
  5. Không truy cập cơ sở dữ liệu.
- **Entity**: người dùng đăng nhập (client WebRTC), hệ thống hạ tầng STUN/TURN
- **Model**: không truy cập model (chỉ DTO `dto.IceServer`, `dto.IceServersResponse`)
- **Databases**: không truy cập DB (chỉ đọc env `ICE_SERVER_URLS`, `TURN_SERVER_URL`, `TURN_USERNAME`, `TURN_CREDENTIAL`)

---

### 146. Lấy lịch sử cuộc gọi (filter + phân trang)

- **Name**: GetCallHistory
- **Description**:
  1. Nhận các tham số lọc và phân trang: giới hạn (mặc định 20, ép trong khoảng 1 đến 100), vị trí bắt đầu (mặc định 0, không âm), loại cuộc gọi, trạng thái, trường sắp xếp và chiều sắp xếp.
  2. Lọc giá trị đầu vào: loại chỉ nhận gọi thoại hoặc gọi video; trạng thái chỉ nhận các trạng thái cuộc gọi cho phép; trường sắp xếp và chiều sắp xếp chỉ nhận tập giá trị chuẩn bị trước, đồng thời kiểm lại ở tầng truy vấn để tránh chèn lệnh tùy ý.
  3. Truy vấn các cuộc gọi mà người dùng là người gọi hoặc người nhận, loại những cuộc gọi họ đã ẩn, kèm bộ lọc và phân trang.
  4. Đếm tổng số cuộc gọi thỏa cùng bộ lọc để trả về cùng danh sách trang.
  5. Tải thông tin hồ sơ của đầu dây bên kia theo lô (tối đa 100 ID mỗi lần, loại trùng lặp trước khi truy vấn) thay vì tra từng bản ghi.
  6. Chuyển từng bản ghi thành mục lịch sử gồm đầu dây bên kia, hướng gọi ra hoặc gọi vào, cờ cuộc gọi nhỡ và thời gian quy đổi.
  7. Trả về danh sách cuộc gọi đã phân trang kèm tổng số, giới hạn và vị trí bắt đầu.
- **Entity**: người dùng xem lịch sử, đầu dây bên kia (other user)
- **Model**: models.Call, models.CallHidden, models.Profile
- **Databases**: calls, call_hidden, profiles

---

### 147. Đếm cuộc gọi nhỡ chưa đọc

- **Name**: GetMissedCallCount
- **Description**:
  1. Đọc mốc thời điểm đọc cuộc gọi nhỡ gần nhất của người dùng; nếu chưa từng đọc thì đếm toàn bộ.
  2. Đếm các cuộc gọi nhỡ mà người dùng là người nhận, loại những cuộc gọi họ đã ẩn, chỉ tính từ mốc thời điểm trở đi.
  3. Dùng phép so sánh lớn hơn hoặc bằng để không bỏ sót cuộc gọi nhỡ xảy ra cùng thời điểm với lúc đánh dấu đã đọc.
  4. Trả về số lượng để hiển thị huy hiệu cuộc gọi nhỡ.
- **Entity**: người dùng (chủ badge), hệ thống đếm
- **Model**: models.Profile, models.Call, models.CallHidden
- **Databases**: profiles, calls, call_hidden

---

### 148. Đánh dấu cuộc gọi nhỡ đã đọc

- **Name**: MarkMissedRead
- **Description**:
  1. Lấy người dùng từ phiên đăng nhập, không cần dữ liệu trong yêu cầu.
  2. Cập nhật mốc thời điểm đọc cuộc gọi nhỡ của người dùng thành thời điểm hiện tại.
  3. Nếu không tìm thấy hồ sơ của người dùng thì báo lỗi.
  4. Mốc mới là ngưỡng đếm cho các lần đọc sau, chỉ tính cuộc gọi nhỡ từ thời điểm này trở đi.
  5. Trả về thông báo đã đánh dấu đã đọc; gọi lại nhiều lần vẫn thành công.
- **Entity**: người dùng (chủ badge cuộc gọi nhỡ)
- **Model**: models.Profile
- **Databases**: profiles

---

### 149. Ẩn cuộc gọi khỏi lịch sử (soft-delete)

- **Name**: HideCall
- **Description**:
  1. Kiểm tra mã cuộc gọi trong đường dẫn không được rỗng.
  2. Tra cứu cuộc gọi: không tồn tại thì báo lỗi cuộc gọi không tồn tại; nếu người dùng không phải người gọi hoặc người nhận thì báo lỗi không phải người tham gia cuộc gọi, không cho ẩn cuộc gọi của người khác.
  3. Ghi một bản ghi ẩn gắn giữa cuộc gọi và người dùng thao tác; nếu đã ẩn rồi thì vẫn thành công, không báo lỗi.
  4. Bản ghi chỉ gắn với người thực hiện thao tác nên cuộc gọi vẫn hiển thị cho phía bên kia; các lần xem lịch sử và đếm cuộc gọi nhỡ đều loại các cuộc gọi đã ẩn.
  5. Dữ liệu cuộc gọi gốc không bị xóa.
  6. Trả về thông báo cuộc gọi đã được ẩn khỏi lịch sử.
- **Entity**: người dùng thao tác (participant), đầu dây bên kia (vẫn thấy lịch sử)
- **Model**: models.Call, models.CallHidden
- **Databases**: calls, call_hidden

---

### 150. Xem chi tiết một cuộc gọi

- **Name**: GetCallDetail
- **Description**:
  1. Kiểm tra mã cuộc gọi trong đường dẫn không được rỗng.
  2. Tra cứu cuộc gọi theo mã; không tìm thấy thì báo lỗi cuộc gọi không tồn tại.
  3. Kiểm tra người dùng phải là người gọi hoặc người nhận; nếu không thì báo lỗi không phải người tham gia cuộc gọi.
  4. Trả về chi tiết cuộc gọi: mã, hai đầu dây, loại cuộc gọi, có phải gọi nhóm hay không, trạng thái, thời gian bắt đầu và kết thúc, thời lượng, trạng thái tắt tiếng và bật camera của từng bên, thời điểm tạo.
- **Entity**: người dùng (participant), cuộc gọi được xem
- **Model**: models.Call
- **Databases**: calls

---

### 151. Khởi tạo cuộc gọi (chống bận)

- **Name**: InitiateCall
- **Description**:
  1. Nhận mã người cần gọi và loại cuộc gọi, chuẩn hóa loại về thoại hoặc video (để trống hoặc giá trị lạ thì coi là gọi thoại), từ chối trường hợp tự gọi chính mình.
  2. Kiểm tra hai người phải là bạn đã chấp nhận kết bạn; nếu không thì báo lỗi.
  3. Tạo bản ghi cuộc gọi mới với trạng thái đang gọi, thời lượng bằng 0 và camera tắt.
  4. Trong một transaction, kiểm tra xem người gọi hoặc người nhận có đang dính vào cuộc gọi đang chờ hoặc đang kết nối không; nếu có thì đánh dấu bận và không tạo bản ghi, khóa các dòng đang hoạt động để hai yêu cầu đồng thời không cùng vượt qua.
  5. Nếu đang bận: gửi sự kiện bận cho người gọi và trả về thông báo người dùng đang bận, không phải lỗi.
  6. Nếu người nhận đang online: gửi sự kiện cuộc gọi đến để họ đổ chuông.
  7. Nếu người nhận offline: chuyển ngay trạng thái thành nhỡ kèm thời điểm kết thúc, báo cho người gọi biết cuộc gọi nhỡ, và tạo thông báo đã gọi nhỡ để người nhận thấy khi quay lại (kiểm tra cài đặt thông báo và gửi push nếu có thiết bị).
  8. Luôn báo trạng thái đang gọi cho người gọi và trả về thông tin cuộc gọi vừa tạo.
  9. Nhánh sự kiện call:initiate qua WebSocket xử lý giống hệt và phản hồi trạng thái đã khởi tạo.
- **Entity**: người gọi (caller), người nghe (callee), server Hub, hệ thống notification
- **Model**: models.Call, models.Friend, models.Notification, models.Profile
- **Databases**: calls, friends, notifications, push_tokens, notification_preferences, profiles

---

### 152. Chấp nhận cuộc gọi (atomic)

- **Name**: AcceptCall
- **Description**:
  1. Kiểm tra mã cuộc gọi trong đường dẫn không được rỗng, lấy người dùng từ phiên đăng nhập.
  2. Chỉ cập nhật trạng thái thành đang kết nối kèm thời điểm bắt đầu khi cuộc gọi còn đang chờ và người thực hiện đúng là người nhận — cập nhật có điều kiện trong một transaction nên không xảy ra đọc, kiểm tra rồi ghi rời rạc.
  3. Nếu không có bản ghi nào được cập nhật thì tra lại để phân biệt hai lỗi: cuộc gọi không tồn tại hay cuộc gọi không ở trạng thái chờ.
  4. Ghi thời điểm chấp nhận làm mốc bắt đầu, dùng để tính thời lượng khi cuộc gọi kết thúc.
  5. Báo trạng thái đã kết nối cho cả hai phía để hiển thị giao diện gọi.
  6. Trả về thông báo cuộc gọi đã được chấp nhận; nếu hai lần chấp nhận chạy đồng thời thì chỉ một lần thành công.
  7. Nhánh sự kiện call:accept qua WebSocket xử lý giống hệt.
- **Entity**: người nghe (callee — người chấp nhận), người gọi (caller), server Hub
- **Model**: models.Call
- **Databases**: calls

---

### 153. Từ chối cuộc gọi (atomic)

- **Name**: RejectCall
- **Description**:
  1. Kiểm tra mã cuộc gọi trong đường dẫn không được rỗng.
  2. Chỉ cập nhật trạng thái thành đã từ chối kèm thời điểm kết thúc khi cuộc gọi còn đang chờ và người thực hiện đúng là người nhận — cập nhật có điều kiện trong một transaction.
  3. Nếu không có bản ghi nào được cập nhật thì tra lại để phân biệt cuộc gọi không tồn tại với cuộc gọi không ở trạng thái chờ.
  4. Khác với chấp nhận ở chỗ chỉ ghi thời điểm kết thúc, không ghi thời điểm bắt đầu và thời lượng giữ bằng 0 vì cuộc gọi không diễn ra.
  5. Báo trạng thái đã từ chối cho cả hai phía.
  6. Trả về thông báo cuộc gọi đã bị từ chối; nếu từ chối và chấp nhận đua nhau thì chỉ một bên thắng dựa trên điều kiện trạng thái.
  7. Nhánh sự kiện call:reject qua WebSocket xử lý giống hệt.
- **Entity**: người nghe (callee — người từ chối), người gọi (caller), server Hub
- **Model**: models.Call
- **Databases**: calls

---

### 154. Bật/tắt tiếng (mute) trong cuộc gọi

- **Name**: ToggleMute (call)
- **Description**:
  1. Kiểm tra mã cuộc gọi và nhận trạng thái tắt tiếng muốn đặt (ghi đè trực tiếp, không phải đảo trạng thái).
  2. Tra cứu cuộc gọi; không tồn tại thì báo lỗi cuộc gọi không tồn tại.
  3. Bắt buộc cuộc gọi đang kết nối; nếu chưa kết nối thì báo lỗi cuộc gọi chưa kết nối, không cho tắt tiếng khi còn đang reng.
  4. Phân biệt theo vai trò: người gọi cập nhật trạng thái tắt tiếng của bản thân, người nhận cập nhật của bản thân, người ngoài thì báo lỗi không phải người tham gia.
  5. Cập nhật chỉ ghi vào trường của người thao tác; nếu không có bản ghi nào bị ảnh hưởng thì báo lỗi.
  6. Báo trạng thái tắt tiếng cho cả hai phía để đầu dây bên kia thấy biểu tượng mute theo thời gian thực.
  7. Trả về thông báo đã cập nhật trạng thái tắt tiếng.
  8. Nhánh sự kiện call:toggle_mute qua WebSocket xử lý giống hệt.
- **Entity**: người thao tác (caller hoặc callee), đầu dây bên kia (nhận sự kiện)
- **Model**: models.Call
- **Databases**: calls

---

### 155. Bật/tắt camera (video) trong cuộc gọi

- **Name**: ToggleVideo
- **Description**:
  1. Kiểm tra mã cuộc gọi và nhận trạng thái camera muốn đặt.
  2. Tra cứu cuộc gọi; không tồn tại thì báo lỗi cuộc gọi không tồn tại.
  3. Bắt buộc đây phải là cuộc gọi video; nếu là cuộc gọi giọng nói thì báo lỗi không phải cuộc gọi video.
  4. Bắt buộc cuộc gọi đang kết nối; nếu chưa kết nối thì báo lỗi cuộc gọi chưa kết nối.
  5. Phân biệt theo vai trò: người gọi cập nhật camera của bản thân, người nhận cập nhật của bản thân, người ngoài thì báo lỗi không phải người tham gia.
  6. Cập nhật chỉ ghi vào trường của người thao tác; nếu không có bản ghi nào bị ảnh hưởng thì báo lỗi.
  7. Báo trạng thái camera cho cả hai phía, dùng làm tín hiệu bật/tắt camera hai chiều.
  8. Trả về thông báo đã cập nhật trạng thái video.
  9. Nhánh sự kiện call:video_toggle qua WebSocket xử lý giống hệt.
- **Entity**: người thao tác (caller hoặc callee), đầu dây bên kia (nhận sự kiện)
- **Model**: models.Call
- **Databases**: calls

## B9a — Communities (user-facing) API Documentation

> Tài liệu hóa 23 endpoint Communities dành cho người dùng (routes/community.routes.go, controllers/community.controller.go, services/community.service.go). Mỗi endpoint gồm Name, Description, Entity, Model, Databases.

---

### 156. Danh sách cộng đồng (khám phá)

- **Name**: ListCommunities
- **Description**:
  1. Người dùng mở trang khám phá cộng đồng, có thể nhập từ khóa tìm theo tên và chọn trang muốn xem.
  2. Nếu request có kèm thông tin đăng nhập hợp lệ thì hệ thống ghi nhận danh tính người xem; không có thông tin đăng nhập vẫn xử lý bình thường.
  3. Hệ thống chuẩn hóa thông tin phân trang (trang mặc định là 1, số mục mỗi trang mặc định là 10, giới hạn trong khoảng 1 đến 50) và tìm cộng đồng theo tên nếu có từ khóa.
  4. Hệ thống chỉ lấy các cộng đồng đang hoạt động, đếm tổng số cộng đồng khớp yêu cầu rồi lấy danh sách theo trang.
  5. Mỗi cộng đồng kèm số lượng thành viên; danh sách sắp xếp cộng đồng đông thành viên nhất lên trước, cùng số thì cộng đồng tạo gần đây nhất lên trước.
  6. Nếu người dùng đang đăng nhập, mỗi mục đánh dấu đây có phải cộng đồng do mình tạo hay không; không trả về thông tin định danh của người tạo cộng đồng.
  7. Trả về danh sách cộng đồng đã phân trang kèm tổng số; thao tác chỉ đọc, không thay đổi dữ liệu.
- **Entity**: người xem (đăng nhập hoặc khách), cộng đồng
- **Model**: models.Community, models.GroupMember
- **Databases**: communities, group_members

### 157. Chi tiết cộng đồng

- **Name**: GetCommunityDetail
- **Description**:
  1. Người dùng mở trang chi tiết của một cộng đồng; hệ thống cho phép xem mà không bắt buộc đăng nhập.
  2. Nếu không có mã cộng đồng hợp lệ thì báo lỗi không tìm thấy cộng đồng.
  3. Hệ thống tìm cộng đồng theo mã; không tồn tại thì báo lỗi không tìm thấy cộng đồng, nếu có thì đếm số lượng thành viên hiện tại.
  4. Nếu người xem đang đăng nhập, hệ thống xác định trạng thái quan hệ của họ với cộng đồng theo thứ tự: là người tạo cộng đồng, là thành viên (kèm vai trò trong cộng đồng), đã gửi yêu cầu tham gia còn chờ duyệt, hoặc chưa tham gia gì.
  5. Hệ thống bổ sung tên hiển thị của người tạo cộng đồng; nếu không lấy được thì bỏ qua, không làm hỏng yêu cầu.
  6. Trả về thông tin chi tiết của cộng đồng; thao tác chỉ đọc, không thay đổi dữ liệu.
- **Entity**: người xem, cộng đồng, người tạo cộng đồng
- **Model**: models.Community, models.GroupMember, models.UserRole, models.CommunityJoinRequest, models.User
- **Databases**: communities, group_members, user_roles, roles, community_join_requests, users

### 158. Bài viết trong cộng đồng

- **Name**: GetCommunityPosts
- **Description**:
  1. Người dùng mở trang bài viết của một cộng đồng; hệ thống cho phép xem mà không bắt buộc đăng nhập.
  2. Nếu không có mã cộng đồng hợp lệ thì báo lỗi không tìm thấy cộng đồng.
  3. Người dùng có thể tải tiếp bài viết cũ bằng con trỏ phân trang (gồm thời điểm tạo bài và mã bài viết, hoặc chỉ mã bài viết); số bài mỗi trang mặc định là 10, giới hạn trong khoảng 1 đến 50.
  4. Hệ thống kiểm tra cộng đồng tồn tại và đang hoạt động; cộng đồng ngừng hoạt động thì báo lỗi không tìm thấy cộng đồng.
  5. Hệ thống diễn giải con trỏ phân trang: nếu con trỏ chỉ có mã bài viết thì tra cứu bài viết đó để lấy thời điểm tạo làm mốc tiếp tục tải.
  6. Hệ thống lấy các bài viết công khai của cộng đồng kèm thông tin tác giả (tên hiển thị, ảnh đại diện); nếu người xem đang đăng nhập thì mỗi bài còn gắn dấu hiệu người xem có theo dõi tác giả hay không.
  7. Bài viết được sắp xếp mới nhất lên trước; trả về danh sách bài viết của cộng đồng, thao tác chỉ đọc, không thay đổi dữ liệu.
- **Entity**: người xem, bài viết của cộng đồng, tác giả bài viết
- **Model**: models.Post, models.Community, models.User, models.Profile
- **Databases**: posts, communities, users, profiles, follows

### 159. Cộng đồng đã tham gia

- **Name**: ListJoinedCommunities
- **Description**:
  1. Người dùng đăng nhập rồi mở danh sách cộng đồng mình đã tham gia.
  2. Nếu request không có thông tin đăng nhập hợp lệ thì hệ thống yêu cầu đăng nhập.
  3. Người dùng có thể nhập từ khóa và chọn trang muốn xem; thông tin phân trang được chuẩn hóa giống trang khám phá cộng đồng.
  4. Hệ thống lấy các cộng đồng đang hoạt động mà người dùng là thành viên, kèm số lượng thành viên của mỗi cộng đồng.
  5. Danh sách sắp xếp cộng đồng tạo gần đây nhất lên trước và phân theo trang.
  6. Trả về danh sách cộng đồng đã tham gia; thao tác chỉ đọc, không thay đổi dữ liệu.
- **Entity**: người dùng đang đăng nhập, cộng đồng mình là thành viên
- **Model**: models.Community, models.GroupMember
- **Databases**: communities, group_members

### 160. Cộng đồng đã tạo

- **Name**: ListCreatedCommunities
- **Description**:
  1. Người dùng đăng nhập rồi mở danh sách cộng đồng mình đã tạo.
  2. Nếu request không có thông tin đăng nhập hợp lệ thì hệ thống yêu cầu đăng nhập.
  3. Người dùng có thể nhập từ khóa và chọn trang muốn xem; thông tin phân trang được chuẩn hóa giống trang khám phá cộng đồng.
  4. Hệ thống lấy các cộng đồng đang hoạt động mà người dùng là người tạo, kèm số lượng thành viên của mỗi cộng đồng.
  5. Danh sách sắp xếp cộng đồng tạo gần đây nhất lên trước và phân theo trang.
  6. Trả về danh sách cộng đồng đã tạo; thao tác chỉ đọc, không thay đổi dữ liệu.
- **Entity**: người dùng đang đăng nhập (creator), cộng đồng do mình tạo
- **Model**: models.Community, models.GroupMember
- **Databases**: communities, group_members

### 161. Tạo cộng đồng

- **Name**: CreateCommunity
- **Description**:
  1. Người dùng đăng nhập rồi mở biểu mẫu tạo cộng đồng mới; hệ thống giới hạn tối đa 10 yêu cầu tạo cộng đồng mỗi phút để chống spam.
  2. Người dùng nhập tên cộng đồng, mô tả, chọn quyền riêng tư, bật hoặc tắt tự động duyệt yêu cầu tham gia, đồng thời có thể chọn ảnh đại diện và ảnh bìa cho cộng đồng.
  3. Hệ thống kiểm tra kích thước dữ liệu gửi lên (quá lớn thì báo file quá lớn) và kiểm tra đúng định dạng biểu mẫu; sai định dạng thì báo lỗi thông tin cộng đồng không hợp lệ.
  4. Hệ thống tải từng ảnh đã chọn lên kho lưu trữ; nếu tải ảnh bìa thất bại thì xóa luôn ảnh đại diện đã tải trước đó để tránh dữ liệu thừa.
  5. Hệ thống kiểm tra thông tin cộng đồng: tên, mô tả, ảnh đại diện, ảnh bìa và quyền riêng tư; quản trị viên cấp cao của hệ thống không được tạo cộng đồng; mỗi người tối đa 10 cộng đồng và tối đa 3 cộng đồng trong 24 giờ; tài khoản phải đang hoạt động; tên cộng đồng đã tồn tại thì báo lỗi trùng tên.
  6. Hệ thống tạo cộng đồng mới và trong cùng một thao tác ghi: ghi người tạo vào danh sách thành viên kèm điểm khởi đầu, gán vai trò quản trị cộng đồng và quản trị nhóm chat cho người tạo, tạo nhóm chat riêng cho cộng đồng kèm khóa mã hóa và thêm người tạo vào nhóm chat với vai trò quản trị.
  7. Hệ thống cài đặt mặc định cho cộng đồng: tạo vài quy tắc mẫu và một mã mời dùng chung không giới hạn số lượt; bước này nếu thất bại chỉ ghi lại thông tin, không làm hỏng yêu cầu.
  8. Hệ thống gửi thông báo cho người tạo rằng cộng đồng và nhóm chat đã được tạo; trả về thông báo thành công kèm mã cộng đồng, thông tin cộng đồng, nhóm chat mặc định và mã mời (nếu cài đặt mặc định thành công).
- **Entity**: người tạo cộng đồng (creator), cộng đồng mới, group chat mặc định, media avatar/background, notification
- **Model**: models.Community, models.GroupMember, models.UserRole, models.Chat, models.ChatParticipant, models.CommunityRule, models.CommunityInviteCode, models.Media, models.User, models.Notification
- **Databases**: communities, group_members, user_roles, roles, chats, chat_participants, community_rules, community_invite_codes, media, users, notifications

### 162. Cập nhật thông tin cộng đồng

- **Name**: UpdateCommunity
- **Description**:
  1. Người dùng đăng nhập rồi gửi yêu cầu sửa thông tin cộng đồng, chỉ cần điền các trường muốn thay đổi: tên (3 đến 100 ký tự), mô tả (tối đa 500 ký tự), quyền riêng tư hoặc bật tắt tự động duyệt.
  2. Nếu dữ liệu gửi lên sai định dạng thì báo lỗi thông tin cộng đồng không hợp lệ.
  3. Hệ thống kiểm tra người dùng phải là người tạo hoặc quản trị viên cộng đồng; thành viên thường không được sửa thông tin cộng đồng.
  4. Hệ thống loại bỏ khoảng trắng thừa ở tên và mô tả rồi kiểm tra tính hợp lệ của các trường được gửi lên.
  5. Chỉ lưu các trường có dữ liệu mới; nếu đổi tên cộng đồng thì kiểm tra tên mới đã bị trùng chưa, tên giữ nguyên thì bỏ qua.
  6. Nếu không có trường nào hợp lệ để cập nhật thì báo lỗi không có dữ liệu nào cần sửa.
  7. Hệ thống chỉ cập nhật các trường được gửi lên; tên trùng do nhiều người đổi cùng lúc thì báo lỗi trùng tên, các lỗi khác thì báo cập nhật thất bại.
  8. Trả về thông báo cập nhật thành công.
- **Entity**: creator/admin cộng đồng, cộng đồng
- **Model**: models.Community, models.UserRole, models.Role
- **Databases**: communities, user_roles, roles

### 163. Cập nhật ảnh bìa (background) cộng đồng

- **Name**: SetCommunityBackground
- **Description**:
  1. Người dùng đăng nhập rồi chọn ảnh bìa mới cho cộng đồng.
  2. Nếu request gửi thiếu file ảnh bìa thì báo lỗi yêu cầu chọn ảnh bìa; kích thước dữ liệu gửi lên cũng bị giới hạn như khi tạo cộng đồng.
  3. Hệ thống kiểm tra người dùng phải là người tạo hoặc quản trị viên cộng đồng, cộng đồng phải tồn tại và tài khoản phải đang hoạt động.
  4. Hệ thống tải ảnh bìa lên kho lưu trữ; tải thất bại thì báo lỗi tải ảnh bìa thất bại.
  5. Hệ thống kiểm tra đường dẫn ảnh vừa tải hợp lệ rồi cập nhật ảnh bìa của cộng đồng.
  6. Nếu cập nhật thất bại sau khi đã tải ảnh thì hệ thống xóa luôn ảnh vừa tải để tránh dữ liệu thừa rồi báo lỗi cập nhật thất bại.
  7. Trả về thông báo cập nhật ảnh bìa thành công.
- **Entity**: creator/admin cộng đồng, cộng đồng, media background
- **Model**: models.Community, models.Media, models.User
- **Databases**: communities, media, user_roles, roles, users

### 164. Yêu cầu tham gia cộng đồng

- **Name**: RequestJoin
- **Description**:
  1. Người dùng đăng nhập rồi gửi yêu cầu tham gia một cộng đồng, có thể kèm mã mời hoặc lời mời nhận trước đó; nếu không có mã cộng đồng hợp lệ thì báo lỗi không tìm thấy cộng đồng.
  2. Hệ thống kiểm tra người dùng tồn tại và tài khoản đang hoạt động, cộng đồng phải tồn tại; nếu đã là thành viên rồi thì báo lỗi đã tham gia cộng đồng.
  3. Nếu cộng đồng chỉ nhận lời mời thì người dùng bắt buộc phải có lời mời hợp lệ: lời mời phải tồn tại, đúng người được mời và còn đang chờ phản hồi; khi hợp lệ hệ thống thêm người dùng vào cộng đồng và nhóm chat, đánh dấu lời mời đã chấp nhận và thông báo cho người đã gửi lời mời.
  4. Nếu cộng đồng công khai và người dùng nhập mã mời thì hệ thống kiểm tra mã còn hiệu lực (chưa bị tắt, chưa quá hạn, chưa hết số lượt dùng) rồi thêm thành viên và nhóm chat, tăng số lần dùng mã và gửi thông báo.
  5. Nếu cộng đồng công khai và người dùng có lời mời thì xử lý như nhánh lời mời ở trên.
  6. Nếu cộng đồng công khai mà không kèm mã hay lời mời, hệ thống kiểm tra cài đặt tự động duyệt của cộng đồng: cộng đồng tự động duyệt thì thêm người dùng vào cộng đồng và nhóm chat ngay và gửi thông báo cho người dùng.
  7. Nếu cộng đồng cần chờ duyệt thì hệ thống chặn khi người dùng đã có yêu cầu đang chờ, xóa các yêu cầu cũ không còn ở trạng thái chờ, tạo yêu cầu tham gia mới ở trạng thái chờ duyệt và thông báo cho người tạo cộng đồng.
  8. Trả về thông báo: đã tham gia ngay hoặc yêu cầu của bạn đang chờ quản trị cộng đồng duyệt.
- **Entity**: người xin tham gia, cộng đồng, creator/admin (nhận notification), mã mời, lời mời
- **Model**: models.User, models.Community, models.CommunityJoinRequest, models.CommunityInviteCode, models.CommunityInvitation, models.GroupMember, models.UserRole, models.Chat, models.ChatParticipant, models.Notification
- **Databases**: users, communities, user_roles, roles, community_join_requests, community_invite_codes, community_invitations, group_members, chats, chat_participants, notifications

### 165. Danh sách yêu cầu tham gia đang chờ

- **Name**: ListPendingJoinRequests
- **Description**:
  1. Người dùng là quản trị cộng đồng mở danh sách các yêu cầu tham gia còn đang chờ duyệt.
  2. Nếu không có mã cộng đồng hợp lệ thì báo lỗi không tìm thấy cộng đồng.
  3. Hệ thống kiểm tra người dùng phải là người tạo hoặc quản trị viên cộng đồng; thành viên thường không được xem danh sách này.
  4. Hệ thống lấy các yêu cầu tham gia đang chờ của cộng đồng theo thứ tự yêu cầu gửi sớm nhất lên trước.
  5. Mỗi yêu cầu kèm thông tin người xin tham gia (tên hiển thị, ảnh đại diện); nếu thiếu thông tin cá nhân thì để trống.
  6. Trả về danh sách yêu cầu đang chờ; thao tác chỉ đọc, không thay đổi dữ liệu.
- **Entity**: creator/admin cộng đồng, người xin tham gia
- **Model**: models.CommunityJoinRequest, models.Profile
- **Databases**: community_join_requests, profiles, user_roles, roles

### 166. Duyệt yêu cầu tham gia

- **Name**: ApproveJoinRequest
- **Description**:
  1. Người dùng là quản trị cộng đồng mở một yêu cầu tham gia đang chờ rồi chọn duyệt.
  2. Nếu không có mã yêu cầu hợp lệ thì báo lỗi không tìm thấy yêu cầu tham gia.
  3. Hệ thống tra cứu yêu cầu theo mã; không tìm thấy thì báo lỗi không tìm thấy yêu cầu tham gia, yêu cầu đã được xử lý rồi thì báo không thể duyệt lại.
  4. Hệ thống xác định cộng đồng thuộc về yêu cầu rồi kiểm tra người thao tác phải là người tạo hoặc quản trị viên cộng đồng; thành viên thường không được duyệt.
  5. Hệ thống tìm nhóm chat mặc định của cộng đồng; nếu không có thì báo lỗi không tìm thấy nhóm chat cộng đồng.
  6. Trong một thao tác ghi, hệ thống đánh dấu yêu cầu đã được duyệt kèm thời điểm phản hồi, thêm người xin vào danh sách thành viên cộng đồng, gán vai trò thành viên và thêm vào nhóm chat mặc định.
  7. Hệ thống gửi thông báo cho người xin tham gia rằng họ đã được duyệt vào cộng đồng và được thêm vào nhóm chat.
  8. Trả về thông báo duyệt yêu cầu thành công.
- **Entity**: creator/admin cộng đồng, người xin tham gia
- **Model**: models.CommunityJoinRequest, models.GroupMember, models.UserRole, models.Chat, models.ChatParticipant, models.Notification
- **Databases**: community_join_requests, group_members, user_roles, roles, chats, chat_participants, notifications

### 167. Từ chối yêu cầu tham gia

- **Name**: RejectJoinRequest
- **Description**:
  1. Người dùng là quản trị cộng đồng mở một yêu cầu tham gia đang chờ rồi chọn từ chối.
  2. Nếu không có mã yêu cầu hợp lệ thì báo lỗi không tìm thấy yêu cầu tham gia.
  3. Hệ thống tra cứu yêu cầu theo mã; không tìm thấy thì báo lỗi không tìm thấy yêu cầu tham gia, yêu cầu đã được xử lý rồi thì báo không thể từ chối lại.
  4. Hệ thống kiểm tra người thao tác phải là người tạo hoặc quản trị viên cộng đồng của yêu cầu.
  5. Hệ thống đánh dấu yêu cầu bị từ chối kèm thời điểm phản hồi; chỉ áp dụng cho yêu cầu còn đang chờ, yêu cầu đã được xử lý trước đó không bị thay đổi.
  6. Hệ thống gửi thông báo cho người xin tham gia rằng yêu cầu của họ đã bị từ chối.
  7. Trả về thông báo từ chối yêu cầu thành công.
- **Entity**: creator/admin cộng đồng, người xin tham gia
- **Model**: models.CommunityJoinRequest, models.Notification
- **Databases**: community_join_requests, user_roles, roles, notifications

### 168. Danh sách thành viên cộng đồng

- **Name**: GetCommunityMembers (communities)
- **Description**:
  1. Người dùng đăng nhập rồi mở danh sách thành viên của một cộng đồng.
  2. Nếu cộng đồng không tồn tại thì báo lỗi không tìm thấy cộng đồng.
  3. Hệ thống lấy toàn bộ thành viên của cộng đồng, xác định vai trò cao nhất của từng người trong cộng đồng và tải thông tin cá nhân (tên hiển thị, ảnh đại diện) theo lô.
  4. Danh sách sắp xếp theo mức ưu tiên vai trò: quản trị cộng đồng lên trước, rồi đến các vai trò thấp hơn.
  5. Mỗi thành viên kèm vai trò, thời điểm tham gia, điểm đóng góp và loại huy hiệu; điểm đóng góp và huy hiệu hiện chưa được hiển thị nên luôn là giá trị rỗng hoặc 0.
  6. Trả về danh sách thành viên; thao tác chỉ đọc, không thay đổi dữ liệu.
- **Entity**: người xem (đã đăng nhập), thành viên cộng đồng
- **Model**: models.GroupMember, models.UserRole, models.Role, models.Profile
- **Databases**: group_members, user_roles, roles, profiles, communities

### 169. Đổi vai trò thành viên

- **Name**: UpdateMemberRole
- **Description**:
  1. Người dùng là quản trị cộng đồng chọn một thành viên rồi gửi yêu cầu đổi vai trò của người đó trong cộng đồng; vai trò hợp lệ chỉ gồm quản trị nhóm, kiểm duyệt viên hoặc thành viên thường.
  2. Nếu dữ liệu gửi lên sai định dạng thì báo lỗi thông tin cộng đồng không hợp lệ; hệ thống kiểm tra cộng đồng phải tồn tại.
  3. Hệ thống kiểm tra người thao tác phải là người tạo hoặc quản trị viên cộng đồng.
  4. Hệ thống áp dụng các ràng buộc: không được tự đổi vai trò của chính mình, người được chọn phải đang là thành viên cộng đồng, không được hạ vai trò của người tạo cộng đồng.
  5. Hệ thống cập nhật vai trò của thành viên trong cộng đồng; nếu không cập nhật được thì báo lỗi đổi vai trò thất bại.
  6. Hệ thống gửi thông báo cho thành viên rằng vai trò của họ trong cộng đồng đã được thay đổi.
  7. Trả về thông báo đổi vai trò thành công.
- **Entity**: creator/admin cộng đồng, thành viên bị đổi vai trò
- **Model**: models.Community, models.UserRole, models.Role, models.Notification
- **Databases**: communities, user_roles, roles, notifications

### 170. Đuổi thành viên khỏi cộng đồng

- **Name**: KickMember
- **Description**:
  1. Người dùng là quản trị cộng đồng chọn một thành viên rồi gửi yêu cầu đuổi người đó, bắt buộc phải kèm lý do từ 3 đến 500 ký tự.
  2. Nếu dữ liệu gửi lên sai định dạng hoặc lý do không hợp lệ (rỗng, quá ngắn, quá dài) thì báo lỗi tương ứng.
  3. Hệ thống kiểm tra cộng đồng phải tồn tại và người thao tác phải là người tạo hoặc quản trị viên cộng đồng.
  4. Hệ thống áp dụng các ràng buộc: không được đuổi chính mình, người bị chọn phải đang là thành viên cộng đồng, không được đuổi người tạo cộng đồng, quản trị viên không được đuổi quản trị viên khác (chỉ người tạo cộng đồng mới làm được).
  5. Trong một thao tác ghi, hệ thống xóa vai trò của thành viên trong cộng đồng và xóa khỏi danh sách thành viên.
  6. Hệ thống gửi thông báo cho người bị đuổi nêu rõ cộng đồng và lý do bị đuổi.
  7. Trả về thông báo đuổi thành viên thành công.
- **Entity**: creator/admin cộng đồng, thành viên bị đuổi
- **Model**: models.Community, models.GroupMember, models.UserRole, models.Role, models.Notification
- **Databases**: communities, group_members, user_roles, roles, notifications

### 171. Rời cộng đồng

- **Name**: LeaveCommunity
- **Description**:
  1. Người dùng đang là thành viên cộng đồng chọn rời khỏi cộng đồng đó; có thể chọn chế độ rời kín đáo.
  2. Nếu request gửi lên sai định dạng thì báo lỗi thông tin cộng đồng không hợp lệ.
  3. Hệ thống kiểm tra cộng đồng phải tồn tại và người dùng phải đang là thành viên; không phải thành viên thì báo lỗi không tìm thấy thành viên.
  4. Nếu người dùng là người tạo cộng đồng thì báo lỗi không được rời cộng đồng, phải chuyển quyền sở hữu cộng đồng cho người khác trước.
  5. Trong một thao tác ghi, hệ thống xóa vai trò của người dùng trong cộng đồng và xóa khỏi danh sách thành viên.
  6. Hệ thống gửi thông báo mọi người biết người dùng đã rời cộng đồng: chế độ kín đáo thì chỉ thông báo cho các quản trị viên, không kín đáo thì thông báo cho toàn bộ thành viên còn lại (trừ chính người rời).
  7. Trả về thông báo rời cộng đồng thành công.
- **Entity**: thành viên rời cộng đồng, admin/member còn lại (nhận notification)
- **Model**: models.Community, models.GroupMember, models.UserRole, models.Notification
- **Databases**: communities, group_members, user_roles, roles, notifications

### 172. Chuyển quyền sở hữu cộng đồng

- **Name**: TransferOwnership (community)
- **Description**:
  1. Người dùng là người tạo cộng đồng chọn một thành viên rồi gửi yêu cầu chuyển quyền sở hữu cộng đồng cho người đó; có thể chọn giữ lại vai trò quản trị cho mình.
  2. Nếu request gửi lên sai định dạng hoặc thiếu người được nhận quyền thì báo lỗi thông tin cộng đồng không hợp lệ; không được tự chuyển quyền cho chính mình và cộng đồng phải tồn tại.
  3. Hệ thống kiểm tra chỉ người tạo cộng đồng mới được chuyển quyền; quản trị viên thường không được phép.
  4. Người được nhận quyền phải đang là thành viên của cộng đồng; không phải thì báo lỗi không tìm thấy thành viên.
  5. Trong một thao tác ghi, hệ thống đổi vai trò người tạo cộng đồng sang người được nhận: người nhận được gán vai trò quản trị cộng đồng và quản trị nhóm chat, người chuyển bị hạ vai trò (nếu chọn không giữ vai trò quản trị thì thành thành viên thường), và cập nhật vai trò quản trị của người nhận trong nhóm chat mặc định của cộng đồng.
  6. Hệ thống gửi thông báo cho người nhận rằng họ đã được chuyển quyền sở hữu cộng đồng.
  7. Trả về thông báo chuyển quyền sở hữu thành công.
- **Entity**: creator (người chuyển), user được nhận quyền, cộng đồng, group chat mặc định
- **Model**: models.Community, models.UserRole, models.Role, models.Chat, models.ChatParticipant, models.Notification
- **Databases**: communities, user_roles, roles, chats, chat_participants, notifications

### 173. Tạo mã mời

- **Name**: CreateInviteCode
- **Description**:
  1. Người dùng là quản trị cộng đồng mở tính năng tạo mã mời rồi gửi yêu cầu, có thể đặt số lượt dùng tối đa (0 nghĩa là không giới hạn) và thời hạn hiệu lực (không đặt thì không hết hạn).
  2. Nếu thiếu mã cộng đồng hợp lệ, request gửi lên sai định dạng hoặc số lượt dùng không hợp lệ thì báo lỗi tương ứng.
  3. Hệ thống kiểm tra người thao tác phải là người tạo hoặc quản trị viên cộng đồng.
  4. Hệ thống sinh mã mời gồm đúng 6 ký tự, dùng bộ ký tự loại bỏ các chữ dễ nhầm lẫn như 0, O, 1, I.
  5. Hệ thống lưu mã mời mới ở trạng thái đang hoạt động, chưa dùng lần nào; lưu thất bại thì báo lỗi tạo mã mời thất bại.
  6. Trả về thông tin mã mời vừa tạo (mã, số lượt dùng tối đa, số lượt đã dùng, thời hạn, trạng thái hoạt động, thời điểm tạo).
- **Entity**: creator/admin cộng đồng, cộng đồng
- **Model**: models.CommunityInviteCode
- **Databases**: community_invite_codes, user_roles, roles

### 174. Danh sách mã mời

- **Name**: ListInviteCodes
- **Description**:
  1. Người dùng là quản trị cộng đồng mở danh sách các mã mời của cộng đồng.
  2. Nếu không có mã cộng đồng hợp lệ thì báo lỗi không tìm thấy cộng đồng.
  3. Hệ thống kiểm tra người thao tác phải là người tạo hoặc quản trị viên cộng đồng; thành viên thường không được xem mã mời.
  4. Hệ thống lấy toàn bộ mã mời của cộng đồng theo thứ tự tạo gần đây nhất lên trước; truy vấn thất bại thì báo lỗi lấy danh sách mã mời thất bại.
  5. Mỗi mã kèm số lượt dùng tối đa, số lượt đã dùng, thời hạn hiệu lực và trạng thái hoạt động.
  6. Trả về danh sách mã mời; thao tác chỉ đọc, không thay đổi dữ liệu.
- **Entity**: creator/admin cộng đồng, cộng đồng
- **Model**: models.CommunityInviteCode
- **Databases**: community_invite_codes, user_roles, roles

### 175. Vô hiệu hóa mã mời

- **Name**: DeactivateInviteCode
- **Description**:
  1. Người dùng là quản trị cộng đồng mở danh sách mã mời rồi chọn vô hiệu hóa một mã.
  2. Nếu không có mã mã mời hợp lệ thì báo lỗi không tìm thấy mã mời.
  3. Hệ thống tra cứu mã mời theo mã; không tìm thấy thì báo lỗi không tìm thấy mã mời.
  4. Hệ thống xác định cộng đồng chứa mã mời rồi kiểm tra người thao tác phải là người tạo hoặc quản trị viên của đúng cộng đồng đó.
  5. Hệ thống đánh dấu mã mời ngừng hoạt động; mã và số lượt đã dùng được giữ nguyên, chỉ chặn việc dùng tiếp.
  6. Trả về thông báo vô hiệu hóa mã mời thành công.
- **Entity**: creator/admin cộng đồng, mã mời
- **Model**: models.CommunityInviteCode
- **Databases**: community_invite_codes, user_roles, roles

### 176. Gửi lời mời tham gia cộng đồng

- **Name**: SendInvitation
- **Description**:
  1. Người dùng là quản trị cộng đồng mở tính năng mời thành viên rồi chọn một người dùng bất kỳ để gửi lời mời tham gia cộng đồng.
  2. Nếu thiếu mã cộng đồng hợp lệ, request gửi lên sai định dạng hoặc thiếu người được mời thì báo lỗi thông tin cộng đồng không hợp lệ.
  3. Hệ thống kiểm tra người gửi lời mời phải là người tạo hoặc quản trị viên cộng đồng.
  4. Hệ thống áp dụng các ràng buộc: không được mời chính mình, người đã là thành viên cộng đồng thì báo đã tham gia, người này đã có lời mời đang chờ phản hồi thì chặn không gửi trùng.
  5. Hệ thống tạo lời mời mới ở trạng thái chờ phản hồi; tạo thất bại thì báo lỗi gửi lời mời thất bại.
  6. Hệ thống gửi thông báo cho người được mời rằng họ nhận được lời mời tham gia cộng đồng.
  7. Trả về thông tin lời mời vừa gửi kèm tên cộng đồng.
- **Entity**: creator/admin cộng đồng (inviter), user được mời (invitee), cộng đồng
- **Model**: models.CommunityInvitation, models.Community, models.Notification
- **Databases**: community_invitations, communities, user_roles, roles, notifications

### 177. Danh sách lời mời của tôi

- **Name**: ListMyInvitations
- **Description**:
  1. Người dùng đăng nhập rồi mở danh sách các lời mời tham gia cộng đồng dành cho mình.
  2. Hệ thống tự lấy danh tính người gọi từ thông tin đăng nhập, không cần truyền thêm tham số.
  3. Hệ thống lấy các lời mời đang chờ phản hồi của người dùng kèm tên cộng đồng, theo thứ tự lời mời nhận gần đây nhất lên trước; truy vấn thất bại thì báo lỗi lấy danh sách lời mời thất bại.
  4. Trả về danh sách lời mời mà người dùng chưa phản hồi; thao tác chỉ đọc, không thay đổi dữ liệu.
- **Entity**: người dùng đang đăng nhập (invitee), cộng đồng mời
- **Model**: models.CommunityInvitation, models.Community
- **Databases**: community_invitations, communities

### 178. Phản hồi lời mời tham gia

- **Name**: RespondInvitation
- **Description**:
  1. Người dùng mở một lời mời tham gia cộng đồng rồi chọn chấp nhận hoặc từ chối.
  2. Nếu request gửi lên sai định dạng hoặc thiếu lựa chọn chấp nhận hoặc từ chối thì báo lỗi thông tin cộng đồng không hợp lệ.
  3. Hệ thống tra cứu lời mời theo mã; không tìm thấy, lời mời không thuộc về người gọi hoặc đã được phản hồi rồi thì báo lỗi tương ứng, không để lộ lời mời của người khác.
  4. Nếu chọn chấp nhận, hệ thống tìm nhóm chat mặc định của cộng đồng (không có thì báo lỗi không tìm thấy nhóm chat cộng đồng), trong một thao tác ghi thêm người dùng vào danh sách thành viên, gán vai trò thành viên và thêm vào nhóm chat mặc định; đồng thời cập nhật lời mời sang trạng thái đã chấp nhận (lỗi cập nhật chỉ ghi lại thông tin, không làm hỏng yêu cầu).
  5. Nếu chọn chấp nhận, hệ thống gửi thông báo cho người gửi lời mời biết người dùng đã tham gia cộng đồng.
  6. Nếu chọn từ chối, hệ thống cập nhật lời mời sang trạng thái đã từ chối kèm thời điểm phản hồi (lỗi cập nhật chỉ ghi lại thông tin).
  7. Trả về thông báo phản hồi lời mời thành công.
- **Entity**: invitee (người được mời), inviter, cộng đồng, group chat mặc định
- **Model**: models.CommunityInvitation, models.GroupMember, models.UserRole, models.Chat, models.ChatParticipant, models.Notification
- **Databases**: community_invitations, group_members, user_roles, roles, chats, chat_participants, notifications

---

**Tổng số mục**: 23 (156–178) — Communities user-facing (list/detail/posts, joined/created, create/update/background, join request, members/roles/kick/leave/transfer, invite codes, invitations).

## B9B — Community Rules + Contribution + Violation Rules (public) + Health/Metrics

Tài liệu hóa **18 endpoint** (số thứ tự 179–196): nội quy cộng đồng, chính sách & điểm đóng góp,
challenge, leaderboard, violation rules (nhóm public) và health/metrics.

## Bối cảnh wiring

- **Route files**: `routes/community_rule.routes.go`, `routes/contribution.routes.go`,
  `routes/violation_rule.routes.go`; `/health` + `/metrics` đăng ký inline trong `cmd/main.go`.
- **`PostService.SetContributionService`** (`services/post.service.go:60`) được gọi ngay sau khi
  `contributionService` khởi tạo → `PostService` cộng điểm `valid_posts` cho member khi đăng bài
  hợp lệ, kéo theo tính lại `contribution_score`.
- **`middlewares.RequireContributionLevel(db, threshold)`** (`middlewares/rbac.middleware.go:62`)
  đọc thẳng `member_contributions.contribution_score` theo `:communityID` param → chặn request
  với 403 nếu điểm đóng góp < ngưỡng (dùng cho các route cần quyền cao trong community).
- **`middlewares.PrometheusMiddleware()`** gắn global cho toàn bộ router (`cmd/main.go:64`).

---

## Nhóm 1: Community Rules (`/api/communities/:communityID/rules`)

### 179. Xem nội quy cộng đồng

- **Name**: GetRules
- **Description**:
  1. Người dùng mở trang xem nội quy của một cộng đồng — đây là thao tác công khai, ai cũng gọi được, kể cả khi chưa đăng nhập.
  2. Nhận mã cộng đồng từ đường dẫn yêu cầu và chuyển yêu cầu xuống tầng xử lý nghiệp vụ.
  3. Truy vấn toàn bộ nội quy của cộng đồng đó, sắp xếp theo danh mục rồi theo thứ tự hiển thị.
  4. Nếu cộng đồng chưa có nội quy nào thì trả về danh sách rỗng, không coi là lỗi.
  5. Trả danh sách nội quy về cho người dùng.
- **Entity**: Người dùng công khai / guest, member community
- **Model**: models.CommunityRule
- **Databases**: community_rules

### 180. Tạo nội quy cộng đồng

- **Name**: CreateRule
- **Description**:
  1. Admin cộng đồng nhập thông tin nội quy mới: danh mục, tiêu đề, nội dung và thứ tự hiển thị.
  2. Yêu cầu đăng nhập bằng token hợp lệ để lấy danh tính người thực hiện.
  3. Kiểm tra quyền: người dùng phải là admin của cộng đồng này, không phải thì từ chối với thông báo không phải admin cộng đồng.
  4. Validate dữ liệu đầu vào: danh mục hợp lệ, tiêu đề và nội dung đúng độ dài, thứ tự hiển thị không âm.
  5. Kiểm tra trùng tiêu đề trong cùng cộng đồng và cùng danh mục (so sánh không phân biệt hoa thường).
  6. Nếu người dùng không chỉ định thứ tự hiển thị thì tự xếp nội quy vào cuối danh mục đó.
  7. Tạo bản ghi nội quy với định danh duy nhất và thời điểm tạo hiện tại, lưu vào database.
  8. Trả thông báo tạo nội quy thành công kèm mã của nội quy vừa tạo.
- **Entity**: Admin của community (COMMUNITY_ADMIN / GROUP_ADMIN), member
- **Model**: models.CommunityRule
- **Databases**: community_rules, user_roles, roles

### 181. Cập nhật nội quy cộng đồng

- **Name**: UpdateRule
- **Description**:
  1. Admin cộng đồng mở nội quy cần sửa và gửi lên các trường muốn thay đổi.
  2. Yêu cầu đăng nhập bằng token hợp lệ để lấy danh tính người thực hiện.
  3. Tìm nội quy theo mã, nếu không có thì báo nội quy không tồn tại.
  4. Kiểm tra quyền trên cộng đồng đang chứa nội quy đó: chỉ admin của đúng cộng đồng đó được sửa.
  5. Cập nhật một phần: chỉ các trường được gửi lên và có giá trị khác rỗng mới bị thay đổi, các trường còn lại giữ nguyên.
  6. Validate từng trường được gửi: tiêu đề, nội dung, danh mục hợp lệ và thứ tự hiển thị không âm.
  7. Ghi thời điểm cập nhật hiện tại rồi lưu thay đổi vào database.
  8. Trả thông báo cập nhật thành công kèm mã nội quy.
- **Entity**: Admin của community
- **Model**: models.CommunityRule
- **Databases**: community_rules, user_roles, roles

### 182. Xoá nội quy cộng đồng

- **Name**: DeleteRule
- **Description**:
  1. Admin cộng đồng chọn nội quy cần xoá.
  2. Yêu cầu đăng nhập bằng token hợp lệ để lấy danh tính người thực hiện.
  3. Tìm nội quy theo mã, nếu không có thì báo nội quy không tồn tại.
  4. Kiểm tra quyền admin trong cộng đồng đang chứa nội quy đó, không phải thì từ chối.
  5. Xoá vĩnh viễn bản ghi nội quy khỏi database.
  6. Trả thông báo xoá nội quy thành công.
- **Entity**: Admin của community
- **Model**: models.CommunityRule
- **Databases**: community_rules, user_roles, roles

---

## Nhóm 2: Contribution Policy + Đóng góp + Challenge (`/api/communities/:communityID/*`)

### 183. Xem chính sách đóng góp của cộng đồng

- **Name**: GetPolicy
- **Description**:
  1. Thành viên đã đăng nhập mở trang xem chính sách đóng góp của một cộng đồng.
  2. Yêu cầu token hợp lệ và truyền đủ mã cộng đồng, thiếu thông tin thì báo lỗi.
  3. Kiểm tra cộng đồng có tồn tại hay không, không có thì báo không tìm thấy.
  4. Kiểm tra người gọi có phải thành viên của cộng đồng đó hay không, không phải thì từ chối.
  5. Đọc chính sách đóng góp của cộng đồng; nếu chưa từng có thì tự tạo chính sách mặc định (trọng số bài viết 10, bình luận 5, phản ứng 2, sự kiện 20; ngưỡng thành viên nổi bật 2500, ngưỡng thăng chức điều hành 5000; tự động thăng chức và huy hiệu đều bật) rồi lưu lại.
  6. Trả về chính sách gồm trọng số của từng hành vi, hai ngưỡng nói trên và trạng thái bật tắt của hai cờ tự động thăng chức cùng huy hiệu.
- **Entity**: Member của community (đã đăng nhập)
- **Model**: models.CommunityPolicy
- **Databases**: community_policies, communities, user_roles, roles

### 184. Cập nhật chính sách đóng góp

- **Name**: UpdatePolicy
- **Description**:
  1. Admin cộng đồng nhập các thông số chính sách muốn thay đổi: trọng số của từng hành vi, hai ngưỡng và hai cờ bật tắt.
  2. Yêu cầu token hợp lệ để lấy danh tính người thực hiện.
  3. Kiểm tra cộng đồng có tồn tại hay không, không có thì báo không tìm thấy.
  4. Kiểm tra người gọi có phải admin của cộng đồng đó hay không, không phải thì từ chối.
  5. Validate dữ liệu: trọng số và ngưỡng không âm, ngưỡng thăng chức điều hành phải hợp lý so với ngưỡng thành viên nổi bật.
  6. Lưu chính sách mới kèm thời điểm cập nhật: đã có thì ghi đè, chưa có thì tạo mới.
  7. Trả thông báo cập nhật thành công; chính sách mới áp dụng ngay cho lần tính điểm kế tiếp, bao gồm gắn huy hiệu và tự động thăng chức điều hành.
- **Entity**: Admin community
- **Model**: models.CommunityPolicy
- **Databases**: community_policies, communities, user_roles, roles

### 185. Bảng xếp hạng đóng góp (leaderboard)

- **Name**: GetLeaderboard
- **Description**:
  1. Người dùng mở trang bảng xếp hạng đóng góp của cộng đồng — đây là thao tác công khai, ai cũng gọi được.
  2. Nhập số trang và kích thước trang (mặc định 1 và 10), giá trị nhỏ hơn 1 thì báo lỗi.
  3. Kiểm tra cộng đồng có tồn tại hay không, không có thì báo không tìm thấy.
  4. Chuẩn hóa phân trang, chặn trang hoặc quy mô vượt giới hạn rồi tính vị trí bắt đầu của trang.
  5. Truy vấn điểm đóng góp của các thành viên, sắp xếp theo điểm giảm dần rồi theo thời điểm tham gia tăng dần, lấy theo trang hiện tại.
  6. Gán thứ hạng tuyệt đối xuyên các trang cho từng người, kèm tên hiển thị, ảnh đại diện, điểm đóng góp và loại huy hiệu.
  7. Trả về trang dữ liệu hiện tại kèm số trang và kích thước trang.
- **Entity**: Mọi người (public), member community
- **Model**: models.MemberContribution (dto.LeaderboardItem)
- **Databases**: member_contributions, profiles, communities

### 186. Xem điểm đóng góp của chính mình

- **Name**: GetMyContribution
- **Description**:
  1. Thành viên đã đăng nhập mở trang xem điểm đóng góp của chính mình trong một cộng đồng.
  2. Yêu cầu token hợp lệ để lấy danh tính người xem.
  3. Kiểm tra cộng đồng có tồn tại và người gọi có phải thành viên của cộng đồng đó hay không, không phải thì từ chối.
  4. Lấy bản ghi đóng góp của người dùng, chưa có thì tạo mới với các bộ đếm bằng không; đồng thời lấy chính sách đóng góp, chưa có thì tạo mặc định.
  5. Tính lại tổng điểm đóng góp từ số bài viết hợp lệ, số bình luận chất lượng, số phản ứng tích cực và số lần tham gia sự kiện, mỗi loại nhân với trọng số tương ứng trong chính sách.
  6. Nếu bật huy hiệu và điểm đạt ngưỡng thì gắn huy hiệu thành viên nổi bật; nếu bật tự động thăng chức và điểm đạt ngưỡng thì thăng người dùng lên vai trò điều hành một cách nguyên tử và tạo thông báo thay đổi vai trò trong cộng đồng.
  7. Ghi thời điểm tính điểm rồi lưu lại, đồng thời gắn tên hiển thị và ảnh đại diện từ hồ sơ người dùng.
  8. Trả về tổng quan điểm đóng góp của người dùng.
- **Entity**: Member đã đăng nhập (chính họ)
- **Model**: models.MemberContribution, models.CommunityPolicy, models.Profile
- **Databases**: member_contributions, community_policies, communities, profiles, user_roles, roles, notifications

### 187. Xem điểm đóng góp của user khác (public)

- **Name**: GetUserContribution
- **Description**:
  1. Người dùng mở trang hồ sơ hoặc cột thành viên để xem điểm đóng góp của người khác — đây là thao tác công khai, ai cũng gọi được.
  2. Nhận mã cộng đồng và mã người dùng cần xem, thiếu thì báo lỗi.
  3. Kiểm tra cộng đồng có tồn tại hay không, không có thì báo không tìm thấy.
  4. Lấy bản ghi đóng góp của người dùng được chỉ định trong cộng đồng đó, chưa có thì tạo mới với các bộ đếm bằng không.
  5. Lấy chính sách đóng góp, chưa có thì tạo mặc định, rồi tính lại điểm, gắn huy hiệu và tự động thăng chức như ở mục xem điểm đóng góp của chính mình, rồi lưu lại.
  6. Gắn tên hiển thị và ảnh đại diện từ hồ sơ của người đó.
  7. Trả về tổng quan điểm đóng góp của người được xem, dùng cho trang hồ sơ và danh sách thành viên công khai.
- **Entity**: Mọi người (public), user được xem
- **Model**: models.MemberContribution, models.CommunityPolicy, models.Profile
- **Databases**: member_contributions, community_policies, communities, profiles, user_roles, roles, notifications

### 188. Danh sách thành viên cộng đồng kèm điểm đóng góp

- **Name**: GetCommunityMembers (contributions)
- **Description**:
  1. Người dùng mở trang danh sách thành viên của cộng đồng kèm điểm đóng góp — đây là thao tác công khai, ai cũng gọi được.
  2. Nhập số trang và kích thước trang (mặc định 1 và 10), giá trị nhỏ hơn 1 thì báo lỗi.
  3. Kiểm tra cộng đồng có tồn tại hay không, không có thì báo không tìm thấy.
  4. Chuẩn hóa phân trang rồi tính vị trí bắt đầu của trang.
  5. Truy vấn danh sách thành viên theo trang, kèm hồ sơ người dùng, vai trò cao nhất trong cộng đồng (admin, điều hành, thành viên) và điểm đóng góp cùng loại huy hiệu của từng người.
  6. Trả về trang dữ liệu gồm thông tin từng thành viên, thời điểm tham gia, vai trò, điểm đóng góp và huy hiệu.
- **Entity**: Mọi người (public), member community
- **Model**: models.MemberContribution (dto.CommunityMemberItem)
- **Databases**: group_members, profiles, user_roles, roles, member_contributions, communities

### 189. Danh sách challenge đang diễn ra

- **Name**: GetActiveChallenges
- **Description**:
  1. Người dùng mở danh sách thử thách của cộng đồng để xem các thẻ thử thách và tiến độ trước khi tham gia — đây là thao tác công khai, ai cũng gọi được.
  2. Chỉ cần mã cộng đồng hợp lệ, không cần đăng nhập.
  3. Kiểm tra cộng đồng có tồn tại hay không, không có thì báo không tìm thấy.
  4. Lấy các thử thách của cộng đồng đang ở trạng thái diễn ra.
  5. Với mỗi thử thách, đếm số người đã tham gia.
  6. Trả về danh sách thử thách kèm số liệu: tiêu đề, mô tả, hashtag, điểm cho mỗi bài viết, ngày bắt đầu và kết thúc, số người tối đa, trạng thái và số người tham gia.
- **Entity**: Mọi người (public), member community
- **Model**: models.CommunityChallenge (dto.ChallengeResponse)
- **Databases**: community_challenges, challenge_participants, communities

### 190. Tạo challenge mới

- **Name**: CreateChallenge
- **Description**:
  1. Admin cộng đồng nhập thông tin thử thách mới: tiêu đề, mô tả, hashtag, điểm cho mỗi bài viết, ngày bắt đầu, ngày kết thúc và số người tối đa.
  2. Yêu cầu token hợp lệ cho thao tác tạo.
  3. Kiểm tra cộng đồng có tồn tại hay không, không có thì báo không tìm thấy.
  4. Kiểm tra người gọi có phải admin của cộng đồng đó hay không, không phải thì từ chối.
  5. Validate dữ liệu: tiêu đề và hashtag hợp lệ, ngày kết thúc phải sau ngày bắt đầu, điểm mỗi bài viết và số người tối đa hợp lệ.
  6. Tạo thử thách với người tạo là admin này, trạng thái là đang diễn ra và thời điểm tạo hiện tại, rồi lưu vào database.
  7. Trả thông báo tạo thử thách thành công.
- **Entity**: Admin community (người tạo), member community (người tham gia sau)
- **Model**: models.CommunityChallenge
- **Databases**: community_challenges, communities, user_roles, roles

### 191. Tham gia challenge

- **Name**: JoinChallenge
- **Description**:
  1. Thành viên cộng đồng bấm tham gia một thử thách.
  2. Yêu cầu token hợp lệ và truyền đủ mã thử thách, thiếu thì báo lỗi.
  3. Tìm thử thách theo mã, nếu không có thì báo thử thách không tồn tại.
  4. Thử thách phải đang ở trạng thái diễn ra, nếu không thì từ chối với thông báo thử thách không hoạt động.
  5. Kiểm tra người dùng có phải thành viên của cộng đồng đang chứa thử thách đó hay không, không phải thì từ chối.
  6. Kiểm tra thời hạn: chưa tới ngày bắt đầu thì báo chưa bắt đầu, đã qua ngày kết thúc thì báo đã kết thúc.
  7. Ghi nhận việc tham gia một cách nguyên tử theo số người tối đa: đã tham gia rồi hoặc đã đủ người thì báo xung đột, không ghi trùng.
  8. Trả thông báo tham gia thành công; về sau khi người dùng đăng bài có hashtag khớp, số bài viết và tổng điểm thưởng của người tham gia được cộng, số lần tham gia sự kiện được cộng dồn vào điểm đóng góp.
- **Entity**: Member community (người tham gia), admin (người tạo challenge)
- **Model**: models.CommunityChallenge, models.ChallengeParticipant
- **Databases**: community_challenges, challenge_participants, communities, user_roles, roles, member_contributions

### 192. Danh sách người tham gia challenge

- **Name**: GetChallengeParticipants
- **Description**:
  1. Người dùng mở danh sách người tham gia của một thử thách — đây là thao tác công khai, ai cũng gọi được.
  2. Nhận mã thử thách, thiếu thì báo lỗi.
  3. Kiểm tra thử thách có tồn tại hay không, không có thì báo thử thách không tồn tại.
  4. Lấy danh sách người tham gia của thử thách kèm hồ sơ của từng người.
  5. Trả về danh sách kèm số bài viết đã đăng, tổng điểm thưởng đã nhận và thời điểm tham gia.
- **Entity**: Mọi người (public), participant challenge
- **Model**: models.ChallengeParticipant
- **Databases**: challenge_participants, community_challenges, profiles

---

## Nhóm 3: Violation Rules (public)

> Đăng ký bởi `routes/violation_rule.routes.go` → `RegisterViolationRuleRoutes`
> (`router.Group("/api/violation-rules")`). Nhóm này là nhóm **dùng cho user thường** (không phải
> nhóm admin `/api/admin/violation-rules` — endpoint admin đã có tài liệu riêng).
> **Lưu ý code**: handler hiện được bảo vệ bằng `middlewares.AuthMiddleware` (yêu cầu Bearer token),
> tức "public" ở đây nghĩa là *công khai cho mọi user đã đăng nhập, không qua phân quyền admin*.

### 193. Danh sách quy tắc vi phạm (đổ dropdown khi report)

- **Name**: ListRules
- **Description**:
  1. Người dùng đã đăng nhập mở danh sách quy tắc vi phạm để chọn lý do khi gửi báo cáo — mọi người dùng đã đăng nhập đều gọi được, không qua phân quyền quản trị.
  2. Nhập bộ lọc loại đối tượng (bài viết, bình luận, người dùng) và từ khóa tìm theo tiêu đề hoặc mô tả.
  3. Chỉ lấy các quy tắc đang được bật, sắp xếp theo thứ tự hiển thị.
  4. Trả về danh sách quy tắc kèm tổng số, mỗi quy tắc gồm tiêu đề, mô tả, phạm vi áp dụng, mức độ nghiêm trọng, thứ tự hiển thị, trạng thái bật tắt và thời điểm tạo.
  5. Quy tắc được chọn sau đó sẽ được kiểm tra lại khi gửi báo cáo: phải tồn tại, đang bật và khớp với loại đối tượng bị báo cáo.
- **Entity**: Người dùng đã đăng nhập (người gửi report), super-admin (người tạo rule)
- **Model**: models.ViolationRule
- **Databases**: violation_rules

### 194. Chi tiết một quy tắc vi phạm

- **Name**: GetRule
- **Description**:
  1. Người dùng mở chi tiết một quy tắc vi phạm để xem mô tả điều khoản trước khi báo cáo, hoặc khi cửa sổ báo cáo hiển thị quy tắc đã chọn.
  2. Nhận mã quy tắc từ đường dẫn, thiếu thì báo lỗi.
  3. Truy vấn quy tắc theo mã đó trong database.
  4. Không tìm thấy thì báo quy tắc vi phạm không tồn tại.
  5. Tìm thấy thì trả về thông tin chi tiết của quy tắc: tiêu đề, mô tả, phạm vi áp dụng, mức độ nghiêm trọng, thứ tự hiển thị, trạng thái bật tắt và các thời điểm tạo cập nhật.
- **Entity**: Người dùng (người gửi report), super-admin (chủ rule)
- **Model**: models.ViolationRule
- **Databases**: violation_rules

---

## Nhóm 4: Health & Metrics

### 195. Kiểm tra sức khỏe server

- **Name**: HealthCheck
- **Description**:
  1. Hệ thống giám sát như nền tảng container, bộ cân bằng tải hoặc đường ống tích diễn gửi yêu cầu kiểm tra sức khỏe tới máy chủ.
  2. Điểm kiểm tra sức khỏe được đăng ký ngay khi khởi động, trước mọi nhóm API nên luôn khả dụng ngay cả khi kết nối database thất bại.
  3. Không thực hiện bất kỳ truy vấn database hay xử lý nghiệp vụ nào, chỉ trả phản hồi ngay.
  4. Trả về trạng thái máy chủ bằng văn bản thuần.
  5. Mọi yêu cầu đều đi qua lớp đo lường nên yêu cầu tới điểm kiểm tra sức khỏe này cũng được đếm vào số liệu.
- **Entity**: Hệ thống (Docker/K8s/LB/monitoring), client nội bộ
- **Model**: — (không model)
- **Databases**: — (không truy vấn DB)

### 196. Prometheus metrics endpoint

- **Name**: MetricsHandler
- **Description**:
  1. Hệ thống giám sát Prometheus định kỳ quét điểm cuối đo lường của máy chủ, điểm cuối này không yêu cầu đăng nhập.
  2. Máy chủ xuất dữ liệu giám sát ra dạng văn bản theo định dạng mà Prometheus đọc được.
  3. Ba loại số liệu được ghi nhận: tổng số yêu cầu theo phương thức, đường dẫn và kết quả xử lý; độ trễ xử lý yêu cầu; số yêu cầu đang được xử lý song song.
  4. Mọi tuyến đường của máy chủ đều được đưa vào đo lường, đường dẫn dùng tên tuyến đã biên dịch chứ không dùng URL thật, nếu không nhận diện được thì ghi nhãn chưa xác định để tránh sinh ra quá nhiều nhãn.
  5. Lượng yêu cầu tới trang kiểm tra sức khỏe và tới chính điểm cuối đo lường này cũng đều được đếm.
  6. Phản hồi trả dữ liệu giám sát dạng văn bản thuần.
- **Entity**: Prometheus server / monitoring stack, DevOps
- **Model**: — (không model)
- **Databases**: — (không truy vấn DB)

---

## Bảng tổng hợp endpoint

| # | Method | Path | Auth |
|---|---|---|---|
| 179 | GET | /api/communities/:communityID/rules | Public |
| 180 | POST | /api/communities/:communityID/rules | Auth (admin community) |
| 181 | PUT | /api/communities/:communityID/rules/:ruleID | Auth (admin community) |
| 182 | DELETE | /api/communities/:communityID/rules/:ruleID | Auth (admin community) |
| 183 | GET | /api/communities/:communityID/policy | Auth (member) |
| 184 | PUT | /api/communities/:communityID/policy | Auth (admin community) |
| 185 | GET | /api/communities/:communityID/contributions/leaderboard | Public |
| 186 | GET | /api/communities/:communityID/contributions/me | Auth (member) |
| 187 | GET | /api/communities/:communityID/contributions/:userID | Public |
| 188 | GET | /api/communities/:communityID/contributions/members | Public |
| 189 | GET | /api/communities/:communityID/challenges | Public |
| 190 | POST | /api/communities/:communityID/challenges | Auth (admin community) |
| 191 | POST | /api/communities/:communityID/challenges/:challengeID/join | Auth (member) |
| 192 | GET | /api/communities/:communityID/challenges/:challengeID/participants | Public |
| 193 | GET | /api/violation-rules | Public (user thường; routes gắn AuthMiddleware) |
| 194 | GET | /api/violation-rules/:id | Public (user thường; routes gắn AuthMiddleware) |
| 195 | GET | /health | Public |
| 196 | GET | /metrics | Public (Prometheus) |

## Bảng database được chạm tới

`community_rules`, `community_policies`, `community_challenges`, `challenge_participants`,
`member_contributions`, `communities`, `profiles`, `users`, `posts`, `comments`, `post_reactions`,
`violation_rules`, `reports`, `notifications` (+ `user_roles`, `roles`, `group_members` phục vụ
kiểm tra quyền/role trong community).
