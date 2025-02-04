package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/chehsunliu/poker"
)

type HandRange struct {
	Hand string `json:"hand"`
}

func calculateHandVsHandEquity(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) float64 {
	// Generate the full deck
	deck := poker.NewDeck()
	fullDeck := deck.Draw(52) // Draw all 52 cards from the deck

	usedCards := append(yourHand, opponentHand...)
	usedCards = append(usedCards, board...)

	remainingDeck := []poker.Card{}
	for _, card := range fullDeck {
		used := false
		for _, usedCard := range usedCards {
			if card == usedCard {
				used = true
				break
			}
		}
		if !used {
			remainingDeck = append(remainingDeck, card)
		}
	}

	// Calculate total number of possible outcomes
	totalOutcomes := 0.0
	winCount := 0.0

	for i := 0; i < len(remainingDeck); i++ {
		for j := i + 1; j < len(remainingDeck); j++ {
			finalBoard := append(board, remainingDeck[i], remainingDeck[j])
			winner := judgeWinner(yourHand, opponentHand, finalBoard)
			if winner == "yourHand" {
				winCount += 1
			} else if winner == "tie" {
				winCount += 0.5
			} else {
				// Do nothing
			}
			totalOutcomes += 1
		}
	}

	// Calculate equity
	equity := winCount / totalOutcomes * 100
	return equity
}

func judgeWinner(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) string {
	// @doc: https://github.com/chehsunliu/poker/blob/72fcd0dd66288388735cc494db3f2bd11b28bfed/lookup.go#L12
	var maxYourHandRank int32 = 7462
	var maxOpponentHandRank int32 = 7462
	var mu sync.Mutex
	var wg sync.WaitGroup

	type result struct {
		yourRank     int32
		opponentRank int32
	}

	results := make(chan result, 60)

	// セマフォを使用して同時に実行されるゴルーチンの数を制限
	semaphore := make(chan struct{}, 8)

	// 手元の4枚から2枚を選ぶ組み合わせを生成
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			// ボードの5枚から3枚を選ぶ組み合わせを生成
			for k := 0; k < 5; k++ {
				for l := k + 1; l < 5; l++ {
					for m := l + 1; m < 5; m++ {
						wg.Add(1)
						go func(i, j, k, l, m int) {
							defer wg.Done()
							// セマフォを取得
							semaphore <- struct{}{}

							// Create a new board
							newBoard := []poker.Card{board[k], board[l], board[m]}

							yourHandRank := poker.Evaluate(append(newBoard, yourHand[i], yourHand[j]))

							opponentHandRank := poker.Evaluate(append(newBoard, opponentHand[i], opponentHand[j]))

							results <- result{
								yourRank:     yourHandRank,
								opponentRank: opponentHandRank,
							}

							// セマフォを解放
							<-semaphore
						}(i, j, k, l, m)
					}
				}
			}
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		if res.yourRank < maxYourHandRank {
			mu.Lock()
			if res.yourRank < maxYourHandRank {
				maxYourHandRank = res.yourRank
			}
			mu.Unlock()
		}

		if res.opponentRank < maxOpponentHandRank {
			mu.Lock()
			if res.opponentRank < maxOpponentHandRank {
				maxOpponentHandRank = res.opponentRank
			}
			mu.Unlock()
		}
	}

	if maxYourHandRank < maxOpponentHandRank {
		return "yourHand"
	} else if maxYourHandRank > maxOpponentHandRank {
		return "opponentHand"
	} else {
		return "tie"
	}
}

func calculateRangeVsRangeEquity(yourHands [][]poker.Card, opponentHands [][]poker.Card) [][]interface{} {
	var results [][]interface{}
	for _, yourHand := range yourHands {
		totalEquity := 0.0

		for _, opponentHand := range opponentHands {
			board := []poker.Card{
				poker.NewCard("2h"),
				poker.NewCard("3d"),
				poker.NewCard("4h"),
			}

			equity := calculateHandVsHandEquity(yourHand, opponentHand, board)

			totalEquity += equity
		}

		averageEquity := totalEquity / float64(len(opponentHands))
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

	var formattedYourHands [][]poker.Card

	for i := 0; i < len(yourHands); i++ {
		var tmpHand string = strings.Split(yourHands[i], "@")[0]
		var tempArray []poker.Card
		if len(tmpHand) == 8 {
			for j := 0; j < 8; j += 2 {
				cardStr := strings.ToUpper(tmpHand[j:j+1]) + strings.ToLower(tmpHand[j+1:j+2])
				tempCard := poker.NewCard(cardStr)
				tempArray = append(tempArray, tempCard)
			}
		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		formattedYourHands = append(formattedYourHands, tempArray)
	}

	opponentHands := strings.Split(requestData.OpponentsHands, ",")

	var formattedOpponentHands [][]poker.Card

	for i := 0; i < len(opponentHands); i++ {
		var tmpHand string = strings.Split(opponentHands[i], "@")[0]
		var tempArray []poker.Card
		if len(tmpHand) == 8 {
			for j := 0; j < 8; j += 2 {
				cardStr := strings.ToUpper(tmpHand[j:j+1]) + strings.ToLower(tmpHand[j+1:j+2])
				tempCard := poker.NewCard(cardStr)
				tempArray = append(tempArray, tempCard)
			}
		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		formattedOpponentHands = append(formattedOpponentHands, tempArray)
	}

	equity := calculateRangeVsRangeEquity(formattedYourHands, formattedOpponentHands)

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
