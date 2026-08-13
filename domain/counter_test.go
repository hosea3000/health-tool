package domain

import (
	"testing"
	"time"
)

func mustTime(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		panic(err)
	}
	return t
}

func TestCounterPeriodKey(t *testing.T) {
	cases := []struct {
		period CounterPeriod
		date   string
		want   string
	}{
		{CounterDay, "2026-08-13", "2026-08-13"},
		{CounterDay, "2026-01-05", "2026-01-05"},
		{CounterMonth, "2026-08-13", "2026-08"},
		{CounterMonth, "2026-12-31", "2026-12"},
		{CounterYear, "2026-08-13", "2026"},
		{CounterNever, "2026-08-13", "all"},
		{CounterNever, "2027-01-01", "all"},
	}
	for _, c := range cases {
		got := Counter{Period: c.period}.PeriodKey(mustTime(c.date))
		if got != c.want {
			t.Errorf("PeriodKey(%s, %s) = %q, want %q", c.period, c.date, got, c.want)
		}
	}
}

func TestCounterCurrentCount(t *testing.T) {
	c := Counter{
		Period: CounterDay,
		Counts: map[string]int{"2026-08-12": 6, "2026-08-13": 5},
	}
	if got := c.CurrentCount(mustTime("2026-08-13")); got != 5 {
		t.Errorf("CurrentCount(08-13) = %d, want 5", got)
	}
	if got := c.CurrentCount(mustTime("2026-08-14")); got != 0 {
		t.Errorf("CurrentCount(08-14) = %d, want 0 (新桶)", got)
	}
}

func TestCounterCurrentCountNever(t *testing.T) {
	c := Counter{
		Period: CounterNever,
		Counts: map[string]int{"all": 42},
	}
	if got := c.CurrentCount(mustTime("2030-05-05")); got != 42 {
		t.Errorf("CurrentCount(never) = %d, want 42", got)
	}
}

func TestCounterValidate(t *testing.T) {
	valid := Counter{Name: "喝水", Period: CounterDay}
	if err := valid.Validate(); err != nil {
		t.Errorf("valid counter rejected: %v", err)
	}
	bad := []Counter{
		{Name: "", Period: CounterDay},
		{Name: "一二三四五六七八九十一二三四五六七八九十一二", Period: CounterDay},
		{Name: "喝水", Period: "weekly"},
		{Name: "喝水", Period: CounterDay, Goal: -1},
	}
	for _, c := range bad {
		if err := c.Validate(); err == nil {
			t.Errorf("invalid counter accepted: %+v", c)
		}
	}
}

func TestCounterPeriodLabel(t *testing.T) {
	cases := []struct {
		period CounterPeriod
		want   string
	}{
		{CounterDay, "今日"},
		{CounterMonth, "本月"},
		{CounterYear, "今年"},
		{CounterNever, "累计"},
	}
	for _, c := range cases {
		if got := (Counter{Period: c.period}).PeriodLabel(); got != c.want {
			t.Errorf("PeriodLabel(%s) = %q, want %q", c.period, got, c.want)
		}
	}
}
