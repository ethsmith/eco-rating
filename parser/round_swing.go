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
	"eco-rating/rating"
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
	involvement := float64(roundStats.Kills)
	involvement += float64(roundStats.Assists) * rating.InvolvementAssistWeight
	involvement += math.Min(1.0, float64(roundStats.Damage)/rating.InvolvementDamageNorm)
	if roundStats.PlantedBomb {
		involvement += rating.InvolvementBombWeight
	}
	if roundStats.DefusedBomb {
		involvement += rating.InvolvementBombWeight
	}
	if roundStats.Survived {
		involvement += rating.InvolvementSurvivalWeight
	}
	involvement = math.Max(0.0, math.Min(involvement/rating.InvolvementNormDivisor, 1.0))

	var baseSwing float64
	if roundStats.TeamWon {
		baseSwing = rating.SwingBaseWin * involvement
	} else {
		baseSwing = rating.SwingBaseLoss * involvement
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

	return math.Max(rating.SwingMinClamp, math.Min(rating.SwingMaxClamp, totalSwing))
}

// calculatePerformanceContribution computes the base performance score from
// kills, damage, assists, and survival status.
func calculatePerformanceContribution(roundStats *model.RoundStats) float64 {
	contribution := 0.0

	if roundStats.Kills > 0 {
		contribution += float64(roundStats.Kills) * rating.KillContribPerKill
		if roundStats.Kills >= 2 {
			contribution += float64(roundStats.Kills-1) * rating.KillContribMultiBonus
		}
	}

	effectiveDamage := roundStats.Damage
	if roundStats.IsExitFrag {
		effectiveDamage = int(float64(effectiveDamage) * rating.ExitFragDamageMultiplier)
	}
	damageContrib := float64(effectiveDamage) / rating.DamageContribDivisor
	contribution += math.Min(damageContrib, rating.DamageContribMax)

	contribution += float64(roundStats.Assists) * rating.AssistContrib
	contribution += float64(roundStats.FlashAssists) * rating.FlashAssistContrib

	if roundStats.Survived {
		if roundStats.TeamWon {
			contribution += rating.SurvivalContribWin
		} else {
			contribution += rating.SurvivalContribLoss
			roundStats.SavedWeapons = true
		}
	} else {
		if roundStats.TradeDeath {
			contribution -= rating.DeathPenaltyTraded
		} else {
			contribution -= rating.DeathPenaltyUntraded
		}
	}

	return contribution
}

// calculateSituationalBonus adds bonuses/penalties for opening kills, trades,
// entry fragging, and round type (pistol, eco, force).
func calculateSituationalBonus(roundStats *model.RoundStats, context *model.RoundContext) float64 {
	bonus := 0.0

	if roundStats.OpeningKill {
		bonus += rating.OpeningKillBonus
		if context.RoundType == "pistol" {
			bonus += rating.OpeningKillPistolBonus
		}
	}

	if roundStats.OpeningDeath {
		if roundStats.TradeDeath {
			bonus -= rating.OpeningDeathTraded
		} else {
			bonus -= rating.OpeningDeathUntraded
		}
	}

	if roundStats.EntryFragger {
		bonus += rating.EntryFragBonus
	}

	if roundStats.TradeKill {
		bonus += rating.TradeKillBonus
	}

	if roundStats.TradeDeath {
		bonus += rating.TradeDeathMitigation
	}

	if roundStats.TradeDenials > 0 {
		bonus += float64(roundStats.TradeDenials) * rating.TradeDenialBonus
	}

	switch context.RoundType {
	case "pistol":
		bonus *= rating.PistolRoundMultiplier
	case "eco":
		if roundStats.TeamWon {
			bonus *= rating.EcoRoundWinMultiplier
		}
	case "force":
		bonus *= rating.ForceRoundMultiplier
	}

	if context.IsOvertimeRound {
		bonus *= rating.OvertimeMultiplier
	}

	return bonus
}

