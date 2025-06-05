// main.go  　(go 1.21 で確認)
package main

import (
	"encoding/binary"
	"equity-distribution-backend/internal/handrank"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"
)

/* -------------------------------------------------------------------------- */
/*  ランダムハンド生成                                                         */
/* -------------------------------------------------------------------------- */

// ランダムに重複しない5枚のカードを生成
func generateRandomHand() [5]int {
	var hand [5]int
	used := make(map[int]bool)

	for i := 0; i < 5; i++ {
		for {
			card := rand.Intn(52)
			if !used[card] {
				hand[i] = card
				used[card] = true
				break
			}
		}
	}

	// 昇順にソート（makeKey22で必要）
	sort.Ints(hand[:])
	return hand
}

/* -------------------------------------------------------------------------- */
/*  1. 22-bit 完全ハッシュ (moved to handrank package)                        */
/* -------------------------------------------------------------------------- */

const (
	tableSize         = 2598960        // 52C5
	binFile           = "ph5_test.bin" // 出力ファイル
	randomTestSamples = 200            // ランダム検証サンプル数
)

/* -------------------------------------------------------------------------- */
/*  2. メイン                                                                  */
/* -------------------------------------------------------------------------- */

func main() {
	start := time.Now()

	table := make([]uint16, tableSize)
	var hand [5]int

	loops := 0

	for a := 0; a < 48; a++ {
		hand[0] = a
		for b := a + 1; b < 49; b++ {
			hand[1] = b
			for d := b + 1; d < 50; d++ {
				hand[2] = d
				for e := d + 1; e < 51; e++ {
					hand[3] = e
					for f := e + 1; f < 52; f++ {
						hand[4] = f

						idx := handrank.MakeKey22(hand)
						table[idx] = uint16(handrank.EvalRank(hand))

						loops++
					}
				}
			}
		}
	}

	fmt.Printf("generated %d hands (%v)\n", loops, time.Since(start))

	/* ------------------------ 検証フェーズ（ランダム200ハンド） ------------- */

	fmt.Printf("verifying with %d random hands... ", randomTestSamples)
	rand.Seed(time.Now().UnixNano()) // ランダムシードを設定

	errors := 0
	for i := 0; i < randomTestSamples; i++ {
		h := generateRandomHand()
		idx := handrank.MakeKey22(h)
		want := handrank.EvalRank(h)
		got := int16(table[idx])
		if want != got {
			if errors < 20 { // 最初の数件だけ表示
				fmt.Printf("\n rank mismatch: hand=%v idx=%d want=%d got=%d",
					h, idx, want, got)
			}
			errors++
		}
	}
	if errors == 0 {
		fmt.Println("OK!")
	} else {
		fmt.Printf("\n%d mismatches detected\n", errors)
	}

	/* ------------------------ ファイル保存 -------------------------------- */

	file, err := os.Create(binFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := binary.Write(file, binary.LittleEndian, table); err != nil {
		panic(err)
	}
	fmt.Printf("table written to %s (%.1f MB)\n",
		binFile, float64(len(table))*2/1024/1024)
}
