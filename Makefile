include version.mk

ARCH := $(shell arch)

ifeq ($(ARCH),x86_64)
	ARCH = amd64
else 
	ifeq ($(ARCH),aarch64)
		ARCH = arm64
	endif
endif
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/dist
GOENVVARS := GOBIN=$(GOBIN) CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH)
GOBINARY := supernets2.0-data-availability
GOCMD := $(GOBASE)/cmd

LDFLAGS += -X 'github.com/0xPolygon/supernets2.0-data-availability.Version=$(VERSION)'
LDFLAGS += -X 'github.com/0xPolygon/supernets2.0-data-availability.GitRev=$(GITREV)'
LDFLAGS += -X 'github.com/0xPolygon/supernets2.0-data-availability.GitBranch=$(GITBRANCH)'
LDFLAGS += -X 'github.com/0xPolygon/supernets2.0-data-availability.BuildDate=$(DATE)'

.PHONY: build
build: ## Builds the binary locally into ./dist
	$(GOENVVARS) go build -ldflags "all=$(LDFLAGS)" -o $(GOBIN)/$(GOBINARY) $(GOCMD)

.PHONY: build-docker
build-docker: ## Builds a docker image with the node binary
	docker build -t supernets2.0-data-availability -f ./Dockerfile .

.PHONY: build-docker-nc
build-docker-nc: ## Builds a docker image with the node binary - but without build cache
	docker build --no-cache=true -t supernets2.0-data-availability -f ./Dockerfile .

.PHONY: install-linter
install-linter: ## Installs the linter
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.52.2

.PHONY: lint
lint: ## Runs the linter
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/golangci-lint run

## Help display.
## Pulls comments from beside commands and prints a nicely formatted
## display with the commands and their usage information.
.DEFAULT_GOAL := help

.PHONY: help
help: ## Prints this help
		@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
