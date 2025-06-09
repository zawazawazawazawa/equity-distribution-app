package poker

import (
	"github.com/chehsunliu/poker"
)

// JudgeWinner determines the winner between two hands
// Automatically detects game type (2 cards = Hold'em, 4 cards = PLO, 5 cards = PLO5)
func JudgeWinner(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) string {
	// Auto-detect game type based on hand size
	if len(yourHand) == 5 && len(opponentHand) == 5 {
		return JudgeWinnerPLO5(yourHand, opponentHand, board)
	} else if len(yourHand) == 4 && len(opponentHand) == 4 {
		return JudgeWinnerPLO(yourHand, opponentHand, board)
	} else if len(yourHand) == 2 && len(opponentHand) == 2 {
		return JudgeWinnerHoldem(yourHand, opponentHand, board)
	} else {
		// Invalid hand sizes - fallback to old behavior for compatibility
		return JudgeWinnerLegacy(yourHand, opponentHand, board)
	}
}

// JudgeWinnerPLO determines the winner between two PLO hands
// Follows strict PLO rules: exactly 2 cards from hand + 3 cards from board
func JudgeWinnerPLO(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) string {
	yourBestRank := evaluatePLOHand(yourHand, board)
	opponentBestRank := evaluatePLOHand(opponentHand, board)

	if yourBestRank < opponentBestRank {
		return "yourHand"
	} else if yourBestRank > opponentBestRank {
		return "opponentHand"
	} else {
		return "tie"
	}
}

// JudgeWinnerPLO5 determines the winner between two 5-card PLO hands
// Follows strict PLO rules: exactly 2 cards from hand + 3 cards from board
func JudgeWinnerPLO5(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) string {
	yourBestRank := evaluatePLO5Hand(yourHand, board)
	opponentBestRank := evaluatePLO5Hand(opponentHand, board)

	if yourBestRank < opponentBestRank {
		return "yourHand"
	} else if yourBestRank > opponentBestRank {
		return "opponentHand"
	} else {
		return "tie"
	}
}

// JudgeWinnerHoldem determines the winner for Texas Hold'em
func JudgeWinnerHoldem(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) string {
	// For Hold'em, use all 7 cards (2 hand + 5 board) and let poker.Evaluate find best 5
	yourCards := append(board, yourHand...)
	opponentCards := append(board, opponentHand...)

	yourRank := poker.Evaluate(yourCards)
	opponentRank := poker.Evaluate(opponentCards)

	if yourRank < opponentRank {
		return "yourHand"
	} else if yourRank > opponentRank {
		return "opponentHand"
	} else {
		return "tie"
	}
}

// JudgeWinnerLegacy preserves the original behavior for compatibility
func JudgeWinnerLegacy(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) string {
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

// evaluatePLOHand evaluates PLO hand following strict PLO rules
// Must use exactly 2 cards from hand and 3 cards from board
func evaluatePLOHand(hand []poker.Card, board []poker.Card) int32 {
	var bestRank int32 = 7462 // worst possible rank

	// Generate all combinations of 2 cards from hand (C(4,2) = 6)
	for i := 0; i < len(hand); i++ {
		for j := i + 1; j < len(hand); j++ {
			handCards := []poker.Card{hand[i], hand[j]}

			// Generate all combinations of 3 cards from board (C(5,3) = 10)
			for x := 0; x < len(board); x++ {
				for y := x + 1; y < len(board); y++ {
					for z := y + 1; z < len(board); z++ {
						boardCards := []poker.Card{board[x], board[y], board[z]}

						// Create 5-card hand: exactly 2 from hand + 3 from board
						fiveCardHand := append(handCards, boardCards...)
						rank := poker.Evaluate(fiveCardHand)

						if rank < bestRank {
							bestRank = rank
						}
					}
				}
			}
		}
	}

	return bestRank
}

// evaluatePLO5Hand evaluates 5-card PLO hand following strict PLO rules
// Must use exactly 2 cards from hand and 3 cards from board
func evaluatePLO5Hand(hand []poker.Card, board []poker.Card) int32 {
	var bestRank int32 = 7462 // worst possible rank

	// Generate all combinations of 2 cards from hand (C(5,2) = 10)
	for i := 0; i < len(hand); i++ {
		for j := i + 1; j < len(hand); j++ {
			handCards := []poker.Card{hand[i], hand[j]}

			// Generate all combinations of 3 cards from board (C(5,3) = 10)
			for x := 0; x < len(board); x++ {
				for y := x + 1; y < len(board); y++ {
					for z := y + 1; z < len(board); z++ {
						boardCards := []poker.Card{board[x], board[y], board[z]}

						// Create 5-card hand: exactly 2 from hand + 3 from board
						fiveCardHand := append(handCards, boardCards...)
						rank := poker.Evaluate(fiveCardHand)

						if rank < bestRank {
							bestRank = rank
						}
					}
				}
			}
		}
	}

	return bestRank
}
