package rating

import (
	"eco-rating/model"
	"math"
)

// ComputeFinalRating calculates a rating using Round Swing as the primary evaluator
// Round Swing (35%) captures per-round evaluation of all cost/output factors
// K/D Differential (15%) directly rewards positive K/D and punishes negative K/D
// Core stats (50%) provide baseline kill/damage/death metrics
func ComputeFinalRating(p *model.PlayerStats) float64 {
	rounds := float64(p.RoundsPlayed)
	if rounds == 0 {
		return 0
	}

	// ==================== CORE OUTPUT METRICS (30%) ====================

	// === Kill Rating (12%) ===
	// Eco-adjusted kills per round
	ecoKPR := p.EcoKillValue / rounds
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

	// === ADR Rating (12%) ===
	adr := float64(p.Damage) / rounds
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

	// === Multi-Kill Rating (8%) ===
	multiKillBonus := float64(sumMulti(p.MultiKillsRaw)) / rounds
	multiKillRating := math.Min(math.Pow(multiKillBonus/BaselineMultiKill, 0.8), 2.0)
	overallPerformance := (ecoKPR/BaselineKPR + (adr / BaselineADR) + p.KAST/BaselineKAST) / 3.0
	if multiKillRating > 1.0 {
		penaltyFactor := math.Pow(math.Min(1.0, overallPerformance), 2)
		multiKillRating = 1.0 + (multiKillRating-1.0)*penaltyFactor
	}

	// ==================== CORE COST METRIC (20%) ====================

	// === Death Rating (20%) ===
	// Survival is critical - deaths hurt the team
	dpr := float64(p.Deaths) / rounds
	deathRatio := dpr / BaselineDPR
	var deathRating float64
	if deathRatio <= 0.5 {
		deathRating = 1.8 // Exceptional survival
	} else if deathRatio <= 0.8 {
		deathRating = 1.6 - (deathRatio * 0.4)
	} else if deathRatio <= 1.0 {
		deathRating = 1.3 - (deathRatio * 0.3)
	} else if deathRatio <= 1.3 {
		deathRating = 1.0 / math.Pow(deathRatio, 1.0)
	} else {
		deathRating = 1.0 / math.Pow(deathRatio, 1.3)
	}
	deathRating = math.Max(0.3, math.Min(1.8, deathRating))

	// ==================== K/D DIFFERENTIAL (15%) ====================
	// Direct reward/penalty for kills vs deaths
	// Positive K/D = rating boost, Negative K/D = rating penalty
	kdDiff := float64(p.Kills-p.Deaths) / rounds // K/D differential per round
	var kdRating float64
	if kdDiff >= 0.15 {
		// Strongly positive K/D: significant boost
		kdRating = 1.3 + (kdDiff-0.15)*2.0
	} else if kdDiff >= 0.05 {
		// Positive K/D: moderate boost
		kdRating = 1.1 + (kdDiff-0.05)*2.0
	} else if kdDiff >= 0 {
		// Slightly positive K/D: small boost
		kdRating = 1.0 + kdDiff*2.0
	} else if kdDiff >= -0.05 {
		// Slightly negative K/D: small penalty
		kdRating = 1.0 + kdDiff*2.0
	} else if kdDiff >= -0.15 {
		// Negative K/D: moderate penalty
		kdRating = 0.9 + (kdDiff+0.05)*2.0
	} else {
		// Strongly negative K/D: significant penalty
		kdRating = 0.7 + (kdDiff+0.15)*1.0
	}
	kdRating = math.Max(0.4, math.Min(1.6, kdRating))

	// ==================== ROUND SWING (35%) ====================
	// Round Swing already captures per-round evaluation of:
	// - Eco deaths (dying with expensive equipment)
	// - Early deaths (death timing penalty)
	// - Untraded opening deaths
	// - Trade efficiency (trade kills, being traded)
	// - Flash assists and utility impact
	// - Clutch performance (wins and failures)
	// - Team flashes (penalty)
	// - Weapon economy (AWP losses, etc.)
	// - Opening duels, bomb plants/defuses
	// - Multi-kills, exit frags
	//
	// Positive swing = net positive contribution to rounds
	// Negative swing = net negative contribution to rounds

	avgSwing := p.RoundSwing / rounds

	// Convert average swing to a rating component
	// Baseline swing is ~0 (neutral), range typically -0.15 to +0.15
	// We want to map this to a rating multiplier around 1.0
	var swingRating float64
	if avgSwing >= 0.10 {
		// Exceptional positive swing: strong boost
		swingRating = 1.3 + (avgSwing-0.10)*2.0
	} else if avgSwing >= 0.05 {
		// Good positive swing: moderate boost
		swingRating = 1.15 + (avgSwing-0.05)*3.0
	} else if avgSwing >= 0 {
		// Slight positive swing: small boost
		swingRating = 1.0 + avgSwing*3.0
	} else if avgSwing >= -0.05 {
		// Slight negative swing: small penalty
		swingRating = 1.0 + avgSwing*3.0
	} else if avgSwing >= -0.10 {
		// Moderate negative swing: stronger penalty
		swingRating = 0.85 + (avgSwing+0.05)*3.0
	} else {
		// Severe negative swing: heavy penalty
		swingRating = 0.70 + (avgSwing+0.10)*1.5
	}
	swingRating = math.Max(0.4, math.Min(1.6, swingRating))

	// ==================== COMBINE COMPONENTS ====================

	rating := killRating*WeightKillRating +
		adrRating*WeightADRRating +
		multiKillRating*WeightMultiKillRating +
		deathRating*WeightDeathRating +
		kdRating*WeightKDRating +
		swingRating*WeightSwingRating

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
// Uses the same simplified formula as ComputeFinalRating with Round Swing as primary evaluator
func ComputeSideRating(rounds int, kills int, deaths int, damage int, ecoKillValue float64,
	roundSwing float64, kast float64, multiKills [6]int, clutchRounds int, clutchWins int) float64 {

	roundsF := float64(rounds)
	if roundsF == 0 {
		return 0
	}

	// === Kill Rating (15%) ===
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

	// === ADR Rating (12%) ===
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

	// === Multi-Kill Rating (8%) ===
	multiKillBonus := float64(sumMulti(multiKills)) / roundsF
	multiKillRating := math.Min(math.Pow(multiKillBonus/BaselineMultiKill, 0.8), 2.0)
	kastPct := kast / roundsF
	overallPerformance := (ecoKPR/BaselineKPR + (adr / BaselineADR) + kastPct/BaselineKAST) / 3.0
	if multiKillRating > 1.0 {
		penaltyFactor := math.Pow(math.Min(1.0, overallPerformance), 2)
		multiKillRating = 1.0 + (multiKillRating-1.0)*penaltyFactor
	}

	// === Death Rating (20%) ===
	dpr := float64(deaths) / roundsF
	deathRatio := dpr / BaselineDPR
	var deathRating float64
	if deathRatio <= 0.5 {
		deathRating = 1.8
	} else if deathRatio <= 0.8 {
		deathRating = 1.6 - (deathRatio * 0.4)
	} else if deathRatio <= 1.0 {
		deathRating = 1.3 - (deathRatio * 0.3)
	} else if deathRatio <= 1.3 {
		deathRating = 1.0 / math.Pow(deathRatio, 1.0)
	} else {
		deathRating = 1.0 / math.Pow(deathRatio, 1.3)
	}
	deathRating = math.Max(0.3, math.Min(1.8, deathRating))

	// === K/D Differential Rating (15%) ===
	kdDiff := float64(kills-deaths) / roundsF
	var kdRating float64
	if kdDiff >= 0.15 {
		kdRating = 1.3 + (kdDiff-0.15)*2.0
	} else if kdDiff >= 0.05 {
		kdRating = 1.1 + (kdDiff-0.05)*2.0
	} else if kdDiff >= 0 {
		kdRating = 1.0 + kdDiff*2.0
	} else if kdDiff >= -0.05 {
		kdRating = 1.0 + kdDiff*2.0
	} else if kdDiff >= -0.15 {
		kdRating = 0.9 + (kdDiff+0.05)*2.0
	} else {
		kdRating = 0.7 + (kdDiff+0.15)*1.0
	}
	kdRating = math.Max(0.4, math.Min(1.6, kdRating))

	// === Round Swing Rating (35%) ===
	avgSwing := roundSwing / roundsF
	var swingRating float64
	if avgSwing >= 0.10 {
		swingRating = 1.3 + (avgSwing-0.10)*2.0
	} else if avgSwing >= 0.05 {
		swingRating = 1.15 + (avgSwing-0.05)*3.0
	} else if avgSwing >= 0 {
		swingRating = 1.0 + avgSwing*3.0
	} else if avgSwing >= -0.05 {
		swingRating = 1.0 + avgSwing*3.0
	} else if avgSwing >= -0.10 {
		swingRating = 0.85 + (avgSwing+0.05)*3.0
	} else {
		swingRating = 0.70 + (avgSwing+0.10)*1.5
	}
	swingRating = math.Max(0.4, math.Min(1.6, swingRating))

	// === Combine Components ===
	rating := killRating*WeightKillRating +
		adrRating*WeightADRRating +
		multiKillRating*WeightMultiKillRating +
		deathRating*WeightDeathRating +
		kdRating*WeightKDRating +
		swingRating*WeightSwingRating

	return math.Max(MinRating, math.Min(MaxRating, rating))
}
