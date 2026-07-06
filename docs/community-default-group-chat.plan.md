# Community Default Group Chat — Kế hoạch triển khai

## A. Mô tả chức năng

### Flow 1 — Tạo Community
1. User tạo community → hệ thống tự động tạo **Group Chat mặc định**
2. Group chat: tên = tên community, avatar = avatar community
3. Creator = **CHAT_ADMIN** của group chat
4. Community: thêm flag **`auto_approve`** (optional, default = false)
5. Gửi **WebSocket notification** cho creator: "Tạo cộng đồng thành công, group chat mặc định đã sẵn sàng"

### Flow 2 — User tham gia Community
- **Nếu `auto_approve = true`**: user join trực tiếp → auto-add vào group_members + chat_participants + WS notification
- **Nếu `auto_approve = false`** (mặc định): user gửi join request → admin duyệt → add vào group_members + chat_participants + WS notification

### Flow 3 — Notification
- **Tạo community xong**: gửi notification cho creator qua `ws.Hub.SendToUser`
- **Join/Auto-join xong**: gửi notification cho user qua `ws.Hub.SendToUser` với type mới `community_group_chat_added`

---

## B. Database changes

### B1. Model: `models/chat.model.go`
```go
type Chat struct {
    ID            string    `json:"id"`
    Type          ChatType  `json:"type"`
    Name          string    `json:"name"`
    AvatarURI     string    `json:"avatar_uri"`
    EncryptionKey string    `json:"-" gorm:"column:encryption_key"`
    CommunityID   *string   `json:"community_id,omitempty"`    // ← NEW
    CreatedAt     time.Time `json:"created_at"`
}
```

### B2. Model: `models/community.model.go`
```go
type Community struct {
    // ... giữ nguyên ...
    AutoApprove   bool       `json:"auto_approve"`              // ← NEW
    // ...
}
```

### B3. Migration: `cmd/seed/schema/main.go`
```sql
-- Sửa CREATE TABLE chats (dòng 118-125), thêm:
community_id VARCHAR(36) NULL,
FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,

-- Sửa CREATE TABLE communities, thêm:
auto_approve TINYINT(1) NOT NULL DEFAULT 0,
```

### B4. Model: `models/notification.model.go`
```go
const (
    NotificationTypeCommunityGroupChatAdded NotificationType = "community_group_chat_added"  // ← NEW
)
```
Thêm case vào `ParseNotificationType` + `isNotificationEnabled` (gán vào nhóm `FriendRequestEnabled`).

---

## C. Repository layer — `repository/community.repository.go`

### C1. Method mới: `CreateCommunityWithDefaultGroupChat`

```go
func (r *CommunityRepository) CreateCommunityWithDefaultGroupChat(
    ctx context.Context,
    community *models.Community,
    member *models.GroupMember,
    userRoles []models.UserRole,
    chat *models.Chat,
    participants []models.ChatParticipant,
) error
```

Transaction:
1. INSERT `communities`
2. INSERT `group_members` (admin)
3. INSERT `user_roles` × 2 (COMMUNITY_ADMIN + GROUP_ADMIN)
4. INSERT `chats` (type=group, name, avatar, encryption_key, community_id)
5. INSERT `chat_participants` (admin)

### C2. Method mới: `FindDefaultGroupChatByCommunity`

```go
func (r *CommunityRepository) FindDefaultGroupChatByCommunity(ctx context.Context, communityID string) (*models.Chat, error)
```
Query: `WHERE community_id = ? AND type = 'group'`

### C3. Method mới: `AddCommunityMemberAndGroupChat`

Dùng cho cả auto-join và approve join. Transaction:
1. INSERT `group_members`
2. INSERT `user_roles` (GROUP_MEMBER)
3. INSERT `chat_participants` (role=CHAT_MEMBER)

```go
func (r *CommunityRepository) AddCommunityMemberAndGroupChat(
    ctx context.Context,
    communityID, userID, chatID string,
) error
```

### C4. Mở rộng `ApproveJoinRequest`

