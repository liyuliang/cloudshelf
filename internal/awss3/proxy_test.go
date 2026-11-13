package awss3

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"cloudshelf/internal/model"
	"cloudshelf/internal/proxy"
)

// TestAWSHTTPClientUsesProxy 验证注入给 AWS SDK 的 HTTP 客户端确实把请求发给
// 配置的代理：起一个本地伪代理，用它访问一个不存在的外网地址（请求根本不会
// 出网，伪代理直接应答），断言伪代理收到过请求。
func TestAWSHTTPClientUsesProxy(t *testing.T) {
	var hit atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit.Store(true)
		if !strings.HasPrefix(r.RequestURI, "http://") {
			t.Errorf("代理应收到绝对 URI，实际 %q", r.RequestURI)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	host, portStr, ok := strings.Cut(strings.TrimPrefix(srv.URL, "http://"), ":")
	if !ok {
		t.Fatalf("无法解析测试代理地址 %q", srv.URL)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatal(err)
	}

	cli := awsHTTPClientFor(proxy.Config{Mode: proxy.ModeHTTP, Host: host, Port: port})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://s3.invalid/", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if !hit.Load() {
		t.Fatal("配置的代理没有收到请求")
	}
}

// TestNewClientGoesThroughProxy 端到端验证 s3.New 注入的 HTTP 客户端确实走配置的
// 代理：目标主机 s3.invalid 不会出网、也不做 DNS 解析，只有代理能给出响应。
func TestNewClientGoesThroughProxy(t *testing.T) {
	var hit atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit.Store(true)
		w.Header().Set("Content-Type", "application/xml")
		_, _ = io.WriteString(w, listBucketsXML)
	}))
	defer srv.Close()

	host, portStr, ok := strings.Cut(strings.TrimPrefix(srv.URL, "http://"), ":")
	if !ok {
		t.Fatalf("无法解析测试代理地址 %q", srv.URL)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatal(err)
	}

	cli := newClientWithProxy(&model.Account{
		Endpoint:        "http://s3.invalid",
		Region:          "us-east-1",
		UsePathStyle:    true,
		AccessKeyID:     "AK",
		SecretAccessKey: "SK",
	}, proxy.Config{Mode: proxy.ModeHTTP, Host: host, Port: port})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := cli.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		t.Fatal(err)
	}
	if !hit.Load() {
		t.Fatal("配置的代理没有收到 S3 请求")
	}
	if len(out.Buckets) != 1 || aws.ToString(out.Buckets[0].Name) != "demo" {
		t.Fatalf("响应解析异常: %+v", out.Buckets)
	}
}

const listBucketsXML = `<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Owner><ID>test</ID><DisplayName>test</DisplayName></Owner><Buckets><Bucket><Name>demo</Name><CreationDate>2024-01-01T00:00:00.000Z</CreationDate></Bucket></Buckets></ListAllMyBucketsResult>`

// TestAWSHTTPClientDirect 确认直连模式不会把请求交给代理。
func TestAWSHTTPClientDirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	cli := awsHTTPClientFor(proxy.Config{Mode: proxy.ModeNone})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := cli.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("状态码 = %d", resp.StatusCode)
	}
}
