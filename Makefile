# terraform-provider-powerdns Makefile
# All Go and Terraform commands run inside the Podman dev container.
# No host Go toolchain is required — see docs/development.md.

COMPOSE_DEV := deployments/compose/compose.dev.yml
COMPOSE_LAB := deployments/compose/compose.lab.yml
DC          := podman-compose -f $(COMPOSE_DEV)
EXEC        := $(DC) exec -T dev

# Provider identity (Terraform Registry namespace/type).
NAMESPACE := dantte-lp
TYPE      := powerdns
BINARY    := terraform-provider-$(TYPE)

VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || cat VERSION 2>/dev/null || echo dev)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -s -w \
  -X main.version=$(VERSION) \
  -X main.commit=$(GIT_COMMIT) \
  -X main.date=$(BUILD_DATE)

# Lab wiring for acceptance runs. Two authoritative endpoints, not one:
# views and networks are unimplemented by gpgsql (ADR 0005).
LAB_ENV := \
  PDNS_SERVER_URL=http://127.0.0.1:18081 \
  PDNS_SERVER_URL_LMDB=http://127.0.0.1:18091 \
  PDNS_RECURSOR_SERVER_URL=http://127.0.0.1:18082 \
  PDNS_API_KEY=labapikey \
  TF_ACC=1

.PHONY: help all verify up down restart logs shell \
        lab-up lab-down lab-status lab-verify \
        build install \
        test test-v test-run testacc \
        lint lint-fix lint-md lint-yaml lint-spell lint-docs tffmt tffmt-check \
        lint-py fmt-py typecheck-py py \
        vulncheck osv-scan gosec \
        docs docs-check \
        tidy download \
        clean versions

help:
	@echo "terraform-provider-powerdns"
	@echo "==========================="
	@echo "Lifecycle:  up down restart logs shell"
	@echo "Lab:        lab-up lab-down lab-status lab-verify"
	@echo "Build:      build install"
	@echo "Test:       test test-v test-run testacc"
	@echo "Quality:    lint lint-fix vulncheck osv-scan gosec"
	@echo "Python:     py (lint-py fmt-py typecheck-py)"
	@echo "Docs:       docs docs-check lint-md lint-yaml lint-spell lint-docs"
	@echo "Terraform:  tffmt tffmt-check"
	@echo "Deps:       tidy download"
	@echo "Gates:      all (pre-PR)   verify (all + lab acceptance)"
	@echo "Misc:       clean versions"

# === Dev container lifecycle ===

up:
	$(DC) up -d --build

down:
	$(DC) down

restart: down up

logs:
	$(DC) logs -f dev

shell:
	$(DC) exec dev bash

# === Lab fixture ===
# Driven through podman-py so a failure is an exception with a status code,
# not a parsed string. See scripts/automation/lab.py.

lab-up:
	python3 scripts/automation/lab.py up

lab-down:
	python3 scripts/automation/lab.py down

lab-status:
	python3 scripts/automation/lab.py status

# Asserts the fixture is the one the tests were written against — pinned
# PowerDNS versions and two genuinely different backends.
lab-verify:
	python3 scripts/automation/lab.py verify

# === Build ===

build:
	$(EXEC) go build -ldflags='$(LDFLAGS)' -o bin/$(BINARY) .

# Install for manual dev_overrides testing.
install:
	$(EXEC) go install -ldflags='$(LDFLAGS)' .

# === Test ===

# Unit tests only. Race detector always on.
test:
	$(EXEC) go test ./... -race -count=1

test-v:
	$(EXEC) go test ./... -race -count=1 -v

test-run:
	@test -n "$(RUN)" || (echo "Usage: make test-run RUN=TestX PKG=./internal/x"; exit 1)
	$(EXEC) go test -run '$(RUN)' $(PKG) -race -count=1 -v

# Acceptance tests against the lab. Requires `make lab-up` first.
testacc:
	$(EXEC) env $(LAB_ENV) go test ./... -race -count=1 -timeout 120m -v

# === Quality ===

lint:
	$(EXEC) golangci-lint run ./...

lint-fix:
	$(EXEC) golangci-lint run --fix ./...

vulncheck:
	$(EXEC) govulncheck ./...

osv-scan:
	$(EXEC) osv-scanner scan --recursive .

gosec:
	$(EXEC) golangci-lint run --enable-only gosec ./...

# === Python (automation scripts) ===
# uv manages the environment; ruff is the linter and formatter; ty is the type
# checker. Same philosophy as the Go gate: explicit allowlist, strict by
# default. Configuration lives in pyproject.toml.

lint-py:
	$(EXEC) uv run ruff check scripts/
	$(EXEC) uv run ruff format --check scripts/

fmt-py:
	$(EXEC) uv run ruff format scripts/
	$(EXEC) uv run ruff check --fix scripts/

# ty is pre-1.0. It runs in the gate because its findings have been accurate
# here, but a ty-only failure is reviewed rather than trusted blindly.
typecheck-py:
	$(EXEC) uv run ty check scripts/

py: lint-py typecheck-py

# === Terraform formatting (examples) ===

tffmt:
	$(EXEC) terraform fmt -recursive examples

tffmt-check:
	$(EXEC) terraform fmt -check -recursive examples

# === Docs ===

docs:
	$(EXEC) tfplugindocs generate --provider-name $(TYPE)

docs-check:
	$(EXEC) tfplugindocs validate --provider-name $(TYPE)

lint-md:
	$(EXEC) markdownlint-cli2 "**/*.md" "#node_modules" "#bin" "#dist" "#vendor"

lint-yaml:
	$(EXEC) yamllint -c .yamllint.yaml .

lint-spell:
	$(EXEC) cspell --no-progress --no-summary --config .cspell.json "**/*.md" "**/*.go"

lint-docs: lint-md lint-yaml lint-spell docs-check

# === Deps ===

tidy:
	$(EXEC) go mod tidy

download:
	$(EXEC) go mod download

# === Aggregate gates ===

# Pre-PR gate: everything that does not need the lab.
all: build test lint py tffmt-check lint-docs vulncheck

# Full gate: `all` plus lab acceptance on both backends.
verify: all lab-verify testacc
	@echo "verify: all gates green"

# === Clean ===

clean:
	rm -rf bin/ dist/ coverage.out coverage.html

# === Info ===

versions:
	@echo "=== Go ==="            && $(EXEC) go version
	@echo "=== Terraform ==="     && $(EXEC) terraform version
	@echo "=== OpenTofu ==="      && $(EXEC) tofu version
	@echo "=== Terragrunt ==="    && $(EXEC) terragrunt --version
	@echo "=== golangci-lint ===" && $(EXEC) golangci-lint version --short
	@echo "=== tfplugindocs ==="  && $(EXEC) tfplugindocs --version 2>/dev/null || echo installed
	@echo "=== uv ==="            && $(EXEC) uv --version
	@echo "=== ruff ==="          && $(EXEC) uv run ruff --version
	@echo "=== ty ==="            && $(EXEC) uv run ty --version
