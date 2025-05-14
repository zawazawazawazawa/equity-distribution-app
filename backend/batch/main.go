package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/chehsunliu/poker"
	"github.com/joho/godotenv"
)

// シナリオ定義
type Scenario struct {
	Name           string
	PresetName     string
	Description    string
	HeroHandRanges []string // ヒーローハンドの範囲（将来的な拡張用）
}

// 利用可能なシナリオのリスト
var scenarios = []Scenario{
	{
		Name:        "SRP BB vs UTG",
		PresetName:  "SRP BB call vs UTG open",
		Description: "シングルレイズポット: BBがUTGオープンに対してコール",
	},
	{
		Name:        "SRP BB vs BTN",
		PresetName:  "SRP BB call vs BTN open",
		Description: "シングルレイズポット: BBがBTNオープンに対してコール",
	},
	{
		Name:        "SRP BTN vs UTG",
		PresetName:  "SRP BTN call vs UTG open",
		Description: "シングルレイズポット: BTNがUTGオープンに対してコール",
	},
	{
		Name:        "3BP UTG vs BB",
		PresetName:  "3BP UTG call vs BB 3bet",
		Description: "3ベットポット: UTGがBBの3ベットに対してコール",
	},
	{
		Name:        "3BP UTG vs BTN",
		PresetName:  "3BP UTG call vs BTN 3bet",
		Description: "3ベットポット: UTGがBTNの3ベットに対してコール",
	},
	{
		Name:        "3BP BTN vs BB",
		PresetName:  "3BP BTN call vs BB 3bet",
		Description: "3ベットポット: BTNがBBの3ベットに対してコール",
	},
}

// FlopEquities represents all equity calculations for a specific flop
type FlopEquities struct {
	Flop     string
	Equities map[string]float64 // handCombination -> equity
}

// HandVsRangeResult represents the result of a hand vs range equity calculation
type HandVsRangeResult struct {
	OpponentHand string  `json:"opponentHand"`
	Equity       float64 `json:"equity"`
}

// バッチ処理の設定
type BatchConfig struct {
	NumCalculations  int    // 計算回数
	NumWorkers       int    // ワーカー数
	DynamoDBEndpoint string // DynamoDBエンドポイント
	DynamoDBRegion   string // DynamoDBリージョン
	LogFile          string // ログファイル
	DataDir          string // データディレクトリ
}

func main() {
	// .envファイルの読み込み
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// コマンドライン引数の解析
	config := parseFlags()

	// ログの設定
	setupLogging(config.LogFile)

	log.Printf("Starting batch processing with %d workers for %d calculations", config.NumWorkers, config.NumCalculations)

	// 乱数生成器の初期化
	rand.Seed(time.Now().UnixNano())

	// ワーカープールの作成
	var wg sync.WaitGroup
	jobs := make(chan int, config.NumCalculations)
	results := make(chan string, config.NumCalculations)

	// ワーカーの起動
	for w := 1; w <= config.NumWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg, config)
	}

	// ジョブの送信
	for j := 1; j <= config.NumCalculations; j++ {
		jobs <- j
	}
	close(jobs)

	// 結果の収集
	go func() {
		for result := range results {
			log.Println(result)
		}
	}()

	// すべてのワーカーが完了するのを待つ
	wg.Wait()
	close(results)

	log.Println("Batch processing completed")
}

// コマンドライン引数を解析する
func parseFlags() *BatchConfig {
	config := &BatchConfig{}

	flag.IntVar(&config.NumCalculations, "n", 100, "Number of equity calculations to perform")
	flag.IntVar(&config.NumWorkers, "w", runtime.NumCPU(), "Number of worker goroutines")
	flag.StringVar(&config.DynamoDBEndpoint, "endpoint", "http://localhost:4566", "DynamoDB endpoint URL")
	flag.StringVar(&config.DynamoDBRegion, "region", "us-east-1", "AWS region")
	flag.StringVar(&config.LogFile, "log", "", "Log file (empty for stdout)")
	flag.StringVar(&config.DataDir, "data", "data", "Directory containing preset data files")

	flag.Parse()

	return config
}

// ログの設定
func setupLogging(logFile string) {
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error opening log file: %v", err)
		}
		log.SetOutput(f)
	}
}

