package poker

import (
	"math"
	"testing"

	"github.com/chehsunliu/poker"
)

func TestCalculateRazzEquity(t *testing.T) {
	tests := []struct {
		name           string
		yourHand       []poker.Card
		opponentHand   []poker.Card
		knownCards     []poker.Card
		expectedMin    float64
		expectedMax    float64
	}{
		{
			name: "Wheel vs high cards - massive favorite",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("4c"), poker.NewCard("5s"), poker.NewCard("6h"), poker.NewCard("7d"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ks"), poker.NewCard("Qd"), poker.NewCard("Jh"),
				poker.NewCard("Tc"), poker.NewCard("9s"), poker.NewCard("9d"), poker.NewCard("8h"),
			},
			knownCards:  []poker.Card{},
			expectedMin: 99.0, // Should win almost always
			expectedMax: 100.0,
		},
		{
			name: "Made 7-low vs drawing hand",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("5c"), poker.NewCard("7s"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ad"), poker.NewCard("4d"), poker.NewCard("6h"),
				poker.NewCard("Kc"), poker.NewCard("Qs"),
			},
			knownCards:  []poker.Card{},
			expectedMin: 95.0, // Made 7-low is huge favorite
			expectedMax: 100.0,
		},
		{
			name: "Mirror hands should be close to 50%",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("4c"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ad"), poker.NewCard("2h"), poker.NewCard("3s"),
				poker.NewCard("4d"),
			},
			knownCards:  []poker.Card{},
			expectedMin: 45.0,
			expectedMax: 55.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equity, err := CalculateRazzEquity(tt.yourHand, tt.opponentHand, tt.knownCards, 10000)
			if err != nil {
				t.Fatalf("Error calculating equity: %v", err)
			}

			if equity < tt.expectedMin || equity > tt.expectedMax {
				t.Errorf("Expected equity between %.1f%% and %.1f%%, got %.1f%%",
					tt.expectedMin, tt.expectedMax, equity)
			}
		})
	}
}

func TestCalculateStudHighEquity(t *testing.T) {
	tests := []struct {
		name           string
		yourHand       []poker.Card
		opponentHand   []poker.Card
		knownCards     []poker.Card
		expectedMin    float64
		expectedMax    float64
	}{
		{
			name: "Rolled up trips vs high cards",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("Ad"), poker.NewCard("Ah"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ks"), poker.NewCard("Qd"), poker.NewCard("Jh"),
			},
			knownCards:  []poker.Card{},
			expectedMin: 80.0, // Trips should be huge favorite
			expectedMax: 95.0,
		},
		{
			name: "Flush draw vs pair",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("Ks"), poker.NewCard("Qs"),
				poker.NewCard("Js"), poker.NewCard("2d"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Kd"), poker.NewCard("Kh"), poker.NewCard("7c"),
				poker.NewCard("6h"), poker.NewCard("5d"),
			},
			knownCards:  []poker.Card{},
			expectedMin: 40.0, // Flush draw should have decent equity
			expectedMax: 60.0,
		},
		{
			name: "Complete hands - flush vs two pair",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("9s"), poker.NewCard("7s"),
				poker.NewCard("5s"), poker.NewCard("3s"), poker.NewCard("2h"), poker.NewCard("2c"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Kh"), poker.NewCard("Kd"), poker.NewCard("Qc"),
				poker.NewCard("Qd"), poker.NewCard("Jh"), poker.NewCard("Tc"), poker.NewCard("9d"),
			},
			knownCards:  []poker.Card{},
			expectedMin: 99.0, // Flush should win
			expectedMax: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equity, err := CalculateStudHighEquity(tt.yourHand, tt.opponentHand, tt.knownCards, 10000)
			if err != nil {
				t.Fatalf("Error calculating equity: %v", err)
			}

			if equity < tt.expectedMin || equity > tt.expectedMax {
				t.Errorf("Expected equity between %.1f%% and %.1f%%, got %.1f%%",
					tt.expectedMin, tt.expectedMax, equity)
			}
		})
	}
}

