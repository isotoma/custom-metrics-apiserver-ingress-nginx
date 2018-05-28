package client

import (
	"reflect"
	"testing"
)

var (
	pod1 = []UpstreamMetric{
		UpstreamMetric{
			Namespace: "default",
			Service:   "foo",
			Port:      "80",
			Metric:    "total_requests",
			Value:     1000,
		},
		UpstreamMetric{
			Namespace: "default",
			Service:   "foo",
			Port:      "81",
			Metric:    "total_requests",
			Value:     1500,
		},
		UpstreamMetric{
			Namespace: "default",
			Service:   "bar",
			Port:      "80",
			Metric:    "total_requests",
			Value:     2000,
		},
	}
	pod2 = []UpstreamMetric{
		UpstreamMetric{
			Namespace: "default",
			Service:   "foo",
			Port:      "80",
			Metric:    "total_requests",
			Value:     3000,
		},
		UpstreamMetric{
			Namespace: "default",
			Service:   "bar",
			Port:      "80",
			Metric:    "total_requests",
			Value:     4000,
		},
	}
	podMetrics = map[string][]UpstreamMetric{
		"pod1": pod1,
		"pod2": pod2,
	}
	pod1foo = TotalKey{
		PodIP:     "pod1",
		Namespace: "default",
		Service:   "foo",
	}
	pod2foo = TotalKey{
		PodIP:     "pod2",
		Namespace: "default",
		Service:   "foo",
	}
	pod1bar = TotalKey{
		PodIP:     "pod1",
		Namespace: "default",
		Service:   "bar",
	}
	pod2bar = TotalKey{
		PodIP:     "pod2",
		Namespace: "default",
		Service:   "bar",
	}
	oldTotals = map[TotalKey]int64{
		pod1foo: 2500,
		pod2foo: 3000,
		pod1bar: 1200,
		pod2bar: 2800,
	}
	newTotals = map[TotalKey]int64{
		pod1foo: 2530,
		pod2foo: 3500,
		pod1bar: 1500,
		pod2bar: 2800,
	}
	foo = RateKey{
		Namespace: "default",
		Service:   "foo",
	}
	bar = RateKey{
		Namespace: "default",
		Service:   "bar",
	}
	newRates = map[RateKey][]int64{
		foo: []int64{3000, 1800, 1200},
		bar: []int64{1200, 0, 300},
	}
	oldRates = map[RateKey]int64{
		foo: 20,
		bar: 2,
	}
)

func TestTotalRequests(t *testing.T) {
	totals := TotalRequests(podMetrics)
	if totals[TotalKey{
		PodIP:     "pod1",
		Namespace: "default",
		Service:   "foo",
	}] != 2500 {
		t.Error("Stores bad")
	}
}

func TestAggregateTotals(t *testing.T) {
	agg := AggregateTotals(newTotals, oldTotals)
	if !reflect.DeepEqual(agg[RateKey{
		Namespace: "default",
		Service:   "foo",
	}], []int64{30, 500}) {
		t.Error("Adds bad")
	}
}

func TestCalculateRates(t *testing.T) {
	rates := CalculateRates(newRates, oldRates, 60, 3)
	if rates[foo] != 25 {
		t.Errorf("Averages bad: %d", rates[foo])
	}
}
