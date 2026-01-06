package main

import (
	"log"
	"os"

	"eco-rating/parser"
)

func main() {
	//if len(os.Args) < 2 {
	//	log.Fatal("Usage: cs2-rating <demo.dem>")
	//}

	demo, err := os.Open("C:\\Users\\ethan\\Downloads\\combine-contender-mid6675-0_de_inferno-2026-01-06_02-42-14.dem\\demos\\combine-contender-mid6675-0_de_inferno-2026-01-06_02-42-14.dem")
	if err != nil {
		log.Fatal(err)
	}
	defer demo.Close()

	p := parser.NewDemoParserWithLogging(demo, true)
	p.Parse()

	p.ExportJSON("match_rating.json")
}
