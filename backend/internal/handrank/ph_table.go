package handrank

import (
	_ "embed" // for go:embed
)

//------------------------------------------------------------
// 1) 生成コマンド
//    go generate ./internal/handrank        ← 再生成したい時
//go:generate go run ../../cmd/gen_table
//------------------------------------------------------------

//  2. 埋め込み
//     Step-3 で ph_table.bin を生成済みならビルドが通る。
//
//go:embed ph_table.bin
var rawPH []byte

// PHTable : upper-24 bits = key, lower-13 bits = rank
var PHTable []uint32

func init() {
	n := len(rawPH) / 4
	PHTable = make([]uint32, n)
	for i := 0; i < n; i++ {
		PHTable[i] = uint32(rawPH[i*4]) |
			uint32(rawPH[i*4+1])<<8 |
			uint32(rawPH[i*4+2])<<16 |
			uint32(rawPH[i*4+3])<<24
	}
}