Đổi signature thành:
```go
func (r *CommunityRepository) ApproveJoinRequest(ctx context.Context, requestID string, chatID *string) error
```

Trong transaction hiện tại, thêm step cuối:
```go
if chatID != nil {
    participant := models.NewChatParticipant(*chatID, req.UserID, models.ChatRoleMember)
    participant.ID = utils.GenerateUUID()
    participant.JoinedAt = now
    if err := tx.Create(&participant).Error; err != nil { ... }
}
```

---

## D. Service layer — `services/community.service.go`

### D1. `CreateCommunity` — mở rộng

```go
func (s *CommunityService) CreateCommunity(
    ctx context.Context,
    creatorID, name, description, avatarURI string,
    autoApprove bool,                    // ← NEW param
) (*models.Community, *models.Chat, error)
```

Thay đổi logic:
1. Build `Community` với `AutoApprove = autoApprove`
2. Build `Chat` (name=community.Name, avatar=community.AvatarURI, CommunityID=&community.ID, EncryptionKey=utils.GenerateEncryptionKey())
3. Build `ChatParticipant` (admin)
4. Gọi `repo.CreateCommunityWithDefaultGroupChat(...)`
5. Gửi notification cho creator:
   ```go
   s.notifService.Create(ctx, creatorID, nil,
       models.NotificationTypeCommunityGroupChatAdded,
       "Tạo cộng đồng thành công! Group chat mặc định đã sẵn sàng",
       nil, nil, nil)
   ```

### D2. `RequestJoin` — mở rộng với auto-approve

```go
func (s *CommunityService) RequestJoin(ctx context.Context, userID, communityID string) (*dto.JoinResult, error)
```

Sau các validation hiện tại, kiểm tra `community.AutoApprove`:

```go
if community.AutoApprove {
    groupChat, err := s.repo.FindDefaultGroupChatByCommunity(ctx, communityID)
    if err != nil { ... }

    if err := s.repo.AddCommunityMemberAndGroupChat(ctx, communityID, userID, groupChat.ID); err != nil {
        return nil, err
    }

    s.notifService.Create(ctx, userID, &community.CreatorID,
        models.NotificationTypeCommunityGroupChatAdded,
        "Bạn đã tham gia cộng đồng và group chat mặc định",
        nil, &communityID, nil)

    return &dto.JoinResult{AutoApproved: true}, nil
}

// Flow cũ: tạo join request
// ...
return &dto.JoinResult{RequestID: joinReq.ID, AutoApproved: false}, nil
```

### D3. `ApproveJoinRequest` — mở rộng

```go
func (s *CommunityService) ApproveJoinRequest(ctx context.Context, adminID, requestID string) error
```

Sau validation:
1. Lấy `groupChat = repo.FindDefaultGroupChatByCommunity(...)`
2. Gọi `repo.ApproveJoinRequest(ctx, requestID, &groupChat.ID)`
3. Gửi notification cho user:
   ```go
   s.notifService.Create(ctx, req.UserID, &adminID,
       models.NotificationTypeCommunityGroupChatAdded,
       "Bạn đã được duyệt vào cộng đồng và thêm vào group chat",
       nil, &req.CommunityID, nil)
   ```

---

## E. Controller layer — `controllers/community.controller.go`

### E1. `CreateCommunity` — update

```go
autoApprove := c.PostForm("auto_approve") == "true"   // ← NEW param

community, groupChat, err := ctrl.communityService.CreateCommunity(
    c.Request.Context(), userID.(string), name, description, avatarURI, autoApprove,
)
if err != nil { ... }

c.JSON(http.StatusOK, gin.H{
    "message":      "Tạo cộng đồng thành công!",
    "community_id": community.ID,
    "auto_approve": community.AutoApprove,
    "default_group_chat": gin.H{
        "id":   groupChat.ID,
        "name": groupChat.Name,
    },
})
```

### E2. `RequestJoin` — update response