// ワーカー関数
func worker(id int, jobs <-chan int, results chan<- string, wg *sync.WaitGroup, config *BatchConfig) {
	defer wg.Done()

	for j := range jobs {
		log.Printf("Worker %d starting job %d", id, j)

		// ジョブIDに基づいてシナリオを選択（6つのシナリオを順番に処理）
		scenarioIndex := (j - 1) % len(scenarios)
		scenario := scenarios[scenarioIndex]
		log.Printf("Selected scenario: %s", scenario.Name)

		// シナリオに基づいてハンドとフロップを生成
		heroHand, opponentRange, flop := generateHandsAndFlop(scenario, config)

		// equity計算
		equities, err := calculateEquity(heroHand, opponentRange, flop, config)
		if err != nil {
			log.Printf("Error calculating equity: %v", err)
			results <- fmt.Sprintf("Job %d failed: %v", j, err)
			continue
		}

		// 結果をDynamoDBに保存
		err = saveResultToDynamoDB(heroHand, opponentRange, flop, equities, config)
		if err != nil {
			log.Printf("Error saving result to DynamoDB: %v", err)
			results <- fmt.Sprintf("Job %d failed: %v", j, err)
			continue
		}

		results <- fmt.Sprintf("Job %d completed: %s - Flop: %s, Hero: %s", j, scenario.Name, generateBoardString(flop), heroHand)
	}
}

// シナリオに基づいてハンドとフロップを生成する
func generateHandsAndFlop(scenario Scenario, config *BatchConfig) (string, string, []poker.Card) {
	// オポーネントレンジはプリセットから読み込む
	opponentRange, err := loadOpponentRangeFromPreset(scenario.PresetName, config.DataDir)
	if err != nil {
		// 失敗したらpanicを投げる
		panic(fmt.Sprintf("Failed to load opponent range: %v", err))
	}

	// アグレッサー側のレンジを読み込む
	aggressorRange, err := loadAggressorRangeFromPreset(scenario.PresetName, config.DataDir)
	if err != nil {
		log.Printf("Error loading aggressor range: %v", err)
		// 失敗したらpanicを投げる
		panic(fmt.Sprintf("Failed to load aggressor range: %v", err))
	}

	// アグレッサー側のレンジからランダムに1ハンドを選ぶ
	aggressorHands := strings.Split(aggressorRange, ",")
	if len(aggressorHands) == 0 {
		// 失敗したらpanicを投げる
		panic("No aggressor hands found")
	}
	heroHand := aggressorHands[rand.Intn(len(aggressorHands))]
	log.Printf("Selected hero hand from aggressor range: %s", heroHand)

	// heroHandに含まれるカードは除外して、flopをランダムに生成
	// ヒーローハンドをpoker.Card形式に変換
	var heroCards []poker.Card
	if len(heroHand) == 8 { // PLOハンド（4枚）
		for j := 0; j < 8; j += 2 {
			cardStr := strings.ToUpper(heroHand[j:j+1]) + strings.ToLower(heroHand[j+1:j+2])
			card := poker.NewCard(cardStr)
			heroCards = append(heroCards, card)
		}
	} else {
		log.Printf("Warning: Unexpected hero hand format: %s", heroHand)
	}

	// 52枚のデッキを生成
	deck := poker.NewDeck()
	fullDeck := deck.Draw(52) // 52枚すべてのカードを取得

	// ヒーローハンドに含まれるカードを除外
	remainingDeck := []poker.Card{}
	for _, card := range fullDeck {
		isInHeroHand := false
		for _, heroCard := range heroCards {
			if card.String() == heroCard.String() {
				isInHeroHand = true
				break
			}
		}
		if !isInHeroHand {
			remainingDeck = append(remainingDeck, card)
		}
	}

	// 残りのカードからランダムに3枚選んでフロップとする
	flop := []poker.Card{}
	for i := 0; i < 3; i++ {
		if len(remainingDeck) == 0 {
			break
		}
		idx := rand.Intn(len(remainingDeck))
		flop = append(flop, remainingDeck[idx])
		// 選んだカードを削除（重複を避けるため）
		remainingDeck = append(remainingDeck[:idx], remainingDeck[idx+1:]...)
	}

	log.Printf("Generated flop: %s", generateBoardString(flop))
	return heroHand, opponentRange, flop
}

