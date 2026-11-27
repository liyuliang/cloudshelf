package main

import (
	"context"
	"sync"
	"sync/atomic"

	"cloudshelf/internal/store"
)

// App 是 Wails 的 Go 绑定层（facade）。
// 只做参数适配与状态管理，所有业务逻辑原样复用 internal/* 下的 7 个纯 Go 包，
// 不在这一层重写任何加解密、序列化或存储逻辑，以保证老数据逐字节兼容。
type App struct {
	ctx   context.Context
	mu    sync.RWMutex
	store *store.Store
	// quitting 标记正在走托盘「退出」流程：置位后 OnBeforeClose 放行真正关闭，
	// 否则关闭按钮一律改为隐藏到托盘。
	quitting atomic.Bool
}

func NewApp() *App {
	return &App{}
}

// startup 保存 Wails runtime 上下文，供对话框、事件推送等 runtime 调用使用。
// 托盘回调可能先于它执行，故用 mu 保护，读侧经 runtimeCtx 获取。
func (a *App) startup(ctx context.Context) {
	a.mu.Lock()
	a.ctx = ctx
	a.mu.Unlock()

	a.initLogForwarder()
	a.initUpdater()
}

// runtimeCtx 返回 runtime 上下文；应用尚未启动完成时返回 nil。
func (a *App) runtimeCtx() context.Context {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ctx
}

// markQuitting / isQuitting 见 App.quitting 注释。
func (a *App) markQuitting()    { a.quitting.Store(true) }
func (a *App) isQuitting() bool { return a.quitting.Load() }

// currentStore 返回已解锁的账号库；未解锁时返回 nil。
func (a *App) currentStore() *store.Store {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.store
}

// setStore 替换当前账号库（解锁、加锁、改库位置时调用）。
func (a *App) setStore(s *store.Store) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.store = s
}
