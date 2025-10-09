.PHONY: build build-release clean install

# Version
VERSION ?= $(shell grep 'Version' meta.go | awk '{print $$3}' | tr -d '"')
REPO := https://github.com/ole3021/ssh-messer

# Build for multiple platforms
build-release:
	@echo "Building for multiple platforms..."
	@mkdir -p dist
	@GOOS=darwin GOARCH=amd64 go build -ldflags "-w -s -X ssh-messer/meta.Version=$(VERSION)" -o dist/ssh-messer-darwin-amd64 ./cmd/tui
	@GOOS=darwin GOARCH=arm64 go build -ldflags "-w -s -X ssh-messer/meta.Version=$(VERSION)" -o dist/ssh-messer-darwin-arm64 ./cmd/tui
	@GOOS=linux GOARCH=amd64 go build -ldflags "-w -s -X ssh-messer/meta.Version=$(VERSION)" -o dist/ssh-messer-linux-amd64 ./cmd/tui
	@GOOS=linux GOARCH=arm64 go build -ldflags "-w -s -X ssh-messer/meta.Version=$(VERSION)" -o dist/ssh-messer-linux-arm64 ./cmd/tui

# Build for local
build:
	@go build -o ssh-messer ./cmd/tui

# Clean
clean:
	@rm -rf dist ssh-messer

# Install to local
install: build
	@cp ssh-messer /usr/local/bin/ssh-messer
	@echo "Installed to /usr/local/bin/ssh-messer"