func TestCalculateStud8Equity(t *testing.T) {
	tests := []struct {
		name             string
		yourHand         []poker.Card
		opponentHand     []poker.Card
		knownCards       []poker.Card
		expectedHighMin  float64
		expectedHighMax  float64
		expectedLowMin   float64
		expectedLowMax   float64
		expectedScoopMin float64
		expectedScoopMax float64
	}{
		{
			name: "Wheel potential vs high cards",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("4c"), poker.NewCard("5s"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ks"), poker.NewCard("Kd"), poker.NewCard("Qh"),
				poker.NewCard("Jc"), poker.NewCard("Ts"),
			},
			knownCards:       []poker.Card{},
			expectedHighMin:  60.0, // Straight should be favorite for high
			expectedHighMax:  80.0,
			expectedLowMin:   90.0, // Should usually make low
			expectedLowMax:   100.0,
			expectedScoopMin: 50.0, // Good chance to scoop
			expectedScoopMax: 80.0,
		},
		{
			name: "High hand only vs low draw",
			yourHand: []poker.Card{
				poker.NewCard("Ks"), poker.NewCard("Kd"), poker.NewCard("Qh"),
				poker.NewCard("Qc"), poker.NewCard("Js"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("6h"),
				poker.NewCard("7c"), poker.NewCard("9s"),
			},
			knownCards:       []poker.Card{},
			expectedHighMin:  85.0, // Two pair should be strong for high
			expectedHighMax:  95.0,
			expectedLowMin:   0.0,  // Can't make low
			expectedLowMax:   5.0,
			expectedScoopMin: 30.0, // Can scoop when opponent doesn't make low
			expectedScoopMax: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateStud8Equity(tt.yourHand, tt.opponentHand, tt.knownCards, 10000)
			if err != nil {
				t.Fatalf("Error calculating equity: %v", err)
			}

			if result.Stud8Result == nil {
				t.Fatal("Expected Stud8Result to be non-nil")
			}

			highEquity := result.Stud8Result.HighEquity
			lowEquity := result.Stud8Result.LowEquity
			scoopEquity := result.Stud8Result.ScoopEquity

			if highEquity < tt.expectedHighMin || highEquity > tt.expectedHighMax {
				t.Errorf("High equity: expected between %.1f%% and %.1f%%, got %.1f%%",
					tt.expectedHighMin, tt.expectedHighMax, highEquity)
			}

			if lowEquity < tt.expectedLowMin || lowEquity > tt.expectedLowMax {
				t.Errorf("Low equity: expected between %.1f%% and %.1f%%, got %.1f%%",
					tt.expectedLowMin, tt.expectedLowMax, lowEquity)
			}

			if scoopEquity < tt.expectedScoopMin || scoopEquity > tt.expectedScoopMax {
				t.Errorf("Scoop equity: expected between %.1f%% and %.1f%%, got %.1f%%",
					tt.expectedScoopMin, tt.expectedScoopMax, scoopEquity)
			}
		})
	}
}

func TestStudEquityValidation(t *testing.T) {
	tests := []struct {
		name         string
		yourHand     []poker.Card
		opponentHand []poker.Card
		expectError  bool
	}{
		{
			name:         "Too few cards in hand",
			yourHand:     []poker.Card{poker.NewCard("As"), poker.NewCard("Kd")},
			opponentHand: []poker.Card{poker.NewCard("Qs"), poker.NewCard("Qd"), poker.NewCard("Qh")},
			expectError:  true,
		},
		{
			name: "Too many cards in hand",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("Kd"), poker.NewCard("Qh"),
				poker.NewCard("Js"), poker.NewCard("Td"), poker.NewCard("9h"),
				poker.NewCard("8s"), poker.NewCard("7d"),
			},
			opponentHand: []poker.Card{poker.NewCard("2s"), poker.NewCard("2d"), poker.NewCard("2h")},
			expectError:  true,
		},
		{
			name:         "Duplicate cards",
			yourHand:     []poker.Card{poker.NewCard("As"), poker.NewCard("As"), poker.NewCard("Kd")},
			opponentHand: []poker.Card{poker.NewCard("Qs"), poker.NewCard("Qd"), poker.NewCard("Qh")},
			expectError:  true,
		},
		{
			name:         "Valid configuration",
			yourHand:     []poker.Card{poker.NewCard("As"), poker.NewCard("Kd"), poker.NewCard("Qh")},
			opponentHand: []poker.Card{poker.NewCard("Js"), poker.NewCard("Td"), poker.NewCard("9h")},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateStudEquity(tt.yourHand, tt.opponentHand, []poker.Card{}, StudRazz, 100)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestEquityConvergence(t *testing.T) {
	// Test that more iterations lead to more stable results
	yourHand := []poker.Card{
		poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"), poker.NewCard("4c"),
	}
	opponentHand := []poker.Card{
		poker.NewCard("5s"), poker.NewCard("6d"), poker.NewCard("7h"), poker.NewCard("8c"),
	}

	var lastEquity float64
	iterations := []int{100, 1000, 10000}

	for _, iter := range iterations {
		equity, err := CalculateRazzEquity(yourHand, opponentHand, []poker.Card{}, iter)
		if err != nil {
			t.Fatalf("Error calculating equity: %v", err)
		}

		if iter > 100 {
			// As iterations increase, the difference should generally decrease
			diff := math.Abs(equity - lastEquity)
			if diff > 10.0 { // Allow for some variance, but not too much
				t.Errorf("Equity variance too high between %d and previous iterations: %.2f%%",
					iter, diff)
			}
		}

		lastEquity = equity
	}
}