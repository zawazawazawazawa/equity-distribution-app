package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chehsunliu/poker"

	"equity-distribution-backend/pkg/models"
)

func BenchmarkHandleEquityCalculation(b *testing.B) {
	reqBody := `{
		"yourHands": "ACADAHAS@100,ACADAH2C@100,ACADAH2D@100",
		"opponentsHands": "KsKh6s5h"
	}`
	req, err := http.NewRequest("POST", "/calculate-equity", strings.NewReader(reqBody))
	if err != nil {
		b.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleEquityCalculation)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

// TestHandleEquityCalculation は handleEquityCalculation 関数の正常系をテストします
func TestHandleEquityCalculation(t *testing.T) {
	// テストケース
	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name: "正常なリクエスト",
			requestBody: `{
				"yourHands": "ACADAHAS@100,ACADAH2C@100,ACADAH2D@100",
				"opponentsHands": "KsKh6s5h",
				"flopCards": ["As", "Kd", "2c"]
			}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/calculate-equity", strings.NewReader(tc.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleEquityCalculation)
			handler.ServeHTTP(rr, req)

			// ステータスコードの確認
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}

			// 正常系の場合はレスポンスの形式を確認
			if tc.expectedStatus == http.StatusOK {
				var response [][]interface{}
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Could not parse response body: %v", err)
				}
			}
		})
	}
}

// TestHandleEquityCalculationErrors は handleEquityCalculation 関数の異常系をテストします
func TestHandleEquityCalculationErrors(t *testing.T) {
	// テストケース
	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "無効なJSONフォーマット",
			requestBody:    `{"yourHands": "ACADAHAS@100", "opponentsHands": "KsKh6s5h", "flopCards": ["As", "Kd", "2c"`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "フロップカードが3枚ではない",
			requestBody: `{
				"yourHands": "ACADAHAS@100",
				"opponentsHands": "KsKh6s5h",
				"flopCards": ["As", "Kd"]
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "無効なハンド形式",
			requestBody: `{
				"yourHands": "INVALID",
				"opponentsHands": "KsKh6s5h",
				"flopCards": ["As", "Kd", "2c"]
			}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/calculate-equity", strings.NewReader(tc.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleEquityCalculation)
			handler.ServeHTTP(rr, req)

			// ステータスコードの確認
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}
		})
	}
}

// BenchmarkHandleHandVsRangeCalculation は handleHandVsRangeCalculation 関数のパフォーマンスを測定します
func BenchmarkHandleHandVsRangeCalculation(b *testing.B) {
	reqBody := `{
		"yourHand": "ACADAHAS@100",
		"selectedPreset": "SRP BB call vs UTG open",
		"flopCards": ["As", "Kd", "2c"]
	}`
	req, err := http.NewRequest("POST", "/calculate-hand-vs-range", strings.NewReader(reqBody))
	if err != nil {
		b.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleHandVsRangeCalculation)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

// TestHandleHandVsRangeCalculation は handleHandVsRangeCalculation 関数の正常系をテストします
func TestHandleHandVsRangeCalculation(t *testing.T) {
	// テストケース
	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name: "プリセットを使用した正常なリクエスト",
			requestBody: `{
				"yourHand": "ACADAHAS@100",
				"selectedPreset": "SRP BB call vs UTG open",
				"flopCards": ["As", "Kd", "2c"]
			}`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "opponentハンドを直接指定した正常なリクエスト",
			requestBody: `{
				"yourHand": "ACADAHAS@100",
				"opponentsHands": "KsKh6s5h,QsQh7s6h",
				"flopCards": ["As", "Kd", "2c"]
			}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/calculate-hand-vs-range", strings.NewReader(tc.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleHandVsRangeCalculation)
			handler.ServeHTTP(rr, req)

			// ステータスコードの確認
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}

			// 正常系の場合はレスポンスの形式を確認
			if tc.expectedStatus == http.StatusOK {
				var response []models.HandVsRangeResult
				err = json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Could not parse response body: %v", err)
				}
			}
		})
	}
}

