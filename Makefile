.DEFAULT_GOAL := help

TEST_PACKAGE = ./...
GO_MODULE = github.com/maxpoletaev/dendy

.PHONY: help
help:  ## print help (this message)
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sed -n 's/^\(.*\): \(.*\)## \(.*\)/\1;\3/p' \
	| column -t  -s ';'

.PHONY: build
build: ## build dendy
	go build -o bin/dendy ./cmd/dendy

.PHONY: nestest
nestest: ## run dendy
	go test -v ./nestest | tee nestest.log
