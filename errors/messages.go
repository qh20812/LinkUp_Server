package errors

// Messages maps error codes to Vietnamese fallback messages.
// Dynamic messages use {{param}} placeholders.
var Messages = map[string]string{
	// Common
	"common.INVALID_INPUT":    "Dữ liệu đầu vào không hợp lệ",
	"common.UNAUTHORIZED":     "Không có quyền truy cập",
	"common.FORBIDDEN":        "Truy cập bị từ chối",
	"common.NOT_FOUND":        "Không tìm thấy tài nguyên",
	"common.INTERNAL_ERROR":   "Lỗi hệ thống",

	// Auth - Validation
	"auth.DISPLAY_NAME_REQUIRED":    "Tên hiển thị không được để trống",
	"auth.DISPLAY_NAME_TOO_SHORT":   "Tên hiển thị phải có ít nhất {{min}} ký tự",
	"auth.DISPLAY_NAME_TOO_LONG":    "Tên hiển thị không được vượt quá {{max}} ký tự",
	"auth.EMAIL_REQUIRED":           "Email không được để trống",
	"auth.EMAIL_INVALID":            "Định dạng email không hợp lệ",
	"auth.PASSWORD_REQUIRED":        "Mật khẩu không được để trống",
	"auth.USERNAME_REQUIRED":        "Tên người dùng không được để trống",
	"auth.USERNAME_TOO_SHORT":       "Tên người dùng phải có ít nhất {{min}} ký tự",
	"auth.USERNAME_TOO_LONG":        "Tên người dùng không được vượt quá {{max}} ký tự",
	"auth.USERNAME_INVALID":         "Tên người dùng chỉ được chứa chữ cái, số, dấu gạch dưới và dấu chấm",
	"auth.PASSWORD_TOO_SHORT":       "Mật khẩu phải có ít nhất {{min}} ký tự",
	"auth.PASSWORD_TOO_LONG":        "Mật khẩu không được vượt quá {{max}} ký tự",
	"auth.PASSWORD_MISSING_UPPER":   "Mật khẩu phải chứa ít nhất một chữ cái in hoa",
	"auth.PASSWORD_MISSING_LOWER":   "Mật khẩu phải chứa ít nhất một chữ cái in thường",
	"auth.PASSWORD_MISSING_DIGIT":   "Mật khẩu phải chứa ít nhất một chữ số",
	"auth.PASSWORD_MISSING_SPECIAL": "Mật khẩu phải chứa ít nhất một ký tự đặc biệt",
	"auth.PASSWORD_SAME_AS_OLD":     "Mật khẩu mới phải khác mật khẩu cũ",

	// Auth - Business
	"auth.INVALID_CREDENTIALS":      "Email hoặc mật khẩu không hợp lệ",
	"auth.EMAIL_EXISTS":             "Email đã tồn tại",
	"auth.ACCOUNT_LOCKED":           "Tài khoản tạm thời bị khóa do nhập sai nhiều lần, vui lòng thử lại sau {{minutes}} phút",
	"auth.ACCOUNT_INACTIVE":         "Tài khoản chưa được kích hoạt",
	"auth.EMAIL_NOT_VERIFIED":       "Vui lòng xác thực email trước khi đăng nhập",
	"auth.MAINTENANCE_MODE":         "Hệ thống đang bảo trì",
	"auth.REGISTRATION_DISABLED":    "Đăng ký đã bị tắt bởi quản trị viên",
	"auth.LOGIN_ATTEMPTS_REMAINING": "Email hoặc mật khẩu không hợp lệ. Bạn còn {{remaining}} lần thử",

	// Auth - Token
	"auth.INVALID_REFRESH_TOKEN":    "Refresh token không hợp lệ hoặc đã hết hạn",
	"auth.SESSION_EXPIRED":          "Phiên đăng nhập đã hết hạn",
	"auth.INVALID_TOKEN":            "Token không hợp lệ",
	"auth.MISSING_AUTHORIZATION":    "Thiếu header authorization",
	"auth.TOKEN_CLAIMS_INVALID":     "Token claims không hợp lệ",
	"auth.TOKEN_NOT_REFRESH":        "Token không phải refresh token",

	// Auth - Google
	"auth.INVALID_GOOGLE_TOKEN":    "Xác thực Google thất bại",
	"auth.GOOGLE_NOT_CONFIGURED":   "Đăng nhập Google chưa được cấu hình",
	"auth.GOOGLE_ACCOUNT_MISMATCH": "Tài khoản Google không khớp với tài khoản hiện tại",

	// Auth - User
	"auth.USER_NOT_FOUND":            "Không tìm thấy người dùng",
	"auth.ACCOUNT_BANNED":            "Tài khoản đang bị ban",
	"auth.ACCOUNT_BANNED_WITH_EXPIRY": "Tài khoản đang bị ban đến {{expiry}}",
	"auth.ACCOUNT_BANNED_PERMANENT":  "Tài khoản bị ban vô thời hạn",

	// Auth - Misc
	"auth.LOGOUT_FAILED":      "Đăng xuất thất bại",
	"auth.CHANGE_PASSWORD_FAILED": "Đổi mật khẩu thất bại",
	"auth.INVALID_ID_TOKEN":   "ID token không được để trống",

	// Post - Validation
	"post.TITLE_REQUIRED":   "Tiêu đề bài viết là bắt buộc",
	"post.TITLE_TOO_SHORT":  "Tiêu đề bài viết phải có ít nhất 5 ký tự",
	"post.TITLE_TOO_LONG":   "Tiêu đề bài viết không được vượt quá 150 ký tự",
	"post.CONTENT_REQUIRED": "Nội dung bài viết là bắt buộc",
	"post.CONTENT_TOO_LONG": "Nội dung bài viết không được vượt quá 5000 ký tự",
	"post.POST_ID_REQUIRED": "Mã bài viết là bắt buộc",
	"post.INVALID_STATUS":   "Trạng thái '{{status}}' không hợp lệ. Chỉ chấp nhận: public, private, friend, hidden",
	"post.INVALID_FORMAT":   "Định dạng dữ liệu gửi lên không hợp lệ",

	// Post - Business
	"post.POST_NOT_FOUND":            "Bài viết không tồn tại",
	"post.POST_HIDDEN_OR_PRIVATE":    "Bài viết đã bị ẩn hoặc ở chế độ riêng tư",
	"post.POST_NOT_ACCESSIBLE":       "Bài viết không tồn tại hoặc không thể truy cập",
	"post.POST_NOT_SHAREABLE":        "Bài viết không tồn tại hoặc không cho phép chia sẻ",
	"post.CANNOT_SHARE_OWN":          "Bạn không thể chia sẻ bài viết của chính mình",
	"post.ALREADY_SHARED":            "Bạn đã chia sẻ bài viết này rồi",
	"post.CANNOT_SAVE_OWN":           "Bạn không thể lưu bài viết của chính mình",
	"post.CANNOT_DELETE_OTHERS":      "Bạn không có quyền xóa bài viết của người khác",
	"post.CONTRIBUTION_SERVICE_NOT_INIT": "Dịch vụ contribution chưa được khởi tạo",

	// Post - Comment
	"post.COMMENT_CONTENT_REQUIRED": "Nội dung bình luận không được để trống",
	"post.COMMENT_CONTENT_TOO_LONG": "Nội dung bình luận không được vượt quá 1000 ký tự",
	"post.COMMENT_NOT_FOUND":        "Bình luận cấp trên không tồn tại hoặc đã bị xóa",
	"post.COMMENT_WRONG_POST":       "Bình luận gốc không thuộc bài viết này",
	"post.CANNOT_COMMENT_HIDDEN_PRIVATE": "Không thể bình luận vào bài viết đã bị ẩn hoặc ở chế độ riêng tư",

	// Post - Reaction
	"post.EMOJI_REQUIRED": "Emoji_id là bắt buộc",
	"post.EMOJI_NOT_FOUND": "Emoji không tồn tại",

	// Post - Pagination
	"post.INVALID_PAGE_SIZE": "page_size phải từ 1 đến 100",

	// Community - Validation
	"community.NAME_REQUIRED":      "Tên cộng đồng không được để trống",
	"community.NAME_TOO_SHORT":     "Tên cộng đồng phải có ít nhất 3 ký tự",
	"community.NAME_TOO_LONG":      "Tên cộng đồng không được vượt quá 100 ký tự",
	"community.DESC_TOO_LONG":      "Mô tả cộng đồng không được vượt quá 500 ký tự",
	"community.AVATAR_INVALID":     "Avatar URI không hợp lệ",
	"community.BACKGROUND_INVALID": "Background URI không hợp lệ",
	"community.INVALID_FORMAT":     "Dữ liệu đầu vào không hợp lệ",

	// Community - Business
	"community.NOT_FOUND":          "Cộng đồng không tồn tại",
	"community.NAME_EXISTS":        "Tên cộng đồng đã tồn tại",
	"community.ADMIN_CANNOT_CREATE": "Quản trị viên không được tạo cộng đồng",
	"community.ENCRYPTION_KEY_FAILED": "Lỗi khi tạo mã khóa mã hóa",
	"community.BACKGROUND_UPLOAD_FAILED": "Tải ảnh background thất bại",
	"community.BACKGROUND_REJECTED":    "Ảnh background vi phạm tiêu chuẩn cộng đồng",
	"community.BACKGROUND_UPDATE_FAILED": "Cập nhật background cộng đồng thất bại",
	"community.UPDATE_FAILED":           "Cập nhật cộng đồng thất bại",
	"community.UPDATE_NO_FIELDS":        "Không có trường nào để cập nhật",
	"community.IMAGE_READ_FAILED":      "Không thể đọc file ảnh",
	"community.JOIN_REQUEST_FAILED":    "Gửi yêu cầu tham gia thất bại",
	"community.INVITE_CODE_REQUIRED":   "Mã mời là bắt buộc",
	"community.INVITATION_REQUIRED":    "Lời mời là bắt buộc",
	"community.GROUP_CHAT_NOT_FOUND":   "Không tìm thấy group chat mặc định",

	// Community - Membership
	"community.ALREADY_MEMBER":         "Bạn đã là thành viên của cộng đồng này",
	"community.JOIN_REQUEST_PENDING":   "Bạn đã có yêu cầu tham gia đang chờ xử lý",
	"community.JOIN_REQUEST_NOT_FOUND": "Yêu cầu tham gia không tồn tại",
	"community.JOIN_REQUEST_HANDLED":   "Yêu cầu tham gia đã được xử lý",
	"community.NOT_ADMIN":              "Bạn không phải quản trị viên của cộng đồng này",
	"community.NOT_MEMBER":             "Bạn không phải thành viên của cộng đồng này",
	"community.MEMBER_NOT_FOUND":       "Thành viên không tồn tại trong cộng đồng",
	"community.INVALID_ROLE":           "Vai trò không hợp lệ",
	"community.CANNOT_CHANGE_OWN_ROLE": "Không thể thay đổi vai trò của chính mình",
	"community.CANNOT_TARGET_ADMIN":    "Không thể thay đổi vai trò của quản trị viên",
	"community.CREATOR_CANNOT_LEAVE":   "Người tạo cộng đồng không thể rời đi, vui lòng chuyển quyền trước",
	"community.CANNOT_KICK_CREATOR":    "Không thể đuổi người tạo cộng đồng",
	"community.CANNOT_KICK_ADMIN":      "Chỉ người tạo cộng đồng mới có quyền đuổi quản trị viên",
	"community.KICK_REASON_REQUIRED":   "Lý do là bắt buộc",
	"community.KICK_REASON_TOO_SHORT":  "Lý do phải có ít nhất 3 ký tự",
	"community.KICK_REASON_TOO_LONG":   "Lý do không được vượt quá 500 ký tự",
	"community.KICK_FAILED":            "Đuổi thành viên thất bại",
	"community.ROLE_UPDATE_FAILED":     "Cập nhật vai trò thất bại",
	"community.MEMBER_CHECK_FAILED":    "Lỗi khi kiểm tra thông tin thành viên",
	"community.TRANSFER_TO_SELF":       "Không thể chuyển quyền sở hữu cho chính mình",
	"community.ONLY_CREATOR_CAN_TRANSFER": "Chỉ người tạo cộng đồng mới có thể chuyển quyền sở hữu",
	"community.TRANSFER_FAILED":        "Chuyển quyền sở hữu thất bại",

	// Community - Invite code
	"community.INVITE_CODE_NOT_FOUND":    "Mã mời không tồn tại",
	"community.INVITE_CODE_EXPIRED":      "Mã mời đã hết hạn",
	"community.INVITE_CODE_INACTIVE":     "Mã mời đã bị vô hiệu hóa",
	"community.INVITE_CODE_MAX_USES":     "Mã mời đã đạt số lần sử dụng tối đa",
	"community.INVITE_CODE_CREATE_FAILED": "Tạo mã mời thất bại",
	"community.INVITE_CODE_SAVE_FAILED":  "Lưu mã mời thất bại",
	"community.INVITE_CODE_DEACTIVATE_FAILED": "Vô hiệu hóa mã mời thất bại",
	"community.INVITE_CODE_LIST_FAILED":  "Lấy danh sách mã mời thất bại",

	// Community - Invitation
	"community.INVITATION_NOT_FOUND":    "Lời mời không tồn tại",
	"community.INVITATION_HANDLED":      "Lời mời đã được xử lý",
	"community.CANNOT_INVITE_SELF":      "Không thể mời chính mình",
	"community.INVITATION_SEND_FAILED":  "Gửi lời mời thất bại",
	"community.INVITATION_LIST_FAILED":  "Lấy danh sách lời mời thất bại",

	// Community - Rule
	"community.RULE_CATEGORY_INVALID": "Danh mục nội quy không hợp lệ",
	"community.RULE_TITLE_REQUIRED":   "Tiêu đề nội quy không được để trống",
	"community.RULE_TITLE_TOO_SHORT":  "Tiêu đề nội quy phải có ít nhất 5 ký tự",
	"community.RULE_TITLE_TOO_LONG":   "Tiêu đề nội quy không được vượt quá 255 ký tự",
	"community.RULE_CONTENT_TOO_LONG": "Nội dung nội quy không được vượt quá 2000 ký tự",
	"community.RULE_POSITION_NEGATIVE": "Vị trí không được âm",
	"community.RULE_TITLE_DUPLICATE":  "Tiêu đề nội quy đã tồn tại trong danh mục này",
	"community.RULE_NOT_FOUND":        "Nội quy không tồn tại",
	"community.RULE_CHECK_FAILED":     "Lỗi khi kiểm tra trùng lặp nội quy",
	"community.RULE_POSITION_FAILED":  "Lỗi khi xác định vị trí nội quy",

	// ── Group Chat ──────────────────────────────────────────────
	"group_chat.INVALID_NAME":         "Tên nhóm chat phải từ 3 đến 50 ký tự",
	"group_chat.ENCRYPTION_KEY_FAILED": "Không thể khởi tạo khóa bảo mật cho nhóm",
	"group_chat.NOT_MEMBER":           "Bạn không phải là thành viên của nhóm này",
	"group_chat.INVALID_LEAVE_MODE":   "leave_mode không hợp lệ",
	"group_chat.INVALID_HISTORY_MODE": "history_mode không hợp lệ",
	"group_chat.SELF_BAN":             "Bạn không thể tự chặn chính mình",
	"group_chat.ADMIN_ONLY":           "Chỉ quản trị viên mới có quyền chặn thành viên",
	"group_chat.ALREADY_BANNED":       "Người dùng này đã bị chặn từ trước",
	"group_chat.SELF_INVITE":          "Không thể tự mời chính mình",
	"group_chat.NOT_ADMIN_INVITE":     "Admin đã tắt quyền thêm thành viên; chỉ có admin mới có thể mời",
	"group_chat.ALREADY_MEMBER":       "Người dùng này đã là thành viên của nhóm",
	"group_chat.PENDING_REQUEST":      "Đã có lời mời đang chờ phản hồi cho người dùng này",
	"group_chat.ADMIN_ONLY_ADD":       "Chỉ admin mới có quyền thêm thành viên",
	"group_chat.ADMIN_ONLY_NAME":      "Chỉ admin mới có quyền đổi tên nhóm",
	"group_chat.INVALID_GROUP_NAME":   "Tên nhóm phải từ 3 đến 50 ký tự",
	"group_chat.ADMIN_ONLY_AVATAR":    "Chỉ admin mới có quyền đổi avatar nhóm",
	"group_chat.ADMIN_ONLY_CONFIG":    "Chỉ admin mới có quyền cấu hình quyền thêm thành viên",
	"group_chat.ADMIN_ONLY_UPDATE":    "Chỉ admin mới có quyền cập nhật cài đặt toàn nhóm",
	"group_chat.TRANSFER_TO_SELF":     "Không thể chuyển quyền cho chính mình",
	"group_chat.ONLY_CREATOR_TRANSFER": "Chỉ người tạo nhóm mới có thể chuyển quyền sở hữu",
	"group_chat.NOT_GROUP_TRANSFER":   "Chỉ hỗ trợ chuyển quyền sở hữu cho nhóm chat",
	"group_chat.TARGET_NOT_MEMBER":    "Người nhận phải là thành viên của nhóm",
	"group_chat.ADMIN_COOLDOWN":       "Quyền admin chỉ có thể chuyển 1 lần mỗi tháng",
	"group_chat.INVALID_MUTE_REASON":  "Lý do tắt tiếng không hợp lệ",
	"group_chat.INVALID_MUTE_DURATION": "Thời lượng tắt tiếng không hợp lệ",
	"group_chat.SELF_MUTE":            "Không thể mute chính mình",
	"group_chat.NOT_MEMBER_MUTE":      "Người dùng không phải thành viên của nhóm",
	"group_chat.REQUEST_NOT_OWN":      "Bạn không phải người được mời",
	"group_chat.REQUEST_ALREADY_HANDLED": "Yêu cầu này đã được xử lý",
	"group_chat.NOT_GROUP_CHAT":       "Chat này không phải nhóm chat",
	"group_chat.BANNED":               "Bạn đã bị chặn khỏi nhóm này",
	"group_chat.MUTED":                "Bạn đã bị tắt tiếng trong nhóm này",
	"group_chat.EMOJI_NOT_FOUND":      "Emoji không tồn tại",
	"group_chat.MEDIA_NOT_FOUND":      "Media không tồn tại",
	"group_chat.MEDIA_NOT_YOURS":      "Media không thuộc về bạn",
	"group_chat.REPLY_NOT_FOUND":      "Tin nhắn gốc không tồn tại",
	"group_chat.REPLY_WRONG_CHAT":     "Tin nhắn gốc không thuộc phòng chat này",
	"group_chat.ENCRYPTION_KEY_NOT_FOUND": "Lấy khóa mã hóa thất bại",

	// ── Chat (DM) ──────────────────────────────────────────────
	"chat.CONTENT_REQUIRED":   "Nội dung tin nhắn, emoji hoặc media là bắt buộc",
	"chat.CONTENT_TOO_LONG":   "Nội dung tin nhắn không được vượt quá 2000 ký tự",
	"chat.SEARCH_EMPTY":       "Từ khóa tìm kiếm là bắt buộc",
	"chat.DELETE_MODE_INVALID": "Chế độ xóa phải là 'all' hoặc 'me'",
	"chat.NOT_PARTICIPANT":    "Bạn không có quyền tham gia chat này",
	"chat.NOT_GROUP_CHAT":     "Đây không phải nhóm chat",
	"chat.SELF_CHAT":          "Không thể tạo chat trực tiếp với chính mình",
	"chat.ALREADY_FRIENDS":    "Đã là bạn, vui lòng mở chat trực tiếp",
	"chat.INVITE_PENDING":     "Lời mời chat đang chờ phản hồi",
	"chat.INVITE_ACCEPTED":    "Đã có lời mời được chấp nhận, không thể gửi",
	"chat.NOT_RECIPIENT":      "Bạn không phải người nhận lời mời này",
	"chat.NOT_SENDER":         "Bạn không phải người gửi tin nhắn này",
	"chat.ALREADY_DELETED":    "Tin nhắn đã bị thu hồi hoặc xóa",
	"chat.MESSAGE_NOT_FOUND":  "Tin nhắn không tồn tại",
	"chat.ACCESS_DENIED":      "Bạn không có quyền truy cập tin nhắn này",

	// ── Friend ─────────────────────────────────────────────────
	"friend.TARGET_REQUIRED":         "Người dùng mục tiêu là bắt buộc",
	"friend.SELF_REQUEST":            "Không thể gửi lời mời kết bạn cho chính mình",
	"friend.USER_NOT_FOUND":          "Người dùng không tồn tại",
	"friend.USER_INACTIVE":           "Không thể gửi lời mời kết bạn đến người dùng này",
	"friend.ADMIN_RESTRICTED":        "Không thể gửi lời mời kết bạn đến admin hoặc super admin",
	"friend.REQUEST_HANDLED":         "Không thể thực hiện hành động trên lời mời đã xử lý",
	"friend.NOT_FOUND":              "Lời mời kết bạn không tồn tại",
	"friend.NOT_AUTHORIZED":          "Bạn không có quyền thực hiện hành động này trên lời mời",
	"friend.ALREADY_PROCESSED":       "Lời mời đã được xử lý trước đó",
	"friend.NOT_FRIENDS":             "Hai người không phải là bạn bè",
	"friend.SELF_UNFRIEND":           "Không thể tự hủy kết bạn với chính mình",
	"friend.TARGET_REQUIRED_UNFRIEND": "userID là bắt buộc",

	// ── Follow ─────────────────────────────────────────────────
	"follow.SELF":              "Không thể follow chính mình",
	"follow.USER_NOT_FOUND":    "Người dùng không tồn tại",
	"follow.USER_INACTIVE":     "Không thể follow người dùng này",
	"follow.SUPERADMIN_RESTRICTED": "Không thể follow superadmin",
	"follow.ADMIN_RESTRICTED":  "Không thể follow admin",

	// ── Block ──────────────────────────────────────────────────
	"block.TARGET_REQUIRED":    "target_user_id là bắt buộc",
	"block.SELF":              "Không thể chặn chính mình",
	"block.ADMIN_RESTRICTED":  "Không thể chặn quản trị viên hoặc siêu quản trị viên",

	// ── Media ──────────────────────────────────────────────────
	"media.FILE_REQUIRED":        "File là bắt buộc",
	"media.FILE_TYPE_NOT_ALLOWED": "Định dạng file không được hỗ trợ",
	"media.FILE_TOO_LARGE":       "File vượt quá giới hạn kích thước",
	"media.INSUFFICIENT_STORAGE": "Dung lượng lưu trữ không đủ",
	"media.STORAGE_QUOTA_EXCEEDED": "Dung lượng lưu trữ đã đầy, vui lòng mua thêm dung lượng",
	"media.NOT_FOUND":            "Media không tồn tại",
	"media.FORBIDDEN":            "Bạn không có quyền xóa media này",
	"media.IMAGE_DECODE":         "Không thể đọc thông tin ảnh, file có thể bị hỏng",
	"media.IMAGE_TOO_SMALL":      "Ảnh quá nhỏ",
	"media.IMAGE_TOO_LARGE":      "Ảnh quá lớn",
	"media.INVALID_ASPECT_RATIO": "Tỉ lệ khung hình không hợp lệ",
	"media.REJECTED":             "Nội dung không phù hợp và đã bị từ chối",

	// ── Contribution ───────────────────────────────────────────
	"contribution.POST_WEIGHT_INVALID":      "Trọng số bài viết phải từ 0 đến 100",
	"contribution.COMMENT_WEIGHT_INVALID":   "Trọng số bình luận phải từ 0 đến 100",
	"contribution.REACTION_WEIGHT_INVALID":  "Trọng số phản hồi phải từ 0 đến 100",
	"contribution.EVENT_WEIGHT_INVALID":     "Trọng số sự kiện phải từ 0 đến 100",
	"contribution.THRESHOLD_INVALID":        "Ngưỡng điểm phải lớn hơn 0",
	"contribution.THRESHOLD_ORDER_INVALID":  "Ngưỡng Moderator phải lớn hơn ngưỡng Top Contributor",
	"contribution.HASHTAG_REQUIRED":         "Hashtag là bắt buộc",
	"contribution.HASHTAG_INVALID_FORMAT":   "Hashtag phải bắt đầu bằng #",
	"contribution.END_DATE_BEFORE_START":    "Ngày kết thúc phải sau ngày bắt đầu",
	"contribution.START_DATE_IN_PAST":       "Ngày bắt đầu phải ở tương lai",
	"contribution.TITLE_REQUIRED":           "Tên challenge là bắt buộc",
	"contribution.TITLE_TOO_SHORT":          "Tên challenge phải có ít nhất 5 ký tự",
	"contribution.TITLE_TOO_LONG":           "Tên challenge không được vượt quá 255 ký tự",
	"contribution.DESC_TOO_LONG":            "Mô tả challenge không được vượt quá 2000 ký tự",
	"contribution.DATE_FORMAT_INVALID":      "Định dạng ngày không hợp lệ, cần dùng RFC3339",
	"contribution.CHALLENGE_NOT_FOUND":      "Challenge không tồn tại",
	"contribution.CHALLENGE_INACTIVE":       "Challenge không còn hoạt động",
	"contribution.CHALLENGE_NOT_STARTED":    "Challenge chưa bắt đầu",
	"contribution.CHALLENGE_ENDED":          "Challenge đã kết thúc",
	"contribution.ALREADY_JOINED":           "Bạn đã tham gia challenge này",
	"contribution.PARTICIPANT_LIMIT_HIT":    "Challenge đã đủ số lượng người tham gia",

	// ── Call ───────────────────────────────────────────────────
	"call.BUSY":           "Người dùng đang bận",
	"call.NOT_FOUND":      "Cuộc gọi không tồn tại",
	"call.NOT_WAITING":    "Cuộc gọi không ở trạng thái chờ",
	"call.SELF_CALL":      "Không thể gọi cho chính mình",
	"call.NOT_FRIEND":     "Chỉ có thể gọi cho bạn bè",
	"call.NOT_PARTICIPANT": "Không phải người tham gia cuộc gọi",
	"call.ALREADY_ENDED":  "Cuộc gọi đã kết thúc",
	"call.NOT_VIDEO":      "Cuộc gọi không phải video call",
	"call.NOT_CONNECTED":  "Cuộc gọi không ở trạng thái kết nối",

	// ── AI Moderation ──────────────────────────────────────────
	"moderation.CLOUDINARY_NOT_INITIALIZED": "Cloudinary chưa được khởi tạo",
	"moderation.UPLOAD_FAILED":             "Tải lên + kiểm duyệt thất bại",

	// ── Profile ────────────────────────────────────────────────
	"profile.PROFILE_NOT_FOUND": "Không tìm thấy hồ sơ",
	"profile.PRIVATE_PROFILE":   "Hồ sơ này ở chế độ riêng tư",
	"profile.PHONE_EXISTS":      "Số điện thoại đã tồn tại",

	// ── Password Reset ─────────────────────────────────────────
	"password_reset.TOKEN_NOT_FOUND": "Token không hợp lệ",
	"password_reset.TOKEN_EXPIRED":   "Token đã hết hạn",
	"password_reset.TOKEN_USED":      "Token đã được sử dụng",
	"password_reset.PASSWORD_TOO_SHORT": "Mật khẩu phải có ít nhất {{min}} ký tự",
	"password_reset.PASSWORD_TOO_LONG":  "Mật khẩu không được vượt quá 50 ký tự",

	// ── Email Verification ─────────────────────────────────────
	"email_verification.TOKEN_NOT_FOUND": "Token xác thực không hợp lệ",
	"email_verification.TOKEN_EXPIRED":   "Token xác thực đã hết hạn",
	"email_verification.TOKEN_USED":      "Token xác thực đã được sử dụng",
	"email_verification.ALREADY_VERIFIED": "Email đã được xác thực trước đó",
	"email_verification.RATE_LIMITED":    "Vui lòng đợi {{seconds}} giây trước khi yêu cầu gửi lại email xác thực",

	// ── User Settings ──────────────────────────────────────────
	"user_settings.INVALID_THEME":      "Chủ đề không hợp lệ",
	"user_settings.INVALID_LANGUAGE":   "Ngôn ngữ không hợp lệ",
	"user_settings.WRONG_PASSWORD":     "Mật khẩu hiện tại không đúng",
	"user_settings.SESSION_NOT_FOUND":  "Phiên đăng nhập không tồn tại",

	// ── Story ──────────────────────────────────────────────────
	"story.INVALID_FORMAT":           "Định dạng file không hỗ trợ, chỉ nhận ảnh (jpg, png) hoặc video (mp4, mov)",
	"story.CONTENT_REQUIRED":         "Story phải có hình ảnh, video hoặc nội dung chữ (caption)",
	"story.NOT_FOUND":                "Không tìm thấy bản tin (story) này",
	"story.INTERACT_NOT_FOUND":       "Không tìm thấy bản tin để tương tác",
	"story.EMOJI_ID_REQUIRED":        "Emoji_id không được để trống khi thực hiện thả cảm xúc",
	"story.REACT_LIMIT_REACHED":      "Bạn đã đạt giới hạn tối đa 5 lần biểu cảm cho story này",
	"story.EMOJI_NOT_FOUND":          "Mã hiệu ứng emoji không tồn tại trong hệ thống",
	"story.REPLY_EMPTY":              "Nội dung tin nhắn phản hồi không thể để trống",
	"story.ANALYTICS_FORBIDDEN":      "Bạn không có quyền truy cập dữ liệu phân tích của story này",

	// ── Search ─────────────────────────────────────────────────
	"search.KEYWORD_TOO_SHORT": "Từ khóa tìm kiếm phải có ít nhất 2 ký tự",
	"search.TYPE_INVALID":      "Loại tìm kiếm phải là 'all', 'users', 'posts', 'hashtags' hoặc 'communities'",

	// ── Ad ─────────────────────────────────────────────────────
	"ad.NO_SUBSCRIPTION":            "Bạn cần đăng ký gói quảng cáo trước khi tạo chiến dịch",
	"ad.SLOTS_EXHAUSTED":            "Bạn đã dùng hết {{used}}/{{max}} slot. Vui lòng nâng cấp gói",
	"ad.FORMAT_NOT_SUPPORTED_VIDEO": "Gói hiện tại của bạn không hỗ trợ quảng cáo định dạng Video",
	"ad.FORMAT_NOT_SUPPORTED_CAROUSEL": "Gói hiện tại của bạn không hỗ trợ quảng cáo định dạng Carousel",
	"ad.NOT_FOUND":                  "Không tìm thấy quảng cáo",
	"ad.NOT_UPDATED":                "Không tìm thấy quảng cáo để cập nhật",
	"ad.NOT_DELETED":                "Quảng cáo không tồn tại",
	"ad.NOT_OWNER":                  "Bạn không có quyền thay đổi quảng cáo này",

	// ── Package ────────────────────────────────────────────────
	"package.SUBSCRIBE_FAILED": "Không thể đăng ký gói",
	"package.NOT_SUBSCRIBED":   "Bạn chưa đăng ký gói quảng cáo nào",

	// ── Notification ───────────────────────────────────────────
	"notification.CREATE_FAILED":   "Không thể tạo thông báo",
	"notification.CREATE_BULK_FAILED": "Không thể tạo thông báo hàng loạt",

	// ── Admin ─────────────────────────────────────────────────
	"admin.ACTION_INVALID":              "Hành động moderation không hợp lệ",
	"admin.REASON_REQUIRED":             "Lý do là bắt buộc",
	"admin.REASON_TOO_SHORT":            "Lý do phải có ít nhất 10 ký tự",
	"admin.REASON_TOO_LONG":             "Lý do không được vượt quá 1000 ký tự",
	"admin.TRANSFER_SELF":              "Không thể chuyển quyền cho chính mình",
	"admin.TRANSFER_EMPTY":             "Người nhận không được để trống",
	"admin.ACTION_NOT_ALLOWED_GROUP":    "Action không được hỗ trợ cho group chat",
	"admin.ACTION_NOT_ALLOWED_COMMUNITY": "Action không được hỗ trợ cho cộng đồng",
	"admin.NO_ACCESS":                   "Không có quyền truy cập",
	"admin.NOT_SUPERADMIN":              "Chỉ superadmin mới có quyền thực hiện thao tác này",
	"admin.INVALID_STATUS":              "Trạng thái cập nhật không hợp lệ",
	"admin.INVALID_BAN_DURATION":        "Thời hạn ban không hợp lệ",
	"admin.NOT_FOUND":                   "Không tìm thấy tài nguyên",
	"admin.NOT_GROUP_CHAT":              "Chỉ có thể xóa group chat",
	"admin.HAS_OTHER_MEMBERS":           "Không thể xóa khi còn thành viên khác; hãy chuyển quyền sở hữu trước",
	"admin.NO_CREATOR":                  "Cộng đồng không có người tạo",
	"admin.ARCHIVED_ACTION_RESTRICTED":  "Không thể thao tác trên cộng đồng đã bị đình chỉ",
	"admin.ALREADY_IN_STATUS":           "Đã ở trạng thái {{status}}",
	"admin.SETTINGS_LOAD_FAILED":        "Không thể tải cài đặt",
	"admin.INVALID_SETTING_KEY":         "Cài đặt '{{key}}' không hợp lệ",
	"admin.INVALID_INT":                 "'{{key}}' phải là số nguyên",
	"admin.VALUE_TOO_LOW":               "'{{key}}' phải >= {{min}}",
	"admin.VALUE_TOO_HIGH":              "'{{key}}' không được vượt quá {{max}}",
	"admin.INVALID_BOOL":                "'{{key}}' phải là 'true' hoặc 'false'",
	"admin.INVALID_EMAIL":               "'{{key}}' không phải email hợp lệ",
	"admin.INVALID_QUERY_PARAMS":        "Tham số truy vấn không hợp lệ",
	"admin.INVALID_INPUT":               "Dữ liệu đầu vào không hợp lệ",
	"admin.SESSIONS_INVALIDATION_FAILED": "Vô hiệu hóa phiên đăng nhập thất bại",

	// ── RBAC ──────────────────────────────────────────────────
	"rbac.AUTHENTICATION_REQUIRED":         "Yêu cầu xác thực",
	"rbac.ROLE_NOT_FOUND":                  "Truy cập bị từ chối - Không tìm thấy vai trò nền tảng",
	"rbac.PERMISSION_DENIED":               "Bạn không có quyền truy cập tài nguyên này",
	"rbac.CONTRIBUTION_LEVEL_INSUFFICIENT": "Điểm đóng góp chưa đủ để thực hiện hành động này",
	"rbac.MISSING_COMMUNITY_ID":            "Thiếu community ID",
	"rbac.CONTRIBUTION_CHECK_FAILED":       "Kiểm tra cấp độ đóng góp thất bại",
	"rbac.MISSING_AD_ID":                   "Thiếu ad ID",
	"rbac.AD_NOT_FOUND":                    "Không tìm thấy quảng cáo",
	"rbac.AD_ACCESS_DENIED":                "Truy cập bị từ chối. Bạn không sở hữu quảng cáo này",

	// ── Ban ───────────────────────────────────────────────────
	"ban.NOT_FOUND": "Không tìm thấy bản ghi ban",
}
