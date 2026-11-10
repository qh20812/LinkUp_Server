package services

import (
	"testing"

	"linkup/models"
)

func TestIsNotificationEnabled(t *testing.T) {
	// Pref "tất cả bật" (như row mới do seed tạo: mọi cột DEFAULT 1).
	allEnabled := &models.NotificationPreference{
		LikeEnabled:          true,
		StoryReactEnabled:    true,
		CommentEnabled:       true,
		FollowEnabled:        true,
		MessageEnabled:       true,
		FriendRequestEnabled: true,
		CommunityEnabled:     true,
		ShareEnabled:         true,
		MediaEnabled:         true,
		VoiceCallEnabled:     true,
	}

	for _, notifType := range []models.NotificationType{
		models.NotificationTypeLike,
		models.NotificationTypeStoryReact,
		models.NotificationTypeComment,
		models.NotificationTypeFollow,
		models.NotificationTypeMessage,
		models.NotificationTypeFriendRequest,
		models.NotificationTypeFriendAccepted,
		models.NotificationTypeCommunityJoinRequest,
		models.NotificationTypeCommunityInvitationAccepted,
		models.NotificationTypeShare,
		models.NotificationTypeMediaApproved,
		models.NotificationTypeMediaRejected,
		models.NotificationTypeMediaFlagged,
		models.NotificationTypeVoiceCall,
		models.NotificationTypeReportSubmitted,
		models.NotificationTypeReportResult,
		models.NotificationTypeModeration,
		models.NotificationTypeWallet,
		models.NotificationTypePackage,
		models.NotificationTypeMessageReact,
	} {
		if !isNotificationEnabled(allEnabled, notifType) {
			t.Errorf("expected enabled for %s with all-enabled pref", notifType)
		}
	}

	// Pref all-false → mọi loại đã map đều bị chặn (loại chưa map fallback true).
	allDisabled := &models.NotificationPreference{}
	for _, notifType := range []models.NotificationType{
		models.NotificationTypeLike,
		models.NotificationTypeStoryReact,
		models.NotificationTypeShare,
		models.NotificationTypeMediaFlagged,
		models.NotificationTypeVoiceCall,
		models.NotificationTypeMessageReact,
	} {
		if isNotificationEnabled(allDisabled, notifType) {
			t.Errorf("expected disabled for %s with all-disabled pref", notifType)
		}
	}
}

func TestIsNotificationEnabledFlagMapping(t *testing.T) {
	cases := []struct {
		name     string
		pref     *models.NotificationPreference
		enabled  []models.NotificationType
		disabled []models.NotificationType
	}{
		{
			name: "story_react",
			pref: &models.NotificationPreference{StoryReactEnabled: true},
			enabled: []models.NotificationType{
				models.NotificationTypeStoryReact,
			},
			disabled: []models.NotificationType{
				models.NotificationTypeLike,
				models.NotificationTypeShare,
				models.NotificationTypeMediaApproved,
			},
		},
		{
			name: "share",
			pref: &models.NotificationPreference{ShareEnabled: true},
			enabled: []models.NotificationType{
				models.NotificationTypeShare,
			},
			disabled: []models.NotificationType{
				models.NotificationTypeLike,
				models.NotificationTypeStoryReact,
			},
		},
		{
			name: "media",
			pref: &models.NotificationPreference{MediaEnabled: true},
			enabled: []models.NotificationType{
				models.NotificationTypeMediaApproved,
				models.NotificationTypeMediaRejected,
				models.NotificationTypeMediaFlagged,
			},
			disabled: []models.NotificationType{
				models.NotificationTypeLike,
			},
		},
		{
			name: "like touches like only not share",
			pref: &models.NotificationPreference{LikeEnabled: true},
			enabled: []models.NotificationType{
				models.NotificationTypeLike,
			},
			disabled: []models.NotificationType{
				models.NotificationTypeShare,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, notifType := range tc.enabled {
				if !isNotificationEnabled(tc.pref, notifType) {
					t.Errorf("expected %s enabled under %s", notifType, tc.name)
				}
			}
			for _, notifType := range tc.disabled {
				if isNotificationEnabled(tc.pref, notifType) {
					t.Errorf("expected %s disabled under %s", notifType, tc.name)
				}
			}
		})
	}
}

func TestParseNotificationTypeStoryReact(t *testing.T) {
	if got := models.ParseNotificationType("story_react"); got != models.NotificationTypeStoryReact {
		t.Fatalf("expected NotificationTypeStoryReact, got %q", got)
	}
}

func TestParseNotificationTypeReport(t *testing.T) {
	if got := models.ParseNotificationType("report_submitted"); got != models.NotificationTypeReportSubmitted {
		t.Fatalf("expected NotificationTypeReportSubmitted, got %q", got)
	}
	if got := models.ParseNotificationType("report_result"); got != models.NotificationTypeReportResult {
		t.Fatalf("expected NotificationTypeReportResult, got %q", got)
	}
}

func TestParseNotificationTypeSystem(t *testing.T) {
	cases := map[string]models.NotificationType{
		"moderation":    models.NotificationTypeModeration,
		"wallet":        models.NotificationTypeWallet,
		"package":       models.NotificationTypePackage,
		"message_react": models.NotificationTypeMessageReact,
	}
	for raw, want := range cases {
		if got := models.ParseNotificationType(raw); got != want {
			t.Errorf("ParseNotificationType(%q) = %q, want %q", raw, got, want)
		}
	}
}

// Report từng bị gắn type "message" nên tắt preference Tin nhắn là mất luôn
// thông báo kết quả. Type report phải độc lập với preference đó.
func TestReportNotificationsIgnoreMessagePreference(t *testing.T) {
	pref := &models.NotificationPreference{MessageEnabled: false}
	for _, notifType := range []models.NotificationType{
		models.NotificationTypeReportSubmitted,
		models.NotificationTypeReportResult,
	} {
		if !isNotificationEnabled(pref, notifType) {
			t.Errorf("expected %s enabled even when message notifications are off", notifType)
		}
	}
	if isNotificationEnabled(pref, models.NotificationTypeMessage) {
		t.Error("expected plain message notification disabled")
	}
}

// Moderation / wallet / package là thông báo hệ thống, không được phụ thuộc
// preference Tin nhắn; còn reaction tin nhắn thì có (là hoạt động chat).
func TestSystemNotificationPreferences(t *testing.T) {
	pref := &models.NotificationPreference{MessageEnabled: false}
	for _, notifType := range []models.NotificationType{
		models.NotificationTypeModeration,
		models.NotificationTypeWallet,
		models.NotificationTypePackage,
	} {
		if !isNotificationEnabled(pref, notifType) {
			t.Errorf("expected %s enabled even when message notifications are off", notifType)
		}
	}
	if isNotificationEnabled(pref, models.NotificationTypeMessageReact) {
		t.Error("expected message_react disabled when message notifications are off")
	}
}