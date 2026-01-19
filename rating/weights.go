// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// =============================================================================

// Package rating implements the eco-rating calculation system.
// This file defines all constants used in rating calculations, including:
// - Component weights for the final rating formula
// - Baseline values for normalization
// - Economic kill/death multipliers
// - Rating bounds
package rating

// Rating component weights - these determine how much each metric contributes
// to the final rating calculation.
const (
	WeightKillRating      = 0.12 // Weight for kills per round contribution
	WeightADRRating       = 0.10 // Weight for average damage per round
	WeightMultiKillRating = 0.08 // Weight for multi-kill rounds
	WeightDeathRating     = 0.20 // Weight for deaths per round (negative impact)
	WeightKDRating        = 0.15 // Weight for K/D ratio
	WeightSwingRating     = 0.35 // Weight for round swing (largest factor)
)

// Baseline values represent average/expected performance levels.
// These are used to normalize metrics so that average performance = 1.0 contribution.
const (
	BaselineKPR                = 0.7   // Average kills per round
	BaselineDPR                = 0.7   // Average deaths per round
	BaselineADR                = 75.0  // Average damage per round
	BaselineOpeningKills       = 0.10  // Opening kills per round
	BaselineMultiKill          = 0.17  // Multi-kill rounds per round
	BaselineAssists            = 0.15  // Assists per round
	BaselineKAST               = 0.70  // KAST percentage (Kill/Assist/Survive/Trade)
	BaselineRoundSwing         = 0.00  // Neutral round swing
	BaselineOpeningSuccessRate = 0.50  // Opening duel win rate
	BaselineTradeKillsPerRound = 0.115 // Trade kills per round
	BaselineUtilityDamage      = 4.0   // Utility damage per round
	BaselineFlashAssists       = 0.38  // Flash assists per round
	BaselineEnemyFlashDur      = 0.9   // Enemy flash duration per round (seconds)
	BaselineEcoDeathValue      = 0.72  // Average eco death penalty
	BaselineEarlyDeaths        = 0.25  // Early deaths per round
	BaselineUntradedOpenings   = 0.04  // Untraded opening deaths per round
	BaselineTeamFlashPerRound  = 0.24  // Team flashes per round
	BaselineFailedClutchRate   = 0.60  // Failed clutch rate
)

// Economic kill value multipliers - rewards kills against better-equipped opponents.
// Higher values mean the kill is worth more to the rating.
const (
	EcoKillPistolVsRifle      = 1.80 // Pistol killing rifle (huge disadvantage)
	EcoKillEcoVsForce         = 1.50 // Eco killing force/full buy
	EcoKillForceVsFullBuy     = 1.25 // Force buy killing full buy
	EcoKillSlightDisadvantage = 1.10 // Slight equipment disadvantage
	EcoKillEqual              = 1.00 // Equal equipment
	EcoKillSlightAdvantage    = 0.95 // Slight equipment advantage
	EcoKillAdvantage          = 0.85 // Clear equipment advantage
	EcoKillRifleVsPistol      = 0.70 // Rifle killing pistol (expected)
)

// Economic death penalty multipliers - penalizes deaths to worse-equipped opponents.
// Higher values mean the death hurts the rating more.
const (
	EcoDeathToPistol           = 1.60 // Rifle dying to pistol (embarrassing)
	EcoDeathToEco              = 1.40 // Full buy dying to eco
	EcoDeathToForceBuy         = 1.20 // Full buy dying to force
	EcoDeathSlightAdvantage    = 1.10 // Slight advantage, still died
	EcoDeathEqual              = 1.00 // Equal equipment death
	EcoDeathSlightDisadvantage = 0.95 // Slight disadvantage death
	EcoDeathDisadvantage       = 0.85 // Clear disadvantage death
	EcoDeathPistolVsRifle      = 0.70 // Pistol dying to rifle (expected)
)

// Minimum equipment value to prevent division by zero in ratio calculations.
const (
	MinEquipmentValue = 100.0
)

// Rating bounds - final ratings are clamped to this range.
const (
	MinRating = 0.20 // Minimum possible rating
	MaxRating = 3.00 // Maximum possible rating
)
