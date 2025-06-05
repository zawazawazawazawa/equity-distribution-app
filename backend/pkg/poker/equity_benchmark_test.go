package poker

import (
	"testing"
	"time"

	"github.com/chehsunliu/poker"
)

func TestEquityCalculationPerformance(t *testing.T) {
	// テスト用のハンドとボードを設定
	yourHand := []poker.Card{
		poker.NewCard("As"),
		poker.NewCard("Ad"),
		poker.NewCard("Kh"),
		poker.NewCard("Kd"),
	}

	// 小さめのレンジでテスト（10ハンド）
	opponentHands := [][]poker.Card{
		{poker.NewCard("Qs"), poker.NewCard("Qd"), poker.NewCard("Jh"), poker.NewCard("Jd")},
		{poker.NewCard("Ts"), poker.NewCard("Td"), poker.NewCard("9h"), poker.NewCard("9d")},
		{poker.NewCard("8s"), poker.NewCard("8d"), poker.NewCard("7h"), poker.NewCard("7d")},
		{poker.NewCard("6s"), poker.NewCard("6d"), poker.NewCard("5h"), poker.NewCard("5d")},
		{poker.NewCard("4s"), poker.NewCard("4d"), poker.NewCard("3h"), poker.NewCard("3d")},
		{poker.NewCard("Ac"), poker.NewCard("Kc"), poker.NewCard("Qc"), poker.NewCard("Jc")},
		{poker.NewCard("9c"), poker.NewCard("8c"), poker.NewCard("7c"), poker.NewCard("6c")},
		{poker.NewCard("Ah"), poker.NewCard("Th"), poker.NewCard("9s"), poker.NewCard("8h")},
		{poker.NewCard("Ks"), poker.NewCard("Js"), poker.NewCard("Ts"), poker.NewCard("9h")},
		{poker.NewCard("Qh"), poker.NewCard("Jh"), poker.NewCard("Tc"), poker.NewCard("9c")},
	}

	board := []poker.Card{
		poker.NewCard("2c"),
		poker.NewCard("7s"),
		poker.NewCard("Js"),
	}

	t.Run("Exhaustive", func(t *testing.T) {
		start := time.Now()
		result, err := CalculateHandVsRangeEquityParallel(yourHand, opponentHands, board)
		duration := time.Since(start)
		
		if err != nil {
			t.Fatalf("Error in exhaustive calculation: %v", err)
		}
		
		t.Logf("Exhaustive calculation took: %v", duration)
		t.Logf("Number of results: %d", len(result))
		
		// 平均エクイティを計算
		var totalEquity float64
		for _, equity := range result {
			totalEquity += equity
		}
		avgEquity := totalEquity / float64(len(result))
		t.Logf("Average equity: %.2f%%", avgEquity)
	})

	t.Run("MonteCarlo", func(t *testing.T) {
		start := time.Now()
		result, err := CalculateHandVsRangeEquityMonteCarloParallel(yourHand, opponentHands, board)
		duration := time.Since(start)
		
		if err != nil {
			t.Fatalf("Error in Monte Carlo calculation: %v", err)
		}
		
		t.Logf("Monte Carlo calculation took: %v", duration)
		t.Logf("Number of results: %d", len(result))
		
		// 平均エクイティを計算
		var totalEquity float64
		for _, equity := range result {
			totalEquity += equity
		}
		avgEquity := totalEquity / float64(len(result))
		t.Logf("Average equity: %.2f%%", avgEquity)
	})
}