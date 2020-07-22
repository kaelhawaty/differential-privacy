//
// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package noise

import (
	"math"
	"testing"

	"github.com/grd/stat"
)

func approxEqual(a, b float64) bool {
	maxMagnitude := math.Max(math.Abs(a), math.Abs(b))
	if math.IsInf(maxMagnitude, +1) {
		return a == b
	}
	return math.Abs(a-b) <= 1e-6*maxMagnitude
}
func TestLaplaceStatistics(t *testing.T) {
	const numberOfSamples = 125000
	for _, tc := range []struct {
		l0Sensitivity                            int64
		lInfSensitivity, epsilon, mean, variance float64
	}{
		{
			l0Sensitivity:   1,
			lInfSensitivity: 1.0,
			epsilon:         1.0,
			mean:            0.0,
			variance:        2.0,
		},
		{
			l0Sensitivity:   1,
			lInfSensitivity: 1.0,
			epsilon:         ln3,
			mean:            0.0,
			variance:        2.0 / (ln3 * ln3),
		},
		{
			l0Sensitivity:   1,
			lInfSensitivity: 1.0,
			epsilon:         ln3,
			mean:            45941223.02107,
			variance:        2.0 / (ln3 * ln3),
		},
		{
			l0Sensitivity:   1,
			lInfSensitivity: 1.0,
			epsilon:         2.0 * ln3,
			mean:            0.0,
			variance:        2.0 / (2.0 * ln3 * 2.0 * ln3),
		},
		{
			l0Sensitivity:   1,
			lInfSensitivity: 2.0,
			epsilon:         2.0 * ln3,
			mean:            0.0,
			variance:        2.0 / (ln3 * ln3),
		},
		{
			l0Sensitivity:   2,
			lInfSensitivity: 1.0,
			epsilon:         2.0 * ln3,
			mean:            0.0,
			variance:        2.0 / (ln3 * ln3),
		},
	} {
		noisedSamples := make(stat.Float64Slice, numberOfSamples)
		for i := 0; i < numberOfSamples; i++ {
			noisedSamples[i] = lap.AddNoiseFloat64(tc.mean, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, 0)
		}
		sampleMean, sampleVariance := stat.Mean(noisedSamples), stat.Variance(noisedSamples)
		// Assuming that the Laplace samples have a mean of 0 and the specified variance of tc.variance,
		// sampleMeanFloat64 and sampleMeanInt64 are approximately Gaussian distributed with a mean of 0
		// and standard deviation of sqrt(tc.variance⁻ / numberOfSamples).
		//
		// The meanErrorTolerance is set to the 99.9995% quantile of the anticipated distribution. Thus,
		// the test falsely rejects with a probability of 10⁻⁵.
		meanErrorTolerance := 4.41717 * math.Sqrt(tc.variance/float64(numberOfSamples))
		// Assuming that the Laplace samples have the specified variance of tc.variance, sampleVarianceFloat64
		// and sampleVarianceInt64 are approximately Gaussian distributed with a mean of tc.variance and a
		// standard deviation of sqrt(5) * tc.variance / sqrt(numberOfSamples).
		//
		// The varianceErrorTolerance is set to the 99.9995% quantile of the anticipated distribution. Thus,
		// the test falsely rejects with a probability of 10⁻⁵.
		varianceErrorTolerance := 4.41717 * math.Sqrt(5.0) * tc.variance / math.Sqrt(float64(numberOfSamples))

		if !nearEqual(sampleMean, tc.mean, meanErrorTolerance) {
			t.Errorf("float64 got mean = %f, want %f (parameters %+v)", sampleMean, tc.mean, tc)
		}
		if !nearEqual(sampleVariance, tc.variance, varianceErrorTolerance) {
			t.Errorf("float64 got variance = %f, want %f (parameters %+v)", sampleVariance, tc.variance, tc)
		}
	}
}

