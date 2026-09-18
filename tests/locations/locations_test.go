package locations_test

import (
	"context"
	"math"
	"testing"

	"linkup/internal/locations"
	"linkup/validations"
)

type stubGeocoder struct {
	display string
	err     error
}

func (s stubGeocoder) ReverseDisplayName(_ context.Context, _, _ float64) (string, error) {
	return s.display, s.err
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Hà Nội", "ha noi"},
		{"ĐÀ NẴNG", "da nang"},
		{"TP. Hồ Chí Minh", "tp ho chi minh"},
		{"Thành phố Hồ Chí Minh", "thanh pho ho chi minh"},
		{"Ba Đình", "ba dinh"},
		{"Xã Ấp-Phú-Q", "xa ap phu q"},
		{"HƯNG YÊN", "hung yen"},
		{"An Giang  ", "an giang"},
	}
	for _, tt := range tests {
		if got := locations.Normalize(tt.in); got != tt.want {
			t.Errorf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestStoreLoadsPost2025Dataset(t *testing.T) {
	store, err := locations.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	provinces := store.Provinces()
	if len(provinces) != 34 {
		t.Errorf("got %d provinces, want 34", len(provinces))
	}
	// "01" is Hanoi and should be first (official code order).
	if len(provinces) > 0 {
		if provinces[0].ID != "01" || provinces[0].Name.Vi != "Hà Nội" {
			t.Errorf("first province = %+v, want Hanoi (01)", provinces[0])
		}
	}

	hanoi, ok := store.ProvinceByID("01")
	if !ok {
		t.Fatal("ProvinceByID(01) missing")
	}
	if hanoi.WardCount == 0 {
		t.Errorf("Hanoi ward count = 0, expected > 0")
	}

	wards := store.Wards("01")
	if len(wards) != hanoi.WardCount {
		t.Errorf("Wards(01) len %d != WardCount %d", len(wards), hanoi.WardCount)
	}
	foundBaDinh := false
	for _, w := range wards {
		if w.ID == "00004" && w.Name.Vi == "Ba Đình" {
			foundBaDinh = true
		}
	}
	if !foundBaDinh {
		t.Error("Hanoi wards missing Ba Đình (00004)")
	}

	totalWards := 0
	for pid := range store.WardsByProvinceID() {
		totalWards += len(store.Wards(pid))
	}
	if totalWards != 3321 {
		t.Errorf("total wards = %d, want 3321", totalWards)
	}
}

func TestServiceListWards(t *testing.T) {
	store, err := locations.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	svc := locations.NewService(store, stubGeocoder{})

	wards, err := svc.ListWards("01")
	if err != nil || len(wards) == 0 {
		t.Fatalf("ListWards(01) = %v, %v", wards, err)
	}
	if wards[0].ID == "" || wards[0].Name.Vi == "" {
		t.Errorf("ward placeholder mismatch: %+v", wards[0])
	}

	if _, err := svc.ListWards("999"); err == nil {
		t.Error("ListWards(999) expected error for unknown province")
	}
}

func TestReverseGeocodeNameMatch(t *testing.T) {
	store, err := locations.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	svc := locations.NewService(store, stubGeocoder{display: "Phường Ba Đình, Quận Ba Đình, Hà Nội, Việt Nam"})

	res, err := svc.ReverseGeocode(context.Background(), 21.03, 105.84)
	if err != nil {
		t.Fatalf("ReverseGeocode: %v", err)
	}
	if !res.Matched {
		t.Fatal("expected matched=true")
	}
	if res.Province == nil || res.Province.ID != "01" || res.Province.Name.Vi != "Hà Nội" {
		t.Errorf("province = %+v, want Hanoi (01)", res.Province)
	}
	if res.Ward == nil || res.Ward.ID != "00004" || res.Ward.Name.Vi != "Ba Đình" {
		t.Errorf("ward = %+v, want Ba Đình (00004)", res.Ward)
	}
}

func TestReverseGeocodeProvinceOnly(t *testing.T) {
	store, err := locations.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	// District part is the most specific token, so only the province matches.
	svc := locations.NewService(store, stubGeocoder{display: "Quận Ba Đình, Hà Nội, Việt Nam"})

	res, err := svc.ReverseGeocode(context.Background(), 21.03, 105.83)
	if err != nil {
		t.Fatalf("ReverseGeocode: %v", err)
	}
	if !res.Matched || res.Province == nil || res.Province.ID != "01" {
		t.Errorf("expected Hanoi match, got %+v", res)
	}
}

func TestReverseGeocodeCentroidFallback(t *testing.T) {
	store, err := locations.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	// Unmatchable names force the nearest-centroid fallback near Hanoi.
	svc := locations.NewService(store, stubGeocoder{display: "Somewhere Else, Middle of Nowhere"})

	res, err := svc.ReverseGeocode(context.Background(), 21.04, 105.836)
	if err != nil {
		t.Fatalf("ReverseGeocode: %v", err)
	}
	if !res.Matched {
		t.Fatal("expected fallback match near Hanoi")
	}
	if res.Ward == nil {
		t.Fatal("expected ward from centroid fallback")
	}
	// Nearest centroid within 50 km: ward must belong to Hanoi.
	if res.Province == nil || res.Province.ID != "01" {
		t.Errorf("fallback province = %+v, want Hanoi (01)", res.Province)
	}
}

func TestReverseGeocodeNoMatch(t *testing.T) {
	store, err := locations.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	// (0,0) is in the South Atlantic — far outside any 50 km threshold.
	svc := locations.NewService(store, stubGeocoder{display: "Middle of the Ocean"})

	res, err := svc.ReverseGeocode(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("ReverseGeocode: %v", err)
	}
	if res.Matched {
		t.Errorf("expected matched=false, got %+v", res)
	}
}

func TestReverseGeocodeGeocoderNetworkFailure(t *testing.T) {
	store, err := locations.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	svc := locations.NewService(store, stubGeocoder{err: context.DeadlineExceeded})

	res, err := svc.ReverseGeocode(context.Background(), 21.04, 105.836)
	if err != nil {
		t.Fatalf("expected no error on geocoder failure, got %v", err)
	}
	if res.Matched {
		t.Error("expected matched=false on network failure")
	}
}

func TestReverseGeocodeInvalidCoordinates(t *testing.T) {
	store, err := locations.NewStore()
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	svc := locations.NewService(store, stubGeocoder{})

	if _, err := svc.ReverseGeocode(context.Background(), 91, 0); err == nil {
		t.Error("expected error for lat=91")
	}
	if _, err := svc.ReverseGeocode(context.Background(), 0, -181); err == nil {
		t.Error("expected error for lon=-181")
	}
}

func TestValidateCoordinates(t *testing.T) {
	if err := validations.ValidateCoordinates(21.04, 105.836); err != nil {
		t.Errorf("valid coords rejected: %v", err)
	}
	if err := validations.ValidateCoordinates(-90, 180); err != nil {
		t.Errorf("boundary coords rejected: %v", err)
	}
	if err := validations.ValidateCoordinates(90.1, 0); err == nil {
		t.Error("lat > 90 accepted")
	}
	if err := validations.ValidateCoordinates(0, -180.1); err == nil {
		t.Error("lon < -180 accepted")
	}
}

func TestHaversine(t *testing.T) {
	// Zero distance.
	if d := locations.HaversineKM(10, 20, 10, 20); d != 0 {
		t.Errorf("self distance = %v, want 0", d)
	}
	// Quarter meridian ≈ 10,007 km.
	d := locations.HaversineKM(0, 0, 90, 0)
	if math.Abs(d-10007) > 50 {
		t.Errorf("quarter meridian = %.1f km, want ~10007", d)
	}
}

func TestProfileEnumValidation(t *testing.T) {
	v := validations.NewProfileValidation()

	if err := v.ValidateWorkCode("it"); err != nil {
		t.Errorf("it should be valid: %v", err)
	}
	if err := v.ValidateWorkCode(""); err != nil {
		t.Errorf("empty work should be valid: %v", err)
	}
	if err := v.ValidateWorkCode("plumber"); err == nil {
		t.Error("plumber should be rejected")
	}

	if err := v.ValidateEducationCode("university"); err != nil {
		t.Errorf("university should be valid: %v", err)
	}
	if err := v.ValidateEducationCode("phd"); err == nil {
		t.Error("phd should be rejected")
	}

	if err := v.ValidateWorkOther("other", ""); err == nil {
		t.Error("other + empty work_other should be rejected")
	}
	if err := v.ValidateWorkOther("other", "Kỹ sư tự do"); err != nil {
		t.Errorf("other + text should be valid: %v", err)
	}
	if err := v.ValidateWorkOther("it", "nothing"); err != nil {
		t.Errorf("non-other + text should be valid: %v", err)
	}
}