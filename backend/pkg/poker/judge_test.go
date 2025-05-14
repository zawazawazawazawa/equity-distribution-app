package poker

import (
	"testing"

	"github.com/chehsunliu/poker"
)

func TestJudgeWinner(t *testing.T) {
	// テストケース1: yourHandが勝つ場合
	t.Run("YourHand wins", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Kd"),
		}
		board := []poker.Card{
			poker.NewCard("Qc"),
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		result := JudgeWinner(yourHand, opponentHand, board)

		// 結果の検証
		if result != "yourHand" {
			t.Errorf("Expected 'yourHand' to win, got %s", result)
		}
	})

	// テストケース2: opponentHandが勝つ場合
	t.Run("OpponentHand wins", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Kd"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		board := []poker.Card{
			poker.NewCard("Qc"),
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		result := JudgeWinner(yourHand, opponentHand, board)

		// 結果の検証
		if result != "opponentHand" {
			t.Errorf("Expected 'opponentHand' to win, got %s", result)
		}
	})

	// テストケース3: 引き分けの場合
	t.Run("Tie", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Kd"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Ad"),
			poker.NewCard("Kh"),
		}
		board := []poker.Card{
			poker.NewCard("Qc"),
			poker.NewCard("Jd"),
			poker.NewCard("Ts"),
		}

		// 関数の実行
		result := JudgeWinner(yourHand, opponentHand, board)

		// 結果の検証
		if result != "tie" {
			t.Errorf("Expected 'tie', got %s", result)
		}
	})

	// テストケース4: フラッシュ vs ストレート
	t.Run("Flush vs Straight", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Kh"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("9d"),
			poker.NewCard("8c"),
		}
		board := []poker.Card{
			poker.NewCard("Qh"),
			poker.NewCard("Jh"),
			poker.NewCard("Th"),
		}

		// 関数の実行
		result := JudgeWinner(yourHand, opponentHand, board)

		// 結果の検証
		if result != "yourHand" {
			t.Errorf("Expected 'yourHand' to win with flush, got %s", result)
		}
	})

	// テストケース5: フルハウス vs フラッシュ
	t.Run("Full House vs Flush", func(t *testing.T) {
		// テストデータの準備
		yourHand := []poker.Card{
			poker.NewCard("Ah"),
			poker.NewCard("Ad"),
		}
		opponentHand := []poker.Card{
			poker.NewCard("Kh"),
			poker.NewCard("Qh"),
		}
		board := []poker.Card{
			poker.NewCard("Ac"),
			poker.NewCard("Kd"),
			poker.NewCard("Kc"),
		}

		// 関数の実行
		result := JudgeWinner(yourHand, opponentHand, board)

		// 結果の検証
		if result != "yourHand" {
			t.Errorf("Expected 'yourHand' to win with full house, got %s", result)
		}
	})
}
