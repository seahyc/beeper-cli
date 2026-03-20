BINARY_NAME = beeper
BUILD_DIR = .
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -s -w \
	-X github.com/yjwong/beeper-cli/internal/cmd.version=$(VERSION) \
	-X github.com/yjwong/beeper-cli/internal/cmd.commit=$(COMMIT) \
	-X github.com/yjwong/beeper-cli/internal/cmd.date=$(DATE)

.PHONY: build install clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/beeper

install:
	go build -ldflags "$(LDFLAGS)" -o $(HOME)/.local/bin/$(BINARY_NAME) ./cmd/beeper

clean:
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
