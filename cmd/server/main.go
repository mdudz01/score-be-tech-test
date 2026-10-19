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
	// Read settings from environment variables, using the default if not set.
	regionalFile := getenv("REGIONAL_FILE", "ONSByRegion2019.xlsx")
	caseFile := getenv("CASE_FILE", "case.json")
	addr := ":" + getenv("PORT", "8080") // e.g. ":8080"

	// Load the ONS spreadsheet into memory. If this fails, stop straight away
	// rather than run with no data. os.Exit(1) means "exited with an error".
	store, err := ons.LoadRegional(regionalFile)
	if err != nil {
		slog.Error("loading ONS data", "err", err)
		os.Exit(1)
	}

	// Load the case file, stopping if it fails.
	repo, err := cases.LoadFile(caseFile)
	if err != nil {
		slog.Error("loading cases", "err", err)
		os.Exit(1)
	}

	// Connect the three real pieces to the HTTP handler. This is the only
	// place that knows about all of them.
	handler := api.NewHandler(repo, region.NewLookup(), store)

	// Configure the server. The timeouts stop slow or broken clients from
	// holding connections open forever.
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,  // time allowed to read the request headers
		ReadTimeout:       10 * time.Second, // time allowed to read the whole request
		WriteTimeout:      10 * time.Second, // time allowed to send the response
	}

	// Start listening. This blocks (keeps running) until the server stops,
	// so the code below only runs if something goes wrong.
	slog.Info("listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

// getenv returns the environment variable key, or fallback if it is empty.
// Go's os.Getenv has no built-in default, so this small helper adds one.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
