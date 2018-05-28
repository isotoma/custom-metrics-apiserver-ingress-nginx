package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// TotalKey is the indexing of totals from the nginx traffic stats, which vary per pod
type TotalKey struct {
	PodIP     string
	Namespace string
	Service   string
}

// RateKey is the key for aggregated rates, which are averaged across all pods
type RateKey struct {
	Namespace string
	Service   string
}

// NginxMetricsClient holds some state about the clients
type NginxMetricsClient struct {
	Label         string
	Port          string
	Path          string
	Duration      string
	MovingSamples int64
	last          map[TotalKey]int64
	lastTime      time.Time
	current       map[RateKey]int64
}

// NewMetricsClient returns a new client and starts collecting metrics
func NewMetricsClient(label, port, path, duration string, movingSamples int64) *NginxMetricsClient {
	return &NginxMetricsClient{
		Label:         label,
		Port:          port,
		Path:          path,
		Duration:      duration,
		MovingSamples: movingSamples,
	}
}

// Do does things
func (c *NginxMetricsClient) Do() {
	go c.Loop()
	go c.LoopForever()
}

// LoopForever starts the regular requests
func (c *NginxMetricsClient) LoopForever() {
	duration, err := time.ParseDuration(c.Duration)
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
			go c.Loop()
		}
	}
}

// GetValue returns average requests per second * 100 (i.e. int millis)
func (c *NginxMetricsClient) GetValue(resource, namespace, name, metricName string) (int64, error) {
	if resource != "services" || metricName != "ingress_requests_per_second" {
		return 0, fmt.Errorf("Request for metric %s for resource %s, that I don't know about", metricName, resource)
	}
	if c.current == nil {
		return 0, fmt.Errorf("No metrics collected (yet)")
	}
	key := RateKey{
		Namespace: namespace,
		Service:   name,
	}
	if value, ok := c.current[key]; ok {
		glog.Infof("GetValue resource=%s namespace=%s name=%s metricName=%s value=%d", resource, namespace, name, metricName, value)
		return value, nil
	}
	return 0, fmt.Errorf("Request for value %s.%s that doesn't exist", name, namespace)
}

// Loop does the fetch-and-update loop
func (c *NginxMetricsClient) Loop() {
	metrics, err := c.Fetch()
	if err != nil {
		fmt.Println(err.Error())
	}
	c.UpdateRates(metrics)
}

func TotalRequests(podMetrics map[string][]UpstreamMetric) map[TotalKey]int64 {
	totals := make(map[TotalKey]int64)
	for podIP, metrics := range podMetrics {
		for _, metric := range metrics {
			// this is the only one we handle atm
			if metric.Metric == "requests_total" {
				key := TotalKey{
					PodIP:     podIP,
					Namespace: metric.Namespace,
					Service:   metric.Service,
				}
				if _, ok := totals[key]; ok {
					totals[key] += metric.Value
				} else {
					totals[key] = metric.Value
				}
			}
		}
	}
	return totals
}

func AggregateTotals(totals map[TotalKey]int64, last map[TotalKey]int64) map[RateKey][]int64 {
	diffs := make(map[RateKey][]int64)
	// Loop through the current totals
	for k, v := range totals {
		// Get the previous total. If there isn't one then we cannot calculate a rate
		if prev, ok := last[k]; ok {
			rateKey := RateKey{
				Namespace: k.Namespace,
				Service:   k.Service,
			}
			// Annoying that there is no defaultdict
			l, ok := diffs[rateKey]
			if !ok {
				l = make([]int64, 0)
			}
			// Add this difference to the list of other ones
			diffs[rateKey] = append(l, v-prev)
		}
	}
	return diffs
}

func average(n []int64) int64 {
	var total int64
	total = 0
	for _, i := range n {
		total += i
	}
	return total / int64(len(n))
}

// CalculateRates calculates the moving average rates per service
func CalculateRates(newRates map[RateKey][]int64, oldRates map[RateKey]int64, period, samples int64) map[RateKey]int64 {
	newCurrent := make(map[RateKey]int64)
	for k, v := range newRates {
		avg := average(v)
		newRate := avg / period
		// Basic average if we don't have the old one
		newCurrent[k] = newRate
		// We can calculate a moving average, if we have the old one
		if oldRate, ok := oldRates[k]; ok {
			newCurrent[k] = oldRate - (oldRate / samples) + (newRate / samples)
		}
	}
	return newCurrent
}

// UpdateRates gets the latest totals and calculates the new rates
func (c *NginxMetricsClient) UpdateRates(podMetrics map[string][]UpstreamMetric) {
	totals := TotalRequests(podMetrics)
	now := time.Now()
	period := int64(now.Sub(c.lastTime).Seconds())
	last := c.last
	c.last = totals
	c.lastTime = now
	if last == nil {
		glog.Warning("No previous metrics available")
		c.current = make(map[RateKey]int64)
		return
	}
	diffs := AggregateTotals(totals, last)
	c.current = CalculateRates(diffs, c.current, period, c.MovingSamples)
}

// Fetch gets the values from the current ingress pods
func (c *NginxMetricsClient) Fetch() (map[string][]UpstreamMetric, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: "app=ingress-nginx",
	})
	if err != nil {
		return nil, err
	}
	metrics := make(map[string][]UpstreamMetric)
	for _, p := range pods.Items {
		url := fmt.Sprintf("http://%s:%s%s", p.Status.PodIP, c.Port, c.Path)
		c := http.Client{}
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		response, err := c.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		upstream, err := Parse(bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}
		metrics[p.Status.PodIP] = upstream
	}
	return metrics, nil
}
