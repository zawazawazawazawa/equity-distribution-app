package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chehsunliu/poker"
	"github.com/joho/godotenv"

	"equity-distribution-backend/pkg/db"
	"equity-distribution-backend/pkg/image"
	"equity-distribution-backend/pkg/storage"
)

// 設定を保持する構造体
type Config struct {
	// 日付と画像生成用パラメータ
	Date         string
	ScenarioName string
	HeroHand     string
	Flop         string

	// R2設定
	R2Endpoint  string
	R2AccessKey string
	R2SecretKey string
	R2Bucket    string

	// PostgreSQL設定
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string

	// その他の設定
	LogFile string
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

	log.Println("Starting image upload test script")

	// 日付の処理
	var targetDate time.Time
	if config.Date == "" {
		// 日付が指定されていない場合は今日の日付を使用
		targetDate = time.Now()
	} else {
		// 指定された日付を解析
		var err error
		targetDate, err = time.Parse("2006-01-02", config.Date)
		if err != nil {
			log.Fatalf("Invalid date format: %v. Please use YYYY-MM-DD format.", err)
		}
	}

	log.Printf("Using date: %s", targetDate.Format("2006-01-02"))

	// PostgreSQL接続の確立
	pgConfig := db.PostgresConfig{
		Host:     config.PostgresHost,
		Port:     config.PostgresPort,
		User:     config.PostgresUser,
		Password: config.PostgresPassword,
		DBName:   config.PostgresDBName,
	}

	pgDB, err := db.GetPostgresConnection(pgConfig)
	if err != nil {
		log.Printf("Warning: Failed to connect to PostgreSQL: %v", err)
		log.Println("Proceeding with command line arguments instead.")
	} else {
		defer pgDB.Close()
		log.Println("Successfully connected to PostgreSQL.")

		// データベースから指定された日付のデータを取得
		results, err := db.GetDailyQuizResultsByDate(pgDB, targetDate)
		if err != nil {
			log.Printf("Warning: Failed to get data from PostgreSQL: %v", err)
			log.Println("Proceeding with command line arguments instead.")
		} else if len(results) > 0 {
			// 1番目のデータを使用
			firstResult := results[0]
			log.Printf("Found %d results in database. Using the first one.", len(results))

			// データベースから取得したデータを使用
			config.ScenarioName = firstResult["scenario"].(string)
			config.HeroHand = firstResult["hero_hand"].(string)
			flopStr := firstResult["flop"].(string)

			log.Printf("Using data from database - Scenario: %s, Hero Hand: %s, Flop: %s",
				config.ScenarioName, config.HeroHand, flopStr)

			// 画像ファイルのパス
			imagePath := filepath.Join("images/daily-quiz", targetDate.Format("2006-01-02")+".png")

			// フロップの処理
			var flopCards []poker.Card
			if flopStr != "" {
				// フロップ文字列からカードを解析（例: "[2c, 3d, 4h]" -> "2c3d4h"）
				cleanFlopStr := strings.ReplaceAll(flopStr, "[", "")
				cleanFlopStr = strings.ReplaceAll(cleanFlopStr, "]", "")
				cleanFlopStr = strings.ReplaceAll(cleanFlopStr, " ", "")
				cleanFlopStr = strings.ReplaceAll(cleanFlopStr, ",", "")
				flopCards = parseFlop(cleanFlopStr)
			} else {
				// フロップが空の場合はデフォルトのフロップを使用
				flopCards = []poker.Card{
					poker.NewCard("2c"),
					poker.NewCard("3d"),
					poker.NewCard("4h"),
				}
			}

			log.Printf("Using flop: %v", formatFlop(flopCards))

			// 画像生成
			log.Println("Generating image...")
			err = image.GenerateDailyQuizImage(
				targetDate,
				config.ScenarioName,
				config.HeroHand,
				flopCards,
			)
			if err != nil {
				log.Fatalf("Error generating daily quiz image: %v", err)
			}
			log.Printf("Successfully generated daily quiz image at: %s", imagePath)

			// R2設定を確認
			if config.R2Endpoint == "" || config.R2AccessKey == "" || config.R2SecretKey == "" || config.R2Bucket == "" {
				log.Println("R2 configuration is incomplete. Skipping upload.")
				log.Println("Image generation completed successfully.")
				return
			}

			// R2設定
			r2Config := storage.R2Config{
				Endpoint:   config.R2Endpoint,
				AccessKey:  config.R2AccessKey,
				SecretKey:  config.R2SecretKey,
				BucketName: config.R2Bucket,
			}

			// R2クライアントを作成
			log.Println("Creating R2 client...")
			r2Client, err := storage.GetR2Client(r2Config)
			if err != nil {
				log.Fatalf("Error creating R2 client: %v", err)
			}

			// 画像をR2にアップロード
			objectKey := "daily-quiz/" + targetDate.Format("2006-01-02") + ".png"
			log.Printf("Uploading image to R2 with key: %s", objectKey)
			err = storage.UploadImageToR2(r2Client, r2Config.BucketName, imagePath, objectKey)
			if err != nil {
				log.Fatalf("Error uploading image to R2: %v", err)
			}
			log.Printf("Successfully uploaded image to R2: %s", objectKey)

			// 公開URLを生成
			publicURL := storage.GetR2ObjectURL(r2Config.Endpoint, r2Config.BucketName, objectKey)
			log.Printf("Image public URL: %s", publicURL)

			log.Println("Image generation and upload completed successfully")
			return
		} else {
			log.Printf("Warning: No data found in PostgreSQL for date: %s", targetDate.Format("2006-01-02"))
			log.Println("Proceeding with command line arguments instead.")
		}
	}

	// データベースからデータが取得できなかった場合はコマンドライン引数を使用
	log.Printf("Using command line arguments - Scenario: %s, Hero Hand: %s", config.ScenarioName, config.HeroHand)

	// フロップの処理
	var flopCards []poker.Card
	if config.Flop != "" {
		// フロップが指定されている場合は解析
		flopCards = parseFlop(config.Flop)
	} else {
		// フロップが指定されていない場合はデフォルトのフロップを使用
		flopCards = []poker.Card{
			poker.NewCard("2c"),
			poker.NewCard("3d"),
			poker.NewCard("4h"),
		}
	}

	log.Printf("Using flop: %v", formatFlop(flopCards))

	// 画像ファイルのパス
	imagePath := filepath.Join("images/daily-quiz", targetDate.Format("2006-01-02")+".png")

	// 画像生成
	log.Println("Generating image...")
	err = image.GenerateDailyQuizImage(
		targetDate,
		config.ScenarioName,
		config.HeroHand,
		flopCards,
	)
	if err != nil {
		log.Fatalf("Error generating daily quiz image: %v", err)
	}
	log.Printf("Successfully generated daily quiz image at: %s", imagePath)

	// R2設定を確認
	if config.R2Endpoint == "" || config.R2AccessKey == "" || config.R2SecretKey == "" || config.R2Bucket == "" {
		log.Println("R2 configuration is incomplete. Skipping upload.")
		log.Println("Image generation completed successfully.")
		return
	}

	// R2設定
	r2Config := storage.R2Config{
		Endpoint:   config.R2Endpoint,
		AccessKey:  config.R2AccessKey,
		SecretKey:  config.R2SecretKey,
		BucketName: config.R2Bucket,
	}

	// R2クライアントを作成
	log.Println("Creating R2 client...")
	r2Client, err := storage.GetR2Client(r2Config)
	if err != nil {
		log.Fatalf("Error creating R2 client: %v", err)
	}

	// 画像をR2にアップロード
	objectKey := "daily-quiz/" + targetDate.Format("2006-01-02") + ".png"
	log.Printf("Uploading image to R2 with key: %s", objectKey)
	err = storage.UploadImageToR2(r2Client, r2Config.BucketName, imagePath, objectKey)
	if err != nil {
		log.Fatalf("Error uploading image to R2: %v", err)
	}
	log.Printf("Successfully uploaded image to R2: %s", objectKey)

	// 公開URLを生成
	publicURL := storage.GetR2ObjectURL(r2Config.Endpoint, r2Config.BucketName, objectKey)
	log.Printf("Image public URL: %s", publicURL)

	log.Println("Image generation and upload completed successfully")
}

