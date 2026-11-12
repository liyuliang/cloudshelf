package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestProxyURL(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		want    string
		wantErr bool
	}{
		{"host+端口", Config{Mode: ModeHTTP, Host: "127.0.0.1", Port: 7890}, "http://127.0.0.1:7890", false},
		{"socks5", Config{Mode: ModeSOCKS5, Host: "10.0.0.1", Port: 1080}, "socks5://10.0.0.1:1080", false},
		{"端口写在地址里", Config{Mode: ModeHTTP, Host: "127.0.0.1:7890"}, "http://127.0.0.1:7890", false},
		{"完整 URL 带认证", Config{Mode: ModeHTTP, Host: "http://u:p@1.2.3.4:8080"}, "http://u:p@1.2.3.4:8080", false},
		{"表单认证覆盖 URL", Config{Mode: ModeHTTP, Host: "1.2.3.4:8080", Username: "a", Password: "b"}, "http://a:b@1.2.3.4:8080", false},
		{"清理首尾空格", Config{Mode: ModeHTTP, Host: " 1.2.3.4 ", Port: 8080}, "http://1.2.3.4:8080", false},
		{"缺端口", Config{Mode: ModeHTTP, Host: "1.2.3.4"}, "", true},
		{"缺地址", Config{Mode: ModeHTTP, Port: 8080}, "", true},
		{"非代理模式", Config{Mode: ModeNone}, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := c.cfg.ProxyURL()
			if c.wantErr {
				if err == nil {
					t.Fatalf("期望报错，实际得到 %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Fatalf("ProxyURL = %q，期望 %q", got, c.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	if err := (Config{Mode: ModeNone}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Config{Mode: ModeEnv}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Config{Mode: ModeHTTP, Host: "1.2.3.4", Port: 8080}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Config{Mode: "http2", Host: "1.2.3.4"}).Validate(); err == nil {
		t.Fatal("未知模式应报错")
	}
	// 空模式视为默认（跟随环境变量）
	if got := (Config{}).Normalize().Mode; got != ModeEnv {
		t.Fatalf("空模式归一化为 %q，期望 env", got)
	}
}

func TestProxyFuncFor(t *testing.T) {
	httpsURL := func(host string) *http.Request {
		return &http.Request{URL: &url.URL{Scheme: "https", Host: host}}
	}
	if fn := ProxyFuncFor(Config{Mode: ModeNone}); fn != nil {
		t.Fatal("直连模式应返回 nil 代理函数")
	}
	if fn := ProxyFuncFor(Config{Mode: ModeEnv}); fn == nil {
		t.Fatal("环境变量模式应返回代理函数")
	}
	if fn := ProxyFuncFor(Config{Mode: ModeHTTP, Host: "", Port: 0}); fn != nil {
		t.Fatal("配置不完整时应退化为直连")
	}

	fn := ProxyFuncFor(Config{Mode: ModeHTTP, Host: "1.2.3.4", Port: 8080, NoProxy: "example.org"})
	u, err := fn(httpsURL("s3.amazonaws.com"))
	if err != nil || u == nil || u.String() != "http://1.2.3.4:8080" {
		t.Fatalf("http 代理 = %v, err=%v", u, err)
	}
	u, err = fn(httpsURL("example.org"))
	if err != nil || u != nil {
		t.Fatalf("NO_PROXY 命中应直连: %v, err=%v", u, err)
	}
	u, err = fn(httpsURL("127.0.0.1:9000"))
	if err != nil || u != nil {
		t.Fatalf("回环地址应直连: %v, err=%v", u, err)
	}

	fn = ProxyFuncFor(Config{Mode: ModeSOCKS5, Host: "1.2.3.4", Port: 1080})
	u, err = fn(httpsURL("s3.amazonaws.com"))
	if err != nil || u == nil || u.String() != "socks5://1.2.3.4:1080" {
		t.Fatalf("socks5 代理 = %v, err=%v", u, err)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.json")

	// 文件不存在 → 默认配置
	got, err := LoadFrom(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != ModeEnv {
		t.Fatalf("默认模式 = %q，期望 env", got.Mode)
	}

	want := Config{Mode: ModeSOCKS5, Host: "10.0.0.1", Port: 1080, Username: "u", Password: "p", NoProxy: "localhost,192.168.0.0/16"}
	if err := SaveTo(p, want); err != nil {
		t.Fatal(err)
	}
	got, err = LoadFrom(p)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("读回 = %+v，期望 %+v", got, want)
	}
	if fi, err := os.Stat(p); err != nil {
		t.Fatal(err)
	} else if runtime.GOOS != "windows" && fi.Mode().Perm() != 0600 {
		t.Fatalf("配置文件权限 = %v，期望 0600", fi.Mode().Perm())
	}

	// 非法配置不落盘
	if err := SaveTo(p, Config{Mode: ModeHTTP, Host: ""}); err == nil {
		t.Fatal("非法配置应拒绝保存")
	}
	got, _ = LoadFrom(p)
	if got != want {
		t.Fatal("非法保存不应覆盖已有配置")
	}

	// 损坏的配置文件：报错且返回默认值
	if err := os.WriteFile(p, []byte("{不是 json"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err = LoadFrom(p)
	if err == nil {
		t.Fatal("损坏配置应报错")
	}
	if got.Mode != ModeEnv {
		t.Fatalf("损坏配置应回退默认值，实际 %q", got.Mode)
	}
}

func TestApplyLocked(t *testing.T) {
	applyLocked(Config{Mode: ModeHTTP, Host: "1.2.3.4", Port: 8080})
	if cur.Mode != ModeHTTP || client == nil {
		t.Fatalf("应用后状态不完整: %+v", cur)
	}
	if fn := client.Transport.(*http.Transport).Proxy; fn == nil {
		t.Fatal("http 代理模式应有代理函数")
	}
	applyLocked(Config{Mode: ModeNone})
	if fn := client.Transport.(*http.Transport).Proxy; fn != nil {
		t.Fatal("直连模式代理函数应为 nil")
	}
}

func TestCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	code, _, err := Check(Config{Mode: ModeNone}, srv.URL, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if code != http.StatusNoContent {
		t.Fatalf("状态码 = %d，期望 204", code)
	}
	if _, _, err := Check(Config{Mode: ModeNone}, "", time.Second); err == nil {
		t.Fatal("空地址应报错")
	}
	// 指向一个没有服务的代理：连接代理就应失败（不会访问外网）
	bad := Config{Mode: ModeHTTP, Host: "127.0.0.1", Port: 1}
	if _, _, err := Check(bad, "http://example.invalid/", 3*time.Second); err == nil {
		t.Fatal("不可达的代理应报错")
	}
}
