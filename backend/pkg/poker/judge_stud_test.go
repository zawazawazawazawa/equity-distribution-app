package poker

import (
	"testing"

	"github.com/chehsunliu/poker"
)

func TestJudgeWinnerRazz(t *testing.T) {
	tests := []struct {
		name         string
		yourHand     []poker.Card
		opponentHand []poker.Card
		expected     string
	}{
		{
			name: "Wheel beats 6-high",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("4c"), poker.NewCard("5s"), poker.NewCard("Kh"), poker.NewCard("Qd"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ad"), poker.NewCard("2h"), poker.NewCard("3s"),
				poker.NewCard("4d"), poker.NewCard("6c"), poker.NewCard("Ks"), poker.NewCard("Qh"),
			},
			expected: "yourHand",
		},
		{
			name: "Same Razz hands tie",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("4c"), poker.NewCard("5s"), poker.NewCard("Kh"), poker.NewCard("Qd"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ad"), poker.NewCard("2h"), poker.NewCard("3s"),
				poker.NewCard("4d"), poker.NewCard("5c"), poker.NewCard("Ks"), poker.NewCard("Qh"),
			},
			expected: "tie",
		},
		{
			name: "7-high loses to 6-high",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("4c"), poker.NewCard("7s"), poker.NewCard("Kh"), poker.NewCard("Qd"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ad"), poker.NewCard("2h"), poker.NewCard("3s"),
				poker.NewCard("4d"), poker.NewCard("6c"), poker.NewCard("Ks"), poker.NewCard("Qh"),
			},
			expected: "opponentHand",
		},
		{
			name: "Pairs don't help in Razz",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("Ah"), poker.NewCard("2d"),
				poker.NewCard("2h"), poker.NewCard("3c"), poker.NewCard("4s"), poker.NewCard("5d"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("6s"), poker.NewCard("7d"), poker.NewCard("8h"),
				poker.NewCard("9c"), poker.NewCard("Ts"), poker.NewCard("Jd"), poker.NewCard("Qh"),
			},
			expected: "yourHand", // A-2-3-4-5 beats 6-7-8-9-T
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JudgeWinnerRazz(tt.yourHand, tt.opponentHand)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestJudgeWinnerStudHigh(t *testing.T) {
	tests := []struct {
		name         string
		yourHand     []poker.Card
		opponentHand []poker.Card
		expected     string
	}{
		{
			name: "Flush beats straight",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("Ks"), poker.NewCard("Qs"),
				poker.NewCard("Js"), poker.NewCard("9s"), poker.NewCard("2d"), poker.NewCard("3h"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Kd"), poker.NewCard("Qh"), poker.NewCard("Jc"),
				poker.NewCard("Ts"), poker.NewCard("9d"), poker.NewCard("2c"), poker.NewCard("3s"),
			},
			expected: "yourHand",
		},
		{
			name: "Full house beats flush",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("Ad"), poker.NewCard("Ah"),
				poker.NewCard("Ks"), poker.NewCard("Kd"), poker.NewCard("2c"), poker.NewCard("3h"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("9s"), poker.NewCard("7s"), poker.NewCard("5s"),
				poker.NewCard("3s"), poker.NewCard("2s"), poker.NewCard("Kh"), poker.NewCard("Qd"),
			},
			expected: "yourHand",
		},
		{
			name: "Same hands tie",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("Kd"), poker.NewCard("Qh"),
				poker.NewCard("Jc"), poker.NewCard("Ts"), poker.NewCard("2d"), poker.NewCard("3h"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ad"), poker.NewCard("Kh"), poker.NewCard("Qc"),
				poker.NewCard("Js"), poker.NewCard("Td"), poker.NewCard("2c"), poker.NewCard("3s"),
			},
			expected: "tie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JudgeWinnerStudHigh(tt.yourHand, tt.opponentHand)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestJudgeWinnerStud8(t *testing.T) {
	tests := []struct {
		name            string
		yourHand        []poker.Card
		opponentHand    []poker.Card
		expectedHigh    string
		expectedLow     string
	}{
		{
			name: "Both qualify for low, different winners",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("4c"), poker.NewCard("5s"), poker.NewCard("Kh"), poker.NewCard("Kd"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ad"), poker.NewCard("Ah"), poker.NewCard("Ac"),
				poker.NewCard("8s"), poker.NewCard("7d"), poker.NewCard("6h"), poker.NewCard("2c"),
			},
			expectedHigh: "yourHand",      // Straight (wheel) beats three aces
			expectedLow:  "yourHand",      // Wheel beats A-2-6-7-8
		},
		{
			name: "Only one qualifies for low",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("4c"), poker.NewCard("5s"), poker.NewCard("6h"), poker.NewCard("7d"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Kd"), poker.NewCard("Kh"), poker.NewCard("Qc"),
				poker.NewCard("Qs"), poker.NewCard("Jd"), poker.NewCard("Th"), poker.NewCard("9c"),
			},
			expectedHigh: "opponentHand", // Two pair beats straight (for high)
			expectedLow:  "yourHand",      // Only your hand qualifies
		},
		{
			name: "No qualifying low",
			yourHand: []poker.Card{
				poker.NewCard("Ks"), poker.NewCard("Kd"), poker.NewCard("Kh"),
				poker.NewCard("Qc"), poker.NewCard("Qs"), poker.NewCard("Jd"), poker.NewCard("Th"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ad"), poker.NewCard("Ah"), poker.NewCard("9c"),
				poker.NewCard("9s"), poker.NewCard("9d"), poker.NewCard("Jh"), poker.NewCard("Tc"),
			},
			expectedHigh: "yourHand",      // Full house (KKK-QQ) beats full house (999-AA)
			expectedLow:  "none",          // No qualifying low
		},
		{
			name: "Scoop - same player wins both",
			yourHand: []poker.Card{
				poker.NewCard("As"), poker.NewCard("2d"), poker.NewCard("3h"),
				poker.NewCard("4c"), poker.NewCard("5s"), poker.NewCard("6h"), poker.NewCard("7d"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ad"), poker.NewCard("2h"), poker.NewCard("3s"),
				poker.NewCard("4d"), poker.NewCard("8c"), poker.NewCard("Ks"), poker.NewCard("Qh"),
			},
			expectedHigh: "yourHand", // Straight beats A-high
			expectedLow:  "yourHand", // 7-high beats 8-high
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			highWinner, lowWinner := JudgeWinnerStud8(tt.yourHand, tt.opponentHand)
			if highWinner != tt.expectedHigh {
				t.Errorf("High: Expected %s, got %s", tt.expectedHigh, highWinner)
			}
			if lowWinner != tt.expectedLow {
				t.Errorf("Low: Expected %s, got %s", tt.expectedLow, lowWinner)
			}
		})
	}
}

func TestJudgeWinnerAutoDetectStud(t *testing.T) {
	// Test that JudgeWinner correctly detects 7-card stud
	yourHand := []poker.Card{
		poker.NewCard("As"), poker.NewCard("Ks"), poker.NewCard("Qs"),
		poker.NewCard("Js"), poker.NewCard("Ts"), poker.NewCard("2d"), poker.NewCard("3h"),
	}
	opponentHand := []poker.Card{
		poker.NewCard("Ad"), poker.NewCard("Ah"), poker.NewCard("Ac"),
		poker.NewCard("Kd"), poker.NewCard("Kh"), poker.NewCard("2c"), poker.NewCard("3s"),
	}
	board := []poker.Card{} // No board for stud

	result := JudgeWinner(yourHand, opponentHand, board)
	expected := "yourHand" // Royal flush beats full house

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}