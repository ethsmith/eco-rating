// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// =============================================================================

// Package model defines the core data structures for player and round statistics.

package model

// RoundStats tracks a player's performance within a single round.
// This struct is populated during demo parsing and used to calculate
// per-round metrics like round swing, KAST, and clutch statistics.
type RoundStats struct {
	Kills              int
	Assists            int
	Damage             int
	Survived           bool
	Traded             bool
	GotKill            bool
	GotAssist          bool
	EconImpact         float64
	AWPKills           int
	AWPOpeningKill     bool
	TeamWon            bool
	PlayersAlive       int
	EnemiesAlive       int
	WasLastAlive       bool
	ClutchKills        int
	PlantedBomb        bool
	DefusedBomb        bool
	OpeningKill        bool
	OpeningDeath       bool
	MultiKillRound     int
	EntryFragger       bool
	ClutchAttempt      bool
	ClutchWon          bool
	ClutchSize         int
	SavedWeapons       bool
	EcoKill            bool
	AntiEcoKill        bool
	FlashAssists       int
	TradeKill          bool
	TradeDeath         bool
	FailedTrades       int
	TradeDenials       int
	SavedByTeammate    bool
	SavedTeammate      bool
	IsSupportRound     bool
	InvolvedInOpening  bool
	UtilityDamage      int
	UtilityKills       int
	SmokeDamage        int
	DeathTime          float64
	TimeAlive          float64
	KillTimes          []float64
	TradeSpeed         float64
	IsExitFrag         bool
	ExitFrags          int
	TeamFlashCount     int
	TeamFlashDuration  float64
	FlashesThrown      int
	EnemyFlashDuration float64
	AWPKill            bool
	KnifeKill          bool
	PistolVsRifleKill  bool
	HadAWP             bool
	LostAWP            bool
	IsPistolRound      bool
	PlayerSide         string
}

// RoundContext provides contextual information about the round state.
// This is used by the round swing calculator to adjust impact values
// based on round importance, score differential, and game situation.
type RoundContext struct {
	RoundNumber     int
	TotalPlayers    int
	BombPlanted     bool
	BombDefused     bool
	RoundType       string
	TimeRemaining   float64
	IsOvertimeRound bool
	MapSide         string
	TeamScore       int
	EnemyScore      int
	ScoreDiff       int
	IsMatchPoint    bool
	IsCloseGame     bool
	RoundImportance float64
	RoundDecided    bool
	RoundDecidedAt  float64
}
