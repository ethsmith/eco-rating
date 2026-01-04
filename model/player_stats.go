package model

type PlayerStats struct {
	SteamID string
	Name    string

	RoundsPlayed int

	Kills        int
	Assists      int
	Deaths       int
	Damage       int
	OpeningKills int
	PerfectKills int
	TradeDenials int
	TradedDeaths int

	MultiKills [6]int // index = kills in round

	RoundImpact float64
	Survival    float64
	KAST        float64
	EconImpact  float64

	// Eco-adjusted values
	EcoKillValue  float64 // Sum of eco-adjusted kill values
	EcoDeathValue float64 // Sum of eco-adjusted death penalties

	// Round Swing - measures contribution to round wins/losses
	RoundSwing    float64 // Cumulative round swing score
	RoundsWon     int     // Rounds where player's team won
	ClutchRounds  int     // Rounds where player was last alive
	ClutchWins    int     // Clutch rounds won

	FinalRating float64
}
