package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/chehsunliu/poker"
	"github.com/joho/godotenv"

	"equity-distribution-backend/pkg/db"
	"equity-distribution-backend/pkg/fileio"
	"equity-distribution-backend/pkg/models"
	pkrlib "equity-distribution-backend/pkg/poker"
)

// HandRange represents a user's hand
type HandRange struct {
	Hand string `json:"hand"`
}

// YourDynamoDBItem represents an item in DynamoDB
type YourDynamoDBItem struct {
	Equity float64 `json:"Equity"`
	// Add more fields as needed
}

// getDynamoDBClient initializes and returns a DynamoDB client
func getDynamoDBClient() *dynamodb.DynamoDB {
	config := db.Config{
		Region:   "us-east-1",
		Endpoint: "http://localhost:4566",
	}
	return db.GetDynamoDBClient(config)
}

// batchQueryDynamoDB retrieves all equity calculations for a specific flop
func batchQueryDynamoDB(flop string) (*models.FlopEquities, error) {
	svc := getDynamoDBClient()

	equities, err := db.BatchQueryDynamoDB(svc, "PloEquity", flop)
	if err != nil {
		return nil, err
	}

	// Create FlopEquities instance
	flopEquities := &models.FlopEquities{
		Flop:     flop,
		Equities: equities,
	}

	return flopEquities, nil
}

// insertDynamoDB inserts or updates an item in DynamoDB
func insertDynamoDB(flop string, handCombination string, equity float64) error {
	svc := getDynamoDBClient()
	return db.InsertDynamoDB(svc, "PloEquity", flop, handCombination, equity)
}

// generateBoardString creates a string representation of the board cards
func generateBoardString(board []poker.Card) string {
	return pkrlib.GenerateBoardString(board)
}

// generateHandCombination creates a unique combination key for the hands
func generateHandCombination(heroHand string, villainHand string) string {
	return pkrlib.GenerateHandCombination(heroHand, villainHand)
}

// hasCardDuplicates checks if there are any duplicate cards across all provided card arrays
func hasCardDuplicates(cards ...[]poker.Card) bool {
	return pkrlib.HasCardDuplicates(cards...)
}

// calculateHandVsHandEquity calculates the equity between two hands
// Returns equity value and whether it was a cache hit
func calculateHandVsHandEquity(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card, flopEquities *models.FlopEquities) (float64, bool) {
	// Convert hands to strings and concatenate
	heroHandStr := ""
	for _, card := range yourHand {
		heroHandStr += card.String()
	}
	villainHandStr := ""
	for _, card := range opponentHand {
		villainHandStr += card.String()
	}
	// Generate key for equity lookup
	handCombination := generateHandCombination(heroHandStr, villainHandStr)

	// Check if equity exists in memory
	if flopEquities != nil {
		if equity, exists := flopEquities.Equities[handCombination]; exists {
			log.Printf("Found cached equity for combination %s: %.2f", handCombination, equity)
			return equity, true
		}
	}

	// Calculate equity since it wasn't found in DynamoDB
	equity, _ := pkrlib.CalculateHandVsHandEquity(yourHand, opponentHand, board)
	return equity, false
}

// judgeWinner determines the winner between two hands
func judgeWinner(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) string {
	return pkrlib.JudgeWinner(yourHand, opponentHand, board)
}

