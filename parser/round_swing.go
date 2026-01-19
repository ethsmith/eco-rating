// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// =============================================================================

// Package parser provides CS2 demo file parsing functionality.
// This file implements the round swing calculation algorithm, which measures
// a player's impact on round outcomes based on their actions, timing, and context.
package parser

import (
	"eco-rating/model"
	"math"
)

// CalculateAdvancedRoundSwing computes a player's impact score for a single round.
// The algorithm considers:
// - Base swing from round outcome and involvement level
// - Performance contribution (kills, damage, assists, survival)
// - Situational bonuses (opening kills, trades, entry fragging)
// - Impact actions (bomb plant/defuse, eco kills)
// - Economy modifiers (equipment advantage/disadvantage)
// - Multi-kill bonuses
// - Clutch performance
// - Utility impact (flashes, grenades)
// - Penalties (team flashes, failed trades, early deaths, exit frags)
//
// Returns a value typically between -0.30 and +0.40 per round.
func CalculateAdvancedRoundSwing(roundStats *model.RoundStats, context *model.RoundContext, playerEquipValue float64, teamEquipValue float64) float64 {
	const baseWin = 0.04
	const baseLoss = -0.04

	involvement := float64(roundStats.Kills)
	involvement += float64(roundStats.Assists) * 0.5
	involvement += math.Min(1.0, float64(roundStats.Damage)/150.0)
	if roundStats.PlantedBomb {
		involvement += 1.0
	}
	if roundStats.DefusedBomb {
		involvement += 1.0
	}
	if roundStats.Survived {
		involvement += 0.5
	}
	involvement = math.Max(0.0, math.Min(involvement/4.0, 1.0))

	var baseSwing float64
	if roundStats.TeamWon {
		baseSwing = baseWin * involvement
	} else {
		baseSwing = baseLoss * involvement
	}

	performanceBonus := calculatePerformanceContribution(roundStats)
	situationalBonus := calculateSituationalBonus(roundStats, context)
	impactBonus := calculateImpactActions(roundStats, context)
	economyModifier := calculateEconomyModifier(roundStats, playerEquipValue, teamEquipValue, context.RoundType)
	multiKillBonus := calculateMultiKillBonus(roundStats)
	clutchModifier := calculateClutchModifier(roundStats)
	utilityBonus := calculateUtilityImpact(roundStats)
	tradeSpeedBonus := calculateTradeSpeedBonus(roundStats)
	exitFragPenalty := calculateExitFragPenalty(roundStats)
	deathTimingPenalty := calculateDeathTimingPenalty(roundStats)
	teamFlashPenalty := calculateTeamFlashPenalty(roundStats)
	failedTradePenalty := calculateFailedTradePenalty(roundStats)
	weaponBonus := calculateWeaponBonus(roundStats)

	totalSwing := baseSwing + performanceBonus + situationalBonus + impactBonus + multiKillBonus + clutchModifier +
		utilityBonus + tradeSpeedBonus - exitFragPenalty - deathTimingPenalty - teamFlashPenalty - failedTradePenalty + weaponBonus

	totalSwing *= economyModifier
	totalSwing *= context.RoundImportance

	return math.Max(-0.30, math.Min(0.40, totalSwing))
}

// calculatePerformanceContribution computes the base performance score from
// kills, damage, assists, and survival status.
func calculatePerformanceContribution(roundStats *model.RoundStats) float64 {
	contribution := 0.0

	if roundStats.Kills > 0 {
		contribution += float64(roundStats.Kills) * 0.04
		if roundStats.Kills >= 2 {
			contribution += float64(roundStats.Kills-1) * 0.02
		}
	}

	effectiveDamage := roundStats.Damage
	if roundStats.IsExitFrag {
		effectiveDamage = int(float64(effectiveDamage) * 0.5)
	}
	damageContrib := float64(effectiveDamage) / 400.0
	contribution += math.Min(damageContrib, 0.08)

	contribution += float64(roundStats.Assists) * 0.015
	contribution += float64(roundStats.FlashAssists) * 0.01

	if roundStats.Survived {
		if roundStats.TeamWon {
			contribution += 0.02
		} else {
			contribution += 0.04
			roundStats.SavedWeapons = true
		}
	} else {
		if roundStats.TradeDeath {
			contribution -= 0.04
		} else {
			contribution -= 0.08
		}
	}

	return contribution
}

