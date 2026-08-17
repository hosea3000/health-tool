package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
		w.Write([]byte(`{"tag_name":"v0.1.2","html_url":"https://github.com/hosea3000/health-tool/releases/tag/v0.1.2","assets":[{"name":"health-tool.exe","browser_download_url":"https://example.com/downloads/health-tool.exe"}]}`))
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
	if result.DownloadURL != "https://example.com/downloads/health-tool.exe" {
		t.Fatalf("download URL = %q, want asset URL", result.DownloadURL)
	}
}

func TestCheckForUpdatesMultiAssetsMatchExe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v0.2.0","html_url":"https://github.com/hosea3000/health-tool/releases/tag/v0.2.0","assets":[{"name":"health-tool.sha256","browser_download_url":"https://example.com/downloads/health-tool.sha256"},{"name":"health-tool.exe","browser_download_url":"https://example.com/downloads/health-tool.exe"},{"name":"readme.txt","browser_download_url":"https://example.com/downloads/readme.txt"}]}`))
	}))
	defer server.Close()

	result := checkForUpdates(server.Client(), "0.1.9", server.URL)
	if result.Status != model.UpdateStatusAvailable {
		t.Fatalf("status = %q, want update-available", result.Status)
	}
	if result.DownloadURL != "https://example.com/downloads/health-tool.exe" {
		t.Fatalf("download URL = %q, want health-tool.exe asset URL", result.DownloadURL)
	}
}

func TestCheckForUpdatesAssetMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v0.3.0","html_url":"https://github.com/hosea3000/health-tool/releases/tag/v0.3.0","assets":[{"name":"other.zip","browser_download_url":"https://example.com/downloads/other.zip"}]}`))
	}))
	defer server.Close()

	result := checkForUpdates(server.Client(), "0.2.0", server.URL)
	if result.Status != model.UpdateStatusAvailable {
		t.Fatalf("status = %q, want update-available", result.Status)
	}
	if result.DownloadURL != "" {
		t.Fatalf("download URL = %q, want empty when asset missing", result.DownloadURL)
	}
}

