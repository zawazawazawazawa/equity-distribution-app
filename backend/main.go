package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	// "math/rand"
	// "time"
)

type HandRange struct {
	Hand string `json:"hand"`
}

func calculateHandEquity(hand string, opponentHandRange []string, board string) float64 {
	// Placeholder for actual PLO equity calculation
	totalEquity := 0.0
	for _, opponentHand := range opponentHandRange {
		// Simulate equity calculation with board
		equity := simulatePLOEquity(hand, opponentHand, board)
		totalEquity += equity
	}
	return totalEquity / float64(len(opponentHandRange))
}

func simulatePLOEquity(hand, opponentHand, board string) float64 {
	// Dummy logic for simulating PLO equity
	// Replace with actual simulation logic
	return 0.5 // Dummy value
}

func calculateEquity(handRange string) [][]interface{} {
	hands := strings.Split(handRange, ",")
	var results [][]interface{}
	for _, hand := range hands {
		equity := calculateHandEquity(hand, []string{}, "")
		results = append(results, []interface{}{hand, equity})
	}
	return [][]interface{}{
		{"AsAh6s5h", 0.85},
		{"KsKh6s5h", 0.80},
		{"QsQh6s5h", 0.75},
		{"JsJh6s5h", 0.70},
	}
}

func handleEquityCalculation(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	var hr HandRange
	err := json.NewDecoder(r.Body).Decode(&hr)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Parse the hand range to extract cards
	hands := strings.Split(hr.Hand, ",")
	for i, hand := range hands {
		hands[i] = strings.Split(hand, "@")[0]
	}

	// Calculate equity
	equity := calculateEquity(hr.Hand)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(equity)
}

func main() {
	http.HandleFunc("/calculate-equity", handleEquityCalculation)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Equity Distribution Backend is running")
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
