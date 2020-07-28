package main

import (
	"fmt"

	"github.com/google/differential-privacy/go/noise"
)

func main() {
	trueSum, trueCount := 50.0, 10.0
	trueMean := trueSum / trueCount
	lap := noise.Laplace()
	// Senstivity parameters
	l0Sensitivity, lInfSensitivity, epsilon := int64(1), float64(1), float64(0.1)
	// Increments for confidence interval
	incrment := 0.01
	valuesCount, value := int(1/incrment), 0.0
	confLevels := make([]float64, valuesCount, valuesCount)
	for i := 0; i < valuesCount; i++ {
		confLevels[i], value = value, value+incrment
	}

	// Estimated Output for difference confidence levels
	estConfLevels := make([]float64, valuesCount, valuesCount)
	numSamples := 100000
	for i, confLevel := range confLevels {
		cntInsideInterval := 0
		for j := 0; j < numSamples; j++ {
			// Getting noised values
			noisedSum, noisedCount := lap.AddNoiseFloat64(trueSum, l0Sensitivity, lInfSensitivity, epsilon, 0), lap.AddNoiseFloat64(trueCount, l0Sensitivity, lInfSensitivity, epsilon, 0)
			// Getting confidence intervals for noised values
			confIntSum, _ := lap.ReturnConfidenceIntervalFloat64(noisedSum, l0Sensitivity, lInfSensitivity, epsilon, 0, confLevel)
			confIntCount, _ := lap.ReturnConfidenceIntervalFloat64(noisedCount, l0Sensitivity, lInfSensitivity, epsilon, 0, confLevel)

			// Computing confidence interval for mean
			confIntMean := noise.ConfidenceIntervalFloat64{LowerBound: confIntSum.LowerBound / confIntCount.UpperBound, UpperBound: confIntSum.UpperBound / confIntCount.LowerBound}
			if confIntMean.LowerBound <= trueMean && trueMean <= confIntMean.UpperBound {
				cntInsideInterval++
			}
		}
		estConfLevels[i] = float64(cntInsideInterval) / float64(numSamples)
	}
	for i := 0; i < valuesCount; i++ {
		fmt.Printf("%f, %f\n", confLevels[i], estConfLevels[i])
	}

}
