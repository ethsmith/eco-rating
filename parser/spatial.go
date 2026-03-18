// Package parser provides CS2 demo file parsing functionality.
// This file implements spatial analysis utilities for ML pipeline features
// including crossfire detection, smoke effectiveness, zone definitions,
// and aim vector tracking.
package parser

import (
	"math"

	"github.com/ethsmith/eco-rating/model"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
)

// MapZones defines spatial zones for each competitive map.
// Zones are used for uncontested advance calculation and smoke effectiveness.
type MapZones struct {
	Name           string
	TSpawn         model.Position
	CTSpawn        model.Position
	BombsiteA      ZoneRect
	BombsiteB      ZoneRect
	MidZone        ZoneRect
	ContestedZones []ZoneRect
	SmokeSpots     []SmokeSpot
}

// ZoneRect defines a rectangular zone on the map.
type ZoneRect struct {
	Name string
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
	MinZ float64
	MaxZ float64
}

// SmokeSpot defines an optimal smoke position with effectiveness radius.
type SmokeSpot struct {
	Name     string
	Position model.Position
	Radius   float64 // Effective radius for smoke placement
	Side     string  // "T", "CT", or "both"
}

// Contains checks if a position is within the zone.
func (z ZoneRect) Contains(pos model.Position) bool {
	return pos.X >= z.MinX && pos.X <= z.MaxX &&
		pos.Y >= z.MinY && pos.Y <= z.MaxY &&
		(z.MinZ == 0 && z.MaxZ == 0 || pos.Z >= z.MinZ && pos.Z <= z.MaxZ)
}

// Center returns the center position of the zone.
func (z ZoneRect) Center() model.Position {
	return model.Position{
		X: (z.MinX + z.MaxX) / 2,
		Y: (z.MinY + z.MaxY) / 2,
		Z: (z.MinZ + z.MaxZ) / 2,
	}
}

