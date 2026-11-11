package services

import (
	"testing"
	"time"

	"linkup/models"
)

func TestStoryActive(t *testing.T) {
	now := time.Now()

	if storyActive(&models.Story{ExpiresAt: nil}) {
		t.Error("story không có expires_at phải coi là hết hạn")
	}
	past := now.Add(-time.Minute)
	if storyActive(&models.Story{ExpiresAt: &past}) {
		t.Error("story quá hạn phải trả false")
	}
	future := now.Add(time.Hour)
	if !storyActive(&models.Story{ExpiresAt: &future}) {
		t.Error("story còn hạn phải trả true")
	}
}
