package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"health-tool/model"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want int
	}{
		{"补丁更新", "0.1.1", "0.1.2", -1},
		{"完全相同", "0.1.1", "0.1.1", 0},
		{"当前更新", "0.2.0", "0.1.9", 1},
		{"主版本更新", "0.9.9", "1.0.0", -1},
		{"带 v 前缀", "0.1.1", "v0.1.2", -1},
		{"两侧都带 v", "v0.1.2", "v0.1.1", 1},
		{"pre-release 低于正式版", "0.2.0", "0.2.0-beta", 1},
		{"pre-release 不高于对应正式版", "0.2.0-beta", "0.2.0", -1},
		{"缺段补零", "0.1", "0.1.0", 0},
		{"非法版本容错", "abc", "0.1.2", 0},
		{"空版本容错", "", "0.1.2", 0},
		{"非法 pre-release 形式", "0.2-beta.1", "0.2.0", -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := compareVersions(tc.a, tc.b); got != tc.want {
				t.Fatalf("compareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestCheckForUpdatesUpdateAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/hosea3000/health-tool/releases/latest" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v0.1.2","html_url":"https://github.com/hosea3000/health-tool/releases/tag/v0.1.2"}`))
	}))
	defer server.Close()

	result := checkForUpdates(server.Client(), "0.1.1", server.URL)
	if result.Status != model.UpdateStatusAvailable {
		t.Fatalf("status = %q, want update-available", result.Status)
	}
	if result.LatestVersion != "0.1.2" {
		t.Fatalf("latest version = %q, want 0.1.2", result.LatestVersion)
	}
	if result.ReleaseURL == "" {
		t.Fatal("release URL must be set")
	}
}

func TestCheckForUpdatesUpToDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"0.1.1","html_url":"https://github.com/hosea3000/health-tool/releases/tag/0.1.1"}`))
	}))
	defer server.Close()

	result := checkForUpdates(server.Client(), "0.1.1", server.URL)
	if result.Status != model.UpdateStatusUpToDate {
		t.Fatalf("status = %q, want up-to-date", result.Status)
	}
	if result.Message == "" {
		t.Fatal("message must not be empty")
	}
}

func TestCheckForUpdatesReleaseNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	result := checkForUpdates(server.Client(), "0.1.1", server.URL)
	if result.Status != model.UpdateStatusError {
		t.Fatalf("status = %q, want error", result.Status)
	}
	if result.Message == "" {
		t.Fatal("message must not be empty")
	}
}

func TestCheckForUpdatesNetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("should never be reached")
	}))
	server.Close()

	// 使用已关闭 server 的地址构造必然失败的请求
	result := checkForUpdates(&http.Client{Timeout: time.Second}, "0.1.1", server.URL)
	if result.Status != model.UpdateStatusError {
		t.Fatalf("status = %q, want error", result.Status)
	}
	if result.Message == "" {
		t.Fatal("message must not be empty")
	}
}

func TestAppCurrentVersion(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	if got := app.CurrentVersion(); got != version {
		t.Fatalf("current version = %q, want %q", got, version)
	}
}

func TestAppCheckForUpdatesDevShortCircuit(t *testing.T) {
	oldVersion := version
	version = "dev"
	defer func() { version = oldVersion }()

	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	result := app.CheckForUpdates()
	if result.Status != model.UpdateStatusUpToDate {
		t.Fatalf("dev status = %q, want up-to-date", result.Status)
	}
	if result.CurrentVersion != "dev" {
		t.Fatalf("dev current version = %q, want dev", result.CurrentVersion)
	}
	if result.Message == "" {
		t.Fatal("dev message must not be empty")
	}
}

func TestAppCheckForUpdatesInjectedVersionPassesThrough(t *testing.T) {
	oldVersion := version
	version = "0.1.1"
	defer func() { version = oldVersion }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"0.1.5","html_url":"https://github.com/hosea3000/health-tool/releases/tag/0.1.5"}`))
	}))
	defer server.Close()

	oldClient := updateClient
	updateClient = server.Client()
	defer func() { updateClient = oldClient }()

	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	result := app.CheckForUpdates()
	if result.Status != model.UpdateStatusAvailable {
		t.Fatalf("status = %q, want update-available", result.Status)
	}
	if result.CurrentVersion != "0.1.1" {
		t.Fatalf("current version = %q, want 0.1.1", result.CurrentVersion)
	}
}

// 确保 CheckForUpdates 与 CurrentVersion 是 App 的导出方法（Wails 绑定依赖导出方法）。
func TestAppExportsUpdateMethods(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	info := app.CheckForUpdates()
	if info.Status == "" {
		t.Fatal("CheckForUpdates must be callable")
	}
	_ = app.CurrentVersion()
}