// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// =============================================================================

// Package parser provides CS2 demo file parsing functionality.
// This file defines the MatchState struct which tracks the current state of a
// match during parsing, including player stats, round stats, and trade detection.
package parser

import (
	"eco-rating/model"
	"fmt"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
)

// MatchState holds all state information during demo parsing.
// It tracks players, current round stats, and various flags for game state.
type MatchState struct {
	Players          map[uint64]*model.PlayerStats
	Round            map[uint64]*model.RoundStats
	RoundHasKill     bool
	MatchStarted     bool
	IsKnifeRound     bool
	IsPistolRound    bool
	RoundNumber      int
	MapName          string
	RecentKills      map[uint64]recentKill
	RoundStartTime   float64
	CurrentSide      string
	TeamScore        int
	EnemyScore       int
	RoundDecided     bool
	RoundDecidedAt   float64
	RecentTeamDeaths map[uint64]float64
	PendingTrades    map[uint64][]pendingTrade
}

// pendingTrade tracks a potential trade opportunity after a kill.
// If the killer is killed within the trade window by a nearby teammate,
// the original death is marked as traded.
type pendingTrade struct {
	KillerID           uint64      // Steam ID of the player who got the kill
	KillerTeam         common.Team // Team of the killer
	TeammateID         uint64      // Steam ID of the nearby teammate who could trade
	DeathTick          int         // Tick when the original death occurred
	TeammatePos        [3]float64  // Position of the teammate at death time
	PotentialTraderPos [3]float64  // Position of the potential trader
}

// recentKill tracks a recent kill for trade detection.
// Used to determine if a subsequent kill is a trade.
type recentKill struct {
	VictimID   uint64      // Steam ID of the victim
	VictimTeam common.Team // Team of the victim
	Tick       int         // Tick when the kill occurred
}

// NewMatchState creates a new MatchState with initialized maps.
func NewMatchState() *MatchState {
	return &MatchState{
		Players:          make(map[uint64]*model.PlayerStats),
		Round:            make(map[uint64]*model.RoundStats),
		RecentKills:      make(map[uint64]recentKill),
		RecentTeamDeaths: make(map[uint64]float64),
		PendingTrades:    make(map[uint64][]pendingTrade),
	}
}

// ensurePlayer returns the PlayerStats for a player, creating it if needed.
func (m *MatchState) ensurePlayer(p *common.Player) *model.PlayerStats {
	id := p.SteamID64
	if _, ok := m.Players[id]; !ok {
		m.Players[id] = &model.PlayerStats{
			SteamID: fmt.Sprintf("%d", id),
			Name:    p.Name,
		}
	}
	return m.Players[id]
}

// ensureRound returns the RoundStats for a player in the current round, creating it if needed.
func (m *MatchState) ensureRound(p *common.Player) *model.RoundStats {
	id := p.SteamID64
	if _, ok := m.Round[id]; !ok {
		m.Round[id] = &model.RoundStats{}
	}
	return m.Round[id]
}
