# 基础变量
BINARY_NAME=clinvarDL
MAIN_PATH=cmd/main.go
BUILD_DIR=bin
CGO_ENABLED=0

# 版本信息
VERSION=1.0.0
BUILD_TIME=$(shell date "+%F %T")
COMMIT_SHA1=$(shell git rev-parse HEAD)

# LDFLAGS
LDFLAGS=-s -w
GCFLAGS=-N -l

.PHONY: all windows linux darwin clean help

# 默认目标
all: windows linux darwin

# Windows 64位
windows:
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-amd64-windows.exe \
		-gcflags "$(GCFLAGS)" \
		-ldflags "$(LDFLAGS)" \
		$(MAIN_PATH)

# Linux 64位
linux:
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-amd64-linux \
		-gcflags "$(GCFLAGS)" \
		-ldflags "$(LDFLAGS)" \
		$(MAIN_PATH)

# macOS 64位
darwin:
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-amd64-darwin \
		-gcflags "$(GCFLAGS)" \
		-ldflags "$(LDFLAGS)" \
		$(MAIN_PATH)

# 清理构建产物
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  all     : Build for all platforms (windows, linux, darwin)"
	@echo "  windows : Build for Windows (amd64)"
	@echo "  linux   : Build for Linux (amd64)"
	@echo "  darwin  : Build for macOS (amd64)"
	@echo "  clean   : Remove all built binaries" 