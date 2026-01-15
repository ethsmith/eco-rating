package rating

// === Rating Component Weights ===
// 60% output metrics, 40% cost metrics
const (
	// Output components (60% total)
	WeightKillRating      = 0.20 // Eco-adjusted kills per round
	WeightADRRating       = 0.14 // Damage per round
	WeightMultiKillRating = 0.08 // Multi-kill rating (explosive moments)
	WeightKASTRating      = 0.06 // Consistency metric
	WeightOpeningRating   = 0.05 // Opening duel performance
	WeightTradeRating     = 0.04 // Trade efficiency and team play
	WeightUtilityRating   = 0.02 // Utility/support impact
	WeightSwingRating     = 0.01 // Round Swing

	// Cost components (40% total)
	WeightDeathRating         = 0.12 // Death penalty (base)
	WeightEcoDeathRating      = 0.10 // Eco-adjusted death value
	WeightEarlyDeathRating    = 0.08 // Early/opening deaths penalty
	WeightUntradedDeathRating = 0.05 // Untraded opening deaths
	WeightTeamFlashRating     = 0.02 // Team flash penalty
	WeightFailedClutchRating  = 0.03 // Failed clutch penalty
)

// === Normalization Baselines ===
// These represent "average" values used to normalize each component to ~1.0
// Calibrated based on actual demo statistics
const (
	BaselineKPR                = 0.65 // Average kills per round (based on demo data)
	BaselineDPR                = 0.65 // Average deaths per round (based on demo data)
	BaselineADR                = 75.0 // Average damage per round (based on demo data)
	BaselineOpeningKills       = 0.12 // Average opening kills per round
	BaselineMultiKill          = 0.20 // Average multi-kill bonus per round
	BaselineAssists            = 0.15 // Average assists per round
	BaselineKAST               = 0.70 // Average KAST percentage (70%)
	BaselineRoundSwing         = 0.00 // Average round swing (zero-sum)
	BaselineOpeningSuccessRate = 0.50 // Average opening duel success rate (50%)
	BaselineTradeKillsPerRound = 0.08 // Average trade kills per round
	BaselineUtilityDamage      = 15.0 // Average utility damage per round
	BaselineFlashAssists       = 0.12 // Average flash assists per round
	BaselineEnemyFlashDur      = 1.2  // Average enemy flash duration per round (seconds)

	// Cost component baselines
	BaselineEcoDeathValue     = 0.65 // Average eco death value per round
	BaselineEarlyDeaths       = 0.08 // Average early deaths per round
	BaselineUntradedOpenings  = 0.04 // Average untraded opening deaths per round
	BaselineTeamFlashPerRound = 0.15 // Average team flashes per round
	BaselineFailedClutchRate  = 0.60 // Average failed clutch rate (1 - win rate)
)

// === Eco Kill Value Multipliers ===
// Based on equipment ratio (victim/attacker)
const (
	EcoKillPistolVsRifle      = 1.80 // Ratio > 4.0: Pistol kill on full buy
	EcoKillEcoVsForce         = 1.50 // Ratio > 2.0: Eco kill on force/full buy
	EcoKillForceVsFullBuy     = 1.25 // Ratio > 1.3: Force buy vs full buy (~$3000 vs $4500)
	EcoKillSlightDisadvantage = 1.10 // Ratio > 1.1: Slight equipment disadvantage
	EcoKillEqual              = 1.00 // Ratio 0.9-1.1: Roughly equal equipment
	EcoKillSlightAdvantage    = 0.95 // Ratio 0.75-0.9: Slight equipment advantage
	EcoKillAdvantage          = 0.85 // Ratio 0.5-0.75: Equipment advantage
	EcoKillRifleVsPistol      = 0.70 // Ratio < 0.5: Large equipment advantage
)

// === Eco Death Penalty Multipliers ===
// Based on equipment ratio (victim/killer) - higher = more embarrassing death
const (
	EcoDeathToPistol           = 1.60 // Ratio > 4.0: Died to pistol with rifle
	EcoDeathToEco              = 1.40 // Ratio > 2.0: Died to eco on force/full buy
	EcoDeathToForceBuy         = 1.20 // Ratio > 1.3: Full buy died to force buy
	EcoDeathSlightAdvantage    = 1.10 // Ratio > 1.1: Slight equipment advantage
	EcoDeathEqual              = 1.00 // Ratio 0.9-1.1: Roughly equal equipment
	EcoDeathSlightDisadvantage = 0.95 // Ratio 0.75-0.9: Slight equipment disadvantage
	EcoDeathDisadvantage       = 0.85 // Ratio 0.5-0.75: Died to better equipped
	EcoDeathPistolVsRifle      = 0.70 // Ratio < 0.5: Eco vs rifle death
)

// === Equipment Thresholds ===
const (
	MinEquipmentValue = 100.0 // Minimum equipment value to avoid division issues
)

// === Rating Clamp Values ===
const (
	MinRating = 0.30 // Minimum possible rating
	MaxRating = 2.50 // Maximum possible rating
)
