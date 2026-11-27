package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloudshelf/internal/version"
)

func TestCheckLatest(t *testing.T) {
	orig := version.Version
	version.Version = "v0.1.0"
	defer func() { version.Version = orig }()

	wantAsset := ExpectedAssetName("windows", "amd64")
	if wantAsset != "cloudshelf-windows-amd64-installer.exe" {
		t.Fatalf("unexpected asset name: %s", wantAsset)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/liyuliang/cloudshelf/releases/latest" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(ghRelease{
			TagName: "v0.2.0",
			Name:    "v0.2.0",
			Body:    "新增自动更新",
			HTMLURL: "https://github.com/liyuliang/cloudshelf/releases/tag/v0.2.0",
			Assets: []ghAsset{
				{Name: "cloudshelf-windows-amd64-installer.exe", Size: 1024, BrowserDownloadURL: "https://example.com/win-installer.exe"},
				{Name: "cloudshelf-macos-universal.zip", Size: 1024, BrowserDownloadURL: "https://example.com/mac.zip"},
				{Name: "cloudshelf-linux-amd64.tar.gz", Size: 1024, BrowserDownloadURL: "https://example.com/linux.tar.gz"},
			},
		})
	}))
	defer srv.Close()

	info, err := CheckLatest(context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("CheckLatest error: %v", err)
	}
	if !info.HasUpdate {
		t.Fatalf("expected HasUpdate=true, got false")
	}
	if info.LatestVersion != "v0.2.0" {
		t.Fatalf("expected LatestVersion=v0.2.0, got %s", info.LatestVersion)
	}
	if !info.CanSelfUpdate || info.AssetURL == "" {
		t.Fatalf("expected CanSelfUpdate=true and non-empty AssetURL, got %+v", info)
	}
}
