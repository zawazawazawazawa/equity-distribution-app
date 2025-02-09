package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/chehsunliu/poker"
	"github.com/joho/godotenv"
)

// HandRange represents a user's hand
type HandRange struct {
	Hand string `json:"hand"`
}

// YourDynamoDBItem represents an item in DynamoDB
type YourDynamoDBItem struct {
	Equity float64 `json:"Equity"`
	// Add more fields as needed
}

// FlopEquities represents all equity calculations for a specific flop
type FlopEquities struct {
	Flop     string
	Equities map[string]float64 // handCombination -> equity
}

// getDynamoDBClient initializes and returns a DynamoDB client
func getDynamoDBClient() *dynamodb.DynamoDB {
	// AWS設定
	config := &aws.Config{
		Region:   aws.String("us-east-1"),             // LocalStackのデフォルトリージョン
		Endpoint: aws.String("http://localhost:4566"), // LocalStackのエンドポイント
	}

	// 認証情報を設定（LocalStackの場合はダミーでOK）
	config.Credentials = credentials.NewStaticCredentials("test", "test", "")

	// セッションを作成
	sess := session.Must(session.NewSession(config))

	// DynamoDBクライアントを作成
	return dynamodb.New(sess)
}

// batchQueryDynamoDB retrieves all equity calculations for a specific flop
func batchQueryDynamoDB(flop string) (*FlopEquities, error) {
	svc := getDynamoDBClient()

	// Query parameters for scanning items with the same flop
	log.Printf("Querying DynamoDB for flop: %s", flop)
	input := &dynamodb.QueryInput{
		TableName:              aws.String("PloEquity"),
		KeyConditionExpression: aws.String("Flop = :flop"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":flop": {
				S: aws.String(flop),
			},
		},
	}

	result, err := svc.Query(input)
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB: %v", err)
	}

	// Create FlopEquities instance
	flopEquities := &FlopEquities{
		Flop:     flop,
		Equities: make(map[string]float64),
	}

	// Unmarshal each item
	for _, item := range result.Items {
		var dbItem struct {
			HandCombination string  `json:"HandCombination"`
			Equity          float64 `json:"Equity"`
		}
		err = dynamodbattribute.UnmarshalMap(item, &dbItem)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal DynamoDB item: %v", err)
		}
		flopEquities.Equities[dbItem.HandCombination] = dbItem.Equity
	}

	return flopEquities, nil
}

// insertDynamoDB inserts or updates an item in DynamoDB
func insertDynamoDB(flop string, handCombination string, equity float64) error {
	log.Printf("Attempting to insert data - Flop: %s, HandCombination: %s, Equity: %.2f", flop, handCombination, equity)
	svc := getDynamoDBClient()
	item := map[string]*dynamodb.AttributeValue{
		"Flop": {
			S: aws.String(flop),
		},
		"HandCombination": {
			S: aws.String(handCombination),
		},
		"Equity": {
			N: aws.String(fmt.Sprintf("%.2f", equity)),
		},
	}

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String("PloEquity"),
	}

	result, err := svc.PutItem(input)
	if err != nil {
		log.Printf("Error inserting data into DynamoDB: %v", err)
		return fmt.Errorf("failed to insert data into DynamoDB: %v", err)
	}
	log.Printf("Successfully inserted data into DynamoDB: %v", result)
	return nil
}

// generateBoardString creates a string representation of the board cards
func generateBoardString(board []poker.Card) string {
	boardStr := ""
	for _, card := range board {
		boardStr += card.String()
	}
	return boardStr
}

// generateHandCombination creates a unique combination key for the hands
func generateHandCombination(heroHand string, villainHand string) string {
	hands := []string{heroHand, villainHand}
	sort.Strings(hands) // Sort alphabetically to ensure uniqueness
	return fmt.Sprintf("%s_%s", hands[0], hands[1])
}

// hasCardDuplicates checks if there are any duplicate cards across all provided card arrays
func hasCardDuplicates(cards ...[]poker.Card) bool {
	seen := make(map[string]bool)
	for _, hand := range cards {
		for _, card := range hand {
			cardStr := card.String()
			if seen[cardStr] {
				return true
			}
			seen[cardStr] = true
		}
	}
	return false
}

// calculateHandVsHandEquity calculates the equity between two hands
func calculateHandVsHandEquity(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card, flopEquities *FlopEquities) float64 {
	// Check for duplicate cards
	if hasCardDuplicates(yourHand, opponentHand, board) {
		return -1 // Return -1 to indicate invalid hand due to duplicate cards
	}

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

	// Convert hands to strings and concatenate
	heroHandStr := ""
	for _, card := range yourHand {
		heroHandStr += card.String()
	}
	villainHandStr := ""
	for _, card := range opponentHand {
		villainHandStr += card.String()
	}
	// Generate key for equity lookup
	handCombination := generateHandCombination(heroHandStr, villainHandStr)

	// Check if equity exists in memory
	if flopEquities != nil {
		if equity, exists := flopEquities.Equities[handCombination]; exists {
			log.Printf("Found cached equity for combination %s: %.2f", handCombination, equity)
			return equity
		}
	}

	// Calculate equity since it wasn't found in DynamoDB
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
			}
			totalOutcomes += 1
		}
	}

	calculatedEquity := (winCount / totalOutcomes) * 100
	return calculatedEquity
}

