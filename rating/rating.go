package rating

import (
	"eco-rating/model"
	"math"
)

// ComputeFinalRating calculates an HLTV 3.0-style rating
// Based on HLTV's 60-40 output/cost balance with recent weight adjustments
func ComputeFinalRating(p *model.PlayerStats) float64 {
	rounds := float64(p.RoundsPlayed)
	if rounds == 0 {
		return 0
	}

	// === Component 1: Kill Rating (28%) ===
	// Eco-adjusted kills per round - primary component per HLTV updates
	ecoKPR := p.EcoKillValue / rounds
	// Enhanced scaling for exceptional fraggers
	killRatio := ecoKPR / BaselineKPR
	var killRating float64
	if killRatio >= 1.5 {
		// Exceptional fraggers: very strong boost (frozen case: 1.51 ratio)
		killRating = 1.0 + (killRatio-1.0)*1.3
	} else if killRatio >= 1.2 {
		// Good fraggers: moderate boost
		killRating = 1.0 + (killRatio-1.0)*0.7
	} else if killRatio >= 0.8 {
		// Average performance: normal scaling
		killRating = math.Pow(killRatio, 0.9)
	} else {
		// Below average: stronger penalty
		killRating = math.Pow(killRatio, 1.1)
	}

	// === Component 2: Death Rating (18%) ===
	// Balanced death penalty - reward low deaths, penalize high deaths
	dpr := float64(p.Deaths) / rounds
	deathRatio := dpr / BaselineDPR
	var deathRating float64
	if deathRatio <= 0.5 {
		// Exceptionally low deaths: very strong reward (frozen: 0.54 ratio)
		deathRating = 2.0 - (deathRatio * 0.2)
	} else if deathRatio <= 0.8 {
		// Very low deaths: strong reward
		deathRating = 1.7 - (deathRatio * 0.4)
	} else if deathRatio <= 1.0 {
		// Below baseline: moderate reward
		deathRating = 1.4 - (deathRatio * 0.3)
	} else if deathRatio <= 1.3 {
		// Above baseline: moderate penalty
		deathRating = 1.0 / math.Pow(deathRatio, 1.0)
	} else {
		// High deaths: stronger penalty
		deathRating = 1.0 / math.Pow(deathRatio, 1.2)
	}
	deathRating = math.Max(0.3, math.Min(1.9, deathRating))

	// === Component 3: ADR Rating (18%) ===
	// Eco-adjusted damage per round - reward high damage dealers
	adr := float64(p.Damage) / rounds
	adrRatio := adr / BaselineADR
	var adrRating float64
	if adrRatio >= 1.4 {
		// Exceptional damage: very strong boost (frozen: 1.46 ratio)
		adrRating = 0.8 + (adrRatio * 0.6)
	} else if adrRatio >= 1.0 {
		// Above baseline: strong scaling for high damage
		adrRating = 0.7 + (adrRatio * 0.5)
	} else if adrRatio >= 0.8 {
		// Below baseline: stronger penalty (ropz case: 0.87 ratio)
		adrRating = 0.4 + (adrRatio * 0.6)
	} else {
		// Low damage: very strong penalty
		adrRating = 0.3 + (adrRatio * 0.5)
	}

	// === Component 4: Round Swing Rating (14%) ===
	// Advanced round swing system - adjust scaling for new range
	avgSwing := p.RoundSwing / rounds
	var swingRating float64

	// New scaling for advanced round swing (range roughly -0.25 to +0.35 per round)
	if avgSwing >= 0.05 {
		// High positive swing: moderate reward (scaled down)
		swingRating = 1.0 + (avgSwing/0.15)*0.4
	} else if avgSwing >= 0 {
		// Low positive swing: small reward
		swingRating = 1.0 + (avgSwing/0.10)*0.2
	} else {
		// Negative swing: penalty (scaled appropriately)
		swingRating = 1.0 + (avgSwing/0.10)*0.3
	}
	swingRating = math.Max(0.6, math.Min(1.4, swingRating))

	// === Component 5: Multi-Kill Rating (12%) ===
	// Separate component for explosive moments - but penalize if overall performance is poor
	multiKillBonus := float64(sumMulti(p.MultiKills)) / rounds
	multiKillRating := math.Min(math.Pow(multiKillBonus/BaselineMultiKill, 0.8), 2.0)

	// Sliding scale: multi-kill bonus proportional to overall performance
	// Prevents stat padding through occasional explosive rounds while rewarding clutch moments proportionally
	overallPerformance := (ecoKPR/BaselineKPR + (adr / BaselineADR) + p.KAST/BaselineKAST) / 3.0
	if multiKillRating > 1.0 {
		penaltyFactor := math.Pow(math.Min(1.0, overallPerformance), 2)
		multiKillRating = 1.0 + (multiKillRating-1.0)*penaltyFactor
	}

	// === Component 6: KAST Rating (8%) ===
	// Consistency metric with penalties for low KAST
	kastRatio := p.KAST / BaselineKAST
	var kastRating float64
	if kastRatio >= 1.2 {
		// Very high KAST: diminishing returns
		kastRating = 1.0 + (kastRatio-1.0)*0.6
	} else if kastRatio >= 0.9 {
		// Good KAST: normal scaling
		kastRating = kastRatio
	} else {
		// Low KAST: stronger penalty
		kastRating = math.Pow(kastRatio, 1.2)
	}

	// === Additional Penalties ===
	// Penalty for failed clutches (clutch attempts with no wins)
	clutchPenalty := 0.0
	if p.ClutchRounds > 0 && p.ClutchWins == 0 {
		// Failed all clutch attempts - penalty scales with attempts
		clutchPenalty = float64(p.ClutchRounds) * 0.02 // -0.02 per failed clutch
	}

	// === Combine Components ===
	// HLTV 3.0 weights: 60% output (kills, damage, multi-kills) + 40% cost/impact (deaths, swing, KAST)
	rating := killRating*WeightKillRating +
		deathRating*WeightDeathRating +
		adrRating*WeightADRRating +
		swingRating*WeightSwingRating +
		multiKillRating*WeightMultiKillRating +
		kastRating*WeightKASTRating -
		clutchPenalty

	// Clamp to reasonable range
	return math.Max(MinRating, math.Min(MaxRating, rating))
}

