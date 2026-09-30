.PHONY: tidy test bench lint gen-r0 gen-iab help

.DEFAULT_GOAL := help

tidy: ## go mod tidy
	go mod tidy

test: ## go test ./...
	go test ./...

bench: ## benchmarks only, with memory stats
	go test -bench=. -benchmem -run=^$$ ./...

lint: ## go vet ./...
	go vet ./...

gen-r0: ## generate r0 dictionaries from r0/datasets
	go run ./r0/cmd/gendict

gen-iab: ## generate iab dictionaries from iab/datasets
	go run ./iab/cmd/gendict
	go run ./iab/cmd/genlight

help: ## show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
