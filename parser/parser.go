// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// =============================================================================

// Package parser provides CS2 demo file parsing functionality.
// This file contains the main DemoParser struct and its methods for parsing
// demo files, computing player statistics, and calculating ratings.
package parser

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/ethsmith/eco-rating/model"
	"github.com/ethsmith/eco-rating/rating"
	"github.com/ethsmith/eco-rating/rating/probability"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
)

// DemoParser wraps the demoinfocs parser and manages match state and logging.
// It processes CS2 demo files and extracts comprehensive player statistics.
type DemoParser struct {
	parser    demoinfocs.Parser
	state     *MatchState
	logger    *Logger
	collector *probability.DataCollector
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
		parser:    p,
		state:     state,
		logger:    NewLogger(enableLogging),
		collector: probability.NewDataCollector(),
	}

	dp.registerHandlers()
	return dp
}

// GetCollector returns the probability data collector for merging in cumulative mode.
func (d *DemoParser) GetCollector() *probability.DataCollector {
	return d.collector
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
// Returns an error if parsing fails. Truncated demos (ErrUnexpectedEndOfDemo)
// are handled gracefully — stats collected up to the truncation point are kept.
func (d *DemoParser) Parse() error {
	if err := d.parser.ParseToEnd(); err != nil {
		if errors.Is(err, demoinfocs.ErrUnexpectedEndOfDemo) {
			log.Printf("Warning: demo truncated (unexpected EOF), using partial data")
		} else {
			return fmt.Errorf("failed to parse demo: %w", err)
		}
	}
	d.computeDerivedStats()
	return nil
}

// computeDerivedStats calculates all derived metrics for each player after parsing.
func (d *DemoParser) computeDerivedStats() {

	for _, p := range d.state.Players {
		if p.RoundsPlayed > 0 {
			rounds := float64(p.RoundsPlayed)
			p.ADR = float64(p.Damage) / rounds
			p.KPR = float64(p.Kills) / rounds
			p.DPR = float64(p.Deaths) / rounds
			p.KAST = p.KAST / rounds
			p.Survival = p.Survival / rounds

			p.AWPKillsPerRound = float64(p.AWPKills) / rounds

			// Calculate HLTV rating using centralized function
			survivals := int(p.Survival * rounds)
			p.HLTVRating = rating.ComputeHLTVRating(rating.HLTVInput{
				RoundsPlayed: p.RoundsPlayed,
				Kills:        p.Kills,
				Deaths:       p.Deaths,
				Survivals:    survivals,
				MultiKills:   p.MultiKillsRaw,
			})

			// Pistol round rating
			if p.PistolRoundsPlayed > 0 {
				p.PistolRoundRating = rating.ComputePistolRoundRating(
					p.PistolRoundsPlayed, p.PistolRoundKills, p.PistolRoundDeaths,
					p.PistolRoundSurvivals, p.PistolRoundMultiKills)
			}

			// Side-specific HLTV ratings
			if p.TRoundsPlayed > 0 {
				p.TRating = rating.ComputeSideHLTVRating(
					p.TRoundsPlayed, p.TKills, p.TDeaths, p.TSurvivals, p.TMultiKills)
			}

			if p.CTRoundsPlayed > 0 {
				p.CTRating = rating.ComputeSideHLTVRating(
					p.CTRoundsPlayed, p.CTKills, p.CTDeaths, p.CTSurvivals, p.CTMultiKills)
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
			p.HeadshotPct = float64(p.Headshots) / float64(p.Kills)
			p.ManAdvantageKillsPct = float64(p.ManAdvantageKills) / float64(p.Kills)
		}

		if p.Deaths > 0 {
			p.ManDisadvantageDeathsPct = float64(p.ManDisadvantageDeaths) / float64(p.Deaths)
		}

		if p.KillsWithTTK > 0 {
			p.AvgTimeToKill = p.TotalTimeToKill / float64(p.KillsWithTTK)
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

		// Compute probability-based swing metrics
		if p.RoundsPlayed > 0 {
			rounds := float64(p.RoundsPlayed)
			p.ProbabilitySwingPerRound = p.ProbabilitySwing / rounds
			// SwingRating: scale swing to rating (0% = 1.0, +4% = 1.4, -3% = 0.7)
			p.SwingRating = 1.0 + (p.ProbabilitySwingPerRound * 10.0)
			if p.SwingRating < 0.5 {
				p.SwingRating = 0.5
			} else if p.SwingRating > 1.5 {
				p.SwingRating = 1.5
			}
		}

		p.FinalRating = rating.ComputeFinalRating(p)

		if p.TRoundsPlayed > 0 {
			p.TEcoRating = rating.ComputeSideRating(
				p.TRoundsPlayed, p.TKills, p.TDeaths, p.TDamage, p.TEcoKillValue,
				p.TProbabilitySwing, p.TKAST, p.TMultiKills, p.TClutchRounds, p.TClutchWins)
		}
		if p.TKills > 0 {
			p.TManAdvantageKillsPct = float64(p.TManAdvantageKills) / float64(p.TKills)
		}
		if p.TDeaths > 0 {
			p.TManDisadvantageDeathsPct = float64(p.TManDisadvantageDeaths) / float64(p.TDeaths)
		}
		if p.CTRoundsPlayed > 0 {
			p.CTEcoRating = rating.ComputeSideRating(
				p.CTRoundsPlayed, p.CTKills, p.CTDeaths, p.CTDamage, p.CTEcoKillValue,
				p.CTProbabilitySwing, p.CTKAST, p.CTMultiKills, p.CTClutchRounds, p.CTClutchWins)
		}
		if p.CTKills > 0 {
			p.CTManAdvantageKillsPct = float64(p.CTManAdvantageKills) / float64(p.CTKills)
		}
		if p.CTDeaths > 0 {
			p.CTManDisadvantageDeathsPct = float64(p.CTManDisadvantageDeaths) / float64(p.CTDeaths)
		}

		// Compute ML Pipeline derived metrics
		d.computeMLFeatures(p)

		d.logger.LogPlayerSummary(p.Name, p.Kills, p.Deaths, p.Damage, p.EcoKillValue, p.EcoDeathValue, p.FinalRating)
	}
}

// computeMLFeatures calculates ML pipeline features for the CS2 Synergy & Role-Gap Engine.
func (d *DemoParser) computeMLFeatures(p *model.PlayerStats) {
	rounds := float64(p.RoundsPlayed)
	if rounds == 0 {
		return
	}

	// Space Creation Index (SCI) components
	p.UncontestedAdvance = p.TotalUncontestedAdvance / rounds
	if p.TradeWindowCount > 0 {
		p.TradeWindowMedian = p.TradeWindowSum / float64(p.TradeWindowCount)
	}
	// SCI = 0.4 * Advance + 0.35 * Displacement + 0.25 * (1 / TradeWindow)
	advanceNorm := p.UncontestedAdvance / 500.0 // Normalize to ~0-1 range (500 units typical max)
	if advanceNorm > 1.0 {
		advanceNorm = 1.0
	}
	displacementNorm := float64(p.CrosshairDisplacement) / rounds / 3.0 // Normalize per round
	if displacementNorm > 1.0 {
		displacementNorm = 1.0
	}
	tradeWindowNorm := 0.0
	if p.TradeWindowMedian > 0 && p.TradeWindowMedian < 5.0 {
		tradeWindowNorm = 1.0 / p.TradeWindowMedian / 2.0 // Normalize: 0.5s = 1.0, 2s = 0.25
		if tradeWindowNorm > 1.0 {
			tradeWindowNorm = 1.0
		}
	}
	p.SpaceCreationIndex = 0.4*advanceNorm + 0.35*displacementNorm + 0.25*tradeWindowNorm

	// Utility Efficiency Score (UES) components
	if p.FlashesThrown > 0 {
		p.BlindToKillConversion = float64(p.FlashesWithKill) / float64(p.FlashesThrown)
	}
	if p.SmokesThrown > 0 {
		p.SmokeEffectiveness = float64(p.EffectiveSmokes) / float64(p.SmokesThrown)
	}
	// UES = 0.3*EFB + 0.3*BKC + 0.2*SmokeEff + 0.2*MolotovDelay
	efbNorm := p.ExpectedFlashBlindness / rounds / 3.0 // Normalize per round
	if efbNorm > 1.0 {
		efbNorm = 1.0
	}
	molotovDelayNorm := p.MolotovDelay / rounds / 5.0 // Normalize per round
	if molotovDelayNorm > 1.0 {
		molotovDelayNorm = 1.0
	}
	p.UtilityEfficiencyScore = 0.3*efbNorm + 0.3*p.BlindToKillConversion + 0.2*p.SmokeEffectiveness + 0.2*molotovDelayNorm

	// CT Anchor Hold Time (AHT)
	if p.AnchorHoldTimeCount > 0 {
		p.AnchorHoldTime = p.AnchorHoldTimeSum / float64(p.AnchorHoldTimeCount)
	}

	// Lurk & Timing Impact
	if p.LurkRounds > 0 {
		p.FlankSuccessRate = float64(p.LurkKills+p.LurkPlants) / float64(p.LurkRounds)
	}
	if p.ClockEfficiencyCount > 0 {
		p.ClockEfficiency = p.ClockEfficiencySum / float64(p.ClockEfficiencyCount)
	}

	// AWP metrics
	buyRounds := rounds - float64(p.PistolRoundsPlayed) // Exclude pistol rounds
	if buyRounds > 0 {
		p.AWPUsageRate = float64(p.AWPBuyRounds) / buyRounds
	}
	if p.AWPOpeningDuelAttempts > 0 {
		p.AWPOpeningDuelWinRate = float64(p.AWPOpeningDuelWins) / float64(p.AWPOpeningDuelAttempts)
	}

	// Economy efficiency
	if rounds > 0 {
		p.ResourceScore = p.TotalEquipmentValue / rounds
	}
	if p.TotalEquipmentValue > 0 {
		p.DamagePerDollar = float64(p.Damage) / p.TotalEquipmentValue
		p.EconomyEfficiency = p.DamagePerDollar * 1000.0 // Scale for readability
	}

	// Spatial metrics
	if p.CrossfireDistanceCount > 0 {
		p.CrossfirePartnerDistance = p.CrossfireDistanceSum / float64(p.CrossfireDistanceCount)
	}
	if p.EntrySpacingCount > 1 {
		mean := p.EntrySpacingSum / float64(p.EntrySpacingCount)
		variance := (p.EntrySpacingSumSq / float64(p.EntrySpacingCount)) - (mean * mean)
		if variance > 0 {
			p.EntrySpacingSD = sqrt(variance)
		}
	}

	// Entry/First Contact metrics
	p.EntryAttemptRate = float64(p.OpeningAttempts) / rounds
	p.FirstContactRate = float64(p.FirstContactRounds) / rounds

	// Mid-round utility (IGL proxy)
	if p.TRoundsPlayed > 0 || p.CTRoundsPlayed > 0 {
		totalMidRound := float64(p.MidRoundUtilityT + p.MidRoundUtilityCT)
		p.MidRoundUtilityUsage = totalMidRound / rounds
	}
}

// sqrt calculates square root using Newton's method.
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
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

// GetCrossfireFlashData returns the per-round crossfire and flash event data.
func (d *DemoParser) GetCrossfireFlashData() *model.CrossfireFlashData {
	return d.state.CrossfireFlashData
}

// InitCrossfireFlashData initializes the crossfire/flash data tracking.
// Call this before parsing if you want to collect per-round crossfire/flash events.
func (d *DemoParser) InitCrossfireFlashData() {
	d.state.CrossfireFlashData = model.NewCrossfireFlashData(d.state.MapName, "")
}
