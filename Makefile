.DEFAULT_GOAL := help

TEST_PACKAGE = ./...
PWD = $(shell pwd)
GO_MODULE = github.com/maxpoletaev/dendy
COMMIT_HASH = $(shell git rev-parse --short HEAD)
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
	CGO_ENABLED=1 GODEBUG=cgocheck=0 go build -pgo=default.pgo -o=bin/dendy ./cmd/dendy
	CGO_ENABLED=0 go build -pgo=off -o=bin/dendy-relay ./cmd/dendy-relay

.PHONY: build-x
build-x:  ## cross compile for linux_amd64 and win_amd64 targets (requires docker)
	@echo "--------- running: $@ ---------"
	docker buildx build -f build/Dockerfile -t dendy-builder .
	docker run --rm -v $(PWD):/src dendy-builder build/build.sh

.PHONY: build-wasm
build-wasm: ## build wasm
	@echo "--------- running: $@ ---------"
	cp "$(shell go env GOROOT)/lib/wasm/wasm_exec.js" ./web
	GOOS=js GOARCH=wasm go build -pgo=default.pgo -o=web/dendy.wasm ./cmd/dendy-wasm

.PHONY: build-wasm-tinygo
build-wasm-tinygo: ## build wasm
	@echo "--------- running: $@ ---------"
	cp "$(shell tinygo env TINYGOROOT)/targets/wasm_exec.js" ./web
	tinygo build -no-debug -gc=conservative -target=wasm -opt=2 -o=web/dendy.wasm ./cmd/dendy-wasm

PHONY: test
test: ## run tests
	@echo "--------- running: $@ ---------"
	@go test -v $(TEST_PACKAGE)

.PHONY: nestest
nestest: ## run nestest rom
	@echo "--------- running: $@ ---------"
	go test -tags testrom -v ./nestest > nestest.log
	sed -i '1d' nestest.log # remove the first line to match the good.log