// calculateHandVsRangeEquity は1つのハンドと複数のハンドのレンジに対してエクイティを計算
func calculateHandVsRangeEquity(yourHand []poker.Card, opponentHands [][]poker.Card, board []poker.Card) []models.HandVsRangeResult {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	boardStr := generateBoardString(board)

	// DynamoDBからフロップに関連するエクイティを取得
	_, err := batchQueryDynamoDB(boardStr)
	if err != nil {
		log.Printf("Error fetching flop equities: %v", err)
	}

	// サンプリングロジック（必要に応じて）
	handsToProcess := opponentHands
	if len(opponentHands) > 10000 {
		log.Printf("Sampling 10000 hands from %d opponent hands", len(opponentHands))
		shuffledOpponentHands := make([][]poker.Card, len(opponentHands))
		copy(shuffledOpponentHands, opponentHands)
		rand.Shuffle(len(shuffledOpponentHands), func(i, j int) {
			shuffledOpponentHands[i], shuffledOpponentHands[j] = shuffledOpponentHands[j], shuffledOpponentHands[i]
		})
		handsToProcess = shuffledOpponentHands[:10000]
	} else {
		log.Printf("Processing all %d opponent hands (less than or equal to 10000)", len(opponentHands))
	}

	// 共通の並列計算関数を使用してequity計算を実行
	equitiesMap, err := pkrlib.CalculateHandVsRangeEquityParallel(yourHand, handsToProcess, board)
	if err != nil {
		log.Printf("Error calculating equities: %v", err)
		return []models.HandVsRangeResult{}
	}

	// 結果をHandVsRangeResultの形式に変換
	var results []models.HandVsRangeResult
	for villainHandStr, equity := range equitiesMap {
		results = append(results, models.HandVsRangeResult{
			OpponentHand: villainHandStr,
			Equity:       equity,
		})
	}

	// エクイティでソート
	sort.Slice(results, func(i, j int) bool {
		return results[i].Equity > results[j].Equity
	})

	return results
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// loadOpponentRangeFromPreset loads opponent range from CSV file based on preset name
func loadOpponentRangeFromPreset(preset string) (string, error) {
	return fileio.LoadOpponentRangeFromPreset(preset, "data")
}

// カードをソートするためのヘルパー関数
func sortCards(cards []string) []string {
	sorted := make([]string, len(cards))
	copy(sorted, cards)
	sort.Strings(sorted)
	return sorted
}

// handleEquityCalculation handles the equity calculation HTTP request
func handleEquityCalculation(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	var requestData struct {
		YourHands      string   `json:"yourHands"`
		OpponentsHands string   `json:"opponentsHands"`
		FlopCards      []string `json:"flopCards"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// フロップカードの検証
	if len(requestData.FlopCards) != 3 {
		http.Error(w, "Exactly 3 flop cards are required", http.StatusBadRequest)
		return
	}

	// フロップカードをソート
	sortedFlopCards := sortCards(requestData.FlopCards)

	yourHands := strings.Split(requestData.YourHands, ",")

	var formattedYourHands [][]poker.Card

	for i := 0; i < len(yourHands); i++ {
		tmpHand := strings.Split(yourHands[i], "@")[0]
		tempArray := []poker.Card{}
		if len(tmpHand) == 8 {
			for j := 0; j < 8; j += 2 {
				cardStr := strings.ToUpper(tmpHand[j:j+1]) + strings.ToLower(tmpHand[j+1:j+2])
				tempCard := poker.NewCard(cardStr)
				tempArray = append(tempArray, tempCard)
			}
		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		formattedYourHands = append(formattedYourHands, tempArray)
	}

	opponentHands := strings.Split(requestData.OpponentsHands, ",")

	var formattedOpponentHands [][]poker.Card

	for i := 0; i < len(opponentHands); i++ {
		tmpHand := strings.Split(opponentHands[i], "@")[0]
		tempArray := []poker.Card{}
		if len(tmpHand) == 8 {
			for j := 0; j < 8; j += 2 {
				cardStr := strings.ToUpper(tmpHand[j:j+1]) + strings.ToLower(tmpHand[j+1:j+2])
				tempCard := poker.NewCard(cardStr)
				tempArray = append(tempArray, tempCard)
			}
		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		formattedOpponentHands = append(formattedOpponentHands, tempArray)
	}

	board := make([]poker.Card, 0, 3)
	for _, cardStr := range sortedFlopCards {
		card := poker.NewCard(cardStr)
		board = append(board, card)
	}

	var equity [][]interface{} // Placeholder for equity

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(equity)
}

// handEquityCalculationのハンドラーを修正
func handleHandVsRangeCalculation(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	var requestData struct {
		YourHand       string   `json:"yourHand"`
		OpponentsHands string   `json:"opponentsHands"`
		SelectedPreset string   `json:"selectedPreset"`
		FlopCards      []string `json:"flopCards"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// フロップカードの検証と変換
	if len(requestData.FlopCards) != 3 {
		http.Error(w, "Exactly 3 flop cards are required", http.StatusBadRequest)
		return
	}

	// ヒーローハンドの変換
	var yourHand []poker.Card
	tmpHand := strings.Split(requestData.YourHand, "@")[0]
	if len(tmpHand) == 8 {
		for j := 0; j < 8; j += 2 {
			cardStr := strings.ToUpper(tmpHand[j:j+1]) + strings.ToLower(tmpHand[j+1:j+2])
			tempCard := poker.NewCard(cardStr)
			yourHand = append(yourHand, tempCard)
		}
	} else {
		http.Error(w, "Invalid hero hand format", http.StatusBadRequest)
		return
	}

	// プリセットが指定されている場合は、CSVファイルからopponent rangeを読み込む
	var opponentRangeStr string
	if requestData.SelectedPreset != "" {
		var err error
		opponentRangeStr, err = loadOpponentRangeFromPreset(requestData.SelectedPreset)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error loading preset: %v", err), http.StatusInternalServerError)
			return
		}
	} else if requestData.OpponentsHands != "" {
		opponentRangeStr = requestData.OpponentsHands
	} else {
		http.Error(w, "Either selectedPreset or opponentsHands must be provided", http.StatusBadRequest)
		return
	}

	// Opponentレンジの変換
	opponentHands := strings.Split(opponentRangeStr, ",")
	var formattedOpponentHands [][]poker.Card
	for _, hand := range opponentHands {
		tmpHand := strings.Split(hand, "@")[0]
		var tempArray []poker.Card
		if len(tmpHand) == 8 {
			for j := 0; j < 8; j += 2 {
				cardStr := strings.ToUpper(tmpHand[j:j+1]) + strings.ToLower(tmpHand[j+1:j+2])
				tempCard := poker.NewCard(cardStr)
				tempArray = append(tempArray, tempCard)
			}
			formattedOpponentHands = append(formattedOpponentHands, tempArray)
		}
	}

	// フロップの変換
	board := make([]poker.Card, 0, 3)
	for _, cardStr := range requestData.FlopCards {
		card := poker.NewCard(cardStr)
		board = append(board, card)
	}

	// エクイティ計算
	results := calculateHandVsRangeEquity(yourHand, formattedOpponentHands, board)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// handleGetDailyQuizResults は日付に基づいてクイズ結果を取得するハンドラー
func handleGetDailyQuizResults(w http.ResponseWriter, r *http.Request) {
	// CORSヘッダーの設定
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// dateパラメータの取得
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "Date parameter is required", http.StatusBadRequest)
		return
	}

	// dateパラメータのバリデーション（yyyy-mm-dd形式）
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format. Use yyyy-mm-dd", http.StatusBadRequest)
		return
	}

	// PostgreSQL接続の設定
	pgConfig := db.PostgresConfig{
		Host:     "localhost", // 環境変数から取得するように変更可能
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "plo_equity",
	}

	// PostgreSQLに接続
	pgDB, err := db.GetPostgresConnection(pgConfig)
	if err != nil {
		log.Printf("Failed to connect to PostgreSQL: %v", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer pgDB.Close()

	// 指定された日付のデータを取得
	results, err := db.GetDailyQuizResultsByDate(pgDB, date)
	if err != nil {
		log.Printf("Error fetching daily quiz results: %v", err)
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	// 各行の「result」フィールドの中身をサンプリング
	for i, item := range results {
		// resultフィールドが存在し、配列である場合
		if resultData, ok := item["result"].([]interface{}); ok && len(resultData) > 1000 {
			log.Printf("Sampling 1000 items from %d total items in result field of row %d", len(resultData), i)

			// 結果をシャッフル
			rand.Seed(time.Now().UnixNano()) // 乱数生成器の初期化
			shuffledResult := make([]interface{}, len(resultData))
			copy(shuffledResult, resultData)
			rand.Shuffle(len(shuffledResult), func(i, j int) {
				shuffledResult[i], shuffledResult[j] = shuffledResult[j], shuffledResult[i]
			})

			// 最初の1万件を選択
			sampledResult := shuffledResult[:1000]

			// サンプリングした配列を元の「result」フィールドに戻す
			results[i]["result"] = sampledResult
		}
	}

	// 全ての結果を返す（各行の「result」フィールドの中身はサンプリング済み）
	resultsToReturn := results

	// 結果をJSONで返す
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resultsToReturn); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// main initializes the HTTP server
func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	http.HandleFunc("/calculate-equity", handleEquityCalculation)
	http.HandleFunc("/calculate-hand-vs-range", handleHandVsRangeCalculation)
	http.HandleFunc("/api/daily-quiz-results", handleGetDailyQuizResults)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Equity Distribution Backend is running")
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
