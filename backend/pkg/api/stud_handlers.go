package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"equity-distribution-backend/pkg/poker"

	pkr "github.com/chehsunliu/poker"
	"github.com/gin-gonic/gin"
)

// CalculateStudEquity handles stud game equity calculation requests
func CalculateStudEquity(c *gin.Context) {
	var req StudEquityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate game type
	gameType, err := parseStudGameType(req.GameType)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_game_type",
			Message: err.Error(),
		})
		return
	}

	// Parse cards
	yourDownCards, err := parseCards(req.YourDownCards)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_your_down_cards",
			Message: "Invalid your down cards: " + err.Error(),
		})
		return
	}

	yourUpCards, err := parseCards(req.YourUpCards)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_your_up_cards",
			Message: "Invalid your up cards: " + err.Error(),
		})
		return
	}

	opponentDownCards, err := parseCards(req.OpponentDownCards)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_opponent_down_cards",
			Message: "Invalid opponent down cards: " + err.Error(),
		})
		return
	}

	opponentUpCards, err := parseCards(req.OpponentUpCards)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_opponent_up_cards",
			Message: "Invalid opponent up cards: " + err.Error(),
		})
		return
	}

	knownDeadCards, err := parseCards(req.KnownDeadCards)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_known_dead_cards",
			Message: "Invalid known dead cards: " + err.Error(),
		})
		return
	}

	// Combine down and up cards
	yourHand := append(yourDownCards, yourUpCards...)
	opponentHand := append(opponentDownCards, opponentUpCards...)

	// Calculate equity
	startTime := time.Now()
	var result poker.StudEquityResult
	var iterations int

	// Determine calculation method
	switch strings.ToLower(req.Precision) {
	case "fast":
		result, err = poker.CalculateStudEquity(yourHand, opponentHand, knownDeadCards, gameType, 1000)
		iterations = result.TotalIterations
	case "accurate":
		result, err = poker.CalculateStudEquity(yourHand, opponentHand, knownDeadCards, gameType, 10000)
		iterations = result.TotalIterations
	case "very_accurate":
		result, err = poker.CalculateStudEquity(yourHand, opponentHand, knownDeadCards, gameType, 20000)
		iterations = result.TotalIterations
	case "extreme":
		result, err = poker.CalculateStudEquity(yourHand, opponentHand, knownDeadCards, gameType, 30000)
		iterations = result.TotalIterations
	case "adaptive":
		config := poker.GetDefaultAdaptiveConfig()
		result, iterations, err = poker.CalculateStudEquityAdaptive(yourHand, opponentHand, knownDeadCards, gameType, config)
	default: // "normal" or unspecified
		result, err = poker.CalculateStudEquity(yourHand, opponentHand, knownDeadCards, gameType, 5000)
		iterations = result.TotalIterations
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "calculation_error",
			Message: "Failed to calculate equity: " + err.Error(),
		})
		return
	}

	calculationTime := time.Since(startTime).Milliseconds()

	// Build response
	response := StudEquityResponse{
		YourEquity:      result.Equity,
		GameType:        gameType.String(),
		TotalIterations: iterations,
		CalculationDetails: CalculationDetails{
			ConvergedAt:       iterations,
			TotalIterations:   iterations,
			ConfidenceLevel:   95.0,
			CalculationTimeMs: calculationTime,
		},
	}

	// Add Hi-Lo details if applicable
	if gameType == poker.StudHighLow8 && result.Stud8Result != nil {
		response.HighLowDetails = &StudHighLowDetails{
			HighEquity:  result.Stud8Result.HighEquity,
			LowEquity:   result.Stud8Result.LowEquity,
			ScoopEquity: result.Stud8Result.ScoopEquity,
		}
	}

	c.JSON(http.StatusOK, response)
}

