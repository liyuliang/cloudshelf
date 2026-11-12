package awss3

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"s3client/internal/model"
	"s3client/internal/proxy"
)

// testClient 返回一个指向本地伪 S3 服务的客户端（直连，不走代理）。
func testClient(endpoint string) *s3.Client {
	return newClientWithProxy(&model.Account{
		Endpoint:        endpoint,
		Region:          "us-east-1",
		UsePathStyle:    true,
		AccessKeyID:     "AK",
		SecretAccessKey: "SK",
	}, proxy.Config{Mode: proxy.ModeNone})
}

func namesOf(objects []Object) []string {
	names := make([]string, len(objects))
	for i, o := range objects {
		names[i] = o.Name
	}
	return names
}

// TestListObjectsSinglePage 验证正常列表只发一次请求、带 delimiter=/，
// 即使服务端返回 IsTruncated=true 也不翻页；截断标志回传给 UI 用于提示。
func TestListObjectsSinglePage(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		q := r.URL.Query()
		if q.Get("continuation-token") != "" {
			t.Errorf("正常列表不应翻页，却带了 continuation-token=%q", q.Get("continuation-token"))
		}
		if q.Get("delimiter") != "/" || q.Get("prefix") != "pre/" || q.Get("list-type") != "2" {
			t.Errorf("查询参数异常: %v", q)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = io.WriteString(w, listPage1XML)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	objects, truncated, err := listObjects(ctx, testClient(srv.URL), "demo", "pre/", "")
	if err != nil {
		t.Fatal(err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("请求次数 = %d，期望 1（不应翻页）", got)
	}
	if !truncated {
		t.Fatal("IsTruncated=true 时应回传截断标志")
	}
	want := []string{"sub/", "b.txt", "a.txt"}
	if got := namesOf(objects); !reflect.DeepEqual(got, want) {
		t.Fatalf("列表 = %v，期望 %v", got, want)
	}
}

// TestSearchObjectsUsesServerPrefix 验证服务端搜索把关键字拼进 prefix、
// 去掉 delimiter，因而能命中首页之外、以及子目录里的文件。
func TestSearchObjectsUsesServerPrefix(t *testing.T) {
	var gotPrefix, gotDelimiter string
	var sawContinuation bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		gotPrefix = q.Get("prefix")
		gotDelimiter = q.Get("delimiter")
		if q.Get("continuation-token") != "" {
			sawContinuation = true
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = io.WriteString(w, searchResultXML)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	objects, truncated, err := listObjects(ctx, testClient(srv.URL), "demo", "pre/", "rep")
	if err != nil {
		t.Fatal(err)
	}
	if gotPrefix != "pre/rep" {
		t.Fatalf("prefix = %q，期望 pre/rep（关键字需拼到当前目录之后）", gotPrefix)
	}
	if gotDelimiter != "" {
		t.Fatalf("搜索不应带 delimiter，实际 %q", gotDelimiter)
	}
	if sawContinuation {
		t.Fatal("搜索不应翻页")
	}
	if truncated {
		t.Fatal("该响应 IsTruncated=false，不应标记截断")
	}
	// pre/reports/2024/jan.pdf 位于子目录，带 delimiter 的首页列表永远看不到它；
	// 服务端前缀搜索能把它找出来，目录占位对象 pre/reports/ 应被跳过。
	want := []string{"reports/2024/jan.pdf", "report.txt"}
	if got := namesOf(objects); !reflect.DeepEqual(got, want) {
		t.Fatalf("搜索结果 = %v，期望 %v", got, want)
	}
}

const listPage1XML = `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>demo</Name><Prefix>pre/</Prefix><KeyCount>3</KeyCount><MaxKeys>1000</MaxKeys>
  <IsTruncated>true</IsTruncated><NextContinuationToken>page2</NextContinuationToken>
  <Contents><Key>pre/a.txt</Key><LastModified>2024-01-01T00:00:00.000Z</LastModified><ETag>etag-a</ETag><Size>1</Size><StorageClass>STANDARD</StorageClass></Contents>
  <Contents><Key>pre/b.txt</Key><LastModified>2024-01-03T00:00:00.000Z</LastModified><ETag>etag-b</ETag><Size>2</Size><StorageClass>STANDARD</StorageClass></Contents>
  <CommonPrefixes><Prefix>pre/sub/</Prefix></CommonPrefixes>
</ListBucketResult>`

const searchResultXML = `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>demo</Name><Prefix>pre/rep</Prefix><KeyCount>3</KeyCount><MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents><Key>pre/report.txt</Key><LastModified>2024-02-01T00:00:00.000Z</LastModified><ETag>e1</ETag><Size>5</Size><StorageClass>STANDARD</StorageClass></Contents>
  <Contents><Key>pre/reports/2024/jan.pdf</Key><LastModified>2024-02-02T00:00:00.000Z</LastModified><ETag>e2</ETag><Size>6</Size><StorageClass>STANDARD</StorageClass></Contents>
  <Contents><Key>pre/reports/</Key><LastModified>2024-01-01T00:00:00.000Z</LastModified><ETag>e3</ETag><Size>0</Size><StorageClass>STANDARD</StorageClass></Contents>
</ListBucketResult>`

// TestResolvePathDirStopsAtLastExisting 验证逐级下钻：有多少级目录就探测多少级，
// 遇到不存在的层级就停在最后一个存在的目录，而不是跳到深层空目录。
func TestResolvePathDirStopsAtLastExisting(t *testing.T) {
	var mu sync.Mutex
	var prefixes []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		mu.Lock()
		prefixes = append(prefixes, q.Get("prefix"))
		mu.Unlock()
		if q.Get("max-keys") != "1" {
			t.Errorf("max-keys = %q，期望 1", q.Get("max-keys"))
		}
		w.Header().Set("Content-Type", "application/xml")
		if q.Get("prefix") == "a/b/c/" {
			_, _ = io.WriteString(w, emptyListXML)
		} else {
			_, _ = io.WriteString(w, oneKeyListXML)
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dir, err := resolvePathDir(ctx, testClient(srv.URL), "demo", []string{"a", "b", "c"})
	if err != nil {
		t.Fatal(err)
	}
	if dir != "a/b/" {
		t.Fatalf("停留目录 = %q，期望 a/b/", dir)
	}
	mu.Lock()
	got := append([]string(nil), prefixes...)
	mu.Unlock()
	want := []string{"a/", "a/b/", "a/b/c/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("逐级请求 = %v，期望 %v", got, want)
	}
}

const oneKeyListXML = `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>demo</Name><Prefix>a/</Prefix><KeyCount>1</KeyCount><MaxKeys>1</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents><Key>a/x.txt</Key><LastModified>2024-01-01T00:00:00.000Z</LastModified><ETag>e</ETag><Size>1</Size><StorageClass>STANDARD</StorageClass></Contents>
</ListBucketResult>`

const emptyListXML = `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>demo</Name><Prefix>a/b/c/</Prefix><KeyCount>0</KeyCount><MaxKeys>1</MaxKeys>
  <IsTruncated>false</IsTruncated>
</ListBucketResult>`
