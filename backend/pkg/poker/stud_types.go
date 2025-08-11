package poker

import "github.com/chehsunliu/poker"

// StudGameType represents the type of stud game
type StudGameType int

const (
	StudRazz StudGameType = iota
	StudHigh
	StudHighLow8
)

// StudHand represents a hand in stud games with down cards and up cards
type StudHand struct {
	DownCards []poker.Card // Hidden cards (hole cards)
	UpCards   []poker.Card // Exposed cards
}

// AllCards returns all cards in the hand
func (sh StudHand) AllCards() []poker.Card {
	allCards := make([]poker.Card, 0, len(sh.DownCards)+len(sh.UpCards))
	allCards = append(allCards, sh.DownCards...)
	allCards = append(allCards, sh.UpCards...)
	return allCards
}

// Stud8EquityResult represents the equity calculation result for Stud Hi-Lo 8 or better
type Stud8EquityResult struct {
	HighEquity  float64 // Probability of winning the high pot (%)
	LowEquity   float64 // Probability of winning the low pot (%)
	ScoopEquity float64 // Probability of winning both high and low (%)
}

// StudEquityResult represents the equity calculation result for stud games
type StudEquityResult struct {
	Equity          float64            // Overall equity (for Razz and Stud High)
	Stud8Result     *Stud8EquityResult // Detailed result for Stud Hi-Lo 8
	TotalIterations int                // Number of iterations run
	GameType        StudGameType       // Type of stud game
}

// GetGameTypeString returns the string representation of the game type
func (sgt StudGameType) String() string {
	switch sgt {
	case StudRazz:
		return "razz"
	case StudHigh:
		return "7card_stud_high"
	case StudHighLow8:
		return "7card_stud_highlow8"
	default:
		return "unknown"
	}
}