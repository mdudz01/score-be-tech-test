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
type fakeRepo struct {
	c   *domain.Case
	err error
}

func (f fakeRepo) GetCase(ctx context.Context, id string) (*domain.Case, error) {
	return f.c, f.err
}

type fakeRegions struct {
	region string
	err    error
}

func (f fakeRegions) ForPostcode(string) (string, error) { return f.region, f.err }

type fakeOns struct {
	out map[string]float64
	err error
}

func (f fakeOns) Outgoings(string) (map[string]float64, error) { return f.out, f.err }

func caseWithPostcode(pc string) *domain.Case {
	return &domain.Case{
		CaseID: "abc",
		Clients: []domain.Client{
			{Details: domain.Details{Address: domain.Address{Postcode: pc}}},
		},
	}
}

func TestCaseOutgoings(t *testing.T) {
	good := fakeOns{out: map[string]float64{"outgoings_food": 123.45}}

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
			h := NewHandler(tc.repo, tc.regions, tc.ons).Routes()

			req := httptest.NewRequest(http.MethodGet, "/cases/abc/outgoings", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("response is not valid JSON: %v", err)
			}
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

func TestHealth(t *testing.T) {
	h := NewHandler(fakeRepo{}, fakeRegions{}, fakeOns{}).Routes()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
