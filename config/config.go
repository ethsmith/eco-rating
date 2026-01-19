// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// =============================================================================

// Package config handles application configuration loading, saving, and validation.
// It supports JSON configuration files and provides sensible defaults for all settings.
package config

import (
	"encoding/json"
	"os"
	"strings"
)

// Config holds all application configuration settings.
// These can be set via JSON config file or command-line flags.
type Config struct {
	Cumulative    bool   `json:"cumulative"`     // Enable batch processing mode
	Tier          string `json:"tier"`           // Competitive tier filter (comma-separated for multiple)
	BaseURL       string `json:"base_url"`       // Cloud bucket base URL
	Prefix        string `json:"prefix"`         // Bucket prefix for demo files
	DemoPath      string `json:"demo_path"`      // Path to single demo file (single mode)
	DemoDir       string `json:"demo_dir"`       // Local directory for downloaded demos
	EnableLogging bool   `json:"enable_logging"` // Enable detailed parsing logs
}

// DefaultConfig returns a Config with sensible default values.
// The defaults point to the CSC demo bucket for season 19 combines.
func DefaultConfig() *Config {
	return &Config{
		Cumulative:    false,
		Tier:          "",
		BaseURL:       "https://cscdemos.nyc3.digitaloceanspaces.com/",
		Prefix:        "s19/Combines/",
		DemoPath:      "",
		DemoDir:       "./demos",
		EnableLogging: true,
	}
}

// LoadConfig reads configuration from a JSON file at the given path.
// If the file doesn't exist, it returns default configuration.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// SaveConfig writes the configuration to a JSON file with pretty formatting.
func SaveConfig(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ValidTiers returns the list of valid competitive tier names.
// Tiers are ordered from highest to lowest skill level.
func ValidTiers() []string {
	return []string{
		"challenger",
		"contender",
		"elite",
		"premier",
		"prospect",
		"recruit",
	}
}

// IsValidTier checks if the given tier name is a recognized competitive tier.
func IsValidTier(tier string) bool {
	for _, t := range ValidTiers() {
		if t == tier {
			return true
		}
	}
	return false
}

// DemoPrefix returns the filename prefix used for demos of a given tier.
// Demo files are named like "combine-contender-map-timestamp.dem.zip".
func DemoPrefix(tier string) string {
	return "combine-" + tier
}

// ParseTiers splits a comma-separated tier string into individual tier names.
// It trims whitespace and filters out empty strings.
func ParseTiers(tierStr string) []string {
	if tierStr == "" {
		return nil
	}
	parts := strings.Split(tierStr, ",")
	tiers := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tiers = append(tiers, t)
		}
	}
	return tiers
}
