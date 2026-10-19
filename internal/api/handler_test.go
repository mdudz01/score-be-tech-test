package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mdudz01/score-be-tech-test/internal/cases"
	"github.com/mdudz01/score-be-tech-test/internal/domain"
	"github.com/mdudz01/score-be-tech-test/internal/region"
)

// Fakes: each satisfies one of the handler's interfaces.
// They return whatever the test tells them to, so no real files or data
// are needed. A type fits an interface just by having the right method.

// fakeRepo stands in for the case repository.
type fakeRepo struct {
	c   *domain.Case
	err error
}

func (f fakeRepo) GetCase(ctx context.Context, id string) (*domain.Case, error) {
	return f.c, f.err
}

// fakeRegions stands in for the postcode -> region lookup.
type fakeRegions struct {
	region string
	err    error
}

func (f fakeRegions) ForPostcode(string) (string, error) { return f.region, f.err }

// fakeOns stands in for the ONS data.
type fakeOns struct {
	out map[string]float64
	err error
}

func (f fakeOns) Outgoings(string) (map[string]float64, error) { return f.out, f.err }

// caseWithPostcode builds a minimal case with one client at the given postcode.
func caseWithPostcode(pc string) *domain.Case {
	return &domain.Case{
		CaseID: "abc",
		Clients: []domain.Client{
			{Details: domain.Details{Address: domain.Address{Postcode: pc}}},
		},
	}
}

// TestCaseOutgoings checks the main endpoint in every situation we handle:
// the success path and each kind of failure.
func TestCaseOutgoings(t *testing.T) {
	// A working ONS fake that returns one known value.
	good := fakeOns{out: map[string]float64{"outgoings_food": 123.45}}

	// One row per scenario: which fakes to use, and the status we expect.
	tests := []struct {
		name       string
		repo       fakeRepo
		regions    fakeRegions
		ons        fakeOns
		wantStatus int
	}{
		{"success", fakeRepo{c: caseWithPostcode("NG1 2ET")}, fakeRegions{region: "East Midlands"}, good, http.StatusOK},
		{"case not found", fakeRepo{err: cases.ErrNotFound}, fakeRegions{}, good, http.StatusNotFound},
		{"repository failure", fakeRepo{err: errors.New("boom")}, fakeRegions{}, good, http.StatusInternalServerError},
		{"no clients", fakeRepo{c: &domain.Case{CaseID: "abc"}}, fakeRegions{}, good, http.StatusUnprocessableEntity},
		{"invalid postcode", fakeRepo{c: caseWithPostcode("???")}, fakeRegions{err: region.ErrInvalidPostcode}, good, http.StatusUnprocessableEntity},
		{"unknown area", fakeRepo{c: caseWithPostcode("QQ1 1AA")}, fakeRegions{err: region.ErrUnknownArea}, good, http.StatusUnprocessableEntity},
		{"ons data mismatch", fakeRepo{c: caseWithPostcode("NG1 2ET")}, fakeRegions{region: "Mars"}, fakeOns{err: errors.New("unknown region")}, http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Build the real handler, but with this row's fakes plugged in.
			h := NewHandler(tc.repo, tc.regions, tc.ons).Routes()

			// Make a fake request and a recorder that captures the response.
			// Nothing goes over a network: the handler is called directly.
			req := httptest.NewRequest(http.MethodGet, "/cases/abc/outgoings", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			// Check 1: the status code.
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}

			// Check 2: the response says it is JSON.
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			// Check 3: the body is valid JSON.
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("response is not valid JSON: %v", err)
			}

			// Check 4: the body has the right content for the outcome.
			if tc.wantStatus == http.StatusOK {
				if body["outgoings_food"] != 123.45 {
					t.Errorf("outgoings_food = %v, want 123.45", body["outgoings_food"])
				}
			} else if _, ok := body["error"]; !ok {
				t.Error("error responses should contain an \"error\" field")
			}
		})
	}
}

// TestHealth checks the /healthz endpoint returns 200.
func TestHealth(t *testing.T) {
	h := NewHandler(fakeRepo{}, fakeRegions{}, fakeOns{}).Routes()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
