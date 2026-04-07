package poker

import (
	"github.com/chehsunliu/poker"
)

// JudgeWinnerRazz determines the winner between two hands in Razz
func JudgeWinnerRazz(yourHand []poker.Card, opponentHand []poker.Card) string {
	yourRank := evaluateBest5CardRazzHand(yourHand)
	opponentRank := evaluateBest5CardRazzHand(opponentHand)
	
	result := CompareRazzHands(yourRank, opponentRank)
	switch result {
	case -1:
		return "yourHand"
	case 1:
		return "opponentHand"
	default:
		return "tie"
	}
}

// JudgeWinnerStudHigh determines the winner between two hands in 7-card Stud High
func JudgeWinnerStudHigh(yourHand []poker.Card, opponentHand []poker.Card) string {
	// For Stud High, use the standard poker.Evaluate on all 7 cards
	yourRank := poker.Evaluate(yourHand)
	opponentRank := poker.Evaluate(opponentHand)
	
	if yourRank < opponentRank {
		return "yourHand"
	} else if yourRank > opponentRank {
		return "opponentHand"
	} else {
		return "tie"
	}
}

// JudgeWinnerStud8 determines the winner between two hands in 7-card Stud Hi-Lo 8 or better
// Returns the winner for high, low, and whether there's a split pot
func JudgeWinnerStud8(yourHand []poker.Card, opponentHand []poker.Card) (highWinner string, lowWinner string) {
	// Evaluate high hands
	yourHighRank := poker.Evaluate(yourHand)
	opponentHighRank := poker.Evaluate(opponentHand)
	
	if yourHighRank < opponentHighRank {
		highWinner = "yourHand"
	} else if yourHighRank > opponentHighRank {
		highWinner = "opponentHand"
	} else {
		highWinner = "tie"
	}
	
	// Evaluate low hands
	yourLowRank := evaluateBest5CardStud8LowHand(yourHand)
	opponentLowRank := evaluateBest5CardStud8LowHand(opponentHand)
	
	lowResult := CompareStud8LowHands(yourLowRank, opponentLowRank)
	
	// Check if there's a qualifying low
	if !yourLowRank.qualifies && !opponentLowRank.qualifies {
		lowWinner = "none" // No qualifying low
	} else {
		switch lowResult {
		case -1:
			lowWinner = "yourHand"
		case 1:
			lowWinner = "opponentHand"
		default:
			lowWinner = "tie"
		}
	}
	
	return highWinner, lowWinner
}