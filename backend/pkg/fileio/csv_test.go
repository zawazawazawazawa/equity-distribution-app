package fileio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRangeFromCSV(t *testing.T) {
	// テストケース1: 正常なCSVファイルの読み込み
	t.Run("Load valid CSV file", func(t *testing.T) {
		// テスト用の一時ファイルを作成
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "test_range.csv")

		// テスト用のCSVデータ
		csvContent := "AA,KK,QQ\nAKs,AQs,AJs\nAKo@50,AQo@40,AJo@30"

		// 一時ファイルに書き込み
		err := os.WriteFile(tempFile, []byte(csvContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test CSV file: %v", err)
		}

		// 関数の実行
		result, err := LoadRangeFromCSV(tempFile)

		// エラーがないことを確認
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// 結果の検証
		expectedHands := []string{"AA", "KK", "QQ", "AKs", "AQs", "AJs", "AKo", "AQo", "AJo"}
		resultHands := strings.Split(result, ",")

		if len(resultHands) != len(expectedHands) {
			t.Errorf("Expected %d hands, got %d", len(expectedHands), len(resultHands))
		}

		// 各ハンドが含まれているか確認
		for _, expectedHand := range expectedHands {
			found := false
			for _, resultHand := range resultHands {
				if resultHand == expectedHand {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected hand %s not found in result", expectedHand)
			}
		}
	})

	// テストケース2: 存在しないファイルの読み込み
	t.Run("Load non-existent file", func(t *testing.T) {
		// 存在しないファイルパス
		nonExistentFile := "/path/to/non/existent/file.csv"

		// 関数の実行
		result, err := LoadRangeFromCSV(nonExistentFile)

		// エラーが発生することを確認
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}

		// 結果が空文字列であることを確認
		if result != "" {
			t.Errorf("Expected empty result, got: %s", result)
		}
	})

	// テストケース3: 空のCSVファイルの読み込み
	t.Run("Load empty CSV file", func(t *testing.T) {
		// テスト用の一時ファイルを作成
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "empty.csv")

		// 空のファイルを作成
		err := os.WriteFile(tempFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create empty CSV file: %v", err)
		}

		// 関数の実行
		result, err := LoadRangeFromCSV(tempFile)

		// エラーがないことを確認
		if err != nil {
			t.Errorf("Expected no error for empty file, got: %v", err)
		}

		// 結果が空文字列であることを確認
		if result != "" {
			t.Errorf("Expected empty result for empty file, got: %s", result)
		}
	})
}

func TestLoadOpponentRangeFromPreset(t *testing.T) {
	// テスト用のデータディレクトリを作成
	tempDir := t.TempDir()
	baseDir := filepath.Join(tempDir, "six_handed_100bb_midrake")

	// 必要なディレクトリ構造を作成
	srDir := filepath.Join(baseDir, "srp")
	bpDir := filepath.Join(baseDir, "3bp")

	err := os.MkdirAll(srDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create srp directory: %v", err)
	}

	err = os.MkdirAll(bpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create 3bp directory: %v", err)
	}

	// テスト用のCSVファイルを作成
	testFiles := map[string]string{
		filepath.Join(srDir, "bb_call_vs_utg.csv"):  "AA,KK,QQ",
		filepath.Join(srDir, "bb_call_vs_btn.csv"):  "AA,KK",
		filepath.Join(srDir, "btn_call_vs_utg.csv"): "AA,AKs",
		filepath.Join(bpDir, "utg_call_vs_bb.csv"):  "KK,QQ",
		filepath.Join(bpDir, "utg_call_vs_btn.csv"): "KK,AKs",
		filepath.Join(bpDir, "btn_call_vs_bb.csv"):  "QQ,AKs",
	}

	for filePath, content := range testFiles {
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filePath, err)
		}
	}

	// テストケース1: 有効なプリセット
	t.Run("Valid presets", func(t *testing.T) {
		testCases := []struct {
			preset         string
			expectedResult string
		}{
			{"SRP BB call vs UTG open", "AA,KK,QQ"},
			{"SRP BB call vs BTN open", "AA,KK"},
			{"SRP BTN call vs UTG open", "AA,AKs"},
			{"3BP UTG call vs BB 3bet", "KK,QQ"},
			{"3BP UTG call vs BTN 3bet", "KK,AKs"},
			{"3BP BTN call vs BB 3bet", "QQ,AKs"},
		}

		for _, tc := range testCases {
			result, err := LoadOpponentRangeFromPreset(tc.preset, tempDir)

			if err != nil {
				t.Errorf("Expected no error for preset %s, got: %v", tc.preset, err)
			}

			if result != tc.expectedResult {
				t.Errorf("For preset %s, expected %s, got: %s", tc.preset, tc.expectedResult, result)
			}
		}
	})

	// テストケース2: 無効なプリセット
	t.Run("Invalid preset", func(t *testing.T) {
		invalidPreset := "Invalid preset name"

		result, err := LoadOpponentRangeFromPreset(invalidPreset, tempDir)

		if err == nil {
			t.Errorf("Expected error for invalid preset, got nil")
		}

		if result != "" {
			t.Errorf("Expected empty result for invalid preset, got: %s", result)
		}
	})
}

