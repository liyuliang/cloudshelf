package main

// 数据兼容回归：Wails 版复用 7 个零改动的后端包，本测试把「老数据格式」钉死成断言，
// 防止后续阶段在 facade 里不小心重写加解密 / 序列化 / schema。
// 全程用临时 HOME 隔离，绝不触碰用户真实的 ~/.cloudshelf。

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"cloudshelf/internal/crypto"
	"cloudshelf/internal/model"
	"cloudshelf/internal/proxy"
	"cloudshelf/internal/share"
	"cloudshelf/internal/store"
)

const testMaster = "compat-test-master-password"

// tempHome 把 os.UserHomeDir() 重定向到临时目录（Windows 读 USERPROFILE，Unix 读 HOME）。
func tempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("USERPROFILE", dir)
	t.Setenv("HOME", dir)
	return dir
}

// TestCompatSchema 断言 accounts/meta 表结构与列顺序不可变（store.ListAccounts 用固定 Scan 顺序）。
func TestCompatSchema(t *testing.T) {
	home := tempHome(t)
	s, err := store.Initialize(testMaster)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	db, err := sql.Open("sqlite", filepath.Join(home, ".cloudshelf", "data.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(`PRAGMA table_info(accounts)`)
	if err != nil {
		t.Fatal(err)
	}
	var cols []string
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		cols = append(cols, name)
	}
	rows.Close()

	want := []string{
		"id", "name", "endpoint", "region", "access_key_id", "secret_access_key",
		"use_path_style", "provider", "tenant_id", "client_id", "client_secret", "site_url",
	}
	if !reflect.DeepEqual(cols, want) {
		t.Fatalf("accounts 列顺序变化 = %v，期望 %v", cols, want)
	}
}

// TestCompatCrypto 断言 KDF 参数、salt/key 长度、AES-256-GCM 密文拼接格式与 verifier 行为。
func TestCompatCrypto(t *testing.T) {
	if crypto.Argon2Time != 3 || crypto.Argon2Memory != 64*1024 || crypto.Argon2Threads != 4 {
		t.Fatalf("Argon2id 参数变化: t=%d m=%d p=%d",
			crypto.Argon2Time, crypto.Argon2Memory, crypto.Argon2Threads)
	}
	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	if len(salt) != 32 {
		t.Fatalf("salt 长度 = %d，期望 32", len(salt))
	}
	key, err := crypto.DeriveKey(testMaster, salt)
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("scrypt key 长度 = %d，期望 32", len(key))
	}

	plain := []byte("cloudshelf-compat-plaintext")
	enc, err := crypto.Encrypt(key, plain)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		t.Fatalf("密文不是标准 base64: %v", err)
	}
	// nonce(12) || ciphertext || tag(16)
	if want := 12 + len(plain) + 16; len(raw) != want {
		t.Fatalf("密文拼接长度 = %d，期望 %d（nonce12 + 明文 + tag16）", len(raw), want)
	}
	dec, err := crypto.Decrypt(key, enc)
	if err != nil {
		t.Fatal(err)
	}
	if string(dec) != string(plain) {
		t.Fatal("加解密往返结果不一致")
	}

	verifier, err := crypto.CreateVerifier(key)
	if err != nil {
		t.Fatal(err)
	}
	if !crypto.CheckVerifier(key, verifier) {
		t.Fatal("正确密钥未能通过 verifier")
	}
	wrong, err := crypto.DeriveKey("wrong-password", salt)
	if err != nil {
		t.Fatal(err)
	}
	if crypto.CheckVerifier(wrong, verifier) {
		t.Fatal("错误密钥竟通过了 verifier")
	}
}

// TestCompatExportImport 断言导出文件外层 JSON 字段与 KDF 参数，并验证跨库导入往返。
func TestCompatExportImport(t *testing.T) {
	tempHome(t)
	s, err := store.Initialize(testMaster)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	accounts := []model.Account{
		{Name: "s3-acc", Provider: "s3", Endpoint: "http://localhost:9000", Region: "us-east-1",
			AccessKeyID: "AK", SecretAccessKey: "SK", UsePathStyle: true},
		{Name: "sp-acc", Provider: "sharepoint", TenantID: "T", ClientID: "C",
			ClientSecret: "CS", SiteURL: "https://contoso.sharepoint.com/sites/TeamA"},
	}
	for i := range accounts {
		if err := s.AddAccount(&accounts[i]); err != nil {
			t.Fatal(err)
		}
	}
	list, err := s.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].SecretAccessKey != "SK" || list[1].ClientSecret != "CS" {
		t.Fatalf("账号字段解密异常: %+v", list)
	}

	data, err := s.ExportAccounts(testMaster)
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Version      int    `json:"version"`
		KDF          string `json:"kdf"`
		ArgonTime    uint32 `json:"argon_time"`
		ArgonMemory  uint32 `json:"argon_memory"`
		ArgonThreads uint8  `json:"argon_threads"`
		Salt         string `json:"salt"`
		Payload      string `json:"payload"`
	}
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	if f.Version != 1 || f.KDF != "argon2id" || f.ArgonTime != 3 ||
		f.ArgonMemory != 65536 || f.ArgonThreads != 4 {
		t.Fatalf("导出文件外层格式变化: %+v", f)
	}
	payloadRaw, err := base64.StdEncoding.DecodeString(f.Payload)
	if err != nil || len(payloadRaw) < 12+16 {
		t.Fatalf("payload 不是 base64(nonce||ct||tag): %v", err)
	}

	// 换一个全新的库导入，验证自包含文件可被旧格式解析。
	home2 := t.TempDir()
	t.Setenv("USERPROFILE", home2)
	t.Setenv("HOME", home2)
	s2, err := store.Initialize(testMaster)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	n, err := s2.ImportAccounts(testMaster, data)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("导入条数 = %d，期望 2", n)
	}
	list2, err := s2.ListAccounts()
	if err != nil {
		t.Fatal(err)
	}
	if len(list2) != 2 || list2[0].SecretAccessKey != "SK" || list2[1].SiteURL != accounts[1].SiteURL {
		t.Fatalf("导入后账号内容不一致: %+v", list2)
	}
}

