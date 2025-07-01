.PHONY: build install clean test release snapshot help

# バージョン情報
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# ビルド設定
BINARY_NAME := mcpjson
MAIN_PACKAGE := .
LDFLAGS := -s -w \
	-X 'github.com/naoto24kawa/mcpjson/cmd.Version=$(VERSION)' \
	-X 'github.com/naoto24kawa/mcpjson/cmd.Commit=$(COMMIT)' \
	-X 'github.com/naoto24kawa/mcpjson/cmd.BuildTime=$(BUILD_TIME)'

# Go設定
GO := go
GOFLAGS := -trimpath

# デフォルトターゲット
all: build

## help: ヘルプを表示
help:
	@echo "使用可能なターゲット:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: バイナリをビルド
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: ./$(BINARY_NAME)"

## install: バイナリをシステムにインストール
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@if [ -w /usr/local/bin ]; then \
		cp $(BINARY_NAME) /usr/local/bin/; \
	else \
		sudo cp $(BINARY_NAME) /usr/local/bin/; \
	fi
	@echo "Installation complete!"

## uninstall: バイナリをアンインストール
uninstall:
	@echo "Removing $(BINARY_NAME) from /usr/local/bin..."
	@if [ -w /usr/local/bin ]; then \
		rm -f /usr/local/bin/$(BINARY_NAME); \
	else \
		sudo rm -f /usr/local/bin/$(BINARY_NAME); \
	fi
	@echo "Uninstallation complete!"

## clean: ビルド成果物をクリーン
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf dist/
	@echo "Clean complete!"

## test: テストを実行
test:
	@echo "Running tests..."
	$(GO) test -v ./...

## test-coverage: カバレッジ付きでテストを実行
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## fmt: コードをフォーマット
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

## vet: コードの静的解析
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

## lint: より詳細な静的解析（golangci-lintが必要）
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Install it from https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

## deps: 依存関係を更新
deps:
	@echo "Updating dependencies..."
	$(GO) mod download
	$(GO) mod tidy

## build-all: 全プラットフォーム向けにビルド
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	# Linux
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	# macOS
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	# Windows
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "Build complete for all platforms!"

## release: GoReleaserを使用してリリース
release:
	@if command -v goreleaser >/dev/null 2>&1; then \
		echo "Creating release with GoReleaser..."; \
		goreleaser release --clean; \
	else \
		echo "goreleaser is not installed. Install it from https://goreleaser.com/install/"; \
		exit 1; \
	fi

## snapshot: GoReleaserでスナップショットリリース（タグなし）
snapshot:
	@if command -v goreleaser >/dev/null 2>&1; then \
		echo "Creating snapshot release with GoReleaser..."; \
		goreleaser release --snapshot --clean; \
	else \
		echo "goreleaser is not installed. Install it from https://goreleaser.com/install/"; \
		exit 1; \
	fi

## version: バージョン情報を表示
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"