package main

import (
	"fmt"
	"log"

	"github.com/mdudz01/score-be-tech-test/internal/ons"
)

func main() {
	// Load the spreadsheet into memory (the same call the server makes).
	store, err := ons.LoadRegional("ONSByRegion2019.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	// Build the 21 outgoings values for one region.
	out, err := store.Outgoings("East Midlands")
	if err != nil {
		log.Fatal(err)
	}

	// Print each field and its monthly value.
	for field, value := range out {
		fmt.Printf("%-60s %8.2f\n", field, value)
	}
}
