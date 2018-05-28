package client

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
)

// UpstreamMetric captures nginx metrics referring to upstreams
type UpstreamMetric struct {
	Namespace string
	Service   string
	Port      string
	Metric    string
	Value     int64
}

func parseLine(line string) *UpstreamMetric {
	re := regexp.MustCompile("^nginx_upstream_([^{]+){[^}]+upstream=\"([^-]+)-(.+)-([0-9]+)\"} (.+)$")
	match := re.FindStringSubmatch(line)
	if len(match) > 0 {
		value, err := strconv.ParseFloat(match[5], 64)
		if err != nil {
			panic(err)
		}
		return &UpstreamMetric{
			Metric:    match[1],
			Namespace: match[2],
			Service:   match[3],
			Port:      match[4],
			Value:     int64(value),
		}
	}
	return nil
}

func Parse(input io.Reader) ([]UpstreamMetric, error) {
	metrics := make([]UpstreamMetric, 0)
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		text := scanner.Text()
		metric := parseLine(text)
		if metric != nil {
			metrics = append(metrics, *metric)
		}
	}
	return metrics, scanner.Err()
}
