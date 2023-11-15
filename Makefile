.PHONY: release

release:
	goreleaser release --clean --snapshot

test:
	go test ./...

tidy:
	go mod tidy

lint: tidy
	golangci-lint run
