package errors

// Common error codes
const (
	ErrCodeInvalidInput = "common.INVALID_INPUT"
	ErrCodeUnauthorized = "common.UNAUTHORIZED"
	ErrCodeForbidden    = "common.FORBIDDEN"
	ErrCodeNotFound     = "common.NOT_FOUND"
	ErrCodeInternal     = "common.INTERNAL_ERROR"
)

// Post error codes
const (
	// Post - Validation
	ErrCodePostTitleRequired   = "post.TITLE_REQUIRED"
	ErrCodePostTitleTooShort   = "post.TITLE_TOO_SHORT"
	ErrCodePostTitleTooLong    = "post.TITLE_TOO_LONG"
	ErrCodePostContentRequired = "post.CONTENT_REQUIRED"
	ErrCodePostContentTooLong  = "post.CONTENT_TOO_LONG"
	ErrCodePostIDRequired      = "post.POST_ID_REQUIRED"
	ErrCodePostInvalidStatus   = "post.INVALID_STATUS"
	ErrCodePostInvalidFormat   = "post.INVALID_FORMAT"

	// Post - Business
	ErrCodePostNotFound            = "post.POST_NOT_FOUND"
	ErrCodePostHiddenOrPrivate     = "post.POST_HIDDEN_OR_PRIVATE"
	ErrCodePostNotAccessible       = "post.POST_NOT_ACCESSIBLE"
	ErrCodePostNotShareable        = "post.POST_NOT_SHAREABLE"
	ErrCodePostCannotShareOwn      = "post.CANNOT_SHARE_OWN"
	ErrCodePostAlreadyShared       = "post.ALREADY_SHARED"
	ErrCodePostCannotSaveOwn       = "post.CANNOT_SAVE_OWN"
	ErrCodePostCannotDeleteOthers  = "post.CANNOT_DELETE_OTHERS"
	ErrCodePostContributionNotInit = "post.CONTRIBUTION_SERVICE_NOT_INIT"

	// Post - Comment
	ErrCodeCommentContentRequired = "post.COMMENT_CONTENT_REQUIRED"
	ErrCodeCommentContentTooLong  = "post.COMMENT_CONTENT_TOO_LONG"
	ErrCodeCommentNotFound        = "post.COMMENT_NOT_FOUND"
	ErrCodeCommentWrongPost       = "post.COMMENT_WRONG_POST"
	ErrCodeCommentHiddenPrivate   = "post.CANNOT_COMMENT_HIDDEN_PRIVATE"

	// Post - Reaction
	ErrCodeEmojiRequired = "post.EMOJI_REQUIRED"
	ErrCodeEmojiNotFound = "post.EMOJI_NOT_FOUND"

	// Post - Pagination
	ErrCodeInvalidPageSize = "post.INVALID_PAGE_SIZE"
)

