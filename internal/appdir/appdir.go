// Package appdir 定义应用数据目录（~/.cloudshelf）。
//
// 项目由 s3client 更名为 cloudshelf，数据目录随之从 ~/.s3client 迁移到
// ~/.cloudshelf：首次访问时若新目录不存在而旧目录存在，整体搬迁（优先
// rename，跨盘退回复制），并修正 db_location 指针文件里仍指向旧目录的
// 绝对路径，保证老用户升级后数据、配置、主密码全部原样可用。
package appdir

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	newDirName = ".cloudshelf"
	oldDirName = ".s3client"
)

// Dir 返回应用数据目录 ~/.cloudshelf（不存在则以 0700 创建）。
// 目录内含 data.db、db_location、config.json 等应用数据。
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	newDir := filepath.Join(home, newDirName)
	oldDir := filepath.Join(home, oldDirName)
	if err := migrate(oldDir, newDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(newDir, 0700); err != nil {
		return "", err
	}
	return newDir, nil
}

// migrate 把旧数据目录整体搬到新目录（幂等：新目录已存在则不动作）。
func migrate(oldDir, newDir string) error {
	if _, err := os.Stat(newDir); err == nil {
		return nil // 新目录已存在（迁移过，或用户手动建过），旧目录原样保留
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if _, err := os.Stat(oldDir); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil // 全新安装，无可迁移
		}
		return err
	}
	if err := moveTree(oldDir, newDir); err != nil {
		return err
	}
	fixDBPointer(oldDir, newDir)
	return nil
}

// moveTree 整体移动目录：优先 rename（同盘瞬间完成），失败退回复制 + 删除。
func moveTree(oldDir, newDir string) error {
	if err := os.Rename(oldDir, newDir); err == nil {
		return nil
	}
	if err := copyTree(oldDir, newDir); err != nil {
		return err
	}
	return os.RemoveAll(oldDir)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0600)
	})
}

// fixDBPointer 修正 db_location 指针文件：数据库被自定义到旧目录内时，
// 搬家后路径要跟着改写到新目录，否则会去已删除的旧路径打开数据库。
func fixDBPointer(oldDir, newDir string) {
	pf := filepath.Join(newDir, "db_location")
	b, err := os.ReadFile(pf)
	if err != nil {
		return // 指针文件不存在 = 数据库就在默认位置，无需修正
	}
	p := strings.TrimSpace(string(b))
	if p == "" {
		return
	}
	np := remapPath(p, oldDir, newDir)
	if np != p {
		_ = os.WriteFile(pf, []byte(np), 0600)
	}
}

// remapPath 把位于 oldDir 之下的绝对路径映射到 newDir 之下；不在其内则原样返回。
func remapPath(p, oldDir, newDir string) string {
	if p == oldDir {
		return newDir
	}
	prefix := oldDir + string(filepath.Separator)
	if strings.HasPrefix(p, prefix) {
		return filepath.Join(newDir, p[len(prefix):])
	}
	return p
}
