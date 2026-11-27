package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cloudshelf/internal/version"
)

const (
	RepoOwner = "liyuliang"
	RepoName  = "cloudshelf"
)

// ReleaseInfo 描述 GitHub Release 最新版本检查结果。
type ReleaseInfo struct {
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

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Name    string    `json:"name"`
	Body    string    `json:"body"`
	HTMLURL string    `json:"html_url"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ExpectedAssetName 返回当前平台对应的桌面端 Release 资产文件名。
func ExpectedAssetName(goos, goarch string) string {
	switch goos {
	case "windows":
		if goarch == "amd64" {
			return "cloudshelf-windows-amd64-installer.exe"
		}
	case "darwin":
		return "cloudshelf-macos-universal.zip"
	case "linux":
		if goarch == "amd64" {
			return "cloudshelf-linux-amd64.tar.gz"
		}
	}
	return ""
}

// CheckLatest 查询 GitHub Releases API 获取最新正式版本并与当前版本比较。
func CheckLatest(ctx context.Context, client *http.Client, apiBase string) (*ReleaseInfo, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", strings.TrimRight(apiBase, "/"), RepoOwner, RepoName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "cloudshelf/"+version.Version)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	info := &ReleaseInfo{
		CurrentVersion: version.Version,
		IsDev:          version.IsDev(version.Version),
		HTMLURL:        fmt.Sprintf("https://github.com/%s/%s/releases", RepoOwner, RepoName),
	}

	if resp.StatusCode == http.StatusNotFound {
		// 仓库尚未发布任何正式 Release（仅 prerelease 或空仓库）
		return info, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("GitHub API 返回 HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("解析 GitHub Release 响应失败: %w", err)
	}

	info.LatestVersion = rel.TagName
	info.ReleaseName = rel.Name
	if info.ReleaseName == "" {
		info.ReleaseName = rel.TagName
	}
	info.ReleaseNotes = rel.Body
	if rel.HTMLURL != "" {
		info.HTMLURL = rel.HTMLURL
	}
	info.HasUpdate = version.IsNewer(version.Version, rel.TagName)

	wantAsset := ExpectedAssetName(runtime.GOOS, runtime.GOARCH)
	for _, a := range rel.Assets {
		if wantAsset != "" && a.Name == wantAsset {
			info.AssetName = a.Name
			info.AssetURL = a.BrowserDownloadURL
			info.AssetSize = a.Size
			info.CanSelfUpdate = true
			break
		}
	}
	return info, nil
}

// CleanupOldBinary 清理上次自更新遗留的 .old 备份文件。
func CleanupOldBinary() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	_ = os.Remove(exe + ".old")
}

// DownloadAndApply 下载对应平台资产并替换当前可执行文件，然后启动新进程。
// 调用方在返回 nil 后应立即退出当前进程。
func DownloadAndApply(ctx context.Context, client *http.Client, assetURL, assetName string, onProgress func(done, total int64)) error {
	if assetURL == "" {
		return errors.New("缺少更新包下载地址")
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Minute}
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("定位当前程序路径失败: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = resolved
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "cloudshelf/"+version.Version)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("下载更新包失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载更新包返回 HTTP %d", resp.StatusCode)
	}

	lowerName := strings.ToLower(assetName)
	isWindowsInstaller := runtime.GOOS == "windows" &&
		(strings.HasSuffix(lowerName, "-installer.exe") || strings.HasSuffix(lowerName, "-setup.exe"))

	if isWindowsInstaller {
		installerPath := filepath.Join(os.TempDir(), filepath.Base(assetName))
		if err := downloadToFile(resp.Body, installerPath, resp.ContentLength, onProgress); err != nil {
			_ = os.Remove(installerPath)
			return err
		}
		return launchWindowsInstaller(installerPath)
	}

	dir := filepath.Dir(exePath)
	tmpDownload := filepath.Join(dir, ".cloudshelf-update.tmp")
	defer os.Remove(tmpDownload)

	if err := downloadToFile(resp.Body, tmpDownload, resp.ContentLength, onProgress); err != nil {
		return err
	}

	switch {
	case strings.HasSuffix(lowerName, ".exe"):
		return applyBinaryReplace(exePath, tmpDownload)
	case strings.HasSuffix(lowerName, ".tar.gz"):
		extracted := filepath.Join(dir, ".cloudshelf-update.bin")
		defer os.Remove(extracted)
		if err := extractBinaryFromTarGz(tmpDownload, "cloudshelf", extracted); err != nil {
			return err
		}
		return applyBinaryReplace(exePath, extracted)
	case strings.HasSuffix(lowerName, ".zip") && runtime.GOOS == "darwin":
		return applyMacOSAppZip(exePath, tmpDownload)
	default:
		return fmt.Errorf("不支持的更新包格式: %s", assetName)
	}
}

func launchWindowsInstaller(installerPath string) error {
	// 通过 cmd start 触发 ShellExecute，以便在需要时正确弹出 UAC 提权提示
	cmd := exec.Command("cmd", "/C", "start", "", installerPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动安装程序失败: %w", err)
	}
	return nil
}

func downloadToFile(r io.Reader, dst string, total int64, onProgress func(done, total int64)) error {
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer f.Close()

	buf := make([]byte, 32*1024)
	var done int64
	var lastEmit time.Time
	for {
		n, readErr := r.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return fmt.Errorf("写入更新包失败: %w", werr)
			}
			done += int64(n)
			if onProgress != nil && (time.Since(lastEmit) >= 120*time.Millisecond || readErr == io.EOF) {
				onProgress(done, total)
				lastEmit = time.Now()
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("读取更新包失败: %w", readErr)
		}
	}
	return f.Close()
}

func applyBinaryReplace(exePath, newBinPath string) error {
	if runtime.GOOS != "windows" {
		if err := os.Chmod(newBinPath, 0o755); err != nil {
			return err
		}
	}
	oldPath := exePath + ".old"
	_ = os.Remove(oldPath)
	if err := os.Rename(exePath, oldPath); err != nil {
		return fmt.Errorf("备份当前程序失败: %w", err)
	}
	if err := os.Rename(newBinPath, exePath); err != nil {
		_ = os.Rename(oldPath, exePath)
		return fmt.Errorf("替换新版本失败: %w", err)
	}

	cmd := exec.Command(exePath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动新版本失败: %w", err)
	}
	return nil
}

func extractBinaryFromTarGz(tarGzPath, targetBase, dstPath string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("解压 tar.gz 失败: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeReg && filepath.Base(hdr.Name) == targetBase {
			out, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			return out.Close()
		}
	}
	return fmt.Errorf("更新包中未找到 %s", targetBase)
}

func applyMacOSAppZip(exePath, zipPath string) error {
	// exePath 通常为 .../cloudshelf.app/Contents/MacOS/cloudshelf
	idx := strings.Index(exePath, ".app/")
	if idx < 0 {
		return errors.New("当前并非从 .app 包运行，请前往 GitHub Release 手动下载")
	}
	appBundle := exePath[:idx+4]
	tmpDir, err := os.MkdirTemp("", "cloudshelf-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// 使用系统 unzip 保留 macOS 权限与符号链接
	if out, err := exec.Command("unzip", "-q", zipPath, "-d", tmpDir).CombinedOutput(); err != nil {
		// 回退到 archive/zip
		if zerr := unzipSimple(zipPath, tmpDir); zerr != nil {
			return fmt.Errorf("解压 macOS 更新包失败: %v (%s)", err, strings.TrimSpace(string(out)))
		}
	}

	newApp := filepath.Join(tmpDir, "cloudshelf.app")
	if _, err := os.Stat(newApp); err != nil {
		return errors.New("更新包中未找到 cloudshelf.app")
	}

	backupApp := appBundle + ".old"
	_ = os.RemoveAll(backupApp)
	if err := os.Rename(appBundle, backupApp); err != nil {
		return fmt.Errorf("备份旧版 .app 失败: %w", err)
	}
	if err := os.Rename(newApp, appBundle); err != nil {
		_ = os.Rename(backupApp, appBundle)
		return fmt.Errorf("替换新版 .app 失败: %w", err)
	}
	_ = os.RemoveAll(backupApp)

	return exec.Command("open", "-n", appBundle).Start()
}

func unzipSimple(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		target := filepath.Join(dst, f.Name)
		if !strings.HasPrefix(target, filepath.Clean(dst)+string(os.PathSeparator)) {
			continue
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		rc.Close()
		out.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}
