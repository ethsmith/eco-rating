// Package extra is the extra family: stats the pre-v6 eco-rating parser had that CSC-Stats doesn't
// store (row.Extra). It is a tier 2 calculator: it reads the round and the combat, openings and
// objectives results already in out.
package extra

import (
	"github.com/ethsmith/eco-rating/engine/fact"
	"github.com/ethsmith/eco-rating/engine/round"
	"github.com/ethsmith/eco-rating/engine/row"
)

// Calc writes each player's row.Extra.
func Calc(r *round.Round, out *row.Round) {
	for id, p := range out.Players {
		// RF is the round won (objectives); kills, damage and the opening kill are combat's and
		// openings'.
		if p.RF == 1 {
			p.KillsInWonRounds = p.Kills
			p.DamageInWonRounds = p.Damage
			if p.Ok == 1 {
				p.RoundsWonAfterOpening = 1
			}
		}
		// Eco's AWP gates, as awpDeaths uses (eco has no bot rows).
		if id != 0 && p.AWPDeaths > 0 && p.EcoAWPKills == 0 {
			p.AWPDeathsNoKill = 1
		}
	}
	teamFlashes(r, out)
	lastAlive(r, out)
}

// teamFlashes counts the flashes a player landed on a teammate (not themself), and the blindness.
func teamFlashes(r *round.Round, out *row.Round) {
	for _, e := range r.Events[:r.Commit] {
		f, ok := e.Fact.(fact.Flashed)
		if !ok || e.Dropped || f.Attacker == nil || f.Player == nil {
			continue
		}
		a, v := f.Attacker, f.Player
		if a.Side != v.Side || a.IsBot || v.IsBot || a.Steam64 == v.Steam64 || !v.Alive {
			continue
		}
		p := out.Players[a.Steam64]
		if p == nil || out.Players[v.Steam64] == nil {
			continue
		}
		p.TeamFlashCount++
		p.TeamFlashTime += f.Duration
	}
}

// lastAlive marks the player who ends the round as their team's only member alive. The members
// alive come from the round's last kill or bomb event; a round with none ends with everyone alive.
// A player who disconnected alive isn't counted (D44), and bots, who share one row, are skipped.
func lastAlive(r *round.Round, out *row.Round) {
	var alive []fact.Player
	seen := false
	for _, e := range r.Events[:r.Commit] {
		if !e.Dropped && e.Alive != nil {
			alive, seen = e.Alive, true
		}
	}
	left := r.Disconnected()

	bySide := map[fact.Side][]uint64{}
	if seen {
		for _, a := range alive {
			if !left[a.Steam64] || a.IsBot {
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
