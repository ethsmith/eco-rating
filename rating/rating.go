package rating

import (
	"eco-rating/model"
	"math"
)

// ComputeFinalRating calculates a rating with 60% output / 40% cost balance
// Output: kills, ADR, multi-kills, KAST, opening duels, trades, utility, swing
// Cost: deaths, eco deaths, early deaths, untraded deaths, team flashes, failed clutches
func ComputeFinalRating(p *model.PlayerStats) float64 {
	rounds := float64(p.RoundsPlayed)
	if rounds == 0 {
		return 0
	}

	// ==================== OUTPUT COMPONENTS (60%) ====================

	// === Kill Rating (20%) ===
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

	// === ADR Rating (14%) ===
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
	multiKillBonus := float64(sumMulti(p.MultiKills)) / rounds
	multiKillRating := math.Min(math.Pow(multiKillBonus/BaselineMultiKill, 0.8), 2.0)
	overallPerformance := (ecoKPR/BaselineKPR + (adr / BaselineADR) + p.KAST/BaselineKAST) / 3.0
	if multiKillRating > 1.0 {
		penaltyFactor := math.Pow(math.Min(1.0, overallPerformance), 2)
		multiKillRating = 1.0 + (multiKillRating-1.0)*penaltyFactor
	}

	// === KAST Rating (6%) ===
	kastRatio := p.KAST / BaselineKAST
	var kastRating float64
	if kastRatio >= 1.2 {
		kastRating = 1.0 + (kastRatio-1.0)*0.6
	} else if kastRatio >= 0.9 {
		kastRating = kastRatio
	} else {
		kastRating = math.Pow(kastRatio, 1.2)
	}

	// === Opening Duel Rating (5%) ===
	openingRating := 1.0
	if p.OpeningAttempts > 0 {
		successRate := float64(p.OpeningSuccesses) / float64(p.OpeningAttempts)
		successRatio := successRate / BaselineOpeningSuccessRate
		winConversion := 0.0
		if p.OpeningSuccesses > 0 {
			winConversion = float64(p.RoundsWonAfterOpening) / float64(p.OpeningSuccesses)
		}
		openingRating = successRatio*0.7 + winConversion*0.6
		openingRating = math.Max(0.5, math.Min(1.6, openingRating))
	}

	// === Trade Efficiency Rating (4%) ===
	tradeRating := 1.0
	tradeKillsPerRound := float64(p.TradeKills) / rounds
	tradeKillRatio := tradeKillsPerRound / BaselineTradeKillsPerRound
	tradeRating += (tradeKillRatio - 1.0) * 0.3
	if p.Deaths > 0 {
		tradedPct := float64(p.TradedDeaths) / float64(p.Deaths)
		tradeRating += tradedPct * 0.2
	}
	savesPerRound := float64(p.SavedTeammate) / rounds
	tradeRating += savesPerRound * 1.5
	tradeRating = math.Max(0.6, math.Min(1.5, tradeRating))

	// === Utility Rating (2%) ===
	utilDmgPerRound := float64(p.UtilityDamage) / rounds
	utilDmgRatio := utilDmgPerRound / BaselineUtilityDamage
	flashAssistsPerRound := float64(p.FlashAssists) / rounds
	flashAssistRatio := flashAssistsPerRound / BaselineFlashAssists
	enemyFlashPerRound := p.EnemyFlashDuration / rounds
	enemyFlashRatio := enemyFlashPerRound / BaselineEnemyFlashDur
	utilityScore := (utilDmgRatio*0.4 + flashAssistRatio*0.3 + enemyFlashRatio*0.3)
	utilityRating := 0.7 + utilityScore*0.3
	utilityRating = math.Max(0.5, math.Min(1.4, utilityRating))

	// === Round Swing Rating (1%) ===
	avgSwing := p.RoundSwing / rounds
	var swingRating float64
	if avgSwing >= 0.05 {
		swingRating = 1.0 + (avgSwing/0.15)*0.4
	} else if avgSwing >= 0 {
		swingRating = 1.0 + (avgSwing/0.10)*0.2
	} else {
		swingRating = 1.0 + (avgSwing/0.10)*0.3
	}
	swingRating = math.Max(0.6, math.Min(1.4, swingRating))

	// ==================== COST COMPONENTS (40%) ====================

	// === Death Rating (12%) ===
	dpr := float64(p.Deaths) / rounds
	deathRatio := dpr / BaselineDPR
	var deathRating float64
	if deathRatio <= 0.5 {
		deathRating = 1.9
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

	// === Eco Death Rating (10%) ===
	// High eco death value = dying with expensive equipment = bad
	ecoDeathPerRound := p.EcoDeathValue / rounds
	ecoDeathRatio := ecoDeathPerRound / BaselineEcoDeathValue
	var ecoDeathRating float64
	if ecoDeathRatio <= 0.5 {
		ecoDeathRating = 1.5 // Low eco deaths = good
	} else if ecoDeathRatio <= 1.0 {
		ecoDeathRating = 1.5 - (ecoDeathRatio * 0.5)
	} else if ecoDeathRatio <= 1.5 {
		ecoDeathRating = 1.0 - (ecoDeathRatio-1.0)*0.4
	} else {
		ecoDeathRating = 0.8 - (ecoDeathRatio-1.5)*0.2
	}
	ecoDeathRating = math.Max(0.4, math.Min(1.5, ecoDeathRating))

	// === Early Death Rating (8%) ===
	// Early deaths in rounds hurt the team
	earlyDeathsPerRound := float64(p.EarlyDeaths) / rounds
	earlyDeathRatio := earlyDeathsPerRound / BaselineEarlyDeaths
	var earlyDeathRating float64
	if earlyDeathRatio <= 0.5 {
		earlyDeathRating = 1.4
	} else if earlyDeathRatio <= 1.0 {
		earlyDeathRating = 1.4 - (earlyDeathRatio * 0.4)
	} else if earlyDeathRatio <= 2.0 {
		earlyDeathRating = 1.0 - (earlyDeathRatio-1.0)*0.3
	} else {
		earlyDeathRating = 0.7 - (earlyDeathRatio-2.0)*0.1
	}
	earlyDeathRating = math.Max(0.4, math.Min(1.4, earlyDeathRating))

	// === Untraded Opening Death Rating (5%) ===
	// Opening deaths that weren't traded are very costly
	untradedOpenings := float64(p.OpeningDeaths - p.OpeningDeathsTraded)
	if untradedOpenings < 0 {
		untradedOpenings = 0
	}
	untradedPerRound := untradedOpenings / rounds
	untradedRatio := untradedPerRound / BaselineUntradedOpenings
	var untradedDeathRating float64
	if untradedRatio <= 0.5 {
		untradedDeathRating = 1.3
	} else if untradedRatio <= 1.0 {
		untradedDeathRating = 1.3 - (untradedRatio * 0.3)
	} else if untradedRatio <= 2.0 {
		untradedDeathRating = 1.0 - (untradedRatio-1.0)*0.25
	} else {
		untradedDeathRating = 0.75 - (untradedRatio-2.0)*0.1
	}
	untradedDeathRating = math.Max(0.5, math.Min(1.3, untradedDeathRating))

	// === Team Flash Rating (2%) ===
	// Flashing teammates is bad
	teamFlashPerRound := float64(p.TeamFlashCount) / rounds
	teamFlashRatio := teamFlashPerRound / BaselineTeamFlashPerRound
	var teamFlashRating float64
	if teamFlashRatio <= 0.5 {
		teamFlashRating = 1.2 // Few team flashes = good
	} else if teamFlashRatio <= 1.0 {
		teamFlashRating = 1.2 - (teamFlashRatio * 0.2)
	} else if teamFlashRatio <= 2.0 {
		teamFlashRating = 1.0 - (teamFlashRatio-1.0)*0.2
	} else {
		teamFlashRating = 0.8 - (teamFlashRatio-2.0)*0.1
	}
	teamFlashRating = math.Max(0.5, math.Min(1.2, teamFlashRating))

	// === Failed Clutch Rating (3%) ===
	// Failed clutches when you had a chance
	var failedClutchRating float64 = 1.0
	if p.ClutchRounds > 0 {
		failedClutches := p.ClutchRounds - p.ClutchWins
		failRate := float64(failedClutches) / float64(p.ClutchRounds)
		failRatio := failRate / BaselineFailedClutchRate
		if failRatio <= 0.5 {
			failedClutchRating = 1.4 // Win most clutches = good
		} else if failRatio <= 1.0 {
			failedClutchRating = 1.4 - (failRatio * 0.4)
		} else if failRatio <= 1.5 {
			failedClutchRating = 1.0 - (failRatio-1.0)*0.3
		} else {
			failedClutchRating = 0.85 - (failRatio-1.5)*0.1
		}
		failedClutchRating = math.Max(0.6, math.Min(1.4, failedClutchRating))
	}

	// ==================== COMBINE COMPONENTS ====================

	// Output components (60%)
	outputRating := killRating*WeightKillRating +
		adrRating*WeightADRRating +
		multiKillRating*WeightMultiKillRating +
		kastRating*WeightKASTRating +
		openingRating*WeightOpeningRating +
		tradeRating*WeightTradeRating +
		utilityRating*WeightUtilityRating +
		swingRating*WeightSwingRating

	// Cost components (40%)
	costRating := deathRating*WeightDeathRating +
		ecoDeathRating*WeightEcoDeathRating +
		earlyDeathRating*WeightEarlyDeathRating +
		untradedDeathRating*WeightUntradedDeathRating +
		teamFlashRating*WeightTeamFlashRating +
		failedClutchRating*WeightFailedClutchRating

	rating := outputRating + costRating

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
// Note: Per-side rating uses simplified formula without opening/trade/utility components
// since those stats aren't tracked per-side currently
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

	// === Component 2: Death Rating (16%) ===
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

	// === Component 4: Round Swing Rating (10%) ===
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

	// === Component 5: Multi-Kill Rating (10%) ===
	multiKillBonus := float64(sumMulti(multiKills)) / roundsF
	multiKillRating := math.Min(math.Pow(multiKillBonus/BaselineMultiKill, 0.8), 2.0)

	kastPct := kast / roundsF
	overallPerformance := (ecoKPR/BaselineKPR + (adr / BaselineADR) + kastPct/BaselineKAST) / 3.0
	if multiKillRating > 1.0 {
		penaltyFactor := math.Pow(math.Min(1.0, overallPerformance), 2)
		multiKillRating = 1.0 + (multiKillRating-1.0)*penaltyFactor
	}

	// === Component 6: KAST Rating (6%) ===
	kastRatio := kastPct / BaselineKAST
	var kastRating float64
	if kastRatio >= 1.2 {
		kastRating = 1.0 + (kastRatio-1.0)*0.6
	} else if kastRatio >= 0.9 {
		kastRating = kastRatio
	} else {
		kastRating = math.Pow(kastRatio, 1.2)
	}

	// === Proportional Clutch Modifier ===
	clutchModifier := 0.0
	if clutchRounds > 0 {
		clutchWinRate := float64(clutchWins) / float64(clutchRounds)
		if clutchWinRate < 0.3 {
			clutchModifier = -float64(clutchRounds) * (0.3 - clutchWinRate) * 0.04
		} else {
			clutchModifier = float64(clutchWins) * 0.015
		}
	}

	// === Combine Components ===
	// Per-side uses adjusted weights (opening/trade/utility default to 1.0)
	// Redistributed: Kill 32%, Death 18%, ADR 20%, Swing 12%, Multi 12%, KAST 6%
	rating := killRating*0.32 +
		deathRating*0.18 +
		adrRating*0.20 +
		swingRating*0.12 +
		multiKillRating*0.12 +
		kastRating*0.06 +
		clutchModifier

	return math.Max(MinRating, math.Min(MaxRating, rating))
}
