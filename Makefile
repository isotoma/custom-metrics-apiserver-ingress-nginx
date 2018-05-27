GO ?= go
BINDIR := $(CURDIR)/bin
GOFLAGS :=

.PHONY: release build dep docker test dep-links 

release: build 

build: 
	CGO_ENABLED=0 $(GO) build

vendor: glide.yaml glide.lock
	glide update --strip-vendor && touch vendor

cert:
	cp ~/projects/chartlocker/certs/ca/intermediate/certs/ca-chain.cert.pem .
