// =============================================================================
// DISCLAIMER: Comments in this file were generated with AI assistance to help
// users find and understand code for reference while building FraGG 3.0.
// =============================================================================

// Package parser provides CS2 demo file parsing functionality.
// This file contains event handlers that process game events (kills, deaths,
// bomb plants, round ends, etc.) and update player statistics accordingly.
package parser

import (
	"eco-rating/model"
	"eco-rating/rating"
	"math"

	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/msg"
)

// registerHandlers sets up all event handlers for demo parsing.
// This is the core of the parsing logic, handling:
// - Match/round lifecycle events (start, end, freeze time)
// - Kill events (including trade detection, opening kills, multi-kills)
// - Damage events
// - Bomb events (plant, defuse, explode)
// - Flash events (enemy blinds, team flashes)
// - Grenade throws
func (d *DemoParser) registerHandlers() {
	d.parser.RegisterNetMessageHandler(func(m *msg.CSVCMsg_ServerInfo) {
		d.state.MapName = m.GetMapName()
	})

	d.parser.RegisterEventHandler(func(e events.MatchStart) {
		d.state.MatchStarted = true
	})

	d.parser.RegisterEventHandler(func(e events.MatchStartedChanged) {
		if e.NewIsStarted {
			d.state.MatchStarted = true
		}
	})

	d.parser.RegisterEventHandler(func(e events.RoundStart) {
		d.state.Round = make(map[uint64]*model.RoundStats)
		d.state.RoundHasKill = false
		d.state.RecentKills = make(map[uint64]recentKill)
		d.state.RecentTeamDeaths = make(map[uint64]float64)
		d.state.PendingTrades = make(map[uint64][]pendingTrade)
		d.state.RoundDecided = false
		d.state.RoundDecidedAt = 0
	})

	d.parser.RegisterEventHandler(func(e events.BombPlanted) {
		if d.state.IsKnifeRound || !d.state.MatchStarted {
			return
		}

		planter := d.state.ensurePlayer(e.Player)
		roundStats := d.state.ensureRound(e.Player)
		roundStats.PlantedBomb = true

		d.logger.LogBombPlant(d.state.RoundNumber, planter.Name)
	})

	d.parser.RegisterEventHandler(func(e events.BombDefused) {
		if d.state.IsKnifeRound || !d.state.MatchStarted {
			return
		}

		defuser := d.state.ensurePlayer(e.Player)
		roundStats := d.state.ensureRound(e.Player)
		roundStats.DefusedBomb = true

		d.logger.LogBombDefuse(d.state.RoundNumber, defuser.Name)
	})

	d.parser.RegisterEventHandler(func(e events.PlayerFlashed) {
		if d.state.IsKnifeRound || !d.state.MatchStarted {
			return
		}

		if e.Attacker != nil && e.Player != nil {
			roundStats := d.state.ensureRound(e.Attacker)
			flashDuration := e.FlashDuration().Seconds()
			if e.Attacker.Team != e.Player.Team {
				roundStats.FlashAssists++
				roundStats.EnemyFlashDuration += flashDuration
			} else if e.Attacker.SteamID64 != e.Player.SteamID64 {
				roundStats.TeamFlashCount++
				roundStats.TeamFlashDuration += flashDuration
			}
		}
	})

	d.parser.RegisterEventHandler(func(e events.GrenadeProjectileThrow) {
		if d.state.IsKnifeRound || !d.state.MatchStarted {
			return
		}

		if e.Projectile != nil && e.Projectile.Thrower != nil {
			if e.Projectile.WeaponInstance != nil && e.Projectile.WeaponInstance.Type == common.EqFlash {
				roundStats := d.state.ensureRound(e.Projectile.Thrower)
				roundStats.FlashesThrown++
			}
		}
	})

	d.parser.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		gs := d.parser.GameState()
		if gs.IsWarmupPeriod() {
			return
		}
		participants := gs.Participants().Playing()
		if len(participants) > 0 {
			firstPlayer := participants[0]
			if firstPlayer.Money()+firstPlayer.MoneySpentThisRound() == 0 {
				d.state.IsKnifeRound = true
				d.logger.LogKnifeRound()
				return
			}
		}
		d.state.IsKnifeRound = false
		d.state.RoundNumber++

		isPistolRound := d.state.RoundNumber == 1 || d.state.RoundNumber == 13 ||
			(d.state.RoundNumber > 24 && (d.state.RoundNumber-25)%6 == 0)
		d.state.IsPistolRound = isPistolRound

		d.state.RoundStartTime = float64(d.parser.CurrentFrame()) / 64.0

		for _, p := range participants {
			if p.Team == common.TeamTerrorists {
				d.state.CurrentSide = "T"
				break
			} else if p.Team == common.TeamCounterTerrorists {
				d.state.CurrentSide = "CT"
				break
			}
		}

		d.logger.LogRoundStart(d.state.RoundNumber)

		for _, p := range participants {
			d.state.ensurePlayer(p)
			roundStats := d.state.ensureRound(p)
			roundStats.IsPistolRound = isPistolRound
			if p.Team == common.TeamTerrorists {
				roundStats.PlayerSide = "T"
			} else if p.Team == common.TeamCounterTerrorists {
				roundStats.PlayerSide = "CT"
			}
		}
	})

	d.parser.RegisterEventHandler(func(e events.Kill) {
		if d.parser.GameState().IsWarmupPeriod() || d.state.IsKnifeRound {
			return
		}

		a := e.Killer
		v := e.Victim

		if a != nil && v != nil && a.SteamID64 == v.SteamID64 {
			return
		}

		if a != nil && v != nil && a.Team == v.Team {
			return
		}

		currentTick := d.parser.CurrentFrame()
		const tradeWindow = 64 * 5

		currentTime := float64(currentTick) / 64.0
		timeInRound := currentTime - d.state.RoundStartTime

		if v != nil {
			victim := d.state.ensurePlayer(v)
			victim.Deaths++
			victimRound := d.state.ensureRound(v)
			victimRound.DeathTime = timeInRound

			d.state.RecentTeamDeaths[v.SteamID64] = timeInRound

			for _, weapon := range v.Weapons() {
				if weapon.Type == common.EqAWP {
					victimRound.HadAWP = true
					victimRound.LostAWP = true
					break
				}
			}

			if a != nil {
				gs := d.parser.GameState()
				victimPos := v.Position()
				for _, teammate := range gs.Participants().Playing() {
					if teammate.Team == v.Team && teammate.IsAlive() && teammate.SteamID64 != v.SteamID64 {
						teammatePos := teammate.Position()
						dx := victimPos.X - teammatePos.X
						dy := victimPos.Y - teammatePos.Y
						distance := math.Sqrt(dx*dx + dy*dy)

						if distance < 1200.0 {
							pt := pendingTrade{
								KillerID:           a.SteamID64,
								KillerTeam:         a.Team,
								TeammateID:         teammate.SteamID64,
								DeathTick:          currentTick,
								TeammatePos:        [3]float64{teammatePos.X, teammatePos.Y, teammatePos.Z},
								PotentialTraderPos: [3]float64{teammatePos.X, teammatePos.Y, teammatePos.Z},
							}
							d.state.PendingTrades[a.SteamID64] = append(d.state.PendingTrades[a.SteamID64], pt)
						}
					}
				}
			}
		}

		if a != nil && v != nil {
			if recent, ok := d.state.RecentKills[v.SteamID64]; ok {
				if recent.VictimTeam == a.Team && currentTick-recent.Tick <= tradeWindow {
					if tradedRound, exists := d.state.Round[recent.VictimID]; exists {
						tradedRound.Traded = true
						tradedRound.SavedByTeammate = true
					}
					tradedPlayerName := ""
					if tradedPlayer, exists := d.state.Players[recent.VictimID]; exists {
						tradedPlayer.TradedDeaths++
						tradedPlayerName = tradedPlayer.Name

						if tradedRound, exists := d.state.Round[recent.VictimID]; exists {
							if tradedRound.OpeningDeath {
								tradedPlayer.OpeningDeathsTraded++
							}
						}
					}
					attackerStats := d.state.ensurePlayer(a)
					attackerStats.TradeDenials++
					attackerStats.SavedTeammate++
					attackerRound := d.state.ensureRound(a)
					attackerRound.SavedTeammate = true

					d.logger.LogTrade(d.state.RoundNumber, a.Name, tradedPlayerName, v.Name)
				}
			}

			delete(d.state.PendingTrades, v.SteamID64)
		}

		for killerID, pendingList := range d.state.PendingTrades {
			var remainingPending []pendingTrade
			expiredCount := 0
			for _, pt := range pendingList {
				if currentTick-pt.DeathTick > tradeWindow {
					if roundStats, exists := d.state.Round[pt.TeammateID]; exists {
						roundStats.FailedTrades++
					}
					expiredCount++
				} else {
					remainingPending = append(remainingPending, pt)
				}
			}
			if expiredCount > 0 {
				if killerRound, exists := d.state.Round[killerID]; exists {
					killerRound.TradeDenials++
				}
			}
			if len(remainingPending) > 0 {
				d.state.PendingTrades[killerID] = remainingPending
			} else {
				delete(d.state.PendingTrades, killerID)
			}
		}

		if a == nil || v == nil {
			return
		}

		d.state.RecentKills[a.SteamID64] = recentKill{
			VictimID:   v.SteamID64,
			VictimTeam: v.Team,
			Tick:       currentTick,
		}

		attacker := d.state.ensurePlayer(a)
		victim := d.state.ensurePlayer(v)
		round := d.state.ensureRound(a)

		attackerEquip := a.EquipmentValueCurrent()
		victimEquip := v.EquipmentValueCurrent()

		killValue := rating.EcoKillValue(float64(attackerEquip), float64(victimEquip))
		deathPenalty := rating.EcoDeathPenalty(float64(victimEquip), float64(attackerEquip))

		d.logger.LogKill(d.state.RoundNumber, a.Name, v.Name, attackerEquip, victimEquip, killValue)
		d.logger.LogDeath(d.state.RoundNumber, v.Name, a.Name, victimEquip, attackerEquip, deathPenalty)

		round.KillTimes = append(round.KillTimes, timeInRound)

		if d.state.RoundDecided {
			round.IsExitFrag = true
			round.ExitFrags++
		}

		round.Kills++
		round.GotKill = true
		round.EconImpact += killValue
		attacker.Kills++
		attacker.EcoKillValue += killValue
		attacker.RoundImpact += killValue
		attacker.EconImpact += killValue

		if killValue < 1.0 {
			attacker.LowBuyKills++
		}
		if killValue <= 0.85 {
			attacker.DisadvantagedBuyKills++
		}

		if e.Weapon != nil {
			switch e.Weapon.Type {
			case common.EqAWP:
				round.AWPKills++
				round.AWPKill = true
				attacker.AWPKills++
			case common.EqKnife:
				round.KnifeKill = true
			case common.EqHE, common.EqMolotov, common.EqIncendiary:
				round.UtilityKills++
				attacker.UtilityKills++
			}

			isPistol := e.Weapon.Type >= common.EqP2000 && e.Weapon.Type <= common.EqRevolver
			victimHadRifle := victimEquip > 3500
			if isPistol && victimHadRifle {
				round.PistolVsRifleKill = true
			}
		}

		victim.EcoDeathValue += deathPenalty

		victimRound := d.state.ensureRound(v)
		if !d.state.RoundHasKill {
			attacker.OpeningKills++
			attacker.OpeningAttempts++
			attacker.OpeningSuccesses++
			round.OpeningKill = true
			round.EntryFragger = true
			round.InvolvedInOpening = true

			if e.Weapon != nil && e.Weapon.Type == common.EqAWP {
				round.AWPOpeningKill = true
				attacker.AWPOpeningKills++
			}

			victim.OpeningDeaths++
			victim.OpeningAttempts++
			victimRound.OpeningDeath = true
			victimRound.InvolvedInOpening = true

			d.state.RoundHasKill = true
			d.logger.LogOpeningKill(d.state.RoundNumber, a.Name, v.Name)
		}

		if recent, ok := d.state.RecentKills[v.SteamID64]; ok {
			if recent.VictimTeam == a.Team && currentTick-recent.Tick <= tradeWindow {
				round.TradeKill = true
				if deathTime, exists := d.state.RecentTeamDeaths[recent.VictimID]; exists {
					round.TradeSpeed = timeInRound - deathTime
				}
			}
		}

		equipRatio := float64(victimEquip) / math.Max(float64(attackerEquip), 500.0)
		if equipRatio > 2.0 {
			round.EcoKill = true
		}
		if equipRatio < 0.5 {
			victimRound.AntiEcoKill = true
		}
		if e.IsHeadshot {
			attacker.PerfectKills++
		}

		if e.Assister != nil {
			assister := d.state.ensurePlayer(e.Assister)
			assister.Assists++
			assistRound := d.state.ensureRound(e.Assister)
			assistRound.GotAssist = true
		}
	})

	d.parser.RegisterEventHandler(func(e events.PlayerHurt) {
		if d.parser.GameState().IsWarmupPeriod() || d.state.IsKnifeRound {
			return
		}

		if e.Attacker == nil || e.Player == nil {
			return
		}

		if e.Attacker.Team != e.Player.Team {
			ps := d.state.ensurePlayer(e.Attacker)
			ps.Damage += int(e.HealthDamageTaken)

			roundStats := d.state.ensureRound(e.Attacker)
			roundStats.Damage += int(e.HealthDamageTaken)

			if e.Weapon != nil {
				switch e.Weapon.Type {
				case common.EqHE, common.EqMolotov, common.EqIncendiary:
					roundStats.UtilityDamage += int(e.HealthDamageTaken)
				}
			}
		}
	})

	d.parser.RegisterEventHandler(func(e events.BombExplode) {
		if d.state.IsKnifeRound || !d.state.MatchStarted {
			return
		}
		currentTime := float64(d.parser.CurrentFrame()) / 64.0
		timeInRound := currentTime - d.state.RoundStartTime
		d.state.RoundDecided = true
		d.state.RoundDecidedAt = timeInRound
	})

	d.parser.RegisterEventHandler(func(e events.Kill) {
		if d.state.IsKnifeRound || !d.state.MatchStarted || d.state.RoundDecided {
			return
		}

		gs := d.parser.GameState()
		tAlive := 0
		ctAlive := 0
		for _, p := range gs.Participants().Playing() {
			if p.IsAlive() {
				if p.Team == common.TeamTerrorists {
					tAlive++
				} else if p.Team == common.TeamCounterTerrorists {
					ctAlive++
				}
			}
		}

		if tAlive == 0 || ctAlive == 0 {
			currentTime := float64(d.parser.CurrentFrame()) / 64.0
			timeInRound := currentTime - d.state.RoundStartTime
			d.state.RoundDecided = true
			d.state.RoundDecidedAt = timeInRound
		}
	})

	d.parser.RegisterEventHandler(func(e events.RoundEnd) {
		if d.parser.GameState().IsWarmupPeriod() || d.state.IsKnifeRound {
			return
		}

		gs := d.parser.GameState()
		winnerTeam := e.Winner

		currentTick := d.parser.CurrentFrame()
		const tradeWindow = 64 * 5

		for _, pendingList := range d.state.PendingTrades {
			for _, pt := range pendingList {
				if currentTick-pt.DeathTick > tradeWindow {
					if roundStats, exists := d.state.Round[pt.TeammateID]; exists {
						roundStats.FailedTrades++
					}
				}
			}
		}

		for steamID, roundStats := range d.state.Round {
			player := d.state.Players[steamID]
			if player == nil {
				continue
			}

			if roundStats.Kills >= 2 && roundStats.Kills <= 5 {
				player.MultiKillsRaw[roundStats.Kills]++
				d.logger.LogMultiKill(d.state.RoundNumber, player.Name, roundStats.Kills)
			}

			if player.RoundsPlayed > 0 {
				player.AWPKillsPerRound = float64(player.AWPKills) / float64(player.RoundsPlayed)
			}
		}

		roundEndTime := float64(d.parser.CurrentFrame()) / 64.0
		roundDuration := roundEndTime - d.state.RoundStartTime

		for _, p := range gs.Participants().Playing() {
			ps := d.state.ensurePlayer(p)
			round := d.state.ensureRound(p)

			teamWon := p.Team == winnerTeam
			round.TeamWon = teamWon
			if teamWon {
				ps.RoundsWon++
			} else {
				ps.RoundsLost++
			}

			if p.IsAlive() {
				ps.Survival++
				round.Survived = true
				round.TimeAlive = roundDuration
				ps.TotalTimeAlive += roundDuration

				if !teamWon {
					ps.SavesOnLoss++
				}
			} else if round.DeathTime > 0 {
				round.TimeAlive = round.DeathTime
				ps.TotalTimeAlive += round.DeathTime
			}
		}

		for _, p := range gs.Participants().Playing() {
			round := d.state.ensureRound(p)
			ps := d.state.ensurePlayer(p)

			aliveTeammates := 0
			aliveEnemies := 0

			for _, other := range gs.Participants().Playing() {
				if other.IsAlive() {
					if other.Team == p.Team {
						aliveTeammates++
					} else {
						aliveEnemies++
					}
				}
			}

			if p.IsAlive() && aliveTeammates == 1 {
				ps.LastAliveRounds++
				round.WasLastAlive = true

				if aliveEnemies > 0 {
					round.ClutchAttempt = true
					round.ClutchSize = aliveEnemies
					round.ClutchKills = round.Kills
					ps.ClutchRounds++

					if aliveEnemies == 1 {
						ps.Clutch1v1Attempts++
						if round.TeamWon {
							ps.Clutch1v1Wins++
						}
					}

					if round.TeamWon {
						round.ClutchWon = true
						ps.ClutchWins++
					}
				}
			}

			if p.IsAlive() && !round.TeamWon {
				round.SavedWeapons = true
			}
		}

		currentTime := float64(d.parser.CurrentFrame()) / 64.0
		timeRemaining := math.Max(0.0, 115.0-(currentTime-d.state.RoundStartTime))

		teamScore := d.state.TeamScore
		enemyScore := d.state.EnemyScore
		scoreDiff := teamScore - enemyScore
		isMatchPoint := teamScore == 12 || enemyScore == 12 || (d.state.RoundNumber > 30 && (teamScore >= 15 || enemyScore >= 15))
		isCloseGame := math.Abs(float64(scoreDiff)) <= 3

		roundImportance := 1.0
		if isMatchPoint {
			roundImportance = 1.3
		} else if isCloseGame {
			roundImportance = 1.15
		} else if math.Abs(float64(scoreDiff)) >= 8 {
			roundImportance = 0.85
		}

		roundContext := &model.RoundContext{
			RoundNumber:     d.state.RoundNumber,
			TotalPlayers:    10,
			BombPlanted:     false,
			BombDefused:     false,
			RoundType:       determineRoundType(d.state.RoundNumber),
			TimeRemaining:   timeRemaining,
			IsOvertimeRound: d.state.RoundNumber > 30,
			MapSide:         d.state.CurrentSide,
			TeamScore:       teamScore,
			EnemyScore:      enemyScore,
			ScoreDiff:       scoreDiff,
			IsMatchPoint:    isMatchPoint,
			IsCloseGame:     isCloseGame,
			RoundImportance: roundImportance,
			RoundDecided:    d.state.RoundDecided,
			RoundDecidedAt:  d.state.RoundDecidedAt,
		}

		for _, roundStats := range d.state.Round {
			if roundStats.PlantedBomb {
				roundContext.BombPlanted = true
			}
			if roundStats.DefusedBomb {
				roundContext.BombDefused = true
			}
		}

		for steamID, roundStats := range d.state.Round {
			player := d.state.Players[steamID]
			if player == nil {
				continue
			}

			roundStats.MultiKillRound = roundStats.Kills

			var playerEquipValue float64
			var teamEquipValue float64

			for _, p := range gs.Participants().Playing() {
				if p.SteamID64 == steamID {
					playerEquipValue = float64(p.EquipmentValueCurrent())

					teamTotal := 0
					teamCount := 0
					for _, teammate := range gs.Participants().Playing() {
						if teammate.Team == p.Team {
							teamTotal += teammate.EquipmentValueCurrent()
							teamCount++
						}
					}
					if teamCount > 0 {
						teamEquipValue = float64(teamTotal)
					}
					break
				}
			}

			if playerEquipValue == 0 {
				playerEquipValue = 3000.0
			}
			if teamEquipValue == 0 {
				teamEquipValue = 15000.0
			}

			swing := CalculateAdvancedRoundSwing(roundStats, roundContext, playerEquipValue, teamEquipValue)
			player.RoundSwing += swing

			if roundStats.PlayerSide == "T" {
				player.TRoundSwing += swing
			} else if roundStats.PlayerSide == "CT" {
				player.CTRoundSwing += swing
			}
		}

		for steamID, roundStats := range d.state.Round {
			player := d.state.Players[steamID]
			if player == nil {
				continue
			}
			if roundStats.GotKill || roundStats.GotAssist || roundStats.Survived || roundStats.Traded {
				player.KAST++
			}

			if roundStats.GotKill {
				player.RoundsWithKill++
				player.AttackRounds++
			}

			if roundStats.Kills >= 2 {
				player.RoundsWithMultiKill++
			}

			if roundStats.TeamWon {
				player.KillsInWonRounds += roundStats.Kills
				player.DamageInWonRounds += roundStats.Damage

				if roundStats.OpeningKill {
					player.RoundsWonAfterOpening++
				}
			}

			if roundStats.AWPKill {
				player.RoundsWithAWPKill++
			}
			if roundStats.AWPKills >= 2 {
				player.AWPMultiKillRounds++
			}

			if roundStats.GotAssist || roundStats.FlashAssists > 0 {
				player.SupportRounds++
				roundStats.IsSupportRound = true
			}

			if roundStats.GotAssist {
				player.AssistedKills += roundStats.Assists
			}

			player.UtilityDamage += roundStats.UtilityDamage
			player.FlashesThrown += roundStats.FlashesThrown
			player.FlashAssists += roundStats.FlashAssists
			player.EnemyFlashDuration += roundStats.EnemyFlashDuration
			player.TeamFlashCount += roundStats.TeamFlashCount
			player.TeamFlashDuration += roundStats.TeamFlashDuration
			player.ExitFrags += roundStats.ExitFrags

			if roundStats.SavedByTeammate {
				player.SavedByTeammate++
			}

			if roundStats.LostAWP {
				player.AWPDeaths++
				if !roundStats.AWPKill {
					player.AWPDeathsNoKill++
				}
			}

			if roundStats.KnifeKill {
				player.KnifeKills++
			}

			if roundStats.PistolVsRifleKill {
				player.PistolVsRifleKills++
			}

			if roundStats.TradeKill {
				player.TradeKills++
				if roundStats.TradeSpeed > 0 && roundStats.TradeSpeed < 2.0 {
					player.FastTrades++
				}
			}

			if roundStats.DeathTime > 0 && roundStats.DeathTime < 30.0 {
				player.EarlyDeaths++
			}

			if roundStats.IsPistolRound {
				player.PistolRoundsPlayed++
				player.PistolRoundKills += roundStats.Kills
				player.PistolRoundDamage += roundStats.Damage
				if roundStats.DeathTime > 0 {
					player.PistolRoundDeaths++
				} else if roundStats.Survived {
					player.PistolRoundSurvivals++
				}
				if roundStats.TeamWon {
					player.PistolRoundsWon++
				}
				if roundStats.Kills >= 2 {
					player.PistolRoundMultiKills++
				}
			}

			if roundStats.PlayerSide == "T" {
				player.TRoundsPlayed++
				player.TKills += roundStats.Kills
				player.TDamage += roundStats.Damage
				player.TEcoKillValue += roundStats.EconImpact
				if roundStats.Survived {
					player.TSurvivals++
				}
				if roundStats.DeathTime > 0 {
					player.TDeaths++
				}
				if roundStats.Kills >= 2 {
					player.TRoundsWithMultiKill++
				}
				if roundStats.Kills >= 0 && roundStats.Kills <= 5 {
					player.TMultiKills[roundStats.Kills]++
				}
				if roundStats.GotKill || roundStats.GotAssist || roundStats.Survived || roundStats.Traded {
					player.TKAST++
				}
				if roundStats.ClutchAttempt {
					player.TClutchRounds++
					if roundStats.ClutchWon {
						player.TClutchWins++
					}
				}
			} else if roundStats.PlayerSide == "CT" {
				player.CTRoundsPlayed++
				player.CTKills += roundStats.Kills
				player.CTDamage += roundStats.Damage
				player.CTEcoKillValue += roundStats.EconImpact
				if roundStats.Survived {
					player.CTSurvivals++
				}
				if roundStats.DeathTime > 0 {
					player.CTDeaths++
				}
				if roundStats.Kills >= 2 {
					player.CTRoundsWithMultiKill++
				}
				if roundStats.Kills >= 0 && roundStats.Kills <= 5 {
					player.CTMultiKills[roundStats.Kills]++
				}
				if roundStats.GotKill || roundStats.GotAssist || roundStats.Survived || roundStats.Traded {
					player.CTKAST++
				}
				if roundStats.ClutchAttempt {
					player.CTClutchRounds++
					if roundStats.ClutchWon {
						player.CTClutchWins++
					}
				}
			}
		}

		for _, p := range d.state.Players {
			p.RoundsPlayed++
		}

		if winnerTeam == common.TeamTerrorists {
			if d.state.CurrentSide == "T" {
				d.state.TeamScore++
			} else {
				d.state.EnemyScore++
			}
		} else if winnerTeam == common.TeamCounterTerrorists {
			if d.state.CurrentSide == "CT" {
				d.state.TeamScore++
			} else {
				d.state.EnemyScore++
			}
		}

		d.logger.LogRoundEnd(d.state.RoundNumber)
	})
}
