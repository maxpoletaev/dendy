.DEFAULT_GOAL := help

TEST_PACKAGE = ./...
PWD = $(shell pwd)
GO_MODULE = github.com/maxpoletaev/dendy
COMMIT_HASH = $(shell git rev-parse --short HEAD)
PROTO_FILES = $(shell find . -type f -name '*.proto')
PGO_PROFILES = $(shell find profiles -type f -name '*.pprof')

.PHONY: help
help: ## print help (this message)
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sed -n 's/^\(.*\): \(.*\)## \(.*\)/\1;\3/p' \
	| column -t  -s ';'

.PHONY: pgo
pgo: ## generate default.pgo
	@echo "--------- running: $@ ---------"
	go tool pprof -proto $(PGO_PROFILES) > default.pgo

.PHONY: build
build: ## build dendy
	@echo "--------- running: $@ ---------"
	CGO_ENABLED=1 GODEBUG=cgocheck=0 go build -pgo=default.pgo -o=dendy ./cmd/dendy

PHONY: test
test: ## run tests
	@echo "--------- running: $@ ---------"
	@go test -v $(TEST_PACKAGE)

.PHONY: nestest
nestest: ## run nestest rom
	@echo "--------- running: $@ ---------"
	go test -tags testrom -v ./nestest > nestest.log
	sed -i '1d' nestest.log # remove the first line to match the good.log
