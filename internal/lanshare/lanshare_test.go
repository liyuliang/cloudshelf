package lanshare

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloudshelf/internal/model"
)

func TestCreateAndListShares(t *testing.T) {
	srv := New()
	defer srv.Stop()

	// S3 share
	s3Acc := &model.Account{ID: 1, Name: "S3-Acc", Provider: "s3"}
	s3Share, err := srv.CreateShare(s3Acc, "my-bucket", "photos/", "My Photos", time.Hour)
	if err != nil {
		t.Fatalf("CreateShare (S3) failed: %v", err)
	}
	if s3Share.Provider != "s3" || s3Share.Bucket != "my-bucket" || s3Share.Prefix != "photos/" {
		t.Errorf("Unexpected S3 share entry: %+v", s3Share)
	}

	// SharePoint share
	spAcc := &model.Account{ID: 2, Name: "SP-Acc", Provider: "sharepoint"}
	spShare, err := srv.CreateSharePointShare(spAcc, "drive-123", "Documents", "Reports/", "SP Reports", time.Hour)
	if err != nil {
		t.Fatalf("CreateSharePointShare failed: %v", err)
	}
	if spShare.Provider != "sharepoint" || spShare.DriveID != "drive-123" || spShare.Bucket != "Documents" || spShare.Prefix != "Reports/" {
		t.Errorf("Unexpected SharePoint share entry: %+v", spShare)
	}

	// RWS share
	rwsAcc := &model.Account{ID: 3, Name: "RWS-Acc", Provider: "rws"}
	rwsShare, err := srv.CreateRWSShare(rwsAcc, "STUDY_01", "Trial Alpha", "", "RWS Alpha", time.Hour)
	if err != nil {
		t.Fatalf("CreateRWSShare failed: %v", err)
	}
	if rwsShare.Provider != "rws" || rwsShare.StudyOID != "STUDY_01" || rwsShare.Bucket != "Trial Alpha" {
		t.Errorf("Unexpected RWS share entry: %+v", rwsShare)
	}

	shares := srv.ListShares()
	if len(shares) != 3 {
		t.Fatalf("Expected 3 active shares, got %d", len(shares))
	}

	// Remove one share
	srv.RemoveShare(spShare.Token)
	shares = srv.ListShares()
	if len(shares) != 2 {
		t.Fatalf("Expected 2 active shares after removal, got %d", len(shares))
	}
}

func TestShareNotFoundAndExpired(t *testing.T) {
	srv := New()
	defer srv.Stop()

	ts := httptest.NewServer(srv.srv.Handler)
	defer ts.Close()

	// 1. Invalid token
	resp, err := http.Get(ts.URL + "/s/nonexistent_token/")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for invalid token, got %d", resp.StatusCode)
	}

	// 2. Expired token
	acc := &model.Account{ID: 1, Name: "Test", Provider: "s3"}
	entry, err := srv.CreateShare(acc, "bucket", "", "expired-share", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	entry.ExpiresAt = time.Now().Add(-time.Hour) // force expired
	resp2, err := http.Get(ts.URL + "/s/" + entry.Token + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for expired share, got %d", resp2.StatusCode)
	}
}

func TestRWSStudyDirectoryListing(t *testing.T) {
	srv := New()
	defer srv.Stop()

	ts := httptest.NewServer(srv.srv.Handler)
	defer ts.Close()

	rwsAcc := &model.Account{
		ID:              10,
		Name:            "RWS Account",
		Provider:        "rws",
		AccessKeyID:     "testuser",
		SecretAccessKey: "testpass",
		Region:          "testdomain",
	}

	// Create study-level share
	share, err := srv.CreateRWSShare(rwsAcc, "STUDY_XYZ", "Study XYZ", "", "Study Share", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Visit root of study: /s/<token>/
	resp, err := http.Get(ts.URL + "/s/" + share.Token + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "datasets/") {
		t.Errorf("Expected datasets/ folder in HTML, got:\n%s", bodyStr)
	}
	if !strings.Contains(bodyStr, "subjects.csv") {
		t.Errorf("Expected subjects.csv in HTML, got:\n%s", bodyStr)
	}
	if !strings.Contains(bodyStr, "sites.csv") {
		t.Errorf("Expected sites.csv in HTML, got:\n%s", bodyStr)
	}

	// 2. Visit datasets without trailing slash: should redirect to datasets/
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/s/"+share.Token+"/datasets", nil)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // don't follow redirect
		},
	}
	respDir, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer respDir.Body.Close()
	if respDir.StatusCode != http.StatusMovedPermanently {
		t.Errorf("Expected 301 redirect for /datasets, got %d", respDir.StatusCode)
	}
}
