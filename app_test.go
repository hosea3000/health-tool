package main

import (
	"context"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"health-tool/domain"
	"health-tool/model"
	"health-tool/store"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
)

func TestAppStartsInWaitingState(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})

	if got := app.Status().State; got != "waiting" {
		t.Fatalf("startup state = %q, want waiting", got)
	}
}

func TestAppSavesReminderSettings(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	app.settingsPath = filepath.Join(t.TempDir(), "settings.json")

	if !app.SaveSettings(45, 5, true) {
		t.Fatal("valid reminder settings were not saved")
	}
	if got := app.GetSettings().ReminderMinutes; got != 45 {
		t.Fatalf("reminder minutes = %d, want 45", got)
	}
	if got := app.GetSettings().RestMinutes; got != 5 {
		t.Fatalf("rest minutes = %d, want 5", got)
	}
	if got := app.Status().ReminderMinutes; got != 45 {
		t.Fatalf("status reminder minutes = %d, want 45", got)
	}
}

func TestAppRecordsOnlyKnownActivity(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})

	app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: time.Unix(0, 0).Add(time.Minute)})
	if got := app.Status().State; got != "working" {
		t.Fatalf("known activity state = %q, want working", got)
	}
}

func TestAppNotifiesOnceWhenWorkSegmentReachesReminder(t *testing.T) {
	start := time.Unix(0, 0)
	now := start
	notifications := 0
	app := newApp(func() time.Time { return now }, func() { notifications++ })
	app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: start})
	for minute := 4; minute < 60; minute += 4 {
		at := start.Add(time.Duration(minute) * time.Minute)
		app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: at})
	}

	now = start.Add(60 * time.Minute)
	status := app.Status()
	app.Status()

	if notifications != 1 {
		t.Fatalf("notifications = %d, want 1", notifications)
	}
	if status.RestMinutes != 3 || status.RestRemainingSeconds != 180 {
		t.Fatalf("rest status = %d minutes, %d seconds; want 3 minutes, 180 seconds", status.RestMinutes, status.RestRemainingSeconds)
	}
}

func TestAppNotifiesWhenAnAutomaticWorkSegmentStarts(t *testing.T) {
	start := time.Unix(0, 0)
	starts := 0
	app := newApp(func() time.Time { return start }, func() {})
	app.notifyStarted = func(int) { starts++ }

	app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: start})
	app.recordActivity(domain.EffectiveActivity{Kind: domain.Click, At: start.Add(time.Minute)})
	if starts != 1 {
		t.Fatalf("startup notifications = %d, want 1", starts)
	}

	app.monitor.PauseForIdle()
	app.recordActivity(domain.EffectiveActivity{Kind: domain.Click, At: start.Add(7 * time.Minute)})
	if starts != 2 {
		t.Fatalf("auto-resume notifications = %d, want 2", starts)
	}
}

func TestAppTimelineStartsWithWorkAndTracksDuration(t *testing.T) {
	start := time.Unix(0, 0)
	now := start
	app := newApp(func() time.Time { return now }, func() {})

	app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: start})
	now = start.Add(90 * time.Second)

	entries := app.Timeline()
	if len(entries) != 1 {
		t.Fatalf("timeline entries = %d, want 1", len(entries))
	}
	entry := entries[0]
	if entry.Kind != "working" || entry.EndedAt != nil || entry.DurationSeconds != 90 {
		t.Fatalf("working entry = %+v, want ongoing 90-second entry", entry)
	}
}

func TestAppTimelineRecordsIdlePauseAndRest(t *testing.T) {
	start := time.Unix(0, 0)
	now := start
	settings := model.Settings{ReminderMinutes: 1, RestMinutes: 1, NotificationsEnabled: true}
	app := newAppWithSettings(func() time.Time { return now }, func() {}, settings)

	app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: start})
	now = start.Add(5 * time.Minute)
	app.Status()
	now = start.Add(6 * time.Minute)
	app.recordActivity(domain.EffectiveActivity{Kind: domain.Click, At: now})
	now = start.Add(7 * time.Minute)
	app.Status()

	entries := app.Timeline()
	if len(entries) != 4 {
		t.Fatalf("timeline entries = %d, want 4", len(entries))
	}
	for i, want := range []string{"working", "idle-paused", "working", "resting"} {
		if entries[i].Kind != want {
			t.Fatalf("timeline entry %d kind = %q, want %q", i, entries[i].Kind, want)
		}
	}
	if entries[0].EndedAt == nil || entries[1].EndedAt == nil || entries[2].EndedAt == nil {
		t.Fatalf("completed timeline entries have missing end time: %+v", entries)
	}

	now = start.Add(8 * time.Minute)
	app.Status()
	entries = app.Timeline()
	if entries[3].EndedAt == nil {
		t.Fatal("rest entry is still open after rest ended")
	}
}

