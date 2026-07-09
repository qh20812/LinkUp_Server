# Admin Group & Community Management — Test Cases

## REST API Endpoints

| Method | Path | Handler | Auth | Description |
|--------|------|---------|------|-------------|
| GET | `/api/admin/groups` | ListGroups | Token (admin+) | Danh sách group chat |
| GET | `/api/admin/groups/:chatID` | GetGroupDetail | Token (admin+) | Chi tiết group chat |
| GET | `/api/admin/groups/:chatID/members` | ListGroupMembers | Token (admin+) | Thành viên group chat |
| GET | `/api/admin/groups/:chatID/logs` | GetGroupModerationLogs | Token (admin+) | Lịch sử kiểm duyệt group |
| POST | `/api/admin/groups/:chatID/hide` | HideGroup | Token (admin+) | Ẩn group chat |
| POST | `/api/admin/groups/:chatID/unhide` | UnhideGroup | Token (admin+) | Bỏ ẩn group chat |
| POST | `/api/admin/groups/:chatID/archive` | ArchiveGroup | Token (admin+) | Đình chỉ group chat |
| POST | `/api/admin/groups/:chatID/warn` | WarnGroup | Token (admin+) | Cảnh báo group chat |
| DELETE | `/api/admin/groups/:chatID` | DeleteGroup | Token (superadmin) | Xoá group chat (chỉ khi 1 member) |
| GET | `/api/admin/communities` | ListCommunities | Token (admin+) | Danh sách cộng đồng |
| GET | `/api/admin/communities/:id` | GetCommunityDetail | Token (admin+) | Chi tiết cộng đồng |
| GET | `/api/admin/communities/:id/members` | ListCommunityMembers | Token (admin+) | Thành viên cộng đồng |
| GET | `/api/admin/communities/:id/logs` | GetCommunityModerationLogs | Token (admin+) | Lịch sử kiểm duyệt cộng đồng |
| POST | `/api/admin/communities/:id/hide` | HideCommunity | Token (admin+) | Ẩn cộng đồng |
| POST | `/api/admin/communities/:id/unhide` | UnhideCommunity | Token (admin+) | Bỏ ẩn cộng đồng |
| POST | `/api/admin/communities/:id/archive` | ArchiveCommunity | Token (admin+) | Đình chỉ cộng đồng |
| POST | `/api/admin/communities/:id/warn` | WarnCommunity | Token (admin+) | Cảnh báo cộng đồng |
| DELETE | `/api/admin/communities/:id` | DeleteCommunity | Token (superadmin) | Xoá cộng đồng (chỉ khi 1 member) |
| POST | `/api/admin/users/:userID/ban` | BanUser | Token (superadmin) | Ban user + auto-transfer ownership |

---

## 1. ListGroups

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| GRP-LST-01 | Lấy danh sách mặc định | GET `/api/admin/groups` với token superadmin | `200` `{groups: [...], total, page:1, page_size:20}`. Sort mặc định `created_at DESC`. | ✅ |
| GRP-LST-02 | Filter theo keyword | GET `/api/admin/groups?keyword=Test` | Chỉ trả group có tên chứa "Test". | ✅ |
| GRP-LST-03 | Filter theo status | GET `/api/admin/groups?status=archived` | Chỉ trả group có status = archived. | ✅ |
| GRP-LST-04 | Phân trang | GET `/api/admin/groups?page=2&page_size=5` | Trả 5 records từ page 2. | ✅ |
| GRP-LST-05 | Admin (không superadmin) có quyền | GET với token admin | `200` thành công. | ✅ |
| GRP-LST-06 | Không có token | GET without Authorization | `401` Unauthorized. | ✅ |
| GRP-LST-07 | Empty list | Không có group nào | `groups: []`, `total: 0`. | ✅ |

## 2. GetGroupDetail

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| GRP-DTL-01 | Thành công | GET `/api/admin/groups/:chatID` | `200` response với đầy đủ `id, name, creator_id, creator_name, type, status, member_count, members[], created_at`. | ✅ |
| GRP-DTL-02 | Group không tồn tại | GET `/api/admin/groups/nonexistent` | `400` `"group chat không tồn tại"`. | ✅ |
| GRP-DTL-03 | chatID là direct chat (type=direct) | GET `/api/admin/groups/:directChatID` | `400` `"group chat không tồn tại"`. | ✅ |
| GRP-DTL-04 | Không có token | GET without Authorization | `401`. | ✅ |

