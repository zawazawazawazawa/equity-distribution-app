package poker

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/chehsunliu/poker"
)

// CalculateStudEquity calculates equity for stud games using Monte Carlo simulation
func CalculateStudEquity(yourHand []poker.Card, opponentHand []poker.Card, knownCards []poker.Card, gameType StudGameType, iterations int) (StudEquityResult, error) {
	// Validate hands
	if len(yourHand) < 3 || len(yourHand) > 7 {
		return StudEquityResult{}, fmt.Errorf("invalid hand size: %d (must be 3-7 cards)", len(yourHand))
	}
	if len(opponentHand) < 3 || len(opponentHand) > 7 {
		return StudEquityResult{}, fmt.Errorf("invalid opponent hand size: %d (must be 3-7 cards)", len(opponentHand))
	}

	// Check for duplicate cards
	allCards := make([]poker.Card, 0)
	allCards = append(allCards, yourHand...)
	allCards = append(allCards, opponentHand...)
	allCards = append(allCards, knownCards...)
	
	cardSet := make(map[poker.Card]bool)
	for _, card := range allCards {
		if cardSet[card] {
			return StudEquityResult{}, fmt.Errorf("invalid card configuration: duplicate cards detected")
		}
		cardSet[card] = true
	}

	// Create deck excluding known cards
	deck := createDeck()
	deck = removeCards(deck, allCards)

	// Initialize random number generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch gameType {
	case StudRazz:
		return calculateRazzEquity(yourHand, opponentHand, deck, iterations, rng)
	case StudHigh:
		return calculateStudHighEquity(yourHand, opponentHand, deck, iterations, rng)
	case StudHighLow8:
		return calculateStud8Equity(yourHand, opponentHand, deck, iterations, rng)
	default:
		return StudEquityResult{}, fmt.Errorf("unknown game type: %v", gameType)
	}
}

// calculateRazzEquity calculates equity for Razz
func calculateRazzEquity(yourHand []poker.Card, opponentHand []poker.Card, deck []poker.Card, iterations int, rng *rand.Rand) (StudEquityResult, error) {
	wins := 0
	ties := 0

	// Calculate how many cards each player needs
	yourNeeds := 7 - len(yourHand)
	opponentNeeds := 7 - len(opponentHand)
	totalNeeded := yourNeeds + opponentNeeds

	if totalNeeded > len(deck) {
		return StudEquityResult{}, fmt.Errorf("not enough cards in deck")
	}

	for i := 0; i < iterations; i++ {
		// Shuffle and deal remaining cards
		shuffledDeck := make([]poker.Card, len(deck))
		copy(shuffledDeck, deck)
		rng.Shuffle(len(shuffledDeck), func(i, j int) {
			shuffledDeck[i], shuffledDeck[j] = shuffledDeck[j], shuffledDeck[i]
		})

		// Complete hands
		yourComplete := make([]poker.Card, len(yourHand))
		copy(yourComplete, yourHand)
		yourComplete = append(yourComplete, shuffledDeck[:yourNeeds]...)

		opponentComplete := make([]poker.Card, len(opponentHand))
		copy(opponentComplete, opponentHand)
		opponentComplete = append(opponentComplete, shuffledDeck[yourNeeds:yourNeeds+opponentNeeds]...)

		// Determine winner
		winner := JudgeWinnerRazz(yourComplete, opponentComplete)
		switch winner {
		case "yourHand":
			wins++
		case "tie":
			ties++
		}
	}

	equity := float64(wins) + float64(ties)*0.5
	equity = equity / float64(iterations) * 100

	return StudEquityResult{
		Equity:          equity,
		TotalIterations: iterations,
		GameType:        StudRazz,
	}, nil
}

