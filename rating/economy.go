package rating

// EcoKillValue calculates the value of a kill based on equipment difference
// Highly rewards pistol kills on rifle players, normal value for equal fights
func EcoKillValue(attackerEquip, victimEquip float64) float64 {
	// Calculate equipment ratio (victim/attacker)
	// Higher ratio = attacker had worse equipment = more impressive kill
	if attackerEquip < MinEquipmentValue {
		attackerEquip = MinEquipmentValue
	}

	ratio := victimEquip / attackerEquip

	if ratio > 4.0 {
		return EcoKillPistolVsRifle
	} else if ratio > 2.0 {
		return EcoKillEcoVsForce
	} else if ratio > 1.3 {
		return EcoKillForceVsFullBuy
	} else if ratio > 1.1 {
		return EcoKillSlightDisadvantage
	} else if ratio > 0.9 {
		return EcoKillEqual
	} else if ratio > 0.75 {
		return EcoKillSlightAdvantage
	} else if ratio > 0.5 {
		return EcoKillAdvantage
	} else {
		return EcoKillRifleVsPistol
	}
}

// EcoDeathPenalty calculates the penalty for dying based on equipment difference
// Highly punishes dying to pistols when you have rifles
func EcoDeathPenalty(victimEquip, killerEquip float64) float64 {
	if killerEquip < MinEquipmentValue {
		killerEquip = MinEquipmentValue
	}

	// Calculate ratio (victim/killer)
	// Higher ratio = victim had better equipment = more embarrassing death
	ratio := victimEquip / killerEquip

	if ratio > 4.0 {
		return EcoDeathToPistol
	} else if ratio > 2.0 {
		return EcoDeathToEco
	} else if ratio > 1.3 {
		return EcoDeathToForceBuy
	} else if ratio > 1.1 {
		return EcoDeathSlightAdvantage
	} else if ratio > 0.9 {
		return EcoDeathEqual
	} else if ratio > 0.75 {
		return EcoDeathSlightDisadvantage
	} else if ratio > 0.5 {
		return EcoDeathDisadvantage
	} else {
		return EcoDeathPistolVsRifle
	}
}

// EconWeight is kept for backward compatibility
func EconWeight(attackerValue, victimValue float64) float64 {
	return EcoKillValue(attackerValue, victimValue)
}

func RoundImportance(teamValue float64) float64 {
	if teamValue < 10000 {
		return 0.7 // eco
	}
	if teamValue < 20000 {
		return 1.0 // force
	}
	return 1.2 // full buy
}
