// Package export provides CSV file export functionality for player statistics.
// This file implements export functionality for per-round crossfire and flash data.
package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ethsmith/eco-rating/model"
)

// CrossfireFlashExporter handles exporting crossfire and flash event data.
type CrossfireFlashExporter struct {
	OutputPath string // Base path for output files (without extension)
}

// NewCrossfireFlashExporter creates a new exporter with the specified output path.
func NewCrossfireFlashExporter(outputPath string) *CrossfireFlashExporter {
	// Remove extension if present
	ext := filepath.Ext(outputPath)
	if ext != "" {
		outputPath = outputPath[:len(outputPath)-len(ext)]
	}
	return &CrossfireFlashExporter{OutputPath: outputPath}
}

// Export writes crossfire and flash event data to CSV files.
func (e *CrossfireFlashExporter) Export(data *model.CrossfireFlashData) error {
	if data == nil {
		return nil
	}

	// Export crossfire events
	if len(data.CrossfireEvents) > 0 {
		if err := e.exportCrossfireEvents(data); err != nil {
			return fmt.Errorf("failed to export crossfire events: %w", err)
		}
	}

	// Export flash kill events
	if len(data.FlashKillEvents) > 0 {
		if err := e.exportFlashKillEvents(data); err != nil {
			return fmt.Errorf("failed to export flash kill events: %w", err)
		}
	}

	// Export combined JSON for programmatic access
	if err := e.exportJSON(data); err != nil {
		return fmt.Errorf("failed to export JSON: %w", err)
	}

	return nil
}

// exportCrossfireEvents writes crossfire events to a CSV file.
func (e *CrossfireFlashExporter) exportCrossfireEvents(data *model.CrossfireFlashData) error {
	path := e.OutputPath + "_crossfire.csv"
	if err := ensureDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	// Write header
	header := []string{
		"round_number",
		"time_in_round",
		"killer_steam_id",
		"killer_name",
		"partner_steam_id",
		"partner_name",
		"victim_steam_id",
		"victim_name",
		"killer_x",
		"killer_y",
		"killer_z",
		"partner_x",
		"partner_y",
		"partner_z",
		"victim_x",
		"victim_y",
		"victim_z",
		"crossfire_angle",
		"partner_distance",
		"killer_view_angle",
		"partner_view_angle",
		"killer_side",
		"map_name",
		"zone",
	}
	if err := w.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write rows
	for _, event := range data.CrossfireEvents {
		row := []string{
			strconv.Itoa(event.RoundNumber),
			fmt.Sprintf("%.2f", event.TimeInRound),
			event.KillerSteamID,
			event.KillerName,
			event.PartnerSteamID,
			event.PartnerName,
			event.VictimSteamID,
			event.VictimName,
			fmt.Sprintf("%.1f", event.KillerPosition.X),
			fmt.Sprintf("%.1f", event.KillerPosition.Y),
			fmt.Sprintf("%.1f", event.KillerPosition.Z),
			fmt.Sprintf("%.1f", event.PartnerPosition.X),
			fmt.Sprintf("%.1f", event.PartnerPosition.Y),
			fmt.Sprintf("%.1f", event.PartnerPosition.Z),
			fmt.Sprintf("%.1f", event.VictimPosition.X),
			fmt.Sprintf("%.1f", event.VictimPosition.Y),
			fmt.Sprintf("%.1f", event.VictimPosition.Z),
			fmt.Sprintf("%.1f", event.CrossfireAngle),
			fmt.Sprintf("%.1f", event.PartnerDistance),
			fmt.Sprintf("%.1f", event.KillerViewAngle),
			fmt.Sprintf("%.1f", event.PartnerViewAngle),
			event.KillerSide,
			event.MapName,
			event.Zone,
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// exportFlashKillEvents writes flash kill events to a CSV file.
func (e *CrossfireFlashExporter) exportFlashKillEvents(data *model.CrossfireFlashData) error {
	path := e.OutputPath + "_flash_kills.csv"
	if err := ensureDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	// Write header
	header := []string{
		"round_number",
		"time_in_round",
		"killer_steam_id",
		"killer_name",
		"flasher_steam_id",
		"flasher_name",
		"victim_steam_id",
		"victim_name",
		"flash_duration",
		"time_from_flash",
		"killer_x",
		"killer_y",
		"killer_z",
		"flasher_x",
		"flasher_y",
		"flasher_z",
		"victim_x",
		"victim_y",
		"victim_z",
		"flash_x",
		"flash_y",
		"flash_z",
		"killer_side",
		"map_name",
		"zone",
		"was_full_blind",
		"was_pop_flash",
	}
	if err := w.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write rows
	for _, event := range data.FlashKillEvents {
		row := []string{
			strconv.Itoa(event.RoundNumber),
			fmt.Sprintf("%.2f", event.TimeInRound),
			event.KillerSteamID,
			event.KillerName,
			event.FlasherSteamID,
			event.FlasherName,
			event.VictimSteamID,
			event.VictimName,
			fmt.Sprintf("%.2f", event.FlashDuration),
			fmt.Sprintf("%.2f", event.TimeFromFlash),
			fmt.Sprintf("%.1f", event.KillerPosition.X),
			fmt.Sprintf("%.1f", event.KillerPosition.Y),
			fmt.Sprintf("%.1f", event.KillerPosition.Z),
			fmt.Sprintf("%.1f", event.FlasherPosition.X),
			fmt.Sprintf("%.1f", event.FlasherPosition.Y),
			fmt.Sprintf("%.1f", event.FlasherPosition.Z),
			fmt.Sprintf("%.1f", event.VictimPosition.X),
			fmt.Sprintf("%.1f", event.VictimPosition.Y),
			fmt.Sprintf("%.1f", event.VictimPosition.Z),
			fmt.Sprintf("%.1f", event.FlashPosition.X),
			fmt.Sprintf("%.1f", event.FlashPosition.Y),
			fmt.Sprintf("%.1f", event.FlashPosition.Z),
			event.KillerSide,
			event.MapName,
			event.Zone,
			strconv.FormatBool(event.WasFullBlind),
			strconv.FormatBool(event.WasPopFlash),
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// exportJSON writes the complete crossfire/flash data to a JSON file.
func (e *CrossfireFlashExporter) exportJSON(data *model.CrossfireFlashData) error {
	path := e.OutputPath + "_crossfire_flash.json"
	if err := ensureDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// GetCrossfireCSVPath returns the path to the crossfire CSV file.
func (e *CrossfireFlashExporter) GetCrossfireCSVPath() string {
	return e.OutputPath + "_crossfire.csv"
}

// GetFlashKillsCSVPath returns the path to the flash kills CSV file.
func (e *CrossfireFlashExporter) GetFlashKillsCSVPath() string {
	return e.OutputPath + "_flash_kills.csv"
}

// GetJSONPath returns the path to the JSON file.
func (e *CrossfireFlashExporter) GetJSONPath() string {
	return e.OutputPath + "_crossfire_flash.json"
}
