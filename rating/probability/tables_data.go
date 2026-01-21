package probability

// DefaultTables returns probability tables with empirically-derived values
// from parsing 35,648 rounds and 255,391 kills across competitive CS2 demos.
func DefaultTables() *ProbabilityTables {
	tables := NewProbabilityTables()

	// === BASE WIN PROBABILITIES ===
	// Format: "TvCT_bombStatus" (e.g., "5v4_none", "3v2_planted")
	// Values are T-side win probability derived from empirical data

	// No bomb planted
	tables.BaseWinProb["5v5_none"] = 0.496 // 17654 / 35555
	tables.BaseWinProb["5v4_none"] = 0.712 // 11826 / 16619
	tables.BaseWinProb["5v3_none"] = 0.890 // 6674 / 7496
	tables.BaseWinProb["5v2_none"] = 0.979 // 3237 / 3307
	tables.BaseWinProb["5v1_none"] = 0.999 // 1354 / 1355
	tables.BaseWinProb["5v0_none"] = 0.602 // 118 / 196 (rare edge case)

	tables.BaseWinProb["4v5_none"] = 0.302 // 5638 / 18676
	tables.BaseWinProb["4v4_none"] = 0.510 // 9059 / 17750
	tables.BaseWinProb["4v3_none"] = 0.745 // 8662 / 11624
	tables.BaseWinProb["4v2_none"] = 0.923 // 5957 / 6453
	tables.BaseWinProb["4v1_none"] = 0.994 // 2770 / 2788
	tables.BaseWinProb["4v0_none"] = 0.599 // 235 / 392 (rare edge case)

	tables.BaseWinProb["3v5_none"] = 0.137 // 1297 / 9481
	tables.BaseWinProb["3v4_none"] = 0.289 // 3924 / 13592
	tables.BaseWinProb["3v3_none"] = 0.514 // 5954 / 11579
	tables.BaseWinProb["3v2_none"] = 0.780 // 5780 / 7410
	tables.BaseWinProb["3v1_none"] = 0.961 // 3353 / 3488
	tables.BaseWinProb["3v0_none"] = 0.591 // 231 / 391 (rare edge case)

	tables.BaseWinProb["2v5_none"] = 0.031 // 148 / 4792
	tables.BaseWinProb["2v4_none"] = 0.098 // 869 / 8876
	tables.BaseWinProb["2v3_none"] = 0.236 // 2174 / 9206
	tables.BaseWinProb["2v2_none"] = 0.491 // 3180 / 6476
	tables.BaseWinProb["2v1_none"] = 0.823 // 2613 / 3175
	tables.BaseWinProb["2v0_none"] = 0.686 // 179 / 261 (rare edge case)

	tables.BaseWinProb["1v5_none"] = 0.004 // 11 / 2593
	tables.BaseWinProb["1v4_none"] = 0.012 // 68 / 5735
	tables.BaseWinProb["1v3_none"] = 0.047 // 320 / 6823
	tables.BaseWinProb["1v2_none"] = 0.148 // 780 / 5256
	tables.BaseWinProb["1v1_none"] = 0.435 // 1126 / 2589
	tables.BaseWinProb["1v0_none"] = 0.605 // 89 / 147 (rare edge case)

	tables.BaseWinProb["0v5_none"] = 0.000
	tables.BaseWinProb["0v4_none"] = 0.500 // 1 / 2 (rare edge case)
	tables.BaseWinProb["0v3_none"] = 0.500 // 1 / 2 (rare edge case)
	tables.BaseWinProb["0v2_none"] = 0.000 // 0 / 1
	tables.BaseWinProb["0v1_none"] = 0.000
	tables.BaseWinProb["0v0_none"] = 0.000

	// Bomb planted - T-side advantage
	tables.BaseWinProb["5v5_planted"] = 0.806 // 170 / 211
	tables.BaseWinProb["5v4_planted"] = 0.855 // 478 / 559
	tables.BaseWinProb["5v3_planted"] = 0.942 // 917 / 973
	tables.BaseWinProb["5v2_planted"] = 0.988 // 1180 / 1194
	tables.BaseWinProb["5v1_planted"] = 0.994 // 1105 / 1112
	tables.BaseWinProb["5v0_planted"] = 1.000

	tables.BaseWinProb["4v5_planted"] = 0.530 // 122 / 230
	tables.BaseWinProb["4v4_planted"] = 0.675 // 686 / 1016
	tables.BaseWinProb["4v3_planted"] = 0.835 // 1837 / 2199
	tables.BaseWinProb["4v2_planted"] = 0.957 // 3028 / 3163
	tables.BaseWinProb["4v1_planted"] = 0.976 // 3055 / 3130
	tables.BaseWinProb["4v0_planted"] = 1.000 // 3 / 3

	tables.BaseWinProb["3v5_planted"] = 0.249 // 50 / 201
	tables.BaseWinProb["3v4_planted"] = 0.403 // 437 / 1085
	tables.BaseWinProb["3v3_planted"] = 0.643 // 1814 / 2819
	tables.BaseWinProb["3v2_planted"] = 0.857 // 4003 / 4669
	tables.BaseWinProb["3v1_planted"] = 0.970 // 4742 / 4888
	tables.BaseWinProb["3v0_planted"] = 1.000 // 2 / 2

	tables.BaseWinProb["2v5_planted"] = 0.076 // 10 / 131
	tables.BaseWinProb["2v4_planted"] = 0.143 // 127 / 886
	tables.BaseWinProb["2v3_planted"] = 0.344 // 868 / 2522
	tables.BaseWinProb["2v2_planted"] = 0.629 // 2968 / 4719
	tables.BaseWinProb["2v1_planted"] = 0.898 // 4871 / 5425
	tables.BaseWinProb["2v0_planted"] = 0.600 // 3 / 5 (rare edge case)

	tables.BaseWinProb["1v5_planted"] = 0.019 // 2 / 105
	tables.BaseWinProb["1v4_planted"] = 0.031 // 19 / 621
	tables.BaseWinProb["1v3_planted"] = 0.087 // 160 / 1832
	tables.BaseWinProb["1v2_planted"] = 0.269 // 988 / 3669
	tables.BaseWinProb["1v1_planted"] = 0.604 // 2778 / 4597
	tables.BaseWinProb["1v0_planted"] = 0.750 // 3 / 4 (rare edge case)

	// 0 T alive but bomb planted - depends on defuse time
	tables.BaseWinProb["0v5_planted"] = 0.000 // 0 / 70
	tables.BaseWinProb["0v4_planted"] = 0.000 // 0 / 340
	tables.BaseWinProb["0v3_planted"] = 0.000 // 0 / 940
	tables.BaseWinProb["0v2_planted"] = 0.001 // 1 / 1729
	tables.BaseWinProb["0v1_planted"] = 0.004 // 6 / 1500
	tables.BaseWinProb["0v0_planted"] = 1.000 // Bomb explodes, T wins

	// === DUEL WIN RATES ===
	// Format: "attacker_vs_defender" (e.g., "rifle_vs_pistol")
	// Values represent attacker win probability from empirical data

	// Starter Pistol attacking
	tables.DuelWinRates["starter_pistol_vs_starter_pistol"] = 0.500  // 19679 / 39358
	tables.DuelWinRates["starter_pistol_vs_upgraded_pistol"] = 0.527 // 2175 / 4129
	tables.DuelWinRates["starter_pistol_vs_smg"] = 0.264             // 1471 / 5569
	tables.DuelWinRates["starter_pistol_vs_rifle"] = 0.252           // 2485 / 9880
	tables.DuelWinRates["starter_pistol_vs_awp"] = 0.268             // 3882 / 14500

	// Upgraded Pistol attacking
	tables.DuelWinRates["upgraded_pistol_vs_starter_pistol"] = 0.473  // 1954 / 4129
	tables.DuelWinRates["upgraded_pistol_vs_upgraded_pistol"] = 0.500 // 208 / 416
	tables.DuelWinRates["upgraded_pistol_vs_smg"] = 0.334             // 513 / 1538
	tables.DuelWinRates["upgraded_pistol_vs_rifle"] = 0.360           // 1646 / 4577
	tables.DuelWinRates["upgraded_pistol_vs_awp"] = 0.346             // 4717 / 13613

	// SMG attacking
	tables.DuelWinRates["smg_vs_starter_pistol"] = 0.736  // 4098 / 5569
	tables.DuelWinRates["smg_vs_upgraded_pistol"] = 0.666 // 1025 / 1538
	tables.DuelWinRates["smg_vs_smg"] = 0.500             // 2821 / 5642
	tables.DuelWinRates["smg_vs_rifle"] = 0.426           // 4931 / 11569
	tables.DuelWinRates["smg_vs_awp"] = 0.402             // 9155 / 22791

	// Rifle attacking
	tables.DuelWinRates["rifle_vs_starter_pistol"] = 0.748  // 7395 / 9880
	tables.DuelWinRates["rifle_vs_upgraded_pistol"] = 0.640 // 2931 / 4577
	tables.DuelWinRates["rifle_vs_smg"] = 0.574             // 6638 / 11569
	tables.DuelWinRates["rifle_vs_rifle"] = 0.500           // 15495 / 30990
	tables.DuelWinRates["rifle_vs_awp"] = 0.467             // 32361 / 69317

	// AWP/Full Buy attacking
	tables.DuelWinRates["awp_vs_starter_pistol"] = 0.732  // 10618 / 14500
	tables.DuelWinRates["awp_vs_upgraded_pistol"] = 0.653 // 8896 / 13613
	tables.DuelWinRates["awp_vs_smg"] = 0.598             // 13636 / 22791
	tables.DuelWinRates["awp_vs_rifle"] = 0.533           // 36956 / 69317
	tables.DuelWinRates["awp_vs_awp"] = 0.500             // 59705 / 119410

	// === MAP T-SIDE WIN RATES ===
	// Empirically derived from demo data

	tables.MapAdjustments["de_ancient"] = 0.507  // 2660 / 5242
	tables.MapAdjustments["de_anubis"] = 0.551   // 927 / 1682 (T-sided)
	tables.MapAdjustments["de_dust2"] = 0.507    // 2555 / 5038
	tables.MapAdjustments["de_inferno"] = 0.514  // 2559 / 4978
	tables.MapAdjustments["de_mirage"] = 0.500   // 2584 / 5171
	tables.MapAdjustments["de_nuke"] = 0.475     // 2404 / 5063 (CT-sided)
	tables.MapAdjustments["de_overpass"] = 0.488 // 2665 / 5457
	tables.MapAdjustments["de_train"] = 0.448    // 1350 / 3017 (CT-sided)
	tables.MapAdjustments["de_vertigo"] = 0.480  // estimated

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
