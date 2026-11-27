package main

import (
	"context"
	"time"

	"cloudshelf/internal/proxy"
	"cloudshelf/internal/updater"
	"cloudshelf/internal/version"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ReleaseInfoDTO 供前端展示的 GitHub Release 检查结果。
type ReleaseInfoDTO struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseName    string `json:"releaseName"`
	ReleaseNotes   string `json:"releaseNotes"`
	HTMLURL        string `json:"htmlUrl"`
	AssetName      string `json:"assetName"`
	AssetURL       string `json:"assetUrl"`
	AssetSize      int64  `json:"assetSize"`
	HasUpdate      bool   `json:"hasUpdate"`
	IsDev          bool   `json:"isDev"`
	CanSelfUpdate  bool   `json:"canSelfUpdate"`
}

// UpdateProgressDTO 推送更新下载进度到前端。
type UpdateProgressDTO struct {
	Done    int64 `json:"done"`
	Total   int64 `json:"total"`
	Percent int   `json:"percent"`
}

func toReleaseInfoDTO(info *updater.ReleaseInfo) ReleaseInfoDTO {
	if info == nil {
		return ReleaseInfoDTO{
			CurrentVersion: version.Version,
			IsDev:          version.IsDev(version.Version),
		}
	}
	return ReleaseInfoDTO{
		CurrentVersion: info.CurrentVersion,
		LatestVersion:  info.LatestVersion,
		ReleaseName:    info.ReleaseName,
		ReleaseNotes:   info.ReleaseNotes,
		HTMLURL:        info.HTMLURL,
		AssetName:      info.AssetName,
		AssetURL:       info.AssetURL,
		AssetSize:      info.AssetSize,
		HasUpdate:      info.HasUpdate,
		IsDev:          info.IsDev,
		CanSelfUpdate:  info.CanSelfUpdate,
	}
}

// initUpdater 在启动时清理上次更新残留的 .old 文件，并异步检查 GitHub Release 新版本。
func (a *App) initUpdater() {
	updater.CleanupOldBinary()
	go func() {
		// 等待前端挂载完成，避免早期事件丢失
		time.Sleep(2500 * time.Millisecond)
		ctx := a.runtimeCtx()
		if ctx == nil || a.isQuitting() {
			return
		}
		checkCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		info, err := updater.CheckLatest(checkCtx, proxy.Client(), "")
		if err != nil || info == nil || !info.HasUpdate {
			return
		}
		wruntime.EventsEmit(ctx, "update:available", toReleaseInfoDTO(info))
	}()
}

// GetAppVersion 返回当前构建版本号。
func (a *App) GetAppVersion() string {
	return version.Version
}

// CheckForUpdate 主动检查 GitHub Release 最新版本。
func (a *App) CheckForUpdate() (ReleaseInfoDTO, error) {
	ctx := a.runtimeCtx()
	if ctx == nil {
		ctx = context.Background()
	}
	checkCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	info, err := updater.CheckLatest(checkCtx, proxy.Client(), "")
	if err != nil {
		return ReleaseInfoDTO{}, err
	}
	return toReleaseInfoDTO(info), nil
}

// ApplyUpdate 下载对应平台的 Release 资产并替换当前可执行文件后重启。
func (a *App) ApplyUpdate(assetURL, assetName string) error {
	ctx := a.runtimeCtx()
	if ctx == nil {
		ctx = context.Background()
	}
	dlCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	err := updater.DownloadAndApply(dlCtx, proxy.Client(), assetURL, assetName, func(done, total int64) {
		pct := 0
		if total > 0 {
			pct = int(done * 100 / total)
			if pct > 100 {
				pct = 100
			}
		}
		if rctx := a.runtimeCtx(); rctx != nil {
			wruntime.EventsEmit(rctx, "update:progress", UpdateProgressDTO{
				Done:    done,
				Total:   total,
				Percent: pct,
			})
		}
	})
	if err != nil {
		return err
	}

	a.markQuitting()
	if rctx := a.runtimeCtx(); rctx != nil {
		wruntime.Quit(rctx)
	}
	return nil
}

// OpenReleasePage 在系统默认浏览器中打开 GitHub Release 页面。
func (a *App) OpenReleasePage(url string) {
	if url == "" {
		url = "https://github.com/" + updater.RepoOwner + "/" + updater.RepoName + "/releases"
	}
	if ctx := a.runtimeCtx(); ctx != nil {
		wruntime.BrowserOpenURL(ctx, url)
	}
}