// Community error codes
const (
	// Community - Validation
	ErrCodeCommunityNameRequired      = "community.NAME_REQUIRED"
	ErrCodeCommunityNameTooShort      = "community.NAME_TOO_SHORT"
	ErrCodeCommunityNameTooLong       = "community.NAME_TOO_LONG"
	ErrCodeCommunityDescTooLong       = "community.DESC_TOO_LONG"
	ErrCodeCommunityAvatarInvalid     = "community.AVATAR_INVALID"
	ErrCodeCommunityBackgroundInvalid = "community.BACKGROUND_INVALID"
	ErrCodeCommunityInvalidFormat     = "community.INVALID_FORMAT"

	// Community - Business
	ErrCodeCommunityNotFound        = "community.NOT_FOUND"
	ErrCodeCommunityNameExists      = "community.NAME_EXISTS"
	ErrCodeAdminCannotCreate        = "community.ADMIN_CANNOT_CREATE"
	ErrCodeEncryptionKeyFailed      = "community.ENCRYPTION_KEY_FAILED"
	ErrCodeBackgroundUploadFailed   = "community.BACKGROUND_UPLOAD_FAILED"
	ErrCodeBackgroundRejected       = "community.BACKGROUND_REJECTED"
	ErrCodeBackgroundUpdateFailed   = "community.BACKGROUND_UPDATE_FAILED"
	ErrCodeImageReadFailed          = "community.IMAGE_READ_FAILED"
	ErrCodeJoinRequestFailed        = "community.JOIN_REQUEST_FAILED"
	ErrCodeInviteCodeRequired       = "community.INVITE_CODE_REQUIRED"
	ErrCodeInvitationRequired       = "community.INVITATION_REQUIRED"
	ErrCodeGroupChatNotFound        = "community.GROUP_CHAT_NOT_FOUND"

	// Community - Membership
	ErrCodeAlreadyMember           = "community.ALREADY_MEMBER"
	ErrCodeJoinRequestPending      = "community.JOIN_REQUEST_PENDING"
	ErrCodeJoinRequestNotFound     = "community.JOIN_REQUEST_NOT_FOUND"
	ErrCodeJoinRequestHandled      = "community.JOIN_REQUEST_HANDLED"
	ErrCodeNotCommunityAdmin       = "community.NOT_ADMIN"
	ErrCodeNotCommunityMember      = "community.NOT_MEMBER"
	ErrCodeMemberNotFound          = "community.MEMBER_NOT_FOUND"
	ErrCodeInvalidRole             = "community.INVALID_ROLE"
	ErrCodeCannotChangeOwnRole     = "community.CANNOT_CHANGE_OWN_ROLE"
	ErrCodeCannotTargetAdmin       = "community.CANNOT_TARGET_ADMIN"
	ErrCodeCreatorCannotLeave      = "community.CREATOR_CANNOT_LEAVE"
	ErrCodeCannotKickCreator       = "community.CANNOT_KICK_CREATOR"
	ErrCodeCannotKickAdmin         = "community.CANNOT_KICK_ADMIN"
	ErrCodeKickReasonRequired      = "community.KICK_REASON_REQUIRED"
	ErrCodeKickReasonTooShort      = "community.KICK_REASON_TOO_SHORT"
	ErrCodeKickReasonTooLong       = "community.KICK_REASON_TOO_LONG"
	ErrCodeKickFailed              = "community.KICK_FAILED"
	ErrCodeRoleUpdateFailed        = "community.ROLE_UPDATE_FAILED"
	ErrCodeMemberCheckFailed       = "community.MEMBER_CHECK_FAILED"
	ErrCodeTransferOwnToSelf       = "community.TRANSFER_TO_SELF"
	ErrCodeOnlyCreatorCanTransfer  = "community.ONLY_CREATOR_CAN_TRANSFER"
	ErrCodeTransferFailed          = "community.TRANSFER_FAILED"

	// Community - Invite code
	ErrCodeInviteCodeNotFound    = "community.INVITE_CODE_NOT_FOUND"
	ErrCodeInviteCodeExpired     = "community.INVITE_CODE_EXPIRED"
	ErrCodeInviteCodeInactive    = "community.INVITE_CODE_INACTIVE"
	ErrCodeInviteCodeMaxUses     = "community.INVITE_CODE_MAX_USES"
	ErrCodeInviteCodeCreateFail  = "community.INVITE_CODE_CREATE_FAILED"
	ErrCodeInviteCodeSaveFail    = "community.INVITE_CODE_SAVE_FAILED"
	ErrCodeInviteCodeDeactFail   = "community.INVITE_CODE_DEACTIVATE_FAILED"
	ErrCodeInviteCodeListFail    = "community.INVITE_CODE_LIST_FAILED"

	// Community - Invitation
	ErrCodeInvitationNotFound      = "community.INVITATION_NOT_FOUND"
	ErrCodeInvitationHandled       = "community.INVITATION_HANDLED"
	ErrCodeCannotInviteSelf        = "community.CANNOT_INVITE_SELF"
	ErrCodeInvitationSendFailed    = "community.INVITATION_SEND_FAILED"
	ErrCodeInvitationListFailed    = "community.INVITATION_LIST_FAILED"

	// Community - Rule
	ErrCodeRuleCategoryRequired = "community.RULE_CATEGORY_INVALID"
	ErrCodeRuleTitleRequired    = "community.RULE_TITLE_REQUIRED"
	ErrCodeRuleTitleTooShort    = "community.RULE_TITLE_TOO_SHORT"
	ErrCodeRuleTitleTooLong     = "community.RULE_TITLE_TOO_LONG"
	ErrCodeRuleContentTooLong   = "community.RULE_CONTENT_TOO_LONG"
	ErrCodeRulePositionNegative = "community.RULE_POSITION_NEGATIVE"
	ErrCodeRuleTitleDuplicate   = "community.RULE_TITLE_DUPLICATE"
	ErrCodeRuleNotFound         = "community.RULE_NOT_FOUND"
	ErrCodeRuleCheckFailed      = "community.RULE_CHECK_FAILED"
	ErrCodeRulePositionFailed   = "community.RULE_POSITION_FAILED"
)

