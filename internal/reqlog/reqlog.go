// Package reqlog 提供 HTTP 请求日志：通过 RoundTripper 包装拦截所有 HTTP 请求，
// 记录方法、URL、状态码、耗时和错误，并通过回调推送给订阅方（如 Wails 前端）。
package reqlog

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Entry 是一条请求日志。
type Entry struct {
	ID       int64  `json:"id"`
	Time     string `json:"time"`     // 本地时间 HH:MM:SS
	Method   string `json:"method"`   // GET / PUT / DELETE …
	URL      string `json:"url"`      // 完整请求 URL
	Status   int    `json:"status"`   // HTTP 状态码；0 表示网络错误
	Duration int64  `json:"duration"` // 毫秒
	Error    string `json:"error"`    // 非空表示请求失败
	Kind     string `json:"kind"`     // 空串为普通请求；KindProbe 为权限探测
}

// KindProbe 标记「预期会失败的权限探测」请求，界面不把它当错误告警。
const KindProbe = "probe"

var (
	seq       atomic.Int64
	mu        sync.RWMutex
	listeners []func(Entry)
)

// Subscribe 注册一个日志监听函数，返回取消函数。
func Subscribe(fn func(Entry)) func() {
	mu.Lock()
	listeners = append(listeners, fn)
	idx := len(listeners) - 1
	mu.Unlock()
	return func() {
		mu.Lock()
		defer mu.Unlock()
		if idx < len(listeners) {
			listeners[idx] = nil
		}
	}
}

func emit(e Entry) {
	mu.RLock()
	defer mu.RUnlock()
	for _, fn := range listeners {
		if fn != nil {
			fn(e)
		}
	}
}

type probeKey struct{}

// WithProbe 标记 ctx：由它发出的请求是权限探测，403 等否定结果属于预期，
// 日志面板会与真实错误分开显示。标记随 context 值传到最终的 *http.Request。
func WithProbe(ctx context.Context) context.Context {
	return context.WithValue(ctx, probeKey{}, true)
}

func kindOf(req *http.Request) string {
	if req != nil && req.Context().Value(probeKey{}) != nil {
		return KindProbe
	}
	return ""
}

func record(method, rawURL string, status int, elapsed time.Duration, err error, kind string) {
	e := Entry{
		ID:       seq.Add(1),
		Time:     time.Now().Format("15:04:05"),
		Method:   method,
		URL:      rawURL,
		Status:   status,
		Duration: elapsed.Milliseconds(),
		Kind:     kind,
	}
	if err != nil {
		e.Error = err.Error()
	}
	emit(e)
}

// ==================== Transport 包装 ====================

// WrapTransport 把一个 http.RoundTripper 包装成带日志的版本。
func WrapTransport(inner http.RoundTripper) http.RoundTripper {
	if inner == nil {
		inner = http.DefaultTransport
	}
	return &loggingTransport{inner: inner}
}

// UnwrapTransport 从 WrapTransport 返回的包装中取出原始 RoundTripper。
// 如果 rt 不是本包的包装类型，原样返回。
func UnwrapTransport(rt http.RoundTripper) http.RoundTripper {
	if lt, ok := rt.(*loggingTransport); ok {
		return lt.inner
	}
	return rt
}

type loggingTransport struct {
	inner http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := t.inner.RoundTrip(req)
	elapsed := time.Since(start)
	status := 0
	if resp != nil {
		status = resp.StatusCode
	}
	record(req.Method, req.URL.String(), status, elapsed, err, kindOf(req))
	return resp, err
}

// ==================== aws.HTTPClient 包装 ====================

// Doer 与 aws.HTTPClient 方法签名一致（Do(*http.Request) (*http.Response, error)）。
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// WrapDoer 把任意 Doer（包括 aws.HTTPClient）包装成带日志的版本。
func WrapDoer(inner Doer) Doer {
	return &loggingDoer{inner: inner}
}

type loggingDoer struct {
	inner Doer
}

func (d *loggingDoer) Do(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := d.inner.Do(req)
	elapsed := time.Since(start)
	status := 0
	if resp != nil {
		status = resp.StatusCode
	}
	record(req.Method, req.URL.String(), status, elapsed, err, kindOf(req))
	return resp, err
}
