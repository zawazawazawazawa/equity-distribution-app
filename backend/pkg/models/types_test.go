package models

import (
	"encoding/json"
	"testing"
)

func TestFlopEquities(t *testing.T) {
	// 構造体の初期化と値の検証
	flop := "AhKdQc"
	equities := map[string]float64{
		"AcAd_KsKh": 65.5,
		"JsTs_9h8h": 34.5,
	}

	flopEquities := FlopEquities{
		Flop:     flop,
		Equities: equities,
	}

	// 値の検証
	if flopEquities.Flop != flop {
		t.Errorf("Expected Flop to be %s, got %s", flop, flopEquities.Flop)
	}

	if len(flopEquities.Equities) != len(equities) {
		t.Errorf("Expected Equities to have %d entries, got %d", len(equities), len(flopEquities.Equities))
	}

	for hand, equity := range equities {
		if flopEquities.Equities[hand] != equity {
			t.Errorf("Expected equity for %s to be %.2f, got %.2f", hand, equity, flopEquities.Equities[hand])
		}
	}

	// JSONシリアライズ/デシリアライズのテスト
	jsonData, err := json.Marshal(flopEquities)
	if err != nil {
		t.Fatalf("Failed to marshal FlopEquities: %v", err)
	}

	var unmarshalled FlopEquities
	err = json.Unmarshal(jsonData, &unmarshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal FlopEquities: %v", err)
	}

	if unmarshalled.Flop != flop {
		t.Errorf("After JSON roundtrip, expected Flop to be %s, got %s", flop, unmarshalled.Flop)
	}

	if len(unmarshalled.Equities) != len(equities) {
		t.Errorf("After JSON roundtrip, expected Equities to have %d entries, got %d", len(equities), len(unmarshalled.Equities))
	}

	for hand, equity := range equities {
		if unmarshalled.Equities[hand] != equity {
			t.Errorf("After JSON roundtrip, expected equity for %s to be %.2f, got %.2f", hand, equity, unmarshalled.Equities[hand])
		}
	}
}

func TestHandVsRangeResult(t *testing.T) {
	// 構造体の初期化と値の検証
	opponentHand := "AcAd"
	equity := 65.5

	result := HandVsRangeResult{
		OpponentHand: opponentHand,
		Equity:       equity,
	}

	// 値の検証
	if result.OpponentHand != opponentHand {
		t.Errorf("Expected OpponentHand to be %s, got %s", opponentHand, result.OpponentHand)
	}

	if result.Equity != equity {
		t.Errorf("Expected Equity to be %.2f, got %.2f", equity, result.Equity)
	}

	// JSONシリアライズ/デシリアライズのテスト
	jsonData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal HandVsRangeResult: %v", err)
	}

	var unmarshalled HandVsRangeResult
	err = json.Unmarshal(jsonData, &unmarshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal HandVsRangeResult: %v", err)
	}

	if unmarshalled.OpponentHand != opponentHand {
		t.Errorf("After JSON roundtrip, expected OpponentHand to be %s, got %s", opponentHand, unmarshalled.OpponentHand)
	}

	if unmarshalled.Equity != equity {
		t.Errorf("After JSON roundtrip, expected Equity to be %.2f, got %.2f", equity, unmarshalled.Equity)
	}
}
