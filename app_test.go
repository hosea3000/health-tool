package main

import (
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
