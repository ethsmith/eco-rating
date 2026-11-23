package fold

import (
	"maps"
	"slices"

	"github.com/ethsmith/eco-rating/engine/row"
)

// The buy scopes a whole-match row is rated over (Core-planning#589, #590).
const (
	scopePistol = iota
	scopeEco
	scopeForce
	scopeFull
	scopeAntiEco
	numBuys
)

// rateBuys rates every whole-match row over the rounds of each of its team's buys, as a side is
// rated: the main CSC formula against the scope's own averages.
func rateBuys(m *Match, buys [numBuys]*scope, tallies map[*row.Player]*tally) {
	for i, s := range buys {
		a := average(s.rows)
		for _, id := range slices.Sorted(maps.Keys(s.rows)) {
			p := s.rows[id]
			_, r := cscRating(p, a, s.teamsWinPoints[id], s.normalizer[id])
			w := m.Whole[id]
			switch i {
			case scopePistol:
				w.PistolRounds, w.PistolRoundsWon, w.PistolRoundRating = p.Rounds, p.RF, r
				w.PistolRoundKills, w.PistolRoundDeaths, w.PistolRoundDamage = p.Kills, p.Deaths, p.Damage
				w.PistolRoundSurvivals = tallies[p].survived // D44: a disconnect alive isn't a survival
				w.PistolRoundMultiKills = p.TwoK + p.ThreeK + p.FourK + p.FiveK
			case scopeEco:
				w.EcoBuyRounds, w.EcoBuyRoundsWon, w.EcoBuyRating = p.Rounds, p.RF, r
			case scopeForce:
				w.ForceBuyRounds, w.ForceBuyRoundsWon, w.ForceBuyRating = p.Rounds, p.RF, r
			case scopeFull:
				w.FullBuyRounds, w.FullBuyRoundsWon, w.FullBuyRating = p.Rounds, p.RF, r
			case scopeAntiEco:
				w.AntiEcoRounds, w.AntiEcoRoundsWon, w.AntiEcoRating = p.Rounds, p.RF, r
			}
		}
	}
}

// buyScopes are the scopes a round of the team's buy counts in.
func buyScopes(own, opponent row.Buy) []int {
	switch own {
	case row.BuyPistol:
		return []int{scopePistol}
	case row.BuyEco:
		return []int{scopeEco}
	case row.BuyForce:
		return []int{scopeForce}
	case row.BuyFull:
		if opponent == row.BuyEco {
			return []int{scopeFull, scopeAntiEco}
		}
		return []int{scopeFull}
	}
	return nil
}