```go
result, err := ctrl.communityService.RequestJoin(c.Request.Context(), userID, communityID)
if err != nil { ... }

if result.AutoApproved {
    c.JSON(http.StatusOK, gin.H{
        "message": "Tham gia cộng đồng thành công!",
    })
} else {
    c.JSON(http.StatusOK, gin.H{
        "message":    "Gửi yêu cầu tham gia cộng đồng thành công!",
        "request_id": result.RequestID,
    })
}
```

---

## F. DTO changes — `dto/community.dto.go`

```go
type JoinResult struct {
    RequestID    string `json:"request_id,omitempty"`
    AutoApproved bool   `json:"auto_approved"`
}
```

---

## G. Seed relationships — `cmd/seed/relationships/main.go`

Sau mỗi `INSERT communities` (dòng 58-66), thêm:

```go
// Tạo group chat mặc định
chatID := internal.UUID()
if err := internal.Exec(database,
    `INSERT INTO chats (id, type, name, avatar_uri, encryption_key, community_id, created_at)
     VALUES (?, 'group', ?, ?, ?, ?, ?)`,
    chatID, c.name,
    fmt.Sprintf("https://api.dicebear.com/7.x/identicon/svg?seed=community%d", i),
    "seed-enc-key", c.id, now,
); err != nil { ... }

// Admin participant
if err := internal.Exec(database,
    `INSERT INTO chat_participants (id, chat_id, user_id, role, joined_at)
     VALUES (?, ?, ?, 'CHAT_ADMIN', ?)`,
    internal.UUID(), chatID, state.UserIDs[c.creatorIdx], now,
); err != nil { ... }

// Lưu chatID để dùng khi seed members
state.CommunityChatIDs = append(state.CommunityChatIDs, chatID)
```

Thêm `CommunityChatIDs []string` vào `internal.SeedState`. Khi seed members, thêm chat_participant cho mỗi member:
```go
if err := internal.Exec(database,
    `INSERT INTO chat_participants (id, chat_id, user_id, role, joined_at)
     VALUES (?, ?, ?, 'CHAT_MEMBER', ?)`,
    internal.UUID(), state.CommunityChatIDs[m.communityIdx], userID, now,
); err != nil { ... }
```

---

## H. Danh sách file thay đổi

| # | File | Loại thay đổi |
|---|---|---|
| 1 | `models/chat.model.go` | + field `CommunityID *string` |
| 2 | `models/community.model.go` | + field `AutoApprove bool` |
| 3 | `models/notification.model.go` | + const `CommunityGroupChatAdded` + case parsing + case `isNotificationEnabled` |
| 4 | `dto/community.dto.go` | + struct `JoinResult` |
| 5 | `repository/community.repository.go` | + `CreateCommunityWithDefaultGroupChat`, + `FindDefaultGroupChatByCommunity`, + `AddCommunityMemberAndGroupChat`, sửa signature `ApproveJoinRequest` |
| 6 | `services/community.service.go` | Mở rộng `CreateCommunity`, `RequestJoin`, `ApproveJoinRequest` |
| 7 | `controllers/community.controller.go` | Update `CreateCommunity`, `RequestJoin` response |
| 8 | `cmd/seed/schema/main.go` | + cột `community_id` + FK trong `chats`, + cột `auto_approve` trong `communities` |
| 9 | `cmd/seed/internal/state.go` | + field `CommunityChatIDs []string` |
| 10 | `cmd/seed/relationships/main.go` | Seed chat + participant + member participants |
| 11 | `validations/community.validation.go` | (optional) Validate `auto_approve` nếu cần |

---

## I. Sơ đồ luồng hoàn chỉnh