// コマンドライン引数を解析する
func parseFlags() *Config {
	config := &Config{}

	// 環境変数からデフォルト値を取得
	r2Endpoint := getEnvOrDefault("R2_ENDPOINT", "")
	r2AccessKey := getEnvOrDefault("R2_ACCESS_KEY", "")
	r2SecretKey := getEnvOrDefault("R2_SECRET_KEY", "")
	r2Bucket := getEnvOrDefault("R2_BUCKET", "")

	// PostgreSQL設定を環境変数から取得
	postgresHost := getEnvOrDefault("POSTGRES_HOST", "localhost")
	postgresPort := getEnvIntOrDefault("POSTGRES_PORT", 5432)
	postgresUser := getEnvOrDefault("POSTGRES_USER", "postgres")
	postgresPassword := getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
	postgresDBName := getEnvOrDefault("POSTGRES_DBNAME", "plo_equity")

	// コマンドライン引数の定義
	flag.StringVar(&config.LogFile, "log", "", "Log file (empty for stdout)")
	flag.StringVar(&config.Date, "date", "", "Date for quiz in YYYY-MM-DD format (default: today)")
	flag.StringVar(&config.ScenarioName, "scenario", "SRP UTG vs BB", "Scenario name (fallback if database lookup fails)")
	flag.StringVar(&config.HeroHand, "hand", "AsKsQsJs", "Hero hand (e.g., AsKsQsJs) (fallback if database lookup fails)")
	flag.StringVar(&config.Flop, "flop", "", "Flop cards (e.g., 2c3d4h) (fallback if database lookup fails)")

	// R2設定
	flag.StringVar(&config.R2Endpoint, "r2-endpoint", r2Endpoint, "R2 endpoint")
	flag.StringVar(&config.R2AccessKey, "r2-access-key", r2AccessKey, "R2 access key")
	flag.StringVar(&config.R2SecretKey, "r2-secret-key", r2SecretKey, "R2 secret key")
	flag.StringVar(&config.R2Bucket, "r2-bucket", r2Bucket, "R2 bucket name")

	// PostgreSQL設定
	flag.StringVar(&config.PostgresHost, "pg-host", postgresHost, "PostgreSQL host")
	flag.IntVar(&config.PostgresPort, "pg-port", postgresPort, "PostgreSQL port")
	flag.StringVar(&config.PostgresUser, "pg-user", postgresUser, "PostgreSQL user")
	flag.StringVar(&config.PostgresPassword, "pg-password", postgresPassword, "PostgreSQL password")
	flag.StringVar(&config.PostgresDBName, "pg-dbname", postgresDBName, "PostgreSQL database name")

	flag.Parse()

	return config
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
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
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

// フロップ文字列を解析してpoker.Cardのスライスに変換する
func parseFlop(flopStr string) []poker.Card {
	var cards []poker.Card

	// 2文字ずつ処理（例: "2c3d4h" -> "2c", "3d", "4h"）
	for i := 0; i < len(flopStr); i += 2 {
		if i+2 <= len(flopStr) {
			cardStr := flopStr[i : i+2]
			// カードの表記を標準化（例: "2c" -> "2c"）
			rank := strings.ToUpper(cardStr[0:1])
			suit := strings.ToLower(cardStr[1:2])
			standardCardStr := rank + suit

			card := poker.NewCard(standardCardStr)
			cards = append(cards, card)
		}
	}

	return cards
}

// フロップカードを文字列形式にフォーマットする
func formatFlop(cards []poker.Card) string {
	var cardStrs []string
	for _, card := range cards {
		cardStrs = append(cardStrs, card.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(cardStrs, ", "))
}
