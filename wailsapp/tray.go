package main

import (
	_ "embed"
	"runtime"

	"fyne.io/systray"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// 托盘图标：Windows 的 Shell_NotifyIcon 只认 .ico；macOS/Linux 用 png。
//
//go:embed build/windows/icon.ico
var trayIconICO []byte

//go:embed build/appicon.png
var trayIconPNG []byte

// startTray 启动系统托盘，返回清理函数（OnShutdown 时调用，幂等）。
//
// 三平台差异：
//   - Windows / Linux：在独立 goroutine 里跑 systray 自带的事件循环。
//     Windows 必须 LockOSThread，保证托盘隐藏窗口的创建与消息泵同线程，
//     否则 goroutine 换线程后消息无人泵，托盘会失灵；
//   - macOS：NSStatusItem 只能在主线程创建，且不能另起 NSApp 事件循环
//     （会与 Wails 抢占 NSApplication），故在 wails.Run 之前用外部事件
//     循环模式在主线程启动。
func startTray(a *App) (cleanup func()) {
	// 左键单击恢复主窗口，右键弹菜单。Linux 的 StatusNotifier 在导出属性时
	// 依据回调是否为空决定 ItemIsMenu，事后注册存在竞态，故 Linux 不注册
	// 左键回调（左右键统一弹菜单，菜单内含「打开主窗口」），行为确定可控。
	if runtime.GOOS != "linux" {
		systray.SetOnTapped(a.showFromTray)
	}

	switch runtime.GOOS {
	case "darwin":
		start, _ := systray.RunWithExternalLoop(a.trayReady, nil)
		start()
	default:
		go func() {
			runtime.LockOSThread()
			systray.Run(a.trayReady, nil)
		}()
	}

	return func() {
		if runtime.GOOS == "darwin" {
			// macOS 进程退出时系统自动移除菜单栏图标；调 systray.Quit 会
			// 执行 [NSApp stop]，可能打断 Wails 收尾，故不调用。
			return
		}
		systray.Quit() // Windows：删图标并结束消息循环；Linux：关闭 DBus
	}
}

// trayReady 在托盘就绪后构建图标与菜单（可能早于 OnStartup，故不碰 ctx）。
func (a *App) trayReady() {
	if runtime.GOOS == "windows" {
		systray.SetIcon(trayIconICO)
	} else {
		systray.SetIcon(trayIconPNG)
	}
	systray.SetTitle("S3 Client")
	systray.SetTooltip("S3 Client")

	mShow := systray.AddMenuItem("打开主窗口", "显示 S3 Client 主窗口")
	mQuit := systray.AddMenuItem("退出", "退出 S3 Client")
	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				a.showFromTray()
			case <-mQuit.ClickedCh:
				a.quitFromTray()
			}
		}
	}()
}

// showFromTray 从托盘恢复主窗口：先还原最小化，再前置显示。
func (a *App) showFromTray() {
	ctx := a.runtimeCtx()
	if ctx == nil {
		return
	}
	wruntime.WindowUnminimise(ctx)
	wruntime.Show(ctx)
}

// quitFromTray 托盘菜单「退出」：二次确认后才真正退出。
// 确认对话框的返回值三平台不一致（Windows/Linux 返回 Yes/No，
// macOS 返回按钮文案），统一按确认集合判断，其余一律视为取消。
func (a *App) quitFromTray() {
	ctx := a.runtimeCtx()
	if ctx == nil {
		return
	}
	defaultButton := "取消"
	if runtime.GOOS == "windows" {
		// Windows 的 QuestionDialog 固定为 MB_YESNO，自定义按钮文案被忽略，
		// 只有 "No" 会映射成默认按钮（默认落在「否」，防误触回车退出）。
		defaultButton = "No"
	}
	result, _ := wruntime.MessageDialog(ctx, wruntime.MessageDialogOptions{
		Type:          wruntime.QuestionDialog,
		Title:         "确认退出",
		Message:       "确定要退出 S3 Client 吗？\n（关闭窗口只会最小化到托盘）",
		Buttons:       []string{"确定", "取消"},
		DefaultButton: defaultButton,
	})
	switch result {
	case "Yes", "Ok", "OK", "确定":
	default:
		return
	}
	a.markQuitting()
	wruntime.Quit(ctx)
}
