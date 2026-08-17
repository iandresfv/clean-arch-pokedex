.DEFAULT_GOAL := help

# Tools installed with `go install` land in GOPATH/bin, which is not always on
# PATH. Resolving it here keeps the targets working on a fresh machine.
GOBIN          := $(shell go env GOPATH)/bin
BINARY         := bin/server
SEEDER         := bin/seeder
VERSION        ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS        := -s -w -X main.version=$(VERSION)
COMPOSE_DEV    := deploy/compose/docker-compose.dev.yml
COMPOSE_PROD   := deploy/compose/docker-compose.yml
MIGRATIONS_DIR := migrations

# Assembled from the same variables the application reads, so `make migrate-up`
# and the server always target the same database.
DB_USER     ?= postgres
DB_PASSWORD ?= postgres
DB_HOST     ?= localhost
DB_PORT     ?= 5432
DB_NAME     ?= pokedex
DB_SSLMODE  ?= disable
DATABASE_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# --- Build and run ---------------------------------------------------------

.PHONY: run
run: ## Run the API server
	go run ./cmd/server/

.PHONY: build
build: ## Build the server binary into bin/
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/server/

.PHONY: build-seeder
build-seeder: ## Build the seeder binary into bin/
	go build -ldflags "$(LDFLAGS)" -o $(SEEDER) ./cmd/seeder/

.PHONY: clean
clean: ## Remove build artifacts and coverage output
	rm -rf bin/ coverage.out coverage.html tmp/

# --- Quality ---------------------------------------------------------------

.PHONY: test
test: ## Run all tests with the race detector
	go test -race ./...

.PHONY: test-short
test-short: ## Run only tests that need no Docker
	go test -race -short ./...

.PHONY: test-cover
test-cover: ## Run tests and write an HTML coverage report
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1

.PHONY: bench
bench: ## Run benchmarks
	go test -bench=. -benchmem -run '^$$' ./...

.PHONY: lint
lint: ## Run golangci-lint
	$(GOBIN)/golangci-lint run

.PHONY: fmt
fmt: ## Format all Go source
	gofmt -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: check
check: fmt vet lint test ## Run the full local quality gate

.PHONY: tidy
tidy: ## Tidy and verify module dependencies
	go mod tidy
	go mod verify

# --- Database --------------------------------------------------------------

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	$(GOBIN)/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Roll back the most recent migration
	$(GOBIN)/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

.PHONY: migrate-drop
migrate-drop: ## Drop everything in the database (destructive)
	$(GOBIN)/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" drop -f

.PHONY: migrate-version
migrate-version: ## Print the current schema version
	$(GOBIN)/migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

.PHONY: migrate-new
migrate-new: ## Create a migration pair: make migrate-new NAME=add_table
	@test -n "$(NAME)" || (echo "NAME is required, e.g. make migrate-new NAME=add_table" && exit 1)
	$(GOBIN)/migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

.PHONY: sqlc
sqlc: ## Regenerate type-safe query code
	cd sqlc && $(GOBIN)/sqlc generate

.PHONY: sqlc-verify
sqlc-verify: ## Fail when generated code is out of date
	cd sqlc && $(GOBIN)/sqlc diff

.PHONY: seed
seed: ## Populate the database from PokeAPI
	go run ./cmd/seeder/ --limit=$(or $(LIMIT),1302)

# --- Docker ----------------------------------------------------------------

.PHONY: docker-dev
docker-dev: ## Start the development stack with hot reload
	docker compose -f $(COMPOSE_DEV) up --build

.PHONY: docker-prod
docker-prod: ## Start the production-like stack
	docker compose -f $(COMPOSE_PROD) up --build -d

.PHONY: docker-down
docker-down: ## Stop all stacks and remove volumes
	-docker compose -f $(COMPOSE_DEV) down -v
	-docker compose -f $(COMPOSE_PROD) down -v

.PHONY: db-up
db-up: ## Start only PostgreSQL
	docker compose -f $(COMPOSE_DEV) up -d postgres

.PHONY: db-shell
db-shell: ## Open a psql shell in the database container
	docker compose -f $(COMPOSE_DEV) exec postgres psql -U $(DB_USER) -d $(DB_NAME)

# --- TLS -------------------------------------------------------------------

.PHONY: tls-cert
tls-cert: ## Generate a self-signed certificate for local HTTPS
	go run ./cmd/tlscert/ --host localhost --out certs/

# --- Kubernetes ------------------------------------------------------------

.PHONY: k8s-dev
k8s-dev: ## Apply the dev overlay to the current cluster
	kubectl apply -k deploy/k8s/overlays/dev

.PHONY: k8s-prod
k8s-prod: ## Apply the production overlay
	kubectl apply -k deploy/k8s/overlays/prod

.PHONY: k8s-delete
k8s-delete: ## Remove the dev overlay
	kubectl delete -k deploy/k8s/overlays/dev --ignore-not-found

.PHONY: tilt-up
tilt-up: ## Start the Tilt development loop on minikube
	tilt up

.PHONY: tilt-down
tilt-down: ## Tear down everything Tilt created
	tilt down

# --- Documentation ---------------------------------------------------------

.PHONY: validate-openapi
validate-openapi: ## Check that every served route is documented, and vice versa
	go test ./internal/router/ -run TestEveryRouteIsDocumented -v

.PHONY: docs-sync
docs-sync: ## Copy the spec into the embedded asset directory
	cp docs/openapi.yaml internal/handler/assets/openapi.yaml

.PHONY: tools
tools: ## Install the development tooling this Makefile expects
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest
