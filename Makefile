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

@PHONY: test
test: ## run tests
	@go test -v $(TEST_PACKAGE)

.PHONY: nestest
nestest: ## run nestest rom
	@go test -tags testrom -v ./nestest | tee nestest.log
	@sed -i '' '1d' nestest.log # remove the first line to match the good log