// judgeWinner determines the winner between two hands
func judgeWinner(yourHand []poker.Card, opponentHand []poker.Card, board []poker.Card) string {
	// @doc: https://github.com/chehsunliu/poker/blob/72fcd0dd66288388735cc494db3f2bd11b28bfed/lookup.go#L12
	var maxYourHandRank int32 = 7462
	var maxOpponentHandRank int32 = 7462

	// Generate all combinations of your hand and board
	for i := 0; i < len(yourHand); i++ {
		for j := i + 1; j < len(yourHand); j++ {
			newBoard := append(board, yourHand[i], yourHand[j])
			yourHandRank := poker.Evaluate(newBoard)
			if yourHandRank < maxYourHandRank {
				maxYourHandRank = yourHandRank
			}
		}
	}

	// Generate all combinations of opponent's hand and board
	for i := 0; i < len(opponentHand); i++ {
		for j := i + 1; j < len(opponentHand); j++ {
			newBoard := append(board, opponentHand[i], opponentHand[j])
			opponentHandRank := poker.Evaluate(newBoard)
			if opponentHandRank < maxOpponentHandRank {
				maxOpponentHandRank = opponentHandRank
			}
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

// calculateRangeVsRangeEquity calculates equity for ranges of hands
func calculateRangeVsRangeEquity(yourHands [][]poker.Card, opponentHands [][]poker.Card, board []poker.Card) [][]interface{} {
	var results [][]interface{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Set semaphore size to number of CPU cores
	numCPU := runtime.NumCPU()
	semaphore := make(chan struct{}, numCPU)

	// Fetch all equities for this flop at once
	boardStr := generateBoardString(board)
	flopEquities, err := batchQueryDynamoDB(boardStr)
	if err != nil {
		log.Printf("Error fetching flop equities: %v", err)
		flopEquities = &FlopEquities{
			Flop:     boardStr,
			Equities: make(map[string]float64),
		}
	}

	// Create a buffered channel for DynamoDB operations
	type dbOperation struct {
		handCombination string
		equity          float64
	}
	const dbBatchSize = 25 // DynamoDB batch size limit
	dbChan := make(chan dbOperation, dbBatchSize)

	// Start a goroutine to handle DynamoDB batch operations
	var dbWg sync.WaitGroup
	dbWg.Add(1)
	go func() {
		defer dbWg.Done()
		batch := make([]dbOperation, 0, dbBatchSize)
		for op := range dbChan {
			batch = append(batch, op)
			if len(batch) >= dbBatchSize {
				// Process batch
				for _, item := range batch {
					if err := insertDynamoDB(boardStr, item.handCombination, item.equity); err != nil {
						log.Printf("Error inserting equity into DynamoDB: %v", err)
					}
				}
				batch = batch[:0] // Clear batch
			}
		}
		// Process remaining items
		for _, item := range batch {
			if err := insertDynamoDB(boardStr, item.handCombination, item.equity); err != nil {
				log.Printf("Error inserting equity into DynamoDB: %v", err)
			}
		}
	}()

	// Process hands in batches
	const batchSize = 1000
	for i := 0; i < len(yourHands); i += batchSize {
		end := i + batchSize
		if end > len(yourHands) {
			end = len(yourHands)
		}

		batch := yourHands[i:end]
		for _, yourHand := range batch {
			wg.Add(1)
			go func(yourHand []poker.Card) {
				defer wg.Done()
				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				totalEquity := 0.0
				validOpponentCount := 0

				for _, opponentHand := range opponentHands {
					heroHandStr := ""
					for _, card := range yourHand {
						heroHandStr += card.String()
					}
					villainHandStr := ""
					for _, card := range opponentHand {
						villainHandStr += card.String()
					}
					handCombination := generateHandCombination(heroHandStr, villainHandStr)

					equity := calculateHandVsHandEquity(yourHand, opponentHand, board, flopEquities)
					if equity != -1 {
						totalEquity += equity
						validOpponentCount++
						// Send to DynamoDB channel
						dbChan <- dbOperation{
							handCombination: handCombination,
							equity:          equity,
						}
					} else {
						log.Printf("Skipping equity calculation for %s vs %s due to duplicate cards", heroHandStr, villainHandStr)
					}
				}

				var averageEquity float64
				if validOpponentCount > 0 {
					averageEquity = totalEquity / float64(validOpponentCount)
				} else {
					averageEquity = -1.0
				}

				if averageEquity != -1 {
					mu.Lock()
					results = append(results, []interface{}{yourHand, averageEquity})
					mu.Unlock()
				}
			}(yourHand)
		}

		// Wait for current batch to complete before processing next batch
		wg.Wait()
	}

	// Close DynamoDB channel and wait for remaining operations
	close(dbChan)
	dbWg.Wait()

	return results
}

// handleEquityCalculation handles the equity calculation HTTP request
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
		tmpHand := strings.Split(yourHands[i], "@")[0]
		tempArray := []poker.Card{}
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
		tmpHand := strings.Split(opponentHands[i], "@")[0]
		tempArray := []poker.Card{}
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

	// Create board
	board := []poker.Card{
		poker.NewCard("2h"),
		poker.NewCard("3d"),
		poker.NewCard("4h"),
	}

	equity := calculateRangeVsRangeEquity(formattedYourHands, formattedOpponentHands, board)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(equity)
}

// main initializes the HTTP server
func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	http.HandleFunc("/calculate-equity", handleEquityCalculation)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Equity Distribution Backend is running")
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
