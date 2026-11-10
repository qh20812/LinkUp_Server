package controllers

import (
	"testing"

	"linkup/internal/locations"
)

func TestValidateTargetLocations(t *testing.T) {
	store, err := locations.NewStore()
	if err != nil {
		t.Fatalf("load location store: %v", err)
	}
	ctrl := &AdController{locationStore: store}

	tests := []struct {
		name    string
		input   []string
		wantErr bool
	}{
		{name: "rỗng = toàn quốc", input: nil},
		{name: "province ID hợp lệ", input: []string{"01", "79"}},
		{name: "bỏ khoảng trắng + trùng lặp", input: []string{" 01 ", "01"}},
		{name: "tên tỉnh (dữ liệu cũ) bị từ chối", input: []string{"Hà Nội"}, wantErr: true},
		{name: "ID không tồn tại bị từ chối", input: []string{"99x"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ctrl.validateTargetLocations(tc.input)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateTargetLocations_NilStoreSkipsLookup(t *testing.T) {
	ctrl := &AdController{}
	if err := ctrl.validateTargetLocations([]string{"anything"}); err != nil {
		t.Fatalf("thiếu store thì chỉ bỏ qua kiểm tra tồn tại: %v", err)
	}
}
