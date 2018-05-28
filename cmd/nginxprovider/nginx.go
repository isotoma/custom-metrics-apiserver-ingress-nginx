package nginxprovider

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/golang/glog"
	"github.com/isotoma/custom-metrics-apiserver-ingress-nginx/pkg/client"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/metrics/pkg/apis/custom_metrics"
)

// The metrics we provide
var allMetrics = []provider.CustomMetricInfo{
	{
		GroupResource: schema.GroupResource{Group: "", Resource: "service"},
		Metric:        "ingress_requests_per_second",
		Namespaced:    true,
	},
}

type nginxProvider struct {
	client dynamic.ClientPool
	nginx  *client.NginxMetricsClient
	mapper apimeta.RESTMapper
	values *map[provider.CustomMetricInfo]int64
}

// New creates a new client
func New(pool dynamic.ClientPool, mapper apimeta.RESTMapper, label, port, path string, interval time.Duration, samples int64) provider.CustomMetricsProvider {
	//nginx := client.NewMetricsClient("ingress-nginx", "10254", "/metrics", "20s", 6)
	nginx := client.NewMetricsClient(label, port, path, interval, samples)
	nginx.Do()
	return &nginxProvider{
		client: pool,
		mapper: mapper,
		nginx:  nginx,
	}
}

func (p *nginxProvider) GetNamespacedMetricByName(groupResource schema.GroupResource, namespace, name, metricName string) (*custom_metrics.MetricValue, error) {
	value, err := p.nginx.GetValue(groupResource.Resource, namespace, name, metricName)
	if err != nil {
		return nil, err
	}
	return p.metricFor(value, groupResource, namespace, name, metricName)
}

func (p *nginxProvider) GetNamespacedMetricBySelector(groupResource schema.GroupResource, namespace string, selector labels.Selector, metricName string) (*custom_metrics.MetricValueList, error) {
	glog.Warningf("Call to GetNamespacedMetricBySelector unsupported")
	return nil, nil
}

func (p *nginxProvider) GetRootScopedMetricByName(groupResource schema.GroupResource, name, metricName string) (*custom_metrics.MetricValue, error) {
	glog.Warningf("Call to GetRootScopedMetricByName is unsupported")
	return nil, nil
}

func (p *nginxProvider) GetRootScopedMetricBySelector(groupResource schema.GroupResource, selector labels.Selector, metricName string) (*custom_metrics.MetricValueList, error) {
	glog.Warningf("Call to GetRootScopedMetricBySelector is unsupported")
	return nil, nil
}

func (p *nginxProvider) ListAllMetrics() []provider.CustomMetricInfo {
	return allMetrics
}

func (p *nginxProvider) metricFor(value int64, groupResource schema.GroupResource, namespace string, name string, metricName string) (*custom_metrics.MetricValue, error) {
	kind, err := p.mapper.KindFor(groupResource.WithVersion(""))
	if err != nil {
		return nil, err
	}

	return &custom_metrics.MetricValue{
		DescribedObject: custom_metrics.ObjectReference{
			APIVersion: groupResource.Group + "/" + runtime.APIVersionInternal,
			Kind:       kind.Kind,
			Name:       name,
			Namespace:  namespace,
		},
		MetricName: metricName,
		Timestamp:  metav1.Time{p.nginx.Timestamp()},
		Value:      *resource.NewQuantity(value, resource.DecimalSI),
	}, nil
}
