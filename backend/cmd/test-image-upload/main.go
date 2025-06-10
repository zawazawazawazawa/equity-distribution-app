package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chehsunliu/poker"
	"github.com/joho/godotenv"

	"equity-distribution-backend/pkg/db"
	"equity-distribution-backend/pkg/image"
	"equity-distribution-backend/pkg/storage"
)

func main() {
	// .envファイルの読み込み
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// コマンドライン引数の定義
	var (
		date       string
		scenario   string
		heroHand   string
		flopStr    string
		logFile    string
		pgHost     string
		pgPort     int
		pgUser     string
		pgPassword string
		pgDBName   string
	)

	// デフォルト値の設定
	flag.StringVar(&date, "date", "", "Date for the quiz (YYYY-MM-DD format, default: today)")
	flag.StringVar(&scenario, "scenario", "PLO5 SRP UTG vs BB", "Scenario name (default: 'PLO5 SRP UTG vs BB')")
	flag.StringVar(&heroHand, "hand", "AsKsQsJsTs", "Hero hand (default: 'AsKsQsJsTs' for 5-card PLO)")
	flag.StringVar(&flopStr, "flop", "2c3d4h", "Flop cards (default: '2c3d4h')")
	flag.StringVar(&logFile, "log", "", "Log file (empty for stdout)")
	flag.StringVar(&pgHost, "pg-host", getEnvOrDefault("POSTGRES_HOST", "localhost"), "PostgreSQL host")
	flag.IntVar(&pgPort, "pg-port", getEnvIntOrDefault("POSTGRES_PORT", 5432), "PostgreSQL port")
	flag.StringVar(&pgUser, "pg-user", getEnvOrDefault("POSTGRES_USER", "postgres"), "PostgreSQL user")
	flag.StringVar(&pgPassword, "pg-password", getEnvOrDefault("POSTGRES_PASSWORD", "postgres"), "PostgreSQL password")
	flag.StringVar(&pgDBName, "pg-dbname", getEnvOrDefault("POSTGRES_DBNAME", "plo_equity"), "PostgreSQL database name")

	flag.Parse()

	// ログの設定
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Error opening log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	// 日付の処理
	var targetDate time.Time
	if date == "" {
		targetDate = time.Now()
	} else {
		var err error
		targetDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			log.Fatalf("Invalid date format: %v. Please use YYYY-MM-DD format.", err)
		}
	}

	log.Printf("Starting image upload test for date: %s", targetDate.Format("2006-01-02"))

	// PostgreSQL接続（既存データを確認するため）
	pgConfig := db.PostgresConfig{
		Host:     pgHost,
		Port:     pgPort,
		User:     pgUser,
		Password: pgPassword,
		DBName:   pgDBName,
	}

	pgDB, err := db.GetPostgresConnection(pgConfig)
	if err != nil {
		log.Printf("Warning: Failed to connect to PostgreSQL: %v", err)
		log.Printf("Continuing with default values...")
	} else {
		defer pgDB.Close()

		// 指定された日付のデータを取得
		existingResults, err := db.GetDailyQuizResultsByDate(pgDB, targetDate)
		if err == nil && len(existingResults) > 0 {
			// 最初の結果を使用
			result := existingResults[0]
			if s, ok := result["scenario"].(string); ok && s != "" {
				scenario = s
			}
			if h, ok := result["hero_hand"].(string); ok && h != "" {
				heroHand = h
			}
			if f, ok := result["flop"].(string); ok && f != "" {
				flopStr = f
			}
			log.Printf("Using data from database: scenario=%s, hero=%s, flop=%s", scenario, heroHand, flopStr)
		}
	}

	// フロップをpoker.Card形式に変換
	var flop []poker.Card
	// フロップ文字列を解析（2文字ずつ処理）
	for i := 0; i < len(flopStr); i += 2 {
		if i+2 <= len(flopStr) {
			cardStr := strings.ToUpper(flopStr[i:i+1]) + strings.ToLower(flopStr[i+1:i+2])
			card := poker.NewCard(cardStr)
			flop = append(flop, card)
		}
	}

	// ゲームタイプを判定
	gameTypeDir := "4card"
	if len(heroHand) == 10 {
		gameTypeDir = "5card"
	}

	// 画像生成
	log.Printf("Generating image...")
	imagePath := filepath.Join("images/daily-quiz", gameTypeDir, targetDate.Format("2006-01-02")+".png")

	err = image.GenerateDailyQuizImage(
		targetDate,
		scenario,
		heroHand,
		flop,
	)
	if err != nil {
		log.Fatalf("Error generating daily quiz image: %v", err)
	}
	log.Printf("Successfully generated daily quiz image for %s", targetDate.Format("2006-01-02"))

	// R2へのアップロード
	r2Config := storage.R2Config{
		Endpoint:   getEnvOrDefault("R2_ENDPOINT", ""),
		AccessKey:  getEnvOrDefault("R2_ACCESS_KEY", ""),
		SecretKey:  getEnvOrDefault("R2_SECRET_KEY", ""),
		BucketName: getEnvOrDefault("R2_BUCKET", ""),
	}

	// R2設定の確認
	if r2Config.Endpoint == "" || r2Config.AccessKey == "" || r2Config.SecretKey == "" || r2Config.BucketName == "" {
		log.Printf("Warning: R2 configuration is incomplete. Skipping upload.")
		log.Printf("Image saved locally at: %s", imagePath)
		fmt.Printf("\n=== Image Generated ===\n")
		fmt.Printf("Game Type: %s\n", gameTypeDir)
		fmt.Printf("Local Path: %s\n", imagePath)
		return
	}

	// R2クライアントを作成
	r2Client, err := storage.GetR2Client(r2Config)
	if err != nil {
		log.Fatalf("Error creating R2 client: %v", err)
	}

	// 画像をR2にアップロード（パスも4card/5cardで分ける）
	objectKey := "daily-quiz/" + gameTypeDir + "/" + targetDate.Format("2006-01-02") + ".png"
	err = storage.UploadImageToR2(r2Client, r2Config.BucketName, imagePath, objectKey)
	if err != nil {
		log.Fatalf("Error uploading image to R2: %v", err)
	}

	// 公開URLを生成
	publicURL := storage.GetR2ObjectURL(r2Config.Endpoint, r2Config.BucketName, objectKey)
	log.Printf("Successfully uploaded image to R2: %s", objectKey)
	log.Printf("Public URL: %s", publicURL)

	fmt.Printf("\n=== Upload Complete ===\n")
	fmt.Printf("Game Type: %s\n", gameTypeDir)
	fmt.Printf("Local Path: %s\n", imagePath)
	fmt.Printf("R2 Object: %s\n", objectKey)
	fmt.Printf("Public URL: %s\n", publicURL)
}

// 環境変数から文字列値を取得するヘルパー関数
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// 環境変数から整数値を取得するヘルパー関数
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		var intValue int
		fmt.Sscanf(value, "%d", &intValue)
		if intValue > 0 {
			return intValue
		}
	}
	return defaultValue
}