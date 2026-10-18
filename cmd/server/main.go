package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/mdudz01/score-be-tech-test/internal/api"
	"github.com/mdudz01/score-be-tech-test/internal/cases"
	"github.com/mdudz01/score-be-tech-test/internal/ons"
	"github.com/mdudz01/score-be-tech-test/internal/region"
)

func main() {
	regionalFile := getenv("REGIONAL_FILE", "ONSByRegion2019.xlsx")
	caseFile := getenv("CASE_FILE", "case.json")
	addr := ":" + getenv("PORT", "8080")

	store, err := ons.LoadRegional(regionalFile)
	if err != nil {
		slog.Error("loading ONS data", "err", err)
		os.Exit(1)
	}

	repo, err := cases.LoadFile(caseFile)
	if err != nil {
		slog.Error("loading cases", "err", err)
		os.Exit(1)
	}

	handler := api.NewHandler(repo, region.NewLookup(), store)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	slog.Info("listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