// TestCompatShareFile 断言 .cloudshelf（旧称 .s3share）外层 JSON、需口令/免口令两种打开方式。
func TestCompatShareFile(t *testing.T) {
	acc := &model.Account{Name: "share-src", Provider: "s3", Endpoint: "e", Region: "r",
		AccessKeyID: "AK", SecretAccessKey: "SK"}

	withPw, err := share.Create("my-share", "my-bucket", acc, 24*time.Hour, "pw123")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Version      int    `json:"version"`
		KDF          string `json:"kdf"`
		ArgonTime    uint32 `json:"argon_time"`
		ArgonMemory  uint32 `json:"argon_memory"`
		ArgonThreads uint8  `json:"argon_threads"`
		Salt         string `json:"salt"`
		Payload      string `json:"payload"`
		UnlockCode   string `json:"unlock_code"`
	}
	if err := json.Unmarshal(withPw, &f); err != nil {
		t.Fatal(err)
	}
	if f.Version != 1 || f.KDF != "argon2id" || f.ArgonTime != 3 ||
		f.ArgonMemory != 65536 || f.ArgonThreads != 4 {
		t.Fatalf("分享文件外层格式变化: %+v", f)
	}
	sess, err := share.Open(withPw, "pw123")
	if err != nil {
		t.Fatal(err)
	}
	if sess.Bucket != "my-bucket" || sess.Provider != "s3" || !sess.Account.ReadOnly {
		t.Fatalf("分享会话内容异常: %+v", sess)
	}
	if _, err := share.Open(withPw, "wrong-pw"); err == nil {
		t.Fatal("错误口令竟能打开分享文件")
	}

	// 免口令：随机口令写进 unlock_code，Open 时无需输入。
	noPw, err := share.Create("free-share", "bucket2", acc, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	var f2 struct {
		UnlockCode string `json:"unlock_code"`
	}
	if err := json.Unmarshal(noPw, &f2); err != nil {
		t.Fatal(err)
	}
	if f2.UnlockCode == "" {
		t.Fatal("免口令分享未写入 unlock_code")
	}
	sess2, err := share.Open(noPw, "")
	if err != nil {
		t.Fatal(err)
	}
	if sess2.Bucket != "bucket2" {
		t.Fatalf("免口令分享 Bucket = %q", sess2.Bucket)
	}
}

// TestCompatProxyConfig 断言 config.json 的 JSON 键与非 Windows 下的 0600 权限。
func TestCompatProxyConfig(t *testing.T) {
	home := tempHome(t)
	cfg := proxy.Config{Mode: proxy.ModeHTTP, Host: "127.0.0.1", Port: 7890,
		Username: "u", Password: "p", NoProxy: "localhost"}
	if err := proxy.Set(cfg); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".cloudshelf", "config.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("config.json 权限 = %o，期望 0600", info.Mode().Perm())
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Proxy struct {
			Mode     string `json:"mode"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			Password string `json:"password"`
			NoProxy  string `json:"no_proxy"`
		} `json:"proxy"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Proxy.Mode != proxy.ModeHTTP || decoded.Proxy.Port != 7890 ||
		decoded.Proxy.NoProxy != "localhost" {
		t.Fatalf("config.json 结构变化: %+v", decoded.Proxy)
	}
}

// TestFacadeDTOConversion 断言 facade 的 DTO 转换不丢任何字段（绑定层唯一的数据变形）。
func TestFacadeDTOConversion(t *testing.T) {
	a := model.Account{ID: 7, Name: "n", Provider: "sharepoint", Endpoint: "e", Region: "r",
		AccessKeyID: "AK", SecretAccessKey: "SK", Bucket: "my-bucket", UsePathStyle: true,
		TenantID: "T", ClientID: "C", ClientSecret: "CS", SiteURL: "https://x"}
	back := accountToDTO(a).toModel()
	if !reflect.DeepEqual(a, back) {
		t.Fatalf("AccountDTO 往返丢字段:\n%+v\n%+v", a, back)
	}

	c := proxy.Config{Mode: proxy.ModeSOCKS5, Host: "h", Port: 1080,
		Username: "u", Password: "p", NoProxy: "n"}
	backCfg := proxyToDTO(c).toConfig()
	if !reflect.DeepEqual(c, backCfg) {
		t.Fatalf("ProxyDTO 往返丢字段:\n%+v\n%+v", c, backCfg)
	}
}
