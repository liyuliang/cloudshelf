package version

import (
	"strconv"
	"strings"
)

// Version 由构建时通过 -ldflags "-X cloudshelf/internal/version.Version=..." 注入；
// 本地开发构建默认为 "dev"。
var Version = "dev"

// IsDev 判断当前是否为未注入正式版本号的开发构建。
func IsDev(v string) bool {
	s := strings.TrimSpace(strings.ToLower(v))
	return s == "" || s == "dev" || strings.HasPrefix(s, "build-")
}

// ParseSemver 解析 vX.Y.Z（可带 v 前缀与 -prerelease/+meta 后缀）为三段数字及是否含预发布后缀。
func ParseSemver(raw string) (major, minor, patch int, prerelease string, ok bool) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(strings.TrimPrefix(s, "v"), "V")
	if idx := strings.IndexByte(s, '+'); idx >= 0 {
		s = s[:idx]
	}
	if idx := strings.IndexByte(s, '-'); idx >= 0 {
		prerelease = s[idx+1:]
		s = s[:idx]
	}
	parts := strings.Split(s, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, 0, 0, "", false
	}
	var err error
	if major, err = strconv.Atoi(parts[0]); err != nil {
		return 0, 0, 0, "", false
	}
	if minor, err = strconv.Atoi(parts[1]); err != nil {
		return 0, 0, 0, "", false
	}
	if len(parts) == 3 {
		if patch, err = strconv.Atoi(parts[2]); err != nil {
			return 0, 0, 0, "", false
		}
	}
	return major, minor, patch, prerelease, true
}

// CompareSemver 比较两个版本号：a < b 返回 -1，a == b 返回 0，a > b 返回 1。
// 若任一方无法解析为 semver，则回退到字符串比较（忽略大小写和 v 前缀）。
func CompareSemver(a, b string) int {
	aMaj, aMin, aPat, aPre, aOK := ParseSemver(a)
	bMaj, bMin, bPat, bPre, bOK := ParseSemver(b)
	if !aOK || !bOK {
		na := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(a)), "v")
		nb := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(b)), "v")
		if na < nb {
			return -1
		}
		if na > nb {
			return 1
		}
		return 0
	}
	if aMaj != bMaj {
		if aMaj < bMaj {
			return -1
		}
		return 1
	}
	if aMin != bMin {
		if aMin < bMin {
			return -1
		}
		return 1
	}
	if aPat != bPat {
		if aPat < bPat {
			return -1
		}
		return 1
	}
	// 正式版 > 预发布版
	if aPre == "" && bPre != "" {
		return 1
	}
	if aPre != "" && bPre == "" {
		return -1
	}
	if aPre < bPre {
		return -1
	}
	if aPre > bPre {
		return 1
	}
	return 0
}

// IsNewer 判断 latest 是否比 current 更新。开发构建（dev / build-*）不视为需要自动提示升级。
func IsNewer(current, latest string) bool {
	if IsDev(current) || strings.TrimSpace(latest) == "" {
		return false
	}
	return CompareSemver(current, latest) < 0
}
