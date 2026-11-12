package main

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"s3client/internal/store"
)

// ---- 会话门（对应旧版 showUnlock / 加锁 / 改主密码 / 改库位置）----

// IsInitialized 报告账号库是否已初始化，决定解锁门显示「解锁」还是「首次设置主密码」。
func (a *App) IsInitialized() (bool, error) {
	return store.IsInitialized()
}

// Unlock 用主密码打开（或首次初始化）账号库，成功后保存为 Go 侧单例。
func (a *App) Unlock(password string) error {
	if password == "" {
		return errors.New("主密码不能为空")
	}
	initialized, err := store.IsInitialized()
	if err != nil {
		return err
	}
	var s *store.Store
	if initialized {
		s, err = store.Open(password)
	} else {
		s, err = store.Initialize(password)
	}
	if err != nil {
		return err
	}
	if old := a.currentStore(); old != nil {
		_ = old.Close()
	}
	a.setStore(s)
	return nil
}

// Lock 加锁：关闭账号库，回到解锁门。
func (a *App) Lock() error {
	s := a.currentStore()
	if s != nil {
		if err := s.Close(); err != nil {
			return err
		}
	}
	a.setStore(nil)
	return nil
}

// VerifyMasterPassword 供导出等敏感操作前校验主密码。
func (a *App) VerifyMasterPassword(password string) (bool, error) {
	s := a.currentStore()
	if s == nil {
		return false, errors.New("账号库未解锁")
	}
	return s.VerifyMasterPassword(password)
}

// ChangeMasterPassword 重新派生密钥并在事务内重加密所有账号。
func (a *App) ChangeMasterPassword(newPassword string) error {
	if newPassword == "" {
		return errors.New("新主密码不能为空")
	}
	s := a.currentStore()
	if s == nil {
		return errors.New("账号库未解锁")
	}
	return s.ChangeMasterPassword(newPassword)
}

// CurrentDBPath 返回当前账号库路径。
func (a *App) CurrentDBPath() (string, error) {
	return store.CurrentDBPath()
}

// ChangeDBLocation 把账号库移动到新位置（旧版流程：先 Close，再 ChangeLocation，然后重新解锁）。
func (a *App) ChangeDBLocation(newPath string) error {
	newPath = filepath.FromSlash(newPath)
	if newPath == "" {
		return errors.New("路径不能为空")
	}
	if s := a.currentStore(); s != nil {
		if err := s.Close(); err != nil {
			return err
		}
	}
	if err := store.ChangeLocation(newPath); err != nil {
		return err
	}
	a.setStore(nil)
	return nil
}

// ---- 加密导入导出（[]byte ⇄ base64）----

// ExportAccounts 返回 base64 编码的加密导出文件内容。
func (a *App) ExportAccounts(password string) (string, error) {
	s := a.currentStore()
	if s == nil {
		return "", errors.New("账号库未解锁")
	}
	data, err := s.ExportAccounts(password)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// ImportAccounts 导入 base64 编码的导出文件，返回导入条数。
func (a *App) ImportAccounts(password, dataB64 string) (int, error) {
	s := a.currentStore()
	if s == nil {
		return 0, errors.New("账号库未解锁")
	}
	data, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil {
		return 0, err
	}
	return s.ImportAccounts(password, data)
}

// ---- 原生文件对话框与本地文件读写（替代旧版 zenity + os.ReadFile/WriteFile）----

// PickSaveFile 弹出保存文件对话框；用户取消时返回空串。
func (a *App) PickSaveFile(title, defaultName string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultName,
	})
}

// PickFile 弹出打开文件对话框；用户取消时返回空串。
func (a *App) PickFile(title string) (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: title})
}

// ReadFileB64 读取本地文件并返回 base64 内容。
func (a *App) ReadFileB64(path string) (string, error) {
	data, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// WriteFileB64 以 0600 权限把 base64 内容写入本地文件。
func (a *App) WriteFileB64(path, dataB64 string) error {
	data, err := base64.StdEncoding.DecodeString(dataB64)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.FromSlash(path), data, 0600)
}

// CopyText 把文本写入系统剪贴板（替代旧版属性弹窗里可选的 Entry + Ctrl+C）。
func (a *App) CopyText(text string) error {
	return runtime.ClipboardSetText(a.ctx, text)
}
