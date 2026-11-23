// Package lanshare provides a lightweight embedded HTTP server for LAN-based
// read-only file sharing. It acts as a reverse proxy to S3: visitors with
// a valid token can browse directories and download files through any browser,
// without needing CloudShelf installed or S3 credentials.
//
// Design:
//   - Random port on 0.0.0.0 (auto-selected by OS)
//   - Token-protected URLs: /s/<token>/ for directory, /s/<token>/<path> for file
//   - 4-hour auto-expiry (configurable per share)
//   - Stream files from S3 directly (no local caching)
//   - Auto-detect LAN IPs for the shareable URL
package lanshare

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"cloudshelf/internal/model"
	"cloudshelf/internal/proxy"
)

// ShareEntry represents one active LAN share.
type ShareEntry struct {
	Token     string    `json:"token"`
	Name      string    `json:"name"`      // human-friendly label
	Bucket    string    `json:"bucket"`
	Prefix    string    `json:"prefix"`    // directory prefix to share (may be "")
	URL       string    `json:"url"`       // full URL for the first LAN IP
	URLs      []string  `json:"urls"`      // URLs for all detected LAN IPs
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
	account   *model.Account
}

// IsExpired reports whether the share has passed its expiry time.
func (e *ShareEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Server is the embedded HTTP file server for LAN sharing.
type Server struct {
	mu       sync.RWMutex
	shares   map[string]*ShareEntry // token -> entry
	srv      *http.Server
	listener net.Listener
	port     int
	done     chan struct{}
}

// New creates a new share server (not yet started).
func New() *Server {
	s := &Server{
		shares: make(map[string]*ShareEntry),
		done:   make(chan struct{}),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/s/", s.handleShare)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	s.srv = &http.Server{Handler: mux}
	return s
}

// Start begins listening on a random port. It returns the port number.
func (s *Server) Start() (int, error) {
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return 0, fmt.Errorf("share server listen: %w", err)
	}
	s.listener = ln
	s.port = ln.Addr().(*net.TCPAddr).Port
	go s.srv.Serve(ln)
	// start background expiry cleaner
	go s.cleanExpired()
	return s.port, nil
}

// Stop gracefully shuts down the share server.
func (s *Server) Stop() {
	select {
	case <-s.done:
		return // already stopped
	default:
		close(s.done)
	}
	if s.srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		s.srv.Shutdown(ctx)
	}
}

// Port returns the listening port, or 0 if not started.
func (s *Server) Port() int {
	return s.port
}

// CreateShare registers a new share and returns the ShareEntry with URLs.
func (s *Server) CreateShare(acc *model.Account, bucket, prefix, name string, ttl time.Duration) (*ShareEntry, error) {
	token, err := randomToken(8)
	if err != nil {
		return nil, err
	}
	if ttl <= 0 {
		ttl = 4 * time.Hour
	}
	now := time.Now()
	entry := &ShareEntry{
		Token:     token,
		Name:      name,
		Bucket:    bucket,
		Prefix:    prefix,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		account:   acc,
	}

	ips := LANAddresses()
	for _, ip := range ips {
		entry.URLs = append(entry.URLs, fmt.Sprintf("http://%s:%d/s/%s/", ip, s.port, token))
	}
	if len(entry.URLs) > 0 {
		entry.URL = entry.URLs[0]
	} else {
		entry.URL = fmt.Sprintf("http://127.0.0.1:%d/s/%s/", s.port, token)
	}

	s.mu.Lock()
	s.shares[token] = entry
	s.mu.Unlock()
	return entry, nil
}

// RemoveShare deletes a share by token.
func (s *Server) RemoveShare(token string) {
	s.mu.Lock()
	delete(s.shares, token)
	s.mu.Unlock()
}

// ListShares returns all active (non-expired) shares.
func (s *Server) ListShares() []*ShareEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*ShareEntry
	for _, e := range s.shares {
		if !e.IsExpired() {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

func (s *Server) getShare(token string) *ShareEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e := s.shares[token]
	if e != nil && e.IsExpired() {
		return nil
	}
	return e
}

// cleanExpired periodically removes expired shares.
func (s *Server) cleanExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			s.mu.Lock()
			for token, e := range s.shares {
				if e.IsExpired() {
					delete(s.shares, token)
				}
			}
			s.mu.Unlock()
		}
	}
}

// ---- HTTP Handler ----

