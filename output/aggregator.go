package output

import (
	"eco-rating/model"
)

// MultiKillStats holds multi-kill counts with explicit labels
type MultiKillStats struct {
	OneK   int `json:"1k"`
	TwoK   int `json:"2k"`
	ThreeK int `json:"3k"`
	FourK  int `json:"4k"`
	FiveK  int `json:"5k"`
}

// AggregatedStats holds cumulative stats for a player across multiple games
type AggregatedStats struct {
	SteamID    string `json:"steam_id"`
	Name       string `json:"name"`
	GamesCount int    `json:"games_count"`

	// Cumulative totals
	RoundsPlayed int `json:"rounds_played"`
	Kills        int `json:"kills"`
	Assists      int `json:"assists"`
	Deaths       int `json:"deaths"`
	Damage       int `json:"damage"`
	OpeningKills int `json:"opening_kills"`

	// Per-round averages (calculated from totals)
	ADR float64 `json:"adr"`
	KPR float64 `json:"kpr"`
	DPR float64 `json:"dpr"`

	PerfectKills int `json:"perfect_kills"`
	TradeDenials int `json:"trade_denials"`
	TradedDeaths int `json:"traded_deaths"`

	// AWP stats
	AWPKills         int     `json:"awp_kills"`
	AWPKillsPerRound float64 `json:"awp_kills_per_round"`

	MultiKills MultiKillStats `json:"multi_kills"`

	// Averages across games
	RoundImpact float64 `json:"round_impact"`
	Survival    float64 `json:"survival"`
	KAST        float64 `json:"kast"`
	EconImpact  float64 `json:"econ_impact"`

	// Eco-adjusted values (cumulative)
	EcoKillValue  float64 `json:"eco_kill_value"`
	EcoDeathValue float64 `json:"eco_death_value"`

	// Round swing stats
	RoundSwing   float64 `json:"round_swing"`
	RoundsWon    int     `json:"rounds_won"`
	ClutchRounds int     `json:"clutch_rounds"`
	ClutchWins   int     `json:"clutch_wins"`

	// Utility and misc stats
	UtilityDamage      int     `json:"utility_damage"`
	TeamFlashCount     int     `json:"team_flash_count"`
	TeamFlashDuration  float64 `json:"team_flash_duration"`
	ExitFrags          int     `json:"exit_frags"`
	AWPDeaths          int     `json:"awp_deaths"`
	AWPDeathsNoKill    int     `json:"awp_deaths_no_kill"`
	KnifeKills         int     `json:"knife_kills"`
	PistolVsRifleKills int     `json:"pistol_vs_rifle_kills"`
	TradeKills         int     `json:"trade_kills"`
	FastTrades         int     `json:"fast_trades"`
	EarlyDeaths        int     `json:"early_deaths"`

	// Rating - average across games
	FinalRating float64 `json:"final_rating"`

	// Track sum of ratings for averaging
	ratingSum float64
}

// Aggregator combines stats from multiple games
type Aggregator struct {
	Players map[string]*AggregatedStats // keyed by SteamID
}

// NewAggregator creates a new stats aggregator
func NewAggregator() *Aggregator {
	return &Aggregator{
		Players: make(map[string]*AggregatedStats),
	}
}

// AddGame adds stats from a single game to the aggregator
func (a *Aggregator) AddGame(players map[uint64]*model.PlayerStats) {
	for _, p := range players {
		agg := a.ensurePlayer(p.SteamID, p.Name)
		agg.GamesCount++

		// Add cumulative stats
		agg.RoundsPlayed += p.RoundsPlayed
		agg.Kills += p.Kills
		agg.Assists += p.Assists
		agg.Deaths += p.Deaths
		agg.Damage += p.Damage
		agg.OpeningKills += p.OpeningKills
		agg.PerfectKills += p.PerfectKills
		agg.TradeDenials += p.TradeDenials
		agg.TradedDeaths += p.TradedDeaths

		// AWP stats
		agg.AWPKills += p.AWPKills

		// Multi-kills
		agg.MultiKills.OneK += p.MultiKills[1]
		agg.MultiKills.TwoK += p.MultiKills[2]
		agg.MultiKills.ThreeK += p.MultiKills[3]
		agg.MultiKills.FourK += p.MultiKills[4]
		agg.MultiKills.FiveK += p.MultiKills[5]

		// Eco values
		agg.EcoKillValue += p.EcoKillValue
		agg.EcoDeathValue += p.EcoDeathValue

		// Round swing stats
		agg.RoundSwing += p.RoundSwing
		agg.RoundsWon += p.RoundsWon
		agg.ClutchRounds += p.ClutchRounds
		agg.ClutchWins += p.ClutchWins

		// Utility and misc
		agg.UtilityDamage += p.UtilityDamage
		agg.TeamFlashCount += p.TeamFlashCount
		agg.TeamFlashDuration += p.TeamFlashDuration
		agg.ExitFrags += p.ExitFrags
		agg.AWPDeaths += p.AWPDeaths
		agg.AWPDeathsNoKill += p.AWPDeathsNoKill
		agg.KnifeKills += p.KnifeKills
		agg.PistolVsRifleKills += p.PistolVsRifleKills
		agg.TradeKills += p.TradeKills
		agg.FastTrades += p.FastTrades
		agg.EarlyDeaths += p.EarlyDeaths

		// Track rating sum for averaging
		agg.ratingSum += p.FinalRating

		// Track weighted sums for percentage-based stats (weighted by rounds played)
		// These are already percentages, so we need to weight them
		rounds := float64(p.RoundsPlayed)
		agg.RoundImpact += p.RoundImpact * rounds
		agg.Survival += p.Survival * rounds
		agg.KAST += p.KAST * rounds
		agg.EconImpact += p.EconImpact * rounds
	}
}

// Finalize calculates final averages and per-round stats
func (a *Aggregator) Finalize() {
	for _, agg := range a.Players {
		if agg.RoundsPlayed > 0 {
			rounds := float64(agg.RoundsPlayed)

			// Calculate per-round stats from totals
			agg.ADR = float64(agg.Damage) / rounds
			agg.KPR = float64(agg.Kills) / rounds
			agg.DPR = float64(agg.Deaths) / rounds
			agg.AWPKillsPerRound = float64(agg.AWPKills) / rounds

			// Calculate weighted averages for percentage stats
			agg.RoundImpact = agg.RoundImpact / rounds
			agg.Survival = agg.Survival / rounds
			agg.KAST = agg.KAST / rounds
			agg.EconImpact = agg.EconImpact / rounds
		}

		// Average rating across games
		if agg.GamesCount > 0 {
			agg.FinalRating = agg.ratingSum / float64(agg.GamesCount)
		}
	}
}

// GetResults returns the aggregated stats map
func (a *Aggregator) GetResults() map[string]*AggregatedStats {
	return a.Players
}

func (a *Aggregator) ensurePlayer(steamID, name string) *AggregatedStats {
	if _, ok := a.Players[steamID]; !ok {
		a.Players[steamID] = &AggregatedStats{
			SteamID: steamID,
			Name:    name,
		}
	}
	return a.Players[steamID]
}
