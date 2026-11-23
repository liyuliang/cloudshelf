// Package proxy 管理软件配置里的网络代理（HTTP/HTTPS、SOCKS5）：
// 从 ~/.cloudshelf/config.json 读写配置，并为标准库 HTTP 客户端（SharePoint）
// 与 AWS SDK（S3）提供代理函数，GUI/CLI 共用同一份配置。
package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http/httpproxy"

	"cloudshelf/internal/appdir"
	"cloudshelf/internal/reqlog"
)

// 代理模式（配置文件里的 mode 取值）。
const (
	ModeNone   = "none"   // 不使用代理，直连
	ModeEnv    = "env"    // 跟随系统环境变量 HTTP_PROXY / HTTPS_PROXY / NO_PROXY
	ModeHTTP   = "http"   // 自定义 HTTP/HTTPS 代理
	ModeSOCKS5 = "socks5" // 自定义 SOCKS5 代理（域名解析交给代理，便于翻墙）
)

// Config 是软件配置里的代理设置。
type Config struct {
	Mode     string `json:"mode"`
	Host     string `json:"host,omitempty"` // 代理地址：host、host:port 或 scheme://[user:pass@]host:port
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	NoProxy  string `json:"no_proxy,omitempty"` // 逗号分隔，命中则直连
}

// fileConfig 是配置文件的顶层结构，预留扩展其它软件配置的位置。
type fileConfig struct {
	Proxy Config `json:"proxy"`
}

// Default 返回出厂默认配置：跟随系统环境变量（与不配置时的行为一致）。
func Default() Config { return Config{Mode: ModeEnv} }

// Normalize 补全缺省值并清理空白。
func (c Config) Normalize() Config {
	c.Mode = strings.TrimSpace(c.Mode)
	if c.Mode == "" {
		c.Mode = ModeEnv
	}
	c.Host = strings.TrimSpace(c.Host)
	c.NoProxy = strings.TrimSpace(c.NoProxy)
	c.Username = strings.TrimSpace(c.Username)
	return c
}

// Validate 校验配置是否可用（模式非法或代理地址不完整时报错）。
func (c Config) Validate() error {
	c = c.Normalize()
	switch c.Mode {
	case ModeNone, ModeEnv:
		return nil
	case ModeHTTP, ModeSOCKS5:
		_, err := c.ProxyURL()
		return err
	default:
		return fmt.Errorf("未知的代理模式 %q", c.Mode)
	}
}

// scheme 返回该模式代理 URL 的协议。
func (c Config) scheme() string {
	if c.Normalize().Mode == ModeSOCKS5 {
		return "socks5"
	}
	return "http"
}

// ProxyURL 返回可直接用于 http.Transport.Proxy 的代理地址，
// 支持 host、host:port、scheme://[user:pass@]host:port 三种写法。
func (c Config) ProxyURL() (string, error) {
	c = c.Normalize()
	if c.Mode != ModeHTTP && c.Mode != ModeSOCKS5 {
		return "", errors.New("当前模式不使用自定义代理")
	}
	s := c.Host
	if s == "" {
		return "", errors.New("请填写代理服务器地址")
	}
	if !strings.Contains(s, "://") {
		s = c.scheme() + "://" + s
	}
	u, err := url.Parse(s)
	if err != nil || u.Hostname() == "" {
		return "", errors.New("代理服务器地址无效，示例：127.0.0.1")
	}
	u.Scheme = c.scheme()
	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	if u.Port() == "" {
		if c.Port <= 0 || c.Port > 65535 {
			return "", errors.New("请填写代理端口（1-65535），示例：7890")
		}
		u.Host = net.JoinHostPort(u.Hostname(), strconv.Itoa(c.Port))
	}
	if c.Username != "" {
		u.User = url.UserPassword(c.Username, c.Password)
	}
	return u.String(), nil
}