// Auth error codes
const (
	ErrCodeInvalidCredentials     = "auth.INVALID_CREDENTIALS"
	ErrCodeEmailExists            = "auth.EMAIL_EXISTS"
	ErrCodePasswordTooShort       = "auth.PASSWORD_TOO_SHORT"
	ErrCodePasswordTooLong        = "auth.PASSWORD_TOO_LONG"
	ErrCodePasswordMissingUpper   = "auth.PASSWORD_MISSING_UPPER"
	ErrCodePasswordMissingLower   = "auth.PASSWORD_MISSING_LOWER"
	ErrCodePasswordMissingDigit   = "auth.PASSWORD_MISSING_DIGIT"
	ErrCodePasswordMissingSpecial = "auth.PASSWORD_MISSING_SPECIAL"
	ErrCodePasswordSameAsOld      = "auth.PASSWORD_SAME_AS_OLD"
	ErrCodeDisplayNameRequired    = "auth.DISPLAY_NAME_REQUIRED"
	ErrCodeDisplayNameTooShort    = "auth.DISPLAY_NAME_TOO_SHORT"
	ErrCodeDisplayNameTooLong     = "auth.DISPLAY_NAME_TOO_LONG"
	ErrCodeEmailRequired          = "auth.EMAIL_REQUIRED"
	ErrCodeEmailInvalid           = "auth.EMAIL_INVALID"
	ErrCodePasswordRequired       = "auth.PASSWORD_REQUIRED"
	ErrCodeUsernameRequired       = "auth.USERNAME_REQUIRED"
	ErrCodeUsernameTooShort       = "auth.USERNAME_TOO_SHORT"
	ErrCodeUsernameTooLong        = "auth.USERNAME_TOO_LONG"
	ErrCodeUsernameInvalid        = "auth.USERNAME_INVALID"
	ErrCodeAccountLocked          = "auth.ACCOUNT_LOCKED"
	ErrCodeAccountInactive        = "auth.ACCOUNT_INACTIVE"
	ErrCodeEmailNotVerified       = "auth.EMAIL_NOT_VERIFIED"
	ErrCodeMaintenanceMode        = "auth.MAINTENANCE_MODE"
	ErrCodeRegistrationDisabled   = "auth.REGISTRATION_DISABLED"
	ErrCodeInvalidRefreshToken    = "auth.INVALID_REFRESH_TOKEN"
	ErrCodeSessionExpired         = "auth.SESSION_EXPIRED"
	ErrCodeInvalidGoogleToken     = "auth.INVALID_GOOGLE_TOKEN"
	ErrCodeGoogleNotConfigured    = "auth.GOOGLE_NOT_CONFIGURED"
	ErrCodeGoogleAccountMismatch  = "auth.GOOGLE_ACCOUNT_MISMATCH"
	ErrCodeLoginAttemptsRemaining = "auth.LOGIN_ATTEMPTS_REMAINING"
	ErrCodeInvalidToken           = "auth.INVALID_TOKEN"
	ErrCodeMissingAuthorization   = "auth.MISSING_AUTHORIZATION"
	ErrCodeLogoutFailed           = "auth.LOGOUT_FAILED"
	ErrCodeInvalidIDToken         = "auth.INVALID_ID_TOKEN"
	ErrCodeTokenClaimsInvalid     = "auth.TOKEN_CLAIMS_INVALID"
	ErrCodeTokenNotRefresh        = "auth.TOKEN_NOT_REFRESH"
	ErrCodeUserNotFound           = "auth.USER_NOT_FOUND"
	ErrCodeAccountBanned          = "auth.ACCOUNT_BANNED"
	ErrCodeAccountBannedWithExpiry = "auth.ACCOUNT_BANNED_WITH_EXPIRY"
	ErrCodeAccountBannedPermanent = "auth.ACCOUNT_BANNED_PERMANENT"

	// ── Group Chat ──────────────────────────────────────────────
	ErrCodeGCInvalidName         = "group_chat.INVALID_NAME"
	ErrCodeGCEncryptionKeyFailed = "group_chat.ENCRYPTION_KEY_FAILED"
	ErrCodeGCNotMember           = "group_chat.NOT_MEMBER"
	ErrCodeGCInvalidLeaveMode    = "group_chat.INVALID_LEAVE_MODE"
	ErrCodeGCInvalidHistoryMode  = "group_chat.INVALID_HISTORY_MODE"
	ErrCodeGCSelfBan             = "group_chat.SELF_BAN"
	ErrCodeGCAdminOnly           = "group_chat.ADMIN_ONLY"
	ErrCodeGCAlreadyBanned       = "group_chat.ALREADY_BANNED"
	ErrCodeGCSelfInvite          = "group_chat.SELF_INVITE"
	ErrCodeGCNotAdminInvite      = "group_chat.NOT_ADMIN_INVITE"
	ErrCodeGCAlreadyMember       = "group_chat.ALREADY_MEMBER"
	ErrCodeGCPendingRequest      = "group_chat.PENDING_REQUEST"
	ErrCodeGCAdminOnlyAdd        = "group_chat.ADMIN_ONLY_ADD"
	ErrCodeGCAdminOnlyName       = "group_chat.ADMIN_ONLY_NAME"
	ErrCodeGCInvalidGroupName    = "group_chat.INVALID_GROUP_NAME"
	ErrCodeGCAdminOnlyAvatar     = "group_chat.ADMIN_ONLY_AVATAR"
	ErrCodeGCAdminOnlyConfig     = "group_chat.ADMIN_ONLY_CONFIG"
	ErrCodeGCAdminOnlyUpdate     = "group_chat.ADMIN_ONLY_UPDATE"
	ErrCodeGCTransferToSelf      = "group_chat.TRANSFER_TO_SELF"
	ErrCodeGCOnlyCreatorTransfer = "group_chat.ONLY_CREATOR_TRANSFER"
	ErrCodeGCNotGroupTransfer    = "group_chat.NOT_GROUP_TRANSFER"
	ErrCodeGCTargetNotMember     = "group_chat.TARGET_NOT_MEMBER"
	ErrCodeGCAdminCooldown       = "group_chat.ADMIN_COOLDOWN"
	ErrCodeGCInvalidMuteReason   = "group_chat.INVALID_MUTE_REASON"
	ErrCodeGCInvalidMuteDuration = "group_chat.INVALID_MUTE_DURATION"
	ErrCodeGCSelfMute            = "group_chat.SELF_MUTE"
	ErrCodeGCNotMemberMute       = "group_chat.NOT_MEMBER_MUTE"
	ErrCodeGCRequestNotOwn       = "group_chat.REQUEST_NOT_OWN"
	ErrCodeGCRequestAlreadyHandled = "group_chat.REQUEST_ALREADY_HANDLED"
	ErrCodeGCNotGroupChat        = "group_chat.NOT_GROUP_CHAT"
	ErrCodeGCBanned              = "group_chat.BANNED"
	ErrCodeGCMuted               = "group_chat.MUTED"
	ErrCodeGCEmojiNotFound       = "group_chat.EMOJI_NOT_FOUND"
	ErrCodeGCMediaNotFound       = "group_chat.MEDIA_NOT_FOUND"
	ErrCodeGCMediaNotYours       = "group_chat.MEDIA_NOT_YOURS"
	ErrCodeGCReplyNotFound       = "group_chat.REPLY_NOT_FOUND"
	ErrCodeGCReplyWrongChat      = "group_chat.REPLY_WRONG_CHAT"
	ErrCodeGCEncryptionKeyNotFound = "group_chat.ENCRYPTION_KEY_NOT_FOUND"

	// ── Chat (DM) ──────────────────────────────────────────────
	ErrCodeChatContentRequired   = "chat.CONTENT_REQUIRED"
	ErrCodeChatContentTooLong    = "chat.CONTENT_TOO_LONG"
	ErrCodeChatSearchEmpty       = "chat.SEARCH_EMPTY"
	ErrCodeChatDeleteModeInvalid = "chat.DELETE_MODE_INVALID"
	ErrCodeChatNotParticipant    = "chat.NOT_PARTICIPANT"
	ErrCodeChatNotGroupChat      = "chat.NOT_GROUP_CHAT"
	ErrCodeChatSelfChat          = "chat.SELF_CHAT"
	ErrCodeChatAlreadyFriends    = "chat.ALREADY_FRIENDS"
	ErrCodeChatInvitePending     = "chat.INVITE_PENDING"
	ErrCodeChatInviteAccepted    = "chat.INVITE_ACCEPTED"
	ErrCodeChatNotRecipient      = "chat.NOT_RECIPIENT"
	ErrCodeChatNotSender         = "chat.NOT_SENDER"
	ErrCodeChatAlreadyDeleted    = "chat.ALREADY_DELETED"
	ErrCodeChatMessageNotFound   = "chat.MESSAGE_NOT_FOUND"
	ErrCodeChatAccessDenied      = "chat.ACCESS_DENIED"
	ErrCodeChatKeyNotFound       = "chat.KEY_NOT_FOUND"

	// ── Friend ─────────────────────────────────────────────────
	ErrCodeFriendTargetRequired       = "friend.TARGET_REQUIRED"
	ErrCodeFriendSelfRequest          = "friend.SELF_REQUEST"
	ErrCodeFriendUserNotFound         = "friend.USER_NOT_FOUND"
	ErrCodeFriendUserInactive         = "friend.USER_INACTIVE"
	ErrCodeFriendAdminRestricted      = "friend.ADMIN_RESTRICTED"
	ErrCodeFriendRequestHandled       = "friend.REQUEST_HANDLED"
	ErrCodeFriendNotFound             = "friend.NOT_FOUND"
	ErrCodeFriendNotAuthorized        = "friend.NOT_AUTHORIZED"
	ErrCodeFriendAlreadyProcessed     = "friend.ALREADY_PROCESSED"
	ErrCodeFriendNotFriends           = "friend.NOT_FRIENDS"
	ErrCodeFriendSelfUnfriend         = "friend.SELF_UNFRIEND"
	ErrCodeFriendTargetRequiredUnfriend = "friend.TARGET_REQUIRED_UNFRIEND"

	// ── Follow ─────────────────────────────────────────────────
	ErrCodeFollowSelf            = "follow.SELF"
	ErrCodeFollowUserNotFound    = "follow.USER_NOT_FOUND"
	ErrCodeFollowUserInactive    = "follow.USER_INACTIVE"
	ErrCodeFollowSuperAdmin      = "follow.SUPERADMIN_RESTRICTED"
	ErrCodeFollowAdminRestricted = "follow.ADMIN_RESTRICTED"

	// ── Block ──────────────────────────────────────────────────
	ErrCodeBlockTargetRequired = "block.TARGET_REQUIRED"
	ErrCodeBlockSelf           = "block.SELF"
	ErrCodeBlockAdminRestricted = "block.ADMIN_RESTRICTED"

	// ── Media ──────────────────────────────────────────────────
	ErrCodeMediaFileRequired        = "media.FILE_REQUIRED"
	ErrCodeMediaFileTypeNotAllowed  = "media.FILE_TYPE_NOT_ALLOWED"
	ErrCodeMediaFileTooLarge        = "media.FILE_TOO_LARGE"
	ErrCodeMediaInsufficientStorage = "media.INSUFFICIENT_STORAGE"
	ErrCodeMediaStorageQuotaExceeded = "media.STORAGE_QUOTA_EXCEEDED"
	ErrCodeMediaNotFound            = "media.NOT_FOUND"
	ErrCodeMediaForbidden           = "media.FORBIDDEN"
	ErrCodeMediaImageDecode         = "media.IMAGE_DECODE"
	ErrCodeMediaImageTooSmall       = "media.IMAGE_TOO_SMALL"
	ErrCodeMediaImageTooLarge       = "media.IMAGE_TOO_LARGE"
	ErrCodeMediaInvalidAspectRatio  = "media.INVALID_ASPECT_RATIO"

	// ── Contribution ───────────────────────────────────────────
	ErrCodeContribPostWeightInvalid     = "contribution.POST_WEIGHT_INVALID"
	ErrCodeContribCommentWeightInvalid  = "contribution.COMMENT_WEIGHT_INVALID"
	ErrCodeContribReactionWeightInvalid = "contribution.REACTION_WEIGHT_INVALID"
	ErrCodeContribEventWeightInvalid    = "contribution.EVENT_WEIGHT_INVALID"
	ErrCodeContribThresholdInvalid      = "contribution.THRESHOLD_INVALID"
	ErrCodeContribThresholdOrderInvalid = "contribution.THRESHOLD_ORDER_INVALID"
	ErrCodeContribHashtagRequired       = "contribution.HASHTAG_REQUIRED"
	ErrCodeContribHashtagInvalidFormat  = "contribution.HASHTAG_INVALID_FORMAT"
	ErrCodeContribEndDateBeforeStart    = "contribution.END_DATE_BEFORE_START"
	ErrCodeContribStartDateInPast       = "contribution.START_DATE_IN_PAST"
	ErrCodeContribTitleRequired         = "contribution.TITLE_REQUIRED"
	ErrCodeContribTitleTooShort         = "contribution.TITLE_TOO_SHORT"
	ErrCodeContribTitleTooLong          = "contribution.TITLE_TOO_LONG"
	ErrCodeContribDescTooLong           = "contribution.DESC_TOO_LONG"
	ErrCodeContribDateFormatInvalid     = "contribution.DATE_FORMAT_INVALID"
	ErrCodeContribChallengeNotFound     = "contribution.CHALLENGE_NOT_FOUND"
	ErrCodeContribChallengeInactive     = "contribution.CHALLENGE_INACTIVE"
	ErrCodeContribChallengeNotStarted   = "contribution.CHALLENGE_NOT_STARTED"
	ErrCodeContribChallengeEnded        = "contribution.CHALLENGE_ENDED"
	ErrCodeContribAlreadyJoined         = "contribution.ALREADY_JOINED"
	ErrCodeContribParticipantLimitHit   = "contribution.PARTICIPANT_LIMIT_HIT"

	// ── Call ───────────────────────────────────────────────────
	ErrCodeCallBusy       = "call.BUSY"
	ErrCodeCallNotFound   = "call.NOT_FOUND"
	ErrCodeCallNotWaiting = "call.NOT_WAITING"
	ErrCodeCallSelfCall   = "call.SELF_CALL"
	ErrCodeCallNotFriend  = "call.NOT_FRIEND"
	ErrCodeCallNotParticipant = "call.NOT_PARTICIPANT"
	ErrCodeCallAlreadyEnded   = "call.ALREADY_ENDED"
	ErrCodeCallNotVideo       = "call.NOT_VIDEO"
	ErrCodeCallNotConnected   = "call.NOT_CONNECTED"

	// ── AI Moderation ──────────────────────────────────────────
	ErrCodeModCloudinaryNotInit = "moderation.CLOUDINARY_NOT_INITIALIZED"
	ErrCodeModUploadFailed      = "moderation.UPLOAD_FAILED"

	// ── Profile ────────────────────────────────────────────────
	ErrCodeProfileNotFound = "profile.PROFILE_NOT_FOUND"
	ErrCodeProfilePrivate  = "profile.PRIVATE_PROFILE"
	ErrCodePhoneExists     = "profile.PHONE_EXISTS"

	// ── Password Reset ─────────────────────────────────────────
	ErrCodeResetTokenNotFound = "password_reset.TOKEN_NOT_FOUND"
	ErrCodeResetTokenExpired  = "password_reset.TOKEN_EXPIRED"
	ErrCodeResetTokenUsed     = "password_reset.TOKEN_USED"
	ErrCodeResetPasswordShort = "password_reset.PASSWORD_TOO_SHORT"
	ErrCodeResetPasswordLong  = "password_reset.PASSWORD_TOO_LONG"

	// ── Email Verification ─────────────────────────────────────
	ErrCodeVerifyTokenNotFound = "email_verification.TOKEN_NOT_FOUND"
	ErrCodeVerifyTokenExpired  = "email_verification.TOKEN_EXPIRED"
	ErrCodeVerifyTokenUsed     = "email_verification.TOKEN_USED"
	ErrCodeVerifyAlreadyDone   = "email_verification.ALREADY_VERIFIED"
	ErrCodeVerifyRateLimited   = "email_verification.RATE_LIMITED"

	// ── User Settings ──────────────────────────────────────────
	ErrCodeSettingsInvalidTheme    = "user_settings.INVALID_THEME"
	ErrCodeSettingsInvalidLanguage = "user_settings.INVALID_LANGUAGE"
	ErrCodeSettingsWrongPassword   = "user_settings.WRONG_PASSWORD"
	ErrCodeSettingsSessionNotFound = "user_settings.SESSION_NOT_FOUND"

	// ── Story ──────────────────────────────────────────────────
	ErrCodeStoryInvalidFormat    = "story.INVALID_FORMAT"
	ErrCodeStoryContentRequired  = "story.CONTENT_REQUIRED"
	ErrCodeStoryNotFound         = "story.NOT_FOUND"
	ErrCodeStoryInteractNotFound = "story.INTERACT_NOT_FOUND"
	ErrCodeStoryEmojiIDRequired  = "story.EMOJI_ID_REQUIRED"
	ErrCodeStoryReactLimitHit    = "story.REACT_LIMIT_REACHED"
	ErrCodeStoryEmojiNotFound    = "story.EMOJI_NOT_FOUND"
	ErrCodeStoryReplyEmpty       = "story.REPLY_EMPTY"
	ErrCodeStoryAnalyticsForbidden = "story.ANALYTICS_FORBIDDEN"

	// ── Search ─────────────────────────────────────────────────
	ErrCodeSearchKeywordTooShort = "search.KEYWORD_TOO_SHORT"
	ErrCodeSearchTypeInvalid     = "search.TYPE_INVALID"

	// ── Ad ─────────────────────────────────────────────────────
	ErrCodeAdNoSubscription      = "ad.NO_SUBSCRIPTION"
	ErrCodeAdSlotsExhausted      = "ad.SLOTS_EXHAUSTED"
	ErrCodeAdFormatNotVideo      = "ad.FORMAT_NOT_SUPPORTED_VIDEO"
	ErrCodeAdFormatNotCarousel   = "ad.FORMAT_NOT_SUPPORTED_CAROUSEL"
	ErrCodeAdNotFound            = "ad.NOT_FOUND"
	ErrCodeAdNotUpdated          = "ad.NOT_UPDATED"
	ErrCodeAdNotDeleted          = "ad.NOT_DELETED"
	ErrCodeAdNotOwner            = "ad.NOT_OWNER"

	// ── Package ────────────────────────────────────────────────
	ErrCodePackageSubscribeFailed = "package.SUBSCRIBE_FAILED"
	ErrCodePackageNotSubscribed   = "package.NOT_SUBSCRIBED"

	// ── Notification ───────────────────────────────────────────
	ErrCodeNotificationCreateFailed = "notification.CREATE_FAILED"
	ErrCodeNotificationBulkFailed   = "notification.CREATE_BULK_FAILED"

	// ── Admin ─────────────────────────────────────────────────
	// Validation (shared group/community)
	ErrCodeAdminActionInvalid         = "admin.ACTION_INVALID"
	ErrCodeAdminReasonRequired        = "admin.REASON_REQUIRED"
	ErrCodeAdminReasonTooShort        = "admin.REASON_TOO_SHORT"
	ErrCodeAdminReasonTooLong         = "admin.REASON_TOO_LONG"
	ErrCodeAdminTransferSelf          = "admin.TRANSFER_SELF"
	ErrCodeAdminTransferEmpty         = "admin.TRANSFER_EMPTY"
	ErrCodeAdminActionNotAllowedGroup = "admin.ACTION_NOT_ALLOWED_GROUP"
	ErrCodeAdminActionNotAllowedComm  = "admin.ACTION_NOT_ALLOWED_COMMUNITY"

	// Service
	ErrCodeAdminNoAccess              = "admin.NO_ACCESS"
	ErrCodeAdminNotSuperadmin         = "admin.NOT_SUPERADMIN"
	ErrCodeAdminInvalidStatus         = "admin.INVALID_STATUS"
	ErrCodeAdminInvalidBanDuration    = "admin.INVALID_BAN_DURATION"
	ErrCodeAdminNotFound              = "admin.NOT_FOUND"
	ErrCodeAdminNotGroupChat          = "admin.NOT_GROUP_CHAT"
	ErrCodeAdminHasOtherMembers       = "admin.HAS_OTHER_MEMBERS"
	ErrCodeAdminNoCreator             = "admin.NO_CREATOR"
	ErrCodeAdminArchivedRestricted    = "admin.ARCHIVED_ACTION_RESTRICTED"
	ErrCodeAdminAlreadyInStatus       = "admin.ALREADY_IN_STATUS"
	ErrCodeAdminSettingsLoadFailed    = "admin.SETTINGS_LOAD_FAILED"
	ErrCodeAdminInvalidSettingKey     = "admin.INVALID_SETTING_KEY"
	ErrCodeAdminInvalidInt            = "admin.INVALID_INT"
	ErrCodeAdminValueTooLow           = "admin.VALUE_TOO_LOW"
	ErrCodeAdminValueTooHigh          = "admin.VALUE_TOO_HIGH"
	ErrCodeAdminInvalidBool           = "admin.INVALID_BOOL"
	ErrCodeAdminInvalidEmail          = "admin.INVALID_EMAIL"
	ErrCodeAdminInvalidQueryParams    = "admin.INVALID_QUERY_PARAMS"
	ErrCodeAdminInvalidInput          = "admin.INVALID_INPUT"
	ErrCodeAdminSessionsInvalidFailed = "admin.SESSIONS_INVALIDATION_FAILED"

	// ── RBAC ──────────────────────────────────────────────────
	ErrCodeRbacAuthRequired         = "rbac.AUTHENTICATION_REQUIRED"
	ErrCodeRbacRoleNotFound         = "rbac.ROLE_NOT_FOUND"
	ErrCodeRbacPermissionDenied     = "rbac.PERMISSION_DENIED"
	ErrCodeRbacContributionInsufficient = "rbac.CONTRIBUTION_LEVEL_INSUFFICIENT"
	ErrCodeRbacMissingCommunityID   = "rbac.MISSING_COMMUNITY_ID"
	ErrCodeRbacContributionCheckFailed = "rbac.CONTRIBUTION_CHECK_FAILED"
	ErrCodeRbacMissingAdID          = "rbac.MISSING_AD_ID"
	ErrCodeRbacAdNotFound           = "rbac.AD_NOT_FOUND"
	ErrCodeRbacAdAccessDenied       = "rbac.AD_ACCESS_DENIED"

	// ── Ban ───────────────────────────────────────────────────
	ErrCodeBanNotFound = "ban.NOT_FOUND"
)
