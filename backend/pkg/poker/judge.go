package poker

import (
	"github.com/chehsunliu/poker"
)

// JudgeWinner determines the winner between two hands
func JudgeWinner(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) string {
	// @doc: https://github.com/chehsunliu/poker/blob/72fcd0dd66288388735cc494db3f2bd11b28bfed/lookup.go#L12
	var maxYourHandRank int32 = 7462
	var maxOpponentHandRank int32 = 7462

	// Generate all combinations of your hand and board
	for i := 0; i < len(yourHand); i++ {
		for j := i + 1; j < len(yourHand); j++ {
			newBoard := append(board, yourHand[i], yourHand[j])
			yourHandRank := poker.Evaluate(newBoard)
			if yourHandRank < maxYourHandRank {
				maxYourHandRank = yourHandRank
			}
		}
	}

	// Generate all combinations of opponent's hand and board
	for i := 0; i < len(opponentHand); i++ {
		for j := i + 1; j < len(opponentHand); j++ {
			newBoard := append(board, opponentHand[i], opponentHand[j])
			opponentHandRank := poker.Evaluate(newBoard)
			if opponentHandRank < maxOpponentHandRank {
				maxOpponentHandRank = opponentHandRank
			}
		}
	}

	if maxYourHandRank < maxOpponentHandRank {
		return "yourHand"
	} else if maxYourHandRank > maxOpponentHandRank {
		return "opponentHand"
	} else {
		return "tie"
	}
}
