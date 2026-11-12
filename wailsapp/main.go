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

	err := wails.Run(&options.App{
		Title:            "S3 Client",
		Width:            1200,
		Height:           800,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 15, G: 17, B: 23, A: 1},
		OnStartup:        app.startup,
		// P5：关闭窗口二次确认，复刻旧版 SetCloseIntercept 行为。
		OnBeforeClose: func(ctx context.Context) bool {
			result, _ := wruntime.MessageDialog(app.ctx, wruntime.MessageDialogOptions{
				Type:          wruntime.QuestionDialog,
				Title:         "确认退出",
				Message:       "确定要关闭 S3 Client 吗？",
				Buttons:       []string{"确定", "取消"},
				DefaultButton: "取消",
			})
			return result != "Yes" && result != "Ok" && result != "确定"
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
