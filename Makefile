GOFMT ?= gofmt
GOVET ?= go vet

.PHONY: fmt vet lint

fmt:
	@echo "Running gofmt..."
	@$(GOFMT) -w ./cmd ./internal

vet:
	@echo "Running go vet..."
	@$(GOVET) ./...

# lint aggregates fmt + vet; add staticcheck here if available.
lint: fmt vet