func TestThresholdLaplace(t *testing.T) {
	// For the l0Sensitivity=1 cases, we make certain that we have implemented
	// both tails of the Laplace distribution. To do so, we write tests in pairs by
	// reflecting the value of delta around the axis 0.5 and the threshold "want"
	// value around the axis lInfSensitivity.
	//
	// This symmetry in the CDF is exploited by implicitly reflecting
	// partitionDelta (the per-partition delta) around the 0.5 axis. When
	// l0Sensitivity is 1, this can be easily expressed in tests since delta ==
	// partitionDelta. When l0Sensitivity != 1, it is no longer easy to express.
	for _, tc := range []struct {
		l0Sensitivity                         int64
		lInfSensitivity, epsilon, delta, want float64
	}{
		// Base test case
		{1, 1, ln3, 1e-10, 21.33},
		{1, 1, ln3, 1 - 1e-10, -19.33},
		// Scale lambda
		{1, 1, 2 * ln3, 1e-10, 11.16},
		{1, 1, 2 * ln3, 1 - 1e-10, -9.16},
		// Scale lInfSensitivity and lambda
		{1, 10, ln3, 1e-10, 213.28},
		{1, 10, ln3, 1 - 1e-10, -193.28},
		// Scale l0Sensitivity
		{10, 10, 10 * ln3, 1e-9, 213.28},
		{10, 10, 10 * ln3, 1 - 1e-9, -2.55}, // l0Sensitivity != 1, not expecting symmetry in "want" threhsold around lInfSensitivity.
		// High precision delta case
		{1, 1, ln3, 1e-200, 419.55},
	} {
		got := lap.Threshold(tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, 0, tc.delta)
		if !nearEqual(got, tc.want, 0.01) {
			t.Errorf("ThresholdForLaplace(%d,%f,%f,%e)=%f, want %f", tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.delta, got, tc.want)
		}
	}
}

func TestDeltaForThresholdLaplace(t *testing.T) {
	// For the l0Sensitivity=1 cases, we make certain that we have implemented
	// both tails of the Laplace distribution. To do so, we write tests in pairs by
	// reflecting the "want" value of delta around the axis 0.5 and the threshold
	// value k around the axis lInfSensitivity.
	//
	// This symmetry in the CDF is exploited by implicitly reflecting
	// partitionDelta (the per-partition delta) around the 0.5 axis. When
	// l0Sensitivity is 1, this can be easily expressed in tests since delta ==
	// partitionDelta. When l0Sensitivity != 1, it is no longer easy to express.
	for _, tc := range []struct {
		l0Sensitivity                     int64
		lInfSensitivity, epsilon, k, want float64
	}{
		// Base test case
		{1, 1, ln3, 20, 4.3e-10},
		{1, 1, ln3, -18, 1 - 4.3e-10},
		// Scale lInfSensitivity, lambda, and k
		{1, 10, ln3, 200, 4.3e-10},
		{1, 10, ln3, -180, 1 - 4.3e-10},
		// Scale lambda and k
		{1, 1, 2 * ln3, 10, 1.29e-9},
		{1, 1, 2 * ln3, -8, 1 - 1.29e-9},
		// Scale l0Sensitivity
		{10, 1, 10 * ln3, 20, 4.3e-9},
		{10, 1, 10 * ln3, -18, 1}, // l0Sensitivity != 1, not expecting symmetry in "want" delta around 0.5.
		// High precision delta case
		{1, 1, ln3, 419.55, 1e-200},
	} {
		got := lap.(laplace).DeltaForThreshold(tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, 0, tc.k)
		if !nearEqual(got, tc.want, 1e-2*tc.want) {
			t.Errorf("ThresholdForLaplace(%d,%f,%f,%f)=%e, want %e", tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.k, got, tc.want)
		}
	}
}
func TestInverseCDFLaplace(t *testing.T) {
	for _, tc := range []struct {
		desc                     string
		mean, lambda, prob, want float64
	}{
		// 2 high-precision calculated random tests.
		{
			desc:   "High-precision random test",
			mean:   64,
			lambda: 4,
			prob:   0.7875404240919761168041,
			want:   67.4234254367},
		{
			desc:   "High-precision random test",
			mean:   23,
			lambda: 2,
			prob:   0.1479685611330654517049,
			want:   20.564783456,
		},
		// Edge cases where p = 0 or p = 1.(Output should be infinite)
		{
			desc:   "Negative infinity output when probablity is 0",
			mean:   0,
			lambda: 1,
			prob:   0,
			want:   math.Inf(-1),
		},
		{
			desc:   "Positive infinity output when probablity is 1",
			mean:   1,
			lambda: 1,
			prob:   1,
			want:   math.Inf(1),
		},
		// Logical testing (with p = 0.5, return value should be mean regardless of lambda)
		{
			desc:   "%50 confidence level, different lambda",
			mean:   5,
			lambda: 5,
			prob:   0.5,
			want:   5,
		},
		{
			desc:   "%50 confidence level, different lambda",
			mean:   5,
			lambda: 10,
			prob:   0.5,
			want:   5,
		},
		// Edge cases where probablity is low or high. (Tests for accuracy)
		{
			desc:   "Low probablity",
			mean:   0,
			lambda: 3,
			prob:   2.88887425971E-8,
			want:   -50,
		},
		{
			desc:   "High probablitiy",
			mean:   0,
			lambda: 3,
			prob:   0.999999971111257,
			want:   50,
		},
	} {
		got := inverseCDFLaplace(tc.mean, tc.lambda, tc.prob)
		if !approxEqual(got, tc.want) {
			t.Errorf("TestInverseCDFLaplace(%f,%f,%f)=%0.12f, want %0.12f, desc: %s", tc.mean, tc.lambda,
				tc.prob, got, tc.want, tc.desc)
		}
	}

}

