VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)

PROJECTNAME := "tracks"
BINARYNAME := "tracks"

# Go related variables.
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

## build: Compile the binary.
build:
	@echo "  >  Building binary..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build $(LDFLAGS) -o $(GOBIN)/$(BINARYNAME) ./cmd/main.go

## install: Install missing dependencies.
install:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go mod download

## clean: Clean build files. Runs `go clean` internally
clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean



## help: Show this help screen
help : Makefile
	@sed -n 's/^##//p' $<