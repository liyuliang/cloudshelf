package awss3

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"cloudshelf/internal/model"
	"cloudshelf/internal/reqlog"
)

// TestReproBucketLogs 复现「进入 S3 账号后一堆 err 日志、bucket 列表不显示」。
func TestReproBucketLogs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("list-type") == "2" {
			// 模拟 bucket 无权限：ListObjectsV2 返回 403
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `<Error><Code>AccessDenied</Code><Message>Access Denied</Message></Error>`)
			return
		}
		// ListBuckets：返回 3 个桶
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
<Owner><ID>1</ID><DisplayName>me</DisplayName></Owner>
<Buckets>
<Bucket><Name>alpha</Name><CreationDate>2024-01-01T00:00:00.000Z</CreationDate></Bucket>
<Bucket><Name>beta</Name><CreationDate>2024-01-01T00:00:00.000Z</CreationDate></Bucket>
<Bucket><Name>gamma</Name><CreationDate>2024-01-01T00:00:00.000Z</CreationDate></Bucket>
</Buckets></ListAllMyBucketsResult>`)
	}))
	defer srv.Close()

	// 关掉环境代理对本地测试服务的干扰

	var mu sync.Mutex
	var entries []reqlog.Entry
	off := reqlog.Subscribe(func(e reqlog.Entry) {
		mu.Lock()
		entries = append(entries, e)
		mu.Unlock()
	})
	defer off()

	acc := &model.Account{
		Provider:        "s3",
		Endpoint:        srv.URL,
		Region:          "us-east-1",
		AccessKeyID:     "test",
		SecretAccessKey: "test",
		UsePathStyle:    true,
	}
	infos, err := ListBucketsWithAccess(acc)
	t.Logf("err=%v infos=%+v", err, infos)
	for _, e := range entries {
		t.Logf("LOG method=%s status=%d kind=%q error=%q url=%s",
			e.Method, e.Status, e.Kind, e.Error, e.URL[strings.Index(e.URL, srv.URL):])
	}
}
