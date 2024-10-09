# Get the latest Git tag, short commit hash, commit date, and branch name
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "unknown")
BUILD := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell git log -1 --format=%cd --date=short 2>/dev/null || echo "unknown")
BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Project and binary names
PROJECTNAME := "tracks"
BINARYNAME := "tracks"

# Go-related variables
GOBASE := $(shell pwd)
GOBIN := $(GOBASE)/bin

# Linker flags to provide version/build settings to the target
LDFLAGS := -ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD) -X=main.Date=$(DATE) -X=main.Branch=$(BRANCH)"

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

# Build the binary
.PHONY: build
## Build the binary
build: check-clean
	@echo "  > Building binary..."
	go build $(LDFLAGS) -o ./build/$(BINARYNAME) ./cmd/main.go
	@echo "  > Binary built at ./build/$(BINARYNAME)..."

# Check if the working directory is clean
.PHONY: check-clean
check-clean:
	@#git diff-index --quiet HEAD -- || (echo "Working directory is not clean. Commit or stash your changes before building." && exit 1)

# Install dependencies
.PHONY: deps
## Install dependencies
deps:
	@echo "  > Installing dependencies..."
	go mod tidy

# Clean the build artifacts
.PHONY: clean
## Clean the build artifacts
clean:
	@echo "  > Cleaning build artifacts..."
	rm -f ./build/$(BINARYNAME)

# Default target
.PHONY: all
all: deps build
