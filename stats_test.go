/*
 * Copyright (c) 2015-2020 by MemSQL. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dbbench

import (
	"fmt"
	"reflect"
	"testing"
)

func assertNear(t *testing.T, expected float64, actual float64, msg string) {
	if actual+.001 < expected || actual-.001 > expected {
		t.Error(msg, "expected", expected, "but got", actual)
	}
}

func TestNormInverseCDF(t *testing.T) {
	type testcase struct {
		p float64
		z float64
	}

	for _, testCase := range []testcase{
		{0.95, 1.645},
		{0.99, 2.326},
	} {
		assertNear(t, testCase.z, NormInverseCDF(testCase.p),
			fmt.Sprint("For", testCase.p))
	}
}

func TestStreamingHistogram(t *testing.T) {
	type testcase struct {
		vals     []uint64
		expected map[uint]uint64
	}

	for _, testCase := range []testcase{
		{[]uint64{1}, map[uint]uint64{1: 1}},
		{[]uint64{0}, map[uint]uint64{0: 1}},
		{[]uint64{1, 3, 3, 3, 4, 16, 257}, map[uint]uint64{1: 1, 2: 3, 3: 1, 5: 1, 9: 1}},
	} {
		var sh StreamingHistogram
		for _, v := range testCase.vals {
			sh.Add(v)
		}

		t.Logf("Testing %d", testCase.vals)
		var expectedBuckets [64]uint64
		for k, v := range testCase.expected {
			expectedBuckets[k] = v
		}

		if !reflect.DeepEqual(expectedBuckets, sh.Buckets) {
			t.Errorf("For buckets\n\texpected %d\n\tbut got %d",
				expectedBuckets, sh.Buckets)
		}
	}
}

func TestStreamingSample(t *testing.T) {
	type testcase struct {
		vals        []float64
		bucketCount int

		min     float64
		max     float64
		buckets []int
	}

	for _, testCase := range []testcase{
		{[]float64{1}, 1, 1, 1, []int{1}},
		{[]float64{1}, 3, 1, 1, []int{1}},
		{[]float64{1, 2}, 1, 1, 2, []int{2}},
		{[]float64{1, 2, 2, 3}, 3, 1, 3, []int{1, 2, 1}},
	} {
		var ss StreamingSample
		for _, v := range testCase.vals {
			ss.Add(v)
		}
		t.Logf("Testing %f", testCase.vals)

		if ss.Count() != len(testCase.vals) {
			t.Error("For count expected", len(testCase.vals),
				"but got got", ss.Count())
		}

		buckets, min, max, _ := ss.Histogram(testCase.bucketCount)
		if !reflect.DeepEqual(testCase.buckets, buckets) {
			t.Errorf("For buckets\n\texpected %d\n\tbut got %d",
				testCase.buckets, buckets)
		}

		if min != testCase.min {
			t.Errorf("For min\n\texpected %f\n\tbut got %f",
				testCase.min, min)
		}

		if max != testCase.max {
			t.Errorf("For max\n\texpected %f\n\tbut got %f",
				testCase.max, max)
		}
	}
}

func TestStreamingStats(t *testing.T) {
	type testcase struct {
		vals   []float64
		mean   float64
		stddev float64
	}

	for _, testCase := range []testcase{
		{[]float64{1, 2}, 1.5, .707},
		{[]float64{1, 2, 3, 4, 5}, 3, 1.581},
		{[]float64{1, 1, 1}, 1, 0},
	} {
		var ss StreamingStats
		for _, v := range testCase.vals {
			ss.Add(v)
		}

		if ss.Count() != len(testCase.vals) {
			t.Error("For count of", testCase.vals,
				"expected", len(testCase.vals),
				"got", ss.Count())
		}
		assertNear(t, testCase.mean, ss.Mean(),
			fmt.Sprint("For mean of", testCase.vals))
		assertNear(t, testCase.stddev, ss.SampleStdDev(),
			fmt.Sprint("For stddev of", testCase.vals))
	}
}
