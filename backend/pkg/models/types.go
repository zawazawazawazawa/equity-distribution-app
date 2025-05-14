package models

// FlopEquities represents all equity calculations for a specific flop
type FlopEquities struct {
	Flop     string
	Equities map[string]float64 // handCombination -> equity
}

// HandVsRangeResult represents the result of a hand vs range equity calculation
type HandVsRangeResult struct {
	OpponentHand string  `json:"opponentHand"`
	Equity       float64 `json:"equity"`
}
