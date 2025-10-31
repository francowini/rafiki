# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

run: 
	go run api/services/partner/main.go | go run api/tooling/logfmt/main.go

help: 
	go run api/services/partner/main.go --help

version:
	go run api/services/partner/main.go --version

curl-ready:
	curl -i http://localhost:3000/readiness

# ==============================================================================
# Define dependencies

GOLANG          := golang:1.25
ALPINE          := alpine:3.22

BASE_IMAGE_NAME := localhost/rafiki
VERSION         := 0.0.1
PARTNER_IMAGE   := $(BASE_IMAGE_NAME)/partner:$(VERSION)

# ==============================================================================
# Modules support

deps-reset:
	git checkout -- go.mod
	go mod tidy
	go mod vendor

tidy:
	go mod tidy
	go mod vendor

deps-list:
	go list -m -u -mod=readonly all

deps-upgrade:
	go get -u -v ./...
	go mod tidy
	go mod vendor

deps-cleancache:
	go clean -modcache

list:
	go list -mod=mod all
