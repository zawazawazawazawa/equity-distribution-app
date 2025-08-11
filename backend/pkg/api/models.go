package api

// EquityRequest represents the request for equity calculation
type EquityRequest struct {
	Hand              string   `json:"hand" binding:"required"`           // User's hand (e.g., "AhKdQcJs")
	OpponentRange     []string `json:"opponent_range,omitempty"`          // List of opponent hands (used when UsePreset is false)
	OpponentPreset    string   `json:"opponent_preset,omitempty"`         // Preset name for opponent range (used when UsePreset is true)
	UsePreset         bool     `json:"use_preset"`                        // Flag to determine whether to use preset or array
	Board             []string `json:"board" binding:"required"`          // Board cards (required, postflop only)
}

// EquityResponse represents the response containing equity calculations
type EquityResponse struct {
	UserEquity     float64            `json:"user_equity"`
	TotalHands     int                `json:"total_hands"`
	WinCount       int                `json:"win_count"`
	LoseCount      int                `json:"lose_count"`
	TieCount       int                `json:"tie_count"`
	GameType       string             `json:"game_type"`       // "PLO4" or "PLO5"
	EquityGraph    []EquityDataPoint  `json:"equity_graph"`
	Calculations   CalculationDetails `json:"calculation_details"`
}

// EquityDataPoint represents a single point on the equity graph
type EquityDataPoint struct {
	Hand   string  `json:"hand"`   // Opponent hand
	Equity float64 `json:"equity"` // User's equity against this specific hand
}

// CalculationDetails provides metadata about the calculation
type CalculationDetails struct {
	ConvergedAt      int     `json:"converged_at"`
	TotalIterations  int     `json:"total_iterations"`
	ConfidenceLevel  float64 `json:"confidence_level"`
	CalculationTimeMs int64   `json:"calculation_time_ms"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}