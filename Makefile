.DEFAULT_GOAL := help

TEST_PACKAGE = ./...
GO_MODULE = github.com/maxpoletaev/dendy
PROTO_FILES = $(shell find . -type f -name '*.proto')

.PHONY: help
help:  ## print help (this message)
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sed -n 's/^\(.*\): \(.*\)## \(.*\)/\1;\3/p' \
	| column -t  -s ';'

.PHONY: build
build: ## build dendy
	@echo "--------- running: $@ ---------"
	CGO_ENABLED=1 GODEBUG=cgocheck=0 go build -o dendy ./cmd/dendy

.PHONY: build_win64
build_win64: ## build dendy for windows
	@echo "--------- running: $@ ---------"
	CGO_ENABLED=1 GODEBUG=cgocheck=0 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_LDFLAGS="-static-libgcc -static -lpthread" go build -o dendy_win64.exe ./cmd/dendy

PHONY: test
test: ## run tests
	@echo "--------- running: $@ ---------"
	@go test -v $(TEST_PACKAGE)

.PHONY: nestest
nestest: ## run nestest rom
	@echo "--------- running: $@ ---------"
	go test -tags testrom -v ./nestest > nestest.log
	sed -i '1d' nestest.log # remove the first line to match the good.log
