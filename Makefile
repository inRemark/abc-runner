# redis-runner Makefile

# 变量定义
BINARY_NAME=redis-runner
OUTPUT_DIR=bin
SOURCE_DIR=.
CONFIG_DIR=conf
DOCS_DIR=docs

# Go相关变量
GO=go
GO_BUILD=$(GO) build
GO_TEST=$(GO) test
GO_CLEAN=$(GO) clean
GO_DEPS=$(GO) mod tidy

# 版本信息
VERSION ?= 0.2.0
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD)

# 构建标志
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE) -X main.gitCommit=$(GIT_COMMIT)"

# 平台列表
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

# 默认目标
.PHONY: all
all: clean deps build

# 构建相关
.PHONY: build
build: clean
	@echo "Building $(BINARY_NAME)..."
	$(GO_BUILD) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) $(SOURCE_DIR)/main.go
	@echo "Build completed successfully!"

.PHONY: build-all
build-all: clean deps
	@echo "Building for all platforms..."
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		OUTPUT_NAME="$(OUTPUT_DIR)/$(BINARY_NAME)-$$OS-$$ARCH"; \
		if [ "$$OS" = "windows" ]; then OUTPUT_NAME="$(OUTPUT_DIR)/$(BINARY_NAME)-$$OS-$$ARCH.exe"; fi; \
		echo "Building for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH $(GO_BUILD) $(LDFLAGS) -o $$OUTPUT_NAME $(SOURCE_DIR)/main.go; \
		echo "Built $$OUTPUT_NAME"; \
	done
	@echo "Cross-platform build completed!"

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux $(SOURCE_DIR)/main.go
	@echo "Linux build completed!"

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin $(SOURCE_DIR)/main.go
	@echo "macOS build completed!"

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows.exe $(SOURCE_DIR)/main.go
	@echo "Windows build completed!"

# 测试相关
.PHONY: test
test:
	@echo "Running tests..."
	$(GO_TEST) -v ./...

.PHONY: test-cover
test-cover:
	@echo "Running tests with coverage..."
	$(GO_TEST) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: integration-test
integration-test:
	@echo "Running integration tests..."
	$(GO_TEST) -v ./test/integration/...

# 清理
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	$(GO_CLEAN)
	rm -rf $(OUTPUT_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean completed!"

# 依赖管理
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GO_DEPS)
	@echo "Dependencies installed!"

.PHONY: vendor
vendor:
	@echo "Vendoring dependencies..."
	$(GO) mod vendor
	@echo "Dependencies vendored!"

# 安装
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(OUTPUT_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation completed!"

# 文档
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@mkdir -p $(DOCS_DIR)/generated
	@echo "Documentation generation completed!"

# 配置文件
.PHONY: config
config:
	@echo "Copying configuration files..."
	@mkdir -p $(CONFIG_DIR)
	@cp -n config/templates/*.yaml $(CONFIG_DIR)/ 2>/dev/null || true
	@echo "Configuration files copied!"

# 发布
.PHONY: release
release: clean deps build-all
	@echo "Creating release v$(VERSION)..."
	@mkdir -p releases/v$(VERSION)
	@cp $(OUTPUT_DIR)/$(BINARY_NAME)-* releases/v$(VERSION)/
	@cp -r $(CONFIG_DIR) releases/v$(VERSION)/
	@cp -r $(DOCS_DIR) releases/v$(VERSION)/
	@cp README.md releases/v$(VERSION)/
	@cp LICENSE releases/v$(VERSION)/
	@echo "Release v$(VERSION) created in releases/v$(VERSION)/"

# 帮助
.PHONY: help
help:
	@echo "redis-runner Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build the project"
	@echo "  make build        Build the project"
	@echo "  make build-all    Build for all platforms"
	@echo "  make test         Run tests"
	@echo "  make clean        Clean build artifacts"
	@echo "  make deps         Install dependencies"
	@echo "  make install      Install the binary"
	@echo "  make release      Create a release"
	@echo ""
	@echo "Targets:"
	@echo "  all             - Clean, install dependencies, and build"
	@echo "  build           - Build the project"
	@echo "  build-all       - Build for all supported platforms"
	@echo "  build-linux     - Build for Linux"
	@echo "  build-darwin    - Build for macOS"
	@echo "  build-windows   - Build for Windows"
	@echo "  test            - Run unit tests"
	@echo "  test-cover      - Run tests with coverage"
	@echo "  integration-test - Run integration tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install dependencies"
	@echo "  vendor          - Vendor dependencies"
	@echo "  install         - Install the binary"
	@echo "  docs            - Generate documentation"
	@echo "  config          - Copy configuration templates"
	@echo "  release         - Create a release"
	@echo "  help            - Show this help message"