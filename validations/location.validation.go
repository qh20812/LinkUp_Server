package validations

import (
	errorsapp "linkup/errors"
)

type LocationValidation struct{}

func NewLocationValidation() *LocationValidation {
	return &LocationValidation{}
}

// ValidateCoordinates checks a latitude/longitude pair is within the valid
// WGS84 range.
func (v *LocationValidation) ValidateCoordinates(lat, lon float64) error {
	return ValidateCoordinates(lat, lon)
}

// ValidateCoordinates is the package-level helper (also used outside structs).
func ValidateCoordinates(lat, lon float64) error {
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return errorsapp.New(errorsapp.ErrCodeInvalidCoordinates)
	}
	return nil
}