// Package parser provides CS2 demo file parsing functionality.
// This file implements flash tracking for detailed flash-kill event analysis.
package parser

import (
	"github.com/ethsmith/eco-rating/model"
)

const (
	// FlashKillWindow is the time window (in seconds) after a flash during which
	// a kill is considered a "flash kill" (assisted by the flash).
	FlashKillWindow = 2.0

	// PopFlashWindow is the time window (in seconds) for a flash to be considered
	// a "pop flash" (detonates very quickly after being thrown).
	PopFlashWindow = 0.5

	// FullBlindThreshold is the minimum flash duration (in seconds) to be considered
	// fully blinded.
	FullBlindThreshold = 2.0
)

// FlashEvent represents a flash that affected an enemy player.
type FlashEvent struct {
	FlasherID       uint64
	FlasherName     string
	VictimID        uint64
	VictimName      string
	FlashDuration   float64
	TimeInRound     float64
	FlasherPosition model.Position
	FlashPosition   model.Position // Where the flash detonated
	VictimPosition  model.Position
	FlasherSide     string
	WasPopFlash     bool
}

// FlashKillTracker tracks flash events and correlates them with kills.
type FlashKillTracker struct {
	// activeFlashes maps victim SteamID -> list of recent flash events affecting them
	activeFlashes map[uint64][]FlashEvent

	// flashThrowTimes maps flash entity ID -> throw time (for pop flash detection)
	flashThrowTimes map[int64]float64

	// flashThrowers maps flash entity ID -> thrower info
	flashThrowers map[int64]FlashThrowerInfo
}

// FlashThrowerInfo stores information about who threw a flash.
type FlashThrowerInfo struct {
	ThrowerID   uint64
	ThrowerName string
	ThrowTime   float64
	Position    model.Position
	Side        string
}

// NewFlashKillTracker creates a new flash kill tracker.
func NewFlashKillTracker() *FlashKillTracker {
	return &FlashKillTracker{
		activeFlashes:   make(map[uint64][]FlashEvent),
		flashThrowTimes: make(map[int64]float64),
		flashThrowers:   make(map[int64]FlashThrowerInfo),
	}
}

// Reset clears all tracking data for a new round.
func (ft *FlashKillTracker) Reset() {
	ft.activeFlashes = make(map[uint64][]FlashEvent)
	ft.flashThrowTimes = make(map[int64]float64)
	ft.flashThrowers = make(map[int64]FlashThrowerInfo)
}

// RecordFlashThrow records when a flash grenade is thrown.
func (ft *FlashKillTracker) RecordFlashThrow(entityID int64, throwerID uint64, throwerName string,
	timeInRound float64, position model.Position, side string) {
	ft.flashThrowTimes[entityID] = timeInRound
	ft.flashThrowers[entityID] = FlashThrowerInfo{
		ThrowerID:   throwerID,
		ThrowerName: throwerName,
		ThrowTime:   timeInRound,
		Position:    position,
		Side:        side,
	}
}

// RecordFlashDetonate records when a flash detonates and affects a player.
func (ft *FlashKillTracker) RecordFlashDetonate(flasherID uint64, flasherName string,
	victimID uint64, victimName string, flashDuration float64, timeInRound float64,
	flasherPos, flashPos, victimPos model.Position, flasherSide string, entityID int64) {

	// Determine if this was a pop flash
	wasPopFlash := false
	if throwTime, ok := ft.flashThrowTimes[entityID]; ok {
		if timeInRound-throwTime <= PopFlashWindow {
			wasPopFlash = true
		}
	}

	event := FlashEvent{
		FlasherID:       flasherID,
		FlasherName:     flasherName,
		VictimID:        victimID,
		VictimName:      victimName,
		FlashDuration:   flashDuration,
		TimeInRound:     timeInRound,
		FlasherPosition: flasherPos,
		FlashPosition:   flashPos,
		VictimPosition:  victimPos,
		FlasherSide:     flasherSide,
		WasPopFlash:     wasPopFlash,
	}

	ft.activeFlashes[victimID] = append(ft.activeFlashes[victimID], event)
}

// GetFlashAssistsForKill returns flash events that could have assisted a kill.
// Returns flashes that occurred within FlashKillWindow seconds before the kill.
func (ft *FlashKillTracker) GetFlashAssistsForKill(victimID uint64, killerID uint64,
	killTime float64) []FlashEvent {

	var assists []FlashEvent

	if flashes, ok := ft.activeFlashes[victimID]; ok {
		for _, flash := range flashes {
			// Check if flash is within the kill window
			timeSinceFlash := killTime - flash.TimeInRound
			if timeSinceFlash >= 0 && timeSinceFlash <= FlashKillWindow {
				// Only count flashes from the killer's teammates (not the killer themselves)
				if flash.FlasherID != killerID {
					assists = append(assists, flash)
				}
			}
		}
	}

	return assists
}

// GetBestFlashAssist returns the most significant flash assist for a kill.
// Prioritizes longer flash durations and more recent flashes.
func (ft *FlashKillTracker) GetBestFlashAssist(victimID uint64, killerID uint64,
	killTime float64) *FlashEvent {

	assists := ft.GetFlashAssistsForKill(victimID, killerID, killTime)
	if len(assists) == 0 {
		return nil
	}

	// Find the best assist (longest duration, most recent)
	var best *FlashEvent
	bestScore := 0.0

	for i := range assists {
		flash := &assists[i]
		timeSinceFlash := killTime - flash.TimeInRound
		// Score based on duration and recency
		score := flash.FlashDuration * (1.0 - timeSinceFlash/FlashKillWindow)
		if score > bestScore {
			bestScore = score
			best = flash
		}
	}

	return best
}

// ClearVictimFlashes removes flash data for a dead player.
func (ft *FlashKillTracker) ClearVictimFlashes(victimID uint64) {
	delete(ft.activeFlashes, victimID)
}

// CleanupOldFlashes removes flash events older than the kill window.
func (ft *FlashKillTracker) CleanupOldFlashes(currentTime float64) {
	for victimID, flashes := range ft.activeFlashes {
		var validFlashes []FlashEvent
		for _, flash := range flashes {
			if currentTime-flash.TimeInRound <= FlashKillWindow {
				validFlashes = append(validFlashes, flash)
			}
		}
		if len(validFlashes) > 0 {
			ft.activeFlashes[victimID] = validFlashes
		} else {
			delete(ft.activeFlashes, victimID)
		}
	}
}

// IsFullBlind returns true if the flash duration indicates full blindness.
func IsFullBlind(duration float64) bool {
	return duration >= FullBlindThreshold
}
