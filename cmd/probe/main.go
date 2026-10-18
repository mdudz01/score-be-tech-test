package main

import (
	"fmt"
	"log"

	"github.com/mdudz01/score-be-tech-test/internal/ons"
)

func main() {
	store, err := ons.LoadRegional("ONSByRegion2019.xlsx") // change to your real filename/path
	if err != nil {
		log.Fatal(err)
	}

	out, err := store.Outgoings("East Midlands")
	if err != nil {
		log.Fatal(err)
	}
	for field, value := range out {
		fmt.Printf("%-60s %8.2f\n", field, value)
	}
}
