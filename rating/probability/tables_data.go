package probability

// DefaultTables returns probability tables with empirically-derived values
// from parsing 54,583 rounds and 390,515 kills across competitive CS2 demos.
func DefaultTables() *ProbabilityTables {
	tables := NewProbabilityTables()

	// === BASE WIN PROBABILITIES ===
	// Format: "TvCT_bombStatus" (e.g., "5v4_none", "3v2_planted")
	// Values are T-side win probability derived from empirical data

	// No bomb planted
	tables.BaseWinProb["5v5_none"] = 0.502 // 27334 / 54432
	tables.BaseWinProb["5v4_none"] = 0.712 // 18275 / 25675
	tables.BaseWinProb["5v3_none"] = 0.887 // 10384 / 11710
	tables.BaseWinProb["5v2_none"] = 0.978 // 5048 / 5164
	tables.BaseWinProb["5v1_none"] = 0.998 // 2130 / 2135
	tables.BaseWinProb["5v0_none"] = 0.610 // 192 / 315 (rare edge case)

	tables.BaseWinProb["4v5_none"] = 0.308 // 8716 / 28314
	tables.BaseWinProb["4v4_none"] = 0.515 // 13931 / 27060
	tables.BaseWinProb["4v3_none"] = 0.746 // 13359 / 17904
	tables.BaseWinProb["4v2_none"] = 0.924 // 9130 / 9876
	tables.BaseWinProb["4v1_none"] = 0.994 // 4366 / 4394
	tables.BaseWinProb["4v0_none"] = 0.610 // 381 / 625 (rare edge case)

	tables.BaseWinProb["3v5_none"] = 0.138 // 1961 / 14256
	tables.BaseWinProb["3v4_none"] = 0.289 // 5908 / 20456
	tables.BaseWinProb["3v3_none"] = 0.514 // 9041 / 17585
	tables.BaseWinProb["3v2_none"] = 0.779 // 8723 / 11203
	tables.BaseWinProb["3v1_none"] = 0.963 // 5026 / 5221
	tables.BaseWinProb["3v0_none"] = 0.610 // 374 / 613 (rare edge case)

	tables.BaseWinProb["2v5_none"] = 0.032 // 232 / 7224
	tables.BaseWinProb["2v4_none"] = 0.098 // 1311 / 13388
	tables.BaseWinProb["2v3_none"] = 0.236 // 3287 / 13907
	tables.BaseWinProb["2v2_none"] = 0.491 // 4865 / 9912
	tables.BaseWinProb["2v1_none"] = 0.820 // 3942 / 4806
	tables.BaseWinProb["2v0_none"] = 0.659 // 261 / 396 (rare edge case)

	tables.BaseWinProb["1v5_none"] = 0.004 // 14 / 3873
	tables.BaseWinProb["1v4_none"] = 0.011 // 93 / 8658
	tables.BaseWinProb["1v3_none"] = 0.046 // 464 / 10176
	tables.BaseWinProb["1v2_none"] = 0.150 // 1180 / 7866
	tables.BaseWinProb["1v1_none"] = 0.440 // 1720 / 3909
	tables.BaseWinProb["1v0_none"] = 0.657 // 138 / 210 (rare edge case)

	tables.BaseWinProb["0v5_none"] = 0.500 // 1 / 2 (rare edge case)
	tables.BaseWinProb["0v4_none"] = 0.500 // 1 / 2 (rare edge case)
	tables.BaseWinProb["0v3_none"] = 0.429 // 3 / 7 (rare edge case)
	tables.BaseWinProb["0v2_none"] = 0.000 // 0 / 1
	tables.BaseWinProb["0v1_none"] = 0.000
	tables.BaseWinProb["0v0_none"] = 0.000

	// Bomb planted - T-side advantage
	tables.BaseWinProb["5v5_planted"] = 0.812 // 315 / 388
	tables.BaseWinProb["5v4_planted"] = 0.857 // 822 / 959
	tables.BaseWinProb["5v3_planted"] = 0.942 // 1515 / 1608
	tables.BaseWinProb["5v2_planted"] = 0.985 // 1870 / 1898
	tables.BaseWinProb["5v1_planted"] = 0.991 // 1761 / 1777
	tables.BaseWinProb["5v0_planted"] = 1.000

	tables.BaseWinProb["4v5_planted"] = 0.530 // 218 / 411
	tables.BaseWinProb["4v4_planted"] = 0.686 // 1146 / 1671
	tables.BaseWinProb["4v3_planted"] = 0.840 // 3007 / 3579
	tables.BaseWinProb["4v2_planted"] = 0.958 // 4736 / 4942
	tables.BaseWinProb["4v1_planted"] = 0.977 // 4809 / 4923
	tables.BaseWinProb["4v0_planted"] = 1.000 // 3 / 3

	tables.BaseWinProb["3v5_planted"] = 0.254 // 86 / 339
	tables.BaseWinProb["3v4_planted"] = 0.416 // 719 / 1730
	tables.BaseWinProb["3v3_planted"] = 0.644 // 2920 / 4531
	tables.BaseWinProb["3v2_planted"] = 0.858 // 6228 / 7254
	tables.BaseWinProb["3v1_planted"] = 0.967 // 7244 / 7489
	tables.BaseWinProb["3v0_planted"] = 1.000 // 3 / 3

	tables.BaseWinProb["2v5_planted"] = 0.074 // 16 / 217
	tables.BaseWinProb["2v4_planted"] = 0.153 // 209 / 1368
	tables.BaseWinProb["2v3_planted"] = 0.338 // 1359 / 4024
	tables.BaseWinProb["2v2_planted"] = 0.630 // 4605 / 7312
	tables.BaseWinProb["2v1_planted"] = 0.895 // 7362 / 8225
	tables.BaseWinProb["2v0_planted"] = 0.600 // 3 / 5 (rare edge case)

	tables.BaseWinProb["1v5_planted"] = 0.018 // 3 / 163
	tables.BaseWinProb["1v4_planted"] = 0.033 // 31 / 940
	tables.BaseWinProb["1v3_planted"] = 0.086 // 247 / 2887
	tables.BaseWinProb["1v2_planted"] = 0.265 // 1496 / 5643
	tables.BaseWinProb["1v1_planted"] = 0.600 // 4190 / 6984
	tables.BaseWinProb["1v0_planted"] = 0.750 // 3 / 4 (rare edge case)

	// 0 T alive but bomb planted - depends on defuse time
	tables.BaseWinProb["0v5_planted"] = 0.000 // 0 / 97
	tables.BaseWinProb["0v4_planted"] = 0.000 // 0 / 512
	tables.BaseWinProb["0v3_planted"] = 0.001 // 1 / 1472
	tables.BaseWinProb["0v2_planted"] = 0.001 // 3 / 2696
	tables.BaseWinProb["0v1_planted"] = 0.004 // 9 / 2321
	tables.BaseWinProb["0v0_planted"] = 1.000 // Bomb explodes, T wins

	// === DUEL WIN RATES ===
	// Format: "attacker_vs_defender" (e.g., "rifle_vs_pistol")
	// Values represent attacker win probability from empirical data

	// Starter Pistol attacking
	tables.DuelWinRates["starter_pistol_vs_starter_pistol"] = 0.500  // 30039 / 60078
	tables.DuelWinRates["starter_pistol_vs_upgraded_pistol"] = 0.520 // 3335 / 6415
	tables.DuelWinRates["starter_pistol_vs_smg"] = 0.257             // 2338 / 9109
	tables.DuelWinRates["starter_pistol_vs_rifle"] = 0.248           // 3748 / 15102
	tables.DuelWinRates["starter_pistol_vs_awp"] = 0.270             // 6037 / 22369

	// Upgraded Pistol attacking
	tables.DuelWinRates["upgraded_pistol_vs_starter_pistol"] = 0.480  // 3080 / 6415
	tables.DuelWinRates["upgraded_pistol_vs_upgraded_pistol"] = 0.500 // 334 / 668
	tables.DuelWinRates["upgraded_pistol_vs_smg"] = 0.348             // 872 / 2508
	tables.DuelWinRates["upgraded_pistol_vs_rifle"] = 0.360           // 2532 / 7040
	tables.DuelWinRates["upgraded_pistol_vs_awp"] = 0.349             // 7108 / 20393

	// SMG attacking
	tables.DuelWinRates["smg_vs_starter_pistol"] = 0.743  // 6771 / 9109
	tables.DuelWinRates["smg_vs_upgraded_pistol"] = 0.652 // 1636 / 2508
	tables.DuelWinRates["smg_vs_smg"] = 0.500             // 4738 / 9476
	tables.DuelWinRates["smg_vs_rifle"] = 0.431           // 8134 / 18879
	tables.DuelWinRates["smg_vs_awp"] = 0.402             // 14264 / 35512

	// Rifle attacking
	tables.DuelWinRates["rifle_vs_starter_pistol"] = 0.752  // 11354 / 15102
	tables.DuelWinRates["rifle_vs_upgraded_pistol"] = 0.640 // 4508 / 7040
	tables.DuelWinRates["rifle_vs_smg"] = 0.569             // 10745 / 18879
	tables.DuelWinRates["rifle_vs_rifle"] = 0.500           // 24356 / 48712
	tables.DuelWinRates["rifle_vs_awp"] = 0.467             // 48770 / 104311

	// AWP/Full Buy attacking
	tables.DuelWinRates["awp_vs_starter_pistol"] = 0.730  // 16332 / 22369
	tables.DuelWinRates["awp_vs_upgraded_pistol"] = 0.651 // 13285 / 20393
	tables.DuelWinRates["awp_vs_smg"] = 0.598             // 21248 / 35512
	tables.DuelWinRates["awp_vs_rifle"] = 0.533           // 55541 / 104311
	tables.DuelWinRates["awp_vs_awp"] = 0.500             // 89410 / 178820

	// === MAP T-SIDE WIN RATES ===
	// Empirically derived from demo data

	tables.MapAdjustments["de_ancient"] = 0.513  // 4120 / 8026
	tables.MapAdjustments["de_anubis"] = 0.564   // 2623 / 4652 (T-sided)
	tables.MapAdjustments["de_dust2"] = 0.519    // 3822 / 7366
	tables.MapAdjustments["de_inferno"] = 0.512  // 3848 / 7517
	tables.MapAdjustments["de_mirage"] = 0.498   // 3711 / 7457
	tables.MapAdjustments["de_nuke"] = 0.480     // 3831 / 7984 (CT-sided)
	tables.MapAdjustments["de_overpass"] = 0.488 // 2665 / 5457
	tables.MapAdjustments["de_train"] = 0.456    // 2793 / 6124 (CT-sided)

	return tables
}

// NormalizationConstants for converting swing to rating contribution.
const (
	// SwingToRatingScale converts average swing per round to a ~1.0 centered rating.
	// A player with +4% avg swing should get ~1.40 rating contribution.
	// A player with -3% avg swing should get ~0.70 rating contribution.
	SwingToRatingScale = 10.0

	// SwingRatingBaseline is the baseline for swing rating (average = 1.0).
	SwingRatingBaseline = 1.0

	// MinSwingRating is the minimum possible swing rating component.
	MinSwingRating = 0.40

	// MaxSwingRating is the maximum possible swing rating component.
	MaxSwingRating = 1.80
)