// calculateStudHighEquity calculates equity for 7-card Stud High
func calculateStudHighEquity(yourHand []poker.Card, opponentHand []poker.Card, deck []poker.Card, iterations int, rng *rand.Rand) (StudEquityResult, error) {
	wins := 0
	ties := 0

	// Calculate how many cards each player needs
	yourNeeds := 7 - len(yourHand)
	opponentNeeds := 7 - len(opponentHand)
	totalNeeded := yourNeeds + opponentNeeds

	if totalNeeded > len(deck) {
		return StudEquityResult{}, fmt.Errorf("not enough cards in deck")
	}

	for i := 0; i < iterations; i++ {
		// Shuffle and deal remaining cards
		shuffledDeck := make([]poker.Card, len(deck))
		copy(shuffledDeck, deck)
		rng.Shuffle(len(shuffledDeck), func(i, j int) {
			shuffledDeck[i], shuffledDeck[j] = shuffledDeck[j], shuffledDeck[i]
		})

		// Complete hands
		yourComplete := make([]poker.Card, len(yourHand))
		copy(yourComplete, yourHand)
		yourComplete = append(yourComplete, shuffledDeck[:yourNeeds]...)

		opponentComplete := make([]poker.Card, len(opponentHand))
		copy(opponentComplete, opponentHand)
		opponentComplete = append(opponentComplete, shuffledDeck[yourNeeds:yourNeeds+opponentNeeds]...)

		// Determine winner
		winner := JudgeWinnerStudHigh(yourComplete, opponentComplete)
		switch winner {
		case "yourHand":
			wins++
		case "tie":
			ties++
		}
	}

	equity := float64(wins) + float64(ties)*0.5
	equity = equity / float64(iterations) * 100

	return StudEquityResult{
		Equity:          equity,
		TotalIterations: iterations,
		GameType:        StudHigh,
	}, nil
}

// calculateStud8Equity calculates equity for 7-card Stud Hi-Lo 8 or better
func calculateStud8Equity(yourHand []poker.Card, opponentHand []poker.Card, deck []poker.Card, iterations int, rng *rand.Rand) (StudEquityResult, error) {
	highWins := 0
	lowWins := 0
	scoops := 0
	highTies := 0
	lowTies := 0
	totalPotShare := 0.0

	// Calculate how many cards each player needs
	yourNeeds := 7 - len(yourHand)
	opponentNeeds := 7 - len(opponentHand)
	totalNeeded := yourNeeds + opponentNeeds

	if totalNeeded > len(deck) {
		return StudEquityResult{}, fmt.Errorf("not enough cards in deck")
	}

	for i := 0; i < iterations; i++ {
		// Shuffle and deal remaining cards
		shuffledDeck := make([]poker.Card, len(deck))
		copy(shuffledDeck, deck)
		rng.Shuffle(len(shuffledDeck), func(i, j int) {
			shuffledDeck[i], shuffledDeck[j] = shuffledDeck[j], shuffledDeck[i]
		})

		// Complete hands
		yourComplete := make([]poker.Card, len(yourHand))
		copy(yourComplete, yourHand)
		yourComplete = append(yourComplete, shuffledDeck[:yourNeeds]...)

		opponentComplete := make([]poker.Card, len(opponentHand))
		copy(opponentComplete, opponentHand)
		opponentComplete = append(opponentComplete, shuffledDeck[yourNeeds:yourNeeds+opponentNeeds]...)

		// Determine winners
		highWinner, lowWinner := JudgeWinnerStud8(yourComplete, opponentComplete)

		// Track results
		yourHighWin := false
		yourLowWin := false
		yourHighTie := false
		yourLowTie := false
		iterationPotShare := 0.0

		switch highWinner {
		case "yourHand":
			highWins++
			yourHighWin = true
		case "tie":
			highTies++
			yourHighTie = true
		}

		switch lowWinner {
		case "yourHand":
			lowWins++
			yourLowWin = true
		case "tie":
			lowTies++
			yourLowTie = true
		case "none":
			// No qualifying low - high winner takes all
		}

		// Calculate pot share for this iteration
		if lowWinner == "none" {
			// No qualifying low - high winner takes entire pot
			if yourHighWin {
				iterationPotShare = 1.0
				scoops++
			} else if yourHighTie {
				iterationPotShare = 0.5
			}
		} else {
			// Pot is split between high and low
			if yourHighWin {
				iterationPotShare += 0.5
			} else if yourHighTie {
				iterationPotShare += 0.25
			}
			
			if yourLowWin {
				iterationPotShare += 0.5
			} else if yourLowTie {
				iterationPotShare += 0.25
			}
			
			// Check for scoop (winning both high and low)
			if yourHighWin && yourLowWin {
				scoops++
			}
		}
		
		totalPotShare += iterationPotShare
	}

	// Calculate probabilities as percentages
	highEquity := (float64(highWins) + float64(highTies)*0.5) / float64(iterations) * 100
	lowEquity := (float64(lowWins) + float64(lowTies)*0.5) / float64(iterations) * 100
	scoopEquity := float64(scoops) / float64(iterations) * 100
	
	// Overall equity is the actual pot share
	totalEquity := totalPotShare / float64(iterations) * 100

	return StudEquityResult{
		Equity: totalEquity,
		Stud8Result: &Stud8EquityResult{
			HighEquity:  highEquity,
			LowEquity:   lowEquity,
			ScoopEquity: scoopEquity,
		},
		TotalIterations: iterations,
		GameType:        StudHighLow8,
	}, nil
}

