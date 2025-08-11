package poker

import (
	"sort"

	"github.com/chehsunliu/poker"
)

// RazzRank represents a rank for Razz (ace-to-five low)
type RazzRank struct {
	cards []int // Card values in ascending order (A=1, 2=2, ..., K=13)
}

// Stud8LowRank represents a rank for 8-or-better low
type Stud8LowRank struct {
	cards    []int // Card values in ascending order (A=1, 2=2, ..., K=13)
	qualifies bool  // Whether the hand qualifies for low (8 or better)
}

// EvaluateRazzHand evaluates a hand for Razz (ace-to-five low)
// Returns a RazzRank where lower is better
func EvaluateRazzHand(cards []poker.Card) RazzRank {
	// Extract unique ranks (pairs don't count in Razz)
	rankSet := make(map[int]bool)
	for _, card := range cards {
		rank := getCardRank(card)
		// In Razz, Ace is always 1
		if rank == 14 {
			rank = 1
		}
		rankSet[rank] = true
	}

	// Convert to slice and sort
	var ranks []int
	for rank := range rankSet {
		ranks = append(ranks, rank)
	}
	sort.Ints(ranks)

	// Take the best 5 cards (lowest)
	bestCards := ranks
	if len(bestCards) > 5 {
		bestCards = bestCards[:5]
	}

	// Pad with high cards if we don't have 5 unique ranks
	for len(bestCards) < 5 {
		// Add the highest possible card that's not already in the hand
		for i := 13; i >= 1; i-- {
			if !contains(bestCards, i) {
				bestCards = append(bestCards, i)
				break
			}
		}
	}

	return RazzRank{cards: bestCards}
}

// EvaluateStud8LowHand evaluates a hand for 8-or-better low
// Returns a Stud8LowRank indicating if it qualifies and its rank
func EvaluateStud8LowHand(cards []poker.Card) Stud8LowRank {
	// Extract unique ranks for low evaluation
	rankSet := make(map[int]bool)
	for _, card := range cards {
		rank := getCardRank(card)
		// In 8-or-better, Ace is always 1 for low
		if rank == 14 {
			rank = 1
		}
		rankSet[rank] = true
	}

	// Convert to slice and sort
	var ranks []int
	for rank := range rankSet {
		ranks = append(ranks, rank)
	}
	sort.Ints(ranks)

	// Check if we have at least 5 cards 8 or lower
	lowCards := []int{}
	for _, rank := range ranks {
		if rank <= 8 {
			lowCards = append(lowCards, rank)
		}
	}

	// If we don't have 5 cards 8 or lower, the hand doesn't qualify
	if len(lowCards) < 5 {
		return Stud8LowRank{
			cards:     []int{},
			qualifies: false,
		}
	}

	// Take the best 5 low cards
	bestCards := lowCards[:5]

	return Stud8LowRank{
		cards:     bestCards,
		qualifies: true,
	}
}

// CompareRazzHands compares two Razz hands
// Returns -1 if hand1 wins, 1 if hand2 wins, 0 if tie
func CompareRazzHands(hand1, hand2 RazzRank) int {
	// Compare each card position from highest to lowest
	// In Razz, we compare starting from the highest card (last in array)
	for i := len(hand1.cards) - 1; i >= 0 && i < len(hand1.cards); i-- {
		if hand1.cards[i] < hand2.cards[i] {
			return -1 // hand1 is lower (better)
		} else if hand1.cards[i] > hand2.cards[i] {
			return 1 // hand2 is lower (better)
		}
	}
	return 0 // tie
}

// CompareStud8LowHands compares two 8-or-better low hands
// Returns -1 if hand1 wins, 1 if hand2 wins, 0 if tie
func CompareStud8LowHands(hand1, hand2 Stud8LowRank) int {
	// If one qualifies and the other doesn't, the qualifying hand wins
	if hand1.qualifies && !hand2.qualifies {
		return -1
	}
	if !hand1.qualifies && hand2.qualifies {
		return 1
	}
	if !hand1.qualifies && !hand2.qualifies {
		return 0 // Neither qualifies, it's a tie for low
	}

	// Both qualify, compare like Razz (from highest to lowest)
	for i := len(hand1.cards) - 1; i >= 0 && i < len(hand1.cards); i-- {
		if hand1.cards[i] < hand2.cards[i] {
			return -1 // hand1 is lower (better)
		} else if hand1.cards[i] > hand2.cards[i] {
			return 1 // hand2 is lower (better)
		}
	}
	return 0 // tie
}

