package extra

import (
	"testing"

	"github.com/ethsmith/eco-rating/engine/fact"
	"github.com/ethsmith/eco-rating/engine/match"
	"github.com/ethsmith/eco-rating/engine/round"
	"github.com/ethsmith/eco-rating/engine/row"
)

func ev(f fact.Fact, x round.Context) round.Event {
	return round.Event{Event: match.Event{Fact: f, Live: true}, Context: x}
}

func roundEnd(winner fact.Side) round.Event {
	return round.Event{Event: match.Event{Fact: fact.RoundEnd{Winner: winner}, Live: true, Win: true}}
}

func pl(id uint64, side fact.Side) fact.Player {
	return fact.Player{Steam64: id, Side: side, Alive: true}
}

func TestPerRoundCounters(t *testing.T) {
	out := &row.Round{Players: map[uint64]*row.Player{
		1: {Combat: row.Combat{Kills: 2, Damage: 180}, Openings: row.Openings{Ok: 1}, Objectives: row.Objectives{RF: 1}},
		2: {Combat: row.Combat{Kills: 1, Damage: 90, AWPDeaths: 1}, Objectives: row.Objectives{RF: 0}},
		3: {Combat: row.Combat{AWPDeaths: 1, EcoAWPKills: 1}, Objectives: row.Objectives{RF: 1}},
		0: {Combat: row.Combat{AWPDeaths: 1}}, // the bot row has no eco stats
	}}
	Calc(&round.Round{}, out)
	if got, want := out.Players[1].Extra, (row.Extra{KillsInWonRounds: 2, DamageInWonRounds: 180, RoundsWonAfterOpening: 1}); got != want {
		t.Errorf("winner with the opening kill: %+v, want %+v", got, want)
	}
	if got, want := out.Players[2].Extra, (row.Extra{AWPDeathsNoKill: 1}); got != want {
		t.Errorf("loser who lost an AWP without an AWP kill: %+v, want %+v", got, want)
	}
	if got := out.Players[3].AWPDeathsNoKill; got != 0 {
		t.Errorf("lost an AWP after an AWP kill: awpDeathsNoKill %d, want 0", got)
	}
	if got := out.Players[0].AWPDeathsNoKill; got != 0 {
		t.Errorf("bot row: awpDeathsNoKill %d, want 0", got)
	}
}

func TestTeamFlashes(t *testing.T) {
	a, b, c := pl(1, fact.SideT), pl(2, fact.SideT), pl(3, fact.SideCT)
	blind := b
	blind.FlashDuration = 4 // already blind for longer than the new flash plus 1 s
	dead := b
	dead.Alive = false
	dropped := ev(fact.Flashed{Attacker: &a, Player: &b, Duration: 2}, round.Context{})
	dropped.Dropped = true
	r := &round.Round{Events: []round.Event{
		ev(fact.Flashed{Attacker: &a, Player: &b, Duration: 2.5}, round.Context{}),   // counts
		ev(fact.Flashed{Attacker: &a, Player: &b, Duration: 0.8}, round.Context{}),   // a flick: under 1 s
		ev(fact.Flashed{Attacker: &a, Player: &blind, Duration: 2}, round.Context{}), // already blinder
		ev(fact.Flashed{Attacker: &a, Player: &dead, Duration: 3}, round.Context{}),  // not alive
		ev(fact.Flashed{Attacker: &a, Player: &a, Duration: 3}, round.Context{}),     // self
		ev(fact.Flashed{Attacker: &a, Player: &c, Duration: 3}, round.Context{}),     // an enemy: ef's
		ev(fact.Flashed{Attacker: &b, Player: &a, Duration: 1.5}, round.Context{}),   // counts for 2
		dropped,
		ev(fact.Flashed{Attacker: &b, Player: &a, Duration: 3}, round.Context{}), // after the commit
	}, Commit: 8}
	out := &row.Round{Players: map[uint64]*row.Player{1: {}, 2: {}, 3: {}}}
	Calc(r, out)
	if p := out.Players[1]; p.TeamFlashCount != 1 || p.TeamFlashTime != 2.5 {
		t.Errorf("player 1: %d team flashes, %v s; want 1, 2.5", p.TeamFlashCount, p.TeamFlashTime)
	}
	if p := out.Players[2]; p.TeamFlashCount != 1 || p.TeamFlashTime != 1.5 {
		t.Errorf("player 2: %d team flashes, %v s; want 1, 1.5", p.TeamFlashCount, p.TeamFlashTime)
	}
	if p := out.Players[3]; p.TeamFlashCount != 0 {
		t.Errorf("player 3: %d team flashes, want 0", p.TeamFlashCount)
	}
}

