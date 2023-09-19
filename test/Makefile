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

.PHONY: run-all
run-all: ## Runs a full node
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


.PHONY: run-node
run-node: ## Runs the node
	docker compose up -d cdk-validium-eth-tx-manager
	docker compose up -d cdk-validium-sync
	sleep 2
	docker compose up -d cdk-validium-sequencer
	docker compose up -d cdk-validium-sequence-sender
	docker compose up -d cdk-validium-l2gaspricer
	docker compose up -d cdk-validium-aggregator
	docker compose up -d cdk-validium-json-rpc

.PHONY: stop-node
stop-node: ## Stops the node
	docker compose stop cdk-validium-sequencer && docker compose rm -f cdk-validium-sequencer
	docker compose stop cdk-validium-sequence-sender && docker compose rm -f cdk-validium-sequence-sender
	docker compose stop cdk-validium-json-rpc && docker compose rm -f cdk-validium-json-rpc
	docker compose stop cdk-validium-l2gaspricer && docker compose rm -f cdk-validium-l2gaspricer
	docker compose stop cdk-validium-aggregator && docker compose rm -f cdk-validium-aggregator
	docker compose stop cdk-validium-sync && docker compose rm -f cdk-validium-sync
	docker compose stop cdk-validium-eth-tx-manager && docker compose rm -f cdk-validium-eth-tx-manager

.PHONY: run-network
run-network: ## Runs the l1 network
	docker compose up -d cdk-validium-mock-l1-network

.PHONY: stop-network
stop-network: ## Stops the l1 network
	docker compose stop cdk-validium-mock-l1-network && docker compose rm -f cdk-validium-mock-l1-network

.PHONY: run-approve-matic
run-approve-matic: ## Runs approve in node container
	docker compose up -d cdk-validium-approve

.PHONY: stop-approve-matic
stop-approve-matic: ## Stops approve in node container
	docker compose stop cdk-validium-approve && docker compose rm -f cdk-validium-approve

.PHONY: stop-test-dac-nodes
stop-test-dac-nodes:
	for i in 0 1 2 3 4 ; do \
  		docker kill cdk-data-availability-$$i && docker rm cdk-data-availability-$$i ; \
  	done

.PHONY: stop-test-dac-dbs
stop-test-dac-dbs:
	for i in 0 1 2 3 4 ; do \
  		docker kill cdk-validium-data-node-db-$$i && docker rm cdk-validium-data-node-db-$$i ; \
  	done

.PHONY: stop-test-dacs
stop-test-dacs: stop-test-dac-nodes stop-test-dac-dbs