## 3. ListGroupMembers

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| GRP-MEM-01 | Thành công | GET `/api/admin/groups/:chatID/members` | `200` `{members: [{user_id, display_name, avatar_uri, role}]}`. Sort theo `joined_at ASC`. | ✅ |
| GRP-MEM-02 | Group không tồn tại | GET `/api/admin/groups/nonexistent/members` | `400` `"group chat không tồn tại"`. | ✅ |
| GRP-MEM-03 | Không có token | GET without Authorization | `401`. | ✅ |
| GRP-MEM-04 | Empty members (unlikely) | Group mới tạo không có participant | `members: []`. | ✅ |

## 4. GetGroupModerationLogs

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| GRP-LOG-01 | Chưa có log | GET `/api/admin/groups/:chatID/logs` | `200` `{logs: [], total: 0, page: 1, page_size: 20}`. | ✅ |
| GRP-LOG-02 | Có log sau khi moderate | 1. Hide group<br>2. GET logs | `logs` có 1 item với action = "hide". | ✅ |
| GRP-LOG-03 | Nhiều log — phân trang | Thực hiện 3 thao tác moderate, GET `?page=1&page_size=2` | Trả 2 logs, `total: 3`. | ✅ |
| GRP-LOG-04 | Group không tồn tại | GET `/api/admin/groups/nonexistent/logs` | `400` `"group chat không tồn tại"`. | ✅ |

## 5. HideGroup

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| GRP-HIDE-01 | Ẩn thành công | POST `/api/admin/groups/:chatID/hide` với `reason` | `200` `"Ẩn group thành công"`. DB: `chats.status = "hidden"`. Tạo moderation log. | ✅ |
| GRP-HIDE-02 | Ẩn group đã ẩn | POST `/api/admin/groups/:chatID/hide` (status đã là hidden) | `200` (idempotent — thông báo "đã ở trạng thái ẩn"). | ✅ |
| GRP-HIDE-03 | Ẩn group đã archived (đình chỉ) | Archive trước → hide | `400` `"không thể thao tác trên group đã bị đình chỉ"`. | ✅ |
| GRP-HIDE-04 | Thiếu reason | Body: `{}` | `400` validation error. | ✅ |
| GRP-HIDE-05 | Group không tồn tại | POST `/api/admin/groups/nonexistent/hide` | `400` `"group chat không tồn tại"`. | ✅ |
| GRP-HIDE-06 | Không có token | POST without Authorization | `401`. | ✅ |

## 6. UnhideGroup

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| GRP-UNHIDE-01 | Bỏ ẩn thành công | POST `/api/admin/groups/:chatID/unhide` (status=hidden) | `200` `"Bỏ ẩn group thành công"`. DB: `chats.status = "active"`. | ✅ |
| GRP-UNHIDE-02 | Bỏ ẩn group active | POST `/api/admin/groups/:chatID/unhide` (status=active) | `200` (idempotent). | ✅ |
| GRP-UNHIDE-03 | Bỏ ẩn group archived | Archive trước → unhide | `400` `"không thể thao tác trên group đã bị đình chỉ"`. | ✅ |
| GRP-UNHIDE-04 | Group không tồn tại | POST `/api/admin/groups/nonexistent/unhide` | `400`. | ✅ |

## 7. ArchiveGroup

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| GRP-ARCH-01 | Đình chỉ thành công | POST `/api/admin/groups/:chatID/archive` với `reason` | `200` `"Đình chỉ group thành công"`. DB: `chats.status = "archived"`. Tạo moderation log. | ✅ |
| GRP-ARCH-02 | Đình chỉ group đã archived | POST lần 2 | `400` `"đã ở trạng thái đình chỉ"`. | ✅ |
| GRP-ARCH-03 | Thiếu reason | Body: `{}` | `400` validation error. | ✅ |
| GRP-ARCH-04 | Group không tồn tại | POST `/api/admin/groups/nonexistent/archive` | `400`. | ✅ |

