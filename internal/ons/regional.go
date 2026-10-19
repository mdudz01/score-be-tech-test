package ons

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// codeRe matches ONS category codes such as "1", "4.3" or "7.2.2".
var codeRe = regexp.MustCompile(`^\d+(\.\d+)*$`)

// Store holds the weekly £ spend as region -> ONS code -> value.
// It is filled once at startup and only read afterwards, so many
// requests can use it at the same time safely.
type Store struct {
	data      map[string]map[string]float64
	colRegion map[int]string
}

// LoadRegional reads the regional spreadsheet at path into a Store.
func LoadRegional(path string) (*Store, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	// Read every row of the first sheet as a list of text cells.
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		return nil, fmt.Errorf("read sheet: %w", err)
	}

	// Find the header row (the one containing "UK") and the region columns.
	headerIdx, firstRegionCol, regionCols := findHeader(rows)
	if headerIdx < 0 {
		return nil, fmt.Errorf("could not find a header row containing %q", "UK")
	}

	// Create the store with an empty map for each region.
	s := &Store{
		data:      make(map[string]map[string]float64),
		colRegion: regionCols,
	}
	for _, region := range regionCols {
		s.data[region] = make(map[string]float64)
	}

	// Go through every row below the header.
	pending := "" // a code seen on a row whose figures are on a later row
	for _, row := range rows[headerIdx+1:] {
		code := findCode(row, firstRegionCol)
		vals := parseValues(row, regionCols)

		// switch with no value runs the first case that is true.
		switch {
		case code != "":
			// Normal row: code and figures together, or the code line of a
			// wrapped label (figures follow on the next row).
			if len(vals) > 0 {
				s.store(code, vals)
				pending = ""
			} else {
				pending = code // remember it for the next row
			}
		case pending != "" && len(vals) > 0:
			// Continuation line of a wrapped label: figures belong to the
			// code we saw on the previous row.
			s.store(pending, vals)
			pending = ""
		case isBlank(row):
			pending = "" // a spacer row ends any wrapped label
		}
	}
	return s, nil
}

// parseValues returns column number -> number for every region cell in the
// row that holds a valid number.
func parseValues(row []string, regionCols map[int]string) map[int]float64 {
	vals := make(map[int]float64)
	for col := range regionCols {
		if col >= len(row) {
			continue
		}
		clean := strings.ReplaceAll(strings.TrimSpace(row[col]), ",", "")
		v, err := strconv.ParseFloat(clean, 64)
		if err != nil {
			continue
		}
		vals[col] = v
	}
	return vals
}

// store saves a row's figures under the given ONS code, for each region.
func (s *Store) store(code string, vals map[int]float64) {
	for col, v := range vals {
		s.data[s.colRegion[col]][code] = v
	}
}

// isBlank reports whether every cell in the row is empty.
func isBlank(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

// findHeader locates the row containing "UK" and returns its row number,
// the column where the regions start, and a map of column number -> region name.
func findHeader(rows [][]string) (int, int, map[int]string) {
	for i, row := range rows {
		for j, cell := range row {
			if strings.TrimSpace(cell) != "UK" {
				continue // not the cell we want, keep looking
			}
			// Found "UK": every named cell from here to the right is a region.
			cols := make(map[int]string)
			for k := j; k < len(row); k++ {
				if name := strings.TrimSpace(row[k]); name != "" {
					cols[k] = name
				}
			}
			return i, j, cols
		}
	}
	return -1, 0, nil
}

// findCode looks in the label columns (everything left of the first
// region column) for an ONS code and returns it, or "" if there is none.
// Stopping at limit means a £ figure can't be mistaken for a code.
func findCode(row []string, limit int) string {
	for j := 0; j < len(row) && j < limit; j++ {
		c := strings.TrimSpace(row[j])
		if codeRe.MatchString(c) {
			return c
		}
	}
	return ""
}

// Weekly returns the weekly £ figure for a region and ONS code.
// The bool is false if there is no such value.
func (s *Store) Weekly(region, code string) (float64, bool) {
	v, ok := s.data[region][code]
	return v, ok
}
