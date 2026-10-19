package region

import (
	"errors"
	"fmt"
	"strings"
)

// Named errors, so the API can tell them apart and reply with a 422.
var (
	ErrInvalidPostcode = errors.New("invalid postcode")
	ErrUnknownArea     = errors.New("unknown postcode area")
)

// areaToRegion maps a postcode area (the leading letters, e.g. "NG") to an
// ONS region. Names must match the column headers in the ONS spreadsheet.
var areaToRegion = map[string]string{
	// North East
	"DH": "North East", "DL": "North East", "NE": "North East", "SR": "North East", "TS": "North East",
	// North West
	"BB": "North West", "BL": "North West", "CA": "North West", "CH": "North West", "CW": "North West",
	"FY": "North West", "L": "North West", "LA": "North West", "M": "North West", "OL": "North West",
	"PR": "North West", "SK": "North West", "WA": "North West", "WN": "North West",
	// Yorkshire and the Humber
	"BD": "Yorkshire and the Humber", "DN": "Yorkshire and the Humber", "HD": "Yorkshire and the Humber",
	"HG": "Yorkshire and the Humber", "HU": "Yorkshire and the Humber", "HX": "Yorkshire and the Humber",
	"LS": "Yorkshire and the Humber", "S": "Yorkshire and the Humber", "WF": "Yorkshire and the Humber",
	"YO": "Yorkshire and the Humber",
	// East Midlands
	"DE": "East Midlands", "LE": "East Midlands", "LN": "East Midlands", "NG": "East Midlands", "NN": "East Midlands",
	// West Midlands
	"B": "West Midlands", "CV": "West Midlands", "DY": "West Midlands", "HR": "West Midlands",
	"ST": "West Midlands", "SY": "West Midlands", "TF": "West Midlands", "WR": "West Midlands",
	"WS": "West Midlands", "WV": "West Midlands",
	// East of England
	"AL": "East", "CB": "East", "CM": "East", "CO": "East", "IP": "East", "LU": "East", "NR": "East",
	"PE": "East", "SG": "East", "SS": "East", "WD": "East",
	// London
	"BR": "London", "CR": "London", "E": "London", "EC": "London", "EN": "London", "HA": "London",
	"IG": "London", "KT": "London", "N": "London", "NW": "London", "RM": "London", "SE": "London",
	"SM": "London", "SW": "London", "TW": "London", "UB": "London", "W": "London", "WC": "London",
	// South East
	"BN": "South East", "CT": "South East", "DA": "South East", "GU": "South East", "HP": "South East",
	"ME": "South East", "MK": "South East", "OX": "South East", "PO": "South East", "RG": "South East",
	"RH": "South East", "SL": "South East", "SO": "South East", "TN": "South East",
	// South West
	"BA": "South West", "BH": "South West", "BS": "South West", "DT": "South West", "EX": "South West",
	"GL": "South West", "PL": "South West", "SN": "South West", "SP": "South West", "TA": "South West",
	"TQ": "South West", "TR": "South West",
	// Wales
	"CF": "Wales", "LD": "Wales", "LL": "Wales", "NP": "Wales", "SA": "Wales",
	// Scotland
	"AB": "Scotland", "DD": "Scotland", "DG": "Scotland", "EH": "Scotland", "FK": "Scotland", "G": "Scotland",
	"HS": "Scotland", "IV": "Scotland", "KA": "Scotland", "KW": "Scotland", "KY": "Scotland", "ML": "Scotland",
	"PA": "Scotland", "PH": "Scotland", "TD": "Scotland", "ZE": "Scotland",
	// Northern Ireland
	"BT": "Northern Ireland",
}

// Lookup finds the region for a postcode. It has no data of its own (the
// table above is shared), but being a type lets the API use it through an
// interface, so a better lookup could replace it later.
type Lookup struct{}

// NewLookup creates a Lookup.
func NewLookup() *Lookup { return &Lookup{} }

// ForPostcode returns the ONS region for a UK postcode.
func (l *Lookup) ForPostcode(postcode string) (string, error) {
	// Step 1: get the area letters (e.g. "NG").
	area, err := AreaFromPostcode(postcode)
	if err != nil {
		return "", err
	}

	// Step 2: look the area up in the table.
	r, ok := areaToRegion[area]
	if !ok {
		// %w wraps the named error so callers can still detect it.
		return "", fmt.Errorf("%w: %q", ErrUnknownArea, area)
	}
	return r, nil
}

// AreaFromPostcode returns the leading letters of a postcode ("NG1 2ET" -> "NG").
func AreaFromPostcode(postcode string) (string, error) {
	// Tidy the input: remove spaces at the ends and use capitals.
	pc := strings.ToUpper(strings.TrimSpace(postcode))

	// Count how many letters the postcode starts with.
	i := 0
	for i < len(pc) && pc[i] >= 'A' && pc[i] <= 'Z' {
		i++
	}

	// Valid if there are 1 or 2 letters, followed by a digit.
	if i == 0 || i > 2 || i >= len(pc) || pc[i] < '0' || pc[i] > '9' {
		return "", fmt.Errorf("%w: %q", ErrInvalidPostcode, postcode)
	}

	// Slice the string to keep just the letters at the start.
	return pc[:i], nil
}
