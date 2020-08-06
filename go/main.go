package main

import (
	"fmt"
	"math"

	"github.com/google/differential-privacy/go/noise"
)

// Senstivity parameters
var l0Sensitivity, lInfSensitivity, epsilon = int64(1), float64(1), float64(1)
var lap = noise.Laplace()

func getConfIntMean(noisedSum float64, noisedCount int64, confLevelSum, confLevelCount float64) noise.ConfidenceIntervalFloat64 {
	confIntSum, _ := lap.ReturnConfidenceIntervalFloat64(noisedSum, l0Sensitivity, lInfSensitivity, epsilon, 0, confLevelSum)
	confIntCount, _ := lap.ReturnConfidenceIntervalInt64(noisedCount, l0Sensitivity, int64(lInfSensitivity), epsilon, 0, confLevelCount)
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
	return confIntMean
}
func main() {
	trueSum, trueCount := 100.0, 1
	trueMean := trueSum / float64(trueCount)

	// Increments for confidence interval
	incrment := 0.01
	valuesCount, value := int(1/incrment), incrment
	confLevels := make([]float64, valuesCount, valuesCount)
	for i := 0; i < valuesCount; i++ {
		confLevels[i], value = value, value+incrment
	}
	// Estimated Output for difference confidence levels
	estConfLevels := make([]float64, valuesCount, valuesCount)
	numSamples := 1000
	//splits := float64(1000.0)
	for i, confLevel := range confLevels {
		cntInsideInterval := 0
		for j := 0; j < numSamples; j++ {
			// Getting noised values
			noisedSum, noisedCount := lap.AddNoiseFloat64(trueSum, l0Sensitivity, lInfSensitivity, epsilon, 0), lap.AddNoiseInt64(int64(trueCount), l0Sensitivity, int64(lInfSensitivity), epsilon, 0)
			// Getting confidence intervals for noised values
			bestConfIntMean := getConfIntMean(noisedSum, noisedCount, math.Sqrt(confLevel), math.Sqrt(confLevel))
			/*ratioInc := (1/confLevel - confLevel) / splits
			var bestConfIntMean noise.ConfidenceIntervalFloat64
			bestTightness := math.MaxFloat64
			for i := 0; i < 1000; i++ {
				curRatio := confLevel + ratioInc*float64(i)
				a := math.Sqrt(confLevel / curRatio)
				curConfInt := getConfIntMean(noisedSum, noisedCount, curRatio*a, a)
				tightness := curConfInt.UpperBound - curConfInt.LowerBound
				if tightness < bestTightness {
					bestConfIntMean = curConfInt
					bestTightness = tightness
				}
			}*/
			if bestConfIntMean.LowerBound <= trueMean && trueMean <= bestConfIntMean.UpperBound {
				cntInsideInterval++
			}
		}
		estConfLevels[i] = float64(cntInsideInterval) / float64(numSamples)
	}
	for i := 0; i < valuesCount; i++ {
		fmt.Printf("%f, %f\n", confLevels[i], estConfLevels[i])
	}

}
