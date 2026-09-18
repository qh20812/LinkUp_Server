package locations

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// Data files come from
// https://github.com/open-admin-data/vietnam-administrative-divisions
// (post-2025 structure: 34 provinces, 3.321 wards — no district level).
//
//go:embed data/all-province.json
var provinceJSON []byte

//go:embed data/all-ward.json
var wardJSON []byte

// Store is an in-memory index over the embedded Vietnamese admin dataset.
type Store struct {
	provinces        []Province
	provinceByID     map[string]Province
	wardsByProvince  map[string][]Ward
}

// NewStore builds the index from the embedded JSON files.
func NewStore() (*Store, error) {
	var rawP []rawDivision
	if err := json.Unmarshal(provinceJSON, &rawP); err != nil {
		return nil, fmt.Errorf("parse provinces: %w", err)
	}
	var rawW []rawDivision
	if err := json.Unmarshal(wardJSON, &rawW); err != nil {
		return nil, fmt.Errorf("parse wards: %w", err)
	}

	s := &Store{
		provinceByID:    make(map[string]Province, len(rawP)),
		wardsByProvince: make(map[string][]Ward),
	}

	for _, r := range rawP {
		p := Province{
			ID:   r.ID,
			Name: PlaceName{Vi: r.Name.Local, En: r.Name.En},
		}
		if r.Geo.Lat != "" {
			if lat, err := strconv.ParseFloat(r.Geo.Lat, 64); err == nil {
				p.Lat = lat
			}
		}
		if r.Geo.Lon != "" {
			if lon, err := strconv.ParseFloat(r.Geo.Lon, 64); err == nil {
				p.Lon = lon
			}
		}
		s.provinceByID[p.ID] = p
		s.provinces = append(s.provinces, p)
	}

	for _, r := range rawW {
		if r.Parent == nil {
			continue
		}
		w := Ward{
			ID:         r.ID,
			Name:       PlaceName{Vi: r.Name.Local, En: r.Name.En},
			ProvinceID: r.Parent.ID,
		}
		if r.Geo.Lat != "" {
			if lat, err := strconv.ParseFloat(r.Geo.Lat, 64); err == nil {
				w.Lat = lat
			}
		}
		if r.Geo.Lon != "" {
			if lon, err := strconv.ParseFloat(r.Geo.Lon, 64); err == nil {
				w.Lon = lon
			}
		}
		s.wardsByProvince[w.ProvinceID] = append(s.wardsByProvince[w.ProvinceID], w)
	}

	// Provinces in official code order; wards sorted by normalized name
	// for friendlier pickers.
	sort.Slice(s.provinces, func(i, j int) bool { return s.provinces[i].ID < s.provinces[j].ID })
	for k := range s.wardsByProvince {
		sort.Slice(s.wardsByProvince[k], func(i, j int) bool {
			return Normalize(s.wardsByProvince[k][i].Name.Vi) < Normalize(s.wardsByProvince[k][j].Name.Vi)
		})
	}
	for i := range s.provinces {
		p := s.provinces[i]
		p.WardCount = len(s.wardsByProvince[p.ID])
		s.provinces[i] = p
		s.provinceByID[p.ID] = p
	}

	return s, nil
}

// Provinces returns all provinces in official code order.
func (s *Store) Provinces() []Province {
	return s.provinces
}

// ProvinceByID returns the province with the given id, or false.
func (s *Store) ProvinceByID(id string) (Province, bool) {
	p, ok := s.provinceByID[id]
	return p, ok
}

// Wards returns the wards of the given province (empty slice if unknown province).
func (s *Store) Wards(provinceID string) []Ward {
	return s.wardsByProvince[provinceID]
}

// WardsByProvinceID maps every province id to its ward count for coverage checks.
func (s *Store) WardsByProvinceID() map[string][]Ward {
	return s.wardsByProvince
}