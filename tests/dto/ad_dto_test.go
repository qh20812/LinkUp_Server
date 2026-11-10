package dto_test

import (
	"reflect"
	"testing"

	"linkup/dto"
)

// Hồi quy: MediaList là field chỉ dùng để render (không phải cột DB). Thiếu
// `gorm:"-"` thì GORM coi đây là quan hệ và `db.Raw(...).Scan(&items)` trả lỗi
// "invalid field found ... define a valid foreign key", làm cả danh sách ads 500.
func TestAdDTOMediaListIgnoredByGorm(t *testing.T) {
	for _, tc := range []struct {
		name string
		typ  reflect.Type
	}{
		{"PartnerAdListItem", reflect.TypeOf(dto.PartnerAdListItem{})},
		{"FeedAdItem", reflect.TypeOf(dto.FeedAdItem{})},
	} {
		field, ok := tc.typ.FieldByName("MediaList")
		if !ok {
			t.Fatalf("%s không còn field MediaList — cập nhật test", tc.name)
		}
		if got := field.Tag.Get("gorm"); got != "-" {
			t.Fatalf("%s.MediaList phải có gorm:\"-\", hiện tại gorm:%q", tc.name, got)
		}
	}
}
