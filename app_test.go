package main

import (
	"context"
	"health-tool/domain"
	"path/filepath"
	"testing"
	"time"
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

	if !app.SaveSettings(45, 5) {
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
	settings := Settings{ReminderMinutes: 1, RestMinutes: 1}
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
	_ = saveTimelineFile(path, timelineFile{
		Date:    now.Format("2006-01-02"),
		SavedAt: savedAt,
		Entries: []TimelineEntry{
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
	_ = saveTimelineFile(path, timelineFile{
		Date:    now.AddDate(0, 0, -1).Format("2006-01-02"),
		SavedAt: now,
		Entries: []TimelineEntry{{Kind: "working", StartedAt: now}},
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

	file, err := loadTimelineFile(path)
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

	file, err := loadTimelineFile(path)
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
	file, err := loadTimelineFile(path)
	if err != nil {
		t.Fatalf("load persisted timeline: %v", err)
	}
	if len(file.Entries) != 0 || file.Date != now.Format("2006-01-02") {
		t.Fatalf("persisted file after midnight = %+v, want empty for new day", file)
	}
}
