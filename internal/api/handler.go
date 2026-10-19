package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/mdudz01/score-be-tech-test/internal/cases"
	"github.com/mdudz01/score-be-tech-test/internal/domain"
	"github.com/mdudz01/score-be-tech-test/internal/region"
)

// These interfaces describe what the handler needs, not how it's provided.
// That keeps the HTTP layer testable with fakes and the data sources swappable.
// Any type with the right method automatically fits; there is no "implements".

// CaseRepository finds a case by ID.
type CaseRepository interface {
	GetCase(ctx context.Context, id string) (*domain.Case, error)
}

// RegionLookup turns a postcode into an ONS region name.
type RegionLookup interface {
	ForPostcode(postcode string) (string, error)
}

// OutgoingsSource returns the monthly outgoings for a region.
type OutgoingsSource interface {
	Outgoings(region string) (map[string]float64, error)
}

// Handler holds the three things the HTTP code depends on.
type Handler struct {
	cases   CaseRepository
	regions RegionLookup
	ons     OutgoingsSource
}

// NewHandler creates a Handler. main.go passes in the real versions,
// and the tests pass in fakes.
func NewHandler(c CaseRepository, r RegionLookup, o OutgoingsSource) *Handler {
	return &Handler{cases: c, regions: r, ons: o}
}

// Routes sets up the URLs the server responds to.
// "GET /cases/{id}/outgoings" includes the HTTP method, and {id} is a
// variable part of the path that handlers read with r.PathValue("id").
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /cases/{id}/outgoings", h.caseOutgoings)
	mux.HandleFunc("GET /regions/{region}/outgoings", h.regionOutgoings)
	return mux
}

// health lets monitoring tools check that the service is running.
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// caseOutgoings is the main endpoint. It follows four steps:
// case -> postcode -> region -> outgoings.
func (h *Handler) caseOutgoings(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Step 1: find the case.
	c, err := h.cases.GetCase(r.Context(), id)
	if err != nil {
		if errors.Is(err, cases.ErrNotFound) {
			writeError(w, http.StatusNotFound, "case not found")
			return
		}
		// Anything else is unexpected: log the details, but send a
		// generic message so internals aren't exposed.
		slog.Error("get case", "case_id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Step 2: get a postcode from the case.
	postcode := firstPostcode(c)
	if postcode == "" {
		writeError(w, http.StatusUnprocessableEntity, "case has no client postcode")
		return
	}

	// Step 3: turn the postcode into an ONS region.
	reg, err := h.regions.ForPostcode(postcode)
	if err != nil {
		if errors.Is(err, region.ErrInvalidPostcode) || errors.Is(err, region.ErrUnknownArea) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		slog.Error("region lookup", "case_id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Step 4: get the outgoings for that region.
	out, err := h.ons.Outgoings(reg)
	if err != nil {
		// The region lookup produced a name the ONS data doesn't have:
		// that's a server-side data mismatch, not a client error.
		slog.Error("outgoings", "case_id", id, "region", reg, "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Success: log it and send the values back.
	slog.Info("outgoings served", "case_id", id, "region", reg)
	writeJSON(w, http.StatusOK, out)
}

// regionOutgoings returns outgoings for a named region. It is mainly
// useful for checking the numbers.
func (h *Handler) regionOutgoings(w http.ResponseWriter, r *http.Request) {
	out, err := h.ons.Outgoings(r.PathValue("region"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// firstPostcode returns the first non-empty client postcode in the case.
// If clients live at different addresses, the first one wins.
func firstPostcode(c *domain.Case) string {
	for _, cl := range c.Clients { // _ ignores the position number
		if pc := strings.TrimSpace(cl.Details.Address.Postcode); pc != "" {
			return pc
		}
	}
	return ""
}

// writeJSON sends a response with the given status code and JSON body.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err)
	}
}

// writeError sends an error in the same shape every time: {"error": "..."}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
