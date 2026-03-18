// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// =============================================================================

// Package model defines the core data structures for player and round statistics.

package model

// Position represents a 3D coordinate in the game world.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Distance calculates the Euclidean distance between two positions.
func (p Position) Distance(other Position) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	dz := p.Z - other.Z
	return sqrt(dx*dx + dy*dy + dz*dz)
}

// Distance2D calculates the 2D distance (ignoring Z) between two positions.
func (p Position) Distance2D(other Position) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return sqrt(dx*dx + dy*dy)
}

// sqrt is a simple square root implementation to avoid importing math in model.
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

// DamageContribution tracks damage dealt by a player to a specific victim.
type DamageContribution struct {
	PlayerID uint64
	Damage   int
}

// FlashAssistInfo tracks flash assist details for a kill.
type FlashAssistInfo struct {
	PlayerID uint64
	Duration float64 // Flash duration in seconds
}

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

	// Probability-based swing tracking (new for v3.0)
	ProbabilitySwing   float64             // Win probability delta contribution
	LastDeathSwing     float64             // Most recent death swing (for trade refund calculation)
	EquipmentValue     float64             // Player's equipment value at round start
	SwingContributions []SwingContribution // Detailed swing events for this round

	// ML Pipeline Features (per-round tracking for CS2 Synergy & Role-Gap Engine)
	// Spatial tracking
	StartPosition          Position // Position at round start (after freeze time)
	FirstContactPosition   Position // Position at first contact/damage
	UncontestedAdvance     float64  // Meters traveled before first contact
	CrosshairDisplacements int      // Enemy aim vectors displaced toward this player

	// Utility tracking
	SmokesThrown          int     // Smokes thrown this round
	EffectiveSmokes       int     // Smokes within 80cm of optimal position
	MolotovsThrown        int     // Molotovs/incendiaries thrown
	MolotovDelayTime      float64 // Seconds enemies delayed by molotovs
	FlashBlindDuration    float64 // Total expected blind duration caused
	FlashesLeadingToKills int     // Flashes that led to teammate kills within 1.5s
	MidRoundUtility       int     // Utility used after first 20s of round

	// Role detection
	IsLurking          bool    // Player was lurking (far from team during execute)
	IsAnchoring        bool    // Player was anchoring a site (CT)
	AnchorHoldTime     float64 // Seconds survived after execute utility landed
	ExecuteUtilityTime float64 // Time when execute utility first landed on anchor's site
	HadAWPAtStart      bool    // Player had AWP at round start
	BuyRoundType       string  // "full", "force", "eco", "pistol"

	// Positioning
	DistanceToEntry          float64 // Distance to designated entry player
	CrossfirePartnerDist     float64 // Distance to crossfire partner
	TeamCentroidDistance     float64 // Distance from team centroid
	WasInCrossfire           bool    // Was in crossfire position during kill
	CrossfirePartnerDistance float64 // Distance to crossfire partner at kill time

	// First contact tracking
	WasFirstContact    bool    // Was first player to take damage this round
	FirstContactTime   float64 // Time of first contact
	ClockTimeAtContact float64 // Round time remaining at first contact (for lurk)
}

// SwingContribution captures a single event's impact on probability swing.
type SwingContribution struct {
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	TimeInRound   float64 `json:"time_in_round,omitempty"`
	Opponent      string  `json:"opponent,omitempty"`
	Weapon        string  `json:"weapon,omitempty"`
	IsTrade       bool    `json:"is_trade,omitempty"`
	IsHeadshot    bool    `json:"is_headshot,omitempty"`
	EcoMultiplier float64 `json:"eco_multiplier,omitempty"`
	Notes         string  `json:"notes,omitempty"`
}

// AddSwingContribution appends a swing contribution entry for the round.
func (r *RoundStats) AddSwingContribution(contribution SwingContribution) {
	if contribution.Amount == 0 {
		return
	}
	r.SwingContributions = append(r.SwingContributions, contribution)
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
