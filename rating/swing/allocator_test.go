package swing

import (
	"math"
	"testing"

	"github.com/ethsmith/eco-rating/rating/probability"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
)

func TestAllocateKillEventZeroSum(t *testing.T) {
	engine := probability.NewDefaultEngine()
	cfg := DefaultConfig()
	allocator := NewPoolAllocator(engine, cfg)

	state := probability.NewRoundState(5, 5, "de_dust2")
	kill := &KillEvent{
		KillerID:            100,
		VictimID:            200,
		KillerSide:          common.TeamTerrorists,
		VictimSide:          common.TeamCounterTerrorists,
		KillerEquip:         5000,
		VictimEquip:         5000,
		TotalDamageToVictim: 100,
		KillerDamageDealt:   5,
		DamageContributors: []DamageContributor{
			{PlayerID: 101, Damage: 95},
		},
	}

	event, result := allocator.AllocateKillEvent(state, KillAllocationInput{Kill: kill})
	if result.RawSwing <= 0 {
		t.Fatalf("expected positive raw swing, got %f", result.RawSwing)
	}

	positive := sumAllocations(event.PositiveAlloc)
	negative := sumAllocations(event.NegativeAlloc)
	if math.Abs(positive+negative) > cfg.ZeroSumTolerance {
		t.Fatalf("kill event not zero-sum: positive=%f negative=%f", positive, negative)
	}

	killerShare := positiveAmount(event.PositiveAlloc, 100)
	helperShare := positiveAmount(event.PositiveAlloc, 101)
	if helperShare <= killerShare {
		t.Fatalf("expected damage helper (%f) to exceed low-damage finisher (%f)", helperShare, killerShare)
	}
}

func TestAllocateKillEventExitFragIgnored(t *testing.T) {
	engine := probability.NewDefaultEngine()
	allocator := NewPoolAllocator(engine, DefaultConfig())

	state := probability.NewRoundState(5, 5, "de_dust2")
	kill := &KillEvent{
		KillerID:   100,
		VictimID:   200,
		KillerSide: common.TeamTerrorists,
		VictimSide: common.TeamCounterTerrorists,
	}

	event, result := allocator.AllocateKillEvent(state, KillAllocationInput{
		Kill:     kill,
		ExitFrag: true,
	})
	if result.RawSwing != 0 && len(event.PositiveAlloc) > 0 && event.PositiveAlloc[0].Amount != 0 {
		t.Fatalf("exit frag should produce zero swing")
	}
}

func TestAllocateKillEventPlantedClutchTransitions(t *testing.T) {
	tests := []struct {
		name            string
		tAlive, ctAlive int
		killerSide      common.Team
		victimSide      common.Team
	}{
		{"T turns 1v2 into 1v1", 1, 2, common.TeamTerrorists, common.TeamCounterTerrorists},
		{"CT turns 1v2 into 1v1", 2, 1, common.TeamCounterTerrorists, common.TeamTerrorists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allocator := NewPoolAllocator(probability.NewDefaultEngine(), DefaultConfig())
			state := probability.NewRoundState(tt.tAlive, tt.ctAlive, "de_mirage")
			state.SetBombPlanted()
			state.TimeRemaining = 40
			_, result := allocator.AllocateKillEvent(state, KillAllocationInput{
				Kill: &KillEvent{
					KillerID: 100, VictimID: 200,
					KillerSide: tt.killerSide, VictimSide: tt.victimSide,
					KillerEquip: 4000, VictimEquip: 4000,
					TotalDamageToVictim: 100, KillerDamageDealt: 100,
				},
			})
			if result.RawSwing <= 0 {
				t.Fatalf("an enemy kill reducing a planted 1v2 to 1v1 should earn positive swing, got %f", result.RawSwing)
			}
		})
	}
}

func TestAllocateKillEventHealthReductionRedistributes(t *testing.T) {
	engine := probability.NewDefaultEngine()
	cfg := DefaultConfig()
	allocator := NewPoolAllocator(engine, cfg)

	state := probability.NewRoundState(5, 5, "de_dust2")
	kill := &KillEvent{
		KillerID:    100,
		VictimID:    200,
		KillerSide:  common.TeamTerrorists,
		VictimSide:  common.TeamCounterTerrorists,
		KillerEquip: 5000,
		VictimEquip: 5000,
	}

	event, _ := allocator.AllocateKillEvent(state, KillAllocationInput{
		Kill:                kill,
		VictimPriorDamage:   90,
		LosingSideTeammates: []uint64{201, 202},
	})

	negative := sumAllocations(event.NegativeAlloc)
	if math.Abs(negative+event.NegativePool) > cfg.ZeroSumTolerance {
		t.Fatalf("negative pool not fully allocated: pool=%f allocated=%f", event.NegativePool, -negative)
	}

	victimPenalty := amountForPlayer(event.NegativeAlloc, 200)
	teammatePenalty := amountForPlayer(event.NegativeAlloc, 201) + amountForPlayer(event.NegativeAlloc, 202)
	if victimPenalty >= 0 {
		t.Fatalf("expected victim penalty, got %f", victimPenalty)
	}
	if teammatePenalty >= 0 {
		t.Fatalf("expected teammate redistribution, got %f", teammatePenalty)
	}
}

func TestAllocateTradeReallocation(t *testing.T) {
	engine := probability.NewDefaultEngine()
	allocator := NewPoolAllocator(engine, DefaultConfig())

	allocs := allocator.AllocateTradeReallocation(200, common.TeamCounterTerrorists, -0.08, []uint64{201, 202})
	total := sumAllocs(allocs)
	if math.Abs(total) > 0.005 {
		t.Fatalf("trade reallocation should be zero-sum, total=%f", total)
	}

	victimRefund := amountForPlayer(allocs, 200)
	if victimRefund <= 0 {
		t.Fatalf("expected victim refund, got %f", victimRefund)
	}
}

func sumAllocations(allocs []PlayerSwingAllocation) float64 {
	total := 0.0
	for _, a := range allocs {
		total += a.Amount
	}
	return total
}

func sumAllocs(allocs []PlayerSwingAllocation) float64 {
	return sumAllocations(allocs)
}

func positiveAmount(allocs []PlayerSwingAllocation, steamID uint64) float64 {
	total := 0.0
	for _, a := range allocs {
		if a.SteamID == steamID {
			total += a.Amount
		}
	}
	return total
}

func amountForPlayer(allocs []PlayerSwingAllocation, steamID uint64) float64 {
	total := 0.0
	for _, a := range allocs {
		if a.SteamID == steamID {
			total += a.Amount
		}
	}
	return total
}
