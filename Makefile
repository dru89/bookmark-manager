BIN       := bk
MODULE    := github.com/dru89/bookmark-manager
CMD       := ./cmd/bk
PREFIX    := /usr/local
GOBIN     ?= $(shell go env GOPATH)/bin

VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -s -w -X main.version=$(VERSION)

.PHONY: all build install uninstall test lint clean

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) $(CMD)

install:
	go build -ldflags "$(LDFLAGS)" -o $(GOBIN)/$(BIN) $(CMD)

uninstall:
	rm -f $(GOBIN)/$(BIN)

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BIN)

# Fetch dependencies
.PHONY: deps
deps:
	go mod tidy
	go mod download