func TestConfidenceIntervalLaplace(t *testing.T) {
	// Tests for getConfidenceIntervalLaplace function
	for _, tc := range []struct {
		desc                    string
		noisedValue             float64
		lambda, confidenceLevel float64
		want                    ConfidenceIntervalFloat64
		wantErr                 bool
	}{
		// 4 Random pre-calculated tests.
		{
			desc:            "Random test",
			noisedValue:     13,
			lambda:          27.33333333333,
			confidenceLevel: 0.95,
			want:            ConfidenceIntervalFloat64{-68.88334881, 94.88334881},
		},
		{
			desc:            "Random test",
			noisedValue:     83.1235,
			lambda:          60,
			confidenceLevel: 0.76,
			want:            ConfidenceIntervalFloat64{-2.503481338, 168.7504813},
		},
		{
			desc:            "Random test",
			noisedValue:     5,
			lambda:          6.6666666666667,
			confidenceLevel: 0.4,
			want:            ConfidenceIntervalFloat64{1.594495842, 8.405504158},
		},
		{
			desc:            "Random test",
			noisedValue:     65.4621,
			lambda:          700,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalFloat64{-90.73838592, 221.6625859},
		},
		// Confidence interval with confidence level of 0 and 1 and Lambda is 10.
		{
			desc:            "Exact point interval, 0 confidence level",
			noisedValue:     0,
			lambda:          10,
			confidenceLevel: 0, // Zero confidence level means confidence interval is an exact point that equals the mean.
			want:            ConfidenceIntervalFloat64{0, 0},
		},
		{
			desc:            "Infinite interval, 1 confidence level",
			noisedValue:     0,
			lambda:          10,
			confidenceLevel: 1, // Probablity of one makes the interval infinite in both directions.
			want:            ConfidenceIntervalFloat64{math.Inf(-1), math.Inf(1)},
		},
		// Near 0 and 1 confidence levels.
		{
			desc:            "Low confidence level",
			noisedValue:     50,
			lambda:          10,
			confidenceLevel: 0.01,
			want:            ConfidenceIntervalFloat64{49.89949664, 50.1005033595},
		},
		{
			desc:            "High confidence level",
			noisedValue:     50,
			lambda:          10,
			confidenceLevel: 0.99,
			want:            ConfidenceIntervalFloat64{3.94829814, 96.05170186},
		},
	} {
		got := getConfidenceIntervalLaplace(tc.noisedValue, tc.lambda, tc.confidenceLevel)
		if !approxEqual(got.LowerBound, tc.want.LowerBound) {
			t.Errorf("TestConfidenceIntervalLaplace(%f, %f, %f)=%0.10f, want %0.10f, desc %s, LowerBound is not equal",
				tc.noisedValue, tc.lambda, tc.confidenceLevel,
				got.LowerBound, tc.want.LowerBound, tc.desc)
		}
		if !approxEqual(got.UpperBound, tc.want.UpperBound) {
			t.Errorf("TestConfidenceIntervalLaplace(%f, %f, %f)=%0.10f, want %0.10f, desc %s, UpperBound is not equal",
				tc.noisedValue, tc.lambda, tc.confidenceLevel,
				got.UpperBound, tc.want.UpperBound, tc.desc)
		}
	}

}

