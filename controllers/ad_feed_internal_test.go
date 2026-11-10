package controllers

import (
	"encoding/json"
	"strings"
	"testing"

	"linkup/models"
)

func TestToFeedAdItems_OnlyPublicFields(t *testing.T) {
	pkgID := "pkg-1"
	ads := []models.Ad{
		{
			ID:          "ad-1",
			PartnerID:   "partner-1",
			PackageID:   &pkgID,
			Title:       "Khuyến mãi",
			Content:     "Nội dung",
			Format:      models.AdFormatImage,
			TargetURL:   "https://linkup.vn/sale",
			Budget:      1000,
			TotalSpent:  120,
			CPMPrice:    25,
			CPCPrice:    2,
			MediaList: []models.AdMedia{
				{URL: "https://cdn/a.png", MediaType: "image"},
			},
		},
	}

	items := toFeedAdItems(ads)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != "ad-1" || items[0].TargetURL != "https://linkup.vn/sale" {
		t.Fatalf("field công khai bị mất: %+v", items[0])
	}
	if len(items[0].MediaList) != 1 || items[0].MediaList[0].URL != "https://cdn/a.png" {
		t.Fatalf("media_list không được map: %+v", items[0].MediaList)
	}

	raw, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	payload := string(raw)
	for _, forbidden := range []string{
		"partner_id", "package_id", "budget", "total_spent",
		"cpm_price", "cpc_price", "max_impressions", "daily_budget",
	} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("feed lộ field nội bộ %q: %s", forbidden, payload)
		}
	}
}

func TestToFeedAdItems_EmptySliceNotNil(t *testing.T) {
	items := toFeedAdItems(nil)
	if items == nil {
		t.Fatalf("phải trả slice rỗng, không phải nil")
	}
	if len(items) != 0 {
		t.Fatalf("expected empty, got %d", len(items))
	}
}
