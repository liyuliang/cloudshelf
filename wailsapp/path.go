package main

import "strings"

// objectPath 与 internal/ui/ui.go 的同名类型逻辑一致。
// 旧版是行为真源；这里复刻一份，因为 P5 之前 internal/ui 保持零改动。
type objectPath struct {
	bucket  string
	dir     string
	keyword string
}

// SearchTargetDTO 是搜索框输入解析后交给前端的结果：
// 目标 bucket、实际停留目录、搜索词，以及是否需要发起服务端搜索。
type SearchTargetDTO struct {
	Bucket   string
	Dir      string
	Keyword  string
	IsSearch bool
}

// parseObjectPath 解析粘贴的对象路径；ok=false 表示这是普通搜索关键字，不做跳转。
// 最后一个冒号之后为 key、其前一字段为 bucket，冒号前缀段数不限：
//
//	artifact:s3:s3.amazonaws.com:<bucket>:<key>
//	s3:<endpoint>:<bucket>:<key>
//	<key>  （当前 bucket + 完整 key；有多少个 / 就下钻多少级目录）
func parseObjectPath(raw, currentBucket string) (objectPath, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return objectPath{}, false
	}
	bucket := currentBucket
	key := s
	if colon := strings.LastIndex(s, ":"); colon >= 0 {
		tail := s[colon+1:]
		if strings.Contains(tail, "/") {
			key = tail
			head := s[:colon]
			if j := strings.LastIndex(head, ":"); j >= 0 {
				bucket = head[j+1:]
			} else {
				bucket = head
			}
		}
	}
	if !strings.Contains(key, "/") {
		return objectPath{}, false
	}
	isDir := strings.HasSuffix(key, "/")
	key = strings.Trim(key, "/")
	if key == "" {
		return objectPath{}, false
	}
	if bucket == "" {
		bucket = currentBucket
	}
	if isDir {
		return objectPath{bucket: bucket, dir: key + "/"}, true
	}
	i := strings.LastIndex(key, "/")
	if i < 0 {
		return objectPath{}, false
	}
	return objectPath{bucket: bucket, dir: key[:i+1], keyword: key[i+1:]}, true
}

// dirSegments 把目录前缀拆成逐级目录名：有多少个 "/" 就有多少级。
func dirSegments(dir string) []string {
	t := strings.Trim(dir, "/")
	if t == "" {
		return nil
	}
	return strings.Split(t, "/")
}

// parentPrefix 返回上一级前缀，如 "a/b/c/" -> "a/b/"（复刻 internal/ui/ui.go:1679）。
func parentPrefix(prefix string) string {
	trimmed := prefix
	if len(trimmed) > 0 && trimmed[len(trimmed)-1] == '/' {
		trimmed = trimmed[:len(trimmed)-1]
	}
	for i := len(trimmed) - 1; i >= 0; i-- {
		if trimmed[i] == '/' {
			return trimmed[:i+1]
		}
	}
	return ""
}

// ResolveSearchInput 复刻旧版 buildObjectView 搜索框的完整路径处理：
// 完整路径先按 / 数量逐级下钻、停在最后找到的目录，再把末段文件名作为搜索词；
// 普通关键字则在当前目录做服务端前缀搜索。
func (a *App) ResolveSearchInput(acc AccountDTO, currentBucket, currentPrefix, raw string) (SearchTargetDTO, error) {
	kw := strings.TrimSpace(raw)
	p, ok := parseObjectPath(raw, currentBucket)
	if !ok {
		return SearchTargetDTO{
			Bucket:   currentBucket,
			Dir:      currentPrefix,
			Keyword:  kw,
			IsSearch: kw != "",
		}, nil
	}
	dir, err := a.ResolvePathDir(acc, p.bucket, dirSegments(p.dir))
	if err != nil {
		return SearchTargetDTO{}, err
	}
	return SearchTargetDTO{
		Bucket:   p.bucket,
		Dir:      dir,
		Keyword:  p.keyword,
		IsSearch: p.keyword != "",
	}, nil
}
