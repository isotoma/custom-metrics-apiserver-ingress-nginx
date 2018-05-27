package nginxmetrics

import (
	"io"
	"time"

	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/cmd/server"
	"github.com/spf13/cobra"
)

// Adapter is the adapter that takes nginx metrics and sends them to k8s
type Adapter struct {
	*server.CustomMetricsAdapterServerOptions
	// RemoteKubeConfigFile is the config used to list pods from the master API server
	RemoteKubeConfigFile string
	// DiscoveryInterval is the interval at which discovery information is refreshed
	DiscoveryInterval time.Duration
	// EnableCustomMetricsAPI switches on sample apiserver for Custom Metrics API
	EnableCustomMetricsAPI bool
	// EnableExternalMetricsAPI switches on sample apiserver for External Metrics API
	EnableExternalMetricsAPI bool
}

// Start starts the server
func Start(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	baseOpts := server.NewCustomMetricsAdapterServerOptions(out, errOut)
	adapter := Adapter{
		CustomMetricsAdapterServerOptions: baseOpts,
		DiscoveryInterval:                 10 * time.Minute,
		EnableCustomMetricsAPI:            true,
		EnableExternalMetricsAPI:          true,
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
	flags.DurationVar(&adapter.DiscoveryInterval, "discovery-interval", adapter.DiscoveryInterval, ""+
		"interval at which to refresh API discovery information")
	flags.BoolVar(&adapter.EnableCustomMetricsAPI, "enable-custom-metrics-api", adapter.EnableCustomMetricsAPI, ""+
		"whether to enable Custom Metrics API")
	flags.BoolVar(&adapter.EnableExternalMetricsAPI, "enable-external-metrics-api", adapter.EnableExternalMetricsAPI, ""+
		"whether to enable External Metrics API")

	return cmd

}

func (a Adapter) Run(stopCh <-chan struct{}) error {
	// config, err := a.Config()
	// if err != nil {
	// return err
	// }
	return nil
}