func (s *Server) handleShare(w http.ResponseWriter, r *http.Request) {
	// URL: /s/<token>/[path...]
	path := strings.TrimPrefix(r.URL.Path, "/s/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	token := parts[0]
	subPath := ""
	if len(parts) > 1 {
		subPath = parts[1]
	}

	entry := s.getShare(token)
	if entry == nil {
		http.Error(w, "404 - 链接无效或已过期", http.StatusNotFound)
		return
	}

	fullPrefix := entry.Prefix + subPath

	// If subPath ends with "/" or is empty, treat as directory listing
	if subPath == "" || strings.HasSuffix(subPath, "/") {
		s.serveDirectory(w, r, entry, fullPrefix, token, subPath)
		return
	}

	// Otherwise, try to serve as file; if it looks like a directory, redirect
	s.serveFile(w, r, entry, fullPrefix, token, subPath)
}

func newS3Client(acc *model.Account) *s3.Client {
	creds := credentials.NewStaticCredentialsProvider(acc.AccessKeyID, acc.SecretAccessKey, "")
	region := acc.Region
	if region == "" {
		region = "us-east-1"
	}
	var ep *string
	if acc.Endpoint != "" {
		ep = &acc.Endpoint
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.Proxy = proxy.ProxyFuncFor(proxy.Current())
	httpClient := &http.Client{Transport: tr}
	return s3.New(s3.Options{
		Region:       region,
		Credentials:  creds,
		UsePathStyle: acc.UsePathStyle,
		BaseEndpoint: ep,
		HTTPClient:   httpClient,
	})
}

func (s *Server) serveDirectory(w http.ResponseWriter, r *http.Request, entry *ShareEntry, prefix, token, subPath string) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	client := newS3Client(entry.account)
	out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:    &entry.Bucket,
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		http.Error(w, "列出目录失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type item struct {
		Name  string
		IsDir bool
		Size  int64
		Time  string
		Href  string
	}
	var items []item

	// Directories (CommonPrefixes)
	for _, cp := range out.CommonPrefixes {
		p := aws.ToString(cp.Prefix)
		name := p[len(prefix):]
		items = append(items, item{
			Name:  name,
			IsDir: true,
			Href:  "/s/" + token + "/" + strings.TrimPrefix(p, entry.Prefix),
		})
	}

	// Files
	for _, obj := range out.Contents {
		k := aws.ToString(obj.Key)
		if k == prefix || strings.HasSuffix(k, "/") {
			continue
		}
		name := k[len(prefix):]
		t := ""
		if obj.LastModified != nil {
			t = obj.LastModified.Local().Format("2006-01-02 15:04")
		}
		items = append(items, item{
			Name:  name,
			IsDir: false,
			Size:  aws.ToInt64(obj.Size),
			Time:  t,
			Href:  "/s/" + token + "/" + strings.TrimPrefix(k, entry.Prefix),
		})
	}

	// Build breadcrumb
	type crumb struct{ Name, Href string }
	var crumbs []crumb
	crumbs = append(crumbs, crumb{entry.Bucket, "/s/" + token + "/"})
	if subPath != "" {
		segs := strings.Split(strings.TrimSuffix(subPath, "/"), "/")
		acc := ""
		for _, seg := range segs {
			if seg == "" {
				continue
			}
			acc += seg + "/"
			crumbs = append(crumbs, crumb{seg, "/s/" + token + "/" + acc})
		}
	}

	remaining := time.Until(entry.ExpiresAt)
	remainStr := formatDuration(remaining)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	dirTmpl.Execute(w, map[string]any{
		"Name":      entry.Name,
		"Items":     items,
		"Crumbs":    crumbs,
		"Remaining": remainStr,
		"Empty":     len(items) == 0,
	})
}

func (s *Server) serveFile(w http.ResponseWriter, r *http.Request, entry *ShareEntry, key, _, subPath string) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	client := newS3Client(entry.account)
	input := &s3.GetObjectInput{
		Bucket: &entry.Bucket,
		Key:    &key,
	}

	// Support Range requests
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		input.Range = &rangeHeader
	}

	out, err := client.GetObject(ctx, input)
	if err != nil {
		// Maybe it's a directory prefix without trailing /
		if !strings.HasSuffix(subPath, "/") {
			http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
			return
		}
		http.Error(w, "获取文件失败: "+err.Error(), http.StatusNotFound)
		return
	}
	defer out.Body.Close()

	// Set Content-Type
	ct := aws.ToString(out.ContentType)
	if ct == "" || ct == "application/octet-stream" || ct == "binary/octet-stream" {
		ext := filepath.Ext(key)
		if mt := mime.TypeByExtension(ext); mt != "" {
			ct = mt
		} else {
			ct = "application/octet-stream"
		}
	}
	w.Header().Set("Content-Type", ct)

	// Content-Disposition for download (but allow inline preview for common types)
	fileName := key
	if idx := strings.LastIndex(key, "/"); idx >= 0 {
		fileName = key[idx+1:]
	}
	if isPreviewable(ct) {
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileName))
	} else {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	}

	if out.ContentLength != nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", *out.ContentLength))
	}
	if rangeHeader != "" && out.ContentRange != nil {
		w.Header().Set("Content-Range", *out.ContentRange)
		w.WriteHeader(http.StatusPartialContent)
	}

	io.Copy(w, out.Body)
}

// ---- LAN IP Detection ----

// LANAddresses returns all non-loopback IPv4 addresses on private subnets.
func LANAddresses() []string {
	var addrs []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return addrs
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		// Skip common virtual interfaces
		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "vmnet") || strings.Contains(name, "vmware") ||
			strings.Contains(name, "vbox") || strings.Contains(name, "docker") ||
			strings.Contains(name, "veth") || strings.Contains(name, "br-") {
			continue
		}
		ifAddrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range ifAddrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP.To4()
			if ip == nil {
				continue
			}
			if isPrivateIP(ip) {
				addrs = append(addrs, ip.String())
			}
		}
	}
	return addrs
}