## 8. WarnGroup

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| GRP-WARN-01 | Cảnh báo thành công | POST `/api/admin/groups/:chatID/warn` với `reason` + `message` | `200` `"Cảnh báo group thành công"`. Tạo moderation log + notification. | ✅ |
| GRP-WARN-02 | Thiếu reason | Body: `{"message": "test"}` | `400` validation error. | ✅ |
| GRP-WARN-03 | Thiếu message | Body: `{"reason": "test"}` | `400` validation error. | ✅ |
| GRP-WARN-04 | Group không tồn tại | POST `/api/admin/groups/nonexistent/warn` | `400`. | ✅ |

## 9. ListCommunities

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| COM-LST-01 | Lấy danh sách mặc định | GET `/api/admin/communities` với token superadmin | `200` `{communities: [...], total, page:1, page_size:20}`. Sort `created_at DESC`. | ✅ |
| COM-LST-02 | Filter theo keyword | GET `/api/admin/communities?keyword=Test` | Chỉ trả community có tên chứa "Test". | ✅ |
| COM-LST-03 | Filter theo status | GET `/api/admin/communities?status=archived` | Chỉ trả community archived. | ✅ |
| COM-LST-04 | Filter theo privacy | GET `/api/admin/communities?privacy=public` | Chỉ trả community public. | ✅ |
| COM-LST-05 | Kết hợp filter | GET `/api/admin/communities?keyword=Test&status=active&privacy=public` | Kết quả đúng tất cả filter. | ✅ |
| COM-LST-06 | Phân trang | GET `/api/admin/communities?page=1&page_size=5` | Trả 5 records. | ✅ |
| COM-LST-07 | Admin (không superadmin) có quyền | GET với token admin | `200`. | ✅ |
| COM-LST-08 | Không có token | GET without Authorization | `401`. | ✅ |

## 10. GetCommunityDetail

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| COM-DTL-01 | Thành công | GET `/api/admin/communities/:id` | `200` với `id, name, description, creator_id, creator_name, privacy, status, auto_approve, member_count, members[], created_at`. | ✅ |
| COM-DTL-02 | Community không tồn tại | GET `/api/admin/communities/nonexistent` | `400` `"cộng đồng không tồn tại"`. | ✅ |
| COM-DTL-03 | Không có token | GET without Authorization | `401`. | ✅ |

## 11. ListCommunityMembers

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| COM-MEM-01 | Thành công | GET `/api/admin/communities/:id/members` | `200` `{members: [{user_id, display_name, avatar_uri, role}]}`. | ✅ |
| COM-MEM-02 | Community không tồn tại | GET `/api/admin/communities/nonexistent/members` | `400`. | ✅ |
| COM-MEM-03 | Empty members | Community không có thành viên | `members: []`. | ✅ |

## 12. GetCommunityModerationLogs

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| COM-LOG-01 | Chưa có log | GET `/api/admin/communities/:id/logs` | `200` `{logs: [], total: 0}`. | ✅ |
| COM-LOG-02 | Có log sau moderate | 1. Hide community<br>2. GET logs | `logs` có 1 item action="hide". | ✅ |
| COM-LOG-03 | Phân trang | GET `?page=1&page_size=2` sau 3 thao tác | Trả 2 logs, `total: 3`. | ✅ |
| COM-LOG-04 | Community không tồn tại | GET `/api/admin/communities/nonexistent/logs` | `400`. | ✅ |

## 13. HideCommunity

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| COM-HIDE-01 | Ẩn thành công | POST `/api/admin/communities/:id/hide` với `reason` | `200` `"Ẩn cộng đồng thành công"`. DB: `communities.status = "hidden"`. Tạo moderation log. | ✅ |
| COM-HIDE-02 | Ẩn community đã ẩn | POST lần 2 | `200` (idempotent). | ✅ |
| COM-HIDE-03 | Ẩn community archived | Archive → hide | `400` `"không thể thao tác trên cộng đồng đã bị đình chỉ"`. | ✅ |
| COM-HIDE-04 | Thiếu reason | Body: `{}` | `400` validation error. | ✅ |
| COM-HIDE-05 | Community không tồn tại | POST `/api/admin/communities/nonexistent/hide` | `400`. | ✅ |

## 14. UnhideCommunity

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| COM-UNHIDE-01 | Bỏ ẩn thành công | POST `/api/admin/communities/:id/unhide` (status=hidden) | `200` `"Bỏ ẩn cộng đồng thành công"`. DB: status="active". | ✅ |
| COM-UNHIDE-02 | Bỏ ẩn community active | POST khi status=active | `200` (idempotent). | ✅ |
| COM-UNHIDE-03 | Bỏ ẩn community archived | Archive → unhide | `400` `"không thể thao tác trên cộng đồng đã bị đình chỉ"`. | ✅ |
| COM-UNHIDE-04 | Community không tồn tại | POST `/api/admin/communities/nonexistent/unhide` | `400`. | ✅ |

