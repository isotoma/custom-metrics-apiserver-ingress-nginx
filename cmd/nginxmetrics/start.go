package nginxmetrics

import (
	"fmt"
	"io"
	"time"

	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/dynamicmapper"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/cmd/server"
	"github.com/spf13/cobra"

	"github.com/isotoma/custom-metrics-apiserver-ingress-nginx/cmd/nginxprovider"
)

// Adapter is the adapter that takes nginx metrics and sends them to k8s
type Adapter struct {
	*server.CustomMetricsAdapterServerOptions
	// RemoteKubeConfigFile is the config used to list pods from the master API server
	RemoteKubeConfigFile string
	// DiscoveryInterval is the interval at which discovery information is refreshed
	DiscoveryInterval time.Duration
	// Label is the label for the ingress-nginx pods
	Label string
	// MetricsPort is the port on the pods that delivers the prometheus format metrics, 10254 by default
	MetricsPort string
	// MetricsPath is the path that delivers the metrics, default /metrics
	MetricsPath string
	// AverageSamples is the number of samples to consider for a moving average
	AverageSamples int64
}

// Start starts the server
func Start(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	baseOpts := server.NewCustomMetricsAdapterServerOptions(out, errOut)
	adapter := Adapter{
		CustomMetricsAdapterServerOptions: baseOpts,
		DiscoveryInterval:                 10 * time.Minute,
	}

	cmd := &cobra.Command{
		Short: "Launch the nginx custom metrics API adapter server",
		Long:  "Launch the nginx custom metrics API adapter server",
		RunE: func(c *cobra.Command, args []string) error {
			if err := adapter.Complete(); err != nil {
				return err
			}
			if err := adapter.Validate(args); err != nil {
				return err
			}
			if err := adapter.Run(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	adapter.SecureServing.AddFlags(flags)
	adapter.Authentication.AddFlags(flags)
	adapter.Authorization.AddFlags(flags)
	adapter.Features.AddFlags(flags)
	flags.StringVar(&adapter.RemoteKubeConfigFile, "lister-kubeconfig", adapter.RemoteKubeConfigFile, ""+
		"kubeconfig file pointing at the 'core' kubernetes server with enough rights to list "+
		"any described objects")
	flags.DurationVar(&adapter.DiscoveryInterval, "discovery-interval", 20000000000, "interval at which to refresh API discovery information")
	flags.StringVar(&adapter.Label, "label", "ingress-nginx", "The label for the ingress pods")
	flags.StringVar(&adapter.MetricsPort, "metrics-port", "10254", "The port on the pods that delivers prometheus style metrics")
	flags.StringVar(&adapter.MetricsPath, "metrics-path", "/metrics", "The path on the metrics port")
	flags.Int64Var(&adapter.AverageSamples, "average-samples", 1, "The number of samples to consider for a moving average estimate")
	return cmd

}

// Run the adapter
func (a Adapter) Run(stopCh <-chan struct{}) error {
	config, err := a.Config()
	if err != nil {
		return err
	}
	var clientConfig *rest.Config
	if len(a.RemoteKubeConfigFile) > 0 {
		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: a.RemoteKubeConfigFile}
		loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})

		clientConfig, err = loader.ClientConfig()
	} else {
		clientConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		return fmt.Errorf("unable to construct lister client config to initialize provider: %v", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("unable to construct discovery client for dynamic client: %v", err)
	}

	// NB: since we never actually look at the contents of
	// the objects we fetch (beyond ObjectMeta), unstructured should be fine
	dynamicMapper, err := dynamicmapper.NewRESTMapper(discoveryClient, apimeta.InterfacesForUnstructured, a.DiscoveryInterval)
	if err != nil {
		return fmt.Errorf("unable to construct dynamic discovery mapper: %v", err)
	}

	clientPool := dynamic.NewClientPool(clientConfig, dynamicMapper, dynamic.LegacyAPIPathResolverFunc)
	if err != nil {
		return fmt.Errorf("unable to construct lister client to initialize provider: %v", err)
	}

	metricsProvider := nginxprovider.New(clientPool, dynamicMapper, a.Label, a.MetricsPort, a.MetricsPath, a.DiscoveryInterval, a.AverageSamples)
	server, err := config.Complete().New("ingress-nginx-custom-metrics-adapter", metricsProvider, nil)
	if err != nil {
		return err
	}
	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}
