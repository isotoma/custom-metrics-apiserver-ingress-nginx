# custom-metrics-apiserver-ingress-nginx

A Kubernetes API Server custom metrics adapter for Ingress Nginx.

This provides http request rate statistics in a format that can be used by the Horizontal Pod Autoscaler to scale a deployment. It uses the prometheus metrics endpoint provided by the ingress controller. This is the case even if you are not using prometheus.

## Setting it up

### Enable custom VTS metrics for prometheus

First you need to turn on the custom VTS metrics for prometheus, thusly:

https://github.com/kubernetes/ingress-nginx/tree/master/docs/examples/customization/custom-vts-metrics-prometheus

You should also do the rather opaque step "Customize ingress".

### Run the custom metrics server

There are some manifests in this repository in custom-metrics.yaml that will get you going. This should be packaged for helm really.

There are some options for the custom metrics server that you should consider:

#### Sample averaging

`--average-samples int`
    
The number of samples to consider for a moving average estimate (default 1)

This provides an estimated exponential smoothing, averaging over the number of samples specified.

#### Discovery interval

`--discovery-interval duration`

Interval at which to refresh API discovery information (default 20s)

This is how frequently the ingress pods are consulted for their current totals.

#### Label

`--label string`

The label for the ingress pods (default "ingress-nginx")

So it can find the ingress pods.

#### Metrics path

`--metrics-path string`

The path on the metrics port (default "/metrics")

#### Metrics port

`--metrics-port string`

The port on the pods that delivers prometheus style metrics (default "10254")

### Getting metrics

You can query the custom metrics api yourself to check the values with:

`kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/NAMESPACE/services/SERVICE/ingress_requests_per_second" | jq`

replace NAMESPACE and SERVICE with the namespace and service names respectively.

### Setting up an HPA

Here is an example HPA manifest using the custom metrics:

    apiVersion: autoscaling/v2beta1
    kind: HorizontalPodAutoscaler
    metadata:
    name: myapp
    namespace: default
    spec:
    maxReplicas: 20
    minReplicas: 2
    scaleTargetRef:
        kind: Deployment
        name: myapp
    metrics:
    - type: Object
        object:
        target:
            kind: Service
            name: myapp
        metricName: ingress_requests_per_second
        targetValue: 20

