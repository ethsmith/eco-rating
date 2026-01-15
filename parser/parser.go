package parser

import (
	"eco-rating/model"
	"eco-rating/output"
	"eco-rating/rating"
	"io"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
)

type DemoParser struct {
	parser demoinfocs.Parser
	state  *MatchState
	logger *Logger
}

// NewDemoParser creates a new demo parser with logging disabled by default
func NewDemoParser(r io.Reader) *DemoParser {
	return NewDemoParserWithLogging(r, false)
}

// NewDemoParserWithLogging creates a new demo parser with optional logging
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

// SetLogging enables or disables logging
func (d *DemoParser) SetLogging(enabled bool) {
	d.logger.SetEnabled(enabled)
}

// SetPlayerFilter sets the list of players to filter logs for
// Only events involving these players will be logged
// Pass empty slice to clear filter and log all events
func (d *DemoParser) SetPlayerFilter(players []string) {
	d.logger.SetPlayerFilter(players)
}

// AddPlayerFilter adds a player to the filter list
func (d *DemoParser) AddPlayerFilter(player string) {
	d.logger.AddPlayerFilter(player)
}

// ClearPlayerFilter clears the player filter (logs all events)
func (d *DemoParser) ClearPlayerFilter() {
	d.logger.ClearPlayerFilter()
}

func (d *DemoParser) Parse() {
	_ = d.parser.ParseToEnd()

	// Compute final stats for all players
	for _, p := range d.state.Players {
		if p.RoundsPlayed > 0 {
			rounds := float64(p.RoundsPlayed)
			// Calculate per-round stats
			p.ADR = float64(p.Damage) / rounds
			p.KPR = float64(p.Kills) / rounds
			p.DPR = float64(p.Deaths) / rounds
			// Convert KAST count to percentage
			p.KAST = p.KAST / rounds
			// Convert Survival count to percentage
			p.Survival = p.Survival / rounds

			// Calculate AWP kills per round
			if p.RoundsPlayed > 0 {
				p.AWPKillsPerRound = float64(p.AWPKills) / rounds
			}

			// Calculate HLTV Rating 1.0
			// Formula: (KillRating + 0.7*SurvivalRating + RoundsWithMultiKillRating) / 2.7
			// Where each component is normalized against average values
			avgKPR := 0.679 // Average kills per round
			avgSPR := 0.317 // Average survival rate
			avgRMK := 0.073 // Average rounds with multi-kill rate

			killRating := p.KPR / avgKPR
			survivalRating := p.Survival / avgSPR
			rmkRating := float64(p.RoundsWithMultiKill) / rounds / avgRMK

			p.HLTVRating = (killRating + 0.7*survivalRating + rmkRating) / 2.7

			// Calculate Pistol Round Rating (HLTV 1.0 for pistol rounds only)
			if p.PistolRoundsPlayed > 0 {
				pistolRounds := float64(p.PistolRoundsPlayed)
				pistolKPR := float64(p.PistolRoundKills) / pistolRounds
				pistolSurvival := float64(p.PistolRoundSurvivals) / pistolRounds
				pistolRMK := float64(p.PistolRoundMultiKills) / pistolRounds

				pistolKillRating := pistolKPR / avgKPR
				pistolSurvivalRating := pistolSurvival / avgSPR
				pistolRMKRating := pistolRMK / avgRMK

				p.PistolRoundRating = (pistolKillRating + 0.7*pistolSurvivalRating + pistolRMKRating) / 2.7
			}

			// Calculate T-side HLTV Rating 1.0
			if p.TRoundsPlayed > 0 {
				tRounds := float64(p.TRoundsPlayed)
				tKPR := float64(p.TKills) / tRounds
				tSurvival := float64(p.TSurvivals) / tRounds
				tRMK := float64(p.TRoundsWithMultiKill) / tRounds

				tKillRating := tKPR / avgKPR
				tSurvivalRating := tSurvival / avgSPR
				tRMKRating := tRMK / avgRMK

				p.TRating = (tKillRating + 0.7*tSurvivalRating + tRMKRating) / 2.7
			}

			// Calculate CT-side HLTV Rating 1.0
			if p.CTRoundsPlayed > 0 {
				ctRounds := float64(p.CTRoundsPlayed)
				ctKPR := float64(p.CTKills) / ctRounds
				ctSurvival := float64(p.CTSurvivals) / ctRounds
				ctRMK := float64(p.CTRoundsWithMultiKill) / ctRounds

				ctKillRating := ctKPR / avgKPR
				ctSurvivalRating := ctSurvival / avgSPR
				ctRMKRating := ctRMK / avgRMK

				p.CTRating = (ctKillRating + 0.7*ctSurvivalRating + ctRMKRating) / 2.7
			}
		}
		p.FinalRating = rating.ComputeFinalRating(p)

		// Calculate per-side Eco Ratings
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

		// Log player summary
		d.logger.LogPlayerSummary(p.Name, p.Kills, p.Deaths, p.Damage, p.EcoKillValue, p.EcoDeathValue, p.FinalRating)
	}
}

func (d *DemoParser) ExportJSON(path string) error {
	return output.Export(d.state.Players, path)
}

// GetPlayers returns the parsed player stats map
func (d *DemoParser) GetPlayers() map[uint64]*model.PlayerStats {
	return d.state.Players
}

// GetMapName returns the map name captured from the demo
func (d *DemoParser) GetMapName() string {
	return d.state.MapName
}

// GetLogs returns the buffered log output from parsing
func (d *DemoParser) GetLogs() string {
	return d.logger.GetOutput()
}