func TestFindAssetDownloadURL(t *testing.T) {
	if got := findAssetDownloadURL(nil, updateAssetName); got != "" {
		t.Fatalf("nil assets = %q, want empty", got)
	}
	if got := findAssetDownloadURL([]githubAsset{{Name: "a.exe", BrowserDownloadURL: "u1"}}, updateAssetName); got != "" {
		t.Fatalf("no match = %q, want empty", got)
	}
	if got := findAssetDownloadURL([]githubAsset{{Name: updateAssetName, BrowserDownloadURL: "u2"}}, updateAssetName); got != "u2" {
		t.Fatalf("exact match = %q, want u2", got)
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

func TestBuildUpdateBat(t *testing.T) {
	content := buildUpdateBat("health-tool.exe")
	for _, fragment := range []string{
		"tasklist",
		"move /y \"%~dp0health-tool.exe.new\" \"%~dp0health-tool.exe\"",
		"del \"%~dp0health-tool.exe.new.version\"",
		"start \"\" \"%~dp0health-tool.exe\"",
		"del \"%~f0\"",
		"ping -n 2 127.0.0.1 >nul",
	} {
		if !strings.Contains(content, fragment) {
			t.Fatalf("bat 缺少关键片段 %q:\n%s", fragment, content)
		}
	}
	// 重试次数与脚本末尾不应残留 .new（成功路径已 move 完成）
	if !strings.Contains(content, "if %tries% lss 3") {
		t.Fatalf("bat 缺少重试逻辑:\n%s", content)
	}
	// timeout 在无控制台进程（CREATE_NO_WINDOW）下会直接报错退出，不得使用
	if strings.Contains(content, "timeout") {
		t.Fatalf("bat 不应使用 timeout 延时:\n%s", content)
	}
}

func TestBuildUpdateBatEscapesSpecialChars(t *testing.T) {
	// 空格、&、! 在引号内安全，不应额外转义，只需引号包裹
	content := buildUpdateBat("health tool&!.exe")
	if !strings.Contains(content, `"%~dp0health tool&!.exe.new"`) {
		t.Fatalf("bat 中文件名未加引号包裹:\n%s", content)
	}
	if strings.Contains(content, "%%") {
		t.Fatalf("文件名不含 %%%% 时不应出现 %%%% 转义:\n%s", content)
	}
	// % 必须转义为 %%（bat 中 %% 表示字面 %）
	content = buildUpdateBat("health%25.exe")
	if !strings.Contains(content, "health%%25.exe") {
		t.Fatalf("bat 未转义文件名中的 %%%%:\n%s", content)
	}
}

func TestPendingUpdateVersion(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "health-tool.exe")
	if v := pendingUpdateVersion(exe); v != "" {
		t.Fatalf("无 .new 时版本 = %q, want empty", v)
	}
	_, newPath, versionPath, _ := exeUpdatePaths(exe)
	if err := os.WriteFile(newPath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if v := pendingUpdateVersion(exe); v != "" {
		t.Fatalf("无版本标记时 = %q, want empty", v)
	}
	if err := os.WriteFile(versionPath, []byte("0.1.6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if v := pendingUpdateVersion(exe); v != "0.1.6" {
		t.Fatalf("版本 = %q, want 0.1.6", v)
	}
}

func TestCleanupPendingUpdateFiles(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "health-tool.exe")
	_, newPath, versionPath, _ := exeUpdatePaths(exe)
	os.WriteFile(newPath, []byte("x"), 0o644)
	os.WriteFile(versionPath, []byte("0.1.6"), 0o644)
	cleanupPendingUpdateFiles(exe)
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		t.Fatal(".new 应被删除")
	}
	if _, err := os.Stat(versionPath); !os.IsNotExist(err) {
		t.Fatal(".new.version 应被删除")
	}
}

func TestCleanupStaleVersionArtifact(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "health-tool.exe")
	_, newPath, versionPath, _ := exeUpdatePaths(exe)
	os.WriteFile(versionPath, []byte("0.1.6"), 0o644)
	cleanupStaleVersionArtifact(exe)
	if _, err := os.Stat(versionPath); !os.IsNotExist(err) {
		t.Fatal("孤儿 .new.version 应被删除")
	}
	// .new 存在时保留版本标记
	os.WriteFile(newPath, []byte("x"), 0o644)
	os.WriteFile(versionPath, []byte("0.1.6"), 0o644)
	cleanupStaleVersionArtifact(exe)
	if _, err := os.Stat(versionPath); err != nil {
		t.Fatal(".new 存在时 .new.version 应保留")
	}
}

func TestDirWritable(t *testing.T) {
	if !dirWritable(t.TempDir()) {
		t.Fatal("临时目录应可写")
	}
}

func TestExeUpdatePaths(t *testing.T) {
	exe := filepath.Join("Program Files", "health-tool", "health-tool.exe")
	part, new, version, bat := exeUpdatePaths(exe)
	if part != filepath.Join("Program Files", "health-tool", "health-tool.exe.part") {
		t.Fatalf("part path = %q", part)
	}
	if new != filepath.Join("Program Files", "health-tool", "health-tool.exe.new") {
		t.Fatalf("new path = %q", new)
	}
	if version != filepath.Join("Program Files", "health-tool", "health-tool.exe.new.version") {
		t.Fatalf("version path = %q", version)
	}
	if bat != filepath.Join("Program Files", "health-tool", "health-tool-update.bat") {
		t.Fatalf("bat path = %q", bat)
	}
}
func TestAppExportsUpdateMethods(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	info := app.CheckForUpdates()
	if info.Status == "" {
		t.Fatal("CheckForUpdates must be callable")
	}
	_ = app.CurrentVersion()
}
