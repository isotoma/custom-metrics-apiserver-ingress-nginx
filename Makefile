GO ?= go
BINDIR := $(CURDIR)/bin
GOFLAGS :=

.PHONY: release build dep docker test dep-links 

release: build 

build: .vendor-links.stamp compile

compile:
	CGO_ENABLED=0 $(GO) build

.vendor-links.stamp: vendor
	mkdir -p vendor/github.com/isotoma/custom-metrics-apiserver-ingress-nginx
	rm -rf vendor/github.com/isotoma/custom-metrics-apiserver-ingress-nginx/pkg
	cd vendor/github.com/isotoma/custom-metrics-apiserver-ingress-nginx && ln -s ../../../../pkg .
	touch .vendor-links.stamp

vendor: glide.yaml glide.lock
	glide update --strip-vendor && touch vendor

cert:
	cp ~/projects/chartlocker/certs/ca/intermediate/certs/ca-chain.cert.pem .
