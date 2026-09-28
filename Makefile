.PHONY: tidy test bench lint help

.DEFAULT_GOAL := help

tidy: ## go mod tidy
	go mod tidy

test: ## go test ./...
	go test ./...

bench: ## benchmarks only, with memory stats
	go test -bench=. -benchmem -run=^$$ ./...

lint: ## go vet ./...
	go vet ./...

help: ## show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-8s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
