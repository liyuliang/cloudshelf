package appdir

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMigrateOldDir 断言旧 ~/.s3client 首次访问时整体迁移到 ~/.cloudshelf，
// 数据库、配置、db_location 指针全部跟过去。
func TestMigrateOldDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)

	oldDir := filepath.Join(home, ".s3client")
	if err := os.MkdirAll(oldDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "data.db"), []byte("db-bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	// 数据库被自定义到旧目录内的子目录，指针文件必须跟着改写。
	sub := filepath.Join(oldDir, "custom")
	if err := os.MkdirAll(sub, 0700); err != nil {
		t.Fatal(err)
	}
	customDB := filepath.Join(sub, "other.db")
	if err := os.WriteFile(filepath.Join(oldDir, "db_location"), []byte(customDB), 0600); err != nil {
		t.Fatal(err)
	}

	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".cloudshelf"); dir != want {
		t.Fatalf("Dir() = %q, want %q", dir, want)
	}
	b, err := os.ReadFile(filepath.Join(dir, "data.db"))
	if err != nil {
		t.Fatalf("迁移后读不到 data.db: %v", err)
	}
	if string(b) != "db-bytes" {
		t.Fatalf("data.db 内容 = %q", b)
	}
	p, err := os.ReadFile(filepath.Join(dir, "db_location"))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "custom", "other.db")
	if string(p) != want {
		t.Fatalf("db_location = %q, want %q", p, want)
	}
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Fatalf("旧目录应已移走, err=%v", err)
	}

	// 幂等：再次调用不报错，且不再动数据。
	if _, err := Dir(); err != nil {
		t.Fatal(err)
	}
}

// TestDirFreshInstall 断言全新安装（无旧目录）直接创建新目录。
func TestDirFreshInstall(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)

	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatalf("新目录未创建: %v", err)
	}
}

// TestRemapPath 断言只有位于旧目录之下的路径才被映射。
func TestRemapPath(t *testing.T) {
	oldDir := filepath.Join("home", ".s3client")
	newDir := filepath.Join("home", ".cloudshelf")
	cases := []struct{ in, want string }{
		{filepath.Join(oldDir, "data.db"), filepath.Join(newDir, "data.db")},
		{filepath.Join(oldDir, "custom", "a.db"), filepath.Join(newDir, "custom", "a.db")},
		{filepath.Join("home", "elsewhere", "a.db"), filepath.Join("home", "elsewhere", "a.db")},
		{oldDir, newDir},
	}
	for _, c := range cases {
		if got := remapPath(c.in, oldDir, newDir); got != c.want {
			t.Errorf("remapPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
