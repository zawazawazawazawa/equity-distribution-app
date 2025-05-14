package poker

import (
	"testing"

	"github.com/chehsunliu/poker"
)

func TestGenerateBoardString(t *testing.T) {
	// テストケース1: 通常のボード
	t.Run("Normal board", func(t *testing.T) {
		// テストデータの準備
		board := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Kd"),
			poker.NewCard("Qc"),
		}

		// 関数の実行
		result := GenerateBoardString(board)

		// 結果の検証
		expected := "AhKdQc"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// テストケース2: 空のボード
	t.Run("Empty board", func(t *testing.T) {
		// テストデータの準備
		board := []poker.Card{}

		// 関数の実行
		result := GenerateBoardString(board)

		// 結果の検証
		expected := ""
		if result != expected {
			t.Errorf("Expected empty string, got %s", result)
		}
	})

	// テストケース3: フルボード（5枚）
	t.Run("Full board", func(t *testing.T) {
		// テストデータの準備
		board := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Kd"),
			poker.NewCard("Qc"),
			poker.NewCard("Js"),
			poker.NewCard("Td"),
		}

		// 関数の実行
		result := GenerateBoardString(board)

		// 結果の検証
		expected := "AhKdQcJsTd"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

func TestGenerateHandCombination(t *testing.T) {
	// テストケース1: 通常のハンド組み合わせ
	t.Run("Normal hand combination", func(t *testing.T) {
		// テストデータの準備
		heroHand := "AhAd"
		villainHand := "KhKd"

		// 関数の実行
		result := GenerateHandCombination(heroHand, villainHand)

		// 結果の検証
		expected := "AhAd_KhKd"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// テストケース2: 逆順のハンド組み合わせ（アルファベット順にソートされるべき）
	t.Run("Reversed hand combination", func(t *testing.T) {
		// テストデータの準備
		heroHand := "KhKd"
		villainHand := "AhAd"

		// 関数の実行
		result := GenerateHandCombination(heroHand, villainHand)

		// 結果の検証
		expected := "AhAd_KhKd"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	// テストケース3: 同じハンドの組み合わせ
	t.Run("Same hands", func(t *testing.T) {
		// テストデータの準備
		heroHand := "AhAd"
		villainHand := "AhAd"

		// 関数の実行
		result := GenerateHandCombination(heroHand, villainHand)

		// 結果の検証
		expected := "AhAd_AhAd"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

func TestHasCardDuplicates(t *testing.T) {
	// テストケース1: 重複がない場合
	t.Run("No duplicates", func(t *testing.T) {
		// テストデータの準備
		hand1 := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		hand2 := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Kd"),
		}
		board := []poker.Card{
			poker.NewCard("Qc"),
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		result := HasCardDuplicates(hand1, hand2, board)

		// 結果の検証
		if result {
			t.Error("Expected no duplicates, got duplicates")
		}
	})

	// テストケース2: ハンド間で重複がある場合
	t.Run("Duplicates between hands", func(t *testing.T) {
		// テストデータの準備
		hand1 := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		hand2 := []poker.Card{
			poker.NewCard("Ah"), // 重複するカード
			poker.NewCard("Kd"),
		}
		board := []poker.Card{
			poker.NewCard("Qc"),
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		result := HasCardDuplicates(hand1, hand2, board)

		// 結果の検証
		if !result {
			t.Error("Expected duplicates, got no duplicates")
		}
	})

	// テストケース3: ハンドとボードの間で重複がある場合
	t.Run("Duplicates between hand and board", func(t *testing.T) {
		// テストデータの準備
		hand1 := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		hand2 := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Kd"),
		}
		board := []poker.Card{
			poker.NewCard("Ad"), // 重複するカード
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		result := HasCardDuplicates(hand1, hand2, board)

		// 結果の検証
		if !result {
			t.Error("Expected duplicates, got no duplicates")
		}
	})

	// テストケース4: 単一の配列内で重複がある場合
	t.Run("Duplicates within single array", func(t *testing.T) {
		// テストデータの準備
		hand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ah"), // 同じカードが重複
		}

		// 関数の実行
		result := HasCardDuplicates(hand)

		// 結果の検証
		if !result {
			t.Error("Expected duplicates, got no duplicates")
		}
	})

	// テストケース5: 空の配列の場合
	t.Run("Empty arrays", func(t *testing.T) {
		// テストデータの準備
		emptyHand1 := []poker.Card{}
		emptyHand2 := []poker.Card{}

		// 関数の実行
		result := HasCardDuplicates(emptyHand1, emptyHand2)

		// 結果の検証
		if result {
			t.Error("Expected no duplicates for empty arrays, got duplicates")
		}
	})
}