// evaluateBest5CardRazzHand evaluates the best 5-card Razz hand from 7 cards
func evaluateBest5CardRazzHand(cards []poker.Card) RazzRank {
	// For 7-card stud, we need to find the best 5-card combination
	// In Razz, this means the 5 lowest unique ranks
	
	// Extract all ranks
	var allRanks []int
	for _, card := range cards {
		rank := getCardRank(card)
		if rank == 14 {
			rank = 1 // Ace is 1 in Razz
		}
		allRanks = append(allRanks, rank)
	}

	// Remove duplicates and sort
	rankSet := make(map[int]bool)
	for _, rank := range allRanks {
		rankSet[rank] = true
	}

	var uniqueRanks []int
	for rank := range rankSet {
		uniqueRanks = append(uniqueRanks, rank)
	}
	sort.Ints(uniqueRanks)

	// Take the best 5 (lowest)
	bestCards := uniqueRanks
	if len(bestCards) > 5 {
		bestCards = bestCards[:5]
	}

	// If we have fewer than 5 unique ranks, we need to use pairs
	if len(bestCards) < 5 {
		// Count occurrences of each rank
		rankCount := make(map[int]int)
		for _, rank := range allRanks {
			rankCount[rank]++
		}

		// Add duplicates of the lowest ranks until we have 5 cards
		for len(bestCards) < 5 {
			for _, rank := range uniqueRanks {
				if rankCount[rank] > 1 && len(bestCards) < 5 {
					// This rank appears multiple times, we can use it again
					bestCards = append(bestCards, rank)
					rankCount[rank]--
					if rankCount[rank] == 1 {
						// We've used all duplicates of this rank
						break
					}
				}
			}
		}
	}

	// Sort the final hand
	sort.Ints(bestCards)

	return RazzRank{cards: bestCards}
}

// evaluateBest5CardStud8LowHand evaluates the best 5-card 8-or-better low hand from 7 cards
func evaluateBest5CardStud8LowHand(cards []poker.Card) Stud8LowRank {
	// Extract all low-eligible ranks (8 or lower)
	var lowRanks []int
	for _, card := range cards {
		rank := getCardRank(card)
		if rank == 14 {
			rank = 1 // Ace is 1 for low
		}
		if rank <= 8 {
			lowRanks = append(lowRanks, rank)
		}
	}

	// Remove duplicates for 8-or-better (no pairs allowed)
	rankSet := make(map[int]bool)
	for _, rank := range lowRanks {
		rankSet[rank] = true
	}

	var uniqueLowRanks []int
	for rank := range rankSet {
		uniqueLowRanks = append(uniqueLowRanks, rank)
	}

	// Check if we have enough cards to qualify
	if len(uniqueLowRanks) < 5 {
		return Stud8LowRank{
			cards:     []int{},
			qualifies: false,
		}
	}

	// Sort and take the best 5
	sort.Ints(uniqueLowRanks)
	bestCards := uniqueLowRanks[:5]

	return Stud8LowRank{
		cards:     bestCards,
		qualifies: true,
	}
}

// Helper function to get card rank from poker.Card
func getCardRank(card poker.Card) int {
	// Use the Rank() method from the poker library
	// Rank returns 0-12 where 0=2, 1=3, ..., 12=A
	rank := int(card.Rank())
	
	// Convert to our representation where 2=2, 3=3, ..., K=13, A=14
	if rank == 12 {
		return 14 // Ace
	}
	return rank + 2 // 2=2, 3=3, ..., K=13
}

// Helper function to check if a slice contains a value
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}