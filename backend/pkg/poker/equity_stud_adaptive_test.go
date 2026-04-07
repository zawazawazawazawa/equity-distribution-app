package poker

import (
	"testing"

	"github.com/chehsunliu/poker"
)

func TestCalculateStudEquityAdaptive(t *testing.T) {
	config := GetDefaultAdaptiveConfig()
	config.MinIterations = 100
	config.TargetPrecision = 2.0 // Allow 2% standard deviation for faster convergence in tests

	tests := []struct {
		name           string
		yourHand       []poker.Card
		opponentHand   []poker.Card
		knownCards     []poker.Card
		gameType       StudGameType
		expectedMin    float64
		expectedMax    float64
		expectConverge bool
	}{
		{
			name: "Razz - Strong made hand should converge quickly",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("4c"), poker.NewCard("5s"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ks"), poker.NewCard("Qd"), poker.NewCard("Jh"),
				poker.NewCard("Tc"), poker.NewCard("9s"),
			},
			knownCards:     []poker.Card{},
			gameType:       StudRazz,
			expectedMin:    95.0,
			expectedMax:    100.0,
			expectConverge: true,
		},
		{
			name: "Stud High - Rolled up trips vs random",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("Ad"), poker.NewCard("Ah"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Jh"), poker.NewCard("Tc"), poker.NewCard("9s"),
			},
			knownCards:     []poker.Card{},
			gameType:       StudHigh,
			expectedMin:    80.0,
			expectedMax:    95.0,
			expectConverge: true,
		},
		{
			name: "Stud8 - Wheel draw vs high only",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("5c"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ks"), poker.NewCard("Kd"), poker.NewCard("Qh"),
				poker.NewCard("Qc"),
			},
			knownCards:     []poker.Card{},
			gameType:       StudHighLow8,
			expectedMin:    40.0, // Overall equity (high + low) / 2
			expectedMax:    60.0,
			expectConverge: true,
		},
		{
			name: "Close matchup - may not converge quickly",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("Ks"), poker.NewCard("Qd"),
				poker.NewCard("Jh"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ad"), poker.NewCard("Kd"), poker.NewCard("Qh"),
				poker.NewCard("Jc"),
			},
			knownCards:     []poker.Card{},
			gameType:       StudHigh,
			expectedMin:    45.0,
			expectedMax:    55.0,
			expectConverge: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, iterations, err := CalculateStudEquityAdaptive(
				tt.yourHand, tt.opponentHand, tt.knownCards, tt.gameType, config,
			)

			if err != nil {
				t.Fatalf("Error calculating equity: %v", err)
			}

			// Check equity is within expected range
			equity := result.Equity
			if tt.gameType == StudHighLow8 && result.Stud8Result != nil {
				// For Stud8, verify the overall equity calculation
				equity = result.Equity
			}

			if equity < tt.expectedMin || equity > tt.expectedMax {
				t.Errorf("Expected equity between %.1f%% and %.1f%%, got %.1f%%",
					tt.expectedMin, tt.expectedMax, equity)
			}

			// Check convergence behavior
			if tt.expectConverge && iterations >= config.MaxIterations {
				t.Logf("Warning: Expected convergence but used all %d iterations", iterations)
			} else if tt.expectConverge && iterations < config.MinIterations {
				t.Errorf("Converged too quickly with only %d iterations", iterations)
			}

			t.Logf("Converged after %d iterations with equity %.2f%%", iterations, equity)
		})
	}
}

func TestStudAdaptiveConvergenceAccuracy(t *testing.T) {
	// Test that adaptive sampling produces accurate results
	yourHand := []poker.Card{
		poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
		poker.NewCard("4c"), poker.NewCard("5s"), // Made wheel in Razz
	}
	opponentHand := []poker.Card{
		poker.NewCard("6s"), poker.NewCard("7d"), poker.NewCard("8h"),
		poker.NewCard("9c"), poker.NewCard("Ts"), // Made T-high
	}

	// Calculate with fixed high iterations for baseline
	baselineResult, err := CalculateStudEquity(yourHand, opponentHand, []poker.Card{}, StudRazz, 10000)
	if err != nil {
		t.Fatalf("Error calculating baseline equity: %v", err)
	}

	// Calculate with adaptive sampling
	config := GetDefaultAdaptiveConfig()
	config.TargetPrecision = 0.5 // Tight precision requirement
	adaptiveResult, iterations, err := CalculateStudEquityAdaptive(
		yourHand, opponentHand, []poker.Card{}, StudRazz, config,
	)
	if err != nil {
		t.Fatalf("Error calculating adaptive equity: %v", err)
	}

	// Results should be very close
	diff := adaptiveResult.Equity - baselineResult.Equity
	if diff < -2.0 || diff > 2.0 {
		t.Errorf("Adaptive result (%.2f%%) differs too much from baseline (%.2f%%)",
			adaptiveResult.Equity, baselineResult.Equity)
	}

	t.Logf("Baseline: %.2f%% (10000 iterations), Adaptive: %.2f%% (%d iterations)",
		baselineResult.Equity, adaptiveResult.Equity, iterations)
}

func TestStud8AdaptiveDetails(t *testing.T) {
	// Test that Stud8 adaptive sampling properly tracks all components
	yourHand := []poker.Card{
		poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
		poker.NewCard("4c"), // Low draw
	}
	opponentHand := []poker.Card{
		poker.NewCard("Ks"), poker.NewCard("Kd"), poker.NewCard("Qh"),
		poker.NewCard("Qc"), // Two pair for high
	}

	config := GetDefaultAdaptiveConfig()
	result, iterations, err := CalculateStudEquityAdaptive(
		yourHand, opponentHand, []poker.Card{}, StudHighLow8, config,
	)

	if err != nil {
		t.Fatalf("Error calculating equity: %v", err)
	}

	if result.Stud8Result == nil {
		t.Fatal("Expected Stud8Result to be populated")
	}

	// Verify all components are reasonable
	if result.Stud8Result.HighEquity < 0 || result.Stud8Result.HighEquity > 100 {
		t.Errorf("Invalid high equity: %.2f%%", result.Stud8Result.HighEquity)
	}
	if result.Stud8Result.LowEquity < 0 || result.Stud8Result.LowEquity > 100 {
		t.Errorf("Invalid low equity: %.2f%%", result.Stud8Result.LowEquity)
	}
	if result.Stud8Result.ScoopEquity < 0 || result.Stud8Result.ScoopEquity > 100 {
		t.Errorf("Invalid scoop equity: %.2f%%", result.Stud8Result.ScoopEquity)
	}

	// Overall equity should be average of high and low
	expectedOverall := (result.Stud8Result.HighEquity + result.Stud8Result.LowEquity) / 2
	if diff := result.Equity - expectedOverall; diff < -0.1 || diff > 0.1 {
		t.Errorf("Overall equity (%.2f%%) doesn't match (high+low)/2 (%.2f%%)",
			result.Equity, expectedOverall)
	}

	t.Logf("Stud8 Results after %d iterations:", iterations)
	t.Logf("  High: %.2f%%, Low: %.2f%%, Scoop: %.2f%%, Overall: %.2f%%",
		result.Stud8Result.HighEquity, result.Stud8Result.LowEquity,
		result.Stud8Result.ScoopEquity, result.Equity)
}