// mapZonesData contains zone definitions for competitive maps.
var mapZonesData = map[string]*MapZones{
	"de_dust2": {
		Name:    "de_dust2",
		TSpawn:  model.Position{X: -672, Y: 2560, Z: 0},
		CTSpawn: model.Position{X: 528, Y: -1568, Z: 0},
		BombsiteA: ZoneRect{
			Name: "A Site",
			MinX: 880, MaxX: 1520,
			MinY: 2080, MaxY: 2720,
		},
		BombsiteB: ZoneRect{
			Name: "B Site",
			MinX: -1920, MaxX: -1200,
			MinY: 2240, MaxY: 2880,
		},
		MidZone: ZoneRect{
			Name: "Mid",
			MinX: -608, MaxX: 176,
			MinY: 640, MaxY: 1600,
		},
		ContestedZones: []ZoneRect{
			{Name: "Long A", MinX: 1040, MaxX: 1680, MinY: 400, MaxY: 1600},
			{Name: "Short A", MinX: 240, MaxX: 880, MinY: 1600, MaxY: 2400},
			{Name: "B Tunnels", MinX: -1600, MaxX: -800, MinY: 1200, MaxY: 2000},
			{Name: "Mid Doors", MinX: -400, MaxX: 200, MinY: 800, MaxY: 1400},
		},
		SmokeSpots: []SmokeSpot{
			{Name: "Xbox", Position: model.Position{X: -400, Y: 1200, Z: 0}, Radius: 100, Side: "T"},
			{Name: "CT Spawn", Position: model.Position{X: 528, Y: -1200, Z: 0}, Radius: 120, Side: "T"},
			{Name: "Long Corner", Position: model.Position{X: 1400, Y: 600, Z: 0}, Radius: 100, Side: "T"},
			{Name: "B Window", Position: model.Position{X: -1400, Y: 2500, Z: 0}, Radius: 80, Side: "T"},
		},
	},
	"de_mirage": {
		Name:    "de_mirage",
		TSpawn:  model.Position{X: -2976, Y: 816, Z: 0},
		CTSpawn: model.Position{X: -400, Y: -2400, Z: 0},
		BombsiteA: ZoneRect{
			Name: "A Site",
			MinX: -1200, MaxX: -400,
			MinY: -2000, MaxY: -1200,
		},
		BombsiteB: ZoneRect{
			Name: "B Site",
			MinX: -2400, MaxX: -1600,
			MinY: -2800, MaxY: -2000,
		},
		MidZone: ZoneRect{
			Name: "Mid",
			MinX: -1600, MaxX: -800,
			MinY: -800, MaxY: 400,
		},
		ContestedZones: []ZoneRect{
			{Name: "A Ramp", MinX: -1600, MaxX: -800, MinY: -1600, MaxY: -800},
			{Name: "Palace", MinX: -400, MaxX: 400, MinY: -1200, MaxY: -400},
			{Name: "B Apps", MinX: -2800, MaxX: -2000, MinY: -1600, MaxY: -800},
			{Name: "Underpass", MinX: -1200, MaxX: -400, MinY: -400, MaxY: 400},
		},
		SmokeSpots: []SmokeSpot{
			{Name: "Window", Position: model.Position{X: -1200, Y: -400, Z: 0}, Radius: 100, Side: "T"},
			{Name: "Connector", Position: model.Position{X: -1000, Y: -1000, Z: 0}, Radius: 100, Side: "T"},
			{Name: "Jungle", Position: model.Position{X: -800, Y: -1400, Z: 0}, Radius: 100, Side: "T"},
			{Name: "CT", Position: model.Position{X: -600, Y: -1800, Z: 0}, Radius: 100, Side: "T"},
		},
	},
	"de_inferno": {
		Name:    "de_inferno",
		TSpawn:  model.Position{X: -1984, Y: 576, Z: 0},
		CTSpawn: model.Position{X: 2176, Y: 576, Z: 0},
		BombsiteA: ZoneRect{
			Name: "A Site",
			MinX: 1600, MaxX: 2400,
			MinY: 400, MaxY: 1200,
		},
		BombsiteB: ZoneRect{
			Name: "B Site",
			MinX: 160, MaxX: 960,
			MinY: 2400, MaxY: 3200,
		},
		MidZone: ZoneRect{
			Name: "Mid",
			MinX: 400, MaxX: 1200,
			MinY: 0, MaxY: 800,
		},
		ContestedZones: []ZoneRect{
			{Name: "Apartments", MinX: 800, MaxX: 1600, MinY: 1600, MaxY: 2400},
			{Name: "Banana", MinX: -400, MaxX: 400, MinY: 1600, MaxY: 2800},
			{Name: "Arch", MinX: 1200, MaxX: 2000, MinY: 800, MaxY: 1600},
		},
		SmokeSpots: []SmokeSpot{
			{Name: "Top Banana", Position: model.Position{X: 0, Y: 2200, Z: 0}, Radius: 100, Side: "CT"},
			{Name: "CT Spawn", Position: model.Position{X: 2000, Y: 600, Z: 0}, Radius: 120, Side: "T"},
			{Name: "Coffins", Position: model.Position{X: 2000, Y: 800, Z: 0}, Radius: 80, Side: "T"},
		},
	},
	"de_nuke": {
		Name:    "de_nuke",
		TSpawn:  model.Position{X: -1024, Y: -1024, Z: -415},
		CTSpawn: model.Position{X: 640, Y: -768, Z: -415},
		BombsiteA: ZoneRect{
			Name: "A Site",
			MinX: 400, MaxX: 1200,
			MinY: -1600, MaxY: -800,
			MinZ: -500, MaxZ: -300,
		},
		BombsiteB: ZoneRect{
			Name: "B Site",
			MinX: 400, MaxX: 1200,
			MinY: -1600, MaxY: -800,
			MinZ: -800, MaxZ: -600,
		},
		MidZone: ZoneRect{
			Name: "Outside",
			MinX: -800, MaxX: 400,
			MinY: -2400, MaxY: -1600,
		},
		ContestedZones: []ZoneRect{
			{Name: "Lobby", MinX: 0, MaxX: 800, MinY: -800, MaxY: 0},
			{Name: "Ramp", MinX: 800, MaxX: 1600, MinY: -2400, MaxY: -1600},
			{Name: "Secret", MinX: -400, MaxX: 400, MinY: -2800, MaxY: -2000},
		},
		SmokeSpots: []SmokeSpot{
			{Name: "Garage", Position: model.Position{X: -200, Y: -2000, Z: -415}, Radius: 100, Side: "T"},
			{Name: "Heaven", Position: model.Position{X: 800, Y: -1200, Z: -350}, Radius: 80, Side: "T"},
		},
	},
	"de_ancient": {
		Name:    "de_ancient",
		TSpawn:  model.Position{X: -2400, Y: -400, Z: 0},
		CTSpawn: model.Position{X: 1200, Y: -400, Z: 0},
		BombsiteA: ZoneRect{
			Name: "A Site",
			MinX: 800, MaxX: 1600,
			MinY: -1200, MaxY: -400,
		},
		BombsiteB: ZoneRect{
			Name: "B Site",
			MinX: -400, MaxX: 400,
			MinY: 800, MaxY: 1600,
		},
		MidZone: ZoneRect{
			Name: "Mid",
			MinX: -800, MaxX: 0,
			MinY: -800, MaxY: 0,
		},
		ContestedZones: []ZoneRect{
			{Name: "Donut", MinX: 0, MaxX: 800, MinY: -800, MaxY: 0},
			{Name: "Cave", MinX: -1600, MaxX: -800, MinY: 0, MaxY: 800},
		},
		SmokeSpots: []SmokeSpot{
			{Name: "CT", Position: model.Position{X: 1000, Y: -600, Z: 0}, Radius: 100, Side: "T"},
			{Name: "Cave", Position: model.Position{X: -1200, Y: 400, Z: 0}, Radius: 100, Side: "T"},
		},
	},
	"de_anubis": {
		Name:    "de_anubis",
		TSpawn:  model.Position{X: -1600, Y: 1200, Z: 0},
		CTSpawn: model.Position{X: 1200, Y: -800, Z: 0},
		BombsiteA: ZoneRect{
			Name: "A Site",
			MinX: 800, MaxX: 1600,
			MinY: 400, MaxY: 1200,
		},
		BombsiteB: ZoneRect{
			Name: "B Site",
			MinX: -800, MaxX: 0,
			MinY: -1200, MaxY: -400,
		},
		MidZone: ZoneRect{
			Name: "Mid",
			MinX: -400, MaxX: 400,
			MinY: -400, MaxY: 400,
		},
		ContestedZones: []ZoneRect{
			{Name: "Main", MinX: 0, MaxX: 800, MinY: 0, MaxY: 800},
			{Name: "Connector", MinX: -800, MaxX: 0, MinY: -400, MaxY: 400},
		},
		SmokeSpots: []SmokeSpot{
			{Name: "CT", Position: model.Position{X: 1000, Y: 0, Z: 0}, Radius: 100, Side: "T"},
			{Name: "Heaven", Position: model.Position{X: 1200, Y: 600, Z: 0}, Radius: 80, Side: "T"},
		},
	},
	"de_vertigo": {
		Name:    "de_vertigo",
		TSpawn:  model.Position{X: -1600, Y: -400, Z: 11700},
		CTSpawn: model.Position{X: -400, Y: 1200, Z: 11700},
		BombsiteA: ZoneRect{
			Name: "A Site",
			MinX: -1200, MaxX: -400,
			MinY: 800, MaxY: 1600,
		},
		BombsiteB: ZoneRect{
			Name: "B Site",
			MinX: -2400, MaxX: -1600,
			MinY: 400, MaxY: 1200,
		},
		MidZone: ZoneRect{
			Name: "Mid",
			MinX: -1600, MaxX: -800,
			MinY: 0, MaxY: 800,
		},
		ContestedZones: []ZoneRect{
			{Name: "Ramp", MinX: -1200, MaxX: -400, MinY: 0, MaxY: 800},
			{Name: "Stairs", MinX: -2000, MaxX: -1200, MinY: -400, MaxY: 400},
		},
		SmokeSpots: []SmokeSpot{
			{Name: "CT", Position: model.Position{X: -600, Y: 1000, Z: 11700}, Radius: 100, Side: "T"},
			{Name: "Elevator", Position: model.Position{X: -800, Y: 600, Z: 11700}, Radius: 80, Side: "T"},
		},
	},
}

