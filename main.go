package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"eco-rating/bucket"
	"eco-rating/config"
	"eco-rating/downloader"
	"eco-rating/model"
	"eco-rating/output"
	"eco-rating/parser"
)

func main() {
	// Command line flags
	configPath := flag.String("config", "", "Path to configuration file (defaults to config.json in executable directory)")
	cumulative := flag.Bool("cumulative", false, "Enable cumulative mode to fetch all demos for a tier")
	tier := flag.String("tier", "", "Tier to filter demos (challenger, contender, elite, premier, prospect, recruit)")
	demoPath := flag.String("demo", "", "Path to a single demo file to parse")
	outputDir := flag.String("output", "./demos", "Output directory for downloaded demos")
	flag.Parse()

	// Determine config path - default to config.json in current working directory
	cfgPath := *configPath
	if cfgPath == "" {
		// First try current working directory
		if _, err := os.Stat("config.json"); err == nil {
			cfgPath = "config.json"
		} else {
			// Fall back to executable directory
			exePath, err := os.Executable()
			if err != nil {
				cfgPath = "config.json"
			} else {
				cfgPath = filepath.Join(filepath.Dir(exePath), "config.json")
			}
		}
	}

	// Load configuration
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override config with command line flags if provided
	if *cumulative {
		cfg.Cumulative = true
	}
	if *tier != "" {
		cfg.Tier = *tier
	}
	if *outputDir != "./demos" {
		cfg.OutputDir = *outputDir
	}
	if *demoPath != "" {
		cfg.DemoPath = *demoPath
	}

	// Cumulative mode - fetch all demos for a tier and aggregate stats
	if cfg.Cumulative {
		if cfg.Tier == "" {
			log.Fatal("Tier must be specified in cumulative mode (use -tier flag or set in config)")
		}
		if !config.IsValidTier(cfg.Tier) {
			log.Fatalf("Invalid tier '%s'. Valid tiers: %v", cfg.Tier, config.ValidTiers())
		}

		runCumulativeMode(cfg)
		return
	}

	// Single demo mode - parse demo from config or flag
	if cfg.DemoPath != "" {
		parseSingleDemo(cfg.DemoPath, cfg.EnableLogging)
		return
	}

	// Default behavior - show usage
	fmt.Println("Usage:")
	fmt.Println("  Cumulative mode: eco-rating -cumulative -tier=contender")
	fmt.Println("  Single demo:     eco-rating -demo=path/to/demo.dem")
	fmt.Println("  Or set demo_path in config.json")
	fmt.Println()
	flag.PrintDefaults()
}

func runCumulativeMode(cfg *config.Config) {
	log.Printf("Running in cumulative mode for tier: %s", cfg.Tier)

	// Create bucket client
	client := bucket.NewClient(cfg.BaseURL)

	// Get all demos for the tier
	log.Printf("Fetching demo list from %s%s...", cfg.BaseURL, cfg.Prefix)
	demos, err := client.GetAllDemosByTier(cfg.Prefix, cfg.Tier)
	if err != nil {
		log.Fatalf("Failed to get demos: %v", err)
	}

	log.Printf("Found %d demos for tier '%s'", len(demos), cfg.Tier)

	// Create downloader and aggregator
	dl := downloader.NewDownloader(cfg.OutputDir)
	aggregator := output.NewAggregator()

	successCount := 0

	// Process each demo
	for i, demo := range demos {
		log.Printf("[%d/%d] Processing: %s", i+1, len(demos), demo.Key)

		// Get download URL
		url := client.GetDownloadURL(demo.Key)

		// Download and extract
		demoPath, err := dl.DownloadAndExtract(url)
		if err != nil {
			log.Printf("  Error downloading: %v", err)
			continue
		}

		log.Printf("  Extracted to: %s", demoPath)

		// Parse the demo and get player stats
		players, err := parseDemo(demoPath, cfg.EnableLogging)
		if err != nil {
			log.Printf("  Parse error: %v", err)
			continue
		}

		// Add to aggregator
		aggregator.AddGame(players)
		successCount++
		log.Printf("  Added to aggregation (players: %d)", len(players))
	}

	// Finalize and export aggregated stats
	aggregator.Finalize()

	outputPath := "match_rating.json"
	if err := output.ExportAggregated(aggregator.GetResults(), outputPath); err != nil {
		log.Fatalf("Failed to export aggregated stats: %v", err)
	}

	log.Printf("Completed processing %d/%d demos", successCount, len(demos))
	log.Printf("Aggregated stats for %d players saved to: %s", len(aggregator.GetResults()), outputPath)
}

func parseSingleDemo(demoPath string, enableLogging bool) {
	demo, err := os.Open(demoPath)
	if err != nil {
		log.Fatalf("Failed to open demo: %v", err)
	}
	defer demo.Close()

	p := parser.NewDemoParserWithLogging(demo, enableLogging)
	p.Parse()

	outputPath := "match_rating.json"
	if err := p.ExportJSON(outputPath); err != nil {
		log.Fatalf("Failed to export JSON: %v", err)
	}

	log.Printf("Results saved to: %s", outputPath)
}

func parseDemo(demoPath string, enableLogging bool) (map[uint64]*model.PlayerStats, error) {
	demo, err := os.Open(demoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open demo: %w", err)
	}
	defer demo.Close()

	p := parser.NewDemoParserWithLogging(demo, enableLogging)
	p.Parse()

	return p.GetPlayers(), nil
}
