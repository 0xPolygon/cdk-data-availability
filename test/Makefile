STOP := docker compose down
RUN-DB := docker compose up -d supernets2.0-data-availability-db

.PHONY: run
run: ## Runs a full data node
	docker compose up -d

.PHONY: run-local
run-local: ## Runs a full data node
	docker compose up -d supernets2.0-data-availability-db
	docker compose up -d zkevm-mock-l1-network
	go run ../cmd run --cfg config/test.local.toml

.PHONY: run-db
run-db: ## Runs the data store
	$(RUN-DB)

.PHONY: zkevm-mock-l1-network
zkevm-mock-l1-network: ## Run the zklidium contracts
	docker compose up -d zkevm-mock-l1-network

.PHONY: run-deps
run-deps: zkevm-mock-l1-network run-db # run the service dependencies

.PHONY: stop
stop: ## Stop a full data node
	$(STOP)

.PHONY: test-e2e
test-e2e: stop ## Runs the E2E tests
	docker compose up -d
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -count=1 -race -v -p 1 -timeout 600s ./e2e/...

.PHONY: test-unit
test-unit: stop ## Runs the unit tests
	$(RUN-DB)
	trap '$(STOP)' EXIT; MallocNanoZone=0 go test -count=1 -short -race -v -p 1 -timeout 600s ../...
