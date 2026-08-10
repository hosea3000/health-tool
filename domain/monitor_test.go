package domain

import (
	"testing"
	"time"
)

func feedActivity(t *testing.T, m *Monitor, start time.Time, until int) {
	t.Helper()
	for minute := 4; minute < until; minute += 4 {
		m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start.Add(time.Duration(minute) * time.Minute)})
	}
}

func TestMonitorStartsWorkingOnFirstActivity(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewMonitor(start, DefaultReminderDuration, DefaultRestDuration)

	if got := m.State(); got != Waiting {
		t.Fatalf("startup state = %s, want %s", got, Waiting)
	}
	m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start.Add(time.Minute)})
	if got := m.State(); got != Working {
		t.Fatalf("first activity state = %s, want %s", got, Working)
	}
}

func TestMonitorPausesAfterFiveMinutesAndRestartsOnActivity(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewMonitor(start, DefaultReminderDuration, DefaultRestDuration)
	m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start})
	m.Advance(start.Add(5 * time.Minute))

	if got := m.State(); got != IdlePaused {
		t.Fatalf("idle state = %s, want %s", got, IdlePaused)
	}
	m.EffectiveActivity(EffectiveActivity{Kind: Click, At: start.Add(6 * time.Minute)})
	if got := m.State(); got != Working {
		t.Fatalf("auto-resumed state = %s, want %s", got, Working)
	}
}

func TestMonitorEntersRestAndIgnoresInputUntilRestEnds(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewMonitor(start, 60*time.Minute, 3*time.Minute)
	m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start})
	feedActivity(t, m, start, 60)

	if result := m.Advance(start.Add(60 * time.Minute)); !result.Reminder {
		t.Fatal("reminder did not start rest")
	}
	if got := m.State(); got != Resting {
		t.Fatalf("rest state = %s, want %s", got, Resting)
	}
	m.EffectiveActivity(EffectiveActivity{Kind: Click, At: start.Add(61 * time.Minute)})
	if got := m.State(); got != Resting {
		t.Fatalf("input during rest state = %s, want %s", got, Resting)
	}
	if got := m.RestRemaining(start.Add(61 * time.Minute)); got != 2*time.Minute {
		t.Fatalf("rest remaining = %s, want 2m", got)
	}
	m.EffectiveActivity(EffectiveActivity{Kind: Click, At: start.Add(63 * time.Minute)})
	if got := m.State(); got != Working {
		t.Fatalf("post-rest activity state = %s, want %s", got, Working)
	}
}

func TestMonitorUsesConfiguredReminderDuration(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewMonitor(start, 30*time.Minute, 3*time.Minute)
	m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start})
	feedActivity(t, m, start, 30)
	if result := m.Advance(start.Add(30 * time.Minute)); !result.Reminder {
		t.Fatal("configured duration did not trigger a reminder")
	}
}

func TestReminderDurationValidation(t *testing.T) {
	for _, duration := range []time.Duration{1 * time.Minute, 60 * time.Minute, 180 * time.Minute} {
		if !ValidReminderDuration(duration) {
			t.Fatalf("duration %s was rejected", duration)
		}
	}
	for _, duration := range []time.Duration{2 * time.Minute, 12 * time.Minute, 185 * time.Minute} {
		if ValidReminderDuration(duration) {
			t.Fatalf("duration %s was accepted", duration)
		}
	}
}

func TestRestDurationValidation(t *testing.T) {
	if !ValidRestDuration(1*time.Minute) || !ValidRestDuration(30*time.Minute) {
		t.Fatal("valid rest duration was rejected")
	}
	if ValidRestDuration(31 * time.Minute) {
		t.Fatal("invalid rest duration was accepted")
	}
}

func TestMonitorRejectsUnknownActivity(t *testing.T) {
	m := NewMonitor(time.Unix(0, 0), DefaultReminderDuration, DefaultRestDuration)
	m.EffectiveActivity(EffectiveActivity{Kind: ActivityKind(255), At: time.Unix(1, 0)})
	if got := m.State(); got != Waiting {
		t.Fatalf("unknown activity state = %s, want %s", got, Waiting)
	}
}

func TestMonitorIgnoresOutOfOrderActivity(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewMonitor(start, DefaultReminderDuration, DefaultRestDuration)
	m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start})
	m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start.Add(time.Minute)})
	m.EffectiveActivity(EffectiveActivity{Kind: Click, At: start.Add(30 * time.Second)})
	if result := m.Advance(start.Add(5*time.Minute + 30*time.Second)); result.Changed {
		t.Fatal("out-of-order activity extended the active period")
	}
}
