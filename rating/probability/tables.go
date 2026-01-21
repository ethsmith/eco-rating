package probability

import (
	"fmt"
)

// ProbabilityTables holds all empirically-derived probability data.
type ProbabilityTables struct {
	// BaseWinProb maps state keys to T-side win probability.
	// Key format: "TvCT_bombStatus" (e.g., "5v4_none", "3v2_planted")
	BaseWinProb map[string]float64

	// DuelWinRates maps economy matchup keys to attacker win probability.
	// Key format: "attacker_vs_defender" (e.g., "awp_vs_pistol", "rifle_vs_smg")
	DuelWinRates map[string]float64

	// MapAdjustments maps map names to T-side win rate adjustments.
	// Value is the T-side win percentage (0.0-1.0) for that map.
	MapAdjustments map[string]float64
}

// NewProbabilityTables creates a new ProbabilityTables with empty maps.
func NewProbabilityTables() *ProbabilityTables {
	return &ProbabilityTables{
		BaseWinProb:    make(map[string]float64),
		DuelWinRates:   make(map[string]float64),
		MapAdjustments: make(map[string]float64),
	}
}

// GetBaseWinProbability returns the T-side win probability for a given state.
func (t *ProbabilityTables) GetBaseWinProbability(tAlive, ctAlive int, bombPlanted bool) float64 {
	bombStatus := "none"
	if bombPlanted {
		bombStatus = "planted"
	}
	key := stateKeyFromComponents(tAlive, ctAlive, bombStatus)

	if prob, ok := t.BaseWinProb[key]; ok {
		return prob
	}

	// Fallback: calculate based on player advantage
	return t.calculateFallbackProbability(tAlive, ctAlive, bombPlanted)
}

// calculateFallbackProbability provides a reasonable estimate when no empirical data exists.
func (t *ProbabilityTables) calculateFallbackProbability(tAlive, ctAlive int, bombPlanted bool) float64 {
	if tAlive == 0 {
		return 0.0
	}
	if ctAlive == 0 {
		return 1.0
	}

	// Base probability from player count ratio
	total := float64(tAlive + ctAlive)
	baseProb := float64(tAlive) / total

	// CT-side advantage in equal situations (CT wins ~52% of 5v5s)
	ctAdvantage := 0.04
	baseProb -= ctAdvantage * (float64(ctAlive) / 5.0)

	// Bomb planted heavily favors T
	if bombPlanted {
		baseProb += 0.25 * (1.0 - baseProb) // Move 25% closer to 1.0
	}

	return clamp(baseProb, 0.01, 0.99)
}

// GetDuelWinRate returns the probability that the attacker wins a duel.
func (t *ProbabilityTables) GetDuelWinRate(attackerEcon, defenderEcon EconomyCategory) float64 {
	key := fmt.Sprintf("%s_vs_%s", attackerEcon.String(), defenderEcon.String())

	if rate, ok := t.DuelWinRates[key]; ok {
		return rate
	}

	// Fallback: calculate based on economy difference
	return t.calculateFallbackDuelRate(attackerEcon, defenderEcon)
}

// calculateFallbackDuelRate provides a reasonable estimate for duel outcomes.
func (t *ProbabilityTables) calculateFallbackDuelRate(attackerEcon, defenderEcon EconomyCategory) float64 {
	diff := int(attackerEcon) - int(defenderEcon)

	// Each economy tier is worth ~7-8% advantage
	adjustment := float64(diff) * 0.07

	return clamp(0.50+adjustment, 0.20, 0.80)
}

// GetMapAdjustment returns the T-side win rate for a map (default 0.50).
func (t *ProbabilityTables) GetMapAdjustment(mapName string) float64 {
	if adj, ok := t.MapAdjustments[mapName]; ok {
		return adj
	}
	return 0.50 // Default balanced
}

// clamp restricts a value to the range [min, max].
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// StateOutcome tracks win/loss outcomes for a particular game state.
type StateOutcome struct {
	TWins  int
	CTWins int
	Total  int
}

// WinRate returns the T-side win rate for this state.
func (o *StateOutcome) WinRate() float64 {
	if o.Total == 0 {
		return 0.50
	}
	return float64(o.TWins) / float64(o.Total)
}

// DuelOutcome tracks win/loss outcomes for a particular duel matchup.
type DuelOutcome struct {
	AttackerWins int
	DefenderWins int
	Total        int
}

// WinRate returns the attacker win rate for this duel type.
func (o *DuelOutcome) WinRate() float64 {
	if o.Total == 0 {
		return 0.50
	}
	return float64(o.AttackerWins) / float64(o.Total)
}

