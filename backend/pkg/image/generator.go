package image

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chehsunliu/poker"
	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
)

// カード画像キャッシュ
var cardImageCache sync.Map

// extractPositionFromScenario はシナリオ文字列からポジションを抽出します
func extractPositionFromScenario(scenario string) string {
	// ポジションを表す一般的な略語（UTG, MP, CO, BTN, SB, BB）を検索
	re := regexp.MustCompile(`\b(UTG|MP|CO|BTN|SB|BB)\b`)
	match := re.FindString(scenario)
	if match == "" {
		return "不明"
	}
	return strings.ToUpper(match)
}

// GenerateDailyQuizImage は日毎のクイズ画像を生成します
func GenerateDailyQuizImage(date time.Time, scenario string, heroHand string, flop []poker.Card) error {
	// シナリオからポジションを抽出
	heroPosition := extractPositionFromScenario(scenario)
	// 1. 適切なサイズでキャンバスを作成（X投稿に最適化）
	const (
		width  = 1200 // 必要十分なサイズに抑える
		height = 630
	)

	// 2. 出力ディレクトリの確保（4card/5cardで分ける）
	gameTypeDir := "4card"
	if len(heroHand) == 10 {
		gameTypeDir = "5card"
	}
	outputDir := filepath.Join("./images/daily-quiz", gameTypeDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// 3. 出力ファイルのパス準備
	outputPath := filepath.Join(outputDir, date.Format("2006-01-02")+".png")

	// 4. 描画コンテキストの作成
	dc := gg.NewContext(width, height)

	// 5. 背景を描画（白色）
	dc.SetRGB(1.0, 1.0, 1.0) // 白色
	dc.Clear()

	// 6. タイトルを描画（4-card or 5-card PLOを判定）
	dateStr := date.Format("01/02")
	gameType := "PLO"
	if len(heroHand) == 10 {
		gameType = "PLO5"
	}
	if err := drawTitle(dc, dateStr+" "+gameType+" EQ Quiz"); err != nil {
		return err
	}

	// 7. シナリオ情報を描画
	if err := drawScenario(dc, scenario, date); err != nil {
		return err
	}

	// 8. ヒーローハンドを描画
	// 5カードPLOの場合は位置を調整
	heroX := 100.0
	flopX := 700.0
	if len(heroHand) == 10 {
		// 5カードPLOの場合、左右の余白を減らして中央のスペースを広げる
		heroX = 50.0   // 左の余白を減らす
		flopX = 750.0  // フロップを右に移動してスペースを確保
	}
	if err := drawHeroHand(dc, heroHand, heroPosition, heroX, 330); err != nil {
		return err
	}

	// 9. フロップを描画（ヒーローハンドとの間に十分なスペースを確保）
	if err := drawFlop(dc, flop, flopX, 330); err != nil {
		return err
	}

	// 10. 画像を保存
	if err := saveImage(dc, outputPath); err != nil {
		return err
	}

	// 11. 処理完了後にGCを促進
	runtime.GC()

	return nil
}

// タイトルを描画する関数
func drawTitle(dc *gg.Context, title string) error {
	// フォントの設定
	if err := dc.LoadFontFace("fonts/Inter-Bold.ttf", 60); err != nil {
		// フォントが見つからない場合はエラーを無視して続行
		log.Printf("Warning: Failed to load font: %v", err)
	}

	// タイトルテキストの色を設定（黒色）
	dc.SetRGB(0, 0, 0)

	// タイトルを中央揃えで描画
	dc.DrawStringAnchored(title, float64(dc.Width())/2, 80, 0.5, 0.5)

	return nil
}

// シナリオ情報を描画する関数
func drawScenario(dc *gg.Context, scenario string, date time.Time) error {
	// フォントの設定（タイトルより小さめ）
	if err := dc.LoadFontFace("fonts/Inter-Regular.ttf", 48); err != nil {
		// フォントが見つからない場合はエラーを無視して続行
		log.Printf("Warning: Failed to load font: %v", err)
	}

	// テキストの色を設定（黒色）
	dc.SetRGB(0, 0, 0)

	// シナリオ名を描画
	dc.DrawStringAnchored("Situation:", 100, 180, 0, 0.5)
	dc.DrawStringAnchored(scenario, 330, 180, 0, 0.5)

	return nil
}

// ヒーローハンドを描画する関数
func drawHeroHand(dc *gg.Context, heroHand string, position string, x, y float64) error {
	// フォントの設定
	if err := dc.LoadFontFace("fonts/Inter-Regular.ttf", 48); err != nil {
		// フォントが見つからない場合はエラーを無視して続行
		log.Printf("Warning: Failed to load font: %v", err)
	}

	// テキストの色を設定（黒色）
	dc.SetRGB(0, 0, 0)
	dc.DrawStringAnchored("Hero: "+position, x, y-60, 0, 0.5)

	// カード数を計算（2文字で1枚のカード）
	numCards := len(heroHand) / 2

	// カードの間隔を動的に計算
	// 4カードの場合は通常の間隔、5カードの場合は狭い間隔
	var cardSpacing float64
	if numCards == 4 {
		cardSpacing = 25
	} else if numCards == 5 {
		// 5カードの場合、利用可能なスペースに収まるように間隔を調整
		// 利用可能スペース: 600px、カード幅: 110px × 5 = 550px
		// 残りスペース: 50px を 4つの間隔で分割 = 12.5px
		cardSpacing = 12
	} else {
		cardSpacing = 25 // デフォルト
	}

	// ヒーローハンドの各カードを処理
	currentX := x
	for i := 0; i < len(heroHand); i += 2 {
		if i+1 >= len(heroHand) {
			break
		}

		rank := string(heroHand[i])
		suit := getSuitName(string(heroHand[i+1]))

		// カード画像を取得（キャッシュ機能付き）
		cardImg, err := getCardImage(rank, suit)
		if err != nil {
			return err
		}

		// カードを描画
		dc.DrawImage(cardImg, int(currentX), int(y))
		currentX += float64(cardImg.Bounds().Dx()) + cardSpacing

		// 不要になった変数を解放
		cardImg = nil

		// 4枚ごとにGCを促進（オプション）
		if i%8 == 6 {
			runtime.GC()
		}
	}

	return nil
}

// フロップを描画する関数
func drawFlop(dc *gg.Context, flop []poker.Card, x, y float64) error {
	// フォントの設定
	if err := dc.LoadFontFace("fonts/Inter-Regular.ttf", 48); err != nil {
		// フォントが見つからない場合はエラーを無視して続行
		log.Printf("Warning: Failed to load font: %v", err)
	}

	// テキストの色を設定（黒色）
	dc.SetRGB(0, 0, 0)

	// "Flop:"ラベルをヒーローハンドとフロップカードの間の中央に表示
	// ヒーローハンドの位置は約250、フロップの位置は約500なので、その中間点に配置
	// フロップカードと重ならないように少し上に配置
	dc.DrawStringAnchored("Flop:", 750, 270, 0.5, 0.5)

	// カードの間隔
	const cardSpacing = 25

	// フロップの各カードを処理
	currentX := x
	for _, card := range flop {
		// カード情報を取得
		cardStr := card.String()
		if len(cardStr) < 2 {
			continue
		}

		rank := cardStr[:1]
		suitChar := cardStr[1:2]
		suit := getSuitName(suitChar)

		// カード画像を取得（キャッシュ機能付き）
		cardImg, err := getCardImage(rank, suit)
		if err != nil {
			return err
		}

		// カードを描画
		dc.DrawImage(cardImg, int(currentX), int(y))
		currentX += float64(cardImg.Bounds().Dx()) + cardSpacing

		// 不要になった変数を解放
		cardImg = nil
	}

	return nil
}

// カード画像を取得する関数（キャッシュ機能付き）
func getCardImage(card string, suit string) (image.Image, error) {
	cacheKey := card + "_" + suit

	// キャッシュから取得を試みる
	if cachedImg, ok := cardImageCache.Load(cacheKey); ok {
		return cachedImg.(image.Image), nil
	}

	// カード画像のパスを構築
	// 10がTで表現されているので、10に変換
	if card == "T" {
		card = "10"
	}
	cardPath := filepath.Join("../backend/images/playing_cards", suit, suit+"_"+card+".png")

	// 画像ファイルを開く
	file, err := os.Open(cardPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open card image: %v (path: %s)", err, cardPath)
	}
	defer file.Close()

	// 画像をデコード
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode card image: %v", err)
	}

	// 必要に応じてリサイズ（メモリ使用量削減）
	const (
		cardWidth = 110
	)
	resizedImg := resize.Resize(cardWidth, 0, img, resize.Lanczos3)

	// キャッシュに保存
	cardImageCache.Store(cacheKey, resizedImg)

	// 元の大きい画像を解放
	img = nil

	return resizedImg, nil
}

// スートの名前を取得する補助関数
func getSuitName(suitChar string) string {
	switch suitChar {
	case "c", "C":
		return "club"
	case "d", "D":
		return "diamond"
	case "h", "H":
		return "heart"
	case "s", "S":
		return "spade"
	default:
		return ""
	}
}

// 画像を保存する処理
func saveImage(dc *gg.Context, outputPath string) error {
	// ファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// PNG圧縮レベルを最適化（メモリ使用量とのバランス）
	encoder := png.Encoder{
		CompressionLevel: png.DefaultCompression,
	}

	// 画像をエンコードして保存
	err = encoder.Encode(file, dc.Image())
	if err != nil {
		return err
	}

	return nil
}
