package ons

import (
	"fmt"
	"math"
)

// Mapping links each outgoings field to the ONS category codes that make
// it up. Codes are chosen so that no two overlap (e.g. we never add "9.4"
// and "9.4.1" together), which avoids double counting.
var Mapping = map[string][]string{
	"outgoings_water":                 {"4.3"},
	"outgoings_communications":        {"8"},
	"outgoings_insurance":             {"12.4"},
	"outgoings_investments":           {"14.6"},
	"outgoings_pension":               {"14.1"},
	"outgoings_council_tax":           {"13.1"}, // ONS bundles this with mortgage interest
	"outgoings_food":                  {"1"},
	"outgoings_clothing":              {"3"},
	"outgoings_household_repairs":     {"4.2"},
	"outgoings_fuel":                  {"4.4"}, // household gas and electricity
	"outgoings_car_costs":             {"7.1", "7.2"},
	"outgoings_other_transport_costs": {"7.3"},
	"outgoings_television_license":    {"9.4.3"},
	"outgoings_sports":                {"9.3.3", "9.4.1"},
	"outgoings_holidays":              {"9.6", "11.2"},
	"outgoings_entertainment":         {"9.1", "9.2", "9.3.1", "9.3.2", "9.3.4", "9.3.5", "9.4.2", "9.4.4", "9.4.6", "9.5", "11.1"},
	"outgoings_child_care":            {"12.3"}, // closest ONS match (social protection)
	"outgoings_other_living_costs":    {"2", "5", "6", "10", "12.1", "12.2", "12.5"},
}

// Unmapped fields have no sensible ONS equivalent and are returned as 0.
var Unmapped = []string{
	"outgoings_mortgage_rent",
	"outgoings_ground_rent_service_charge_shared_equity_rent",
	"outgoings_additional_details",
}

// ONS figures are weekly; outgoings are assumed to be monthly.
const weeksPerMonth = 52.0 / 12.0

// Outgoings returns estimated monthly £ values for every outgoings field
// for the given region.
func (s *Store) Outgoings(region string) (map[string]float64, error) {
	if _, ok := s.data[region]; !ok {
		return nil, fmt.Errorf("unknown region %q", region)
	}

	out := make(map[string]float64, len(Mapping)+len(Unmapped))

	for field, codes := range Mapping {
		var weekly float64
		for _, code := range codes {
			if v, ok := s.Weekly(region, code); ok {
				weekly += v
			}
		}
		out[field] = round2(weekly * weeksPerMonth)
	}
	for _, field := range Unmapped {
		out[field] = 0
	}
	return out, nil
}

func round2(f float64) float64 {
	return math.Round(f*100) / 100
}