// ProbabilityDataCollector aggregates data from demo parsing to build probability tables.
type ProbabilityDataCollector struct {
	StateOutcomes map[string]*StateOutcome
	DuelOutcomes  map[string]*DuelOutcome
	MapData       map[string]*MapProbabilityData
}

// MapProbabilityData tracks win rates for a specific map.
type MapProbabilityData struct {
	TWins  int
	CTWins int
	Total  int
}

// NewProbabilityDataCollector creates a new collector.
func NewProbabilityDataCollector() *ProbabilityDataCollector {
	return &ProbabilityDataCollector{
		StateOutcomes: make(map[string]*StateOutcome),
		DuelOutcomes:  make(map[string]*DuelOutcome),
		MapData:       make(map[string]*MapProbabilityData),
	}
}

// RecordStateOutcome records the outcome of a round from a particular state.
func (c *ProbabilityDataCollector) RecordStateOutcome(state *RoundState, tWon bool) {
	key := state.StateKey()

	if _, ok := c.StateOutcomes[key]; !ok {
		c.StateOutcomes[key] = &StateOutcome{}
	}

	c.StateOutcomes[key].Total++
	if tWon {
		c.StateOutcomes[key].TWins++
	} else {
		c.StateOutcomes[key].CTWins++
	}
}

// RecordDuelOutcome records the outcome of a duel.
func (c *ProbabilityDataCollector) RecordDuelOutcome(attackerEcon, defenderEcon EconomyCategory, attackerWon bool) {
	key := fmt.Sprintf("%s_vs_%s", attackerEcon.String(), defenderEcon.String())

	if _, ok := c.DuelOutcomes[key]; !ok {
		c.DuelOutcomes[key] = &DuelOutcome{}
	}

	c.DuelOutcomes[key].Total++
	if attackerWon {
		c.DuelOutcomes[key].AttackerWins++
	} else {
		c.DuelOutcomes[key].DefenderWins++
	}
}

// RecordMapOutcome records a round outcome for a specific map.
func (c *ProbabilityDataCollector) RecordMapOutcome(mapName string, tWon bool) {
	if _, ok := c.MapData[mapName]; !ok {
		c.MapData[mapName] = &MapProbabilityData{}
	}

	c.MapData[mapName].Total++
	if tWon {
		c.MapData[mapName].TWins++
	} else {
		c.MapData[mapName].CTWins++
	}
}

// GenerateTables creates probability tables from collected data.
func (c *ProbabilityDataCollector) GenerateTables(minSampleSize int) *ProbabilityTables {
	tables := NewProbabilityTables()

	// Generate base win probabilities
	for key, outcome := range c.StateOutcomes {
		if outcome.Total >= minSampleSize {
			tables.BaseWinProb[key] = outcome.WinRate()
		}
	}

	// Generate duel win rates
	for key, outcome := range c.DuelOutcomes {
		if outcome.Total >= minSampleSize {
			tables.DuelWinRates[key] = outcome.WinRate()
		}
	}

	// Generate map adjustments
	for mapName, data := range c.MapData {
		if data.Total >= minSampleSize {
			tables.MapAdjustments[mapName] = float64(data.TWins) / float64(data.Total)
		}
	}

	return tables
}

// Merge combines another collector's data into this one.
func (c *ProbabilityDataCollector) Merge(other *ProbabilityDataCollector) {
	for key, outcome := range other.StateOutcomes {
		if _, ok := c.StateOutcomes[key]; !ok {
			c.StateOutcomes[key] = &StateOutcome{}
		}
		c.StateOutcomes[key].TWins += outcome.TWins
		c.StateOutcomes[key].CTWins += outcome.CTWins
		c.StateOutcomes[key].Total += outcome.Total
	}

	for key, outcome := range other.DuelOutcomes {
		if _, ok := c.DuelOutcomes[key]; !ok {
			c.DuelOutcomes[key] = &DuelOutcome{}
		}
		c.DuelOutcomes[key].AttackerWins += outcome.AttackerWins
		c.DuelOutcomes[key].DefenderWins += outcome.DefenderWins
		c.DuelOutcomes[key].Total += outcome.Total
	}

	for mapName, data := range other.MapData {
		if _, ok := c.MapData[mapName]; !ok {
			c.MapData[mapName] = &MapProbabilityData{}
		}
		c.MapData[mapName].TWins += data.TWins
		c.MapData[mapName].CTWins += data.CTWins
		c.MapData[mapName].Total += data.Total
	}
}