func TestReturnConfidenceIntervalFloat64(t *testing.T) {
	for _, tc := range []struct {
		desc                                      string
		noisedValue                               float64
		l0Sensitivity                             int64
		lInfSensitivity, epsilon, confidenceLevel float64
		want                                      ConfidenceIntervalFloat64
	}{
		{
			desc:            "Random test",
			noisedValue:     83.1235,
			l0Sensitivity:   3,
			lInfSensitivity: 2,
			epsilon:         0.1,
			confidenceLevel: 0.76,
			want:            ConfidenceIntervalFloat64{-2.503481338, 168.7504813},
		},
	} {
		got, err := lap.ReturnConfidenceIntervalFloat64(tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity,
			tc.epsilon, 0, tc.confidenceLevel)
		if err != nil {
			t.Errorf("ReturnConfidenceIntervalFloat64: when %s for err got %v", tc.desc, err)
		}
		if !approxEqual(got.LowerBound, tc.want.LowerBound) {
			t.Errorf("TestReturnConfidenceIntervalFloat64(%f, %d, %f, %f, %f)=%0.10f, want %0.10f, desc %s, LowerBound is not equal",
				tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.confidenceLevel,
				got.UpperBound, tc.want.LowerBound, tc.desc)
		}
		if !approxEqual(got.UpperBound, tc.want.UpperBound) {
			t.Errorf("TestReturnConfidenceIntervalFloat64(%f, %d, %f, %f, %f)=%0.10f, want %0.10f, desc %s, UpperBound is not equal",
				tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.confidenceLevel,
				got.UpperBound, tc.want.LowerBound, tc.desc)
		}
	}
}
func TestReturnConfidenceIntervalInt64(t *testing.T) {
	for _, tc := range []struct {
		desc                                        string
		noisedValue, l0Sensitivity, lInfSensitivity int64
		epsilon, confidenceLevel                    float64
		want                                        ConfidenceIntervalInt64
	}{
		{
			desc:            "Random test",
			noisedValue:     65,
			l0Sensitivity:   7,
			lInfSensitivity: 10,
			epsilon:         0.1,
			confidenceLevel: 0.2,
			want:            ConfidenceIntervalInt64{-91, 221},
		},
		{
			desc:            "Random test",
			noisedValue:     5,
			l0Sensitivity:   1,
			lInfSensitivity: 2,
			epsilon:         0.3,
			confidenceLevel: 0.4,
			want:            ConfidenceIntervalInt64{2, 8},
		},
	} {
		got, err := lap.ReturnConfidenceIntervalInt64(tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity,
			tc.epsilon, 0, tc.confidenceLevel)
		if err != nil {
			t.Errorf("ReturnConfidenceIntervalInt64: when %s for err got %v", tc.desc, err)
		}
		if got.LowerBound != tc.want.LowerBound {
			t.Errorf("TestReturnConfidenceIntervalInt64(%d, %d, %d, %f, %f)=%d, want %d, desc %s, LowerBound is not equal",
				tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.confidenceLevel,
				got.UpperBound, tc.want.LowerBound, tc.desc)
		}
		if got.UpperBound != tc.want.UpperBound {
			t.Errorf("TestReturnConfidenceIntervalInt64(%d, %d, %d, %f, %f)=%d, want %d, desc %s, UpperBound is not equal",
				tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity, tc.epsilon, tc.confidenceLevel,
				got.UpperBound, tc.want.LowerBound, tc.desc)
		}
	}
}

