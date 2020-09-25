package com.google.privacy.differentialprivacy;

import java.io.*;

public class main {
  private static double MAX_EPSILON = 2;
  private static Noise NOISE = new GaussianNoise();
  private static Double DEFAULT_DELTA = 0.123;


  private static double[] getEpsilons(double increment) {
    int size = (int) Math.floor(MAX_EPSILON / increment);
    double[] arr = new double[size];
    double curEpsilon = increment;
    for (int i = 0; i < size; i++, curEpsilon += increment) {
      arr[i] = curEpsilon;
    }
    return arr;
  }

  private static BoundedMean getBoundedMean(double trueValue, int cnt, double epsilon) {
      BoundedMean mean =
          BoundedMean.builder()
              .epsilon(epsilon)
              .delta(DEFAULT_DELTA)
              .maxContributionsPerPartition(1)
              .maxPartitionsContributed(1)
              .noise(NOISE)
              .lower(0)
              .upper(2*trueValue)
              .build();
    for (int i = 0; i < cnt; i++) {
      mean.addEntry(trueValue);
    }
    return mean;
  }

  private static BoundedMeanCountFloat getBoundedMeanCountFloat(double trueValue, int cnt, double epsilon) {
    BoundedMeanCountFloat mean =
            BoundedMeanCountFloat.builder()
                    .epsilon(epsilon)
                    .delta(DEFAULT_DELTA)
                    .maxContributionsPerPartition(1)
                    .maxPartitionsContributed(1)
                    .noise(NOISE)
                    .lower(0)
                    .upper(2*trueValue)
                    .build();
    for (int i = 0; i < cnt; i++) {
      mean.addEntry(trueValue);
    }
    return mean;
  }

  public static void main(String[] args) throws IOException {
    File file = new File("/home/" + System.getProperty("user.name") + "/output.txt");
    file.createNewFile();
    PrintStream out = new PrintStream(
            new FileOutputStream(file, true), true);
    out.println();
    int numSamples = 10000000;
    double trueMean = 100.0; // Arbitrary value but should be reasonably high to see the small changes in count.
    int cnt = 20; // Add trueMean to BoundedMean cnt times

    double[] epsilons = getEpsilons(0.1);
    double[] variance = new double[epsilons.length];

    for (int i = 0; i < epsilons.length; i++) {
        System.out.println("Progress: BoundedMean " + i);
        BoundedMean mean = getBoundedMean(trueMean, cnt, epsilons[i]);
        double sumOfSquares = 0.0;
        for (int j = 0; j < numSamples; j++) {
          mean.reset();
          double noisedValue = mean.computeResult();
          sumOfSquares += (noisedValue - trueMean) * (noisedValue - trueMean);
        }
        variance[i] = sumOfSquares / numSamples;
    }

    for (int i = 0; i < epsilons.length; i++) {
      out.printf("%f\n", variance[i]);
    }

    out.println("=======================================");

    for (int i = 0; i < epsilons.length; i++) {
      System.out.println("Progress: BoundedMeanCountFloat " + i);
      BoundedMeanCountFloat mean = getBoundedMeanCountFloat(trueMean, cnt, epsilons[i]);
      double sumOfSquares = 0.0;
      for (int j = 0; j < numSamples; j++) {
        mean.reset();
        double noisedValue = mean.computeResult();
        sumOfSquares += (noisedValue - trueMean) * (noisedValue - trueMean);
      }
      variance[i] = sumOfSquares / numSamples;
    }

    for (int i = 0; i < epsilons.length; i++) {
      out.printf("%f\n", variance[i]);
    }
  }
}