func isPrivateIP(ip net.IP) bool {
	if ip[0] == 10 {
		return true
	}
	if ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31 {
		return true
	}
	if ip[0] == 192 && ip[1] == 168 {
		return true
	}
	return false
}

// ---- Helpers ----

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func isPreviewable(ct string) bool {
	return strings.HasPrefix(ct, "image/") ||
		strings.HasPrefix(ct, "text/") ||
		strings.HasPrefix(ct, "video/") ||
		strings.HasPrefix(ct, "audio/") ||
		ct == "application/pdf"
}

func formatSize(b int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "已过期"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%d 小时 %d 分钟", h, m)
	}
	return fmt.Sprintf("%d 分钟", m)
}

// ---- HTML Template ----

var dirTmpl = template.Must(template.New("dir").Funcs(template.FuncMap{
	"formatSize": formatSize,
}).Parse(`<!doctype html>
<html lang="zh">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.Name}} - CloudShelf 分享</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,"Segoe UI","Microsoft YaHei",sans-serif;background:#0f1117;color:#e7eaf2;line-height:1.6}
.wrap{max-width:900px;margin:0 auto;padding:20px 24px}
.header{display:flex;align-items:center;gap:12px;padding:16px 0;border-bottom:1px solid #272c3c;margin-bottom:16px}
.logo{width:28px;height:28px;border-radius:8px;background:linear-gradient(135deg,#5b8cff,#8b6bff);flex:0 0 28px}
.title{font-size:16px;font-weight:700;flex:1}
.expire{font-size:11px;color:#6b7389;background:#1c2030;padding:4px 10px;border-radius:12px}
.crumbs{display:flex;align-items:center;gap:4px;font-size:13px;padding:8px 0 12px;flex-wrap:wrap}
.crumbs a{color:#7aa0ff;text-decoration:none;padding:2px 4px;border-radius:4px}
.crumbs a:hover{background:rgba(91,140,255,.14)}
.crumbs .sep{color:#6b7389}
table{width:100%;border-collapse:collapse}
th{text-align:left;padding:8px 12px;font-size:11px;font-weight:600;color:#6b7389;border-bottom:1px solid #272c3c;text-transform:uppercase}
td{padding:8px 12px;border-bottom:1px solid rgba(39,44,60,.5);font-size:13px}
tr:hover{background:#1c2030}
a.file-link{color:#e7eaf2;text-decoration:none;display:flex;align-items:center;gap:8px}
a.file-link:hover{color:#7aa0ff}
.icon{flex:0 0 18px;font-size:14px}
.size{color:#99a1b7;font-size:12px;text-align:right}
.time{color:#6b7389;font-size:12px;text-align:right}
.empty{text-align:center;padding:40px 0;color:#6b7389}
.footer{text-align:center;padding:20px 0;font-size:11px;color:#6b7389;border-top:1px solid #272c3c;margin-top:24px}
.dl-btn{display:inline-flex;align-items:center;gap:4px;padding:3px 10px;border-radius:4px;font-size:11px;color:#5b8cff;background:rgba(91,140,255,.1);text-decoration:none;border:1px solid rgba(91,140,255,.2)}
.dl-btn:hover{background:rgba(91,140,255,.2)}
</style>
</head>
<body>
<div class="wrap">
  <div class="header">
    <div class="logo"></div>
    <div class="title">{{.Name}}</div>
    <div class="expire">剩余 {{.Remaining}}</div>
  </div>
  <div class="crumbs">
    {{range $i, $c := .Crumbs}}
      {{if $i}}<span class="sep">/</span>{{end}}
      <a href="{{$c.Href}}">{{$c.Name}}</a>
    {{end}}
  </div>
  {{if .Empty}}
    <div class="empty">此目录为空</div>
  {{else}}
  <table>
    <thead><tr><th>名称</th><th style="width:90px;text-align:right">大小</th><th style="width:140px;text-align:right">修改时间</th><th style="width:60px"></th></tr></thead>
    <tbody>
    {{range .Items}}
    <tr>
      <td><a class="file-link" href="{{.Href}}"><span class="icon">{{if .IsDir}}📁{{else}}📄{{end}}</span>{{.Name}}</a></td>
      <td class="size">{{if .IsDir}}&mdash;{{else}}{{formatSize .Size}}{{end}}</td>
      <td class="time">{{if .Time}}{{.Time}}{{else}}&mdash;{{end}}</td>
      <td>{{if not .IsDir}}<a class="dl-btn" href="{{.Href}}" download>下载</a>{{end}}</td>
    </tr>
    {{end}}
    </tbody>
  </table>
  {{end}}
  <div class="footer">由 CloudShelf 提供 · 只读分享 · 链接到期后自动失效</div>
</div>
</body>
</html>`))
