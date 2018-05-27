FROM alpine
ADD custom-metrics-apiserver-ingress-nginx /
ENTRYPOINT /custom-metrics-apiserver-ingress-nginx