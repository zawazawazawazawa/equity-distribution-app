package poker

import (
	"github.com/chehsunliu/poker"
)

// GenerateBoardString creates a string representation of the board cards
func GenerateBoardString(board []poker.Card) string {
	boardStr := ""
	for _, card := range board {
		boardStr += card.String()
	}
	return boardStr
}


// HasCardDuplicates checks if there are any duplicate cards across all provided card arrays
func HasCardDuplicates(cards ...[]poker.Card) bool {
	seen := make(map[string]bool)
	for _, hand := range cards {
		for _, card := range hand {
			cardStr := card.String()
			if seen[cardStr] {
				return true
			}
			seen[cardStr] = true
		}
	}
	return false
}
