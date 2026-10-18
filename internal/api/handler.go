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
type CaseRepository interface {
	GetCase(ctx context.Context, id string) (*domain.Case, error)
}

type RegionLookup interface {
	ForPostcode(postcode string) (string, error)
}

type OutgoingsSource interface {
	Outgoings(region string) (map[string]float64, error)
}

type Handler struct {
	cases   CaseRepository
	regions RegionLookup
	ons     OutgoingsSource
}

func NewHandler(c CaseRepository, r RegionLookup, o OutgoingsSource) *Handler {
	return &Handler{cases: c, regions: r, ons: o}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /cases/{id}/outgoings", h.caseOutgoings)
	mux.HandleFunc("GET /regions/{region}/outgoings", h.regionOutgoings)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) caseOutgoings(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	c, err := h.cases.GetCase(r.Context(), id)
	if err != nil {
		if errors.Is(err, cases.ErrNotFound) {
			writeError(w, http.StatusNotFound, "case not found")
			return
		}
		slog.Error("get case", "case_id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	postcode := firstPostcode(c)
	if postcode == "" {
		writeError(w, http.StatusUnprocessableEntity, "case has no client postcode")
		return
	}

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

	out, err := h.ons.Outgoings(reg)
	if err != nil {
		// The region lookup produced a name the ONS data doesn't have:
		// that's a server-side data mismatch, not a client error.
		slog.Error("outgoings", "case_id", id, "region", reg, "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	slog.Info("outgoings served", "case_id", id, "region", reg)
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) regionOutgoings(w http.ResponseWriter, r *http.Request) {
	out, err := h.ons.Outgoings(r.PathValue("region"))
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// firstPostcode returns the first non-empty client postcode in the case.
func firstPostcode(c *domain.Case) string {
	for _, cl := range c.Clients {
		if pc := strings.TrimSpace(cl.Details.Address.Postcode); pc != "" {
			return pc
		}
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
