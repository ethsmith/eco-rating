# Eco-Rating Parser

> **DISCLAIMER**: Comments throughout this codebase were generated with AI assistance to help users find and understand code for reference while building FraGG 3.0. There may be mistakes in the comments. Please verify accuracy.

A CS2 demo parser that calculates advanced player performance ratings based on probability-based impact metrics, economic context, and round swing analysis.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Adding New Stats](#adding-new-stats)
- [Updating Weights](#updating-weights)
- [Rating Formula](#rating-formula)
- [Key Concepts](#key-concepts)

---

## Overview

This parser processes CS2 demo files and computes comprehensive player statistics including:

- **Probability Swing**: How much each action affected win probability
- **Economic Impact**: Equipment-adjusted kill values
- **HLTV Rating**: Standard HLTV 2.0 rating for comparison
- **Round Swing**: Per-round impact score
- **140+ tracked statistics**: Opening kills, trades, clutches, utility, AWP stats, etc.

### Usage

```bash
# Single demo
eco-rating -demo=path/to/demo.dem

# Cumulative mode (batch process from cloud bucket)
eco-rating -cumulative -tier=contender
```

---

## Architecture

```
eco-rating/
├── main.go                 # Entry point, CLI handling
├── config/                 # Configuration loading
├── bucket/                 # Cloud storage client
├── downloader/             # Demo download & extraction
├── parser/                 # Demo parsing (core logic)
│   ├── parser.go           # Main DemoParser struct
│   ├── handlers.go         # Event handlers (kills, damage, rounds)
│   ├── round.go            # MatchState management
│   ├── round_swing.go      # Round swing calculation
│   ├── side_stats.go       # T/CT side stat updates
│   ├── trade_detector.go   # Trade kill detection
│   ├── swing_tracker.go    # Probability swing tracking
│   └── damage_tracker.go   # Damage attribution
├── model/                  # Data structures
│   ├── player_stats.go     # PlayerStats struct (all tracked stats)
│   ├── round_stats.go      # RoundStats struct (per-round data)
│   └── round_context_builder.go
├── rating/                 # Rating calculations
│   ├── rating.go           # Final rating computation
│   ├── weights.go          # ALL constants and weights
│   ├── economy.go          # Economic kill/death values
│   ├── hltv.go             # HLTV 2.0 rating calculation
│   ├── probability/        # Win probability engine
│   └── swing/              # Swing calculation & attribution
├── output/                 # Statistics aggregation
│   └── aggregator.go       # Multi-game stat aggregation
└── export/                 # Export to CSV/JSON
```

---

## Adding New Stats

### Step 1: Add the Field to PlayerStats

Edit `model/player_stats.go` to add your new stat:

```go
type PlayerStats struct {
    // ... existing fields ...
    
    // Your new stat
    MyNewStat     int     `json:"my_new_stat"`
    MyNewStatPct  float64 `json:"my_new_stat_pct"`  // If it needs a percentage
}
```

### Step 2: Add to RoundStats (if tracked per-round)

If your stat is tracked per-round, add it to `model/round_stats.go`:

```go
type RoundStats struct {
    // ... existing fields ...
    
    MyNewStatThisRound int
}
```

### Step 3: Track the Stat in Event Handlers

Edit `parser/handlers.go` to track your stat during parsing. Find the appropriate handler:

- **Kill events**: `handleKill()` or create a new `processMyNewStat()` function
- **Damage events**: `handlePlayerHurt()`
- **Round events**: `handleRoundEnd()`
- **Bomb events**: `handleBombPlanted()`, `handleBombDefused()`

Example - tracking a new kill-related stat:

```go
// In handlers.go, add to processKillerStats or create new function
func (d *DemoParser) processMyNewStat(ctx *killContext) {
    if someCondition {
        attacker := d.state.ensurePlayer(ctx.attacker)
        round := d.state.ensureRound(ctx.attacker)
        
        attacker.MyNewStat++
        round.MyNewStatThisRound++
    }
}

// Call it from handleKill()
func (d *DemoParser) handleKill(e events.Kill) {
    // ... existing code ...
    d.processMyNewStat(ctx)
}
```

### Step 4: Calculate Derived Metrics

If your stat needs a per-round rate or percentage, add it to `parser/parser.go` in `computeDerivedStats()`:

```go
func (d *DemoParser) computeDerivedStats() {
    for _, p := range d.state.Players {
        if p.RoundsPlayed > 0 {
            rounds := float64(p.RoundsPlayed)
            // ... existing calculations ...
            
            // Your new derived metric
            p.MyNewStatPct = float64(p.MyNewStat) / rounds
        }
    }
}
```

### Step 5: Add to Aggregator (for cumulative mode)

Edit `output/aggregator.go`:

1. Add field to `AggregatedStats` struct
2. Add accumulation in `AddGame()`:
   ```go
   agg.MyNewStat += p.MyNewStat
   ```
3. Add derived calculation in `Finalize()` if needed

### Step 6: Add to Export

Edit `export/file.go`:

1. Add column to `getSingleGameHeader()`:
   ```go
   return []string{
       // ... existing headers ...
       "My New Stat", "My New Stat Pct",
   }
   ```

2. Add value to `getSingleGameRow()`:
   ```go
   return []string{
       // ... existing values ...
       strconv.Itoa(p.MyNewStat),
       formatFloat(p.MyNewStatPct),
   }
   ```

3. Repeat for `getAggregatedHeader()` and `getAggregatedRow()` if used in cumulative mode.

---

## Updating Weights

All weights and constants are centralized in `rating/weights.go`. This makes tuning the rating formula straightforward.

### Weight Categories

#### 1. Rating Formula Weights

These control how stats contribute to the final rating:

```go
// rating/weights.go

// ADR contribution (per point above/below baseline)
ADRContribAbove = 0.005  // Bonus per ADR point above 77
ADRContribBelow = 0.004  // Penalty per ADR point below 77

// KAST contribution
KASTContribAbove = 0.20  // Bonus per % above 72%
KASTContribBelow = 0.25  // Penalty per % below 72%

// Probability swing multiplier (core metric)
ProbSwingContribMultiplier = 2.5  // How much swing affects rating
```

**To adjust**: Change these values to shift importance between stats.

#### 2. Economic Kill/Death Values

These control how equipment advantage affects kill value:

```go
// Kills against better-equipped opponents worth more
EcoKillPistolVsRifle = 1.80  // Pistol kills rifle = 1.8x value
EcoKillEcoVsForce    = 1.50  // Eco kills force = 1.5x value
EcoKillEqual         = 1.00  // Equal equipment = 1.0x

// Deaths to worse-equipped opponents penalized more  
EcoDeathToPistol     = 1.60  // Rifle dies to pistol = 1.6x penalty
EcoDeathEqual        = 1.00  // Equal equipment = 1.0x
```

**To adjust**: Modify multipliers in `EcoKillValue()` and `EcoDeathPenalty()` in `rating/economy.go`.

#### 3. Round Swing Weights

These control per-round impact scoring:

```go
// Performance contribution
KillContribPerKill    = 0.04   // Base swing per kill
KillContribMultiBonus = 0.02   // Extra per kill after first
DamageContribMax      = 0.08   // Max damage contribution
SurvivalContribWin    = 0.02   // Surviving a won round
SurvivalContribLoss   = 0.04   // Surviving a lost round (save)

// Situational bonuses
OpeningKillBonus      = 0.06   // First kill of round
EntryFragBonus        = 0.04   // Entry fragging
TradeKillBonus        = 0.02   // Trading a teammate
BombPlantBonus        = 0.08   // Planting bomb
BombDefuseBonus       = 0.10   // Defusing bomb

// Multi-kill bonuses
MultiKill2KBonus = 0.03  // Double kill
MultiKill3KBonus = 0.08  // Triple kill
MultiKill4KBonus = 0.15  // Quad kill
MultiKill5KBonus = 0.25  // Ace

// Penalties
DeathPenaltyUntraded  = 0.08   // Dying without trade
EarlyDeathPenalty     = 0.08   // Dying in first 15s
```

**To adjust**: Change values in `rating/weights.go`, then the `RoundSwingCalculator` in `parser/round_swing.go` uses them.

#### 4. HLTV Rating Constants

Standard HLTV 2.0 formula constants (usually don't change):

```go
HLTVBaselineKPR    = 0.679  // Average KPR in pro matches
HLTVBaselineSPR    = 0.317  // Average survival rate
HLTVBaselineRMK    = 1.277  // Average multi-kill points
HLTVSurvivalWeight = 0.7    // Survival component weight
HLTVRatingDivisor  = 2.7    // Final divisor
```

### How to Tune Weights

1. **Identify the behavior to change**
   - Rating too high for fraggers? Reduce `KillContribPerKill`
   - Opening kills undervalued? Increase `OpeningKillBonus`
   - Deaths not penalized enough? Increase `DeathPenaltyUntraded`

2. **Make small changes** (5-10% at a time)

3. **Test on sample demos**
   ```bash
   eco-rating -demo=test.dem -output=test.csv
   ```

4. **Compare before/after** ratings for known good/bad performances

---

## Rating Formula

The final rating is computed in `rating/rating.go`:

```go
rating = 1.0                          // Baseline
       + adrContrib                   // ADR above/below 77
       + kastContrib                  // KAST above/below 72%
       + probSwingContrib             // Probability swing (core)
```

### Probability Swing (Core Metric)

This measures how much a player's actions affected their team's win probability:

1. **Before each action**: Calculate team's win probability (e.g., 45%)
2. **After action**: Calculate new win probability (e.g., 55%)
3. **Swing**: The delta (+10% in this case)

Actions that affect swing:
- Kills (biggest impact)
- Deaths (negative swing)
- Bomb plants
- Bomb defuses

Swing is adjusted for:
- **Economy**: Hard kills (pistol vs rifle) worth more
- **Trade status**: Trade kills worth less (expected)
- **Timing**: Exit frags worth less

---

## Key Concepts

### KAST
**K**ill, **A**ssist, **S**urvive, or **T**raded. Percentage of rounds where player contributed.

### Trade
A kill that avenges a teammate's death within 5 seconds.

### Round Swing
Per-round impact score combining kills, damage, utility, and survival.

### Probability Swing  
Win probability delta from player actions. A kill that moves win probability from 30% to 50% = +20% swing.

### Economic Impact
Kill value adjusted for equipment advantage. Killing a rifle player with a pistol is worth 1.8x; killing a pistol player with a rifle is worth 0.7x.

---

## Files Quick Reference

| File | Purpose |
|------|---------|
| `model/player_stats.go` | Add new stat fields |
| `model/round_stats.go` | Add per-round tracking fields |
| `parser/handlers.go` | Track stats during parsing |
| `parser/parser.go` | Calculate derived metrics |
| `output/aggregator.go` | Accumulate stats across games |
| `export/file.go` | Add to CSV export |
| `rating/weights.go` | **All weights and constants** |
| `rating/rating.go` | Final rating formula |
| `rating/economy.go` | Economic kill/death values |

---

## Questions?

Review the inline comments in each file. Comments were generated with AI assistance to help explain the code, though there may be mistakes.
