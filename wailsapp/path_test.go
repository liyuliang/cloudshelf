package main

import (
	"reflect"
	"testing"
)

// 与 internal/ui/path_test.go 完全相同的用例，确保复刻的解析逻辑没有漂移。
func TestParseObjectPathParity(t *testing.T) {
	const cur = "cur-bucket"
	const dir = "summit-raw-3003-cdz79/summit-raw-3003-cdz79-rave-api-3750677447/"
	const file = "rave_api_data.tgz"

	cases := []struct {
		name    string
		raw     string
		bucket  string
		dir     string
		keyword string
		ok      bool
	}{
		{"带 artifact 前缀的完整路径", "artifact:s3:s3.amazonaws.com:s3-dashboard-summit-prod:" + dir + file, "s3-dashboard-summit-prod", dir, file, true},
		{"不带 artifact 前缀的完整路径", "s3:s3.amazonaws.com:s3-dashboard-summit-prod:" + dir + file, "s3-dashboard-summit-prod", dir, file, true},
		{"短路径按当前 bucket", dir + file, cur, dir, file, true},
		{"前后空白", "  " + dir + file + "  ", cur, dir, file, true},
		{"目录路径无搜索词", dir, cur, dir, "", true},
		{"单级目录也被识别为路径", "sub/" + file, cur, "sub/", file, true},
		{"普通关键字不跳转", "rave", "", "", "", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p, ok := parseObjectPath(c.raw, cur)
			if ok != c.ok {
				t.Fatalf("ok = %v，期望 %v", ok, c.ok)
			}
			if !ok {
				return
			}
			want := objectPath{bucket: c.bucket, dir: c.dir, keyword: c.keyword}
			if p != want {
				t.Fatalf("解析结果 = %+v，期望 %+v", p, want)
			}
		})
	}
}

func TestDirSegments(t *testing.T) {
	cases := []struct {
		dir  string
		want []string
	}{
		{"a/b/c/", []string{"a", "b", "c"}},
		{"a/", []string{"a"}},
		{"", nil},
		{"/", nil},
	}
	for _, c := range cases {
		if got := dirSegments(c.dir); !reflect.DeepEqual(got, c.want) {
			t.Fatalf("dirSegments(%q) = %v，期望 %v", c.dir, got, c.want)
		}
	}
}

func TestParentPrefix(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a/b/c/", "a/b/"},
		{"a/b/", "a/"},
		{"a/", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := parentPrefix(c.in); got != c.want {
			t.Fatalf("parentPrefix(%q) = %q，期望 %q", c.in, got, c.want)
		}
	}
}
