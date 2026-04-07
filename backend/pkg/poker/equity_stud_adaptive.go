package poker

import (
	"fmt"
	"math/rand"

	"github.com/chehsunliu/poker"
)

func calculateStud8EquityAdaptive(yourHand []poker.Card, opponentHand []poker.Card, deck []poker.Card, config EquityCalculationConfig, rng *rand.Rand) (StudEquityResult, int, error) {
	highWins := 0.0
	lowWins := 0.0
	scoops := 0.0
	highTies := 0.0
	lowTies := 0.0
	totalPotShare := 0.0
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
		total++

		// Check for convergence
		if i >= config.MinIterations && i%config.ConvergenceCheck == 0 {
			// Use actual pot share for convergence check
			currentHighEquity := (highWins + highTies*0.5) / total * 100
			currentLowEquity := (lowWins + lowTies*0.5) / total * 100
			currentOverallEquity := totalPotShare / total * 100
			
			recentResults = append(recentResults, currentOverallEquity)

			if len(recentResults) >= 5 {
				// Calculate standard deviation of recent results
				if standardDeviation(recentResults) < config.TargetPrecision {
					return StudEquityResult{
						Equity: currentOverallEquity,
						Stud8Result: &Stud8EquityResult{
							HighEquity:  currentHighEquity,
							LowEquity:   currentLowEquity,
							ScoopEquity: scoops / total * 100,
						},
						TotalIterations: i + 1,
						GameType:        StudHighLow8,
					}, i + 1, nil
				}
				recentResults = recentResults[1:] // Remove old result
			}
		}
	}

	// Calculate final probabilities and equity
	finalHighEquity := (highWins + highTies*0.5) / total * 100
	finalLowEquity := (lowWins + lowTies*0.5) / total * 100
	finalOverallEquity := totalPotShare / total * 100

	return StudEquityResult{
		Equity: finalOverallEquity,
		Stud8Result: &Stud8EquityResult{
			HighEquity:  finalHighEquity,
			LowEquity:   finalLowEquity,
			ScoopEquity: scoops / total * 100,
		},
		TotalIterations: config.MaxIterations,
		GameType:        StudHighLow8,
	}, config.MaxIterations, nil
}