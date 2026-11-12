APP     := s3client
CLI     := s3cli
CLI_PKG := ./cmd/cli
GOOS    := $(shell go env GOOS)

.PHONY: build build-cli dev run clean

## build: 构建 Wails v2 桌面应用（产出 wailsapp/build/bin/s3client.exe）
build:
	cd wailsapp && wails build

## build-cli: 构建纯 Go CLI（跨平台，无需 CGO）
build-cli:
	CGO_ENABLED=0 go build -o $(CLI) $(CLI_PKG)
	@echo "Built $(CLI)"

## dev: 启动 Wails v2 开发模式（Go + Vue 热重载）
dev:
	cd wailsapp && wails dev

## run: 直接运行已构建的桌面应用（Windows 为 .exe）
run:
ifeq ($(GOOS),windows)
	wailsapp/build/bin/$(APP).exe
else
	wailsapp/build/bin/$(APP)
endif

## clean: 清理构建产物
clean:
	rm -rf wailsapp/build/bin/
	rm -f $(CLI) $(CLI).exe
