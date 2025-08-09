package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCalculateStudEquity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupRoutes(router)

	tests := []struct {
		name           string
		request        StudEquityRequest
		expectedStatus int
		checkResponse  func(t *testing.T, resp StudEquityResponse)
	}{
		{
			name: "Valid Razz request",
			request: StudEquityRequest{
				YourDownCards:     []string{"As", "2d", "3h"},
				YourUpCards:       []string{"4c", "5s", "6h", "7d"},
				OpponentDownCards: []string{"Ks", "Qd", "Jh"},
				OpponentUpCards:   []string{"Tc", "9s", "8d", "8h"},
				GameType:          "razz",
				Precision:         "accurate",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp StudEquityResponse) {
				assert.Equal(t, "razz", resp.GameType)
				expectedEquity := 99.5 // 7-low should dominate
				assert.InDelta(t, expectedEquity, resp.YourEquity, 0.5, "Equity should be within ±0.5% of expected value")
				assert.Nil(t, resp.HighLowDetails)
			},
		},
		{
			name: "Valid Stud High request",
			request: StudEquityRequest{
				YourDownCards:     []string{"As", "Ad"},
				YourUpCards:       []string{"Ah"},
				OpponentDownCards: []string{"Ks", "Qd"},
				OpponentUpCards:   []string{"Jh"},
				GameType:          "stud_high",
				Precision:         "accurate",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp StudEquityResponse) {
				assert.Equal(t, "7card_stud_high", resp.GameType)
				expectedEquity := 92.7 // Trips should be strong (actual calculation result)
				assert.InDelta(t, expectedEquity, resp.YourEquity, 0.5, "Equity should be within ±0.5% of expected value")
			},
		},
		{
			name: "Valid Stud Hi-Lo 8 request",
			request: StudEquityRequest{
				YourDownCards:     []string{"As", "2d"},
				YourUpCards:       []string{"3h", "5c"},
				OpponentDownCards: []string{"Ks", "Kd"},
				OpponentUpCards:   []string{"Qh", "Qc"},
				GameType:          "stud_highlow8",
				Precision:         "accurate",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp StudEquityResponse) {
				assert.Equal(t, "7card_stud_highlow8", resp.GameType)
				assert.NotNil(t, resp.HighLowDetails)
				expectedLowEquity := 75.0 // Should have good low
				assert.InDelta(t, expectedLowEquity, resp.HighLowDetails.LowEquity, 0.5, "Low equity should be within ±0.5% of expected value")
				expectedTotalEquity := 50.9 // Combined equity (actual calculation result)
				assert.InDelta(t, expectedTotalEquity, resp.YourEquity, 0.5, "Total equity should be within ±0.5% of expected value")
			},
		},
		{
			name: "Invalid game type",
			request: StudEquityRequest{
				YourDownCards:     []string{"As", "Kd"},
				OpponentDownCards: []string{"Qs", "Qd"},
				GameType:          "invalid_game",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Duplicate cards",
			request: StudEquityRequest{
				YourDownCards:     []string{"As", "As"}, // Duplicate
				OpponentDownCards: []string{"Ks", "Kd"},
				GameType:          "razz",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid card format",
			request: StudEquityRequest{
				YourDownCards:     []string{"AA", "KK"}, // Invalid format
				OpponentDownCards: []string{"Qs", "Qd"},
				GameType:          "stud_high",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/stud/equity", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp StudEquityResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				if tt.checkResponse != nil {
					tt.checkResponse(t, resp)
				}

				// Common checks for successful responses
				assert.Greater(t, resp.TotalIterations, 0)
				assert.GreaterOrEqual(t, resp.YourEquity, 0.0)
				assert.LessOrEqual(t, resp.YourEquity, 100.0)
			}
		})
	}
}

func TestCalculateStudRangeEquity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupRoutes(router)

	tests := []struct {
		name           string
		request        StudRangeEquityRequest
		expectedStatus int
		checkResponse  func(t *testing.T, resp StudRangeEquityResponse)
	}{
		{
			name: "Valid range request",
			request: StudRangeEquityRequest{
				YourDownCards: []string{"As", "Ad"},
				YourUpCards:   []string{"Ah"},
				OpponentRange: []StudHand{
					{DownCards: []string{"Ks", "Kd"}, UpCards: []string{"Kh"}},
					{DownCards: []string{"Qs", "Qd"}, UpCards: []string{"Qh"}},
					{DownCards: []string{"Js", "Jd"}, UpCards: []string{"Jh"}},
				},
				GameType:  "stud_high",
				Precision: "accurate",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp StudRangeEquityResponse) {
				assert.Equal(t, 3, resp.TotalHands)
				assert.Equal(t, 3, len(resp.EquityGraph))
				expectedEquity := 72.7 // Aces should beat most pairs (actual calculation result)
				assert.InDelta(t, expectedEquity, resp.YourEquity, 0.5, "Range equity should be within ±0.5% of expected value")
			},
		},
		{
			name: "Empty range",
			request: StudRangeEquityRequest{
				YourDownCards: []string{"As", "Ad"},
				OpponentRange: []StudHand{},
				GameType:      "razz",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Mixed valid and invalid hands in range",
			request: StudRangeEquityRequest{
				YourDownCards: []string{"As", "2d", "3h"},
				OpponentRange: []StudHand{
					{DownCards: []string{"Ks", "Kd", "Kh"}},     // Valid
					{DownCards: []string{"XX", "YY", "ZZ"}},     // Invalid (bad rank)
					{DownCards: []string{"Qs", "Qd", "Qh"}},     // Valid
				},
				GameType: "razz",
				Precision: "fast",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp StudRangeEquityResponse) {
				assert.Equal(t, 2, resp.TotalHands) // Only 2 valid hands
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/stud/range-equity", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Logf("Response body: %s", w.Body.String())
			}
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp StudRangeEquityResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)

				if tt.checkResponse != nil {
					tt.checkResponse(t, resp)
				}
			}
		})
	}
}

func TestParseStudGameType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"razz", "razz", false},
		{"RAZZ", "razz", false},
		{"stud_high", "7card_stud_high", false},
		{"7card_stud_high", "7card_stud_high", false},
		{"stud_highlow8", "7card_stud_highlow8", false},
		{"stud8", "7card_stud_highlow8", false},
		{"invalid", "", true},
		{"holdem", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gameType, err := parseStudGameType(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, gameType.String())
			}
		})
	}
}