// Wrapper functions for specific game types

// CalculateRazzEquity calculates equity for Razz
func CalculateRazzEquity(yourHand []poker.Card, opponentHand []poker.Card, knownCards []poker.Card, iterations int) (float64, error) {
	result, err := CalculateStudEquity(yourHand, opponentHand, knownCards, StudRazz, iterations)
	if err != nil {
		return 0, err
	}
	return result.Equity, nil
}

// CalculateStudHighEquity calculates equity for 7-card Stud High
func CalculateStudHighEquity(yourHand []poker.Card, opponentHand []poker.Card, knownCards []poker.Card, iterations int) (float64, error) {
	result, err := CalculateStudEquity(yourHand, opponentHand, knownCards, StudHigh, iterations)
	if err != nil {
		return 0, err
	}
	return result.Equity, nil
}

// CalculateStud8Equity calculates equity for 7-card Stud Hi-Lo 8 or better
func CalculateStud8Equity(yourHand []poker.Card, opponentHand []poker.Card, knownCards []poker.Card, iterations int) (StudEquityResult, error) {
	return CalculateStudEquity(yourHand, opponentHand, knownCards, StudHighLow8, iterations)
}

// Helper function to create a standard 52-card deck
func createDeck() []poker.Card {
	suits := []string{"s", "h", "d", "c"}
	ranks := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K"}
	
	deck := make([]poker.Card, 0, 52)
	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, poker.NewCard(rank+suit))
		}
	}
	return deck
}

// Helper function to remove cards from deck
func removeCards(deck []poker.Card, cardsToRemove []poker.Card) []poker.Card {
	result := make([]poker.Card, 0, len(deck))
	
	for _, card := range deck {
		found := false
		for _, removeCard := range cardsToRemove {
			if card == removeCard {
				found = true
				break
			}
		}
		if !found {
			result = append(result, card)
		}
	}
	
	return result
}

// CalculateStudEquityAdaptive calculates equity for stud games with adaptive sampling
func CalculateStudEquityAdaptive(yourHand []poker.Card, opponentHand []poker.Card, knownCards []poker.Card, gameType StudGameType, config EquityCalculationConfig) (StudEquityResult, int, error) {
	// Validate hands
	if len(yourHand) < 3 || len(yourHand) > 7 {
		return StudEquityResult{}, 0, fmt.Errorf("invalid hand size: %d (must be 3-7 cards)", len(yourHand))
	}
	if len(opponentHand) < 3 || len(opponentHand) > 7 {
		return StudEquityResult{}, 0, fmt.Errorf("invalid opponent hand size: %d (must be 3-7 cards)", len(opponentHand))
	}

	// Check for duplicate cards
	allCards := make([]poker.Card, 0)
	allCards = append(allCards, yourHand...)
	allCards = append(allCards, opponentHand...)
	allCards = append(allCards, knownCards...)
	
	cardSet := make(map[poker.Card]bool)
	for _, card := range allCards {
		if cardSet[card] {
			return StudEquityResult{}, 0, fmt.Errorf("invalid card configuration: duplicate cards detected")
		}
		cardSet[card] = true
	}

	// Create deck excluding known cards
	deck := createDeck()
	deck = removeCards(deck, allCards)

	// Initialize random number generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch gameType {
	case StudRazz:
		return calculateRazzEquityAdaptive(yourHand, opponentHand, deck, config, rng)
	case StudHigh:
		return calculateStudHighEquityAdaptive(yourHand, opponentHand, deck, config, rng)
	case StudHighLow8:
		return calculateStud8EquityAdaptive(yourHand, opponentHand, deck, config, rng)
	default:
		return StudEquityResult{}, 0, fmt.Errorf("unknown game type: %v", gameType)
	}
}

