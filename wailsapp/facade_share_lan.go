package main

import (
	"fmt"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"cloudshelf/internal/lanshare"
)

// ---- LAN 分享 DTO ----

type LANShareDTO struct {
	Token     string   `json:"token"`
	Name      string   `json:"name"`
	Bucket    string   `json:"bucket"`
	Prefix    string   `json:"prefix"`
	URL       string   `json:"url"`
	URLs      []string `json:"urls"`
	CreatedAt string   `json:"createdAt"`
	ExpiresAt string   `json:"expiresAt"`
	Remaining string   `json:"remaining"`
}

func shareEntryToDTO(e *lanshare.ShareEntry) LANShareDTO {
	remaining := time.Until(e.ExpiresAt)
	remStr := "已过期"
	if remaining > 0 {
		h := int(remaining.Hours())
		m := int(remaining.Minutes()) % 60
		if h > 0 {
			remStr = fmt.Sprintf("%d 小时 %d 分钟", h, m)
		} else {
			remStr = fmt.Sprintf("%d 分钟", m)
		}
	}
	return LANShareDTO{
		Token:     e.Token,
		Name:      e.Name,
		Bucket:    e.Bucket,
		Prefix:    e.Prefix,
		URL:       e.URL,
		URLs:      e.URLs,
		CreatedAt: e.CreatedAt.Local().Format("2006-01-02 15:04"),
		ExpiresAt: e.ExpiresAt.Local().Format("2006-01-02 15:04"),
		Remaining: remStr,
	}
}

// ---- Wails 绑定方法 ----

// lanServer 在 App 层持有，启动一次后常驻到应用退出。
var lanServer *lanshare.Server

func ensureLANServer() (*lanshare.Server, error) {
	if lanServer != nil {
		return lanServer, nil
	}
	srv := lanshare.New()
	_, err := srv.Start()
	if err != nil {
		return nil, fmt.Errorf("启动分享服务失败: %w", err)
	}
	lanServer = srv
	return srv, nil
}

// CreateLANShare 在局域网上创建一个目录级只读分享，返回分享信息（含 URL）。
func (a *App) CreateLANShare(acc AccountDTO, bucket, prefix, name string, hours int) (LANShareDTO, error) {
	srv, err := ensureLANServer()
	if err != nil {
		return LANShareDTO{}, err
	}
	ma := toAccount(acc)
	// Make a deep copy so the share holds its own credentials
	acctCopy := *ma
	if hours <= 0 {
		hours = 4
	}
	entry, err := srv.CreateShare(&acctCopy, bucket, prefix, name, time.Duration(hours)*time.Hour)
	if err != nil {
		return LANShareDTO{}, err
	}
	return shareEntryToDTO(entry), nil
}

// ListLANShares 返回所有活跃的局域网分享。
func (a *App) ListLANShares() []LANShareDTO {
	if lanServer == nil {
		return nil
	}
	entries := lanServer.ListShares()
	out := make([]LANShareDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, shareEntryToDTO(e))
	}
	return out
}

// RemoveLANShare 删除指定 token 的分享。
func (a *App) RemoveLANShare(token string) {
	if lanServer != nil {
		lanServer.RemoveShare(token)
	}
}

// GetLANAddresses 返回本机检测到的局域网 IP 列表（供前端展示）。
func (a *App) GetLANAddresses() []string {
	return lanshare.LANAddresses()
}

// GetLANSharePort 返回分享服务监听的端口号（未启动时返回 0）。
func (a *App) GetLANSharePort() int {
	if lanServer == nil {
		return 0
	}
	return lanServer.Port()
}

// CopyToClipboard 复制文本到系统剪贴板。
func (a *App) CopyToClipboard(text string) {
	ctx := a.runtimeCtx()
	if ctx != nil {
		wruntime.ClipboardSetText(ctx, text)
	}
}
