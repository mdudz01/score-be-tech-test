package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/mdudz01/score-be-tech-test/internal/ons"
)

func main() {
	path := os.Getenv("REGIONAL_FILE")
	if path == "" {
		path = "ONSByRegion2019.xlsx"
	}

	store, err := ons.LoadRegional(path)
	if err != nil {
		log.Fatalf("loading ONS data: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("GET /regions/{region}/outgoings", func(w http.ResponseWriter, r *http.Request) {
		out, err := store.Outgoings(r.PathValue("region"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