func TestAppTimelineStartsEmpty(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})

	if entries := app.Timeline(); len(entries) != 0 {
		t.Fatalf("timeline entries = %d, want empty", len(entries))
	}
}

func TestAppAllowsRequestedQuit(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	app.requestQuit()

	if app.beforeClose(context.Background()) {
		t.Fatal("requested quit was intercepted as a window hide")
	}
}

func TestTimelineRestoresSameDayRecords(t *testing.T) {
	now := time.Unix(1000000, 0)
	path := filepath.Join(t.TempDir(), "timeline.json")
	savedAt := now.Add(time.Minute)
	_ = store.SaveTimelineFile(path, store.TimelineFile{
		Date:    now.Format("2006-01-02"),
		SavedAt: savedAt,
		Entries: []model.TimelineEntry{
			{Kind: "working", StartedAt: now, EndedAt: &now},
			{Kind: "working", StartedAt: now.Add(time.Minute)},
		},
	})
	app := newApp(func() time.Time { return now }, func() {})
	app.timelinePath = path

	app.loadTimelineLocked(now)
	entries := app.Timeline()
	if len(entries) != 2 {
		t.Fatalf("restored entries = %d, want 2", len(entries))
	}
	if entries[1].EndedAt == nil || !entries[1].EndedAt.Equal(savedAt) {
		t.Fatalf("open record should be closed at savedAt, got %+v", entries[1])
	}
}

func TestTimelineIgnoresStaleDayFile(t *testing.T) {
	now := time.Unix(1000000, 0)
	path := filepath.Join(t.TempDir(), "timeline.json")
	_ = store.SaveTimelineFile(path, store.TimelineFile{
		Date:    now.AddDate(0, 0, -1).Format("2006-01-02"),
		SavedAt: now,
		Entries: []model.TimelineEntry{{Kind: "working", StartedAt: now}},
	})
	app := newApp(func() time.Time { return now }, func() {})
	app.timelinePath = path

	app.loadTimelineLocked(now)
	if entries := app.Timeline(); len(entries) != 0 {
		t.Fatalf("stale-day entries = %d, want empty", len(entries))
	}
}

func TestTimelinePersistsOnTransition(t *testing.T) {
	start := time.Unix(0, 0)
	path := filepath.Join(t.TempDir(), "timeline.json")
	app := newApp(func() time.Time { return start }, func() {})
	app.timelinePath = path

	app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: start.Add(time.Minute)})

	file, err := store.LoadTimelineFile(path)
	if err != nil {
		t.Fatalf("load persisted timeline: %v", err)
	}
	if len(file.Entries) != 1 || file.Entries[0].Kind != "working" || file.Entries[0].EndedAt != nil {
		t.Fatalf("persisted entries = %+v, want one open working record", file.Entries)
	}
}

func TestTimelineShutdownClosesOpenRecord(t *testing.T) {
	start := time.Unix(0, 0)
	path := filepath.Join(t.TempDir(), "timeline.json")
	now := start
	app := newApp(func() time.Time { return now }, func() {})
	app.timelinePath = path
	app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: start.Add(time.Minute)})

	now = start.Add(2 * time.Minute)
	app.shutdown(context.Background())

	file, err := store.LoadTimelineFile(path)
	if err != nil {
		t.Fatalf("load persisted timeline: %v", err)
	}
	if len(file.Entries) != 1 || file.Entries[0].EndedAt == nil || !file.Entries[0].EndedAt.Equal(now) {
		t.Fatalf("persisted entries = %+v, want open record closed at shutdown", file.Entries)
	}
}

