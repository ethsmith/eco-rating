// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// There may be mistakes in the comments. Please verify accuracy.
// =============================================================================

// Package output provides JSON export functionality for player statistics.
// This file contains standalone export functions for writing stats to JSON files.
package output

import (
	"encoding/json"
	"os"

	"eco-rating/model"
)

// Export writes single-game player statistics to a JSON file with pretty formatting.
// The output is a map of Steam ID (uint64) to PlayerStats.
func Export(players map[uint64]*model.PlayerStats, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(players)
}

// ExportAggregated writes aggregated multi-game statistics to a JSON file.
// The output is a map of player key ("SteamID:Tier") to AggregatedStats.
func ExportAggregated(players map[string]*AggregatedStats, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(players)
}
