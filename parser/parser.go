// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// =============================================================================

// Package parser provides CS2 demo file parsing functionality.
// This file contains the main DemoParser struct and its methods for parsing
// demo files, computing player statistics, and calculating ratings.
package parser

import (
	"eco-rating/model"
	"eco-rating/rating"
	"io"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
)

// DemoParser wraps the demoinfocs parser and manages match state and logging.
// It processes CS2 demo files and extracts comprehensive player statistics.
type DemoParser struct {
	parser demoinfocs.Parser
	state  *MatchState
	logger *Logger
}

// NewDemoParser creates a new DemoParser with logging disabled.
func NewDemoParser(r io.Reader) *DemoParser {
	return NewDemoParserWithLogging(r, false)
}

// NewDemoParserWithLogging creates a new DemoParser with configurable logging.
// The parser is initialized with event handlers but Parse() must be called to process.
func NewDemoParserWithLogging(r io.Reader, enableLogging bool) *DemoParser {
	p := demoinfocs.NewParser(r)
	state := NewMatchState()

	dp := &DemoParser{
		parser: p,
		state:  state,
		logger: NewLogger(enableLogging),
	}

	dp.registerHandlers()
	return dp
}

// SetLogging enables or disables detailed parsing logs.
func (d *DemoParser) SetLogging(enabled bool) {
	d.logger.SetEnabled(enabled)
}

// SetPlayerFilter limits logging to events involving the specified players.
func (d *DemoParser) SetPlayerFilter(players []string) {
	d.logger.SetPlayerFilter(players)
}

// AddPlayerFilter adds a player to the logging filter.
func (d *DemoParser) AddPlayerFilter(player string) {
	d.logger.AddPlayerFilter(player)
}

// ClearPlayerFilter removes all player filters, logging all events.
func (d *DemoParser) ClearPlayerFilter() {
	d.logger.ClearPlayerFilter()
}

