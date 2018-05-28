package client

import "testing"

func TestParseLineComment(t *testing.T) {
	metrics := parseLine("# HELP cpu_seconds_total Cpu usage in seconds")
	if metrics != nil {
		t.Error("Parses comment incorrectly")
	}
}

func TestParseLineMiss(t *testing.T) {
	metrics := parseLine("go_gc_duration_seconds{quantile=\"0.5\"} 5.6742e-05")
	if metrics != nil {
		t.Error("Parses miss incorrectly")
	}
}

func TestParseLineHit(t *testing.T) {
	metrics := parseLine("nginx_upstream_backup{ingress_class=\"nginx\",namespace=\"\",server=\"100.96.10.10:80\",upstream=\"default-foo-bar-80\"} 4.591337e+06")
	good := UpstreamMetric{
		Metric:    "backup",
		Namespace: "default",
		Service:   "foo-bar",
		Port:      "80",
		Value:     4591337,
	}
	if *metrics != good {
		t.Errorf("Parses hit incorrectly %+v", *metrics)
	}
}
