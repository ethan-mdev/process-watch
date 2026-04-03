.PHONY: build build-all clean help

PROJECT_NAME = process-watch
OUTPUT_DIR = dist
VERSION ?= dev

# Build targets
LINUX_AMD64 = $(OUTPUT_DIR)/$(PROJECT_NAME)-linux-amd64
DARWIN_AMD64 = $(OUTPUT_DIR)/$(PROJECT_NAME)-darwin-amd64
DARWIN_ARM64 = $(OUTPUT_DIR)/$(PROJECT_NAME)-darwin-arm64
WINDOWS_AMD64 = $(OUTPUT_DIR)/$(PROJECT_NAME)-windows-amd64.exe

help:
	@echo "process-watch build targets:"
	@echo "  make build-all    - Build for all platforms (linux, darwin, windows)"
	@echo "  make $(LINUX_AMD64)"
	@echo "  make $(DARWIN_AMD64)"
	@echo "  make $(DARWIN_ARM64)"
	@echo "  make $(WINDOWS_AMD64)"
	@echo "  make clean        - Remove dist directory"

build-all: $(LINUX_AMD64) $(DARWIN_AMD64) $(DARWIN_ARM64) $(WINDOWS_AMD64)
	@echo ""
	@echo "✓ Build complete! Binaries in $(OUTPUT_DIR)/:"
	@ls -lh $(OUTPUT_DIR)/$(PROJECT_NAME)-*

$(OUTPUT_DIR):
	@mkdir -p $(OUTPUT_DIR)

$(LINUX_AMD64): $(OUTPUT_DIR)
	@echo "Building for linux/amd64..."
	GOOS=linux GOARCH=amd64 go build -o $@ .
	chmod +x $@

$(DARWIN_AMD64): $(OUTPUT_DIR)
	@echo "Building for darwin/amd64..."
	GOOS=darwin GOARCH=amd64 go build -o $@ .
	chmod +x $@

$(DARWIN_ARM64): $(OUTPUT_DIR)
	@echo "Building for darwin/arm64..."
	GOOS=darwin GOARCH=arm64 go build -o $@ .
	chmod +x $@

$(WINDOWS_AMD64): $(OUTPUT_DIR)
	@echo "Building for windows/amd64..."
	GOOS=windows GOARCH=amd64 go build -o $@ .

clean:
	@echo "Cleaning up..."
	rm -rf $(OUTPUT_DIR)
	@echo "✓ Clean complete"