```
                    TẠO COMMUNITY
                  =================
POST /api/communities (form: name, description, avatar, auto_approve)
  │
  ├─ Controller.CreateCommunity
  │   ├─ Parse form (name, desc, avatar file → upload → avatarURI)
  │   ├─ Parse auto_approve (string → bool)
  │   └─ Call Service.CreateCommunity(creatorID, name, desc, avatarURI, autoApprove)
  │
  ├─ Service.CreateCommunity
  │   ├─ ValidateCreateCommunity(name, desc, avatarURI)
  │   ├─ Check role (không phải admin)
  │   ├─ Check name availability
  │   ├─ Build Community (with AutoApprove)
  │   ├─ Build Chat (type=group, name, avatar, encryption_key, community_id)
  │   ├─ Build ChatParticipant (admin)
  │   ├─ Repo.CreateCommunityWithDefaultGroupChat [1 TRANSACTION]
  │   │   ├─ INSERT communities
  │   │   ├─ INSERT group_members
  │   │   ├─ INSERT user_roles × 2
  │   │   ├─ INSERT chats
  │   │   └─ INSERT chat_participants
  │   ├─ notifService.Create(creator, type=community_group_chat_added)
  │   │   └─ ws.Hub.SendToUser → "notification" event
  │   └─ return (*Community, *Chat, nil)
  │
  └─ Controller → JSON { community_id, auto_approve, default_group_chat: { id, name } }


                USER THAM GIA COMMUNITY (auto_approve=true)
              =============================================
POST /api/communities/:id/join
  │
  ├─ Controller.RequestJoin
  │
  ├─ Service.RequestJoin
  │   ├─ Validate user exists, active
  │   ├─ Check not already member
  │   ├─ Check community.AutoApprove == true
  │   ├─ FindDefaultGroupChatByCommunity
  │   ├─ Repo.AddCommunityMemberAndGroupChat [1 TRANSACTION]
  │   │   ├─ INSERT group_members
  │   │   ├─ INSERT user_roles (GROUP_MEMBER)
  │   │   └─ INSERT chat_participants (CHAT_MEMBER)
  │   ├─ notifService.Create(user, type=community_group_chat_added)
  │   │   └─ ws.Hub.SendToUser → "notification" event
  │   └─ return { AutoApproved: true }
  │
  └─ Controller → JSON { message: "Tham gia cộng đồng thành công!" }


          USER GỬI JOIN REQUEST (auto_approve=false → admin duyệt)
        ========================================================
POST /api/communities/:id/join
  │
  ├─ (giống validate trên)
  ├─ Kiểm tra community.AutoApprove == false
  ├─ Tạo join request (logic cũ)
  └─ notifService.Create(creator, "đã gửi yêu cầu tham gia")

PUT /api/communities/:id/join-requests/:requestID/approve
  │
  ├─ Controller.ApproveJoinRequest
  │
  ├─ Service.ApproveJoinRequest
  │   ├─ Find join request
  │   ├─ Check admin role
  │   ├─ FindDefaultGroupChatByCommunity
  │   ├─ Repo.ApproveJoinRequest(requestID, &chatID) [1 TRANSACTION]
  │   │   ├─ UPDATE join_requests (status=approved)
  │   │   ├─ INSERT group_members
  │   │   ├─ INSERT user_roles (GROUP_MEMBER)
  │   │   └─ INSERT chat_participants (CHAT_MEMBER)   ← NEW
  │   ├─ notifService.Create(user, type=community_group_chat_added)
  │   │   └─ ws.Hub.SendToUser → "notification" event
  │   └─ return nil
  │
  └─ Controller → JSON { message: "Chấp nhận yêu cầu tham gia thành công!" }
```

---

## J. Ghi chú kỹ thuật

| Vấn đề | Giải pháp |
|---|---|
| Circular dependency | Không inject `GroupChatService` vào `CommunityService` — logic tạo chat mặc định đơn giản, gọi trực tiếp `utils.GenerateEncryptionKey()` + repository. Pattern tương tự `PostService.SetContributionService` không cần thiết ở đây |
| Transaction atomicity | Toàn bộ community + chat + participant trong 1 transaction — không lo inconsistent state |
| Edge: community không có group chat `FindDefaultGroupChatByCommunity` trả về `nil` | Xử lý graceful: bỏ qua, không block luồng chính (log warning) |
| Notification preference type mới | Gắn `community_group_chat_added` vào cùng nhóm `FriendRequestEnabled` như các community notification khác |
| Auto_approve mặc định | `false` — giữ nguyên behavior hiện tại, chỉ community nào chủ động bật mới có auto-join |
| Seed encryption key | Seed dùng key giả `"seed-enc-key"` — không ảnh hưởng production |
