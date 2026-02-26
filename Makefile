BINARY=ado
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build install test lint clean fmt vet

build:
	go build $(LDFLAGS) -o bin/$(BINARY) .

install: build
	cp bin/$(BINARY) $(GOPATH)/bin/$(BINARY)

test:
	go test ./... -v

lint:
	golangci-lint run

fmt:
	gofmt -s -w .

vet:
	go vet ./...

clean:
	rm -rf bin/ dist/

all: fmt vet lint test build
