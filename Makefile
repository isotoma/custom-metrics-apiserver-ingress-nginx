.PHONY: release build vendor

release: build 

build: 
	CGO_ENABLED=0 go build
	strip ./custom-metrics-apiserver-ingress-nginx

vendor: glide.yaml glide.lock
	glide update --strip-vendor && touch vendor
