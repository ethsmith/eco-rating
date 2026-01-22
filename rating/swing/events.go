// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// There may be mistakes in the comments. Please verify accuracy.
// =============================================================================

// Package swing implements probability-based player impact calculation.
// This file defines the event types that can affect round swing:
// kills, bomb plants, bomb defuses, and bomb explosions.
package swing

import "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"

// RoundEvent is the interface for all events that can affect round swing.
type RoundEvent interface {
	GetTimeInRound() float64
	GetType() EventType
}

// EventType identifies the type of round event.
type EventType int

const (
	EventKill EventType = iota
	EventBombPlant
	EventBombDefuse
	EventBombExplode
)

// KillEvent represents a kill during the round.
type KillEvent struct {
	TimeInRound         float64
	KillerID            uint64
	VictimID            uint64
	KillerSide          common.Team
	VictimSide          common.Team
	KillerEquip         float64
	VictimEquip         float64
	IsTradeKill         bool
	IsHeadshot          bool
	TotalDamageToVictim int
	KillerDamageDealt   int
	DamageContributors  []DamageContributor
	FlashAssists        []FlashAssist
}

func (e *KillEvent) GetTimeInRound() float64 { return e.TimeInRound }
func (e *KillEvent) GetType() EventType      { return EventKill }

// DamageContributor tracks a player's damage contribution to a kill.
type DamageContributor struct {
	PlayerID uint64
	Damage   int
}

// FlashAssist tracks flash assist details.
type FlashAssist struct {
	PlayerID uint64
	Duration float64 // Seconds the victim was flashed
}

// BombPlantEvent represents a bomb plant.
type BombPlantEvent struct {
	TimeInRound float64
	PlanterID   uint64
}

func (e *BombPlantEvent) GetTimeInRound() float64 { return e.TimeInRound }
func (e *BombPlantEvent) GetType() EventType      { return EventBombPlant }

// BombDefuseEvent represents a bomb defuse.
type BombDefuseEvent struct {
	TimeInRound float64
	DefuserID   uint64
}

func (e *BombDefuseEvent) GetTimeInRound() float64 { return e.TimeInRound }
func (e *BombDefuseEvent) GetType() EventType      { return EventBombDefuse }

// BombExplodeEvent represents a bomb explosion.
type BombExplodeEvent struct {
	TimeInRound float64
}

func (e *BombExplodeEvent) GetTimeInRound() float64 { return e.TimeInRound }
func (e *BombExplodeEvent) GetType() EventType      { return EventBombExplode }

// RoundResult contains the outcome of a round for swing calculation.
type RoundResult struct {
	Winner       common.Team
	EndReason    RoundEndReason
	Survivors    []uint64
	SurvivorSide common.Team
}

// RoundEndReason identifies how the round ended.
type RoundEndReason int

const (
	ReasonElimination RoundEndReason = iota
	ReasonBombExploded
	ReasonBombDefused
	ReasonTimeExpired
)
