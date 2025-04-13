.PHONY: build clean test

build:
	@go build -ldflags="-s -w" -o bin/rotate ./cmd/rotate

test:
	@go test -v ./...

clean:
	@rm -rf bin/