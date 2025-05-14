package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/chehsunliu/poker"
	"github.com/joho/godotenv"

	"equity-distribution-backend/pkg/db"
	"equity-distribution-backend/pkg/fileio"
	pkrlib "equity-distribution-backend/pkg/poker"
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

// バッチ処理の設定
type BatchConfig struct {
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

	log.Printf("Starting sequential processing for %d scenarios", len(scenarios))

	// 乱数生成器の初期化
	rand.Seed(time.Now().UnixNano())

	// 直列で各シナリオを実行
	for i, scenario := range scenarios {
		log.Printf("Starting scenario %d: %s", i+1, scenario.Name)
		log.Printf("Selected scenario: %s", scenario.Name)

		// シナリオに基づいてハンドとフロップを生成
		heroHand, opponentRange, flop := generateHandsAndFlop(scenario, config)

		// equity計算
		equities, err := calculateEquity(heroHand, opponentRange, flop, config)
		if err != nil {
			log.Printf("Error calculating equity: %v", err)
			log.Printf("Scenario %d failed: %v", i+1, err)
			continue
		}

		// 結果をDynamoDBに保存
		err = saveResultToDynamoDB(heroHand, opponentRange, flop, equities, config)
		if err != nil {
			log.Printf("Error saving result to DynamoDB: %v", err)
			log.Printf("Scenario %d failed: %v", i+1, err)
			continue
		}

		log.Printf("Scenario %d completed: %s - Flop: %s, Hero: %s", i+1, scenario.Name, pkrlib.GenerateBoardString(flop), heroHand)
	}

	log.Println("All scenarios processed successfully")
}

// コマンドライン引数を解析する
func parseFlags() *BatchConfig {
	config := &BatchConfig{}

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

// シナリオに基づいてハンドとフロップを生成する
func generateHandsAndFlop(scenario Scenario, config *BatchConfig) (string, string, []poker.Card) {
	// オポーネントレンジはプリセットから読み込む
	opponentRange, err := fileio.LoadOpponentRangeFromPreset(scenario.PresetName, config.DataDir)
	if err != nil {
		// 失敗したらpanicを投げる
		panic(fmt.Sprintf("Failed to load opponent range: %v", err))
	}

	// アグレッサー側のレンジを読み込む
	aggressorRange, err := fileio.LoadAggressorRangeFromPreset(scenario.PresetName, config.DataDir)
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

	log.Printf("Generated flop: %s", pkrlib.GenerateBoardString(flop))
	return heroHand, opponentRange, flop
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
		if pkrlib.HasCardDuplicates(yourHand, opponentHand, flop) {
			continue
		}

		// ハンド文字列の生成
		villainHandStr := ""
		for _, card := range opponentHand {
			villainHandStr += card.String()
		}

		// equity計算
		equity, _ := pkrlib.CalculateHandVsHandEquity(yourHand, opponentHand, flop)
		if equity != -1 {
			equities[villainHandStr] = equity
		}
	}

	if len(equities) == 0 {
		return nil, fmt.Errorf("no valid equity calculations")
	}

	return equities, nil
}

// 結果をDynamoDBに保存する
func saveResultToDynamoDB(heroHand string, opponentRange string, flop []poker.Card, equities map[string]float64, config *BatchConfig) error {
	// フロップ文字列の生成
	flopStr := pkrlib.GenerateBoardString(flop)

	// DynamoDBクライアントを取得
	dbConfig := db.Config{
		Region:   config.DynamoDBRegion,
		Endpoint: config.DynamoDBEndpoint,
	}
	svc := db.GetDynamoDBClient(dbConfig)

	// 各equity結果をDynamoDBに保存
	for villainHand, equity := range equities {
		handCombination := pkrlib.GenerateHandCombination(heroHand, villainHand)
		err := db.InsertDynamoDB(svc, "PloEquity", flopStr, handCombination, equity)
		if err != nil {
			return fmt.Errorf("failed to insert data into DynamoDB: %v", err)
		}
	}

	return nil
}
