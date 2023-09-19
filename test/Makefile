STOP := docker compose down
RUN-DB := docker compose up -d cdk-data-availability-db

.PHONY: run
run: ## Runs a full data node
	docker compose up -d

.PHONY: run-local
run-local: ## Runs a full data node
	docker compose up -d cdk-data-availability-db
	docker compose up -d cdk-validium-mock-l1-network
	go run ../cmd run --cfg config/test.local.toml

.PHONY: run-dev
run-dev: run-db
	go run ../cmd run --cfg config/test.dev.toml

.PHONY: run-db
run-db: ## Runs the data store
	$(RUN-DB)

.PHONY: cdk-validium-mock-l1-network
cdk-validium-mock-l1-network: ## Run the zklidium contracts
	docker compose up -d cdk-validium-mock-l1-network

.PHONY: run-deps
run-deps: cdk-validium-mock-l1-network run-db ## Runs the service dependencies

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

.PHONY: run-node
run-node: ## Runs a full node
	docker compose up -d cdk-validium-state-db
	docker compose up -d cdk-validium-pool-db
	docker compose up -d cdk-validium-event-db
	docker compose up -d cdk-validium-mock-l1-network
	sleep 1
	docker compose up -d cdk-validium-prover
	docker compose up -d cdk-validium-approve
	sleep 3
	docker compose up -d cdk-validium-sync
	sleep 2
	docker compose up -d cdk-validium-eth-tx-manager
	docker compose up -d cdk-validium-sequencer
	docker compose up -d cdk-validium-sequence-sender
	docker compose up -d cdk-validium-l2gaspricer
	docker compose up -d cdk-validium-aggregator
	docker compose up -d cdk-validium-json-rpc


