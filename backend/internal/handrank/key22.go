package handrank

// choose テーブル & 22-bit 完全ハッシュ
var choose [53][6]int32

func init() {
	for n := 0; n <= 52; n++ {
		choose[n][0] = 1
		for k := 1; k <= 5 && k <= n; k++ {
			choose[n][k] = choose[n-1][k-1] + choose[n-1][k]
		}
	}
}

// MakeKey22 converts 5 cards in ascending order to a unique key (0..2598959)
// Arguments: c [5]int - 5 cards in ascending order (0-51)
// Returns: int32 - unique key for the 5-card combination
func MakeKey22(c [5]int) int32 {
	return choose[c[0]][1] +
		choose[c[1]][2] +
		choose[c[2]][3] +
		choose[c[3]][4] +
		choose[c[4]][5]
}
