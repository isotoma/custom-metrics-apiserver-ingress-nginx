package parser

import (
	"bufio"
	"io"
	"regexp"
)

// UpstreamMetric captures nginx metrics referring to upstreams
type UpstreamMetric struct {
	Namespace string
	Service   string
	Port      string
	Metric    string
	Value     string
}

func parseLine(line string) *UpstreamMetric {
	re := regexp.MustCompile("^nginx_upstream_([^{]+){[^}]+upstream=\"([^-]+)-(.+)-([0-9]+)\"} (.+)$")
	match := re.FindStringSubmatch(line)
	if len(match) > 0 {
		return &UpstreamMetric{
			Metric:    match[1],
			Namespace: match[2],
			Service:   match[3],
			Port:      match[4],
			Value:     match[5],
		}
	}
	return nil
}

func parse(input io.Reader) ([]UpstreamMetric, error) {
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

/*
func main() {
	f, err := os.Open("metrics")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	metrics, err := parse(f)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", metrics)
}
*/
