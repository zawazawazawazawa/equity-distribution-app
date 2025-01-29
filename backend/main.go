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
	return 0.5

	// // Generate a deck of 52 cards
	// fullDeck := []string{
	// 	"2c", "2d", "2h", "2s",
	// 	"3c", "3d", "3h", "3s",
	// 	"4c", "4d", "4h", "4s",
	// 	"5c", "5d", "5h", "5s",
	// 	"6c", "6d", "6h", "6s",
	// 	"7c", "7d", "7h", "7s",
	// 	"8c", "8d", "8h", "8s",
	// 	"9c", "9d", "9h", "9s",
	// 	"Tc", "Td", "Th", "Ts",
	// 	"Jc", "Jd", "Jh", "Js",
	// 	"Qc", "Qd", "Qh", "Qs",
	// 	"Kc", "Kd", "Kh", "Ks",
	// 	"Ac", "Ad", "Ah", "As",
	// }

	// // Remove cards that are in yourHand, opponentsHand, and board
	// usedCards := strings.Split(yourHand+opponentsHand+board, "")
	// remainingDeck := []string{}
	// for _, card := range fullDeck {
	// 	if !strings.Contains(yourHand+opponentsHand+board, card) {
	// 		remainingDeck = append(remainingDeck, card)
	// 	}
	// }

	// // Calculate win probability
	// winCount := 0
	// totalCount := 0

	// for _, turn := range turnAndRiverCards {
	// 	for _, river := range turnAndRiverCards {
	// 		if turn != river && !strings.Contains(yourHand+opponentsHand+board, turn) && !strings.Contains(yourHand+opponentsHand+board, river) {
	// 			// Simulate the outcome with the current turn and river
	// 			// This is a placeholder for actual hand evaluation logic
	// 			if yourHand > opponentsHand { // Simplified comparison
	// 				winCount++
	// 			}
	// 			totalCount++
	// 		}
	// 	}
	// }

	// // Calculate probability
	// return float64(winCount) / float64(totalCount)
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

	yourHands := strings.Split(requestData.YourHands, ",")
	for i, hand := range yourHands {
		yourHands[i] = strings.Split(hand, "@")[0]
		if len(yourHands[i]) == 8 {
			yourHands[i] = strings.ToUpper(yourHands[i][0:1]) + strings.ToLower(yourHands[i][1:2]) + strings.ToUpper(yourHands[i][2:3]) + strings.ToLower(yourHands[i][3:4]) + strings.ToUpper(yourHands[i][4:5]) + strings.ToLower(yourHands[i][5:6]) + strings.ToUpper(yourHands[i][6:7]) + strings.ToLower(yourHands[i][7:8])
		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	}

	opponentHands := strings.Split(requestData.OpponentsHands, ",")
	for i, hand := range opponentHands {
		opponentHands[i] = strings.Split(hand, "@")[0]
		if len(opponentHands[i]) == 8 {
			opponentHands[i] = strings.ToUpper(opponentHands[i][0:1]) + strings.ToLower(opponentHands[i][1:2]) + strings.ToUpper(opponentHands[i][2:3]) + strings.ToLower(opponentHands[i][3:4]) + strings.ToUpper(opponentHands[i][4:5]) + strings.ToLower(opponentHands[i][5:6]) + strings.ToUpper(opponentHands[i][6:7]) + strings.ToLower(opponentHands[i][7:8])
		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
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