// calculateImpactActions rewards bomb plants/defuses and eco kills,
// while penalizing anti-eco deaths.
func calculateImpactActions(roundStats *model.RoundStats, context *model.RoundContext) float64 {
	bonus := 0.0

	if roundStats.PlantedBomb {
		bonus += rating.BombPlantBonus
		if context.TimeRemaining < rating.LateRoundTimeThreshold {
			bonus += rating.BombPlantLateBonus
		}
	}

	if roundStats.DefusedBomb {
		bonus += rating.BombDefuseBonus
		if context.TimeRemaining < rating.ClutchDefuseThreshold {
			bonus += rating.BombDefuseClutchBonus
		}
	}

	if roundStats.EcoKill {
		bonus += rating.EcoKillBonus
	}

	if roundStats.AntiEcoKill {
		bonus -= rating.AntiEcoDeathPenalty
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
		return rating.MultiKill2KBonus
	case 3:
		return rating.MultiKill3KBonus
	case 4:
		return rating.MultiKill4KBonus
	case 5:
		return rating.MultiKill5KBonus
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
				modifier += rating.ClutchWin4KBonus
			case roundStats.ClutchKills >= 3:
				modifier += rating.ClutchWin3KBonus
			case roundStats.ClutchKills >= 2:
				modifier += rating.ClutchWin2KBonus
			default:
				modifier += rating.ClutchWin1KBonus
			}
		} else {
			modifier -= rating.ClutchLossPenalty
			modifier += float64(roundStats.ClutchKills) * rating.ClutchLossKillMitigation
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
	equipRatio := playerEquip / math.Max(avgTeamEquip, rating.MinTeamEquipValue)

	if equipRatio < rating.LowEquipRatioThreshold {
		if roundStats.TeamWon {
			modifier *= rating.LowEquipWinMultiplier
		}
		if roundStats.Kills > 0 {
			modifier *= rating.LowEquipKillMultiplier
		}
	} else if equipRatio > rating.HighEquipRatioThreshold {
		if !roundStats.TeamWon && roundStats.Kills == 0 {
			modifier *= rating.HighEquipFailMultiplier
		}
	}

	switch roundType {
	case "eco":
		if roundStats.Kills > 0 {
			modifier *= rating.EcoRoundKillMultiplier
		}
	case "force":
		modifier *= rating.ForceRoundModifier
	}

	return modifier
}

// calculateUtilityImpact rewards effective utility usage (damage, flashes).
func calculateUtilityImpact(roundStats *model.RoundStats) float64 {
	bonus := 0.0

	if roundStats.UtilityDamage > 0 {
		utilityContrib := float64(roundStats.UtilityDamage) / 100.0 * rating.UtilityDamageContribRate
		bonus += math.Min(utilityContrib, rating.UtilityDamageContribMax)
	}

	if roundStats.EnemyFlashDuration > 0 {
		flashContrib := roundStats.EnemyFlashDuration / 3.0 * rating.FlashDurationContribRate
		bonus += math.Min(flashContrib, rating.FlashDurationContribMax)
	}

	if roundStats.FlashAssists > 0 {
		bonus += float64(roundStats.FlashAssists) * rating.FlashAssistBonusRate
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
	case roundStats.TradeSpeed < rating.FastTradeThreshold:
		return rating.FastTradeBonus
	case roundStats.TradeSpeed < rating.MediumTradeThreshold:
		return rating.MediumTradeBonus
	case roundStats.TradeSpeed < rating.SlowTradeThreshold:
		return rating.SlowTradeBonus
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

	return float64(roundStats.ExitFrags) * rating.ExitFragPenaltyRate
}

// calculateDeathTimingPenalty penalizes early deaths more heavily than late deaths.
// Dying in the first 15 seconds is most penalized.
func calculateDeathTimingPenalty(roundStats *model.RoundStats) float64 {
	if roundStats.Survived || roundStats.DeathTime <= 0 {
		return 0.0
	}

	switch {
	case roundStats.DeathTime < rating.EarlyDeathThreshold:
		return rating.EarlyDeathPenalty
	case roundStats.DeathTime < rating.MidDeathThreshold:
		return rating.MidDeathPenalty
	case roundStats.DeathTime < rating.LateDeathThreshold:
		return rating.LateDeathPenalty
	default:
		return 0.0
	}
}

// calculateTeamFlashPenalty penalizes flashing teammates.
func calculateTeamFlashPenalty(roundStats *model.RoundStats) float64 {
	if roundStats.TeamFlashCount == 0 {
		return 0.0
	}

	countPenalty := float64(roundStats.TeamFlashCount) * rating.TeamFlashCountPenalty
	durationPenalty := roundStats.TeamFlashDuration * rating.TeamFlashDurationPenalty

	return math.Min(countPenalty+durationPenalty, rating.TeamFlashPenaltyMax)
}

// calculateFailedTradePenalty penalizes failing to trade a nearby teammate's death.
func calculateFailedTradePenalty(roundStats *model.RoundStats) float64 {
	if roundStats.FailedTrades == 0 {
		return 0.0
	}

	return float64(roundStats.FailedTrades) * rating.FailedTradePenalty
}

// calculateWeaponBonus rewards/penalizes based on weapon-specific performance.
// AWP kills are rewarded, losing an AWP without a kill is penalized.
// Knife kills and pistol vs rifle kills receive bonuses.
func calculateWeaponBonus(roundStats *model.RoundStats) float64 {
	bonus := 0.0

	if roundStats.AWPKill {
		if roundStats.LostAWP {
			bonus += rating.AWPKillBonusWithDeath
		} else {
			bonus += rating.AWPKillBonusWithSurvive
		}
	} else if roundStats.LostAWP {
		bonus -= rating.AWPLostNokillPenalty
	}

	if roundStats.KnifeKill {
		bonus += rating.KnifeKillBonus
	}

	if roundStats.PistolVsRifleKill {
		bonus += rating.PistolVsRifleKillBonus
	}

	return bonus
}
