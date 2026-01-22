// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// There may be mistakes in the comments. Please verify accuracy.
// =============================================================================

// Package parser provides CS2 demo file parsing functionality.
// This file implements damage tracking to attribute kill credit to players
// who dealt damage to a victim before they were killed.
package parser

import (
	"eco-rating/rating/swing"
)

// DamageTracker tracks damage dealt to each player during a round.
// Used to attribute kill credit to damage contributors.
type DamageTracker struct {
	// damageDealt maps victim SteamID -> attacker SteamID -> damage
	damageDealt map[uint64]map[uint64]int

	// flashedPlayers maps victim SteamID -> list of flash assists
	flashedPlayers map[uint64][]FlashInfo
}

// FlashInfo tracks flash duration on a victim from an attacker.
type FlashInfo struct {
	AttackerID uint64
	Duration   float64
}

// NewDamageTracker creates a new damage tracker.
func NewDamageTracker() *DamageTracker {
	return &DamageTracker{
		damageDealt:    make(map[uint64]map[uint64]int),
		flashedPlayers: make(map[uint64][]FlashInfo),
	}
}

// Reset clears all tracking data for a new round.
func (dt *DamageTracker) Reset() {
	dt.damageDealt = make(map[uint64]map[uint64]int)
	dt.flashedPlayers = make(map[uint64][]FlashInfo)
}

// RecordDamage records damage dealt from attacker to victim.
func (dt *DamageTracker) RecordDamage(attackerID, victimID uint64, damage int) {
	if dt.damageDealt[victimID] == nil {
		dt.damageDealt[victimID] = make(map[uint64]int)
	}
	dt.damageDealt[victimID][attackerID] += damage
}

// RecordFlash records that an attacker flashed a victim.
func (dt *DamageTracker) RecordFlash(attackerID, victimID uint64, duration float64) {
	dt.flashedPlayers[victimID] = append(dt.flashedPlayers[victimID], FlashInfo{
		AttackerID: attackerID,
		Duration:   duration,
	})
}

// GetDamageContributors returns all players who damaged a victim and their damage totals.
func (dt *DamageTracker) GetDamageContributors(victimID uint64) []swing.DamageContributor {
	contributors := make([]swing.DamageContributor, 0)

	if damages, ok := dt.damageDealt[victimID]; ok {
		for attackerID, damage := range damages {
			contributors = append(contributors, swing.DamageContributor{
				PlayerID: attackerID,
				Damage:   damage,
			})
		}
	}

	return contributors
}

// GetTotalDamageToVictim returns the total damage dealt to a victim.
func (dt *DamageTracker) GetTotalDamageToVictim(victimID uint64) int {
	total := 0
	if damages, ok := dt.damageDealt[victimID]; ok {
		for _, damage := range damages {
			total += damage
		}
	}
	return total
}

// GetKillerDamage returns the damage the killer dealt to the victim.
func (dt *DamageTracker) GetKillerDamage(killerID, victimID uint64) int {
	if damages, ok := dt.damageDealt[victimID]; ok {
		return damages[killerID]
	}
	return 0
}

// GetFlashAssists returns flash assists for a victim's death.
// Only returns recent flashes (within the flash duration window).
func (dt *DamageTracker) GetFlashAssists(victimID uint64) []swing.FlashAssist {
	assists := make([]swing.FlashAssist, 0)

	if flashes, ok := dt.flashedPlayers[victimID]; ok {
		for _, flash := range flashes {
			// Only count flashes with significant duration
			if flash.Duration >= 0.5 {
				assists = append(assists, swing.FlashAssist{
					PlayerID: flash.AttackerID,
					Duration: flash.Duration,
				})
			}
		}
	}

	return assists
}

// ClearVictimData removes tracking data for a dead player.
// Called after processing a kill to prevent double-counting.
func (dt *DamageTracker) ClearVictimData(victimID uint64) {
	delete(dt.damageDealt, victimID)
	delete(dt.flashedPlayers, victimID)
}