func sumMulti(m [6]int) int {
	// Weight multi-kills with exponential scaling like HLTV 3.0
	// Higher kill counts are exponentially more valuable
	weights := [6]int{0, 0, 1, 3, 7, 15} // 0, 0, double=1, triple=3, quad=7, ace=15
	total := 0
	for i := 2; i <= 5; i++ {
		total += m[i] * weights[i]
	}
	return total
}

// ComputeSideRating calculates eco rating for a specific side (T or CT)
// Uses the same formula as ComputeFinalRating but with side-specific stats
func ComputeSideRating(rounds int, kills int, deaths int, damage int, ecoKillValue float64,
	roundSwing float64, kast float64, multiKills [6]int, clutchRounds int, clutchWins int) float64 {

	roundsF := float64(rounds)
	if roundsF == 0 {
		return 0
	}

	// === Component 1: Kill Rating (28%) ===
	ecoKPR := ecoKillValue / roundsF
	killRatio := ecoKPR / BaselineKPR
	var killRating float64
	if killRatio >= 1.5 {
		killRating = 1.0 + (killRatio-1.0)*1.3
	} else if killRatio >= 1.2 {
		killRating = 1.0 + (killRatio-1.0)*0.7
	} else if killRatio >= 0.8 {
		killRating = math.Pow(killRatio, 0.9)
	} else {
		killRating = math.Pow(killRatio, 1.1)
	}

	// === Component 2: Death Rating (18%) ===
	dpr := float64(deaths) / roundsF
	deathRatio := dpr / BaselineDPR
	var deathRating float64
	if deathRatio <= 0.5 {
		deathRating = 2.0 - (deathRatio * 0.2)
	} else if deathRatio <= 0.8 {
		deathRating = 1.7 - (deathRatio * 0.4)
	} else if deathRatio <= 1.0 {
		deathRating = 1.4 - (deathRatio * 0.3)
	} else if deathRatio <= 1.3 {
		deathRating = 1.0 / math.Pow(deathRatio, 1.0)
	} else {
		deathRating = 1.0 / math.Pow(deathRatio, 1.2)
	}
	deathRating = math.Max(0.3, math.Min(1.9, deathRating))

	// === Component 3: ADR Rating (18%) ===
	adr := float64(damage) / roundsF
	adrRatio := adr / BaselineADR
	var adrRating float64
	if adrRatio >= 1.4 {
		adrRating = 0.8 + (adrRatio * 0.6)
	} else if adrRatio >= 1.0 {
		adrRating = 0.7 + (adrRatio * 0.5)
	} else if adrRatio >= 0.8 {
		adrRating = 0.4 + (adrRatio * 0.6)
	} else {
		adrRating = 0.3 + (adrRatio * 0.5)
	}

	// === Component 4: Round Swing Rating (14%) ===
	avgSwing := roundSwing / roundsF
	var swingRating float64
	if avgSwing >= 0.05 {
		swingRating = 1.0 + (avgSwing/0.15)*0.4
	} else if avgSwing >= 0 {
		swingRating = 1.0 + (avgSwing/0.10)*0.2
	} else {
		swingRating = 1.0 + (avgSwing/0.10)*0.3
	}
	swingRating = math.Max(0.6, math.Min(1.4, swingRating))

	// === Component 5: Multi-Kill Rating (12%) ===
	multiKillBonus := float64(sumMulti(multiKills)) / roundsF
	multiKillRating := math.Min(math.Pow(multiKillBonus/BaselineMultiKill, 0.8), 2.0)

	kastPct := kast / roundsF
	overallPerformance := (ecoKPR/BaselineKPR + (adr / BaselineADR) + kastPct/BaselineKAST) / 3.0
	if multiKillRating > 1.0 {
		penaltyFactor := math.Pow(math.Min(1.0, overallPerformance), 2)
		multiKillRating = 1.0 + (multiKillRating-1.0)*penaltyFactor
	}

	// === Component 6: KAST Rating (8%) ===
	kastRatio := kastPct / BaselineKAST
	var kastRating float64
	if kastRatio >= 1.2 {
		kastRating = 1.0 + (kastRatio-1.0)*0.6
	} else if kastRatio >= 0.9 {
		kastRating = kastRatio
	} else {
		kastRating = math.Pow(kastRatio, 1.2)
	}

	// === Additional Penalties ===
	clutchPenalty := 0.0
	if clutchRounds > 0 && clutchWins == 0 {
		clutchPenalty = float64(clutchRounds) * 0.02
	}

	// === Combine Components ===
	rating := killRating*WeightKillRating +
		deathRating*WeightDeathRating +
		adrRating*WeightADRRating +
		swingRating*WeightSwingRating +
		multiKillRating*WeightMultiKillRating +
		kastRating*WeightKASTRating -
		clutchPenalty

	return math.Max(MinRating, math.Min(MaxRating, rating))
}