func TestLoadAggressorRangeFromPreset(t *testing.T) {
	// テスト用のデータディレクトリを作成
	tempDir := t.TempDir()
	baseDir := filepath.Join(tempDir, "six_handed_100bb_midrake")

	// 必要なディレクトリ構造を作成
	srDir := filepath.Join(baseDir, "srp")
	bpDir := filepath.Join(baseDir, "3bp")

	err := os.MkdirAll(srDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create srp directory: %v", err)
	}

	err = os.MkdirAll(bpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create 3bp directory: %v", err)
	}

	// テスト用のCSVファイルを作成
	testFiles := map[string]string{
		filepath.Join(srDir, "utg_open.csv"):      "AA,KK,QQ,JJ",
		filepath.Join(srDir, "btn_open.csv"):      "AA,KK,QQ,JJ,TT",
		filepath.Join(bpDir, "bb_3b_vs_utg.csv"):  "AA,KK,QQ",
		filepath.Join(bpDir, "bb_3b_vs_btn.csv"):  "AA,KK,QQ,JJ",
		filepath.Join(bpDir, "btn_3b_vs_utg.csv"): "AA,KK",
	}

	for filePath, content := range testFiles {
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filePath, err)
		}
	}

	// テストケース1: 有効なプリセット
	t.Run("Valid presets", func(t *testing.T) {
		testCases := []struct {
			preset         string
			expectedResult string
		}{
			{"SRP BB call vs UTG open", "AA,KK,QQ,JJ"},    // UTGがアグレッサー
			{"SRP BB call vs BTN open", "AA,KK,QQ,JJ,TT"}, // BTNがアグレッサー
			{"SRP BTN call vs UTG open", "AA,KK,QQ,JJ"},   // UTGがアグレッサー
			{"3BP UTG call vs BB 3bet", "AA,KK,QQ"},       // BBがアグレッサー
			{"3BP UTG call vs BTN 3bet", "AA,KK"},         // BTNがアグレッサー
			{"3BP BTN call vs BB 3bet", "AA,KK,QQ,JJ"},    // BBがアグレッサー
		}

		for _, tc := range testCases {
			result, err := LoadAggressorRangeFromPreset(tc.preset, tempDir)

			if err != nil {
				t.Errorf("Expected no error for preset %s, got: %v", tc.preset, err)
			}

			if result != tc.expectedResult {
				t.Errorf("For preset %s, expected %s, got: %s", tc.preset, tc.expectedResult, result)
			}
		}
	})

	// テストケース2: 無効なプリセット
	t.Run("Invalid preset", func(t *testing.T) {
		invalidPreset := "Invalid preset name"

		result, err := LoadAggressorRangeFromPreset(invalidPreset, tempDir)

		if err == nil {
			t.Errorf("Expected error for invalid preset, got nil")
		}

		if result != "" {
			t.Errorf("Expected empty result for invalid preset, got: %s", result)
		}
	})
}
