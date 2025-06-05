package handrank

import (
	"testing"
)

func TestEval5WithTable(t *testing.T) {
	// テストケース: ロイヤルフラッシュ (As Ks Qs Js Ts)
	// スペードのロイヤルフラッシュ
	royalFlush := [5]int{
		12*4 + 0, // As (A=12, spades=0)
		11*4 + 0, // Ks (K=11, spades=0)
		10*4 + 0, // Qs (Q=10, spades=0)
		9*4 + 0,  // Js (J=9, spades=0)
		8*4 + 0,  // Ts (T=8, spades=0)
	}

	rank, err := Eval5WithTable(royalFlush)
	if err != nil {
		t.Fatalf("Eval5WithTable failed: %v", err)
	}

	// ロイヤルフラッシュは高いランクであることを確認
	// 実際の値をログで確認
	if rank <= 0 {
		t.Errorf("Expected positive rank for royal flush, got %d", rank)
	}

	// テストケース: ハイカード (2s 4h 6d 8c Ts)
	highCard := [5]int{
		0*4 + 0, // 2s (2=0, spades=0)
		2*4 + 1, // 4h (4=2, hearts=1)
		4*4 + 2, // 6d (6=4, diamonds=2)
		6*4 + 3, // 8c (8=6, clubs=3)
		8*4 + 0, // Ts (T=8, spades=0)
	}

	rank2, err := Eval5WithTable(highCard)
	if err != nil {
		t.Fatalf("Eval5WithTable failed: %v", err)
	}

	// ハイカードは低いランク（実際の値を確認）
	if rank2 <= 0 {
		t.Errorf("Expected positive rank for high card, got %d", rank2)
	}

	// ロイヤルフラッシュの方が高いランクであることを確認
	if rank <= rank2 {
		t.Errorf("Royal flush rank (%d) should be higher than high card rank (%d)", rank, rank2)
	}

	t.Logf("Royal flush rank: %d", rank)
	t.Logf("High card rank: %d", rank2)
}

func TestEval5WithTableVsEvalRank(t *testing.T) {
	// 複数のハンドで新旧の実装を比較
	testHands := [][5]int{
		// ペア (As As 2h 3d 4c)
		{12*4 + 0, 12*4 + 1, 0*4 + 1, 1*4 + 2, 2*4 + 3},
		// ツーペア (As As 2h 2d 3c)
		{12*4 + 0, 12*4 + 1, 0*4 + 1, 0*4 + 2, 1*4 + 3},
		// スリーカード (As As As 2h 3d)
		{12*4 + 0, 12*4 + 1, 12*4 + 2, 0*4 + 1, 1*4 + 2},
		// ストレート (As 2h 3d 4c 5s)
		{12*4 + 0, 0*4 + 1, 1*4 + 2, 2*4 + 3, 3*4 + 0},
	}

	for i, hand := range testHands {
		newRank, err := Eval5WithTable(hand)
		if err != nil {
			t.Fatalf("Test %d: Eval5WithTable failed: %v", i, err)
		}

		oldRank := EvalRank(hand)

		if int16(newRank) != oldRank {
			t.Errorf("Test %d: Rank mismatch - Eval5WithTable: %d, EvalRank: %d", i, newRank, oldRank)
		}

		t.Logf("Test %d: Hand %v - Rank: %d", i, hand, newRank)
	}
}

func BenchmarkEval5WithTable(b *testing.B) {
	hand := [5]int{12*4 + 0, 11*4 + 0, 10*4 + 0, 9*4 + 0, 8*4 + 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Eval5WithTable(hand)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEvalRank(b *testing.B) {
	hand := [5]int{12*4 + 0, 11*4 + 0, 10*4 + 0, 9*4 + 0, 8*4 + 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		EvalRank(hand)
	}
}