// CalculateStudRangeEquity handles stud game equity vs range calculation requests
func CalculateStudRangeEquity(c *gin.Context) {
	var req StudRangeEquityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate game type
	gameType, err := parseStudGameType(req.GameType)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_game_type",
			Message: err.Error(),
		})
		return
	}

	// Parse your cards
	yourDownCards, err := parseCards(req.YourDownCards)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_your_down_cards",
			Message: "Invalid your down cards: " + err.Error(),
		})
		return
	}

	yourUpCards, err := parseCards(req.YourUpCards)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_your_up_cards",
			Message: "Invalid your up cards: " + err.Error(),
		})
		return
	}

	knownDeadCards, err := parseCards(req.KnownDeadCards)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_known_dead_cards",
			Message: "Invalid known dead cards: " + err.Error(),
		})
		return
	}

	yourHand := append(yourDownCards, yourUpCards...)

	// Parse opponent range
	var equityPoints []StudEquityPoint
	totalEquity := 0.0
	totalHighEquity := 0.0
	totalLowEquity := 0.0
	totalScoopEquity := 0.0
	validHands := 0

	startTime := time.Now()

	for _, oppHand := range req.OpponentRange {
		oppDownCards, err := parseCards(oppHand.DownCards)
		if err != nil {
			continue // Skip invalid hands
		}

		oppUpCards, err := parseCards(oppHand.UpCards)
		if err != nil {
			continue // Skip invalid hands
		}

		opponentHand := append(oppDownCards, oppUpCards...)

		// Calculate equity against this specific hand
		var result poker.StudEquityResult
		switch strings.ToLower(req.Precision) {
		case "fast":
			result, err = poker.CalculateStudEquity(yourHand, opponentHand, knownDeadCards, gameType, 1000)
		case "accurate":
			result, err = poker.CalculateStudEquity(yourHand, opponentHand, knownDeadCards, gameType, 10000)
		case "very_accurate":
			result, err = poker.CalculateStudEquity(yourHand, opponentHand, knownDeadCards, gameType, 20000)
		case "extreme":
			result, err = poker.CalculateStudEquity(yourHand, opponentHand, knownDeadCards, gameType, 30000)
		default: // "normal"
			result, err = poker.CalculateStudEquity(yourHand, opponentHand, knownDeadCards, gameType, 5000)
		}

		if err != nil {
			continue // Skip hands with errors
		}

		equityPoints = append(equityPoints, StudEquityPoint{
			Hand:   oppHand,
			Equity: result.Equity,
		})

		totalEquity += result.Equity
		if gameType == poker.StudHighLow8 && result.Stud8Result != nil {
			totalHighEquity += result.Stud8Result.HighEquity
			totalLowEquity += result.Stud8Result.LowEquity
			totalScoopEquity += result.Stud8Result.ScoopEquity
		}
		validHands++
	}

	if validHands == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "no_valid_hands",
			Message: "No valid opponent hands in range",
		})
		return
	}

	calculationTime := time.Since(startTime).Milliseconds()

	// Calculate averages
	avgEquity := totalEquity / float64(validHands)

	response := StudRangeEquityResponse{
		YourEquity:   avgEquity,
		GameType:     gameType.String(),
		TotalHands:   validHands,
		EquityGraph:  equityPoints,
		CalculationDetails: CalculationDetails{
			ConvergedAt:       validHands * 5000, // Approximate
			TotalIterations:   validHands * 5000,
			ConfidenceLevel:   95.0,
			CalculationTimeMs: calculationTime,
		},
	}

	// Add Hi-Lo details if applicable
	if gameType == poker.StudHighLow8 {
		response.HighLowDetails = &StudHighLowDetails{
			HighEquity:  totalHighEquity / float64(validHands),
			LowEquity:   totalLowEquity / float64(validHands),
			ScoopEquity: totalScoopEquity / float64(validHands),
		}
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions

func parseStudGameType(gameTypeStr string) (poker.StudGameType, error) {
	switch strings.ToLower(gameTypeStr) {
	case "razz":
		return poker.StudRazz, nil
	case "stud_high", "7card_stud_high":
		return poker.StudHigh, nil
	case "stud_highlow8", "7card_stud_highlow8", "stud8":
		return poker.StudHighLow8, nil
	default:
		return 0, fmt.Errorf("invalid game type: %s (must be 'razz', 'stud_high', or 'stud_highlow8')", gameTypeStr)
	}
}

func parseCards(cardStrs []string) ([]pkr.Card, error) {
	cards := make([]pkr.Card, 0, len(cardStrs))
	for _, cardStr := range cardStrs {
		if cardStr == "" {
			continue
		}
		if len(cardStr) != 2 {
			return nil, fmt.Errorf("invalid card format: %s", cardStr)
		}
		
		// Validate rank and suit
		rank := cardStr[0]
		suit := cardStr[1]
		
		validRanks := "23456789TJQKA"
		validSuits := "shdc"
		
		if !strings.ContainsRune(validRanks, rune(rank)) {
			return nil, fmt.Errorf("invalid rank: %c", rank)
		}
		if !strings.ContainsRune(validSuits, rune(suit)) {
			return nil, fmt.Errorf("invalid suit: %c", suit)
		}
		
		card := pkr.NewCard(cardStr)
		cards = append(cards, card)
	}
	return cards, nil
}