// GetMapZones returns the zone definitions for a map.
func GetMapZones(mapName string) *MapZones {
	if zones, ok := mapZonesData[mapName]; ok {
		return zones
	}
	return nil
}

// SpatialAnalyzer handles spatial calculations for ML features.
type SpatialAnalyzer struct {
	mapZones *MapZones
	mapName  string
}

// NewSpatialAnalyzer creates a new spatial analyzer for the given map.
func NewSpatialAnalyzer(mapName string) *SpatialAnalyzer {
	return &SpatialAnalyzer{
		mapZones: GetMapZones(mapName),
		mapName:  mapName,
	}
}

// IsInContestedZone checks if a position is in a contested zone.
func (s *SpatialAnalyzer) IsInContestedZone(pos model.Position) bool {
	if s.mapZones == nil {
		return false
	}
	for _, zone := range s.mapZones.ContestedZones {
		if zone.Contains(pos) {
			return true
		}
	}
	return false
}

// GetContestedZoneName returns the name of the contested zone containing the position.
func (s *SpatialAnalyzer) GetContestedZoneName(pos model.Position) string {
	if s.mapZones == nil {
		return ""
	}
	for _, zone := range s.mapZones.ContestedZones {
		if zone.Contains(pos) {
			return zone.Name
		}
	}
	return ""
}