## 15. ArchiveCommunity

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| COM-ARCH-01 | Đình chỉ thành công | POST `/api/admin/communities/:id/archive` với `reason` | `200` `"Đình chỉ cộng đồng thành công"`. DB: status="archived". | ✅ |
| COM-ARCH-02 | Đình chỉ community đã archived | POST lần 2 | `400` `"đã ở trạng thái đình chỉ"`. | ✅ |
| COM-ARCH-03 | Thiếu reason | Body: `{}` | `400`. | ✅ |
| COM-ARCH-04 | Community không tồn tại | POST `/api/admin/communities/nonexistent/archive` | `400`. | ✅ |

## 16. WarnCommunity

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| COM-WARN-01 | Cảnh báo thành công | POST `/api/admin/communities/:id/warn` với `reason` + `message` | `200` `"Cảnh báo cộng đồng thành công"`. Tạo moderation log + notification. | ✅ |
| COM-WARN-02 | Thiếu reason | Body: `{"message": "test"}` | `400`. | ✅ |
| COM-WARN-03 | Thiếu message | Body: `{"reason": "test"}` | `400`. | ✅ |
| COM-WARN-04 | Community không tồn tại | POST `/api/admin/communities/nonexistent/warn` | `400`. | ✅ |

## 17. DeleteGroup (Phase 4)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| GRP-DEL-01 | Xoá thành công (1 member) | DELETE `/api/admin/groups/:chatID` với `reason` khi group chỉ có 1 participant | `200` `"Xóa group chat thành công"`. Xoá messages, participants, chat record. Tạo moderation log. | ✅ |
| GRP-DEL-02 | Lỗi — còn nhiều member | DELETE khi group có ≥ 2 participants | `400` `"không thể xóa group chat còn thành viên khác; hãy chuyển quyền sở hữu trước"`. | ✅ |
| GRP-DEL-03 | Group không tồn tại | DELETE `/api/admin/groups/nonexistent` | `400` `"không tìm thấy chat"`. | ✅ |
| GRP-DEL-04 | Không phải superadmin | DELETE với token admin | `400` `"chỉ có superadmin mới có được phép"`. | ✅ |
| GRP-DEL-05 | Thiếu reason | Body: `{}` | `400` validation error. | ✅ |
| GRP-DEL-06 | chat là direct chat | DELETE với direct chat ID | `400` `"chỉ có thể xóa group chat"`. | ✅ |

## 18. DeleteCommunity (Phase 4)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| COM-DEL-01 | Xoá thành công (1 member) | DELETE `/api/admin/communities/:id` với `reason` khi community chỉ có 1 group_member | `200` `"Xóa cộng đồng thành công"`. Xoá user_roles + group_member. Tạo moderation log. | ✅ |
| COM-DEL-02 | Lỗi — còn nhiều member | DELETE khi community có ≥ 2 group_members | `400` `"không thể xóa cộng đồng còn thành viên khác; hãy chuyển quyền sở hữu trước"`. | ✅ |
| COM-DEL-03 | Community không tồn tại | DELETE `/api/admin/communities/nonexistent` | `400` `"cộng đồng không tồn tại"`. | ✅ |
| COM-DEL-04 | Không phải superadmin | DELETE với token admin | `400` `"chỉ có superadmin mới có được phép"`. | ✅ |
| COM-DEL-05 | Thiếu reason | Body: `{}` | `400`. | ✅ |

