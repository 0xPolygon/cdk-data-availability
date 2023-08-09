STOP := docker compose down
RUN-DB := docker compose up -d supernets2-data-availability-db

.PHONY: run
run: ## Runs a full data node
	docker compose up -d

.PHONY: run-local
run-local: ## Runs a full data node
	docker compose up -d supernets2-data-availability-db
	docker compose up -d supernets2-mock-l1-network
	go run ../cmd run --cfg config/test.local.toml

.PHONY: run-dev
run-dev: run-db
	go run ../cmd run --cfg config/test.dev.toml

.PHONY: run-db
run-db: ## Runs the data store
	$(RUN-DB)

.PHONY: supernets2-mock-l1-network
supernets2-mock-l1-network: ## Run the zklidium contracts
	docker compose up -d supernets2-mock-l1-network

.PHONY: run-deps
run-deps: supernets2-mock-l1-network run-db ## Runs the service dependencies

.PHONY: stop
stop: ## Stop a full data node
	$(STOP)

.PHONY: test-e2e
test-e2e: stop ## Runs the E2E tests
	docker compose up -d
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -count=1 -race -v -p 1 -timeout 600s ./e2e/...

.PHONY: test-unit
test-unit: ## Runs the unit tests, skips e2e
	go test `go list ../... | grep -v /e2e`

## Help display.
## Pulls comments from beside commands and prints a nicely formatted
## display with the commands and their usage information.
.DEFAULT_GOAL := help

.PHONY: help
help: ## Prints this help
		@grep -h -E '^[a-zA-Z0-9_-]*:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
