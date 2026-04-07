package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"equity-distribution-backend/pkg/fileio"
	"equity-distribution-backend/pkg/poker"

	pkr "github.com/chehsunliu/poker"
	"github.com/gin-gonic/gin"
)

// CalculateEquity handles the equity calculation request
func CalculateEquity(c *gin.Context) {
	var req EquityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Validate and parse user's hand
	userHand, err := parseHand(req.Hand)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_hand",
			Message: "Invalid hand format: " + err.Error(),
		})
		return
	}

	// Handle opponent range based on UsePreset flag
	var opponentHands [][]pkr.Card
	
	if req.UsePreset {
		// Load opponent range from preset
		if req.OpponentPreset == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "empty_preset",
				Message: "Preset name cannot be empty when use_preset is true",
			})
			return
		}
		
		// Validate preset name
		if !IsValidPreset(req.OpponentPreset) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_preset",
				Message: "Invalid preset name. Valid presets: " + strings.Join(ValidPresets(), ", "),
			})
			return
		}
		
		// Get data directory (relative to executable)
		dataDir := filepath.Join(".", "data")
		
		// Load range from preset
		rangeStr, err := fileio.LoadOpponentRangeFromPreset(req.OpponentPreset, dataDir)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_preset",
				Message: "Failed to load preset '" + req.OpponentPreset + "': " + err.Error(),
			})
			return
		}
		
		// Parse the loaded range
		hands := strings.Split(rangeStr, ",")
		for _, handStr := range hands {
			hand, err := parseHand(strings.TrimSpace(handStr))
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "invalid_preset_hand",
					Message: "Invalid hand in preset '" + handStr + "': " + err.Error(),
				})
				return
			}
			opponentHands = append(opponentHands, hand)
		}
	} else {
		// Use provided opponent range array
		if len(req.OpponentRange) == 0 {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "empty_range",
				Message: "Opponent range cannot be empty when use_preset is false",
			})
			return
		}

		for _, handStr := range req.OpponentRange {
			hand, err := parseHand(handStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Error:   "invalid_opponent_hand",
					Message: "Invalid opponent hand '" + handStr + "': " + err.Error(),
				})
				return
			}
			opponentHands = append(opponentHands, hand)
		}
	}

	// Parse board cards (required for postflop)
	if len(req.Board) < 3 || len(req.Board) > 5 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_board",
			Message: "Board must contain 3-5 cards for postflop calculation",
		})
		return
	}
	
	var board []pkr.Card
	for _, cardStr := range req.Board {
		card, err := parseCard(cardStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_board",
				Message: "Invalid board card '" + cardStr + "': " + err.Error(),
			})
			return
		}
		board = append(board, card)
	}


	// Detect game type from hand length
	gameType := "PLO4"
	if len(userHand) == 5 {
		gameType = "PLO5"
	}
	
	// Validate all opponent hands have the same card count
	for _, oppHand := range opponentHands {
		if len(oppHand) != len(userHand) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "mixed_game_types",
				Message: fmt.Sprintf("All hands must have the same number of cards. User has %d cards but opponent has %d cards", len(userHand), len(oppHand)),
			})
			return
		}
	}
	
	// Perform equity calculation with adaptive precision
	startTime := time.Now()
	response, err := calculateEquityWithAdaptivePrecision(userHand, opponentHands, board, gameType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "calculation_error",
			Message: "Error calculating equity: " + err.Error(),
		})
		return
	}
	
	// Add calculation time
	response.Calculations.CalculationTimeMs = time.Since(startTime).Milliseconds()

	c.JSON(http.StatusOK, response)
}

// parseHand parses a hand string (e.g., "AhKdQcJs" for PLO4 or "AhKdQcJsTd" for PLO5) into poker cards
func parseHand(handStr string) ([]pkr.Card, error) {
	// Check for PLO4 (8 chars) or PLO5 (10 chars)
	if len(handStr) != 8 && len(handStr) != 10 {
		return nil, fmt.Errorf("hand must be 8 characters (PLO4) or 10 characters (PLO5), got %d", len(handStr))
	}

	var cards []pkr.Card
	for i := 0; i < len(handStr); i += 2 {
		cardStr := handStr[i : i+2]
		card, err := parseCard(cardStr)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	
	return cards, nil
}

// parseCard parses a single card string (e.g., "Ah") into a poker card
func parseCard(cardStr string) (pkr.Card, error) {
	if len(cardStr) != 2 {
		var emptyCard pkr.Card
		return emptyCard, fmt.Errorf("card must be 2 characters, got %d", len(cardStr))
	}
	
	// Validate card format (rank + suit)
	rank := cardStr[0]
	suit := cardStr[1]
	
	// Validate rank
	validRanks := "23456789TJQKA"
	rankValid := false
	for _, r := range validRanks {
		if rank == byte(r) {
			rankValid = true
			break
		}
	}
	
	// Validate suit
	validSuits := "hdcs"
	suitValid := false
	for _, s := range validSuits {
		if suit == byte(s) {
			suitValid = true
			break
		}
	}
	
	if !rankValid || !suitValid {
		var emptyCard pkr.Card
		return emptyCard, fmt.Errorf("invalid card format '%s'. Expected format: rank[2-9TJQKA] + suit[hdcs]", cardStr)
	}
	
	// Create card using poker package
	card := pkr.NewCard(cardStr)
	return card, nil
}


// calculateEquityWithAdaptivePrecision performs equity calculation with adaptive precision control
func calculateEquityWithAdaptivePrecision(userHand []pkr.Card, opponentHands [][]pkr.Card, board []pkr.Card, gameType string) (*EquityResponse, error) {
	// Use adaptive sampling for equity calculation
	config := poker.DefaultAdaptiveConfig()
	equities, avgEquity, samplesUsed, err := poker.CalculateHandVsRangeAdaptiveWithDetails(userHand, opponentHands, board, config)
	if err != nil {
		return nil, err
	}

	// Build equity graph
	var equityGraph []EquityDataPoint
	for handStr, equity := range equities {
		if equity >= 0 { // Skip invalid hands (equity = -1)
			equityGraph = append(equityGraph, EquityDataPoint{
				Hand:   handStr,
				Equity: equity,
			})
		}
	}

	// Use the calculated average equity
	userEquity := avgEquity

	// Create response
	response := &EquityResponse{
		UserEquity:     userEquity,
		TotalHands:     len(opponentHands),
		GameType:       gameType,
		EquityGraph:    equityGraph,
		Calculations: CalculationDetails{
			ConvergedAt:      samplesUsed,
			TotalIterations:  samplesUsed,
			ConfidenceLevel:  95.0, // 95% confidence interval from default config
			CalculationTimeMs: 0,   // Will be set by caller
		},
	}

	return response, nil
}