func TestLastAliveAtRoundEnd(t *testing.T) {
	t1, t2, ct1, ct2 := pl(1, fact.SideT), pl(2, fact.SideT), pl(3, fact.SideCT), pl(4, fact.SideCT)
	r := &round.Round{Events: []round.Event{
		// CT 4 dies: T 1 and 2 and CT 3 alive.
		ev(fact.Kill{Killer: &t1, Victim: &ct2}, round.Context{Alive: []fact.Player{t1, t2, ct1}}),
		// T 2 dies: T 1 is the T side's last alive, CT 3 the CT side's.
		ev(fact.Kill{Killer: &ct1, Victim: &t2}, round.Context{Alive: []fact.Player{t1, ct1}}),
		roundEnd(fact.SideCT),
		// After RoundEnd, CT 3 hunts T 1 down: it doesn't change who ended the round last alive.
		ev(fact.Kill{Killer: &ct1, Victim: &t1}, round.Context{Alive: []fact.Player{ct1}}),
	}, Commit: 4, Wins: []int{3}}
	out := &row.Round{Players: map[uint64]*row.Player{1: {}, 2: {}, 3: {}, 4: {}}}
	Calc(r, out)
	for id, want := range map[uint64]int{1: 1, 2: 0, 3: 1, 4: 0} {
		if got := out.Players[id].LastAliveRounds; got != want {
			t.Errorf("player %d: lastAliveRounds %d, want %d", id, got, want)
		}
	}
}

func TestLastAliveEdges(t *testing.T) {
	t1, t2, ct1 := pl(1, fact.SideT), pl(2, fact.SideT), pl(3, fact.SideCT)
	// Two Ts alive at RoundEnd; one is killed afterwards. Neither ended the round as last alive.
	r := &round.Round{Events: []round.Event{
		ev(fact.Kill{Killer: &t1, Victim: &fact.Player{Steam64: 4, Side: fact.SideCT}}, round.Context{Alive: []fact.Player{t1, t2, ct1}}),
		roundEnd(fact.SideT),
		ev(fact.Kill{Killer: &ct1, Victim: &t2}, round.Context{Alive: []fact.Player{t1, ct1}}),
	}, Commit: 3, Wins: []int{2}}
	out := &row.Round{Players: map[uint64]*row.Player{1: {}, 2: {}, 3: {}, 4: {}}}
	Calc(r, out)
	if out.Players[1].LastAliveRounds != 0 || out.Players[2].LastAliveRounds != 0 {
		t.Errorf("two Ts alive at RoundEnd: %d, %d; want 0, 0", out.Players[1].LastAliveRounds, out.Players[2].LastAliveRounds)
	}
	if out.Players[3].LastAliveRounds != 1 {
		t.Errorf("the lone CT at RoundEnd: %d, want 1", out.Players[3].LastAliveRounds)
	}

	// A player who disconnected alive is no survivor (D44): their teammate is last alive.
	r = &round.Round{Events: []round.Event{
		ev(fact.Kill{Killer: &ct1, Victim: &fact.Player{Steam64: 5, Side: fact.SideT}}, round.Context{Alive: []fact.Player{t1, t2, ct1}}),
		ev(fact.Disconnect{Steam64: 2}, round.Context{}),
		roundEnd(fact.SideCT),
	}, Commit: 3, Wins: []int{2}}
	out = &row.Round{Players: map[uint64]*row.Player{1: {}, 2: {}, 3: {}, 5: {}}}
	Calc(r, out)
	if out.Players[1].LastAliveRounds != 1 || out.Players[2].LastAliveRounds != 0 {
		t.Errorf("teammate disconnected alive: %d, %d; want 1, 0", out.Players[1].LastAliveRounds, out.Players[2].LastAliveRounds)
	}

	// No RoundEnd before the commit: no round result, no last alive.
	r = &round.Round{Events: []round.Event{
		ev(fact.Kill{Killer: &ct1, Victim: &t2}, round.Context{Alive: []fact.Player{t1, ct1}}),
	}, Commit: 1}
	out = &row.Round{Players: map[uint64]*row.Player{1: {}, 2: {}, 3: {}}}
	Calc(r, out)
	if out.Players[1].LastAliveRounds != 0 {
		t.Errorf("no RoundEnd: lastAliveRounds %d, want 0", out.Players[1].LastAliveRounds)
	}
}