// calculateSituationalBonus adds bonuses/penalties for opening kills, trades,
// entry fragging, and round type (pistol, eco, force).
func calculateSituationalBonus(roundStats *model.RoundStats, context *model.RoundContext) float64 {
	bonus := 0.0

	if roundStats.OpeningKill {
		bonus += 0.06
		if context.RoundType == "pistol" {
			bonus += 0.02
		}
	}

	if roundStats.OpeningDeath {
		if roundStats.TradeDeath {
			bonus -= 0.04
		} else {
			bonus -= 0.15
		}
	}

	if roundStats.EntryFragger {
		bonus += 0.04
	}

	if roundStats.TradeKill {
		bonus += 0.02
	}

	if roundStats.TradeDeath {
		bonus += 0.015
	}

	if roundStats.TradeDenials > 0 {
		bonus += float64(roundStats.TradeDenials) * 0.04
	}

	switch context.RoundType {
	case "pistol":
		bonus *= 1.3
	case "eco":
		if roundStats.TeamWon {
			bonus *= 1.4
		}
	case "force":
		bonus *= 1.1
	}

	if context.IsOvertimeRound {
		bonus *= 1.2
	}

	return bonus
}

// calculateImpactActions rewards bomb plants/defuses and eco kills,
// while penalizing anti-eco deaths.
func calculateImpactActions(roundStats *model.RoundStats, context *model.RoundContext) float64 {
	bonus := 0.0

	if roundStats.PlantedBomb {
		bonus += 0.08
		if context.TimeRemaining < 30.0 {
			bonus += 0.02
		}
	}

	if roundStats.DefusedBomb {
		bonus += 0.10
		if context.TimeRemaining < 10.0 {
			bonus += 0.03
		}
	}

	if roundStats.EcoKill {
		bonus += 0.04
	}

	if roundStats.AntiEcoKill {
		bonus -= 0.10
	}

	return bonus
}

// calculateMultiKillBonus returns bonus points for multi-kill rounds.
// 2K=0.03, 3K=0.08, 4K=0.15, 5K(ACE)=0.25
func calculateMultiKillBonus(roundStats *model.RoundStats) float64 {
	if roundStats.MultiKillRound < 2 {
		return 0.0
	}

	switch roundStats.MultiKillRound {
	case 2:
		return 0.03
	case 3:
		return 0.08
	case 4:
		return 0.15
	case 5:
		return 0.25
	default:
		return 0.0
	}
}

// calculateClutchModifier rewards successful clutches and partially rewards
// clutch attempts based on kills achieved.
func calculateClutchModifier(roundStats *model.RoundStats) float64 {
	modifier := 0.0

	if roundStats.ClutchAttempt {
		if roundStats.ClutchWon {
			switch {
			case roundStats.ClutchKills >= 4:
				modifier += 0.20
			case roundStats.ClutchKills >= 3:
				modifier += 0.15
			case roundStats.ClutchKills >= 2:
				modifier += 0.10
			default:
				modifier += 0.06
			}
		} else {
			modifier -= 0.02
			modifier += float64(roundStats.ClutchKills) * 0.02
		}
	}

	return modifier
}

// determineRoundType categorizes a round as pistol, eco, force, or full buy
// based on the round number.
func determineRoundType(roundNumber int) string {
	switch {
	case roundNumber == 1 || roundNumber == 16:
		return "pistol"
	case roundNumber <= 3 || (roundNumber >= 16 && roundNumber <= 18):
		return "eco"
	case roundNumber%3 == 0:
		return "force"
	default:
		return "full"
	}
}

