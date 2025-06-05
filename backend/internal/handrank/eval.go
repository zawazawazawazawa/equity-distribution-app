package handrank

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"errors"
	"sync"
)


//go:embed ph5_test.bin
var ph5TestBin []byte

var (
	ph5Table    []uint16
	ph5InitOnce sync.Once
)

// ph5テーブルを初期化（最初のEval5WithTable呼び出しで1回だけ実行）
func loadPh5Table() {
	const tableSize = 2598960 // 52C5
	ph5Table = make([]uint16, tableSize)
	if len(ph5TestBin) != tableSize*2 {
		panic("ph5_test.bin size mismatch")
	}
	if err := binary.Read(
		bytes.NewReader(ph5TestBin),
		binary.LittleEndian, ph5Table,
	); err != nil {
		panic(err)
	}
}

// Eval5WithTable: 5枚のカード配列から役順位(rank)を返す（ph5_test.binを使用）
func Eval5WithTable(cards [5]int) (int16, error) {
	ph5InitOnce.Do(loadPh5Table)

	// カードの有効性チェック
	for _, card := range cards {
		if card < 0 || card > 51 {
			return 0, errors.New("invalid card value: must be 0-51")
		}
	}

	// カードを昇順にソート（MakeKey22で必要）
	sortedCards := cards
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 5; j++ {
			if sortedCards[i] > sortedCards[j] {
				sortedCards[i], sortedCards[j] = sortedCards[j], sortedCards[i]
			}
		}
	}

	// 22bitキー生成
	idx := MakeKey22(sortedCards)
	if idx < 0 || int(idx) >= len(ph5Table) {
		return 0, errors.New("invalid table index")
	}

	return int16(ph5Table[idx]), nil
}

