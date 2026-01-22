// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// There may be mistakes in the comments. Please verify accuracy.
// =============================================================================

// Package rating implements the eco-rating calculation system.
// This file provides a centralized DerivedMetricsCalculator that computes
// per-round and percentage metrics from raw statistics. This eliminates
// duplication between parser.go and aggregator.go.
package rating

// DerivedMetricsInput contains the raw statistics needed to compute derived metrics.
type DerivedMetricsInput struct {
	RoundsPlayed int
	RoundsWon    int
	RoundsLost   int
	Kills        int
	Deaths       int
	Damage       int
	Assists      int

	// Time and flash stats
	TotalTimeAlive     float64
	EnemyFlashDuration float64
	TeamFlashDuration  float64

	// Round-based stats
	RoundsWithKill      int
	RoundsWithMultiKill int
	SavedByTeammate     int
	TradedDeaths        int
	SupportRounds       int
	SavedTeammate       int
	TradeKills          int
	OpeningKills        int
	OpeningDeaths       int
	OpeningAttempts     int
	OpeningSuccesses    int
	AttackRounds        int
	ClutchWins          int
	LastAliveRounds     int
	RoundsWithAWPKill   int
	AWPMultiKillRounds  int
	AWPOpeningKills     int
	UtilityDamage       int
	UtilityKills        int
	FlashesThrown       int
	FlashAssists        int
	LowBuyKills         int
	DisadvantagedKills  int
	AWPKills            int
	AssistedKills       int
	RoundsWonAfterOpen  int
	Clutch1v1Attempts   int
	Clutch1v1Wins       int
	SavesOnLoss         int

	// Survival tracking
	SurvivalCount float64
	KASTCount     float64
}

// DerivedMetricsOutput contains all computed per-round and percentage metrics.
type DerivedMetricsOutput struct {
	ADR                        float64
	KPR                        float64
	DPR                        float64
	AWPKillsPerRound           float64
	TimeAlivePerRound          float64
	EnemyFlashDurationPerRound float64
	TeamFlashDurationPerRound  float64
	RoundsWithKillPct          float64
	RoundsWithMultiKillPct     float64
	SavedByTeammatePerRound    float64
	TradedDeathsPerRound       float64
	TradedDeathsPct            float64
	AssistsPerRound            float64
	SupportRoundsPct           float64
	SavedTeammatePerRound      float64
	TradeKillsPerRound         float64
	TradeKillsPct              float64
	OpeningKillsPerRound       float64
	OpeningDeathsPerRound      float64
	OpeningAttemptsPct         float64
	OpeningSuccessPct          float64
	WinPctAfterOpeningKill     float64
	AttacksPerRound            float64
	ClutchPointsPerRound       float64
	LastAlivePct               float64
	Clutch1v1WinPct            float64
	SavesPerRoundLoss          float64
	RoundsWithAWPKillPct       float64
	AWPMultiKillRoundsPerRound float64
	AWPOpeningKillsPerRound    float64
	UtilityDamagePerRound      float64
	UtilityKillsPer100Rounds   float64
	FlashesThrownPerRound      float64
	FlashAssistsPerRound       float64
	LowBuyKillsPct             float64
	DisadvantagedBuyKillsPct   float64
	AssistedKillsPct           float64
	DamagePerKill              float64
	Survival                   float64
	KAST                       float64
}

// CalculateDerivedMetrics computes all derived metrics from raw statistics.
// This is the single source of truth for derived metric calculations.
func CalculateDerivedMetrics(input DerivedMetricsInput) DerivedMetricsOutput {
	output := DerivedMetricsOutput{}

	if input.RoundsPlayed == 0 {
		return output
	}

	rounds := float64(input.RoundsPlayed)

	// Basic per-round metrics
	output.ADR = float64(input.Damage) / rounds
	output.KPR = float64(input.Kills) / rounds
	output.DPR = float64(input.Deaths) / rounds
	output.AWPKillsPerRound = float64(input.AWPKills) / rounds

	// Time and flash metrics
	output.TimeAlivePerRound = input.TotalTimeAlive / rounds
	output.EnemyFlashDurationPerRound = input.EnemyFlashDuration / rounds
	output.TeamFlashDurationPerRound = input.TeamFlashDuration / rounds

	// Round-based percentages
	output.RoundsWithKillPct = float64(input.RoundsWithKill) / rounds
	output.RoundsWithMultiKillPct = float64(input.RoundsWithMultiKill) / rounds
	output.SavedByTeammatePerRound = float64(input.SavedByTeammate) / rounds
	output.TradedDeathsPerRound = float64(input.TradedDeaths) / rounds
	output.AssistsPerRound = float64(input.Assists) / rounds
	output.SupportRoundsPct = float64(input.SupportRounds) / rounds
	output.SavedTeammatePerRound = float64(input.SavedTeammate) / rounds
	output.TradeKillsPerRound = float64(input.TradeKills) / rounds
	output.OpeningKillsPerRound = float64(input.OpeningKills) / rounds
	output.OpeningDeathsPerRound = float64(input.OpeningDeaths) / rounds
	output.OpeningAttemptsPct = float64(input.OpeningAttempts) / rounds
	output.AttacksPerRound = float64(input.AttackRounds) / rounds
	output.ClutchPointsPerRound = float64(input.ClutchWins) / rounds
	output.LastAlivePct = float64(input.LastAliveRounds) / rounds
	output.RoundsWithAWPKillPct = float64(input.RoundsWithAWPKill) / rounds
	output.AWPMultiKillRoundsPerRound = float64(input.AWPMultiKillRounds) / rounds
	output.AWPOpeningKillsPerRound = float64(input.AWPOpeningKills) / rounds
	output.UtilityDamagePerRound = float64(input.UtilityDamage) / rounds
	output.UtilityKillsPer100Rounds = float64(input.UtilityKills) / rounds * 100
	output.FlashesThrownPerRound = float64(input.FlashesThrown) / rounds
	output.FlashAssistsPerRound = float64(input.FlashAssists) / rounds

	// Conditional percentages
	if input.Deaths > 0 {
		output.TradedDeathsPct = float64(input.TradedDeaths) / float64(input.Deaths)
	}

	if input.Kills > 0 {
		output.TradeKillsPct = float64(input.TradeKills) / float64(input.Kills)
		output.LowBuyKillsPct = float64(input.LowBuyKills) / float64(input.Kills)
		output.DisadvantagedBuyKillsPct = float64(input.DisadvantagedKills) / float64(input.Kills)
		output.AssistedKillsPct = float64(input.AssistedKills) / float64(input.Kills)
		output.DamagePerKill = float64(input.Damage) / float64(input.Kills)
	}

	if input.OpeningAttempts > 0 {
		output.OpeningSuccessPct = float64(input.OpeningSuccesses) / float64(input.OpeningAttempts)
	}

	if input.OpeningKills > 0 {
		output.WinPctAfterOpeningKill = float64(input.RoundsWonAfterOpen) / float64(input.OpeningKills)
	}

	if input.Clutch1v1Attempts > 0 {
		output.Clutch1v1WinPct = float64(input.Clutch1v1Wins) / float64(input.Clutch1v1Attempts)
	}

	if input.RoundsLost > 0 {
		output.SavesPerRoundLoss = float64(input.SavesOnLoss) / float64(input.RoundsLost)
	}

	// Survival and KAST (already normalized if passed as ratios)
	output.Survival = input.SurvivalCount / rounds
	output.KAST = input.KASTCount / rounds

	return output
}