// loadAggressorRangeFromPreset loads aggressor range from CSV file based on preset name
func loadAggressorRangeFromPreset(preset string, dataDir string) (string, error) {
	var filePath string
	baseDir := fmt.Sprintf("%s/six_handed_100bb_midrake", dataDir)

	// プリセット値に基づいてアグレッサー側のレンジファイルパスを決定
	switch preset {
	case "SRP BB call vs UTG open":
		filePath = fmt.Sprintf("%s/srp/utg_open.csv", baseDir) // UTGがアグレッサー
	case "SRP BB call vs BTN open":
		filePath = fmt.Sprintf("%s/srp/btn_open.csv", baseDir) // BTNがアグレッサー
	case "SRP BTN call vs UTG open":
		filePath = fmt.Sprintf("%s/srp/utg_open.csv", baseDir) // UTGがアグレッサー
	case "3BP UTG call vs BB 3bet":
		filePath = fmt.Sprintf("%s/3bp/bb_3b_vs_utg.csv", baseDir) // BBがアグレッサー
	case "3BP UTG call vs BTN 3bet":
		filePath = fmt.Sprintf("%s/3bp/btn_3b_vs_utg.csv", baseDir) // BTNがアグレッサー
	case "3BP BTN call vs BB 3bet":
		filePath = fmt.Sprintf("%s/3bp/bb_3b_vs_btn.csv", baseDir) // BBがアグレッサー
	default:
		return "", fmt.Errorf("unknown preset: %s", preset)
	}

	log.Printf("Loading aggressor range from file: %s", filePath)

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

// loadOpponentRangeFromPreset loads opponent range from CSV file based on preset name
func loadOpponentRangeFromPreset(preset string, dataDir string) (string, error) {
	var filePath string
	baseDir := fmt.Sprintf("%s/six_handed_100bb_midrake", dataDir)

	// プリセット値に基づいてファイルパスを決定
	switch preset {
	case "SRP BB call vs UTG open":
		filePath = fmt.Sprintf("%s/srp/bb_call_vs_utg.csv", baseDir)
	case "SRP BB call vs BTN open":
		filePath = fmt.Sprintf("%s/srp/bb_call_vs_btn.csv", baseDir)
	case "SRP BTN call vs UTG open":
		filePath = fmt.Sprintf("%s/srp/btn_call_vs_utg.csv", baseDir)
	case "3BP UTG call vs BB 3bet":
		filePath = fmt.Sprintf("%s/3bp/utg_call_vs_bb.csv", baseDir)
	case "3BP UTG call vs BTN 3bet":
		filePath = fmt.Sprintf("%s/3bp/utg_call_vs_btn.csv", baseDir)
	case "3BP BTN call vs BB 3bet":
		filePath = fmt.Sprintf("%s/3bp/btn_call_vs_bb.csv", baseDir)
	default:
		return "", fmt.Errorf("unknown preset: %s", preset)
	}

	log.Printf("Loading opponent range from file: %s", filePath)

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

// equity計算を実行する
func calculateEquity(heroHand string, opponentRange string, flop []poker.Card, config *BatchConfig) (map[string]float64, error) {
	// ヒーローハンドをpoker.Card形式に変換
	var yourHand []poker.Card
	if len(heroHand) == 8 {
		for j := 0; j < 8; j += 2 {
			cardStr := strings.ToUpper(heroHand[j:j+1]) + strings.ToLower(heroHand[j+1:j+2])
			tempCard := poker.NewCard(cardStr)
			yourHand = append(yourHand, tempCard)
		}
	} else {
		return nil, fmt.Errorf("invalid hero hand format: %s", heroHand)
	}

	// オポーネントレンジをpoker.Card形式に変換
	opponentHands := strings.Split(opponentRange, ",")
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

	// 結果を格納するマップ
	equities := make(map[string]float64)

	// 各オポーネントハンドに対してequity計算
	for _, opponentHand := range formattedOpponentHands {
		// カード重複チェック
		if hasCardDuplicates(yourHand, opponentHand, flop) {
			continue
		}

		// ハンド文字列の生成
		villainHandStr := ""
		for _, card := range opponentHand {
			villainHandStr += card.String()
		}

		// equity計算
		equity, _ := calculateHandVsHandEquity(yourHand, opponentHand, flop)
		if equity != -1 {
			equities[villainHandStr] = equity
		}
	}

	if len(equities) == 0 {
		return nil, fmt.Errorf("no valid equity calculations")
	}

	return equities, nil
}

// calculateHandVsHandEquity calculates the equity between two hands
func calculateHandVsHandEquity(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) (float64, bool) {
	// Check for duplicate cards
	if hasCardDuplicates(yourHand, opponentHand, board) {
		return -1, false // Return -1 to indicate invalid hand due to duplicate cards
	}

	// Generate the full deck
	deck := poker.NewDeck()
	fullDeck := deck.Draw(52) // Draw all 52 cards from the deck

	usedCards := append(yourHand, opponentHand...)
	usedCards = append(usedCards, board...)

	remainingDeck := []poker.Card{}
	for _, card := range fullDeck {
		used := false
		for _, usedCard := range usedCards {
			if card == usedCard {
				used = true
				break
			}
		}
		if !used {
			remainingDeck = append(remainingDeck, card)
		}
	}

	// Calculate equity
	totalOutcomes := 0.0
	winCount := 0.0

	for i := 0; i < len(remainingDeck); i++ {
		for j := i + 1; j < len(remainingDeck); j++ {
			finalBoard := append(board, remainingDeck[i], remainingDeck[j])
			winner := judgeWinner(yourHand, opponentHand, finalBoard)
			if winner == "yourHand" {
				winCount += 1
			} else if winner == "tie" {
				winCount += 0.5
			}
			totalOutcomes += 1
		}
	}

	calculatedEquity := (winCount / totalOutcomes) * 100
	return calculatedEquity, false
}

// judgeWinner determines the winner between two hands
func judgeWinner(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) string {
	// @doc: https://github.com/chehsunliu/poker/blob/72fcd0dd66288388735cc494db3f2bd11b28bfed/lookup.go#L12
	var maxYourHandRank int32 = 7462
	var maxOpponentHandRank int32 = 7462

	// Generate all combinations of your hand and board
	for i := 0; i < len(yourHand); i++ {
		for j := i + 1; j < len(yourHand); j++ {
			newBoard := append(board, yourHand[i], yourHand[j])
			yourHandRank := poker.Evaluate(newBoard)
			if yourHandRank < maxYourHandRank {
				maxYourHandRank = yourHandRank
			}
		}
	}

	// Generate all combinations of opponent's hand and board
	for i := 0; i < len(opponentHand); i++ {
		for j := i + 1; j < len(opponentHand); j++ {
			newBoard := append(board, opponentHand[i], opponentHand[j])
			opponentHandRank := poker.Evaluate(newBoard)
			if opponentHandRank < maxOpponentHandRank {
				maxOpponentHandRank = opponentHandRank
			}
		}
	}

	if maxYourHandRank < maxOpponentHandRank {
		return "yourHand"
	} else if maxYourHandRank > maxOpponentHandRank {
		return "opponentHand"
	} else {
		return "tie"
	}
}

// hasCardDuplicates checks if there are any duplicate cards across all provided card arrays
func hasCardDuplicates(cards ...[]poker.Card) bool {
	seen := make(map[string]bool)
	for _, hand := range cards {
		for _, card := range hand {
			cardStr := card.String()
			if seen[cardStr] {
				return true
			}
			seen[cardStr] = true
		}
	}
	return false
}

// generateBoardString creates a string representation of the board cards
func generateBoardString(board []poker.Card) string {
	boardStr := ""
	for _, card := range board {
		boardStr += card.String()
	}
	return boardStr
}

// generateHandCombination creates a unique combination key for the hands
func generateHandCombination(heroHand string, villainHand string) string {
	hands := []string{heroHand, villainHand}
	sort.Strings(hands) // Sort alphabetically to ensure uniqueness
	return fmt.Sprintf("%s_%s", hands[0], hands[1])
}

// 結果をDynamoDBに保存する
func saveResultToDynamoDB(heroHand string, opponentRange string, flop []poker.Card, equities map[string]float64, config *BatchConfig) error {
	// フロップ文字列の生成
	flopStr := generateBoardString(flop)

	// 各equity結果をDynamoDBに保存
	for villainHand, equity := range equities {
		handCombination := generateHandCombination(heroHand, villainHand)
		err := insertDynamoDB(flopStr, handCombination, equity, config)
		if err != nil {
			return fmt.Errorf("failed to insert data into DynamoDB: %v", err)
		}
	}

	return nil
}

// getDynamoDBClient initializes and returns a DynamoDB client
func getDynamoDBClient(config *BatchConfig) *dynamodb.DynamoDB {
	// AWS設定
	awsConfig := &aws.Config{
		Region:   aws.String(config.DynamoDBRegion),
		Endpoint: aws.String(config.DynamoDBEndpoint),
	}

	// 認証情報を設定（LocalStackの場合はダミーでOK）
	awsConfig.Credentials = credentials.NewStaticCredentials("test", "test", "")

	// セッションを作成
	sess := session.Must(session.NewSession(awsConfig))

	// DynamoDBクライアントを作成
	return dynamodb.New(sess)
}

// insertDynamoDB inserts or updates an item in DynamoDB
func insertDynamoDB(flop string, handCombination string, equity float64, config *BatchConfig) error {
	log.Printf("Inserting data - Flop: %s, HandCombination: %s, Equity: %.2f", flop, handCombination, equity)
	svc := getDynamoDBClient(config)
	item := map[string]*dynamodb.AttributeValue{
		"Flop": {
			S: aws.String(flop),
		},
		"HandCombination": {
			S: aws.String(handCombination),
		},
		"Equity": {
			N: aws.String(fmt.Sprintf("%.2f", equity)),
		},
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String("PloEquity"),
	}

	_, err := svc.PutItem(input)
	if err != nil {
		log.Printf("Error inserting data into DynamoDB: %v", err)
		return fmt.Errorf("failed to insert data into DynamoDB: %v", err)
	}
	log.Printf("Successfully inserted data into DynamoDB for %s", handCombination)
	return nil
}
