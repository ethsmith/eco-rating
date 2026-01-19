// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// =============================================================================

// Package rating implements the eco-rating calculation system.
// This file contains the core rating computation functions that combine
// various performance metrics into a final player rating.
package rating

import (
	"eco-rating/model"
	"math"
)

// ComputeFinalRating calculates the overall eco-rating for a player.
// The rating is based on a 1.0 baseline with contributions from:
// - KPR (kills per round) - asymmetric scaling favoring high performers
// - DPR (deaths per round) - penalizes high death rates
// - ADR (average damage per round) - rewards consistent damage output
// - KAST percentage - rewards round involvement
// - Round swing - measures actual impact on round outcomes
// - Impact metrics (opening kills, multi-kills)
//
// Returns a value typically between 0.20 and 3.00.
func ComputeFinalRating(p *model.PlayerStats) float64 {
	rounds := float64(p.RoundsPlayed)
	if rounds == 0 {
		return 0
	}

	kpr := float64(p.Kills) / rounds
	dpr := float64(p.Deaths) / rounds
	adr := float64(p.Damage) / rounds
	kast := p.KAST
	avgSwing := p.RoundSwing / rounds

	openingKillsPerRound := float64(p.OpeningKills) / rounds
	multiKillRoundsPerRound := float64(p.RoundsWithMultiKill) / rounds

	var kprContrib float64
	if kpr >= 0.7 {
		kprContrib = (kpr - 0.7) * 0.75
	} else {
		kprContrib = (kpr - 0.7) * 0.55
	}

	var dprContrib float64
	if dpr <= 0.7 {
		dprContrib = (0.7 - dpr) * 0.15
	} else {
		dprContrib = (0.7 - dpr) * 0.55
	}

	var adrContrib float64
	if adr >= 75.0 {
		adrContrib = (adr - 75.0) * 0.015
	} else {
		adrContrib = (adr - 75.0) * 0.004
	}

	var kastContrib float64
	if kast >= 0.70 {
		kastContrib = (kast - 0.70) * 0.20
	} else {
		kastContrib = (kast - 0.70) * 0.35
	}

	var swingContrib float64
	if avgSwing >= 0 {
		swingContrib = avgSwing * 0.75
	} else {
		swingContrib = avgSwing * 1.00
	}

	impactContrib := openingKillsPerRound*0.3 + multiKillRoundsPerRound*0.15

	multiKillBonus := float64(sumMulti(p.MultiKillsRaw)) / rounds
	multiContrib := multiKillBonus * 0.015

	rating := 1.0 + kprContrib + dprContrib + adrContrib + kastContrib + swingContrib + impactContrib + multiContrib

	return math.Max(MinRating, math.Min(MaxRating, rating))
}

// sumMulti calculates a weighted sum of multi-kill rounds.
// Higher kill counts receive exponentially higher weights.
func sumMulti(m [6]int) int {
	weights := [6]int{0, 0, 2, 6, 14, 30}
	total := 0
	for i := 2; i <= 5; i++ {
		total += m[i] * weights[i]
	}
	return total
}

// ComputeSideRating calculates a rating for a specific side (T or CT).
// Uses the same formula as ComputeFinalRating but with side-specific stats.
// This allows comparison of player performance on different sides.
func ComputeSideRating(rounds int, kills int, deaths int, damage int, ecoKillValue float64,
	roundSwing float64, kast float64, multiKills [6]int, clutchRounds int, clutchWins int) float64 {

	roundsF := float64(rounds)
	if roundsF == 0 {
		return 0
	}

	kpr := float64(kills) / roundsF
	dpr := float64(deaths) / roundsF
	adr := float64(damage) / roundsF
	kastPct := kast / roundsF
	avgSwing := roundSwing / roundsF

	var kprContrib float64
	if kpr >= 0.7 {
		kprContrib = (kpr - 0.7) * 0.75
	} else {
		kprContrib = (kpr - 0.7) * 0.55
	}

	var dprContrib float64
	if dpr <= 0.7 {
		dprContrib = (0.7 - dpr) * 0.15
	} else {
		dprContrib = (0.7 - dpr) * 0.55
	}

	var adrContrib float64
	if adr >= 75.0 {
		adrContrib = (adr - 75.0) * 0.015
	} else {
		adrContrib = (adr - 75.0) * 0.004
	}

	var kastContrib float64
	if kastPct >= 0.70 {
		kastContrib = (kastPct - 0.70) * 0.20
	} else {
		kastContrib = (kastPct - 0.70) * 0.35
	}

	var swingContrib float64
	if avgSwing >= 0 {
		swingContrib = avgSwing * 0.75
	} else {
		swingContrib = avgSwing * 1.00
	}

	multiKillBonus := float64(sumMulti(multiKills)) / roundsF
	multiContrib := multiKillBonus * 0.015

	rating := 1.0 + kprContrib + dprContrib + adrContrib + kastContrib + swingContrib + multiContrib

	return math.Max(MinRating, math.Min(MaxRating, rating))
}