// calculateEconomyModifier adjusts swing based on equipment advantage/disadvantage.
// Rewards performance when under-equipped, penalizes poor performance when over-equipped.
func calculateEconomyModifier(roundStats *model.RoundStats, playerEquip, teamEquip float64, roundType string) float64 {
	modifier := 1.0

	avgTeamEquip := teamEquip / 5.0
	equipRatio := playerEquip / math.Max(avgTeamEquip, 500.0)

	if equipRatio < 0.5 {
		if roundStats.TeamWon {
			modifier *= 1.3
		}
		if roundStats.Kills > 0 {
			modifier *= 1.2
		}
	} else if equipRatio > 1.5 {
		if !roundStats.TeamWon && roundStats.Kills == 0 {
			modifier *= 0.8
		}
	}

	switch roundType {
	case "eco":
		if roundStats.Kills > 0 {
			modifier *= 1.4
		}
	case "force":
		modifier *= 1.1
	}

	return modifier
}

// calculateUtilityImpact rewards effective utility usage (damage, flashes).
func calculateUtilityImpact(roundStats *model.RoundStats) float64 {
	bonus := 0.0

	if roundStats.UtilityDamage > 0 {
		utilityContrib := float64(roundStats.UtilityDamage) / 100.0 * 0.03
		bonus += math.Min(utilityContrib, 0.06)
	}

	if roundStats.EnemyFlashDuration > 0 {
		flashContrib := roundStats.EnemyFlashDuration / 3.0 * 0.02
		bonus += math.Min(flashContrib, 0.04)
	}

	if roundStats.FlashAssists > 0 {
		bonus += float64(roundStats.FlashAssists) * 0.015
	}

	return bonus
}

// calculateTradeSpeedBonus rewards fast trades (killing the enemy who killed a teammate).
// Faster trades receive higher bonuses.
func calculateTradeSpeedBonus(roundStats *model.RoundStats) float64 {
	if !roundStats.TradeKill || roundStats.TradeSpeed <= 0 {
		return 0.0
	}

	switch {
	case roundStats.TradeSpeed < 2.0:
		return 0.025
	case roundStats.TradeSpeed < 3.0:
		return 0.015
	case roundStats.TradeSpeed < 5.0:
		return 0.008
	default:
		return 0.0
	}
}

// calculateExitFragPenalty penalizes kills that occur after the round is decided
// (exit frags have less impact on round outcome).
func calculateExitFragPenalty(roundStats *model.RoundStats) float64 {
	if !roundStats.IsExitFrag {
		return 0.0
	}

	return float64(roundStats.ExitFrags) * 0.02
}

// calculateDeathTimingPenalty penalizes early deaths more heavily than late deaths.
// Dying in the first 15 seconds is most penalized.
func calculateDeathTimingPenalty(roundStats *model.RoundStats) float64 {
	if roundStats.Survived || roundStats.DeathTime <= 0 {
		return 0.0
	}

	switch {
	case roundStats.DeathTime < 15.0:
		return 0.08
	case roundStats.DeathTime < 30.0:
		return 0.05
	case roundStats.DeathTime < 60.0:
		return 0.02
	default:
		return 0.0
	}
}

// calculateTeamFlashPenalty penalizes flashing teammates.
func calculateTeamFlashPenalty(roundStats *model.RoundStats) float64 {
	if roundStats.TeamFlashCount == 0 {
		return 0.0
	}

	countPenalty := float64(roundStats.TeamFlashCount) * 0.02
	durationPenalty := roundStats.TeamFlashDuration * 0.008

	return math.Min(countPenalty+durationPenalty, 0.10)
}

// calculateFailedTradePenalty penalizes failing to trade a nearby teammate's death.
func calculateFailedTradePenalty(roundStats *model.RoundStats) float64 {
	if roundStats.FailedTrades == 0 {
		return 0.0
	}

	return float64(roundStats.FailedTrades) * 0.08
}

// calculateWeaponBonus rewards/penalizes based on weapon-specific performance.
// AWP kills are rewarded, losing an AWP without a kill is penalized.
// Knife kills and pistol vs rifle kills receive bonuses.
func calculateWeaponBonus(roundStats *model.RoundStats) float64 {
	bonus := 0.0

	if roundStats.AWPKill {
		if roundStats.LostAWP {
			bonus += 0.005
		} else {
			bonus += 0.02
		}
	} else if roundStats.LostAWP {
		bonus -= 0.05
	}

	if roundStats.KnifeKill {
		bonus += 0.03
	}

	if roundStats.PistolVsRifleKill {
		bonus += 0.025
	}

	return bonus
}
