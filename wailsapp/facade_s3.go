package main

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"cloudshelf/internal/awss3"
	"cloudshelf/internal/model"
)

// ---- 传输 DTO ----

type BucketDTO struct {
	Name       string
	Accessible bool
}

// ObjectDTO 把 awss3.Object 的 time.Time 转成展示字符串，避免绑定生成的不确定性。
type ObjectDTO struct {
	Key          string
	Name         string
	IsDir        bool
	Size         int64
	Type         string
	LastModified string
}

// ObjectListDTO 把「条目 + 是否被 S3 单页上限截断」一起返回。
// Wails v2 绑定会丢弃第三个返回值之外的多返回值，因此不能直接用 ([]ObjectDTO, bool, error)。
type ObjectListDTO struct {
	Items     []ObjectDTO
	Truncated bool
}

func objectToDTO(o awss3.Object) ObjectDTO {
	lm := ""
	if !o.LastModified.IsZero() {
		lm = o.LastModified.Local().Format("2006-01-02 15:04")
	}
	return ObjectDTO{
		Key:          o.Key,
		Name:         o.Name,
		IsDir:        o.IsDir,
		Size:         o.Size,
		Type:         o.Type,
		LastModified: lm,
	}
}

func objectsToDTO(list []awss3.Object) []ObjectDTO {
	out := make([]ObjectDTO, 0, len(list))
	for _, o := range list {
		out = append(out, objectToDTO(o))
	}
	return out
}

func toAccount(d AccountDTO) *model.Account {
	m := d.toModel()
	return &m
}

// ---- Bucket 列表 ----

func (a *App) ListBuckets(acc AccountDTO) ([]BucketDTO, error) {
	infos, err := awss3.ListBucketsWithAccess(toAccount(acc))
	if err != nil {
		return nil, err
	}
	out := make([]BucketDTO, 0, len(infos))
	for _, b := range infos {
		out = append(out, BucketDTO{Name: b.Name, Accessible: b.Accessible})
	}
	return out, nil
}

// ---- 单层对象列表 / 服务端前缀搜索 ----

func (a *App) ListObjects(acc AccountDTO, bucket, prefix string) (ObjectListDTO, error) {
	list, truncated, err := awss3.ListObjects(toAccount(acc), bucket, prefix)
	if err != nil {
		return ObjectListDTO{}, err
	}
	return ObjectListDTO{Items: objectsToDTO(list), Truncated: truncated}, nil
}

func (a *App) SearchObjects(acc AccountDTO, bucket, prefix, keyword string) (ObjectListDTO, error) {
	list, truncated, err := awss3.SearchObjects(toAccount(acc), bucket, prefix, keyword)
	if err != nil {
		return ObjectListDTO{}, err
	}
	return ObjectListDTO{Items: objectsToDTO(list), Truncated: truncated}, nil
}

// ListObjectsRecursive 供文件夹下载使用（P3 并行传输会用到）。
func (a *App) ListObjectsRecursive(acc AccountDTO, bucket, prefix string) ([]ObjectDTO, error) {
	list, err := awss3.ListObjectsRecursive(toAccount(acc), bucket, prefix)
	if err != nil {
		return nil, err
	}
	return objectsToDTO(list), nil
}

// ResolvePathDir 按目录段逐级下钻，停在最后一个存在的目录。
func (a *App) ResolvePathDir(acc AccountDTO, bucket string, dirs []string) (string, error) {
	return awss3.ResolvePathDir(toAccount(acc), bucket, dirs)
}

// ---- 记住每个 bucket 上次打开的目录（meta 表，按账号 + bucket 隔离）----

func (a *App) GetBucketPrefix(acc AccountDTO, bucket string) (string, error) {
	s := a.currentStore()
	if s == nil {
		return "", nil
	}
	return s.GetBucketPrefix(acc.ID, bucket)
}

func (a *App) SetBucketPrefix(acc AccountDTO, bucket, prefix string) error {
	s := a.currentStore()
	if s == nil {
		return nil
	}
	return s.SetBucketPrefix(acc.ID, bucket, prefix)
}

// ---- 写操作 / 预签名 URL ----

func (a *App) PresignURL(acc AccountDTO, bucket, key string, minutes int) (string, error) {
	if minutes <= 0 {
		minutes = 10
	}
	return awss3.PresignURL(toAccount(acc), bucket, key, time.Duration(minutes)*time.Minute)
}

func (a *App) DeleteObject(acc AccountDTO, bucket, key string) error {
	return awss3.DeleteObject(toAccount(acc), bucket, key)
}

func (a *App) CreateFolder(acc AccountDTO, bucket, prefix, name string) error {
	return awss3.CreateFolder(toAccount(acc), bucket, prefix, name)
}

// ---- 文件对话框（多选/目录）与打开资源管理器 ----

func (a *App) PickDirectory(title string) (string, error) {
	return wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{Title: title})
}

func (a *App) PickMultipleFiles(title string) ([]string, error) {
	return wruntime.OpenMultipleFilesDialog(a.ctx, wruntime.OpenDialogOptions{Title: title})
}

// RevealInExplorer 复刻 internal/ui/ui.go:66 的平台分支。
func (a *App) RevealInExplorer(path string) error {
	path = filepath.FromSlash(path)
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", "/select,"+path).Start()
	case "darwin":
		return exec.Command("open", "-R", path).Start()
	default:
		return exec.Command("xdg-open", filepath.Dir(path)).Start()
	}
}

// ---- 单文件传输（P2 简化版：无进度条、不并行；P3 升级为多选并行 + 进度事件）----

func (a *App) UploadFile(acc AccountDTO, bucket, key, localPath string) error {
	return awss3.UploadFile(toAccount(acc), bucket, key, filepath.FromSlash(localPath), nil)
}

func (a *App) DownloadFile(acc AccountDTO, bucket, key, localPath string) error {
	return awss3.DownloadFile(toAccount(acc), bucket, key, filepath.FromSlash(localPath), nil)
}
