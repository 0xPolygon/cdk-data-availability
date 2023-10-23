STOP := docker compose down

.PHONY: run
run: stop ## Runs a full node for e2e
	docker compose up -d zkevm-state-db
	docker compose up -d zkevm-pool-db
	docker compose up -d cdk-data-availability-db
	docker compose up -d l1
	sleep 1
	docker compose up -d cdk-data-availability
	docker compose up -d zkevm-prover
	docker compose up -d zkevm-node

.PHONY: stop
stop: ## Stop a full data node
	make stop-dacs
	$(STOP)

.PHONY: stop-dacs
stop-dacs:
	@./stop-dacs > /dev/null 2>&1

.PHONY: test-e2e
test-e2e: run ## Runs the E2E tests
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

### node commands called by testing operations ###

.PHONY: run-node
run-node: ## Runs the node
	docker compose up -d zkevm-node

.PHONY: stop-node
stop-node: ## Stops the node
	docker compose stop zkevm-node && docker compose rm -f zkevm-node

.PHONY: run-network
run-network: ## Runs the l1 network
	docker compose up -d l1

.PHONY: stop-network
stop-network: ## Stops the l1 network
	docker compose stop l1 && docker compose rm -f l1

.PHONY: run-local
run-local: run-network ## Run network and db for local instance
	docker compose up -d cdk-data-availability-db

.PHONY: stop-local
stop-local: stop-network ## Run network and db for local instance
	docker compose stop cdk-data-availability-db && docker compose rm -f cdk-data-availability-db