// calculateRazzEquityAdaptive calculates Razz equity with adaptive sampling
func calculateRazzEquityAdaptive(yourHand []poker.Card, opponentHand []poker.Card, deck []poker.Card, config EquityCalculationConfig, rng *rand.Rand) (StudEquityResult, int, error) {
	wins := 0.0
	ties := 0.0
	total := 0.0
	var recentResults []float64

	// Calculate how many cards each player needs
	yourNeeds := 7 - len(yourHand)
	opponentNeeds := 7 - len(opponentHand)
	totalNeeded := yourNeeds + opponentNeeds

	if totalNeeded > len(deck) {
		return StudEquityResult{}, 0, fmt.Errorf("not enough cards in deck")
	}

	for i := 0; i < config.MaxIterations; i++ {
		// Shuffle and deal remaining cards
		shuffledDeck := make([]poker.Card, len(deck))
		copy(shuffledDeck, deck)
		rng.Shuffle(len(shuffledDeck), func(i, j int) {
			shuffledDeck[i], shuffledDeck[j] = shuffledDeck[j], shuffledDeck[i]
		})

		// Complete hands
		yourComplete := make([]poker.Card, len(yourHand))
		copy(yourComplete, yourHand)
		yourComplete = append(yourComplete, shuffledDeck[:yourNeeds]...)

		opponentComplete := make([]poker.Card, len(opponentHand))
		copy(opponentComplete, opponentHand)
		opponentComplete = append(opponentComplete, shuffledDeck[yourNeeds:yourNeeds+opponentNeeds]...)

		// Determine winner
		winner := JudgeWinnerRazz(yourComplete, opponentComplete)
		switch winner {
		case "yourHand":
			wins++
		case "tie":
			ties++
		}
		total++

		// Check for convergence
		if i >= config.MinIterations && i%config.ConvergenceCheck == 0 {
			currentEquity := (wins + ties*0.5) / total * 100
			recentResults = append(recentResults, currentEquity)

			if len(recentResults) >= 5 {
				// Calculate standard deviation of recent results
				if standardDeviation(recentResults) < config.TargetPrecision {
					return StudEquityResult{
						Equity:          currentEquity,
						TotalIterations: i + 1,
						GameType:        StudRazz,
					}, i + 1, nil
				}
				recentResults = recentResults[1:] // Remove old result
			}
		}
	}

	finalEquity := (wins + ties*0.5) / total * 100
	return StudEquityResult{
		Equity:          finalEquity,
		TotalIterations: config.MaxIterations,
		GameType:        StudRazz,
	}, config.MaxIterations, nil
}

// calculateStudHighEquityAdaptive calculates Stud High equity with adaptive sampling
func calculateStudHighEquityAdaptive(yourHand []poker.Card, opponentHand []poker.Card, deck []poker.Card, config EquityCalculationConfig, rng *rand.Rand) (StudEquityResult, int, error) {
	wins := 0.0
	ties := 0.0
	total := 0.0
	var recentResults []float64

	// Calculate how many cards each player needs
	yourNeeds := 7 - len(yourHand)
	opponentNeeds := 7 - len(opponentHand)
	totalNeeded := yourNeeds + opponentNeeds

	if totalNeeded > len(deck) {
		return StudEquityResult{}, 0, fmt.Errorf("not enough cards in deck")
	}

	for i := 0; i < config.MaxIterations; i++ {
		// Shuffle and deal remaining cards
		shuffledDeck := make([]poker.Card, len(deck))
		copy(shuffledDeck, deck)
		rng.Shuffle(len(shuffledDeck), func(i, j int) {
			shuffledDeck[i], shuffledDeck[j] = shuffledDeck[j], shuffledDeck[i]
		})

		// Complete hands
		yourComplete := make([]poker.Card, len(yourHand))
		copy(yourComplete, yourHand)
		yourComplete = append(yourComplete, shuffledDeck[:yourNeeds]...)

		opponentComplete := make([]poker.Card, len(opponentHand))
		copy(opponentComplete, opponentHand)
		opponentComplete = append(opponentComplete, shuffledDeck[yourNeeds:yourNeeds+opponentNeeds]...)

		// Determine winner
		winner := JudgeWinnerStudHigh(yourComplete, opponentComplete)
		switch winner {
		case "yourHand":
			wins++
		case "tie":
			ties++
		}
		total++

		// Check for convergence
		if i >= config.MinIterations && i%config.ConvergenceCheck == 0 {
			currentEquity := (wins + ties*0.5) / total * 100
			recentResults = append(recentResults, currentEquity)

			if len(recentResults) >= 5 {
				// Calculate standard deviation of recent results
				if standardDeviation(recentResults) < config.TargetPrecision {
					return StudEquityResult{
						Equity:          currentEquity,
						TotalIterations: i + 1,
						GameType:        StudHigh,
					}, i + 1, nil
				}
				recentResults = recentResults[1:] // Remove old result
			}
		}
	}

	finalEquity := (wins + ties*0.5) / total * 100
	return StudEquityResult{
		Equity:          finalEquity,
		TotalIterations: config.MaxIterations,
		GameType:        StudHigh,
	}, config.MaxIterations, nil
}

// calculateStud8EquityAdaptive calculates Stud8 equity with adaptive sampling
// calculateStud8EquityAdaptive is now in equity_stud_adaptive.go