// TestHandleHandVsRangeCalculationErrors は handleHandVsRangeCalculation 関数の異常系をテストします
func TestHandleHandVsRangeCalculationErrors(t *testing.T) {
	// テストケース
	testCases := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "無効なJSONフォーマット",
			requestBody:    `{"yourHand": "ACADAHAS@100", "selectedPreset": "SRP BB call vs UTG open", "flopCards": ["As", "Kd", "2c"`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "フロップカードが3枚ではない",
			requestBody: `{
				"yourHand": "ACADAHAS@100",
				"selectedPreset": "SRP BB call vs UTG open",
				"flopCards": ["As", "Kd"]
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "無効なハンド形式",
			requestBody: `{
				"yourHand": "INVALID",
				"selectedPreset": "SRP BB call vs UTG open",
				"flopCards": ["As", "Kd", "2c"]
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "プリセットもopponentハンドも指定されていない",
			requestBody: `{
				"yourHand": "ACADAHAS@100",
				"flopCards": ["As", "Kd", "2c"]
			}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "存在しないプリセット",
			requestBody: `{
				"yourHand": "ACADAHAS@100",
				"selectedPreset": "NON_EXISTENT_PRESET",
				"flopCards": ["As", "Kd", "2c"]
			}`,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/calculate-hand-vs-range", strings.NewReader(tc.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleHandVsRangeCalculation)
			handler.ServeHTTP(rr, req)

			// ステータスコードの確認
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}
		})
	}
}

// TestJudgeWinner は judgeWinner 関数をテストします
func TestJudgeWinner(t *testing.T) {
	// テストケース
	testCases := []struct {
		name         string
		yourHand     []poker.Card
		opponentHand []poker.Card
		board        []poker.Card
		expected     string
	}{
		{
			name: "あなたの手札が勝つ場合",
			yourHand: []poker.Card{
				poker.NewCard("Ah"),
				poker.NewCard("Ad"),
				poker.NewCard("Kh"),
				poker.NewCard("Qh"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ks"),
				poker.NewCard("Kd"),
				poker.NewCard("Qd"),
				poker.NewCard("Jd"),
			},
			board: []poker.Card{
				poker.NewCard("2h"),
				poker.NewCard("3d"),
				poker.NewCard("7c"),
				poker.NewCard("4s"),
				poker.NewCard("5c"),
			},
			expected: "yourHand",
		},
		{
			name: "相手の手札が勝つ場合",
			yourHand: []poker.Card{
				poker.NewCard("Ks"),
				poker.NewCard("Kd"),
				poker.NewCard("Qd"),
				poker.NewCard("Jd"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ah"),
				poker.NewCard("Ad"),
				poker.NewCard("Kh"),
				poker.NewCard("Qh"),
			},
			board: []poker.Card{
				poker.NewCard("2h"),
				poker.NewCard("3d"),
				poker.NewCard("7c"),
				poker.NewCard("4s"),
				poker.NewCard("5c"),
			},
			expected: "opponentHand",
		},
		{
			name: "引き分けの場合",
			yourHand: []poker.Card{
				poker.NewCard("Ah"),
				poker.NewCard("Kh"),
				poker.NewCard("Qh"),
				poker.NewCard("Jh"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("As"),
				poker.NewCard("Ks"),
				poker.NewCard("Qs"),
				poker.NewCard("Js"),
			},
			board: []poker.Card{
				poker.NewCard("2h"),
				poker.NewCard("3d"),
				poker.NewCard("7c"),
				poker.NewCard("4s"),
				poker.NewCard("5c"),
			},
			expected: "tie",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := judgeWinner(tc.yourHand, tc.opponentHand, tc.board)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

// TestHasCardDuplicates は hasCardDuplicates 関数をテストします
func TestHasCardDuplicates(t *testing.T) {
	// テストケース
	testCases := []struct {
		name     string
		cards    [][]poker.Card
		expected bool
	}{
		{
			name: "重複がない場合",
			cards: [][]poker.Card{
				{
					poker.NewCard("Ah"),
					poker.NewCard("Kh"),
				},
				{
					poker.NewCard("Qh"),
					poker.NewCard("Jh"),
				},
			},
			expected: false,
		},
		{
			name: "重複がある場合",
			cards: [][]poker.Card{
				{
					poker.NewCard("Ah"),
					poker.NewCard("Kh"),
				},
				{
					poker.NewCard("Ah"), // 重複
					poker.NewCard("Qh"),
				},
			},
			expected: true,
		},
		{
			name: "3つの配列で重複がある場合",
			cards: [][]poker.Card{
				{
					poker.NewCard("Ah"),
					poker.NewCard("Kh"),
				},
				{
					poker.NewCard("Qh"),
					poker.NewCard("Jh"),
				},
				{
					poker.NewCard("Kh"), // 重複
					poker.NewCard("Th"),
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := hasCardDuplicates(tc.cards...)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestCalculateHandVsHandEquity は calculateHandVsHandEquity 関数をテストします
func TestCalculateHandVsHandEquity(t *testing.T) {
	// テストケース
	testCases := []struct {
		name         string
		yourHand     []poker.Card
		opponentHand []poker.Card
		board        []poker.Card
		expectError  bool
	}{
		{
			name: "正常なケース",
			yourHand: []poker.Card{
				poker.NewCard("Ah"),
				poker.NewCard("Ad"),
				poker.NewCard("Kh"),
				poker.NewCard("Qh"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ks"),
				poker.NewCard("Kd"),
				poker.NewCard("Qd"),
				poker.NewCard("Jd"),
			},
			board: []poker.Card{
				poker.NewCard("2h"),
				poker.NewCard("3d"),
				poker.NewCard("7c"),
			},
			expectError: false,
		},
		{
			name: "カードが重複するケース",
			yourHand: []poker.Card{
				poker.NewCard("Ah"),
				poker.NewCard("Ad"),
				poker.NewCard("Kh"),
				poker.NewCard("Qh"),
			},
			opponentHand: []poker.Card{
				poker.NewCard("Ah"), // 重複
				poker.NewCard("Kd"),
				poker.NewCard("Qd"),
				poker.NewCard("Jd"),
			},
			board: []poker.Card{
				poker.NewCard("2h"),
				poker.NewCard("3d"),
				poker.NewCard("7c"),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			equity, _ := calculateHandVsHandEquity(tc.yourHand, tc.opponentHand, tc.board, nil)
			if tc.expectError && equity != -1 {
				t.Errorf("expected error (equity = -1), got %f", equity)
			} else if !tc.expectError && equity == -1 {
				t.Errorf("expected valid equity, got error (equity = -1)")
			}
		})
	}
}

// TestGenerateBoardString は generateBoardString 関数をテストします
func TestGenerateBoardString(t *testing.T) {
	// テストケース
	testCases := []struct {
		name     string
		board    []poker.Card
		expected string
	}{
		{
			name: "3枚のボード",
			board: []poker.Card{
				poker.NewCard("Ah"),
				poker.NewCard("Kd"),
				poker.NewCard("2c"),
			},
			expected: "AhKd2c",
		},
		{
			name: "5枚のボード",
			board: []poker.Card{
				poker.NewCard("Ah"),
				poker.NewCard("Kd"),
				poker.NewCard("2c"),
				poker.NewCard("Jh"),
				poker.NewCard("Ts"),
			},
			expected: "AhKd2cJhTs",
		},
		{
			name:     "空のボード",
			board:    []poker.Card{},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := generateBoardString(tc.board)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

// TestGenerateHandCombination は generateHandCombination 関数をテストします
func TestGenerateHandCombination(t *testing.T) {
	// テストケース
	testCases := []struct {
		name        string
		heroHand    string
		villainHand string
		expected    string
	}{
		{
			name:        "通常のケース",
			heroHand:    "AhAdKhQh",
			villainHand: "KsKdQdJd",
			expected:    "AhAdKhQh_KsKdQdJd",
		},
		{
			name:        "アルファベット順で後のハンドが先に来る場合",
			heroHand:    "KsKdQdJd",
			villainHand: "AhAdKhQh",
			expected:    "AhAdKhQh_KsKdQdJd",
		},
		{
			name:        "同じハンドの場合",
			heroHand:    "AhAdKhQh",
			villainHand: "AhAdKhQh",
			expected:    "AhAdKhQh_AhAdKhQh",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := generateHandCombination(tc.heroHand, tc.villainHand)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}