// Parse processes the entire demo file and computes all player statistics.
// After parsing, it calculates derived metrics (ADR, KPR, ratings, etc.)
// and the final eco-rating for each player.
func (d *DemoParser) Parse() {
	_ = d.parser.ParseToEnd()

	for _, p := range d.state.Players {
		if p.RoundsPlayed > 0 {
			rounds := float64(p.RoundsPlayed)
			p.ADR = float64(p.Damage) / rounds
			p.KPR = float64(p.Kills) / rounds
			p.DPR = float64(p.Deaths) / rounds
			p.KAST = p.KAST / rounds
			p.Survival = p.Survival / rounds

			if p.RoundsPlayed > 0 {
				p.AWPKillsPerRound = float64(p.AWPKills) / rounds
			}

			killRating := p.KPR / rating.HLTVBaselineKPR
			survived := p.Survival * rounds
			survivalRating := ((survived - float64(p.Deaths)) / rounds) / rating.HLTVBaselineSPR
			rmkPoints := float64(p.MultiKillsRaw[1]*1 + p.MultiKillsRaw[2]*4 + p.MultiKillsRaw[3]*9 + p.MultiKillsRaw[4]*16 + p.MultiKillsRaw[5]*25)
			rmkRating := (rmkPoints / rounds) / rating.HLTVBaselineRMK

			p.HLTVRating = (killRating + rating.HLTVSurvivalWeight*survivalRating + rmkRating) / rating.HLTVRatingDivisor

			if p.PistolRoundsPlayed > 0 {
				pistolRounds := float64(p.PistolRoundsPlayed)
				pistolKPR := float64(p.PistolRoundKills) / pistolRounds
				pistolSurvivalRating := ((float64(p.PistolRoundSurvivals) - float64(p.PistolRoundDeaths)) / pistolRounds) / rating.HLTVBaselineSPR
				pistolRMKPoints := float64(p.PistolRoundMultiKills) * 4.0
				pistolRMKRating := (pistolRMKPoints / pistolRounds) / rating.HLTVBaselineRMK

				pistolKillRating := pistolKPR / rating.HLTVBaselineKPR
				p.PistolRoundRating = (pistolKillRating + rating.HLTVSurvivalWeight*pistolSurvivalRating + pistolRMKRating) / rating.HLTVRatingDivisor
			}

			if p.TRoundsPlayed > 0 {
				tRounds := float64(p.TRoundsPlayed)
				tKPR := float64(p.TKills) / tRounds
				tSurvivalRating := ((float64(p.TSurvivals) - float64(p.TDeaths)) / tRounds) / rating.HLTVBaselineSPR
				tRMKPoints := float64(p.TMultiKills[1]*1 + p.TMultiKills[2]*4 + p.TMultiKills[3]*9 + p.TMultiKills[4]*16 + p.TMultiKills[5]*25)
				tRMKRating := (tRMKPoints / tRounds) / rating.HLTVBaselineRMK

				tKillRating := tKPR / rating.HLTVBaselineKPR
				p.TRating = (tKillRating + rating.HLTVSurvivalWeight*tSurvivalRating + tRMKRating) / rating.HLTVRatingDivisor
			}

			if p.CTRoundsPlayed > 0 {
				ctRounds := float64(p.CTRoundsPlayed)
				ctKPR := float64(p.CTKills) / ctRounds
				ctSurvivalRating := ((float64(p.CTSurvivals) - float64(p.CTDeaths)) / ctRounds) / rating.HLTVBaselineSPR
				ctRMKPoints := float64(p.CTMultiKills[1]*1 + p.CTMultiKills[2]*4 + p.CTMultiKills[3]*9 + p.CTMultiKills[4]*16 + p.CTMultiKills[5]*25)
				ctRMKRating := (ctRMKPoints / ctRounds) / rating.HLTVBaselineRMK

				ctKillRating := ctKPR / rating.HLTVBaselineKPR
				p.CTRating = (ctKillRating + rating.HLTVSurvivalWeight*ctSurvivalRating + ctRMKRating) / rating.HLTVRatingDivisor
			}

			p.TimeAlivePerRound = p.TotalTimeAlive / rounds
			p.EnemyFlashDurationPerRound = p.EnemyFlashDuration / rounds
			p.TeamFlashDurationPerRound = p.TeamFlashDuration / rounds
			p.RoundsWithKillPct = float64(p.RoundsWithKill) / rounds
			p.RoundsWithMultiKillPct = float64(p.RoundsWithMultiKill) / rounds
			p.SavedByTeammatePerRound = float64(p.SavedByTeammate) / rounds
			p.TradedDeathsPerRound = float64(p.TradedDeaths) / rounds
			p.AssistsPerRound = float64(p.Assists) / rounds
			p.SupportRoundsPct = float64(p.SupportRounds) / rounds
			p.SavedTeammatePerRound = float64(p.SavedTeammate) / rounds
			p.TradeKillsPerRound = float64(p.TradeKills) / rounds
			p.OpeningKillsPerRound = float64(p.OpeningKills) / rounds
			p.OpeningDeathsPerRound = float64(p.OpeningDeaths) / rounds
			p.OpeningAttemptsPct = float64(p.OpeningAttempts) / rounds
			p.AttacksPerRound = float64(p.AttackRounds) / rounds
			p.ClutchPointsPerRound = float64(p.ClutchWins) / rounds
			p.LastAlivePct = float64(p.LastAliveRounds) / rounds
			p.RoundsWithAWPKillPct = float64(p.RoundsWithAWPKill) / rounds
			p.AWPMultiKillRoundsPerRound = float64(p.AWPMultiKillRounds) / rounds
			p.AWPOpeningKillsPerRound = float64(p.AWPOpeningKills) / rounds
			p.UtilityDamagePerRound = float64(p.UtilityDamage) / rounds
			p.UtilityKillsPer100Rounds = float64(p.UtilityKills) * 100 / rounds
			p.FlashesThrownPerRound = float64(p.FlashesThrown) / rounds
			p.FlashAssistsPerRound = float64(p.FlashAssists) / rounds
		}

		if p.RoundsWon > 0 {
			p.KillsPerRoundWin = float64(p.KillsInWonRounds) / float64(p.RoundsWon)
			p.DamagePerRoundWin = float64(p.DamageInWonRounds) / float64(p.RoundsWon)
		}

		if p.RoundsLost > 0 {
			p.SavesPerRoundLoss = float64(p.SavesOnLoss) / float64(p.RoundsLost)
		}

		if p.Deaths > 0 {
			p.TradedDeathsPct = float64(p.TradedDeaths) / float64(p.Deaths)
		}

		if p.OpeningDeaths > 0 {
			p.OpeningDeathsTradedPct = float64(p.OpeningDeathsTraded) / float64(p.OpeningDeaths)
		}

		if p.Kills > 0 {
			p.TradeKillsPct = float64(p.TradeKills) / float64(p.Kills)
			p.AssistedKillsPct = float64(p.AssistedKills) / float64(p.Kills)
			p.DamagePerKill = float64(p.Damage) / float64(p.Kills)
			p.AWPKillsPct = float64(p.AWPKills) / float64(p.Kills)
			p.LowBuyKillsPct = float64(p.LowBuyKills) / float64(p.Kills)
			p.DisadvantagedBuyKillsPct = float64(p.DisadvantagedBuyKills) / float64(p.Kills)
		}

		if p.OpeningAttempts > 0 {
			p.OpeningSuccessPct = float64(p.OpeningSuccesses) / float64(p.OpeningAttempts)
		}

		if p.OpeningKills > 0 {
			p.WinPctAfterOpeningKill = float64(p.RoundsWonAfterOpening) / float64(p.OpeningKills)
		}

		if p.Clutch1v1Attempts > 0 {
			p.Clutch1v1WinPct = float64(p.Clutch1v1Wins) / float64(p.Clutch1v1Attempts)
		}

		p.MultiKills.OneK = p.MultiKillsRaw[1]
		p.MultiKills.TwoK = p.MultiKillsRaw[2]
		p.MultiKills.ThreeK = p.MultiKillsRaw[3]
		p.MultiKills.FourK = p.MultiKillsRaw[4]
		p.MultiKills.FiveK = p.MultiKillsRaw[5]

		p.FinalRating = rating.ComputeFinalRating(p)

		if p.TRoundsPlayed > 0 {
			p.TEcoRating = rating.ComputeSideRating(
				p.TRoundsPlayed, p.TKills, p.TDeaths, p.TDamage, p.TEcoKillValue,
				p.TRoundSwing, p.TKAST, p.TMultiKills, p.TClutchRounds, p.TClutchWins)
		}
		if p.CTRoundsPlayed > 0 {
			p.CTEcoRating = rating.ComputeSideRating(
				p.CTRoundsPlayed, p.CTKills, p.CTDeaths, p.CTDamage, p.CTEcoKillValue,
				p.CTRoundSwing, p.CTKAST, p.CTMultiKills, p.CTClutchRounds, p.CTClutchWins)
		}

		d.logger.LogPlayerSummary(p.Name, p.Kills, p.Deaths, p.Damage, p.EcoKillValue, p.EcoDeathValue, p.FinalRating)
	}
}

// GetPlayers returns the map of all player statistics keyed by Steam ID.
func (d *DemoParser) GetPlayers() map[uint64]*model.PlayerStats {
	return d.state.Players
}

// GetMapName returns the name of the map played (e.g., "de_dust2").
func (d *DemoParser) GetMapName() string {
	return d.state.MapName
}

// GetLogs returns all captured log output from parsing.
func (d *DemoParser) GetLogs() string {
	return d.logger.GetOutput()
}
