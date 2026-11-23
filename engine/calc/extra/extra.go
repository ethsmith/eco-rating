// Package extra is the extra family: stats the pre-v6 eco-rating parser had that CSC-Stats doesn't
// store (row.Extra). It is a tier 2 calculator: it reads the round and the combat, openings and
// objectives results already in out.
//
// Where the old parser's definition was a timing artifact, this follows the engine's own rules
// instead: a round's result and its survivors are taken at RoundEnd, and a teammate flash is
// filtered like an enemy one (CSC's ef).
package extra

import (
	"github.com/ethsmith/eco-rating/engine/fact"
	"github.com/ethsmith/eco-rating/engine/round"
	"github.com/ethsmith/eco-rating/engine/row"
)

// Calc writes each player's row.Extra.
func Calc(r *round.Round, out *row.Round) {
	for id, p := range out.Players {
		// RF is the round won (objectives). Kills and damage are combat's and count up to the
		// round's commit, as the kills and damage columns do, so these stay a subset of them.
		if p.RF == 1 {
			p.KillsInWonRounds = p.Kills
			p.DamageInWonRounds = p.Damage
			if p.Ok == 1 { // the round's first enemy kill (openings)
				p.RoundsWonAfterOpening = 1
			}
		}
		// Died carrying an AWP without an AWP kill this round, by eco's gates as awpDeaths uses
		// (eco has no bot rows).
		if id != 0 && p.AWPDeaths > 0 && p.EcoAWPKills == 0 {
			p.AWPDeathsNoKill = 1
		}
	}
	teamFlashes(r, out)
	lastAlive(r, out)
}

// teamFlashes counts the teammates a player blinded and the blindness, with the filter CSC's
// enemy flashes (ef) use (round.Deriver.flashed) except for the side: a blind over 1 s on an alive
// player who isn't already blind for longer. A sub-second flick isn't a team flash, and the two
// counts compare like for like.
func teamFlashes(r *round.Round, out *row.Round) {
	for _, e := range r.Events[:r.Commit] {
		f, ok := e.Fact.(fact.Flashed)
		if !ok || e.Dropped || f.Attacker == nil || f.Player == nil {
			continue
		}
		a, v := f.Attacker, f.Player
		if a.Side != v.Side || (a.Steam64 == v.Steam64 && a.UserID == v.UserID) { // bots share Steam64 0
			continue
		}
		if f.Duration <= 1 || !v.Alive || v.FlashDuration >= f.Duration+1 {
			continue
		}
		p := out.Players[a.Steam64]
		if p == nil || out.Players[v.Steam64] == nil { // both round members
			continue
		}
		p.TeamFlashCount++
		p.TeamFlashTime += f.Duration
	}
}

// lastAlive marks the player who ends the round, at RoundEnd, as their team's only member alive.
// Kills after RoundEnd (hunting a saving player before the official end) don't change who was
// last alive when the round was decided. The members alive come from the last kill or bomb event
// before RoundEnd; with none, everyone who hadn't disconnected was alive. A player who
// disconnected alive isn't counted (D44), and bots, who share one row, are skipped.
func lastAlive(r *round.Round, out *row.Round) {
	end := -1
	for _, w := range r.WinList() {
		if w.Event < r.Commit {
			end = w.Event // the latest RoundEnd before the commit is the round's (Q67)
		}
	}
	if end < 0 {
		return
	}

	var alive []fact.Player
	seen := false
	left := map[uint64]bool{}
	for _, e := range r.Events[:end] {
		if e.Dropped {
			continue
		}
		if e.Alive != nil {
			alive, seen = e.Alive, true
		}
		if d, ok := e.Fact.(fact.Disconnect); ok {
			left[d.Steam64] = true
		}
	}

	bySide := map[fact.Side][]uint64{}
	if seen {
		for _, a := range alive {
			if a.IsBot || !left[a.Steam64] {
				bySide[a.Side] = append(bySide[a.Side], a.Steam64)
			}
		}
	} else {
		for id, p := range out.Players {
			if !left[id] {
				bySide[fact.Side(p.Side)] = append(bySide[fact.Side(p.Side)], id)
			}
		}
	}
	for _, ids := range bySide {
		if len(ids) != 1 || ids[0] == 0 {
			continue
		}
		if p := out.Players[ids[0]]; p != nil {
			p.LastAliveRounds = 1
		}
	}
}
