SHELL := /bin/sh
GO ?= go
SERVICES := api-gateway identity facility booking notification
PROTOC ?= protoc
PROTO_INCLUDE ?= $(shell $(PROTOC) --print_include_path 2>/dev/null)
ifeq ($(PROTO_INCLUDE),)
PROTO_INCLUDE := /usr/include
endif
PROTO_FLAGS := --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative
PROTO_FILES := services/identity/pkg/grpc/identitypb/identity.proto \
	services/facility/pkg/grpc/facilitypb/facility.proto

.PHONY: help tidy run run-% tidy-% run-all run-frontend compose-up compose-down compose-migrate proto

help:
	@echo "Available targets:"
	@echo "  tidy        - go mod tidy for every service"
	@echo "  tidy-<svc>  - go mod tidy for a specific service (e.g. make tidy-identity)"
	@echo "  run-<svc>   - go run ./cmd/<svc> inside service folder"
	@echo "  run-all     - sequentially start all services (blocking)"
	@echo "  run-frontend- build & start nginx frontend container"
	@echo "  proto       - regenerate Go gRPC stubs from .proto files"
	@echo "  compose-up  - docker compose up --build"
	@echo "  compose-down- docker compose down -v"
	@echo "  compose-migrate - apply SQL migrations via docker compose exec"

# Run go mod tidy for every service module
 tidy: $(SERVICES:%=tidy-%)

# Run a single go mod tidy inside services/<name>
 tidy-%:
	@echo "==> tidy $*"
	@cd services/$* && $(GO) mod tidy

# Run go run ./cmd/<name> inside services/<name>
 run-%:
	@echo "==> run $*"
	@cd services/$* && $(GO) run ./cmd/$*

# Sequentially start all services (use separate terminals for concurrent runs)
 run-all:
	@for svc in $(SERVICES); do \
		$(MAKE) run-$$svc || exit 1; \
	done

run-frontend:
	@echo "==> docker compose up frontend"
	@docker compose up -d --build frontend

proto:
	@for file in $(PROTO_FILES); do \
		echo "==> protoc $$file"; \
		dir=$$(dirname $$file); \
		$(PROTOC) -I $$dir -I $(PROTO_INCLUDE) $(PROTO_FLAGS) $$file || exit 1; \
	done

# Build & run everything via Docker Compose
 compose-up:
	@docker compose up -d --build

# Stop containers and drop named volumes
 compose-down:
	@docker compose down

# Apply all SQL migrations using the postgres container
 compose-migrate:
	@echo "==> applying identity migrations"
	@docker compose exec -T postgres psql -U postgres -d sporthub < services/identity/migrations/001_init.sql
	@echo "==> applying facility migrations"
	@docker compose exec -T postgres psql -U postgres -d sporthub < services/facility/migrations/001_init.sql
	@echo "==> applying booking migrations"
	@docker compose exec -T postgres psql -U postgres -d sporthub < services/booking/migrations/001_init.sql
