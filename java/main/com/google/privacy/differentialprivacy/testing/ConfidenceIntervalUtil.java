package com.google.privacy.differentialprivacy.testing;

import com.google.privacy.differentialprivacy.ConfidenceInterval;

import static com.google.common.truth.Truth.assertThat;

public class ConfidenceIntervalUtil {
  /** Number of samples used to run confidence interval simulation */
  private static final int numSamples = 100000;

  /**
   * Interface to encapsulate the behavior of generating confidence intervals for noise generators
   * and aggregations.
   */
  public interface ConfidenceIntervalGenerator {
    /** Return alpha level of the confidence intervals that are being generated */
    double getAlpha();
    /** Return the true value of the statistic */
    double getTrueValue();
    /**
     * This function should encapsulate the generation of noised values and return a confidence
     * interval around the generated noised value with confidence level equal to 1 - {@link
     * ConfidenceIntervalGenerator#getAlpha()}
     */
    ConfidenceInterval computeConfidenceInterval();
  }

  public static double runSimulation(ConfidenceIntervalGenerator confIntGen) {
    int cntOutsideInterval = 0;
    for (int i = 0; i < numSamples; i++) {
      ConfidenceInterval confInt = confIntGen.computeConfidenceInterval();
      double trueValue = confIntGen.getTrueValue();
      if (confInt.lowerBound() > trueValue || confInt.upperBound() < trueValue) {
        cntOutsideInterval++;
      }
    }
    return (double) cntOutsideInterval / numSamples;
  }

  public static void runTwoSidedTest(ConfidenceIntervalGenerator confIntGen) {
    double alphaEstimator = runSimulation(confIntGen);
    double requestedAlpha = confIntGen.getAlpha();
    double variance = requestedAlpha * (1 - requestedAlpha) / numSamples;

    // The tolerance is chosen according to the 99.999995% quantile of the anticipated distributions
    // of the sample mean. Thus, the test falsely rejects with a probability of 10^-5.
    double tolerance = 4.4171734 * Math.sqrt(variance);

    assertThat(alphaEstimator).isWithin(tolerance).of(requestedAlpha);
  }

  public static void runOneSidedTest(ConfidenceIntervalGenerator confIntGen) {
    double alphaEstimator = runSimulation(confIntGen);
    double requestedAlpha = confIntGen.getAlpha();
    double variance = requestedAlpha * (1 - requestedAlpha) / numSamples;

    // The tolerance is chosen according to the 99.99999% quantile of the anticipated distributions
    // of the sample mean. Thus, the test falsely rejects with a probability of 10^-5.
    double tolerance = 4.2648908 * Math.sqrt(variance);

    assertThat(alphaEstimator).isAtMost(tolerance + requestedAlpha);
  }
}
