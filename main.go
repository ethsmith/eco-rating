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

	demo, err := os.Open("C:\\Program Files (x86)\\Steam\\steamapps\\common\\Counter-Strike Global Offensive\\game\\csgo\\replays\\match730_003795643948026822985_1027935699_410.dem")
	if err != nil {
		log.Fatal(err)
	}
	defer demo.Close()

	p := parser.NewDemoParserWithLogging(demo, true)
	p.Parse()

	p.ExportJSON("match_rating.json")
}