## 19. BanUser — Auto-transfer Ownership (Phase 4)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| BAN-AT-01 | Ban user + auto-transfer community | 1. User A tạo community (có member B)<br>2. POST `/api/admin/users/:userA_id/ban` với reason+duration<br>3. Kiểm tra community | `200` ban thành công. DB: `communities.creator_id` ≠ userA_id (đã chuyển sang B hoặc admin khác). | ✅ |
| BAN-AT-02 | Ban user + auto-transfer group chat | User A tạo group chat (có member B) | DB: `chats.creator_id` ≠ userA_id (đã chuyển sang participant khác). | ✅ |
| BAN-AT-03 | Fallback — admin → oldest member → highest contribution | Community có nhiều admin + member | Ưu tiên admin khác, sau đó oldest member, cuối cùng highest contribution. | ✅ |
| BAN-AT-04 | Ban user không sở hữu gì | Ban user B (không phải creator của community/group nào) | `200`. Không ảnh hưởng tới community/group khác. | ✅ |
| BAN-AT-05 | Ban user đã bị ban | POST ban lần 2 | `400` `"người dùng đã bị ban"`. | ✅ |
| BAN-AT-06 | Không phải superadmin | Ban với token admin | `400` `"chỉ có superadmin mới có được phép"`. | ✅ |
| BAN-AT-07 | User không tồn tại | POST `/api/admin/users/nonexistent/ban` | `400`. | ✅ |
| BAN-AT-08 | Thiếu reason | Body: `{"duration": "permanent"}` | `400` validation error. | ✅ |
| BAN-AT-09 | Thiếu duration | Body: `{"reason": "test"}` | `400` validation error. | ✅ |
| BAN-AT-10 | Duration không hợp lệ | Body: `{"reason": "test", "duration": "forever"}` | `400` `"thời hạn ban không hợp lệ"`. | ✅ |
| BAN-AT-11 | Ban temporary (7 ngày) | Body: `{"reason": "test", "duration": "1w"}` | `200`. DB: `bans.expires_at` không null, cách `created_at` ~7 ngày. | ✅ |
| BAN-AT-12 | Ban permanent | Body: `{"reason": "test", "duration": "permanent"}` | `200`. DB: `bans.expires_at IS NULL`. | ✅ |
| BAN-AT-13 | Orphaned community (không ai để transfer) | User A là creator + sole member của community, không có admin khác | Owner không đổi (không crash). | ✅ |
| BAN-AT-14 | Orphaned group chat (không ai để transfer) | User A là creator + sole participant | Owner không đổi (không crash). | ✅ |

## 20. TransferOwnership — Community (Phase 2)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| TRF-COM-01 | Chuyển thành công (keep_admin=false) | POST `/api/communities/:id/transfer-ownership` `{target_user_id, keep_admin: false}` | `200`. DB: `creator_id` đổi. Old creator mất COMMUNITY_ADMIN role, giữ GROUP_MEMBER. New creator có COMMUNITY_ADMIN + GROUP_ADMIN. | ✅ |
| TRF-COM-02 | Chuyển thành công (keep_admin=true) | POST với `keep_admin: true` | Old creator giữ COMMUNITY_ADMIN role. | ✅ |
| TRF-COM-03 | Target không phải member | transfer cho user không trong group_members | `400`. | ✅ |
| TRF-COM-04 | Không phải creator hiện tại | User khác gọi transfer | `400`. | ✅ |

## 21. TransferOwnership — Group Chat (Phase 2)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| TRF-GRP-01 | Chuyển thành công (keep_admin=false) | POST `/api/group-chats/:chatID/transfer-ownership` `{target_user_id, keep_admin: false}` | `200`. DB: `creator_id` đổi. Old creator role → CHAT_MEMBER. New creator → CHAT_ADMIN. Cập nhật `group_chat_settings.last_admin_transfer_at`. | ✅ |
| TRF-GRP-02 | Chuyển thành công (keep_admin=true) | POST với `keep_admin: true` | Old creator giữ CHAT_ADMIN role. | ✅ |
| TRF-GRP-03 | Target không phải participant | transfer cho user ngoài group | `400`. | ✅ |
| TRF-GRP-04 | Không phải creator hiện tại | User khác gọi transfer | `400`. | ✅ |

## 22. TransferAdmin — Group Chat (Phase 2)

| ID | Scenario | Steps | Expected Result | Status |
|----|----------|-------|-----------------|--------|
| TRF-ADM-01 | Admin uỷ quyền thành công | POST `/api/group-chats/:chatID/transfer-admin` `{target_user_id}` (caller là CHAT_ADMIN) | `200`. Target user → CHAT_ADMIN. Cập nhật `last_admin_transfer_at`. | ✅ |
| TRF-ADM-02 | Không phải admin | User CHAT_MEMBER gọi transfer-admin | `400`. | ✅ |
| TRF-ADM-03 | Target không phải participant | `{target_user_id}` không trong group | `400`. | ✅ |
