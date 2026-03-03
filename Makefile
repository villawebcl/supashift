APP=supashift

.PHONY: build test lint fmt tidy ci smoke

build:
	go build -o $(APP) ./cmd/$(APP)

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

ci: fmt lint test build

smoke: build
	./scripts/smoke.sh ./$(APP)
