package notification_test

import (
	"testing"

	"linkup/models"
)

func TestNewDefaultNotificationPreference_AllEnabled(t *testing.T) {
	pref := models.NewDefaultNotificationPreference("user-1")
	if pref.UserID != "user-1" {
		t.Errorf("UserID = %q", pref.UserID)
	}
	flags := map[string]bool{
		"LikeEnabled": pref.LikeEnabled, "CommentEnabled": pref.CommentEnabled,
		"FollowEnabled": pref.FollowEnabled, "MessageEnabled": pref.MessageEnabled,
		"FriendRequestEnabled": pref.FriendRequestEnabled, "CommunityEnabled": pref.CommunityEnabled,
		"VoiceCallEnabled": pref.VoiceCallEnabled, "StoryReactEnabled": pref.StoryReactEnabled,
		"ShareEnabled": pref.ShareEnabled, "MediaEnabled": pref.MediaEnabled,
	}
	for name, on := range flags {
		if !on {
			t.Errorf("%s = false, want true", name)
		}
	}
}