// CalculateUncontestedAdvance calculates the distance traveled before entering a contested zone.
func (s *SpatialAnalyzer) CalculateUncontestedAdvance(startPos, contactPos model.Position, side string) float64 {
	if s.mapZones == nil {
		// Fallback: use simple distance
		return startPos.Distance2D(contactPos)
	}

	// Get spawn position for the side
	var spawnPos model.Position
	if side == "T" {
		spawnPos = s.mapZones.TSpawn
	} else {
		spawnPos = s.mapZones.CTSpawn
	}

	// Calculate distance from spawn to first contact
	// This represents how far the player advanced before engagement
	return spawnPos.Distance2D(contactPos)
}

// CalculateSmokeEffectiveness calculates how effective a smoke placement is.
// Returns a value between 0 and 1.
func (s *SpatialAnalyzer) CalculateSmokeEffectiveness(smokePos model.Position, throwerSide string) float64 {
	if s.mapZones == nil {
		return 0.5 // Default effectiveness when no map data
	}

	bestScore := 0.0
	for _, spot := range s.mapZones.SmokeSpots {
		// Check if this smoke spot is relevant for the thrower's side
		if spot.Side != "both" && spot.Side != throwerSide {
			continue
		}

		// Calculate distance from optimal position
		distance := smokePos.Distance2D(spot.Position)
		if distance <= spot.Radius {
			// Perfect placement
			score := 1.0
			if score > bestScore {
				bestScore = score
			}
		} else if distance <= spot.Radius*2 {
			// Good placement
			score := 1.0 - (distance-spot.Radius)/(spot.Radius)
			if score > bestScore {
				bestScore = score
			}
		} else if distance <= spot.Radius*3 {
			// Acceptable placement
			score := 0.5 - (distance-spot.Radius*2)/(spot.Radius*2)
			if score > 0 && score > bestScore {
				bestScore = score
			}
		}
	}

	return bestScore
}

// CrossfireAnalyzer handles crossfire pair detection.
type CrossfireAnalyzer struct {
	playerPositions map[uint64]model.Position
	playerAngles    map[uint64]float64 // View angle in degrees
	playerTeams     map[uint64]common.Team
}

// NewCrossfireAnalyzer creates a new crossfire analyzer.
func NewCrossfireAnalyzer() *CrossfireAnalyzer {
	return &CrossfireAnalyzer{
		playerPositions: make(map[uint64]model.Position),
		playerAngles:    make(map[uint64]float64),
		playerTeams:     make(map[uint64]common.Team),
	}
}

// UpdatePlayer updates the position and angle for a player.
func (c *CrossfireAnalyzer) UpdatePlayer(steamID uint64, pos model.Position, viewAngle float64, team common.Team) {
	c.playerPositions[steamID] = pos
	c.playerAngles[steamID] = viewAngle
	c.playerTeams[steamID] = team
}

// Reset clears all player data.
func (c *CrossfireAnalyzer) Reset() {
	c.playerPositions = make(map[uint64]model.Position)
	c.playerAngles = make(map[uint64]float64)
	c.playerTeams = make(map[uint64]common.Team)
}

// FindCrossfirePairs finds pairs of teammates who are set up in crossfire positions.
// Returns a map of player ID to their crossfire partner ID and distance.
func (c *CrossfireAnalyzer) FindCrossfirePairs(team common.Team) map[uint64]CrossfirePair {
	pairs := make(map[uint64]CrossfirePair)

	// Get all players on the team
	var teamPlayers []uint64
	for id, t := range c.playerTeams {
		if t == team {
			teamPlayers = append(teamPlayers, id)
		}
	}

	// Check each pair of teammates
	for i := 0; i < len(teamPlayers); i++ {
		for j := i + 1; j < len(teamPlayers); j++ {
			p1 := teamPlayers[i]
			p2 := teamPlayers[j]

			pos1 := c.playerPositions[p1]
			pos2 := c.playerPositions[p2]
			angle1 := c.playerAngles[p1]
			angle2 := c.playerAngles[p2]

			// Check if they form a crossfire
			if c.isCrossfire(pos1, pos2, angle1, angle2) {
				distance := pos1.Distance2D(pos2)
				pairs[p1] = CrossfirePair{PartnerID: p2, Distance: distance}
				pairs[p2] = CrossfirePair{PartnerID: p1, Distance: distance}
			}
		}
	}

	return pairs
}

