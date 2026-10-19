package region

import (
	"errors"
	"testing"
)

// TestForPostcode checks the postcode -> region lookup with a table of
// examples. Each row is one test case: an input and the expected result.
// Test functions start with "Test" and take a *testing.T.
func TestForPostcode(t *testing.T) {
	l := NewLookup()

	// A slice of anonymous structs: one entry per test case.
	tests := []struct {
		name     string
		postcode string
		want     string
		wantErr  error
	}{
		{"full postcode", "NG1 2ET", "East Midlands", nil},
		{"lowercase and padded", "  ng1 2et ", "East Midlands", nil},
		{"single letter area", "M1 1AA", "North West", nil},
		{"london", "SW1A 1AA", "London", nil},
		{"northern ireland", "BT1 5GS", "Northern Ireland", nil},
		{"empty", "", "", ErrInvalidPostcode},
		{"letters only", "HELLO", "", ErrInvalidPostcode},
		{"digits only", "12345", "", ErrInvalidPostcode},
		{"unknown area", "QQ1 1AA", "", ErrUnknownArea},
	}

	// Run every row as its own named subtest, so a failure says which one broke.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := l.ForPostcode(tc.postcode)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}

			if got != tc.want {
				t.Errorf("region = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestAllRegionNamesAreValid catches typos in the area table. A misspelt
// region name would only fail at runtime for that postcode area, because
// the ONS data wouldn't have a region with that name.
func TestAllRegionNamesAreValid(t *testing.T) {
	// The region names used as column headers in the ONS spreadsheet.
	valid := map[string]bool{
		"North East": true, "North West": true, "Yorkshire and the Humber": true,
		"East Midlands": true, "West Midlands": true, "East": true, "London": true,
		"South East": true, "South West": true, "Wales": true, "Scotland": true,
		"Northern Ireland": true,
	}

	// Check every region name in the lookup table against the list.
	for area, r := range areaToRegion {
		if !valid[r] {
			t.Errorf("area %q maps to unrecognised region %q", area, r)
		}
	}
}
