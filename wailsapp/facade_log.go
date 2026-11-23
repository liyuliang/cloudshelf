package main

import (
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"cloudshelf/internal/reqlog"
)

// initLogForwarder 把 reqlog 的日志条目通过 Wails EventsEmit 推送到前端。
// 在 App.startup 中调用一次即可。
func (a *App) initLogForwarder() {
	reqlog.Subscribe(func(e reqlog.Entry) {
		ctx := a.runtimeCtx()
		if ctx == nil {
			return
		}
		wruntime.EventsEmit(ctx, "request:log", e)
	})
}
