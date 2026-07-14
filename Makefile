.PHONY: fmt vet lint test build check release-snapshot

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test ./...

build:
	go build -trimpath -o bin/debugdoc ./cmd/debugdoc

check: vet lint test build

release-snapshot:
	goreleaser release --snapshot --clean
