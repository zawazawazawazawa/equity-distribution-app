package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type HandRange struct {
	Hand string `json:"hand"`
}

func calculateEquity(handRange string) float64 {
	// Dummy equity calculation logic
	// In a real scenario, implement the actual equity calculation
	return 50.0
}

func handleEquityCalculation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
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

	// Calculate equity (dummy value for now)
	equity := calculateEquity(hr.Hand)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"equity": equity})
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
