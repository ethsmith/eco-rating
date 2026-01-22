// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// There may be mistakes in the comments. Please verify accuracy.
// =============================================================================

// Package swing implements probability-based player impact calculation.
// This file contains the Attributor which distributes swing credit among
// players who contributed to a kill (killer, damage dealers, flash assists).
package swing

import "math"

// Credit attribution constants for kill swing distribution.
const (
	// KillerBaseCredit is the base credit given to the player who got the kill.
	KillerBaseCredit = 0.60

	// DamageShareCredit is the portion of credit distributed based on damage share.
	DamageShareCredit = 0.25

	// FlashAssistMaxCredit is the maximum credit for flash assists.
	FlashAssistMaxCredit = 0.15

	// TradeKillPenalty reduces credit for trade kills (kill was "expected").
	TradeKillPenalty = 0.30

	// PlantCreditShare is the portion of plant swing given to the planter.
	PlantCreditShare = 0.60

	// DefuseCreditShare is the portion of defuse swing given to the defuser.
	DefuseCreditShare = 0.80

	// SavePenalty is the small negative swing for surviving a lost round.
	SavePenalty = 0.02
)

// Attributor handles the distribution of kill credit among contributors.
type Attributor struct{}

// NewAttributor creates a new Attributor.
func NewAttributor() *Attributor {
	return &Attributor{}
}

// AttributeKillCredit distributes swing credit for a kill among all contributors.
func (a *Attributor) AttributeKillCredit(
	playerSwing map[uint64]float64,
	kill *KillEvent,
	delta float64,
) {
	// Calculate killer's credit
	killerCredit := a.calculateKillerCredit(kill)

	// Apply trade kill penalty
	if kill.IsTradeKill {
		killerCredit *= (1.0 - TradeKillPenalty)
	}

	// Award killer
	playerSwing[kill.KillerID] += delta * killerCredit

	// Award damage contributors (excluding killer)
	a.attributeDamageContributors(playerSwing, kill, delta)

	// Award flash assists
	a.attributeFlashAssists(playerSwing, kill, delta)
}

// calculateKillerCredit computes the killer's share of credit.
func (a *Attributor) calculateKillerCredit(kill *KillEvent) float64 {
	credit := KillerBaseCredit

	// Add damage share bonus if killer dealt significant damage
	if kill.TotalDamageToVictim > 0 && kill.KillerDamageDealt > 0 {
		damageRatio := float64(kill.KillerDamageDealt) / float64(kill.TotalDamageToVictim)
		credit += DamageShareCredit * damageRatio
	}

	return credit
}

// attributeDamageContributors distributes credit to players who dealt damage.
func (a *Attributor) attributeDamageContributors(
	playerSwing map[uint64]float64,
	kill *KillEvent,
	delta float64,
) {
	if kill.TotalDamageToVictim <= 0 || len(kill.DamageContributors) == 0 {
		return
	}

	for _, contributor := range kill.DamageContributors {
		// Skip the killer (they already got credit)
		if contributor.PlayerID == kill.KillerID {
			continue
		}

		// Calculate damage share
		share := float64(contributor.Damage) / float64(kill.TotalDamageToVictim)

		// Award proportional credit
		playerSwing[contributor.PlayerID] += delta * DamageShareCredit * share
	}
}

// attributeFlashAssists distributes credit to players who flash assisted.
func (a *Attributor) attributeFlashAssists(
	playerSwing map[uint64]float64,
	kill *KillEvent,
	delta float64,
) {
	for _, flash := range kill.FlashAssists {
		// Flash credit scales with duration (up to 3 seconds for full credit)
		flashCredit := math.Min(flash.Duration/3.0, 1.0) * FlashAssistMaxCredit

		playerSwing[flash.PlayerID] += delta * flashCredit
	}
}
