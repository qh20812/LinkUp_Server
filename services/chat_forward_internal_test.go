package services

import (
	"testing"
)

func TestForwardHasContent(t *testing.T) {
	// Rỗng → không có nội dung để chuyển tiếp.
	if forwardHasContent("", nil, nil, nil) {
		t.Fatal("expected false for empty forward")
	}
	// Chỉ khoảng trắng → cũng là rỗng.
	ws := "  \t "
	if forwardHasContent(ws, nil, nil, nil) {
		t.Fatal("expected false for whitespace-only content")
	}

	// Có text → hợp lệ.
	if !forwardHasContent("hello", nil, nil, nil) {
		t.Fatal("expected true for text content")
	}

	// Có media → hợp lệ.
	mediaID := "m1"
	if !forwardHasContent("", &mediaID, nil, nil) {
		t.Fatal("expected true for media forward")
	}

	// Có emoji → hợp lệ.
	emojiID := "e1"
	if !forwardHasContent("", nil, &emojiID, nil) {
		t.Fatal("expected true for emoji forward")
	}

	// Có shared post → hợp lệ.
	postID := "p1"
	if !forwardHasContent("", nil, nil, &postID) {
		t.Fatal("expected true for shared post forward")
	}
}