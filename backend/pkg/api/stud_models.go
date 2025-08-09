package api

// StudEquityRequest represents the request for stud game equity calculation
type StudEquityRequest struct {
	YourDownCards     []string `json:"your_down_cards" binding:"required"`    // Your hidden cards
	YourUpCards       []string `json:"your_up_cards,omitempty"`               // Your exposed cards (optional)
	OpponentDownCards []string `json:"opponent_down_cards" binding:"required"` // Opponent's hidden cards
	OpponentUpCards   []string `json:"opponent_up_cards,omitempty"`           // Opponent's exposed cards (optional)
	KnownDeadCards    []string `json:"known_dead_cards,omitempty"`            // Other known cards (folded, etc.)
	GameType          string   `json:"game_type" binding:"required"`           // "razz", "stud_high", or "stud_highlow8"
	Precision         string   `json:"precision,omitempty"`                    // "fast", "normal", "accurate", "very_accurate", "extreme", or "adaptive"
}

// StudEquityResponse represents the response for stud game equity calculation
type StudEquityResponse struct {
	YourEquity       float64                 `json:"your_equity"`
	GameType         string                  `json:"game_type"`
	TotalIterations  int                     `json:"total_iterations"`
	HighLowDetails   *StudHighLowDetails     `json:"highlow_details,omitempty"` // Only for Stud Hi-Lo 8
	CalculationDetails CalculationDetails    `json:"calculation_details"`
}

// StudHighLowDetails provides detailed equity breakdown for Stud Hi-Lo 8 or better
type StudHighLowDetails struct {
	HighEquity  float64 `json:"high_equity"`  // Probability of winning high pot (%)
	LowEquity   float64 `json:"low_equity"`   // Probability of winning low pot (%)
	ScoopEquity float64 `json:"scoop_equity"` // Probability of winning both pots (%)
}

// StudRangeEquityRequest represents the request for stud game equity vs range
type StudRangeEquityRequest struct {
	YourDownCards     []string   `json:"your_down_cards" binding:"required"`    // Your hidden cards
	YourUpCards       []string   `json:"your_up_cards,omitempty"`               // Your exposed cards (optional)
	OpponentRange     []StudHand `json:"opponent_range" binding:"required"`     // List of opponent hands
	KnownDeadCards    []string   `json:"known_dead_cards,omitempty"`            // Other known cards
	GameType          string     `json:"game_type" binding:"required"`           // "razz", "stud_high", or "stud_highlow8"
	Precision         string     `json:"precision,omitempty"`                    // Calculation precision
}

// StudHand represents a hand in the opponent's range
type StudHand struct {
	DownCards []string `json:"down_cards" binding:"required"` // Hidden cards
	UpCards   []string `json:"up_cards,omitempty"`            // Exposed cards
}

// StudRangeEquityResponse represents the response for range equity calculation
type StudRangeEquityResponse struct {
	YourEquity      float64              `json:"your_equity"`
	GameType        string               `json:"game_type"`
	TotalHands      int                  `json:"total_hands"`
	EquityGraph     []StudEquityPoint    `json:"equity_graph"`
	HighLowDetails  *StudHighLowDetails  `json:"highlow_details,omitempty"`
	CalculationDetails CalculationDetails `json:"calculation_details"`
}

// StudEquityPoint represents a single point on the equity graph
type StudEquityPoint struct {
	Hand   StudHand `json:"hand"`   // Opponent hand
	Equity float64  `json:"equity"` // Your equity against this specific hand
}