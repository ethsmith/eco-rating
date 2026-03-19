// Package model defines the core data structures for player and round statistics.
// This file defines structures for per-round crossfire and flash event tracking.
package model

// CrossfireEvent represents a crossfire kill event where two teammates
// were in crossfire positions when a kill occurred.
type CrossfireEvent struct {
	RoundNumber      int      `json:"round_number"`
	TimeInRound      float64  `json:"time_in_round"`
	KillerSteamID    string   `json:"killer_steam_id"`
	KillerName       string   `json:"killer_name"`
	PartnerSteamID   string   `json:"partner_steam_id"`
	PartnerName      string   `json:"partner_name"`
	VictimSteamID    string   `json:"victim_steam_id"`
	VictimName       string   `json:"victim_name"`
	KillerPosition   Position `json:"killer_position"`
	PartnerPosition  Position `json:"partner_position"`
	VictimPosition   Position `json:"victim_position"`
	CrossfireAngle   float64  `json:"crossfire_angle"`   // Angle between killer and partner view directions
	PartnerDistance  float64  `json:"partner_distance"`  // Distance between killer and partner
	KillerViewAngle  float64  `json:"killer_view_angle"` // Killer's view direction
	PartnerViewAngle float64  `json:"partner_view_angle"`
	KillerSide       string   `json:"killer_side"` // "T" or "CT"
	MapName          string   `json:"map_name"`
	Zone             string   `json:"zone"` // Map zone where the kill occurred
}

// FlashKillEvent represents a kill that occurred while the victim was flashed
// by a teammate of the killer.
type FlashKillEvent struct {
	RoundNumber     int      `json:"round_number"`
	TimeInRound     float64  `json:"time_in_round"`
	KillerSteamID   string   `json:"killer_steam_id"`
	KillerName      string   `json:"killer_name"`
	FlasherSteamID  string   `json:"flasher_steam_id"`
	FlasherName     string   `json:"flasher_name"`
	VictimSteamID   string   `json:"victim_steam_id"`
	VictimName      string   `json:"victim_name"`
	FlashDuration   float64  `json:"flash_duration"`  // How long the victim was flashed
	TimeFromFlash   float64  `json:"time_from_flash"` // Time between flash and kill
	KillerPosition  Position `json:"killer_position"`
	FlasherPosition Position `json:"flasher_position"`
	VictimPosition  Position `json:"victim_position"`
	FlashPosition   Position `json:"flash_position"` // Where the flash detonated
	KillerSide      string   `json:"killer_side"`
	MapName         string   `json:"map_name"`
	Zone            string   `json:"zone"`
	WasFullBlind    bool     `json:"was_full_blind"` // Victim was fully blinded (>2s duration)
	WasPopFlash     bool     `json:"was_pop_flash"`  // Flash detonated within 0.5s of throw
}

// CrossfireFlashData holds all crossfire and flash events for a match.
type CrossfireFlashData struct {
	MapName         string           `json:"map_name"`
	MatchID         string           `json:"match_id"`
	CrossfireEvents []CrossfireEvent `json:"crossfire_events"`
	FlashKillEvents []FlashKillEvent `json:"flash_kill_events"`
}

// NewCrossfireFlashData creates a new CrossfireFlashData instance.
func NewCrossfireFlashData(mapName, matchID string) *CrossfireFlashData {
	return &CrossfireFlashData{
		MapName:         mapName,
		MatchID:         matchID,
		CrossfireEvents: make([]CrossfireEvent, 0),
		FlashKillEvents: make([]FlashKillEvent, 0),
	}
}

// AddCrossfireEvent adds a crossfire event to the data.
func (d *CrossfireFlashData) AddCrossfireEvent(event CrossfireEvent) {
	d.CrossfireEvents = append(d.CrossfireEvents, event)
}

// AddFlashKillEvent adds a flash kill event to the data.
func (d *CrossfireFlashData) AddFlashKillEvent(event FlashKillEvent) {
	d.FlashKillEvents = append(d.FlashKillEvents, event)
}

// GetCrossfiresByRound returns all crossfire events for a specific round.
func (d *CrossfireFlashData) GetCrossfiresByRound(roundNumber int) []CrossfireEvent {
	var events []CrossfireEvent
	for _, e := range d.CrossfireEvents {
		if e.RoundNumber == roundNumber {
			events = append(events, e)
		}
	}
	return events
}

// GetFlashKillsByRound returns all flash kill events for a specific round.
func (d *CrossfireFlashData) GetFlashKillsByRound(roundNumber int) []FlashKillEvent {
	var events []FlashKillEvent
	for _, e := range d.FlashKillEvents {
		if e.RoundNumber == roundNumber {
			events = append(events, e)
		}
	}
	return events
}

// GetPlayerCrossfires returns all crossfire events where the player was the killer.
func (d *CrossfireFlashData) GetPlayerCrossfires(steamID string) []CrossfireEvent {
	var events []CrossfireEvent
	for _, e := range d.CrossfireEvents {
		if e.KillerSteamID == steamID {
			events = append(events, e)
		}
	}
	return events
}

// GetPlayerFlashAssists returns all flash kill events where the player was the flasher.
func (d *CrossfireFlashData) GetPlayerFlashAssists(steamID string) []FlashKillEvent {
	var events []FlashKillEvent
	for _, e := range d.FlashKillEvents {
		if e.FlasherSteamID == steamID {
			events = append(events, e)
		}
	}
	return events
}
