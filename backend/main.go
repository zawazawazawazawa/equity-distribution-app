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

func calculateHandEquity(yourHand string, opponentsHand string, board string) float64 {	
	// TODO: Simulate PLO equity calculation for your hand against opponent's hand

	// response dummy value
	return 0.85
}

func calculateEquity(yourHands []string, OpponentsHands []string) [][]interface{} {
	var results [][]interface{}
	for _, yourHand := range yourHands {
		totalEquity := 0.0

		for _, opponentsHand := range OpponentsHands {

			// print yourHand and opponentsHand
			log.Printf("Your Hand: %s", yourHand)
			log.Printf("Opponents Hand: %s", opponentsHand)

			// define the board
			board := "2c3d4h"

			equity := calculateHandEquity(yourHand, opponentsHand, board)
			totalEquity += equity
		}

		averageEquity := totalEquity / float64(len(OpponentsHands))
		results = append(results, []interface{}{yourHand, averageEquity})
	}

	return results
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

	var requestData struct {
		YourHands      string `json:"yourHands"`
		OpponentsHands string `json:"opponentsHands"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Parse the hand range to extract cards
	yourHands := strings.Split(requestData.YourHands, ",")
	for i, hand := range yourHands {
		yourHands[i] = strings.Split(hand, "@")[0]
	}

	opponentHands := strings.Split(requestData.OpponentsHands, ",")
	for i, hand := range opponentHands {
		opponentHands[i] = strings.Split(hand, "@")[0]
	}

	equity := calculateEquity(yourHands, opponentHands)

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
