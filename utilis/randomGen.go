package utilis

import (
	"fmt"
	"math/rand"
)

func GenerateRandomWithFavour(lowerBound, upperBound int, favourableSet [2]int, favourableProbability float64) int {

	if lowerBound > upperBound || favourableProbability < 0 || favourableProbability > 1 {
		fmt.Println("Invalid parameters")
		return 0
	}

	// Calculate total range and the favourable range
	totalRange := upperBound - lowerBound + 1
	favourableRange := favourableSet[1] - favourableSet[0] + 1

	if favourableRange <= 0 || favourableRange > totalRange {
		fmt.Println("Invalid favourable set")
		return 0
	}

	// Check if the favourable set is within the total range
	if favourableSet[0] < lowerBound || favourableSet[1] > upperBound || favourableRange <= 0 {
		fmt.Println("Invalid favourable set")
		return 0
	}

	// Calculate the number of favourable outcomes based on the probability
	favourableOutcomes := int(favourableProbability * float64(totalRange))
	if favourableOutcomes < favourableRange {
		favourableOutcomes = favourableRange
	}

	// Generate a random number and adjust for favourable outcomes
	randNum := rand.Intn(totalRange)
	if randNum < favourableOutcomes {
		// Map the first `favourableOutcomes` to the favourable range
		randNum = randNum%favourableRange + favourableSet[0]
	} else {
		// Adjust the random number to exclude the favourable range and map to the rest of the range
		randNum = randNum%favourableOutcomes + lowerBound
		if randNum >= favourableSet[0] && randNum <= favourableSet[1] {
			randNum = favourableSet[1] + 1 + (randNum - favourableSet[0])
		}
	}

	return randNum
}
