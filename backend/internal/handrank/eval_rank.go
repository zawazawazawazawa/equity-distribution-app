package handrank

import "sort"

// 13ランクごとに割り当てる素数（2〜A）
var Prime = [...]int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41}

func EvalRank(cards [5]int) int16 {
	// 標準役判定: 役順位は「High Card < One Pair < ... < Royal Flush」
	// ランクごとにカウント
	rankCnt := make([]int, 13)
	suitCnt := make([]int, 4)
	for _, c := range cards {
		rankCnt[c%13]++
		suitCnt[c/13]++
	}
	// フラッシュ
	flush := false
	for _, cnt := range suitCnt {
		if cnt == 5 {
			flush = true
			break
		}
	}
	// ストレート
	ranks := []int{}
	for r, cnt := range rankCnt {
		if cnt > 0 {
			ranks = append(ranks, r)
		}
	}
	sort.Ints(ranks)
	straight := false
	highRank := 0
	if len(ranks) == 5 {
		if ranks[4]-ranks[0] == 4 {
			straight = true
			highRank = ranks[4]
		} else if ranks[0] == 0 && ranks[1] == 1 && ranks[2] == 2 && ranks[3] == 3 && ranks[4] == 12 {
			straight = true
			highRank = 3 // 5-high
		}
	}
	// 判定ロジック
	switch {
	case flush && straight && highRank == 12:
		return 7461 // Royal Flush
	case flush && straight:
		return 7000 + int16(highRank) // Straight Flush
	case contains(rankCnt, 4):
		return 6000 + int16(indexOf(rankCnt, 4)) // Four of a Kind
	case contains(rankCnt, 3) && contains(rankCnt, 2):
		return 5000 + int16(indexOf(rankCnt, 3))*13 + int16(indexOf(rankCnt, 2)) // Full House
	case flush:
		return 4000 + int16(ranks[4])
	case straight:
		return 3000 + int16(highRank)
	case contains(rankCnt, 3):
		return 2000 + int16(indexOf(rankCnt, 3))
	case countOf(rankCnt, 2) == 2:
		return 1000 + int16(indexOf(rankCnt, 2))*13 + int16(secondIndexOf(rankCnt, 2))
	case contains(rankCnt, 2):
		return 500 + int16(indexOf(rankCnt, 2))
	default:
		return int16(ranks[4]) // High Card
	}
}

// 配列内に値があるか
func contains(a []int, x int) bool {
	for _, v := range a {
		if v == x {
			return true
		}
	}
	return false
}

func countOf(a []int, x int) int {
	cnt := 0
	for _, v := range a {
		if v == x {
			cnt++
		}
	}
	return cnt
}
func indexOf(a []int, x int) int {
	for i, v := range a {
		if v == x {
			return i
		}
	}
	return -1
}
func secondIndexOf(a []int, x int) int {
	cnt := 0
	for i, v := range a {
		if v == x {
			if cnt == 1 {
				return i
			}
			cnt++
		}
	}
	return -1
}
