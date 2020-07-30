package main

import (
	"fmt"
	"math"

	"github.com/google/differential-privacy/go/noise"
)

func main() {
	trueSum, trueCount := 10000000.0, 50
	trueMean := trueSum / float64(trueCount)
	lap := noise.Laplace()
	// Senstivity parameters
	l0Sensitivity, lInfSensitivity, epsilon := int64(1), float64(1), float64(1)
	// Increments for confidence interval
	incrment := 0.01
	valuesCount, value := int(1/incrment), 0.0
	confLevels := make([]float64, valuesCount, valuesCount)
	for i := 0; i < valuesCount; i++ {
		confLevels[i], value = value, value+incrment
	}
	prev := int64(0)
	// Estimated Output for difference confidence levels
	estConfLevels := make([]float64, valuesCount, valuesCount)
	numSamples := 10000
	for i, confLevel := range confLevels {
		cntInsideInterval := 0
		for j := 0; j < numSamples; j++ {
			// Getting noised values
			noisedSum, noisedCount := lap.AddNoiseFloat64(trueSum, l0Sensitivity, lInfSensitivity, epsilon, 0), lap.AddNoiseInt64(int64(trueCount), l0Sensitivity, int64(lInfSensitivity), epsilon, 0)
			// Getting confidence intervals for noised values
			confIntSum, _ := lap.ReturnConfidenceIntervalFloat64(noisedSum, l0Sensitivity, lInfSensitivity, epsilon, 0, math.Sqrt(confLevel))
			confIntCount, _ := lap.ReturnConfidenceIntervalInt64(int64(noisedCount), l0Sensitivity, int64(lInfSensitivity), epsilon, 0, math.Sqrt(confLevel))
			confIntCount.LowerBound, confIntCount.UpperBound = int64(math.Max(1, float64(confIntCount.LowerBound))), int64(math.Max(1, float64(confIntCount.UpperBound)))
			// Computing confidence interval for mean
			var estLowerBound, estUpperBound float64
			if confIntSum.LowerBound >= 0 {
				estLowerBound = confIntSum.LowerBound / float64(confIntCount.UpperBound)
			} else {
				estLowerBound = confIntSum.LowerBound / float64(confIntCount.LowerBound)
			}
			if confIntSum.UpperBound >= 0 {
				estUpperBound = confIntSum.UpperBound / float64(confIntCount.LowerBound)
			} else {
				estUpperBound = confIntSum.UpperBound / float64(confIntCount.UpperBound)
			}
			confIntMean := noise.ConfidenceIntervalFloat64{LowerBound: estLowerBound, UpperBound: estUpperBound}
			if confIntMean.LowerBound <= trueMean && trueMean <= confIntMean.UpperBound {
				cntInsideInterval++
			}
			if prev < (confIntCount.UpperBound - confIntCount.LowerBound) {
				prev = confIntCount.UpperBound - confIntCount.LowerBound
				fmt.Println(i)
			}
		}
		estConfLevels[i] = float64(cntInsideInterval) / float64(numSamples)
	}
	for i := 0; i < valuesCount; i++ {
		fmt.Printf("%f, %f\n", confLevels[i], estConfLevels[i])
	}

}
