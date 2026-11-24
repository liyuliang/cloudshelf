package main

import (
	"context"
	"path/filepath"
	"time"

	"cloudshelf/internal/sharepoint"
)

// ---- SharePoint DTO ----

type DriveDTO struct {
	ID   string
	Name string
}

type SPItemDTO struct {
	ID             string
	Name           string
	IsFolder       bool
	Size           int64
	CreatedTime    string
	LastModified   string
	CreatedBy      string
	LastModifiedBy string
	DownloadURL    string
}

type SPTransferItemDTO struct {
	ID       string // 前端生成的唯一任务 id
	Name     string // 显示名
	ItemID   string // Graph driveItem ID（下载用）
	LocalPath string // 本地完整路径（上传用）
}

func spItemToDTO(it sharepoint.Item) SPItemDTO {
	ct := ""
	if !it.CreatedTime.IsZero() {
		ct = it.CreatedTime.Local().Format("2006-01-02 15:04")
	}
	lm := ""
	if !it.LastModified.IsZero() {
		lm = it.LastModified.Local().Format("2006-01-02 15:04")
	}
	return SPItemDTO{
		ID:             it.ID,
		Name:           it.Name,
		IsFolder:       it.IsFolder,
		Size:           it.Size,
		CreatedTime:    ct,
		LastModified:   lm,
		CreatedBy:      it.CreatedBy,
		LastModifiedBy: it.LastModifiedBy,
		DownloadURL:    it.DownloadURL,
	}
}

// ---- SharePoint 文档库列表 ----

func (a *App) SPListDrives(acc AccountDTO) ([]DriveDTO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	drives, err := sharepoint.ListDrives(ctx, spConfig(toAccount(acc)))
	if err != nil {
		return nil, err
	}
	out := make([]DriveDTO, 0, len(drives))
	for _, d := range drives {
		out = append(out, DriveDTO{ID: d.ID, Name: d.Name})
	}
	return out, nil
}

// ---- SharePoint 子项列表 ----

func (a *App) SPListChildren(acc AccountDTO, driveID, itemPath string) ([]SPItemDTO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	items, err := sharepoint.ListChildren(ctx, spConfig(toAccount(acc)), driveID, itemPath)
	if err != nil {
		return nil, err
	}
	out := make([]SPItemDTO, 0, len(items))
	for _, it := range items {
		out = append(out, spItemToDTO(it))
	}
	return out, nil
}

// ---- SharePoint 写操作 ----

func (a *App) SPCreateFolder(acc AccountDTO, driveID, parentPath, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return sharepoint.CreateFolder(ctx, spConfig(toAccount(acc)), driveID, parentPath, name)
}

func (a *App) SPDeleteItem(acc AccountDTO, driveID, itemID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return sharepoint.DeleteItem(ctx, spConfig(toAccount(acc)), driveID, itemID)
}

// ---- SharePoint 四路并发传输 ----

func (a *App) SPUploadFiles(acc AccountDTO, driveID, itemPath string, items []SPTransferItemDTO) (TransferResultDTO, error) {
	cfg := spConfig(toAccount(acc))
	transferItems := make([]TransferItemDTO, len(items))
	for i, it := range items {
		transferItems[i] = TransferItemDTO{ID: it.ID, Name: it.Name, LocalPath: it.LocalPath}
	}
	return a.runTransfers(transferItems, func(it TransferItemDTO, progress func(int64, int64)) error {
		return sharepoint.UploadItem(context.Background(), cfg, driveID, itemPath, filepath.FromSlash(it.LocalPath), progress)
	}), nil
}

func (a *App) SPDownloadFiles(acc AccountDTO, driveID, dir string, items []SPTransferItemDTO) (TransferResultDTO, error) {
	cfg := spConfig(toAccount(acc))
	transferItems := make([]TransferItemDTO, len(items))
	for i, it := range items {
		transferItems[i] = TransferItemDTO{ID: it.ID, Name: it.Name, Key: it.ItemID}
	}
	return a.runTransfers(transferItems, func(it TransferItemDTO, progress func(int64, int64)) error {
		local := filepath.Join(filepath.FromSlash(dir), it.Name)
		return sharepoint.DownloadItem(context.Background(), cfg, driveID, it.Key, local, progress)
	}), nil
}