// ProxyFuncFor 把给定配置翻译成 http.Transport.Proxy 需要的函数；返回 nil 表示直连。
func ProxyFuncFor(c Config) func(*http.Request) (*url.URL, error) {
	switch c.Normalize().Mode {
	case ModeEnv:
		return http.ProxyFromEnvironment
	case ModeHTTP, ModeSOCKS5:
		u, err := c.ProxyURL()
		if err != nil {
			return nil
		}
		cfg := &httpproxy.Config{
			HTTPProxy:  u,
			HTTPSProxy: u,
			NoProxy:    c.Normalize().NoProxy,
		}
		fn := cfg.ProxyFunc()
		return func(req *http.Request) (*url.URL, error) {
			if req == nil || req.URL == nil {
				return nil, nil
			}
			return fn(req.URL)
		}
	default:
		return nil
	}
}

// clientFor 用给定配置构造一个独立的 HTTP 客户端（独立连接池，便于测试与保存前试连）。
func clientFor(c Config) *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.Proxy = ProxyFuncFor(c)
	return &http.Client{Transport: reqlog.WrapTransport(tr)}
}

// ==================== 持久化 ====================

// ConfigPath 返回配置文件路径（~/.cloudshelf/config.json）。
func ConfigPath() (string, error) {
	dir, err := appdir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadFrom 读取指定路径的配置；文件不存在返回默认配置。
func LoadFrom(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Default(), err
	}
	var f fileConfig
	if err := json.Unmarshal(b, &f); err != nil {
		return Default(), fmt.Errorf("配置文件已损坏: %w", err)
	}
	return f.Proxy.Normalize(), nil
}

// SaveTo 把配置原子写入指定路径（临时文件 + rename，权限 0600）。
func SaveTo(path string, c Config) error {
	if err := c.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(fileConfig{Proxy: c.Normalize()}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// Load 读取默认路径的配置。
func Load() (Config, error) {
	p, err := ConfigPath()
	if err != nil {
		return Default(), err
	}
	return LoadFrom(p)
}

// Save 把配置写入默认路径。
func Save(c Config) error {
	p, err := ConfigPath()
	if err != nil {
		return err
	}
	return SaveTo(p, c)
}

// ==================== 运行时状态 ====================

var (
	stateMu sync.RWMutex
	loaded  bool
	cur     Config
	client  *http.Client
)

func applyLocked(c Config) {
	cur = c.Normalize()
	client = clientFor(cur)
}

func loadLocked() {
	c, err := Load()
	if err != nil {
		// 配置读不出来不阻塞启动，退回默认（跟随环境变量）
		c = Default()
	}
	applyLocked(c)
	loaded = true
}

// Current 返回当前生效的配置（首次调用时从配置文件加载）。
func Current() Config {
	stateMu.RLock()
	if loaded {
		c := cur
		stateMu.RUnlock()
		return c
	}
	stateMu.RUnlock()

	stateMu.Lock()
	defer stateMu.Unlock()
	if !loaded {
		loadLocked()
	}
	return cur
}

// Set 校验并保存配置，保存成功后立即生效（进行中的传输不受影响，新请求走新代理）。
func Set(c Config) error {
	c = c.Normalize()
	if err := c.Validate(); err != nil {
		return err
	}
	if err := Save(c); err != nil {
		return err
	}
	stateMu.Lock()
	applyLocked(c)
	stateMu.Unlock()
	return nil
}

// Client 返回当前配置生效的 HTTP 客户端，供标准库请求（SharePoint 等）使用。
func Client() *http.Client {
	Current()
	stateMu.RLock()
	defer stateMu.RUnlock()
	return client
}

// ==================== 连通性测试 ====================

// Check 用配置 c 向 rawURL 发一次 GET，返回状态码与耗时。
// 用于界面的「测试连接」，不会修改当前生效的配置。
func Check(c Config, rawURL string, timeout time.Duration) (int, time.Duration, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return 0, 0, errors.New("请填写测试地址")
	}
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}
	if timeout <= 0 {
		timeout = 12 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, 0, err
	}
	cli := clientFor(c)
	defer cli.CloseIdleConnections()
	start := time.Now()
	resp, err := cli.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return 0, elapsed, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	return resp.StatusCode, elapsed, nil
}
