package ons

import (
	"math"
	"strings"
	"testing"
)

func testStore() *Store {
	return &Store{data: map[string]map[string]float64{
		"East Midlands": {"1": 60, "4.3": 12, "8": 6},
	}}
}

func approx(a, b float64) bool { return math.Abs(a-b) < 0.005 }

func TestOutgoings(t *testing.T) {
	out, err := testStore().Outgoings("East Midlands")
	if err != nil {
		t.Fatal(err)
	}

	// Weekly figures are converted to monthly: weekly * 52 / 12.
	checks := map[string]float64{
		"outgoings_food":           260.00, // 60 * 52/12
		"outgoings_water":          52.00,  // 12 * 52/12
		"outgoings_communications": 26.00,  // 6 * 52/12
		"outgoings_clothing":       0,      // code missing from the data
		"outgoings_mortgage_rent":  0,      // unmapped by design
	}
	for field, want := range checks {
		got, ok := out[field]
		if !ok {
			t.Errorf("%s missing from output", field)
			continue
		}
		if !approx(got, want) {
			t.Errorf("%s = %.2f, want %.2f", field, got, want)
		}
	}

	// All 21 fields from the brief must always be present.
	if len(out) != 21 {
		t.Errorf("got %d fields, want 21", len(out))
	}
}

func TestOutgoingsUnknownRegion(t *testing.T) {
	if _, err := testStore().Outgoings("Narnia"); err == nil {
		t.Error("expected an error for an unknown region")
	}
}

// Guards against double counting: no code may be used twice, and no code may
// be used alongside its own parent (e.g. "9.4" together with "9.4.1").
func TestMappingHasNoOverlap(t *testing.T) {
	owner := map[string]string{}
	for field, codes := range Mapping {
		for _, c := range codes {
			if prev, dup := owner[c]; dup {
				t.Errorf("code %s used by both %s and %s", c, prev, field)
			}
			owner[c] = field
		}
	}
	for parent, pf := range owner {
		for child, cf := range owner {
			if strings.HasPrefix(child, parent+".") {
				t.Errorf("code %s (%s) contains %s (%s)", parent, pf, child, cf)
			}
		}
	}
}
