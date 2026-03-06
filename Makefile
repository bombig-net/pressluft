SHELL := /bin/sh

GO ?= go
NPM ?= pnpm
APP_BINARY ?= bin/pressluft
AGENT_BINARY ?= bin/pressluft-agent
DEV_API_PORT ?= 8081
DEV_UI_PORT ?= 8080
DEV_UI_HOST ?= 0.0.0.0
GO_TEST ?= $(GO) test

WEB_DIR := web
EMBED_DIST_DIR := internal/server/dist

.PHONY: build dev run clean format fmt-check lint test test-unit test-integration test-fast test-smoke smoke check check-fast agent agent-dev all validate-profiles validate-profile-schema validate-profile-consistency ansible-syntax ansible-check ansible-validate

build:
	@if [ ! -d "$(WEB_DIR)/node_modules" ]; then $(NPM) --prefix "$(WEB_DIR)" install; fi
	$(NPM) --prefix "$(WEB_DIR)" run generate
	test -f "$(WEB_DIR)/.output/public/index.html"
	rm -rf "$(EMBED_DIST_DIR)"
	mkdir -p "$(EMBED_DIST_DIR)"
	touch "$(EMBED_DIST_DIR)/.gitkeep"
	cp -R "$(WEB_DIR)/.output/public/." "$(EMBED_DIST_DIR)/"
	mkdir -p "$(dir $(APP_BINARY))"
	$(GO) build -o "$(APP_BINARY)" ./cmd

agent:
	CGO_ENABLED=0 $(GO) build -o "$(AGENT_BINARY)" ./cmd/pressluft-agent

agent-dev:
	CGO_ENABLED=0 $(GO) build -tags dev -o "$(AGENT_BINARY)" ./cmd/pressluft-agent

all: build agent

dev: agent-dev
	@if [ ! -d "$(WEB_DIR)/node_modules" ]; then $(NPM) --prefix "$(WEB_DIR)" install; fi
	DEV_API_PORT="$(DEV_API_PORT)" DEV_UI_PORT="$(DEV_UI_PORT)" DEV_UI_HOST="$(DEV_UI_HOST)" WEB_DIR="$(WEB_DIR)" NPM="$(NPM)" GO="$(GO)" ./ops/scripts/dev.sh

run: build
	./$(APP_BINARY)

format:
	$(GO) fmt ./...

fmt-check:
	@unformatted="$$(gofmt -l cmd internal)"; \
	if [ -n "$$unformatted" ]; then \
		printf '%s\n' "$$unformatted"; \
		exit 1; \
	fi

lint:
	$(GO) vet ./...

test:
	$(MAKE) test-unit
	$(MAKE) test-integration

test-unit:
	$(GO_TEST) ./cmd ./cmd/pressluft-agent ./internal/agent ./internal/agentauth ./internal/agentcommand ./internal/database ./internal/dispatch ./internal/platform ./internal/registration ./internal/worker ./internal/ws

test-integration:
	$(GO_TEST) -count=1 ./internal/activity ./internal/orchestrator ./internal/provider/... ./internal/runner/ansible ./internal/server ./internal/server/profiles

test-fast:
	$(MAKE) fmt-check
	$(MAKE) lint
	$(GO) test ./...

test-smoke:
	./ops/tests/smoke/run.sh

smoke: test-smoke

validate-profile-schema:
	$(GO_TEST) -count=1 ./internal/server/profiles -run TestProfileArtifactsSatisfySchema

validate-profile-consistency:
	$(GO_TEST) -count=1 ./internal/server/profiles -run TestRegistryMatchesProfileArtifacts

validate-profiles:
	$(MAKE) validate-profile-schema
	$(MAKE) validate-profile-consistency

ansible-syntax:
	ansible-playbook -i localhost, -c local --syntax-check ops/ansible/playbooks/configure.yml
	ansible-playbook -i localhost, -c local --syntax-check ops/ansible/playbooks/delete_server.yml
	ansible-playbook -i localhost, -c local --syntax-check ops/ansible/playbooks/rebuild_server.yml
	ansible-playbook -i localhost, -c local --syntax-check ops/ansible/playbooks/resize_server.yml

ansible-check:
	./ops/scripts/ansible_check_configure.sh

ansible-validate:
	$(MAKE) ansible-syntax
	$(MAKE) validate-profile-schema
	$(MAKE) validate-profile-consistency

check-fast:
	$(MAKE) fmt-check
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) ansible-validate

check: check-fast build

clean:
	rm -f "$(APP_BINARY)" "$(AGENT_BINARY)"
