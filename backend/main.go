package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/chehsunliu/poker"
)

type HandRange struct {
	Hand string `json:"hand"`
}

func calculateHandEquity(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) float64 {
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

	// usedCards := []string{}
	// combined := yourHand + opponentsHand + board
	// for i := 0; i < len(combined); i += 2 {
	// 	usedCards = append(usedCards, combined[i:i+2])
	// }

	// remainingDeck := []string{}
	// for _, card := range fullDeck {
	// 	if !contains(usedCards, card) {
	// 		remainingDeck = append(remainingDeck, card)
	// 	}
	// }

	// // print the length of remainingDeck
	// log.Printf("Length of remainingDeck: %d", len(remainingDeck))

	// // Calculate win probability
	// winCount := 0.0
	// totalCount := 0

	// // Generate all combinations of two cards from the remainingDeck
	// for i := 0; i < len(remainingDeck); i++ {
	// 	for j := i + 1; j < len(remainingDeck); j++ {
	// 		turn := remainingDeck[i]
	// 		river := remainingDeck[j]

	// 		finalBoard := board + turn + river

	// 		// Evaluate hands according to PLO rules
	// 		yourHandValue := evaluateHand(yourHand, finalBoard)
	// 		opponentsHandValue := evaluateHand(opponentsHand, finalBoard)

	// 		if yourHandValue > opponentsHandValue {
	// 			winCount++
	// 		} else if yourHandValue == opponentsHandValue {
	// 			winCount += 0.5
	// 		}
	// 		totalCount++
	// 	}
	// }

	// Calculate probability
	// return winCount / float64(totalCount)

	return 0.0
}

func evaluateHand(hand string, board string) int {
	// print the hand
	log.Printf("Hand: %s", hand)

	// handCards := strings.Split(hand, "")
	// cards := make([]poker.Card, len(handCards)/2)
	// handRank := poker.Evaluate(cards)
	// return int(handRank)
	return int(1)
}

// Helper function to check if a slice contains a specific element
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func calculateEquity(yourHands [][]poker.Card, OpponentHands [][]poker.Card) [][]interface{} {
	var results [][]interface{}
	for _, yourHand := range yourHands {
		totalEquity := 0.0

		for _, opponentHand := range OpponentHands {

			// print yourHand and opponentsHand
			// log.Printf("Your Hand: %s", yourHand)
			// log.Printf("Opponents Hand: %s", opponentHand)

			// define the board
			board := []poker.Card{}

			equity := calculateHandEquity(yourHand, opponentHand, board)
			totalEquity += equity
		}

		averageEquity := totalEquity / float64(len(OpponentHands))
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

	var formatedYourHands [][]poker.Card

	for i := 0; i < len(yourHands); i += 1 {
		var tmpHand string = strings.Split(yourHands[i], "@")[0]
		var tempArray []poker.Card
		if len(tmpHand) == 8 {
			for j := 0; j < 8; j += 2 {
				tempArray = append(tempArray, poker.NewCard(strings.ToUpper(tmpHand[j:j+1])+strings.ToLower(tmpHand[j+1:j+2])))
			}
		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		formatedYourHands = append(formatedYourHands, tempArray)
	}

	// print the formatedYourHands
	log.Printf("Formated Your Hands: %s", formatedYourHands)

	opponentHands := strings.Split(requestData.OpponentsHands, ",")

	var formatedOpponentHands [][]poker.Card

	for i := 0; i < len(opponentHands); i += 1 {
		var tmpHand string = strings.Split(opponentHands[i], "@")[0]
		var tempArray []poker.Card
		if len(tmpHand) == 8 {
			for j := 0; j < 8; j += 2 {
				tempArray = append(tempArray, poker.NewCard(strings.ToUpper(tmpHand[j:j+1])+strings.ToLower(tmpHand[j+1:j+2])))
			}
		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
		formatedOpponentHands = append(formatedOpponentHands, tempArray)
	}

	// print the formatedOpponentHands
	log.Printf("Formated Opponent Hands: %s", formatedOpponentHands)

	equity := calculateEquity(formatedYourHands, formatedOpponentHands)

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
