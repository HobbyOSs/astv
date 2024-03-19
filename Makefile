# Go parameters
GOBUILD=go build
GOTEST=gotest
BIN=astv

.PHONY: all

all: dep build test

build: dep
	$(GOBUILD) -v
	go install -v ./...

test:
	go install -v github.com/rakyll/gotest@latest
	$(GOTEST) -v ./...

clean:
	go clean

run: build
	./$(BIN)

#gen:
# 	go generate ./...

fmt:
	go fmt ./...

dep:
	go mod download
	go mod tidy
