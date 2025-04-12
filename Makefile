.PHONY: build build-all clean release install

# Program variables
PROGRAM_NAME=mocker
DOCKER_PLUGIN_NAME=docker-model
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
LDFLAGS=-ldflags "-X main.AppVersion=$(VERSION)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Directories
DIST_DIR=dist

# Build targets
build:
	$(GOBUILD) $(LDFLAGS) -o $(DOCKER_PLUGIN_NAME) -v

# Cross-platform builds
build-all: clean setup-dist build-linux build-mac build-windows

setup-dist:
	mkdir -p $(DIST_DIR)

build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(DOCKER_PLUGIN_NAME)-linux-amd64 -v
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(DOCKER_PLUGIN_NAME)-linux-arm64 -v
	tar -czf $(DIST_DIR)/$(DOCKER_PLUGIN_NAME)-linux-amd64.tar.gz -C $(DIST_DIR) $(DOCKER_PLUGIN_NAME)-linux-amd64
	tar -czf $(DIST_DIR)/$(DOCKER_PLUGIN_NAME)-linux-arm64.tar.gz -C $(DIST_DIR) $(DOCKER_PLUGIN_NAME)-linux-arm64

build-mac:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(DOCKER_PLUGIN_NAME)-darwin-amd64 -v
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(DOCKER_PLUGIN_NAME)-darwin-arm64 -v
	tar -czf $(DIST_DIR)/$(DOCKER_PLUGIN_NAME)-darwin-amd64.tar.gz -C $(DIST_DIR) $(DOCKER_PLUGIN_NAME)-darwin-amd64
	tar -czf $(DIST_DIR)/$(DOCKER_PLUGIN_NAME)-darwin-arm64.tar.gz -C $(DIST_DIR) $(DOCKER_PLUGIN_NAME)-darwin-arm64

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(DOCKER_PLUGIN_NAME)-windows-amd64.exe -v
	zip -j $(DIST_DIR)/$(DOCKER_PLUGIN_NAME)-windows-amd64.zip $(DIST_DIR)/$(DOCKER_PLUGIN_NAME)-windows-amd64.exe

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(DIST_DIR)
	rm -f $(DOCKER_PLUGIN_NAME)*

# Create a release (builds all platforms and packages them)
release: build-all

# Install to the local Docker CLI plugins directory
install: build
	mkdir -p ~/.docker/cli-plugins
	cp $(DOCKER_PLUGIN_NAME) ~/.docker/cli-plugins/
	chmod +x ~/.docker/cli-plugins/$(DOCKER_PLUGIN_NAME)