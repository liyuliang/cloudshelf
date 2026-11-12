package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	// 系统托盘：关闭窗口不退出，缩到任务栏/菜单栏托盘图标（三平台一致）。
	stopTray := startTray(app)

	err := wails.Run(&options.App{
		Title:            "S3 Client",
		Width:            1200,
		Height:           800,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 15, G: 17, B: 23, A: 1},
		OnStartup:        app.startup,
		// 关闭按钮 → 最小化到系统托盘；只有托盘菜单「退出」（确认后置位
		// quitting）才放行真正关闭。
		OnBeforeClose: func(ctx context.Context) bool {
			if app.isQuitting() {
				return false
			}
			wruntime.Hide(ctx)
			return true
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
