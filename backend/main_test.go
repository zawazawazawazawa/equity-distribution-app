package main

import (
	"testing"
)

func TestCalculateHandEquity(t *testing.T) {
	hand := "AsAh6s5h"
	opponentHandRange := []string{"KsKh6s5h", "QsQh6s5h", "JsJh6s5h"}
	board := "2c3d4h"

	equity := calculateHandEquity(hand, opponentHandRange, board)

	if equity <= 0 || equity > 1 {
		t.Errorf("Expected equity to be between 0 and 1, got %f", equity)
	}
}

func TestSimulatePLOEquity(t *testing.T) {
	hand := "AsAh6s5h"
	opponentHand := "KsKh6s5h"
	board := "2c3d4h"

	equity := simulatePLOEquity(hand, opponentHand, board)

	if equity <= 0 || equity > 1 {
		t.Errorf("Expected equity to be between 0 and 1, got %f", equity)
	}
}
