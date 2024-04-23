VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)

PROJECTNAME := "tracks"
BINARYNAME := "tracks"

# Go-related variables
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin

# Linker flags to provide version/build settings to the target
LDFLAGS := -ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

# Build the binary
.PHONY: build
## Build the binary
build:
	@echo "  > Building binary..."
	go build $(LDFLAGS) -o ./$(BINARYNAME) ./cmd/main.go

# Install dependencies
.PHONY: install
## Install missing dependencies
install:
	@echo "  > Installing dependencies..."
	go mod download

# Clean build files
.PHONY: clean
## Clean build cache
clean:
	@echo "  > Cleaning build cache..."
	go clean

# Show help information
.PHONY: help
## Show this help screen
help:
	@sed -n 's/^##//p' Makefile
