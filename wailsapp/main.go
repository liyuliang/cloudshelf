package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	// 系统托盘：关闭窗口不退出，缩到任务栏/菜单栏托盘图标（三平台一致）。
	stopTray := startTray(app)

	err := wails.Run(&options.App{
		Title:            "CloudShelf",
		Width:            1200,
		Height:           800,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 15, G: 17, B: 23, A: 1},
		OnStartup:        app.startup,
		// 关闭按钮 → 隐藏到系统托盘并重置应用状态；托盘菜单「退出」已置位 quitting 直接放行。
		OnBeforeClose: func(ctx context.Context) bool {
			if app.isQuitting() {
				return false // 托盘退出已确认，放行
			}
			app.hideToTrayAndReset(ctx)
			return true // 取消关闭，保持驻留托盘
		},
		OnShutdown: func(ctx context.Context) {
			stopTray()
		},
		Bind: []interface{}{
			app,
		},
	})
	stopTray() // 兜底清理（幂等，OnShutdown 已调用过则为空操作）

	if err != nil {
		println("Error:", err.Error())
	}
}