func TestChecksConfidenceInterval(t *testing.T) {
	for _, tc := range []struct {
		desc                                      string
		noisedValue                               float64
		l0Sensitivity                             int64
		lInfSensitivity, epsilon, confidenceLevel float64
		want                                      ConfidenceIntervalFloat64
	}{
		{
			desc:            "Negative Distribution parameter for Laplace args checks",
			noisedValue:     0,
			l0Sensitivity:   7,
			lInfSensitivity: -1,
			epsilon:         0.1,
			confidenceLevel: 0.2,
		},
		{
			desc:            "Negative confidence level",
			noisedValue:     0,
			l0Sensitivity:   1,
			lInfSensitivity: 2,
			epsilon:         0.3,
			confidenceLevel: -1,
		},
		{
			desc:            "Greater than 1 confidence level",
			noisedValue:     0,
			l0Sensitivity:   1,
			lInfSensitivity: 2,
			epsilon:         0.3,
			confidenceLevel: 2,
		},
	} {
		_, err := lap.ReturnConfidenceIntervalInt64(int64(tc.noisedValue), tc.l0Sensitivity, int64(tc.lInfSensitivity),
			tc.epsilon, 0, tc.confidenceLevel)
		if err == nil {
			t.Errorf("ReturnConfidenceIntervalInt64: didn't return an error, desc %s", tc.desc)
		}
		_, err = lap.ReturnConfidenceIntervalFloat64(tc.noisedValue, tc.l0Sensitivity, tc.lInfSensitivity,
			tc.epsilon, 0, tc.confidenceLevel)
		if err == nil {
			t.Errorf("ReturnConfidenceIntervalFloat64: didn't return an error, desc %s", tc.desc)
		}
	}
}
func TestGeometricStatistics(t *testing.T) {
	const numberOfSamples = 125000
	for _, tc := range []struct {
		lambda float64
		mean   float64
		stdDev float64
	}{
		{
			lambda: 0.1,
			mean:   10.50833,
			stdDev: 9.99583,
		},
		{
			lambda: 0.0001,
			mean:   10000.50001,
			stdDev: 9999.99999,
		},
		{
			lambda: 0.0000001,
			mean:   10000000.5,
			stdDev: 9999999.99999,
		},
	} {
		geometricSamples := make(stat.IntSlice, numberOfSamples)
		for i := 0; i < numberOfSamples; i++ {
			geometricSamples[i] = geometric(tc.lambda)
		}
		sampleMean := stat.Mean(geometricSamples)
		// Assuming that the geometric samples have the specified mean tc.mean and the standard
		// deviation of tc.stdDev, sampleMean is approximately Gaussian distributed with a mean
		// of tc.stdDev and standard deviation of tc.stdDev / sqrt(numberOfSamples).
		//
		// The meanErrorTolerance is set to the 99.9995% quantile of the anticipated distribution
		// of sampleMean. Thus, the test falsely rejects with a probability of 10⁻⁵.
		meanErrorTolerance := 4.41717 * tc.stdDev / math.Sqrt(float64(numberOfSamples))

		if !nearEqual(sampleMean, tc.mean, meanErrorTolerance) {
			t.Errorf("got mean = %f, want %f (parameters %+v)", sampleMean, tc.mean, tc)
		}
	}
}
