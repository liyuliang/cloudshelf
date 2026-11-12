package main

import (
	"context"
	"sync"

	"s3client/internal/store"
)

// App 是 Wails 的 Go 绑定层（facade）。
// 只做参数适配与状态管理，所有业务逻辑原样复用 internal/* 下的 7 个纯 Go 包，
// 不在这一层重写任何加解密、序列化或存储逻辑，以保证老数据逐字节兼容。
type App struct {
	ctx   context.Context
	mu    sync.RWMutex
	store *store.Store
}

func NewApp() *App {
	return &App{}
}

// startup 保存 Wails runtime 上下文，供对话框、事件推送等 runtime 调用使用。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

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