// CrossfirePair represents a crossfire partnership.
type CrossfirePair struct {
	PartnerID uint64
	Distance  float64
}

// isCrossfire checks if two players are set up in a crossfire position.
// A crossfire exists when two players can cover the same area from different angles.
func (c *CrossfireAnalyzer) isCrossfire(pos1, pos2 model.Position, angle1, angle2 float64) bool {
	// Calculate the angle between the two players
	dx := pos2.X - pos1.X
	dy := pos2.Y - pos1.Y
	angleBetween := math.Atan2(dy, dx) * 180 / math.Pi

	// Normalize angles to 0-360
	angle1 = normalizeAngle(angle1)
	angle2 = normalizeAngle(angle2)
	angleBetween = normalizeAngle(angleBetween)

	// Check if players are looking at roughly the same area
	// but from different angles (crossfire condition)

	// Calculate the angle difference between their view directions
	viewAngleDiff := math.Abs(angleDiff(angle1, angle2))

	// For a crossfire, players should be looking in somewhat different directions
	// (between 30 and 150 degrees apart)
	if viewAngleDiff < 30 || viewAngleDiff > 150 {
		return false
	}

	// Check if they're at a reasonable distance for crossfire (50-500 units)
	distance := pos1.Distance2D(pos2)
	if distance < 50 || distance > 500 {
		return false
	}

	// Check if their view angles converge on a common area
	// This is a simplified check - in reality would need line-of-sight
	return true
}

// normalizeAngle normalizes an angle to 0-360 degrees.
func normalizeAngle(angle float64) float64 {
	for angle < 0 {
		angle += 360
	}
	for angle >= 360 {
		angle -= 360
	}
	return angle
}

// angleDiff calculates the smallest difference between two angles.
func angleDiff(a1, a2 float64) float64 {
	diff := a1 - a2
	for diff > 180 {
		diff -= 360
	}
	for diff < -180 {
		diff += 360
	}
	return diff
}

// AimTracker tracks crosshair displacement for players.
type AimTracker struct {
	lastAngles        map[uint64]float64 // Last recorded view angle
	displacementCount map[uint64]int     // Number of significant crosshair movements
	lastUpdateTick    map[uint64]int     // Last tick when angle was updated
}

// NewAimTracker creates a new aim tracker.
func NewAimTracker() *AimTracker {
	return &AimTracker{
		lastAngles:        make(map[uint64]float64),
		displacementCount: make(map[uint64]int),
		lastUpdateTick:    make(map[uint64]int),
	}
}

// Reset clears all tracking data.
func (a *AimTracker) Reset() {
	a.lastAngles = make(map[uint64]float64)
	a.displacementCount = make(map[uint64]int)
	a.lastUpdateTick = make(map[uint64]int)
}

// UpdateAngle updates the view angle for a player and tracks displacement.
// Returns true if a significant displacement was detected.
func (a *AimTracker) UpdateAngle(steamID uint64, viewAngle float64, currentTick int) bool {
	lastAngle, exists := a.lastAngles[steamID]
	lastTick := a.lastUpdateTick[steamID]

	a.lastAngles[steamID] = viewAngle
	a.lastUpdateTick[steamID] = currentTick

	if !exists {
		return false
	}

	// Only count displacement if enough time has passed (avoid counting micro-adjustments)
	tickDiff := currentTick - lastTick
	if tickDiff < 8 { // About 0.125 seconds at 64 tick
		return false
	}

	// Calculate angle change
	angleChange := math.Abs(angleDiff(viewAngle, lastAngle))

	// Significant displacement is > 30 degrees
	if angleChange > 30 {
		a.displacementCount[steamID]++
		return true
	}

	return false
}

// GetDisplacementCount returns the number of significant crosshair displacements for a player.
func (a *AimTracker) GetDisplacementCount(steamID uint64) int {
	return a.displacementCount[steamID]
}

// GetAllDisplacements returns displacement counts for all tracked players.
func (a *AimTracker) GetAllDisplacements() map[uint64]int {
	result := make(map[uint64]int)
	for id, count := range a.displacementCount {
		result[id] = count
	}
	return result
}