func TestTimelineRollsOverAtMidnight(t *testing.T) {
	start := time.Unix(0, 0)
	path := filepath.Join(t.TempDir(), "timeline.json")
	now := start
	app := newApp(func() time.Time { return now }, func() {})
	app.timelinePath = path
	app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: start.Add(time.Minute)})
	app.currentDate = now.Format("2006-01-02")

	now = start.Add(24 * time.Hour)
	app.rolloverLocked(now)

	if entries := app.Timeline(); len(entries) != 0 {
		t.Fatalf("entries after midnight = %d, want empty", len(entries))
	}
	file, err := store.LoadTimelineFile(path)
	if err != nil {
		t.Fatalf("load persisted timeline: %v", err)
	}
	if len(file.Entries) != 0 || file.Date != now.Format("2006-01-02") {
		t.Fatalf("persisted file after midnight = %+v, want empty for new day", file)
	}
}

func TestCountdownCRUDAndOrdering(t *testing.T) {
	now := time.Date(2026, 8, 12, 10, 0, 0, 0, time.Local)
	app := newApp(func() time.Time { return now }, func() {})

	date := func(s string) domain.Rule { return domain.Rule{Type: domain.RuleDate, Target: s} }
	if msg := app.AddCountdown("一周后", date("2026-08-19")); msg != "" {
		t.Fatalf("add 一周后 failed: %s", msg)
	}
	if msg := app.AddCountdown("今天到期", date("2026-08-12")); msg != "" {
		t.Fatalf("add 今天到期 failed: %s", msg)
	}
	if msg := app.AddCountdown("已过", date("2026-08-01")); msg != "" {
		t.Fatalf("add 已过 failed: %s", msg)
	}
	if msg := app.AddCountdown("", domain.Rule{Type: domain.RuleMonthly, Day: 32}); msg == "" {
		t.Fatal("invalid countdown should be rejected")
	}

	views := app.CountdownEvents()
	if len(views) != 3 {
		t.Fatalf("views = %d, want 3", len(views))
	}
	got := []string{views[0].Title, views[1].Title, views[2].Title}
	want := []string{"今天到期", "一周后", "已过"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
	if views[0].RemainingDays != 0 || views[2].RemainingDays != -11 {
		t.Fatalf("remaining days = %d, %d; want 0, -11", views[0].RemainingDays, views[2].RemainingDays)
	}

	if msg := app.UpdateCountdown(views[1].ID, "改标题", date("2026-08-20")); msg != "" {
		t.Fatalf("update failed: %s", msg)
	}
	if got := app.CountdownEvents()[1].Title; got != "改标题" {
		t.Fatalf("updated title = %q, want 改标题", got)
	}

	if !app.DeleteCountdown(views[0].ID) {
		t.Fatal("delete failed")
	}
	if got := len(app.CountdownEvents()); got != 2 {
		t.Fatalf("after delete views = %d, want 2", got)
	}
}

func TestCardOrderSaveLoadAndRoundTrip(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	app.cardOrderPath = filepath.Join(t.TempDir(), "card_order.json")

	if got := app.GetCardOrder(); len(got) != 0 {
		t.Fatalf("initial order = %v, want empty", got)
	}
	order := []string{"countdown:2", "reminder", "countdown:1"}
	if !app.SaveCardOrder(order) {
		t.Fatal("save failed")
	}
	if got := app.GetCardOrder(); len(got) != 3 || got[0] != "countdown:2" || got[2] != "countdown:1" {
		t.Fatalf("order = %v, want %v", got, order)
	}

	app2 := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	app2.cardOrderPath = app.cardOrderPath
	app2.loadCardOrderLocked()
	if got := app2.GetCardOrder(); len(got) != 3 || got[0] != "countdown:2" {
		t.Fatalf("round-trip order = %v, want %v", got, order)
	}
}

func TestAppDownloadAndApplyUpdateDevShortCircuit(t *testing.T) {
	oldVersion := version
	version = "dev"
	defer func() { version = oldVersion }()

	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	if msg := app.DownloadAndApplyUpdate(); msg == "" {
		t.Fatal("dev 版本应返回不支持自动更新的提示")
	}
}

func TestAppApplyUpdateAndRestartDevShortCircuit(t *testing.T) {
	oldVersion := version
	version = "dev"
	defer func() { version = oldVersion }()

	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	if msg := app.ApplyUpdateAndRestart(); msg == "" {
		t.Fatal("dev 版本应返回不支持自动更新的提示")
	}
}

func TestAppDownloadAndApplyUpdatePlatformUnsupported(t *testing.T) {
	if goruntime.GOOS == "windows" {
		t.Skip("非 Windows 平台行为测试，Windows 上跳过")
	}
	oldVersion := version
	version = "0.1.0"
	defer func() { version = oldVersion }()

	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	if msg := app.DownloadAndApplyUpdate(); msg == "" || !strings.Contains(msg, "不支持") {
		t.Fatalf("非 Windows 平台应返回不支持提示，got %q", msg)
	}
}

func TestAppDownloadAndApplyUpdateNoCachedURL(t *testing.T) {
	oldVersion := version
	version = "0.1.0"
	defer func() { version = oldVersion }()

	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	// 未调用 CheckForUpdates，无缓存下载地址
	msg := app.DownloadAndApplyUpdate()
	if msg == "" {
		t.Fatal("无缓存下载地址时应返回错误提示")
	}
	if goruntime.GOOS == "windows" && !strings.Contains(msg, "检查更新") {
		t.Fatalf("Windows 未缓存下载地址时应提示先检查更新，got %q", msg)
	}
}

func TestAppPendingUpdateInfoDevShortCircuit(t *testing.T) {
	oldVersion := version
	version = "dev"
	defer func() { version = oldVersion }()

	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	if info := app.PendingUpdateInfo(); info.Exists {
		t.Fatal("dev 版本不应暴露待更新状态")
	}
}

func TestAppPendingUpdateInfo(t *testing.T) {
	if goruntime.GOOS != "windows" {
		t.Skip("非 Windows 平台不暴露待更新状态")
	}
	oldVersion := version
	version = "0.1.6"
	defer func() { version = oldVersion }()

	exePath, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	_, newPath, versionPath, _ := exeUpdatePaths(exePath)
	// 清理可能的历史残留
	os.Remove(newPath)
	os.Remove(versionPath)
	defer func() {
		os.Remove(newPath)
		os.Remove(versionPath)
	}()

	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	if info := app.PendingUpdateInfo(); info.Exists {
		t.Fatal("无 .new 时不应暴露待更新状态")
	}
	if err := os.WriteFile(newPath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(versionPath, []byte("0.1.6"), 0o644); err != nil {
		t.Fatal(err)
	}
	info := app.PendingUpdateInfo()
	if !info.Exists || info.Version != "0.1.6" {
		t.Fatalf("PendingUpdateInfo = %+v, want exists+0.1.6", info)
	}
}

func TestAppCheckForUpdatesUpToDateCleansPending(t *testing.T) {
	oldVersion := version
	version = "0.1.6"
	defer func() { version = oldVersion }()

	exePath, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	_, newPath, versionPath, _ := exeUpdatePaths(exePath)
	os.Remove(newPath)
	os.Remove(versionPath)
	defer func() {
		os.Remove(newPath)
		os.Remove(versionPath)
	}()
	if err := os.WriteFile(newPath, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(versionPath, []byte("0.1.6"), 0o644); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"0.1.6","html_url":"https://example.com/release"}`))
	}))
	defer server.Close()

	oldClient := updateClient
	updateClient = server.Client()
	defer func() { updateClient = oldClient }()

	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	result := app.CheckForUpdates()
	if result.Status != model.UpdateStatusUpToDate {
		t.Fatalf("status = %q, want up-to-date", result.Status)
	}
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		t.Fatal("已是最新时 .new 应被清理")
	}
	if _, err := os.Stat(versionPath); !os.IsNotExist(err) {
		t.Fatal("已是最新时 .new.version 应被清理")
	}
}

func TestAppSilentModeKeepsWorkingWithoutNotification(t *testing.T) {
	start := time.Unix(0, 0)
	now := start
	notifications := 0
	app := newAppWithSettings(func() time.Time { return now }, func() { notifications++ }, model.Settings{ReminderMinutes: 1, RestMinutes: 1, NotificationsEnabled: false})

	app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: start})
	now = start.Add(90 * time.Second)
	status := app.Status()
	app.Status()

	if notifications != 0 {
		t.Fatalf("silent notifications = %d, want 0", notifications)
	}
	if status.State != "working" {
		t.Fatalf("silent state = %q, want working", status.State)
	}
	if status.ElapsedSeconds != 90 {
		t.Fatalf("silent elapsed = %d, want 90 (计时持续增长)", status.ElapsedSeconds)
	}
	if status.NotificationsEnabled {
		t.Fatal("status should report notifications disabled")
	}
	if entries := app.Timeline(); len(entries) != 1 || entries[0].Kind != "working" {
		t.Fatalf("silent timeline entries = %+v, want a single ongoing working entry", entries)
	}
}

func TestAppNotificationToggleTakesEffectAsynchronously(t *testing.T) {
	start := time.Unix(0, 0)
	now := start
	notifications := 0
	app := newApp(func() time.Time { return now }, func() { notifications++ })
	app.settingsPath = filepath.Join(t.TempDir(), "settings.json")

	app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: start})
	// 关闭：立即生效，到点不提醒
	now = start.Add(50 * time.Second)
	if !app.SaveSettings(1, 1, false) {
		t.Fatal("saving disabled notifications failed")
	}
	now = start.Add(60 * time.Second)
	app.Status()
	app.Status()
	if notifications != 0 {
		t.Fatalf("after disable notifications = %d, want 0 (关闭立即生效)", notifications)
	}
	if got := app.Status().State; got != "working" {
		t.Fatalf("after disable state = %q, want working", got)
	}
	// 静默 75 秒后开启：从开启时刻重新计满 60 秒，不补弹
	now = start.Add(75 * time.Second)
	if !app.SaveSettings(1, 1, true) {
		t.Fatal("saving enabled notifications failed")
	}
	app.Status()
	if notifications != 0 {
		t.Fatalf("after re-enable notifications = %d, want 0 (开启不补弹)", notifications)
	}
	now = start.Add(135 * time.Second)
	app.Status()
	if notifications != 1 {
		t.Fatalf("postponed notifications = %d, want 1 (顺延 75s+60s 时触发)", notifications)
	}
	if got := app.Status().State; got != "resting" {
		t.Fatalf("postponed state = %q, want resting", got)
	}
}

func TestAppSaveSettingsPersistsNotificationsEnabled(t *testing.T) {
	app := newApp(func() time.Time { return time.Unix(0, 0) }, func() {})
	app.settingsPath = filepath.Join(t.TempDir(), "settings.json")

	if !app.SaveSettings(45, 5, false) {
		t.Fatal("saving silent settings failed")
	}
	if got := app.GetSettings().NotificationsEnabled; got {
		t.Fatal("saved notificationsEnabled = true, want false")
	}
	loaded, err := store.LoadSettings(app.settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.NotificationsEnabled {
		t.Fatal("persisted notificationsEnabled = true, want false")
	}
}

func TestAppSilentModeSuppressesStartNotification(t *testing.T) {
	start := time.Unix(0, 0)
	starts := 0
	app := newAppWithSettings(func() time.Time { return start }, func() {}, model.Settings{ReminderMinutes: 1, RestMinutes: 1, NotificationsEnabled: false})
	app.notifyStarted = func(int) { starts++ }

	// 第一次打开软件后的首次有效活动：静默模式下不弹「新的工作段已开始」通知
	app.recordActivity(domain.EffectiveActivity{Kind: domain.KeyPress, At: start})
	if starts != 0 {
		t.Fatalf("silent start notifications = %d, want 0", starts)
	}
	if got := app.Status().State; got != "working" {
		t.Fatalf("silent state = %q, want working", got)
	}

	// 重新开启后恢复弹出
	now := start.Add(time.Minute)
	app.settings.NotificationsEnabled = true
	app.monitor.SetNotificationsEnabled(true, now)
	app.monitor.PauseForIdle()
	app.recordActivity(domain.EffectiveActivity{Kind: domain.Click, At: now.Add(time.Minute)})
	if starts != 1 {
		t.Fatalf("re-enabled start notifications = %d, want 1", starts)
	}
}
