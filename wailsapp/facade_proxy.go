package main

import (
	"errors"
	"fmt"
	"time"

	"s3client/internal/proxy"
)

// ProxyDTO 与 proxy.Config 字段一一对应，作为绑定传输结构（避免 json tag 影响绑定）。
type ProxyDTO struct {
	Mode     string
	Host     string
	Port     int
	Username string
	Password string
	NoProxy  string
}

func (d ProxyDTO) toConfig() proxy.Config {
	return proxy.Config{
		Mode:     d.Mode,
		Host:     d.Host,
		Port:     d.Port,
		Username: d.Username,
		Password: d.Password,
		NoProxy:  d.NoProxy,
	}
}

func proxyToDTO(c proxy.Config) ProxyDTO {
	return ProxyDTO{
		Mode:     c.Mode,
		Host:     c.Host,
		Port:     c.Port,
		Username: c.Username,
		Password: c.Password,
		NoProxy:  c.NoProxy,
	}
}

// ProxyModes 返回四种代理模式及其显示标签，供前端下拉框使用（与旧版 settings.go 一致）。
func (a *App) ProxyModes() []ProxyModeDTO {
	return []ProxyModeDTO{
		{Mode: proxy.ModeEnv, Label: "跟随系统环境变量（HTTP_PROXY 等）"},
		{Mode: proxy.ModeNone, Label: "不使用代理（直连）"},
		{Mode: proxy.ModeHTTP, Label: "HTTP/HTTPS 代理"},
		{Mode: proxy.ModeSOCKS5, Label: "SOCKS5 代理"},
	}
}

type ProxyModeDTO struct {
	Mode  string
	Label string
}

// GetProxy 读取当前生效的代理配置。
func (a *App) GetProxy() ProxyDTO {
	return proxyToDTO(proxy.Current())
}

// SetProxy 校验并保存代理配置，保存后立即生效（进行中的传输不受影响）。
func (a *App) SetProxy(d ProxyDTO) error {
	return proxy.Set(d.toConfig())
}

// ValidateProxy 在保存/测试前校验配置（旧版 proxySettingsContent.build 调用 c.Validate()）。
func (a *App) ValidateProxy(d ProxyDTO) error {
	return d.toConfig().Validate()
}

// CheckProxy 测试连通性；timeoutSec<=0 时沿用旧版 12 秒默认值。
func (a *App) CheckProxy(d ProxyDTO, testURL string, timeoutSec int) (string, error) {
	if testURL == "" {
		return "", errors.New("请填写测试地址")
	}
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 12 * time.Second
	}
	code, elapsed, err := proxy.Check(d.toConfig(), testURL, timeout)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("代理返回 HTTP %d，耗时 %s", code, elapsed.Round(time.Millisecond)), nil
}
