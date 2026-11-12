package main

import (
	"encoding/base64"
	"time"

	"s3client/internal/share"
)

// ---- 分享配置 DTO ----

type ShareSessionDTO struct {
	Name      string
	Provider  string // "s3" 或 "sharepoint"
	Bucket    string
	DriveID   string
	DriveName string
	ItemPath  string
	ExpiresAt string // 空串 = 不限期
	Remaining string
	Expired   bool
	ReadOnly  bool
	// 用于前端构造只读浏览标签的账号信息
	Account AccountDTO
}

func shareSessionToDTO(sess *share.Session) ShareSessionDTO {
	exp := ""
	if !sess.ExpiresAt.IsZero() {
		exp = sess.ExpiresAt.Local().Format("2006-01-02 15:04")
	}
	return ShareSessionDTO{
		Name:      sess.Name,
		Provider:  sess.Provider,
		Bucket:    sess.Bucket,
		DriveID:   sess.DriveID,
		DriveName: sess.DriveName,
		ItemPath:  sess.ItemPath,
		ExpiresAt: exp,
		Remaining: sess.Remaining(),
		Expired:   sess.Expired(),
		ReadOnly:  true,
		Account:   accountToDTO(*sess.Account),
	}
}

// OpenShareFile 打开分享配置文件并用口令解密，返回只读会话。
// password 为空时若文件是免口令则直接打开，否则返回 "need_password" 错误。
func (a *App) OpenShareFile(dataB64, password string) (ShareSessionDTO, error) {
	data, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil {
		return ShareSessionDTO{}, err
	}
	sess, err := share.Open(data, password)
	if err != nil {
		return ShareSessionDTO{}, err
	}
	return shareSessionToDTO(sess), nil
}

// NeedSharePassword 检查分享配置是否需要口令（用于前端决定是否弹口令输入框）。
func (a *App) NeedSharePassword(dataB64 string) (bool, error) {
	data, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil {
		return false, err
	}
	_, err = share.Open(data, "")
	if err == share.ErrPasswordRequired {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

// CreateS3Share 生成 S3 bucket 的只读分享配置并返回 base64 编码内容。
func (a *App) CreateS3Share(acc AccountDTO, bucket, name string, hours int, password string) (string, error) {
	ttl := time.Duration(0)
	if hours > 0 {
		ttl = time.Duration(hours) * time.Hour
	}
	data, err := share.Create(name, bucket, toAccount(acc), ttl, password)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// CreateSharePointShare 生成 SharePoint 文档库的只读分享配置。
func (a *App) CreateSharePointShare(acc AccountDTO, driveID, driveName, itemPath, name string, hours int, password string) (string, error) {
	ttl := time.Duration(0)
	if hours > 0 {
		ttl = time.Duration(hours) * time.Hour
	}
	data, err := share.CreateSharePoint(name, toAccount(acc), driveID, driveName, itemPath, ttl, password)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
