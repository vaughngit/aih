.PHONY: build test lint install clean tidy fmt

BIN := aih
PKG := ./cmd/aih

build:
	go build -o $(BIN) $(PKG)

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install $(PKG)

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

clean:
	rm -f $(BIN)
	rm -rf dist/
