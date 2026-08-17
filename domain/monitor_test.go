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

func TestMonitorDefaultsToNotificationsEnabled(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewMonitor(start, 60*time.Minute, 3*time.Minute)
	if !m.NotificationsEnabled() {
		t.Fatal("notifications should default to enabled")
	}
}

func TestMonitorSilentModeSkipsReminderAtThreshold(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewMonitor(start, 60*time.Minute, 3*time.Minute)
	m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start})
	feedActivity(t, m, start, 60)
	m.SetNotificationsEnabled(false, start.Add(50*time.Minute))

	result := m.Advance(start.Add(60 * time.Minute))
	if result.Changed || result.Reminder {
		t.Fatalf("silent advance result = %+v, want empty", result)
	}
	if got := m.State(); got != Working {
		t.Fatalf("silent state at threshold = %s, want %s", got, Working)
	}
	if got := m.Elapsed(start.Add(90 * time.Minute)); got != 90*time.Minute {
		t.Fatalf("silent elapsed = %s, want 90m (计时持续增长)", got)
	}
	// 持续活动与推进都不应触发提醒
	m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start.Add(92 * time.Minute)})
	if result := m.Advance(start.Add(95 * time.Minute)); result.Changed || result.Reminder {
		t.Fatalf("silent ongoing result = %+v, want empty", result)
	}
}

func TestMonitorSilentModeIdlePauseStillWorks(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewMonitor(start, 60*time.Minute, 3*time.Minute)
	m.SetNotificationsEnabled(false, start)
	m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start})
	m.Advance(start.Add(5 * time.Minute))
	if got := m.State(); got != IdlePaused {
		t.Fatalf("silent idle state = %s, want %s", got, IdlePaused)
	}
	m.EffectiveActivity(EffectiveActivity{Kind: Click, At: start.Add(6 * time.Minute)})
	if got := m.State(); got != Working {
		t.Fatalf("silent resumed state = %s, want %s", got, Working)
	}
}

func TestMonitorReEnablePostponesReminder(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewMonitor(start, 60*time.Minute, 3*time.Minute)
	m.SetNotificationsEnabled(false, start)
	m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start})
	feedActivity(t, m, start, 80)

	// 静默超时 75 分钟后开启：从当前时刻重新计满 60 分钟，不补弹
	now := start.Add(75 * time.Minute)
	m.SetNotificationsEnabled(true, now)
	if result := m.Advance(now); result.Changed || result.Reminder {
		t.Fatalf("re-enable immediate result = %+v, want empty", result)
	}
	if got := m.State(); got != Working {
		t.Fatalf("re-enable state = %s, want %s", got, Working)
	}
	// 顺延后的提醒时点：已工作 75 分钟 + 设置 60 分钟 = 135 分钟
	feedActivity(t, m, start, 135)
	if result := m.Advance(start.Add(135 * time.Minute)); !result.Reminder {
		t.Fatal("postponed reminder did not fire at elapsed+reminder")
	}
	if got := m.State(); got != Resting {
		t.Fatalf("postponed rest state = %s, want %s", got, Resting)
	}
}

func TestMonitorReEnableOutsideWorkingUsesConfiguredDuration(t *testing.T) {
	start := time.Unix(0, 0)
	m := NewMonitor(start, 60*time.Minute, 3*time.Minute)
	m.SetNotificationsEnabled(false, start)
	m.SetNotificationsEnabled(true, start) // Waiting 状态下开启：无顺延副作用
	m.EffectiveActivity(EffectiveActivity{Kind: KeyPress, At: start})
	feedActivity(t, m, start, 60)
	if result := m.Advance(start.Add(60 * time.Minute)); !result.Reminder {
		t.Fatal("reminder after waiting-state re-enable did not fire on schedule")
	}
}
