package main

import (
	"context"
	"embed"
	"runtime"

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
		Title:            "CloudShelf",
		Width:            1200,
		Height:           800,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 15, G: 17, B: 23, A: 1},
		OnStartup:        app.startup,
		// 关闭按钮 → 弹出确认对话框；托盘菜单「退出」已置位 quitting 直接放行。
		OnBeforeClose: func(ctx context.Context) bool {
			if app.isQuitting() {
				return false // 托盘退出已确认，放行
			}
			// 主窗口关闭按钮：弹出确认对话框
			defaultButton := "取消"
			if runtime.GOOS == "windows" {
				defaultButton = "No"
			}
			result, _ := wruntime.MessageDialog(ctx, wruntime.MessageDialogOptions{
				Type:          wruntime.QuestionDialog,
				Title:         "确认退出",
				Message:       "确定要退出 CloudShelf 吗？",
				Buttons:       []string{"确定", "取消"},
				DefaultButton: defaultButton,
			})
			switch result {
			case "Yes", "Ok", "OK", "确定":
				app.markQuitting()
				return false // 放行关闭
			default:
				return true // 取消关闭
			}
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
