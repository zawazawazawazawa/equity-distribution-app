package fileio

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LoadRangeFromCSV loads a range from a CSV file
func LoadRangeFromCSV(filePath string) (string, error) {
	log.Printf("Loading range from file: %s", filePath)

	// CSVファイルを読み込む
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading CSV file: %v", err)
		return "", fmt.Errorf("failed to read CSV file: %v", err)
	}

	// CSVの内容をカンマ区切りの文字列に変換
	lines := strings.Split(string(content), "\n")
	var hands []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// CSVの各行からすべてのハンドを抽出
		parts := strings.Split(line, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			// @記号がある場合は、その前の部分だけを使用
			handParts := strings.Split(part, "@")
			hand := handParts[0]

			if hand != "" {
				hands = append(hands, hand)
			}
		}
	}

	// カンマ区切りの文字列に変換
	return strings.Join(hands, ","), nil
}

// LoadOpponentRangeFromPreset loads opponent range from CSV file based on preset name
func LoadOpponentRangeFromPreset(preset string, dataDir string) (string, error) {
	var filePath string
	var gameType string
	
	// プリセット名からゲームタイプを判定
	if strings.HasPrefix(preset, "PLO5") {
		gameType = "plo5"
	} else {
		gameType = "plo4"
	}
	
	baseDir := fmt.Sprintf("%s/%s/six_handed_100bb_midrake", dataDir, gameType)

	// プリセット値に基づいてファイルパスを決定
	switch preset {
	case "SRP BB call vs UTG open", "PLO5 SRP BB call vs UTG open":
		filePath = fmt.Sprintf("%s/srp/bb_call_vs_utg.csv", baseDir)
	case "SRP BB call vs BTN open", "PLO5 SRP BB call vs BTN open":
		filePath = fmt.Sprintf("%s/srp/bb_call_vs_btn.csv", baseDir)
	case "SRP BTN call vs UTG open", "PLO5 SRP BTN call vs UTG open":
		filePath = fmt.Sprintf("%s/srp/btn_call_vs_utg.csv", baseDir)
	case "3BP UTG call vs BB 3bet", "PLO5 3BP UTG call vs BB 3bet":
		filePath = fmt.Sprintf("%s/3bp/utg_call_vs_bb.csv", baseDir)
	case "3BP UTG call vs BTN 3bet", "PLO5 3BP UTG call vs BTN 3bet":
		filePath = fmt.Sprintf("%s/3bp/utg_call_vs_btn.csv", baseDir)
	case "3BP BTN call vs BB 3bet", "PLO5 3BP BTN call vs BB 3bet":
		filePath = fmt.Sprintf("%s/3bp/btn_call_vs_bb.csv", baseDir)
	default:
		return "", fmt.Errorf("unknown preset: %s", preset)
	}

	return LoadRangeFromCSV(filePath)
}

// LoadAggressorRangeFromPreset loads aggressor range from CSV file based on preset name
func LoadAggressorRangeFromPreset(preset string, dataDir string) (string, error) {
	var filePath string
	var gameType string
	
	// プリセット名からゲームタイプを判定
	if strings.HasPrefix(preset, "PLO5") {
		gameType = "plo5"
	} else {
		gameType = "plo4"
	}
	
	baseDir := fmt.Sprintf("%s/%s/six_handed_100bb_midrake", dataDir, gameType)

	// プリセット値に基づいてアグレッサー側のレンジファイルパスを決定
	switch preset {
	case "SRP BB call vs UTG open", "PLO5 SRP BB call vs UTG open":
		filePath = fmt.Sprintf("%s/srp/utg_open.csv", baseDir) // UTGがアグレッサー
	case "SRP BB call vs BTN open", "PLO5 SRP BB call vs BTN open":
		filePath = fmt.Sprintf("%s/srp/btn_open.csv", baseDir) // BTNがアグレッサー
	case "SRP BTN call vs UTG open", "PLO5 SRP BTN call vs UTG open":
		filePath = fmt.Sprintf("%s/srp/utg_open.csv", baseDir) // UTGがアグレッサー
	case "3BP UTG call vs BB 3bet", "PLO5 3BP UTG call vs BB 3bet":
		filePath = fmt.Sprintf("%s/3bp/bb_3b_vs_utg.csv", baseDir) // BBがアグレッサー
	case "3BP UTG call vs BTN 3bet", "PLO5 3BP UTG call vs BTN 3bet":
		filePath = fmt.Sprintf("%s/3bp/btn_3b_vs_utg.csv", baseDir) // BTNがアグレッサー
	case "3BP BTN call vs BB 3bet", "PLO5 3BP BTN call vs BB 3bet":
		filePath = fmt.Sprintf("%s/3bp/bb_3b_vs_btn.csv", baseDir) // BBがアグレッサー
	default:
		return "", fmt.Errorf("unknown preset: %s", preset)
	}

	return LoadRangeFromCSV(filePath)
}
