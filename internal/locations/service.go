package locations

import (
	"context"
	"math"
	"strings"

	errorsapp "linkup/errors"
	"linkup/validations"
)

const (
	earthRadiusKM = 6371.0
	nearestLimitKM = 50.0
)

// Service answers province/ward lookups and reverse-geocoding.
type Service struct {
	store *Store
	geo   Geocoder
}

func NewService(store *Store, geo Geocoder) *Service {
	return &Service{store: store, geo: geo}
}

// ListProvinces returns all 34 provinces/cities in official code order.
func (s *Service) ListProvinces() []Place {
	provinces := s.store.Provinces()
	out := make([]Place, 0, len(provinces))
	for _, p := range provinces {
		out = append(out, toPlace(p.ID, p.Name))
	}
	return out
}

// ListWards returns the wards of the given province. Unknown province ids
// yield a province-not-found error.
func (s *Service) ListWards(provinceID string) ([]Place, error) {
	if _, ok := s.store.ProvinceByID(provinceID); !ok {
		return nil, errorsapp.New(errorsapp.ErrCodeProvinceNotFound)
	}
	wards := s.store.Wards(provinceID)
	out := make([]Place, 0, len(wards))
	for _, w := range wards {
		out = append(out, toPlace(w.ID, w.Name))
	}
	return out, nil
}

// ReverseGeocode resolves a coordinate to a province and (best-effort) ward.
// It never fails on network trouble — an unusable answer yields matched=false.
func (s *Service) ReverseGeocode(ctx context.Context, lat, lon float64) (GeoResult, error) {
	if err := validations.ValidateCoordinates(lat, lon); err != nil {
		return GeoResult{}, err
	}

	displayName, err := s.geo.ReverseDisplayName(ctx, lat, lon)
	if err != nil || strings.TrimSpace(displayName) == "" {
		return GeoResult{}, nil
	}

	parts := tokenize(displayName)

	// Province = first non-country part scanning from the end.
	var province *Province
	var provincePart string
	for i := len(parts) - 1; i >= 0; i-- {
		if isCountryPart(parts[i]) {
			continue
		}
		provincePart = parts[i]
		break
	}
	if provincePart != "" {
		if p, ok := s.matchProvince(provincePart); ok {
			province = &p
		}
	}

	// Ward = most specific (first) part, matched within the found province.
	var ward *Ward
	wardPart := ""
	if len(parts) > 0 {
		wardPart = parts[0]
	}
	if wardPart != "" && province != nil {
		if w, ok := s.matchWard(province.ID, wardPart); ok {
			ward = &w
		}
	}

	// Fallback: nearest ward centroid within nearestLimitKM.
	if ward == nil {
		if w, ok := s.nearestWard(lat, lon); ok {
			ward = &w
			if province == nil {
				if p, ok := s.store.ProvinceByID(w.ProvinceID); ok {
					province = &p
				}
			}
		}
	}

	result := GeoResult{Matched: province != nil}
	if province != nil {
		p := toPlace(province.ID, province.Name)
		result.Province = &p
	}
	if ward != nil {
		w := toPlace(ward.ID, ward.Name)
		result.Ward = &w
	}
	return result, nil
}

func (s *Service) matchProvince(part string) (Province, bool) {
	provinces := s.store.Provinces()
	for i := range provinces {
		if matchName(provinces[i].Name.Vi, provinces[i].Name.En, part) {
			return provinces[i], true
		}
	}
	return Province{}, false
}

func (s *Service) matchWard(provinceID, part string) (Ward, bool) {
	for _, w := range s.store.Wards(provinceID) {
		if matchName(w.Name.Vi, w.Name.En, part) {
			return w, true
		}
	}
	return Ward{}, false
}

// nearestWard returns the ward whose centroid is closest to the coordinate,
// provided it is within nearestLimitKM.
func (s *Service) nearestWard(lat, lon float64) (Ward, bool) {
	var best Ward
	bestDist := math.Inf(1)
	found := false
	for _, wards := range s.store.WardsByProvinceID() {
		for _, w := range wards {
			d := HaversineKM(lat, lon, w.Lat, w.Lon)
			if d < bestDist {
				bestDist = d
				best = w
				found = true
			}
		}
	}
	if !found || bestDist > nearestLimitKM {
		return Ward{}, false
	}
	return best, true
}

func toPlace(id string, n PlaceName) Place {
	return Place{ID: id, Name: n}
}

// tokenize splits a display_name into comma-separated, trimmed parts.
func tokenize(displayName string) []string {
	raw := strings.Split(displayName, ",")
	parts := make([]string, 0, len(raw))
	for _, p := range raw {
		if p = strings.TrimSpace(p); p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

// isCountryPart reports whether a display_name part is just the country.
func isCountryPart(part string) bool {
	switch Normalize(part) {
	case "viet nam", "vietnam", "viet", "viet-nam", "vn":
		return true
	}
	return false
}

// haversineKM returns the great-circle distance in kilometers.
func HaversineKM(lat1, lon1, lat2, lon2 float64) float64 {
	lat1, lon1, lat2, lon2 = rad(lat1), rad(lon1), rad(lat2), rad(lon2)
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	return 2 * earthRadiusKM * math.Asin(math.Sqrt(a))
}

func rad(deg float64) float64 { return deg * math.Pi / 180 }