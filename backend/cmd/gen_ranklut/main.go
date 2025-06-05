package main

import (
	"equity-distribution-backend/internal/handrank"
	"fmt"
	"os"
)

func main() {
	const lutSize = 4888
	lut := make([]int16, lutSize)
	deck := [52]int{}
	for i := 0; i < 52; i++ {
		deck[i] = i
	}
	var c [5]int
	total := 0
	for a := 0; a < 48; a++ {
		c[0] = deck[a]
		for b := a + 1; b < 49; b++ {
			c[1] = deck[b]
			for d := b + 1; d < 50; d++ {
				c[2] = deck[d]
				for e := d + 1; e < 51; e++ {
					c[3] = deck[e]
					for f := e + 1; f < 52; f++ {
						c[4] = deck[f]
						prod := handrank.Prime[c[0]%13] * handrank.Prime[c[1]%13] * handrank.Prime[c[2]%13] * handrank.Prime[c[3]%13] * handrank.Prime[c[4]%13]
						rank := handrank.EvalRank(c)
						idx := prod % lutSize
						if rank > lut[idx] {
							lut[idx] = rank
						}
						total++
					}
				}
			}
		}
	}
	fmt.Printf("Total hands: %d\n", total)

	// 保存: internal/handrank/rankLUT.go
	// 生成したLUTをGoコードとして保存

	f, err := os.Create("internal/handrank/rankLUT.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		return
	}
	defer f.Close()
	fmt.Fprintln(f, "package handrank\n")
	fmt.Fprintln(f, "// rankLUT: prime-product LUT for hand ranking")
	fmt.Fprintln(f, "var RankLUT = [4888]int16{")
	for i, v := range lut {
		if i%16 == 0 {
			fmt.Fprint(f, "\n\t")
		}
		fmt.Fprintf(f, "%d,", v)
	}
	fmt.Fprintln(f, "\n}")
}
