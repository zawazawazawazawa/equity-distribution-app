package poker

import (
	"reflect"
	"testing"

	"github.com/chehsunliu/poker"
)

func TestEvaluateRazzHand(t *testing.T) {
	tests := []struct {
		name     string
		cards    []poker.Card
		expected []int
	}{
		{
			name: "Perfect Razz hand (wheel)",
			cards: []poker.Card{
				poker.NewCard("As"),
				poker.NewCard("2d"),
				poker.NewCard("3h"),
				poker.NewCard("4c"),
				poker.NewCard("5s"),
			},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name: "Razz hand with pairs (pairs ignored)",
			cards: []poker.Card{
				poker.NewCard("As"),
				poker.NewCard("Ah"),
				poker.NewCard("2d"),
				poker.NewCard("3h"),
				poker.NewCard("4c"),
				poker.NewCard("5s"),
				poker.NewCard("5d"),
			},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name: "High cards only",
			cards: []poker.Card{
				poker.NewCard("Ks"),
				poker.NewCard("Qd"),
				poker.NewCard("Jh"),
				poker.NewCard("Tc"),
				poker.NewCard("9s"),
			},
			expected: []int{9, 10, 11, 12, 13},
		},
		{
			name: "Mixed hand",
			cards: []poker.Card{
				poker.NewCard("As"),
				poker.NewCard("5d"),
				poker.NewCard("7h"),
				poker.NewCard("Jc"),
				poker.NewCard("Ks"),
				poker.NewCard("2s"),
				poker.NewCard("8d"),
			},
			expected: []int{1, 2, 5, 7, 8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateBest5CardRazzHand(tt.cards)
			if !reflect.DeepEqual(result.cards, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result.cards)
			}
		})
	}
}

func TestEvaluateStud8LowHand(t *testing.T) {
	tests := []struct {
		name      string
		cards     []poker.Card
		expected  []int
		qualifies bool
	}{
		{
			name: "Perfect 8-or-better (wheel)",
			cards: []poker.Card{
				poker.NewCard("As"),
				poker.NewCard("2d"),
				poker.NewCard("3h"),
				poker.NewCard("4c"),
				poker.NewCard("5s"),
				poker.NewCard("Ks"),
				poker.NewCard("Qd"),
			},
			expected:  []int{1, 2, 3, 4, 5},
			qualifies: true,
		},
		{
			name: "8-high low",
			cards: []poker.Card{
				poker.NewCard("As"),
				poker.NewCard("2d"),
				poker.NewCard("3h"),
				poker.NewCard("5c"),
				poker.NewCard("8s"),
				poker.NewCard("Ks"),
				poker.NewCard("Jd"),
			},
			expected:  []int{1, 2, 3, 5, 8},
			qualifies: true,
		},
		{
			name: "No qualifying low (high cards only)",
			cards: []poker.Card{
				poker.NewCard("Ks"),
				poker.NewCard("Qd"),
				poker.NewCard("Jh"),
				poker.NewCard("Tc"),
				poker.NewCard("9s"),
				poker.NewCard("9d"),
				poker.NewCard("Th"),
			},
			expected:  []int{},
			qualifies: false,
		},
		{
			name: "Borderline - not enough low cards",
			cards: []poker.Card{
				poker.NewCard("As"),
				poker.NewCard("2d"),
				poker.NewCard("3h"),
				poker.NewCard("4c"),
				poker.NewCard("Ks"),
				poker.NewCard("Qd"),
				poker.NewCard("Jh"),
			},
			expected:  []int{},
			qualifies: false,
		},
		{
			name: "Low with pairs (pairs disqualify for 8-or-better)",
			cards: []poker.Card{
				poker.NewCard("As"),
				poker.NewCard("Ad"),
				poker.NewCard("2h"),
				poker.NewCard("3c"),
				poker.NewCard("4s"),
				poker.NewCard("5d"),
				poker.NewCard("6h"),
			},
			expected:  []int{1, 2, 3, 4, 5},
			qualifies: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluateBest5CardStud8LowHand(tt.cards)
			if result.qualifies != tt.qualifies {
				t.Errorf("Expected qualifies=%v, got %v", tt.qualifies, result.qualifies)
			}
			if !reflect.DeepEqual(result.cards, tt.expected) {
				t.Errorf("Expected cards %v, got %v", tt.expected, result.cards)
			}
		})
	}
}

func TestCompareRazzHands(t *testing.T) {
	tests := []struct {
		name     string
		hand1    RazzRank
		hand2    RazzRank
		expected int
	}{
		{
			name:     "Wheel beats 6-high",
			hand1:    RazzRank{cards: []int{1, 2, 3, 4, 5}},
			hand2:    RazzRank{cards: []int{1, 2, 3, 4, 6}},
			expected: -1,
		},
		{
			name:     "Same hands tie",
			hand1:    RazzRank{cards: []int{1, 2, 3, 4, 5}},
			hand2:    RazzRank{cards: []int{1, 2, 3, 4, 5}},
			expected: 0,
		},
		{
			name:     "7-high loses to 6-high",
			hand1:    RazzRank{cards: []int{1, 2, 3, 4, 7}},
			hand2:    RazzRank{cards: []int{1, 2, 3, 4, 6}},
			expected: 1,
		},
		{
			name:     "Compare by second card",
			hand1:    RazzRank{cards: []int{1, 3, 4, 5, 6}},
			hand2:    RazzRank{cards: []int{1, 2, 4, 5, 6}},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareRazzHands(tt.hand1, tt.hand2)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCompareStud8LowHands(t *testing.T) {
	tests := []struct {
		name     string
		hand1    Stud8LowRank
		hand2    Stud8LowRank
		expected int
	}{
		{
			name:     "Qualifying hand beats non-qualifying",
			hand1:    Stud8LowRank{cards: []int{1, 2, 3, 4, 5}, qualifies: true},
			hand2:    Stud8LowRank{cards: []int{}, qualifies: false},
			expected: -1,
		},
		{
			name:     "Two non-qualifying hands tie",
			hand1:    Stud8LowRank{cards: []int{}, qualifies: false},
			hand2:    Stud8LowRank{cards: []int{}, qualifies: false},
			expected: 0,
		},
		{
			name:     "Wheel beats 6-high",
			hand1:    Stud8LowRank{cards: []int{1, 2, 3, 4, 5}, qualifies: true},
			hand2:    Stud8LowRank{cards: []int{1, 2, 3, 4, 6}, qualifies: true},
			expected: -1,
		},
		{
			name:     "8-high loses to 7-high",
			hand1:    Stud8LowRank{cards: []int{1, 2, 3, 5, 8}, qualifies: true},
			hand2:    Stud8LowRank{cards: []int{1, 2, 3, 5, 7}, qualifies: true},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareStud8LowHands(tt.hand1, tt.hand2)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}