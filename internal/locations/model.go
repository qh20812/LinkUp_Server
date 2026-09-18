package locations

// PlaceName holds the local (Vietnamese) and English display names.
type PlaceName struct {
	Vi string `json:"vi"`
	En string `json:"en"`
}

// Place is an admin division (province or ward) exposed by the API.
type Place struct {
	ID   string    `json:"id"`
	Name PlaceName `json:"name"`
}

// Province is a level-1 division (tỉnh / thành phố trực thuộc TW).
type Province struct {
	ID        string
	Name      PlaceName
	Lat, Lon  float64
	WardCount int
}

// Ward is a level-2 division (xã / phường) belonging to one province.
type Ward struct {
	ID         string
	Name       PlaceName
	ProvinceID string
	Lat, Lon   float64
}

// GeoResult is the reverse-geocoding answer for a coordinate.
type GeoResult struct {
	Matched  bool   `json:"matched"`
	Province *Place `json:"province,omitempty"`
	Ward     *Place `json:"ward,omitempty"`
}

// rawDivision mirrors a single entry of the open-admin-data JSON documents.
type rawDivision struct {
	ID     string         `json:"id"`
	Level  int            `json:"level"`
	Name   rawName        `json:"name"`
	Parent *rawParent     `json:"parent"`
	Geo    rawGeo         `json:"geo"`
}

type rawName struct {
	Local string `json:"local"`
	En    string `json:"en"`
	Slug  string `json:"slug"`
}

type rawParent struct {
	ID   string  `json:"id"`
	Level int    `json:"level"`
	Name rawName `json:"name"`
}

type rawGeo struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}